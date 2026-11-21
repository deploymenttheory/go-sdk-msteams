# Portable authentication and operation access

Select one `auth.Config.Method` when constructing the client. Each provider is lazy
and acquires Graph and Teams tokens separately under the operation's context.
There is no automatic fallback between users and applications.

| Method | Mode | Portable inputs and limits | Tests |
|---|---|---|---|
| `Browser` | Delegated | Caller opens browser; SDK owns loopback callback with state/PKCE; no Windows broker dependency | [tests](../../msteams/auth) |
| `DeviceCode` | Delegated | Caller displays code/verification URI; context cancels polling | [tests](../../msteams/auth) |
| `Password` | Delegated | Explicit username/password; MFA and conditional access can reject this flow; federated endpoints require explicit trust | [tests](../../msteams/auth) |
| `Certificate` | Application | PEM certificate/private key or supported PFX subset; no OS certificate store needed | [tests](../../msteams/auth) |
| `ClientSecret` | Application | Explicit secret; native provider corresponds to the module's externally acquired app-token path | [tests](../../msteams/auth) |
| `ManagedIdentity` | Application | Azure host supplies system-assigned or selected user-assigned identity; commercial profile | [tests](../../msteams/auth) |
| `WorkloadIdentity` | Application | Assertion file or callback; Entra federation trust required; SDK addition rather than a claimed Connect parameter | [tests](../../msteams/auth) |
| `SuppliedTokens` | Explicit single mode | Audience-keyed values and expiries; caller renews them; both audiences must belong to the same declared identity/mode | [Configuration](client-configuration.md) |
| `ExternalProvider` | Explicit single mode | Borrowed provider or Azure credential adapter; caller owns credential lifecycle | [Configuration](client-configuration.md) |

All SDK-native flows work without PowerShell. Host availability, tenant policies,
consent and credential format support still apply. Offline contracts run on Linux,
macOS and Windows; that is separate from live flow qualification. Secret and PEM
certificate flows have lab evidence. Delegated and Azure managed identity live
qualification remain outstanding; GitHub federation setup awaits explicit approval.

Successful token acquisition does not establish operation authorization. Use the
offline `msteams/capabilities` catalog for dated documentation observations:

```go
catalog, err := capabilities.NewCatalog()
if err != nil {
    return err
}
report, err := catalog.Assess("New-Team", auth.Application, "commercial")
if err != nil {
    return err
}
// report.Support is ModeDocumentedUnsupported. Other commands may be
// ModeUnqualified: unknown never means authorized or implemented.
```

At the operation boundary, use `errors.Is` with `auth.ErrUnauthenticated` and
`auth.ErrForbidden` for HTTP 401 and 403. A 403 cannot identify a missing role by
itself. `ErrUnsupportedOperationMode` and `ErrTenantCapabilityUnavailable` are
reserved for resource-specific evidence; generic response text does not trigger
them. Keep `errors.As` access to `*client.APIError` for status and diagnostic codes.
`ServiceErrorCodes` retains numeric PolicyRP `errorCode` values independently of
string `Codes`. The compiled PolicyRP adapter recognizes HTTP 400 / `ClientError`
/ 40003 together with the explicit type-flighting message as
`ErrTenantCapabilityUnavailable`, preserving the original API error. Generic 400s,
403s and arbitrary unsupported-feature text retain their original classifications.

The capability catalog keeps source dates and documented permission boundaries.
Resource tickets track the authorization and acceptance tests still required.


## Environments and local sessions

Set Config.Cloud to commercial (default), gcc, gcch, dod, china, france or germany.
Pinned Teams environment names are accepted too. Config.Endpoints can explicitly
override HTTPS origins and the Teams token audience; Cloud custom requires all four
fields. Graph's selected origin is also its token audience. Treat overrides as trusted
configuration and keep credentials matched to the selected authority.

Client.Environment and Client.SessionInfo return copies without network access.
Close clears only this client's owned state; changing environment means constructing
a new client. WithAuthConfig replaces environment-loaded defaults. No machine-global
PowerShell settings or internal support-region headers are reproduced.

Sovereign profiles have unit coverage; live qualification is still required.
Managed identity remains commercial-only. The capability catalog records dated
commercial/GCC/GCCH/DoD documentation observations.
