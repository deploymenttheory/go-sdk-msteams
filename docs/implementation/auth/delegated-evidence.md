# AUTH-03 delegated authentication evidence

Status: AUTH-03 implemented with offline qualification. The Azure credential adapter and
built-in delegated providers are wired into SDK sessions. No delegated tenant sign-in has
been performed; full Connect-MicrosoftTeams coverage remains with AUTH-06 and lab work.

## Entra and Graph Go SDK references

Following the user's direction, inspected the dependencies already pinned in this SDK:
Azure Identity `v1.15.0`, Azure Core `v1.23.3` and MSAL Go `v1.10.1`. Also inspected
Microsoft Kiota's Azure authentication provider at immutable revision
`204916181aec7a7f559161d9a48f0e57e6312591`.

| Reference | Adopted pattern and Teams-specific requirement |
|---|---|
| [Kiota Azure token provider](https://github.com/microsoft/kiota-authentication-azure-go/blob/204916181aec7a7f559161d9a48f0e57e6312591/azure_identity_access_token_provider.go) | Borrow `azcore.TokenCredential`, pass context and explicit scopes, validate destination before authentication. Teams needs the `48ac35b8-9aa8-4d74-927d-1f4a14a0b239/.default` scope; deriving it from the API hostname would be wrong. |
| [Azure Identity public client](https://github.com/Azure/azure-sdk-for-go/blob/sdk/azidentity/v1.15.0/sdk/azidentity/public_client.go) | Silent acquisition uses an authenticated account; interactive acquisition is a distinct operation. Keep account selection consistent across the two audiences and retain cancellation while waiting for shared acquisition. |
| [Azure Identity device credential](https://github.com/Azure/azure-sdk-for-go/blob/sdk/azidentity/v1.15.0/sdk/azidentity/device_code_credential.go) | A context-aware prompt callback receives user instructions. Our built-in SDK flow must require explicit presentation behavior and avoid printing instructions automatically. |
| [Azure Identity browser credential](https://github.com/Azure/azure-sdk-for-go/blob/sdk/azidentity/v1.15.0/sdk/azidentity/interactive_browser_credential.go) | Tenant/client options, explicit authentication, local account record and an in-memory cache are useful patterns. Its public options lack a custom browser-opening callback; MSAL exposes one for the SDK's caller-controlled browser flow. |
| [MSAL public client](https://github.com/AzureAD/microsoft-authentication-library-for-go/blob/v1.10.1/apps/public/public.go) | Use Microsoft's browser PKCE, device-code and password implementations; retain caller context and avoid an SDK credential fallback chain. |

`NewAzureTokenProvider(binding, credential)` now bridges existing Azure Identity or other
`azcore.TokenCredential` implementations to the SDK's provider contract. It binds tenant,
cloud, identity and mode before calling the borrowed credential, accepts only the two
commercial audiences, and maps each to its own `/.default` scope. It delegates caching
and lifetime to the credential. It does not instantiate `DefaultAzureCredential`, select
a browser fallback, inspect an opaque token, or silently enable additional tenants.

The caller remains responsible for the borrowed credential's identity and authentication
mode. The existing bound-provider wrapper validates returned expiry/material and redacts
ordinary errors; explicitly unwrapped third-party errors can contain sensitive diagnostics.
Tests cover constructor purity, tenant/scope mapping, binding rejection before acquisition,
typed-nil credentials, cancellation, expiry and error formatting. Azure Core is promoted
from an indirect dependency to a direct dependency without a version change.

## PowerShell and protocol requirements

The [Connect-MicrosoftTeams documentation](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/connect-microsoftteams?view=teams-ps)
describes interactive, device and credential connections and requires separate Graph and
Teams resource tokens. Its delegated example uses each resource's `/.default` scope.
The checked 8.1.0 runtime remains authoritative for baseline parameter sets; current
documentation is supporting protocol evidence, not a reason to change the baseline.

[ROPC documentation](https://learn.microsoft.com/en-us/entra/identity-platform/v2-oauth-ropc)
establishes that password authentication cannot satisfy MFA and has account/federation
restrictions. A denied password flow must return an actionable error, never automatically
switch to a different authentication mode. Browser/device alternatives are explicit.
Portable browser authentication replaces Windows broker assumptions; certificate files
already replace platform-specific certificate-store lookup in AUTH-02.

## Delivered delegated provider

`NewDelegatedProvider` uses MSAL Go's public client for account/cache management, silent
refresh, device-code acquisition, password acquisition and authorization-code redemption.
The root session owns this provider for `Browser`, `DeviceCode` and `Password`. All
constructors remain free of network, browser, callback and credential-store activity.

`Config.OpenBrowser(ctx, URL)` and `Config.DevicePrompt(ctx, DeviceCodeMessage)` are explicit
required callbacks for their respective methods. Callbacks must cooperate with context.
`LoginTimeout` defaults to 15 minutes; caller cancellation and session closure also cancel
acquisition. A context-aware gate serializes account selection across resource audiences.
AUTH-01 coalesces requests per audience; cancelling one waiter does not abort other waiters.
The selected MSAL account is pinned for both audiences. Silent refresh responses without
an ID token retain the account used for that refresh. A different returned account fails.
Silent transport failures return errors without launching a fresh sign-in prompt.

The SDK loopback browser receiver uses MSAL to build the authorization URL and exchange
the code, adding random state and S256 PKCE. It validates state on both success and error
callbacks, limits POST bodies, does not reflect untrusted values, and forcibly closes
active connections on cancellation. This ownership is deliberate: the pinned MSAL
`AcquireTokenInteractive` receiver calls `Shutdown(context.Background())`, which can wait
on an active slow POST. The SDK's tests include that stalled-body scenario. It invokes
only the caller's browser callback after validating the authorization authority/path.
Register the portable `http://localhost` redirect URI on the public client application.

Device instructions expose only the user code, verification URL, message and expiry;
the private device code stays inside MSAL. Verification URLs are explicitly checked.
The adapter enforces the service poll interval and increases it by five seconds on
`slow_down`, compensating for the pinned library's internal 50 ms to five-second poller.
Callbacks and polling stop on expiry/cancellation. No fallback to another login method
occurs. Error bodies are captured before MSAL consumes them; only recognized OAuth error
values and numeric Entra diagnostic codes survive. The [Entra error guidance](https://learn.microsoft.com/en-us/entra/identity-platform/reference-error-codes)
identifies OAuth values as the control-flow contract and numeric codes as diagnostics.
MFA/conditional-access failures preserve those codes without retaining descriptions.

Password authentication supports managed accounts and the library's WS-Trust 1.3 federated
exchange. Federation requires exact HTTPS MEX and password URLs in
`Config.FederationEndpoints`, copied at construction. Metadata cannot expand this list,
redirects and cookies are disabled, and response buffers are bounded. Without explicit
endpoint trust, federation fails before a password is forwarded. The synthetic MEX/SAML
fixture follows the pinned MSAL XML models and verifies the complete SOAP-to-Entra grant.
MSAL does not implement WS-Trust 2005; this is an explicit library limitation, not a claim
of qualified support for every tenant federation deployment.

Owned Close cancels workers and drops configuration/account/cache references. Wait allows
callers to await cooperative callback exit. Borrowed Azure credentials retain their own
lifecycle. Standard formatted errors never include password, callback URL, device code,
server description or token. Explicit callback data is for the signing-in user's UI.

## Verification map

| Tests | Contract checked |
|---|---|
| `TestAzureCredentialAdapter*`, `ExampleNewAzureTokenProvider` | Graph-compatible borrowed TokenCredential, exact tenant/scopes, cancellation and lifecycle boundary |
| `TestSessionOwnsDelegatedProvidersWithoutStartingLogin` | Root/session wiring, constructor purity and owned close for all three methods |
| `TestDelegatedPasswordAndSilentSecondAudience` | Password first grant; second audience silently uses the selected account |
| `TestDelegatedOAuthFailuresPreserveDiagnosticsWithoutFallback` | MFA, conditional access, consent, denial and expiry fixtures preserve codes without retry/downgrade |
| `TestDelegatedDeviceFlowAndCallback`, `TestDevicePollingHonorsIntervalAndSlowDown` | Prompt/poll behavior, pending state, interval and slow-down |
| `TestDelegatedDeviceRejections`, `TestDeviceCodeExpiresWhilePromptWaits`, `TestDevicePromptCanCancelWithoutLeakingCallbackError` | Verification authority, expiry and explicit callback cancellation |
| `TestBrowserPKCEAndConcurrentAudiencesShareAccount` | Real local callback, state/PKCE redemption and concurrent audience reuse |
| `TestBrowserDenialCallbackErrorAndUntrustedAuthority`, `TestBrowserCallbackRequiresStateBeforeErrorOrCode` | Denial, callback failure, hostile authority, state/host/body checks |
| `TestBrowserCancellationClosesSlowCallbackConnection` | Cancellation closes an active incomplete form POST and worker exits |
| `TestDelegatedAccountCannotChangeBetweenAudiences`, `TestDelegatedSilentRefreshWithoutIDTokenKeepsSelectedAccount` | Account consistency and normal ID-token omission |
| `TestDelegatedSilentTransportFailureDoesNotPrompt`, `TestDelegatedCancellationWhileWaitingAndClose` | No new login on transport failure; cancelled waiters/close release work |
| `TestTrustedFederationPasswordUsesMicrosoftWSExchange`, `TestFederationEndpointValidation`, `TestDelegatedFederationDoesNotForwardPassword` | MSAL federation interoperability and endpoint trust |

Fixtures use synthetic unsigned ID-token representations only inside the fake trusted
identity transport, to exercise MSAL account extraction. They are not tenant credentials
or evidence of live authentication. Access tokens remain opaque. No ledger command is
promoted to live-tested or complete by this foundation alone.
