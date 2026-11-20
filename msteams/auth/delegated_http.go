package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// delegatedHTTP adds only the evidenced device-code and user-realm routes to the
// application authority guard. It captures safe OAuth classification before MSAL
// consumes an error body and enforces the device service's polling interval.
type delegatedHTTP struct {
	federation  map[string]*url.URL
	base        *authorityHTTP
	username    string
	mu          sync.Mutex
	lastFailure *DelegatedError
	polling     bool
	interval    time.Duration
	nextPoll    time.Time
	now         func() time.Time
	wait        func(context.Context, time.Duration) error
}

func newDelegatedHTTP(client *http.Client, tenant, username string) *delegatedHTTP {
	return &delegatedHTTP{base: &authorityHTTP{client: client, tenant: tenant}, username: username, now: time.Now, wait: waitForLogin}
}

func (*delegatedHTTP) CloseIdleConnections() {}
func (a *delegatedHTTP) resetFailure()       { a.mu.Lock(); a.lastFailure = nil; a.mu.Unlock() }
func (a *delegatedHTTP) failure() *DelegatedError {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.lastFailure
}

func (a *delegatedHTTP) tokenPath() string {
	a.base.mu.RLock()
	defer a.base.mu.RUnlock()
	return a.base.canonicalTokenPath
}

func (a *delegatedHTTP) Do(request *http.Request) (*http.Response, error) {
	if request != nil && request.URL != nil {
		if target, ok := a.federation[request.URL.String()]; ok {
			return a.sendFederation(request, target)
		}
	}
	if request == nil || !commercialAuthorityURL(request.URL) || request.Host != "" && request.Host != "login.microsoftonline.com" {
		return rejectAuthorityRequest(request)
	}
	u := request.URL
	tokenPath := a.tokenPath()
	isToken := tokenPath != "" && request.Method == http.MethodPost && u.Path == tokenPath
	if isToken {
		if err := a.beforePoll(request.Context()); err != nil {
			if request.Body != nil {
				_ = request.Body.Close()
			}
			return nil, err
		}
	}
	var response *http.Response
	var err error
	devicePath := strings.TrimSuffix(tokenPath, "token") + "devicecode"
	userRealm := a.username != "" && request.Method == http.MethodGet && u.Path == "/common/UserRealm/"+a.username && u.RawQuery == "api-version=1.0"
	device := tokenPath != "" && request.Method == http.MethodPost && u.Path == devicePath && u.RawQuery == "" && !u.ForceQuery
	if userRealm || device {
		response, err = a.base.send(request, u.Path, false)
	} else {
		response, err = a.base.Do(request)
	}
	if err != nil {
		// A failed silent HTTP exchange is not a cache miss and must not launch
		// an interactive flow. Retain only a safe failure classification.
		a.mu.Lock()
		a.lastFailure = &DelegatedError{}
		a.mu.Unlock()
		return nil, err
	}
	content, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if readErr != nil || closeErr != nil {
		return nil, ErrDelegatedAuthentication
	}
	if userRealm && response.StatusCode >= 200 && response.StatusCode < 300 {
		var realm struct {
			AccountType string `json:"account_type"`
			MetadataURL string `json:"federation_metadata_url"`
		}
		if json.Unmarshal(content, &realm) != nil {
			return nil, ErrDelegatedAuthentication
		}
		if realm.AccountType == "Federated" && a.federation[realm.MetadataURL] == nil {
			return nil, ErrFederatedPassword
		}
	}
	if response.StatusCode >= 400 {
		failure := delegatedFailure(content)
		a.mu.Lock()
		a.lastFailure = failure
		a.mu.Unlock()
		if isToken {
			var payload struct {
				Error string `json:"error"`
			}
			if json.Unmarshal(content, &payload) == nil && payload.Error == "slow_down" {
				a.mu.Lock()
				if a.polling {
					a.interval += 5 * time.Second
					a.nextPoll = a.now().Add(a.interval)
				}
				a.mu.Unlock()
			}
		}
	}
	response.Body = io.NopCloser(bytes.NewReader(content))
	return response, nil
}

func (a *delegatedHTTP) startPolling(seconds int) {
	if seconds <= 0 {
		seconds = 5
	}
	// Device codes expire long before a day; bounding conversion prevents overflow
	// from malformed service metadata while retaining cancellation during the wait.
	if seconds > 86400 {
		seconds = 86400
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.polling = true
	a.interval = time.Duration(seconds) * time.Second
	a.nextPoll = a.now().Add(a.interval)
}
func (a *delegatedHTTP) stopPolling() {
	a.mu.Lock()
	a.polling = false
	a.nextPoll = time.Time{}
	a.mu.Unlock()
}
func (a *delegatedHTTP) beforePoll(ctx context.Context) error {
	a.mu.Lock()
	polling, next := a.polling, a.nextPoll
	a.mu.Unlock()
	if !polling {
		return nil
	}
	if err := a.wait(ctx, next.Sub(a.now())); err != nil {
		return err
	}
	a.mu.Lock()
	a.nextPoll = a.now().Add(a.interval)
	a.mu.Unlock()
	return nil
}
func waitForLogin(ctx context.Context, duration time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if duration <= 0 {
		return nil
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func validDeviceVerificationURL(value string) bool {
	return value == "https://microsoft.com/devicelogin" || value == "https://www.microsoft.com/devicelogin" || value == "https://login.microsoftonline.com/common/oauth2/deviceauth"
}

func (a *delegatedHTTP) validAuthorizationURL(u *url.URL) bool {
	path := a.tokenPath()
	return commercialAuthorityURL(u) && path != "" && u.Path == strings.TrimSuffix(path, "token")+"authorize"
}
