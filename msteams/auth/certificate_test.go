package auth

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"
)

func certificateFixture(t *testing.T) ([]byte, []byte, time.Time) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
	template := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Synthetic SDK fixture"}, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return cert, keyPEM, now
}

func TestPEMCertificatePairValidityAndOwnership(t *testing.T) {
	t.Parallel()
	cert, key, now := certificateFixture(t)
	// Spare capacity exposes accidental append into the caller's certificate buffer.
	backing := bytes.Repeat([]byte{'X'}, len(cert)+len(key)+1)
	copy(backing, cert)
	cert = backing[:len(cert)]
	originalBacking := bytes.Clone(backing)
	originalCert, originalKey := string(cert), string(key)
	parsed, err := parsePEMCertificate(cert, key)
	if err != nil {
		t.Fatal(err)
	}
	if string(cert) != originalCert || string(key) != originalKey {
		t.Fatal("parser modified caller material")
	}
	if !bytes.Equal(backing, originalBacking) {
		t.Fatal("parser wrote into spare capacity of the caller's certificate buffer")
	}
	for _, test := range []struct {
		at   time.Time
		want error
	}{
		{now, nil}, {now.Add(-time.Hour), nil}, {now.Add(-time.Hour - time.Nanosecond), ErrCertificateNotYetValid},
		{now.Add(time.Hour), ErrCertificateExpired},
	} {
		if err := parsed.validateTime(test.at); !errors.Is(err, test.want) {
			t.Fatalf("validity: %v", err)
		}
	}
	for _, output := range []string{fmt.Sprintf("%v", parsed), fmt.Sprintf("%+v", parsed), fmt.Sprintf("%#v", parsed)} {
		if strings.Contains(output, "PRIVATE") || strings.Contains(output, "Synthetic") {
			t.Fatal("credential diagnostic exposed source material")
		}
	}
}

func TestPEMCertificateAcceptsPKCS8RSAKey(t *testing.T) {
	t.Parallel()
	cert, key, now := certificateFixture(t)
	block, _ := pem.Decode(key)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	pkcs8, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8})
	parsed, err := parsePEMCertificate(cert, keyPEM)
	if err != nil {
		t.Fatal(err)
	}
	if err := parsed.validateTime(now); err != nil {
		t.Fatal(err)
	}
}

func TestPEMInvalidAndMismatchedMaterialIsRedacted(t *testing.T) {
	t.Parallel()
	cert, key, _ := certificateFixture(t)
	_, otherKey, _ := certificateFixture(t)
	for _, test := range []struct{ cert, key []byte }{
		{cert, otherKey}, {cert, nil}, {nil, key}, {[]byte("sensitive-invalid-certificate"), []byte("sensitive-invalid-key")},
		{append(append([]byte{}, cert...), cert...), key},
		{cert, append(append([]byte{}, key...), key...)},
	} {
		_, err := parsePEMCertificate(test.cert, test.key)
		if !errors.Is(err, ErrInvalidCertificate) {
			t.Fatalf("invalid pair: %v", err)
		}
		if strings.Contains(err.Error(), "sensitive") {
			t.Fatal("parse error included source material")
		}
	}
}

func TestPEMRejectsUnsupportedAndEncryptedKeys(t *testing.T) {
	t.Parallel()
	cert, key, _ := certificateFixture(t)
	block, _ := pem.Decode(key)
	block.Headers = map[string]string{"DEK-Info": "DES-EDE3-CBC,0000000000000000", "Proc-Type": "4,ENCRYPTED"}
	if _, err := parsePEMCertificate(cert, pem.EncodeToMemory(block)); !errors.Is(err, ErrInvalidCertificate) {
		t.Fatalf("encrypted key: %v", err)
	}
	ec, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := x509.MarshalPKCS8PrivateKey(ec)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parsePEMCertificate(cert, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: raw})); !errors.Is(err, ErrInvalidCertificate) {
		t.Fatalf("unsupported key: %v", err)
	}
}

func TestInvalidPFXIsRedacted(t *testing.T) {
	t.Parallel()
	for _, data := range [][]byte{nil, []byte("sensitive-invalid-pfx")} {
		if _, err := parsePFXCertificate(data, []byte("sensitive-password")); !errors.Is(err, ErrInvalidCertificate) {
			t.Fatalf("invalid PFX: %v", err)
		}
	}
}

func TestPEMRejectsIgnoredJunkAndBundledKeys(t *testing.T) {
	t.Parallel()
	cert, key, _ := certificateFixture(t)
	join := func(parts ...[]byte) []byte { return bytes.Join(parts, nil) }
	brokenCertificate := []byte("-----BEGIN CERTIFICATE-----\nnot-base64!\n-----END CERTIFICATE-----\n")
	brokenKey := []byte("-----BEGIN RSA PRIVATE KEY-----\nnot-base64!\n-----END RSA PRIVATE KEY-----\n")
	cases := []struct {
		name      string
		cert, key []byte
	}{
		{name: "certificate prefix junk", cert: join([]byte("junk"), cert), key: key},
		{name: "certificate suffix junk", cert: join(cert, []byte("junk")), key: key},
		{name: "key prefix junk", cert: cert, key: join([]byte("junk"), key)},
		{name: "key suffix junk", cert: cert, key: join(key, []byte("junk"))},
		{name: "malformed certificate before valid block", cert: join(brokenCertificate, cert), key: key},
		{name: "malformed key before valid block", cert: cert, key: join(brokenKey, key)},
		{name: "unfinished certificate before valid block", cert: join([]byte("-----BEGIN CERTIFICATE-----\n"), cert), key: key},
		{name: "unfinished key before valid block", cert: cert, key: join([]byte("-----BEGIN RSA PRIVATE KEY-----\n"), key)},
		{name: "hidden key in certificate", cert: join(cert, key), key: []byte("ignored-junk")},
		{name: "hidden certificate in key", cert: []byte("ignored-junk"), key: join(cert, key)},
		{name: "certificate in key input", cert: cert, key: join(key, cert)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := parsePEMCertificate(tc.cert, tc.key); !errors.Is(err, ErrInvalidCertificate) {
				t.Fatalf("expected strict PEM rejection, got %v", err)
			}
		})
	}
}

func TestPEMAcceptsSurroundingWhitespace(t *testing.T) {
	t.Parallel()
	cert, key, now := certificateFixture(t)
	padded := func(data []byte) []byte { return bytes.Join([][]byte{[]byte(" \r\n\t"), data, []byte("\t \r\n")}, nil) }
	parsed, err := parsePEMCertificate(padded(cert), padded(key))
	if err != nil {
		t.Fatal(err)
	}
	if err := parsed.validateTime(now); err != nil {
		t.Fatal(err)
	}
}

func TestPFXRejectsPEMFallback(t *testing.T) {
	t.Parallel()
	cert, key, _ := certificateFixture(t)
	bundle := bytes.Join([][]byte{cert, key}, nil)
	if _, err := parsePFXCertificate(bundle, nil); !errors.Is(err, ErrInvalidCertificate) {
		t.Fatalf("PFX parser accepted a PEM bundle: %v", err)
	}
	// A sequence tag alone cannot prevent the upstream PEM scanner from skipping
	// leading junk and selecting a later PEM bundle instead of decoding PKCS#12.
	prefixedBundle := bytes.Join([][]byte{{0x30}, bundle}, nil)
	if _, err := parsePFXCertificate(prefixedBundle, nil); !errors.Is(err, ErrInvalidCertificate) {
		t.Fatalf("PFX parser accepted a prefixed PEM bundle: %v", err)
	}
}
