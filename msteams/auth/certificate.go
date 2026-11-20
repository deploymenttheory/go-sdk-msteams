package auth

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

var (
	ErrInvalidCertificate     = errors.New("certificate material must contain one matching RSA signing certificate and private key")
	ErrCertificateExpired     = errors.New("signing certificate has expired")
	ErrCertificateNotYetValid = errors.New("signing certificate is not yet valid")
)

// parsedCertificate retains only Microsoft's credential and the signing certificate's
// validity window. It does not consult operating-system certificate stores.
type parsedCertificate struct {
	credential confidential.Credential
	notBefore  time.Time
	notAfter   time.Time
}

func (parsedCertificate) String() string     { return "msteams certificate credential (redacted)" }
func (c parsedCertificate) GoString() string { return c.String() }

func (c parsedCertificate) validateTime(now time.Time) error {
	if now.Before(c.notBefore) {
		return ErrCertificateNotYetValid
	}
	if !now.Before(c.notAfter) {
		return ErrCertificateExpired
	}
	return nil
}

func parsePEMCertificate(certificate, key []byte) (parsedCertificate, error) {
	certificate = bytes.TrimSpace(certificate)
	key = bytes.TrimSpace(key)
	if !validPEMBlocks(certificate, []string{"CERTIFICATE"}, false) ||
		!validPEMBlocks(key, []string{"PRIVATE KEY", "RSA PRIVATE KEY"}, true) {
		return parsedCertificate{}, ErrInvalidCertificate
	}
	// Never append into caller-owned buffers, including when spare capacity exists.
	material := make([]byte, 0, len(certificate)+len(key)+1)
	material = append(material, certificate...)
	material = append(material, '\n')
	material = append(material, key...)
	defer clear(material)
	certs, privateKey, err := confidential.CertFromPEM(material, "")
	if err != nil {
		return parsedCertificate{}, ErrInvalidCertificate
	}
	return validateCertificate(certs, privateKey)
}

// validPEMBlocks prevents pem.Decode's permissive search from skipping malformed
// blocks, foreign material or junk. Microsoft still performs semantic DER parsing.
func validPEMBlocks(data []byte, allowedTypes []string, single bool) bool {
	count := 0
	for len(bytes.TrimSpace(data)) > 0 {
		data = bytes.TrimSpace(data)
		var blockType string
		for _, candidate := range allowedTypes {
			if bytes.HasPrefix(data, []byte("-----BEGIN "+candidate+"-----")) {
				blockType = candidate
				break
			}
		}
		if blockType == "" {
			return false
		}
		endMarker := []byte("-----END " + blockType + "-----")
		end := bytes.Index(data, endMarker)
		if end < 0 {
			return false
		}
		end += len(endMarker)
		beginMarker := []byte("-----BEGIN " + blockType + "-----")
		if bytes.Contains(data[len(beginMarker):end], []byte("-----BEGIN ")) {
			return false
		}
		block, remaining := pem.Decode(data[:end])
		if block == nil || block.Type != blockType || len(bytes.TrimSpace(remaining)) != 0 {
			return false
		}
		count++
		if single && count > 1 {
			return false
		}
		data = data[end:]
	}
	return count > 0
}

func parsePFXCertificate(data, password []byte) (parsedCertificate, error) {
	// PKCS#12 is a DER SEQUENCE. Reject PEM rather than allowing the upstream
	// auto-detection fallback to override the caller's explicit format selection.
	if len(data) == 0 || data[0] != 0x30 {
		return parsedCertificate{}, ErrInvalidCertificate
	}
	if block, _ := pem.Decode(data); block != nil {
		return parsedCertificate{}, ErrInvalidCertificate
	}
	// The pinned Azure Identity parser supports PKCS#12 with SHA1 MAC. Modern SHA256
	// MAC bundles are rejected; callers can supply equivalent unencrypted PEM instead.
	certs, key, err := azidentity.ParseCertificates(data, password)
	if err != nil {
		return parsedCertificate{}, ErrInvalidCertificate
	}
	return validateCertificate(certs, key)
}

func validateCertificate(certs []*x509.Certificate, key crypto.PrivateKey) (parsedCertificate, error) {
	privateKey, ok := key.(*rsa.PrivateKey)
	if !ok || privateKey == nil {
		return parsedCertificate{}, ErrInvalidCertificate
	}
	if err := privateKey.Validate(); err != nil {
		return parsedCertificate{}, ErrInvalidCertificate
	}
	var signing *x509.Certificate
	for _, cert := range certs {
		if cert == nil {
			return parsedCertificate{}, ErrInvalidCertificate
		}
		publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
		if ok && privateKey.PublicKey.Equal(publicKey) {
			// Multiple matching leaf certificates make validity/renewal selection ambiguous.
			if signing != nil {
				return parsedCertificate{}, ErrInvalidCertificate
			}
			signing = cert
		}
	}
	if signing == nil || signing.NotBefore.IsZero() || !signing.NotAfter.After(signing.NotBefore) {
		return parsedCertificate{}, ErrInvalidCertificate
	}
	credential, err := confidential.NewCredFromCert(certs, privateKey)
	if err != nil {
		return parsedCertificate{}, ErrInvalidCertificate
	}
	return parsedCertificate{credential: credential, notBefore: signing.NotBefore, notAfter: signing.NotAfter}, nil
}
