package auth

import (
	"context"
	"encoding/json"
	"errors"
)

func loginCallbackError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}
	return ErrLoginCallback
}

var (
	ErrDelegatedAuthentication = errors.New("delegated authentication failed; check application consent and the selected sign-in method")
	ErrAuthenticationRequired  = errors.New("interactive authentication or consent is required; select browser or device-code sign-in explicitly")
	ErrAuthenticationDenied    = errors.New("authentication was denied; inspect Entra diagnostics and tenant sign-in policy")
	ErrDeviceCodeExpired       = errors.New("device code expired; begin a new device-code sign-in")
	ErrLoginCallback           = errors.New("sign-in callback failed")
	ErrFederatedPassword       = errors.New("federated password endpoint is not trusted; configure exact federation endpoints or select browser/device sign-in")
)

// DelegatedError exposes OAuth classification and numeric Entra diagnostic codes.
// Numeric codes are diagnostic only; OAuth error values drive classification.
// Descriptions, claims, URLs, codes and tokens from the response are never retained.
type DelegatedError struct {
	OAuthCode  string
	ErrorCodes []int
}

func (e DelegatedError) Unwrap() error {
	switch e.OAuthCode {
	case "interaction_required", "consent_required", "login_required":
		return ErrAuthenticationRequired
	case "access_denied", "authorization_declined", "invalid_grant":
		return ErrAuthenticationDenied
	case "expired_token":
		return ErrDeviceCodeExpired
	default:
		return ErrDelegatedAuthentication
	}
}
func (e DelegatedError) Error() string    { return e.Unwrap().Error() }
func (e DelegatedError) GoString() string { return e.Error() }

func delegatedFailure(content []byte) *DelegatedError {
	var result struct {
		Error string `json:"error"`
		Codes []int  `json:"error_codes"`
	}
	if json.Unmarshal(content, &result) != nil {
		return &DelegatedError{}
	}
	if result.Error == "authorization_pending" || result.Error == "slow_down" {
		return nil
	}
	failure := &DelegatedError{}
	switch result.Error {
	case "interaction_required", "consent_required", "login_required", "access_denied", "authorization_declined", "invalid_grant", "expired_token", "invalid_request", "invalid_client", "invalid_scope", "unauthorized_client", "unsupported_grant_type", "temporarily_unavailable", "server_error", "bad_verification_code":
		failure.OAuthCode = result.Error
	}
	for _, code := range result.Codes {
		if code > 0 && code < 100000000 {
			failure.ErrorCodes = append(failure.ErrorCodes, code)
		}
	}
	return failure
}
