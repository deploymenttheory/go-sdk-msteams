package client

import "github.com/deploymenttheory/go-sdk-msteams/msteams/auth"

// SessionInfo is credential-free configuration metadata for one local client.
// Cloud is the selected isolation profile, not proof that its routing or any
// operation is supported. Tenant and identity identifiers require explicit handling.
type SessionInfo struct {
	TenantID   string
	Cloud      string
	IdentityID string
	Mode       auth.Mode
	Method     auth.Method
}

// Info copies local configuration under the session lock without authenticating,
// querying a tenant, decoding a token or inspecting another client's state.
func (s *Session) Info() (SessionInfo, error) {
	if s == nil {
		return SessionInfo{}, ErrClosed
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed || s.auth == nil {
		return SessionInfo{}, ErrClosed
	}
	return SessionInfo{TenantID: s.binding.TenantID, Cloud: s.binding.Cloud, IdentityID: s.binding.IdentityID,
		Mode: s.binding.Mode, Method: s.auth.Method}, nil
}

func (SessionInfo) String() string        { return "msteams session information (redacted)" }
func (info SessionInfo) GoString() string { return info.String() }
