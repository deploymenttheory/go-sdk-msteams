# AUTH-06 session helper investigation

Status: implemented session/helper mapping with fixture coverage. Cloud routing
availability is distinct from live qualification and individual command parity.

## Pinned observation

The accepted 8.1.0 runtime inventory observes four Connect parameter sets on macOS
PowerShell Core: `AccessToken`, `ManagedServiceLogin`, `ServicePrincipalCertificate`
and `UserCredential`. Disconnect and Set/Clear-TeamsEnvironmentConfig each expose
`__AllParameterSets`. The five assigned OCE/region helpers are conditional exports,
not observed in that default import.

`custom/CmdletConfig.json` restricts all five OCE/region helpers to `ExportsTo: [Torus]`.
This is corroborated by the pinned proxy declarations and the actual method bodies,
inspected offline without executing any configuration mutation or tenant command.

| Assigned helper | Pinned behavior | SDK mapping / disposition |
|---|---|---|
| Connect-MicrosoftTeams | Explicit user, certificate, supplied-token and managed identity parameter sets | Root construction plus resource-aware provider acquisition; portable flow mappings already implemented, explicit environment tuple and per-parameter mapping below |
| Disconnect-MicrosoftTeams | Local connection helper | Root `Close` cancels owned work and clears owned credentials; does not revoke Entra tokens or clear borrowed providers |
| Set-TeamsEnvironmentConfig | Core assembly throws an unsupported-on-PowerShell-Core exception; help describes machine-local environment defaults, primarily for special-cloud migration | Explicit per-client configuration is the portable design; named profiles plus explicit endpoint overrides below; no machine-global defaults |
| Clear-TeamsEnvironmentConfig | Core assembly likewise throws; help describes clearing those machine-local defaults | Caller discards/replaces its configuration; no machine-global registry/state mutation |
| Set-CsOCEContext | Torus-only; changes static tenant/application/environment and support-header context; may resolve a tenant domain | Internal-context disposition; not a delegated or application authentication shortcut |
| Clear-CsOCEContext | Torus-only; calls static `OCEContext.Clear` | Internal-context disposition; SDK lifecycle remains scoped to one client |
| Set-CsRegionContext | Torus-only; warns to use Set-CsRoutingContext, then assigns static `_Region` | Internal support routing, not an ordinary tenant data-residency selector |
| Get-CsRegionContext | Torus-only; warns to use Get-CsRoutingContext, then reads static `_Region` | Internal-context disposition; any SDK configuration inspection must be per-client and credential-free |
| Clear-CsRegionContext | Torus-only; warns to use Clear-CsRoutingContext, then clears static `_Region` | Internal-context disposition; no process-wide routing cache is reproduced |

No helper is removed from the candidate ledger or silently counted as implemented
by this classification. In particular, a conditional internal context is not evidence
that its underlying management commands are excluded from the SDK's parity scope.

## Method-body findings

The Core `SetTeamsEnvironmentConfig.ProcessRecord` and
`ClearTeamsEnvironmentConfig.ProcessRecord` methods log and throw an unsupported
PowerShell Core exception. Their `ProcessInternal` methods merely delegate to
ProcessRecord. This explains why export discovery alone cannot prove usable behavior.
The pinned help says Connect parameters override saved defaults and describes
EndpointUris and TeamsEnvironmentName, including TeamsChina migration scenarios.
Windows Desktop persistence has not been executed or qualified here.

The ConfigAPI region enum is `Noam`, `Emea`, `Apac`. `SetRegionContext.ProcessRecord`
sets the static nullable region at IL offset 64; clear resets it at offset 61; get
reads it at offset 58. Deprecation warnings point to the routing-context replacements.

`SendAsyncCallback.AddOrUpdateOCERelatedHeaders` first checks `IsTorusEnvironment`.
Its region path uses `X-MS-InternalSupport-TargetRegion` (literal at offset 95),
and its ring path uses `X-MS-InternalSupport-TargetRing` (offset 199). It also merges
OCE context headers. Other consumers include modern forest discovery and OCE
authorization adapters. These observations do not authorize adding support-only
headers or selecting internal deployment rings in ordinary tenant requests.

`SetOCEContext.ProcessRecord` handles tenant domain/ID, clears/repopulates static
context, retains routing overrides when its environment matches, selects a configured
or supplied app ID, stores headers/system-tenant state and sets `_IsOCEContextSet`.
`OCEContext.Clear` clears tenant, application, forest, environment, region, ring,
cloud and header references. This behavior belongs to the module's host context;
it is distinct from acquiring or revoking user/application access tokens.

## Sources

All observations above are from the accepted package whose SHA256 is recorded in
`contracts/powershell/8.1.0/provenance.json`. Relevant assembly hashes:

- Core Connect: `c1c2b2d1c4a5306fb70f373df7ed5bfca4f9db58e90be358afd8a1d48ed46aea`
- Core CmdletHostContract: `477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af`
- ConfigAPI private: `8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6`
- Connect help XML: `1fc87439f1bfdbf53e836458d46d2b4144e0fb65975c5bcaa73ec8040f1ee8ff`

Reconciliation records preserve the conditional export configuration. Proxy
declarations expose Set-CsOCEContext's AppId, TenantDomain, TenantId, Environment,
Headers and IsSystemTenant inputs, and Set-CsRegionContext's mandatory Region.
The runtime artifact retains all observed Connect parameter details.

## Environment evidence collected for implementation

The pinned `DeploymentConfiguration` dictionaries map TeamsCloud to Public,
TeamsGCCH to GCCH, TeamsDOD to DoD, TeamsChina to Gallatin, TeamsFrance to Bleu and
TeamsGermany to Delos. The SDK resolves these names and aliases to explicit endpoint profiles.

| Pinned configuration environment | ConfigAPI host | Authority in pinned CmdletSettings |
|---|---|---|
| Public / MultiTenant | api.interfaces.records.teams.microsoft.com | login.microsoftonline.com |
| GCCH / Itar | api.interfaces.records.gov.teams.microsoft.us | login.microsoftonline.us |
| DoD | api.interfaces.records.dod.teams.microsoft.us | login.microsoftonline.us |
| Gallatin | api.interfaces.records.teams.microsoftonline.cn | login.partner.microsoftonline.cn |
| Bleu | api.interfaces.records.communications.svc.sovcloud.fr | Not present in that settings file |
| Delos | api.interfaces.records.communications.svc.sovcloud.de | Not present in that settings file |

The settings resource ID for the listed Public, GCCH, DoD and Gallatin entries is
`48ac35b8-9aa8-4d74-927d-1f4a14a0b239`. Internal INT/AG08/AG09 entries are not silently
promoted to public SDK environments. The source is
`netcoreapp3.1/CmdletSettings.json` plus the hashed host-contract assembly above.

Current [Graph deployment documentation](https://learn.microsoft.com/en-us/graph/deployments),
reviewed 2026-11-20, distinguishes worldwide, GCC High, DoD and China Graph origins:
`graph.microsoft.com`, `graph.microsoft.us`, `dod-graph.microsoft.us`, and
`microsoftgraph.chinacloudapi.cn`. GCC uses worldwide endpoints. China documentation
uses `login.chinacloudapi.cn`, which differs from the pinned settings authority alias;
the implementation must resolve that distinction explicitly without trusting arbitrary
cross-host redirects or mixing token audiences. The SDK uses the documented China authority by default; callers needing the pinned
legacy alias can explicitly override Authority. It never follows cross-host aliases.

The current [environment-helper documentation](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/set-teamsenvironmentconfig?view=teams-ps)
retains saved-default precedence and the special-cloud migration purpose. It is
current documentation, separate from the pinned Core methods' observed exception.


The pinned Connect assembly SHA256 is
`3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24`.
`Models.AzureEnvironment..cctor` contains Germany authority/Graph literals at IL
431/455 and France at 497/521; `Models.TeamsEnvironment.setMsGraphEndPoint`
selects France/Germany Graph at 212/250. The configured tuples are:

| SDK profile | Authority | Graph |
|---|---|---|
| commercial / gcc | login.microsoftonline.com | graph.microsoft.com |
| gcch | login.microsoftonline.us | graph.microsoft.us |
| dod | login.microsoftonline.us | dod-graph.microsoft.us |
| china | login.chinacloudapi.cn | microsoftgraph.chinacloudapi.cn |
| france | login.sovcloud-identity.fr | graph.svc.sovcloud.fr |
| germany | login.sovcloud-identity.de | graph.svc.sovcloud.de |

The pinned `DeploymentConfiguration.ConfigApiEnvironmentToAppId` dictionary
confirms resource `48ac35b8-9aa8-4d74-927d-1f4a14a0b239` for Public, GCCH, DoD,
Gallatin, Bleu and Delos. This is the resource audience; the distinct
`EnvironmentToAppId` dictionary holds module client IDs and is not reused by the SDK.

## Portable parameter mapping

| Connect input / helper behavior | SDK equivalent |
|---|---|
| TenantId | Config.TenantID; immutable tenant binding; GUID or DNS domain for native exchanges |
| TeamsEnvironmentName | Config.Cloud; pinned Teams names plus commercial/gcc/gcch/dod/china/france/germany labels |
| EndpointUris | Config.Endpoints: typed Authority, Graph, TeamsAdmin origins and TeamsAdminResource audience; explicit nonempty fields override the selected profile |
| ApplicationId / Certificate | Caller-owned ClientID and Certificate method using PEM pair or supported PFX; OS thumbprint/store lookup is replaced by explicit portable material |
| AccessTokens | Audience-keyed tokens with explicit expiry, IdentityID and single TokenMode; Graph origin is its audience, Teams audience is independent of API origin |
| Credential | Explicit Password method and Username/Password; no implicit fallback if MFA/conditional access rejects it |
| UseDeviceAuthentication | DeviceCode method and context-aware DevicePrompt callback |
| Default interactive / DisableWAM | Browser method, caller OpenBrowser callback and SDK state/PKCE loopback; no WAM or OS broker dependency |
| AccountId (user) | Browser LoginHint is a hint, not verified identity; MSAL account locks subsequent delegated acquisition; password uses Username |
| Identity / AccountId (managed service) | ManagedIdentity; empty ClientID selects system identity, explicit ClientID selects user-assigned identity |
| ManagedServiceHostName/Port/Secret | Supported Azure host environment contracts are snapshotted by the managed identity provider; arbitrary host/secret overrides are deliberately not exposed |
| LogFilePath / LogLevel | Caller-owned Zap logger through WithLogger; file destination and level belong to caller; detailed request instrumentation is HTTP-06 |
| Verbose / Debug / InformationAction / InformationVariable | Caller logger and explicit Go return values; no PowerShell stream variables |
| ErrorAction / ErrorVariable / WarningAction / WarningVariable | Returned errors, errors.Is/As, context cancellation and caller logging; no shell-global action preferences |
| OutVariable / OutBuffer / PipelineVariable / ProgressAction | Typed return values and context-aware polling; no shell pipeline buffers or progress preferences |
| Disconnect / Clear saved environment | Close releases only owned session state; create a new client with replacement config to change settings; no Entra revocation or machine registry mutation |

Set-TeamsEnvironmentConfig persistence is replaced by caller configuration storage.
NewClient uses only explicit config; NewClientFromEnv reads MSTEAMS variables before
applying explicit options, and WithAuthConfig replaces that entire configuration.
There is no mutable global environment. Cloud and endpoint overrides are copied;
all four resolved endpoint fields contribute to override cache isolation.
`Client.SessionInfo` and `Client.Environment` return local copies without credential
parsing, discovery or token acquisition. Ordinary formatted diagnostics redact them.

The five Torus helpers have a reviewed internal-host-context disposition based on
the export restriction and method/consumer evidence above. There is no public Region
field that would pretend Noam/Emea/Apac support headers select tenant residency.
Supplying those strings as Cloud fails local validation. Tenant/application settings
belong to Config, clear belongs to Close, and inspection belongs to the local metadata
methods. No internal support headers or privileged host contexts are synthesized.
This disposition does not change candidate counts or exempt their resource services.

## Implementation and qualification

`auth.ResolveEnvironment` validates HTTPS origins without paths, query strings or
credentials. `custom` requires a complete tuple. Overrides are trusted caller input;
never populate them from API responses. Native MSAL providers disable instance-alias
and automatic regional discovery. OIDC issuer and token endpoints remain bound to
the selected authority and tenant. Browser authorization and device verification URLs
are checked separately; device codes accept the known microsoft.com device-login URL
or the selected authority's exact `/common/oauth2/deviceauth` endpoint. Additional
sovereign short-link formats require evidence before acceptance.

The Azure TokenCredential adapter supports the same profiles and explicit overrides
through NewAzureTokenProviderWithEndpoints. The caller configures the borrowed
credential's authority independently. Managed identity remains commercial-only under
its existing documented platform contract. The dated capability catalog still covers
commercial/GCC/GCCH/DoD documentation only; enabling another profile does not invent
capability evidence. No sovereign tenant, delegated login or Azure host was exercised.

| Verification | Evidence |
|---|---|
| Concurrent tenant/identity isolation, copied metadata, owner-only Close, lazy inspection | TestSessionInformationAndCloseRemainClientLocal; TestSessionInfoIsLazyAndNilSafe |
| Six native flows against every profile and custom origins/ports; distinct Graph/Teams scopes; certificate assertion authority; no instance-alias discovery | TestEnvironmentAuthenticationFlows |
| Cross-authority issuer/token/browser metadata rejected before credential POST | TestEnvironmentRejectsForeignIdentityMetadataBeforeCredentialPOST |
| Explicit endpoint validation, full-tuple cache isolation, equivalent origin canonicalization, internal region rejection, redaction | TestEnvironmentOverridesValidateAndIsolateBindings |
| Backend routes, hostile continuation rejection before auth, custom Graph version guard, relative operation locations, immutable copied config | TestEnvironmentRoutesAndContinuationsStayBound |
| Borrowed Azure adapter scope and custom authority binding | TestAzureCredentialCustomEnvironmentBinding |
| Configuration precedence and session lifecycle regression | Existing msteams/client and msteams constructor/session tests |
| Commercial app-token and operation access regression | Opt-in TestApplicationAccessLab; live result recorded in PR evidence |

These are foundation contracts. The coverage ledger remains unchanged: this ticket
does not promote the whole Connect surface, conditional helpers or resource commands
to release-ready parity. Windows machine-default persistence is intentionally replaced
by explicit Go configuration; it has not been executed as a PowerShell oracle here.
