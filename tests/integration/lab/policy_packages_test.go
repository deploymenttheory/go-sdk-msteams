package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var errPackageIdentity = errors.New("invalid private delegated package identity")

type packageIdentity struct {
	Tenant   string    `json:"tenant_id"`
	Identity string    `json:"identity_id"`
	Token    string    `json:"access_token"`
	Expires  time.Time `json:"expires_at"`
	Cloud    string    `json:"cloud"`
	User     string    `json:"user_id"`
}

func (packageIdentity) String() string     { return "delegated package identity (redacted)" }
func (v packageIdentity) GoString() string { return v.String() }

func decodePackageIdentity(raw []byte, tenant string, now time.Time) (packageIdentity, error) {
	var v packageIdentity
	if len(raw) > labkit.MaxOracleBytes || !labkit.DirectoryID.MatchString(tenant) {
		return v, errPackageIdentity
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !strings.EqualFold(v.Tenant, tenant) || v.Identity == "" || v.Token == "" || !v.Expires.After(now.Add(time.Minute)) {
		return packageIdentity{}, errPackageIdentity
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return packageIdentity{}, errPackageIdentity
	}
	return v, nil
}

func delegatedPackageClient(t *testing.T) (*msteams.Client, string) {
	t.Helper()
	raw, err := labkit.PrivateFile(os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	if err != nil {
		t.Fatal("private lab tenant binding unavailable")
	}
	defer clear(raw)
	var expected struct {
		Tenant string `json:"tenant_id"`
	}
	if json.Unmarshal(raw, &expected) != nil {
		t.Fatal("invalid lab tenant binding")
	}
	data, err := labkit.PrivateFile(os.Getenv("MSTEAMS_LAB_PACKAGE_DELEGATED_FILE"))
	if err != nil {
		t.Fatal("private delegated package identity unavailable")
	}
	defer clear(data)
	v, err := decodePackageIdentity(data, expected.Tenant, time.Now())
	if err != nil {
		t.Fatal("invalid delegated package identity or expiry")
	}
	environment, err := auth.ResolveEnvironment(v.Cloud, nil)
	if err != nil {
		t.Fatal("invalid explicit delegated cloud")
	}
	sdk, err := msteams.NewClient(&auth.Config{Method: auth.SuppliedTokens, TenantID: v.Tenant, IdentityID: v.Identity, TokenMode: auth.Delegated, Cloud: v.Cloud, AccessTokens: map[string]auth.Token{environment.Endpoints.TeamsAdminResource: {AccessToken: v.Token, ExpiresAt: v.Expires}}})
	v.Token = ""
	if err != nil {
		t.Fatal("delegated package client unavailable")
	}
	t.Cleanup(func() { _ = sdk.Close() })
	return sdk, v.User
}

// TestAccPolicyPackagesReadOnly requires a separately supplied delegated context.
// With the existing SPN it verifies the documented local mode guard, then skips
// tenant qualification. It never creates, changes or assigns a package.
func TestAccPolicyPackagesReadOnly(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	if os.Getenv("MSTEAMS_LAB_PACKAGE_DELEGATED_FILE") == "" {
		sdk, _, _ := labkit.SDKClient(t)
		_, r, e := sdk.TeamsAdminAPI.PolicyPackages.List(ctx)
		if !errors.Is(e, client.ErrAuthenticationMode) || r != nil {
			t.Fatal("configured lab identity did not enforce the documented package mode restriction")
		}
		t.Skip("application lab identity rejected locally; delegated package context is not configured; no tenant request ran")
	}
	sdk, user := delegatedPackageClient(t)
	values, r, e := sdk.TeamsAdminAPI.PolicyPackages.List(ctx)
	if e != nil {
		status := 0
		if r != nil {
			status = r.StatusCode()
		}
		t.Fatalf("delegated package list failed (HTTP %d); no payload retained", status)
	}
	checked := 0
	for _, v := range values {
		if checked == 5 {
			break
		}
		if v == nil || v.Name == nil || *v.Name == "" {
			t.Fatal("package list returned no stable name")
		}
		item, _, e := sdk.TeamsAdminAPI.PolicyPackages.GetByName(ctx, *v.Name)
		if e != nil || item == nil || item.Name == nil || *item.Name != *v.Name {
			t.Fatal("selected package read did not preserve its identity")
		}
		checked++
	}
	if user != "" {
		if _, _, e := sdk.TeamsAdminAPI.PolicyPackages.GetForUser(ctx, user); e != nil {
			t.Fatal("selected user package read unavailable")
		}
		if _, _, e := sdk.TeamsAdminAPI.PolicyPackages.RecommendationsForUser(ctx, user); e != nil {
			t.Fatal("selected user package recommendations unavailable")
		}
	}
	t.Logf("delegated package list and %d selected reads observed; mutation/assignment and effective behavior remain separately qualified", checked)
}

func TestUnitPackageDelegatedIdentityRejectsWrongTenantExpiryAndExtraData(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	now := time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)
	valid := `{"tenant_id":"` + tenant + `","identity_id":"fixture-user","access_token":"fixture-token","expires_at":"2026-11-21T13:00:00Z"}`
	if _, e := decodePackageIdentity([]byte(valid), tenant, now); e != nil {
		t.Fatal("valid private identity rejected")
	}
	for _, raw := range []string{`{}`, strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, "13:00:00", "12:00:00", 1), strings.Replace(valid, `"fixture-token"`, `""`, 1), valid + ` {}`, strings.TrimSuffix(valid, "}") + `,"unexpected":false}`} {
		if _, e := decodePackageIdentity([]byte(raw), tenant, now); e == nil {
			t.Fatal("invalid private identity accepted")
		}
	}
}
