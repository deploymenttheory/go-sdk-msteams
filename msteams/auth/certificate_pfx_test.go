package auth

import (
	"errors"
	"os"
	"testing"
	"time"
)

func TestPFXSupportedFormatAndPasswordFailures(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/client-credentials-sha1.pfx")
	if err != nil {
		t.Fatal(err)
	}
	original := string(data)
	password := []byte("fixture-only")
	parsed, err := parsePFXCertificate(data, password)
	if err != nil {
		t.Fatal(err)
	}
	if err := parsed.validateTime(time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	if string(data) != original || string(password) != "fixture-only" {
		t.Fatal("PFX parser mutated caller input")
	}
	for _, wrong := range [][]byte{nil, []byte("wrong-password")} {
		if _, err := parsePFXCertificate(data, wrong); !errors.Is(err, ErrInvalidCertificate) {
			t.Fatalf("password failure: %v", err)
		}
	}
}

func TestPFXUnsupportedSHA256MACFailsExplicitly(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/client-credentials-sha256.pfx")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parsePFXCertificate(data, []byte("fixture-only")); !errors.Is(err, ErrInvalidCertificate) {
		t.Fatalf("unsupported PKCS12 format: %v", err)
	}
}
