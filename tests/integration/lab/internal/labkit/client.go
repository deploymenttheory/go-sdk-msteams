package labkit

import (
	"encoding/json"
	"io"
	"os"
	"runtime"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

func PrivateFile(path string) ([]byte, error) {
	//nolint:gosec // G304: explicitly selected private lab files, after live opt-in.
	file, err := os.Open(path)
	if err != nil {
		return nil, lab.ErrConfig
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > 1<<20 {
		return nil, lab.ErrConfig
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		return nil, lab.ErrConfig
	}
	data, err := io.ReadAll(io.LimitReader(file, (1<<20)+1))
	if err != nil || len(data) > 1<<20 {
		return nil, lab.ErrConfig
	}
	return data, nil
}

func SDKClient(t *testing.T) (*msteams.Client, string, string) {
	t.Helper()
	file := os.Getenv("MSTEAMS_LAB_IDENTITY_FILE")
	if file == "" {
		t.Fatal("explicit private lab identity file is required")
	}
	raw, err := PrivateFile(file)
	if err != nil {
		t.Fatal("private lab identity unavailable")
	}
	defer clear(raw)
	var identity struct {
		Tenant      string `json:"tenant_id"`
		Client      string `json:"client_id"`
		Secret      string `json:"client_secret"`
		Certificate string `json:"certificate_file"`
		Key         string `json:"private_key_file"`
	}
	if json.Unmarshal(raw, &identity) != nil {
		t.Fatal("invalid lab identity")
	}
	method := os.Getenv("MSTEAMS_LAB_AUTH_METHOD")
	cfg := &auth.Config{Method: auth.Certificate, TenantID: identity.Tenant, ClientID: identity.Client}
	switch method {
	case "", "certificate":
		cfg.Certificate, err = PrivateFile(identity.Certificate)
		if err != nil {
			t.Fatal("private lab certificate unavailable")
		}
		cfg.PrivateKey, err = PrivateFile(identity.Key)
		if err != nil {
			t.Fatal("private lab key unavailable")
		}
		defer clear(cfg.PrivateKey)
	case "client-secret":
		cfg.Method = auth.ClientSecret
		cfg.ClientSecret = identity.Secret
	default:
		t.Fatal("unsupported explicit lab auth method")
	}
	sdk, err := msteams.NewClient(cfg)
	identity.Secret = ""
	cfg.ClientSecret = ""
	if err != nil {
		t.Fatal("lab SDK construction failed")
	}
	t.Cleanup(func() {
		if err := sdk.Close(); err != nil {
			t.Error("lab SDK cleanup failed")
		}
	})
	return sdk, identity.Tenant + "|commercial|" + identity.Client, identity.Client
}
