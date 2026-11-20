# AUTH-03 delegated authentication evidence

Status: implementation in progress. HTTP-01 is merged. The Azure credential adapter is
implemented and tested locally; built-in delegated providers and their qualification are
still required before closing AUTH-03. No delegated tenant sign-in has been performed.

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

## Remaining implementation checks

- Add built-in provider wiring for browser, device and password modes without network
  activity in constructors. Use explicit callback configuration and bounded operation
  contexts, and clear provider-owned account/cache references on close.
- Pin the selected delegated account for both audiences. Serialize account selection with
  a context-aware gate; an ordinary mutex around an interactive wait can delay cancelled
  callers. Separate tenants/clients retain independent sessions.
- Extend the authority guard using actual MSAL request paths. Browser authorization URLs
  need validation before invoking a caller's browser callback. Password realm discovery
  can branch to federation; never forward passwords to an unreviewed metadata endpoint.
- Exercise the pinned MSAL device poller, which currently starts at 50 ms and caps its
  internal delay at five seconds. Verify/enforce the service-provided interval and
  `slow_down` handling at the adapter boundary. The [device protocol](https://learn.microsoft.com/en-us/entra/identity-platform/v2-oauth2-device-code)
  defines the interval, expiry and terminal errors; importing a library does not replace
  those behavior checks.
- Preserve actionable denial/MFA/conditional-access classifications while discarding raw
  OAuth descriptions, device codes and callback URLs from formatted errors. MSAL's device
  error inspection consumes response bodies, so capture safe classification before that
  inspection rather than assuming the body remains readable afterwards.
- Test browser denial/cancellation and PKCE callback handling, device pending/slow-down,
  expiry, password denial, concurrent audience acquisition and session closure. Lab reads
  remain separate from offline qualification; no command coverage state is advanced here.
