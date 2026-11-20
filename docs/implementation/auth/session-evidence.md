# AUTH-06 session helper investigation

Status: in progress. This records pinned evidence and the remaining implementation
work; it does not establish complete connection/environment parity.

## Pinned observation

The accepted 8.1.0 runtime inventory observes four Connect parameter sets on macOS
PowerShell Core: `AccessToken`, `ManagedServiceLogin`, `ServicePrincipalCertificate`
and `UserCredential`. Disconnect and Set/Clear-TeamsEnvironmentConfig each expose
`__AllParameterSets`. The five assigned OCE/region helpers are conditional exports,
not observed in that default import.

`custom/CmdletConfig.json` restricts all five OCE/region helpers to `ExportsTo: [Torus]`.
This is corroborated by the pinned proxy declarations and the actual method bodies,
inspected offline without executing any configuration mutation or tenant command.

| Assigned helper | Pinned behavior | SDK mapping / outstanding work |
|---|---|---|
| Connect-MicrosoftTeams | Explicit user, certificate, supplied-token and managed identity parameter sets | Root construction plus resource-aware provider acquisition; portable flow mappings already implemented, full environment/parameter mapping still under review |
| Disconnect-MicrosoftTeams | Local connection helper | Root `Close` cancels owned work and clears owned credentials; does not revoke Entra tokens or clear borrowed providers |
| Set-TeamsEnvironmentConfig | Core assembly throws an unsupported-on-PowerShell-Core exception; help describes machine-local environment defaults, primarily for special-cloud migration | Explicit per-client configuration is the portable design; special-cloud/custom-endpoint behavior remains unresolved |
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

## Remaining work

1. Finish the per-parameter mapping, including endpoint/environment precedence,
   caller-selected account identity, portable logging and certificate equivalents.
2. Credential-free `Client.SessionInfo` / `Session.Info` inspection is now implemented.
   `TestSessionInformationAndCloseRemainClientLocal` and `TestSessionInfoIsLazyAndNilSafe`
   pass under the race detector: caller config mutations do not rebind clients,
   concurrent token/metadata reads stay isolated, Close affects only its owner,
   copied info cannot mutate the session, and inspection performs no credential parsing.
   Complete the remaining per-parameter/lifecycle mappings before closing this ticket.
3. Resolve special-cloud/custom-endpoint support explicitly. Current providers and
   transport enable only the commercial profile; unknown cloud labels cannot be
   treated as successful environment configuration or operation capability.
4. Attach exact test and symbol mappings without promoting incomplete Connect or
   environment helpers to full command parity. Internal-context disposition must
   retain its pinned evidence and remain distinct from a new exclusion.

## Environment evidence collected for implementation

The pinned `DeploymentConfiguration` dictionaries map TeamsCloud to Public,
TeamsGCCH to GCCH, TeamsDOD to DoD, TeamsChina to Gallatin, TeamsFrance to Bleu and
TeamsGermany to Delos. These are explicit module mappings, not runtime SDK support.

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
cross-host redirects or mixing token audiences. France/Germany authority/Graph tuples
and custom EndpointUris semantics still need evidence before enabling those routes.

The current [environment-helper documentation](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/set-teamsenvironmentconfig?view=teams-ps)
retains saved-default precedence and the special-cloud migration purpose. It is
current documentation, separate from the pinned Core methods' observed exception.
