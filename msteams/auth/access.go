package auth

import "errors"

// Operation failures distinguish credential rejection from authorization and
// evidenced operation restrictions. A token exchange never proves operation access.
var (
	ErrUnauthenticated             = errors.New("operation authentication failed")
	ErrForbidden                   = errors.New("operation access is forbidden")
	ErrUnsupportedOperationMode    = errors.New("operation does not support this authentication mode")
	ErrTenantCapabilityUnavailable = errors.New("required tenant capability is unavailable")
)
