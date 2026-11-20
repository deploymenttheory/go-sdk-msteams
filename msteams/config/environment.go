// Package config loads explicitly requested SDK configuration.
package config

import (
	"os"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

// Environment is unvalidated configuration. Explicit client options are applied
// before parsing so they can replace malformed environment defaults.
type Environment struct {
	Auth       *auth.Config
	TokensJSON string
	Timeout    string
}

// FromEnv reads only documented MSTEAMS variables. It never reads credential files.
func FromEnv() Environment {
	return Environment{
		Auth: &auth.Config{
			Method: auth.Method(os.Getenv("MSTEAMS_AUTH_METHOD")), TenantID: os.Getenv("MSTEAMS_TENANT_ID"),
			ClientID: os.Getenv("MSTEAMS_CLIENT_ID"), ClientSecret: os.Getenv("MSTEAMS_CLIENT_SECRET"),
			Username: os.Getenv("MSTEAMS_USERNAME"), Password: os.Getenv("MSTEAMS_PASSWORD"),
			Certificate: []byte(os.Getenv("MSTEAMS_CERTIFICATE_PEM")), PrivateKey: []byte(os.Getenv("MSTEAMS_PRIVATE_KEY_PEM")),
			TokenMode: auth.Mode(os.Getenv("MSTEAMS_TOKEN_MODE")), IdentityID: os.Getenv("MSTEAMS_IDENTITY_ID"),
			AssertionFile: os.Getenv("MSTEAMS_ASSERTION_FILE"),
		},
		TokensJSON: os.Getenv("MSTEAMS_ACCESS_TOKENS"),
		Timeout:    os.Getenv("MSTEAMS_HTTP_TIMEOUT"),
	}
}
