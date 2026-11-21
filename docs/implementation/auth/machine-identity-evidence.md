# AUTH-04 machine identity evidence

Status: implemented and offline qualified. A manual GitHub federation lab workflow is
prepared; its external trust/identifier setup awaits explicit approval after automatic
approval review rejected those persistent security changes. No managed-identity hosting
environment is available for live qualification. Existing application secret/certificate
live results are recorded separately and do not prove federation or managed identity.

## Source contracts

The pinned dependencies are Azure Identity `v1.15.0`, Azure Core `v1.23.3` and MSAL Go
`v1.10.1`. Inspected their managed identity selection, HTTP pipeline, platform request
builders, token cache and confidential-client assertion callback implementations.

- [Azure managed identity credential](https://github.com/Azure/azure-sdk-for-go/blob/sdk/azidentity/v1.15.0/sdk/azidentity/managed_identity_credential.go): explicit identity selection and `/.default` to resource conversion.
- [Azure managed identity pipeline](https://github.com/Azure/azure-sdk-for-go/blob/sdk/azidentity/v1.15.0/sdk/azidentity/managed_identity_client.go): HTTP injection, retry policy, Service Fabric client augmentation and claim forwarding.
- [MSAL managed identity](https://github.com/AzureAD/microsoft-authentication-library-for-go/blob/v1.10.1/apps/managedidentity/managedidentity.go): source discovery, cache, Arc file validation and identity echo, endpoint request builders.
- [MSAL confidential client](https://github.com/AzureAD/microsoft-authentication-library-for-go/blob/v1.10.1/apps/confidential/confidential.go): assertion credential callback and client-credentials exchange.
- [Entra federated identity credentials](https://learn.microsoft.com/en-us/graph/api/resources/federatedidentitycredential): issuer, subject and audience matching.
- [GitHub OIDC reference](https://docs.github.com/en/actions/reference/security/oidc): runtime assertion acquisition, audience and repository/ref subject.

These support the explicit machine identity implementation. They do not alter the
MicrosoftTeams 8.1.0 command baseline or qualify undocumented Teams management routes.

## Managed identity boundary

`NewManagedIdentityProvider` is owned by the SDK session for `ManagedIdentity`. Constructor
validation/snapshotting performs no HTTP calls, callbacks, filesystem reads or Azure Arc
discovery. On acquisition, Azure Identity selects the environment-specific endpoint and
supports a user-assigned client ID where available. Empty `ClientID` selects the host's
default identity, including Azure ML's platform-selected default. Other selector kinds
(object/resource ID) are not currently exposed by this SDK configuration.

The environment snapshot includes `IDENTITY_ENDPOINT`, `IDENTITY_HEADER`,
`IDENTITY_SERVER_THUMBPRINT`, `MSI_ENDPOINT`, `MSI_SECRET`, `IMDS_ENDPOINT`,
`DEFAULT_IDENTITY_CLIENT_ID` and Windows `ProgramData` (Arc installation/secret directory).
Both cache hits and new exchanges reject changes. The first discovered platform source
is pinned. An Azure pipeline guard validates endpoint, method, resource and identity
selection before transport, rejects redirects and bounds response buffers to 1 MiB.
Echoed resource/client IDs are checked when present. The source remains the authority
for actual identity/tenant; opaque tokens are not decoded as proof. Tenant configuration
must identify the host identity's home tenant, and never causes a cross-tenant switch.

Service Fabric retains the library's certificate pinning through the configured HTTP
client; custom unsupported transports fail closed. Arc challenge file restrictions are
owned by the pinned library. The SDK neither broadens those directories nor captures
file secrets. Endpoint failures are redacted to unavailable/rejected sentinels. Acquisition
is bounded to 30 seconds, caller/session cancellation wins, and hidden Azure/MSAL retries
are disabled. The provider never falls back to another credential family.

### Pinned library cache behavior

MSAL's managed identity cache is process-global and omits the hosting endpoint and SDK
tenant binding. It cannot serve as this SDK's isolation boundary. The SDK therefore
uses its own `RefreshProvider` for binding-aware caching/coalescing, and supplies an empty
JSON claims object on each underlying Azure acquisition. In the pinned implementation,
nonempty claims bypass the library cache lookup and are not forwarded to the local
managed identity protocol. No additional claims are requested. Tests require a real
endpoint exchange for each fresh SDK acquisition, even after another session populated
the library cache. Dependency upgrades must preserve or replace this verified behavior.

MSAL still writes its library-owned process cache; the SDK does not read it or clear
another consumer's entries. Close cancels work and drops the SDK's own credentials/cache
references, not every copy held inside third-party runtime state. The HTTP transport
remains caller-owned when injected.

## Workload federation

`ApplicationProvider` now also accepts `WorkloadIdentity`, preserving the existing
commercial tenant/authority guard and fresh MSAL confidential client per exchange.
`AssertionFile` and `Assertion` are mutually exclusive. Files are re-opened and bounded;
known nonregular sources are rejected before opening. The caller must control the source
path and containing directory. Callback raw output is bounded before whitespace trimming.
Read/callback failures discard sensitive diagnostics. Entra, not local JWT decoding,
validates issuer trust and assertion expiry. Cached access tokens do not force needless
assertion reads; an expired access token or a different audience requires a fresh source.
The application provider's optional rejected-token replacement capability also applies.

## Offline verification map

| Tests | Evidence |
|---|---|
| `TestUnitWorkloadRotatedFileAndAudienceIsolation` | File rotation, cached token reuse, expiry refresh and separate Graph/Teams exchanges |
| `TestUnitWorkloadCallbackCoalescesAndRespectsBinding` | Concurrent callback coalescing and tenant rejection before source access |
| `TestUnitWorkloadInvalidAndExpiredAssertions` | Missing/empty/oversized sources, callback errors and issuer rejection of expired assertions |
| `TestUnitWorkloadCancellationAndOwnedClose` | Context propagation, close and owned callback release |
| `TestUnitManagedIdentityPlatformSelection` | IMDS system/user identities, App Service, Azure ML and Cloud Shell request formats |
| `TestUnitManagedIdentityDoesNotReadAnotherSessionsGlobalCache` | Cross-session cache isolation, concurrent coalescing and forced expiry refresh |
| `TestUnitManagedIdentityEnvironmentChangeRejectsCachedAndFreshTokens` | Endpoint/secret/ProgramData changes fail before cache or HTTP use |
| `TestUnitManagedIdentityUnavailableAndRejectedResponses` | Absent/rejected endpoints, wrong identity/resource echoes, expiry, response bounds and redirects |
| `TestUnitManagedIdentityCancellationAndClose` | Endpoint cancellation and owned environment release |
| `TestUnitManagedIdentityServiceFabricKeepsCertificatePinning` | Real local TLS endpoint accepts the configured pin and rejects a wrong pin despite a prior cached token |
| `TestUnitManagedIdentityArcRejectsUntrustedSecretFileChallenge` | Untrusted Arc secret-file challenge cannot trigger a credential-bearing second request |
| `TestUnitManagedIdentityGuardRejectsChangedDestinationAndIdentity` | Destination, path, resource, identity and configured URL validation |

Examples compile with the SDK on Linux/macOS/Windows. The manual live workflow runs
file/callback assertions against both token audiences and performs only an organization
read. It requests OIDC only for `main`, stores assertions in runner temporary storage,
removes them after the test, and does not upload them as artifacts. Enabling its Entra
trust and masked tenant/client identifiers is distinct from merging the implementation.
No resource ledger entries are promoted by this authentication foundation alone.
