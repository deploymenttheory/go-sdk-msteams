# CORE-01 targeting-policy eligibility and ConfigAPI routing evidence

Research revision: 2026-11-20; pinned MicrosoftTeams **8.1.0**. This supplements
[template/policy/deployment contracts](template-policy-deployment-evidence.md),
[AUTH-06 session evidence](../auth/session-evidence.md) and
[HTTP-01 routing evidence](../http/routing-evidence.md). It traces the three
`Get/Set/Remove-CsTeamsTargetingPolicy` commands and the ConfigAPI routing shared
by templates and batch deployment. No credentials, live tenant calls, sign-in,
management operations or remote sessions were used.

**The native modern path has an evidenced HTTP discovery/bootstrap contract and
does not require creating a TRPS session. Universal REST acceptance for tenants
that the pinned client does not select for modern execution remains unproven.**
The flight list is a server-provided client dispatch setting, not an evidenced
service authorization prerequisite. Neither assuming universal acceptance nor
declaring every absent flight an unsupported tenant is justified by this source.
All three commands remain in parity scope; the unresolved row below blocks a
claim of unconditional tenant parity.

## Sources and reproduction

All package references are to [MicrosoftTeams 8.1.0 on PowerShell Gallery](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0),
package SHA256
`78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`.
Assembly type and method names identify the primary evidence independently of
decompiler formatting.

| ID | Package path | SHA256 |
| --- | --- | --- |
| P | `bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll` | `8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6` |
| H | `netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll` | `477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af` |
| K | `netcoreapp3.1/Microsoft.TeamsCmdlets.PowerShell.Connect.dll` | `c1c2b2d1c4a5306fb70f373df7ed5bfca4f9db58e90be358afd8a1d48ed46aea` |
| F | `custom/CmdletConfig.json` | `754d149c1a5904158e9f01eb55ba44d15c17b7e66f7556385ea894f4ff84d906` |
| C | `custom/Merged_custom_PsExt.ps1` | `03dbc3a85e12c98dd1807d604782c6a90d0656569541cd9540738825a7fcb255` |
| I | `internal/Merged_internal.ps1` | `6ff95158174bab031dc1fc24112a2bc6ae81cf1e2f59db4ecb58afbefc8c2bd9` |

Method bodies were inspected offline with ILSpy CLI **11.2.0.10050**, for example:

```sh
env DOTNET_ROOT=/opt/homebrew/opt/dotnet/libexec \
  DOTNET_CLI_HOME=/private/tmp/msteams-dotnet-home \
  /private/tmp/msteams-dotnet-tools/ilspycmd \
  -t Microsoft.Teams.ConfigApi.Cmdlets.SessionConfiguration \
  /private/tmp/msteams-baseline-810/module/bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll
```

The same procedure was used for the types cited below, with H/K substituted for
their owning assemblies. Temporary decompilations are outside the repository.
`shasum -a 256` checked the original files. An offline reflection read of P's
manifest resource `Microsoft.Teams.ConfigAPI.Cmdlets.custom.CmdletConfig.json`
produced the same SHA256 as F; this matters because
`CmdletConfiguration.InitializeOnceInProcess` reads that embedded resource, not
the adjacent loose JSON file. This read did not import the module or invoke
cmdlets/static initializers.

P types below are in `Microsoft.Teams.ConfigApi.Cmdlets` unless marked
`Generated`, whose namespace is `Microsoft.Teams.ConfigAPI.Cmdlets.Generated`.
H types are in `Microsoft.Teams.ConfigAPI.CmdletHostContract`. K references use
their full class names where needed. Confidence is high for these static client
contracts and unobserved for actual service acceptance.

## Origin and shape of the modern-command selection

The normal module connection path is:

1. K `ConnectMicrosoftTeams.ProcessRecord` establishes identity/context, installs
   `TPMCmdletHostContract`, then calls `WriteConfigApiTelemetry(account.Type)`.
2. `WriteConfigApiTelemetry` invokes `Invoke-CsInternalPSTelemetry` with the
   account-type string. Its exception is logged and swallowed by Connect; a
   successful login therefore does not prove that selection data was obtained.
3. P `CsInternalPsTelemetry.ProcessRecord/InvokeCsInternalTelemetryRelay` invokes
   **`Invoke-CsInternalTelemetryRelay_InternalExpanded`**, passing module
   versions, old connector version (`none` when absent), and auth type.
4. `Generated.ConfigApiBasedCmdlets.TelemetryRelayInternal` sends **POST
   `/Teams.InternalSupport/TelemetryRelay`** through the normal ConfigAPI
   pipeline. It uses UTF-8 JSON and `Content-Type: application/json`. The 200
   response is `TelemetryRelayResponse`; 400 is a raw string, other statuses are
   undeclared responses in this generated method.
5. The generated command's `onOk` emits `response.SessionConfiguration`.
   `CsInternalPsTelemetry` constructs `SessionConfiguration` from it and stores
   it through `SessionStateStore.SetCsSessionConfiguration`. An exception in
   this step sets the stored configuration to null and rethrows.

Sources: K `Microsoft.TeamsCmdlets.PowerShell.Connect.ConnectMicrosoftTeams`;
P `CsInternalPsTelemetry`, `Generated.Cmdlets.InvokeCsInternalTelemetryRelay_InternalExpanded`,
`Generated.ConfigApiBasedCmdlets.TelemetryRelayInternal/_Call`.

The request model has exactly these optional string keys; null strings are
omitted, while empty strings remain:

```json
{
  "MicrosoftTeamsPSVersion": "8.1.0",
  "SfBOnlineConnectorPSVersion": "none",
  "ConfigApiPowershellModuleVersion": "9.1116.2",
  "TeamsModuleAuthTypeUsed": "User"
}
```

This is a **synthetic package-derived example**, not a capture. The pinned
manifests declare MicrosoftTeams 8.1.0 and ConfigAPI 9.1116.2. Actual values are
obtained from module metadata, with a prerelease suffix added when appropriate.
K `AzureAccount.AccountType` also names `ServicePrincipal`, `AccessToken`,
`ManagedService`, `Device` and `IntegratedWindows`; the auth value describes the
module's authentication path, not authorization for a particular policy.
[P `Generated.Models.ConnectPowershellTelemetry.ToJson`; `CsInternalPsTelemetry`;
K `Microsoft.TeamsCmdlets.Powershell.Connect.Models.AzureAccount`; package
`MicrosoftTeams.psd1`, `Microsoft.Teams.ConfigAPI.Cmdlets.psd1`.]

The response shape consumed by the client is:

```json
{
  "SessionConfiguration": {
    "Tpm-ConfigApiModuleConfiguration": {
      "RemotingCmdletsFlightedForAutoRest": "Get-CsTeamsTargetingPolicy,Set-CsTeamsTargetingPolicy,Remove-CsTeamsTargetingPolicy"
    }
  }
}
```

Again this is **synthetic**, showing the parser's expected types. The flight
value is one comma-delimited **string**, not a JSON array or Boolean map.
`SessionConfiguration` splits on comma without trimming and constructs a
case-insensitive `HashSet<string>`. Missing/blank value leaves the set empty;
missing outer configuration key also leaves initialized defaults. The model is
an associative JSON object, and the constructor expects the nested value to be
a dictionary. A malformed nested shape can throw; it is not equivalent to a
well-formed empty flight list. [P `SessionConfiguration` constructor and
`Generated.Models.TelemetryRelayResponse/TelemetryRelayResponseSessionConfiguration`.]

The constructor also parses `RemotingHost`, `ClientAction`/`ClientMessage`,
`CaptureTelemetry`, `CmdletsEnabledForPipelineInputBatching` and
`PipelineInputBatchSize`. `ClientAction` can cause a warning or minimum-version
block in `CsInternalPsTelemetry` and Connect. The source contains **no
server-side selection rules** establishing how tenant, cloud, role, module
version or rollout state determine the returned flight string. It is a
module-version/auth-context bootstrap exchange, not a documented general SDK
capability API. A Go implementation must not silently equate its own version
with a Microsoft PowerShell module version. [Same P methods; K
`ConnectMicrosoftTeams.IsTpmVersionBlocked`.]

P also has `TelemetryRelayAppInternal` at
`/Teams.InternalSupport/TelemetryRelayApp` with the same request/response model.
**That is not the route selected by the inspected Connect path, even for
`ServicePrincipal`.** Its name does not justify switching app-only bootstrap to
it. [P `CsInternalPsTelemetry.LazyInvokeInternalTelemetryRelayCmdletInfo` and
`Generated.ConfigApiBasedCmdlets.TelemetryRelayAppInternal`.]

## The selection algorithm and local overrides

F maps each of the three commands to `Get/Set/Remove-CsConfigurationModern`,
respectively, with `DefaultAutoRestParameters.ConfigType=TeamsTargetingPolicy`.
All three have `CmdletType=Remoting` and export to TeamsGA, TeamsPreview,
SelfHost and Torus. `CmdletConfiguration` loads these definitions lazily;
`TrpsCmdletAutoRestCommandInfo` resolves the mapped command from the module.
This provides the local modern implementation, separately from whether it is
selected. [F entries at lines 1822, 3804, 4566; P
`CmdletConfiguration`, `TrpsCmdletAutoRestCommandInfo`.]

P `FlightingUtils.GetFlightingCommandInfo` applies this order:

| Condition | Selected path |
| --- | --- |
| Explicit `Legacy` override | Disabled modern flight; TRPS processor |
| Torus/internal host and mapped modern command exists | Modern regardless of flight membership |
| Ordinary host, mapped modern command exists, and either flight membership or explicit modern override | Modern |
| Otherwise | TRPS processor |

The command lookup uses a non-null local
`SessionConfiguration.CmdletConfigOverride` entry before the embedded F mapping.
That dictionary is initialized empty. **The telemetry constructor does not
parse it from the server response.** The inspected origin is
`SetCsSessionState.AddCmdletConfigOverrideForAutorest`, which deserializes local
JSON to `CmdletConfig` objects and resolves local command objects. That helper
also permits adding/removing local flight names and mocks; F exports
`Set-CsSessionState` to **SelfHost only**. It is not an ordinary tenant
negotiation header or remote policy input. [P `FlightingUtils.GetModernCommandInfoForInvokedCommand`,
`SessionConfiguration`, `SetCsSessionState.ProcessRecord`; F lines 83–88.]

`TrpsPSCmdlet.BeginProcessing` accepts the exact internal values `Legacy`,
`TryModern`, `ForceModern` and removes that parameter from transmitted
application parameters. `TryGetFlightedApi` maps **both modern values to the
same `OverrideFlightMode.Modern`**. `ProcessRecord/getProcessor` chooses one
processor once. The modern processor invokes the local modern command; the
exception path rethrows. There is **no automatic modern-error-to-TRPS retry**
in these methods. The official [Get policy reference](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/get-csteamstargetingpolicy?view=teams-ps),
checked 2026-11-20, labels this processing-mode parameter for internal use.

No flight string, modern override or selected-command object is serialized in
the evidenced policy REST request. Thus selection is demonstrably client-side.
It remains possible that an unselected modern request works; it is also
possible that the absent flight reflects a rollout or incompatibility. The
client source cannot distinguish those cases. Positive flight membership is
evidence that the pinned client selects REST, **not proof of server permission**;
absence is **not proof of a missing tenant capability**. [P `FlightingUtils`,
`TrpsPSCmdlet`, `SyncProcessor`; C modern wrappers and P generated configuration
methods cited in the companion contract.]

## Forest discovery and `X-MS-Target-Uri`

The actual modern forest-discovery request is:

| Property | Evidenced value |
| --- | --- |
| Method | GET; default `HttpWebRequest` method |
| URL | `{selected ConfigApiEndpoint origin}/Teams.Tenant/serviceDiscovery` |
| Authorization | `Bearer {same session Teams/ConfigAPI access token}` |
| Accept | `application/json` |
| Body/query | None |
| Ordinary tenant headers | No `X-MS-Target-Uri`, tenant-ID or internal-region header added by this direct discovery method |
| Successful response | 200 JSON object; read `Endpoints.AdminServiceEndpoint` |

H `ConfigApiAutoDForestProvider.ModernAutodiscoverInternal` concatenates the
session endpoint's `Uri.ToString()` (including its trailing slash) and
`Teams.Tenant/serviceDiscovery`. The selected endpoint is an origin, not the
generated request's development host. The consumed response is structurally:

```json
{"Endpoints":{"AdminServiceEndpoint":"admin.example.invalid"}}
```

This synthetic hostname illustrates shape only. The client prefixes `https://`
to the field, so the service field is expected to be a **bare endpoint**, not an
already absolute HTTPS URL. Empty/missing endpoint returns null and
`AutodiscoverInternal` throws; HTTP/parsing exceptions propagate. There is no
ordinary-tenant fallback to a constant forest or to legacy webdir discovery.
[H `ConfigApiAutoDForestProvider.ModernAutodiscoverInternal/AutodiscoverInternal`.]

`GetForestUri` caches by `LoginInfo.AdministeredDomain` using case-insensitive
keys. Ordinary external environments permit discovery; system-tenant and
NPEINT/Local contexts are exceptions. Torus has additional tenant-ID caching
and internal headers. Its region-context 404 fallback to
`ConfigApiEnvironmentToDefaultForest` is **internal-context behavior**, not a
safe fallback for external tenants. Public default `admin2a.online.lync.com`
is therefore not a universal forest choice. [H `GetForestUri`,
`IsServiceDiscoveryAllowed`, `IsOCERegionContextServiceDiscoveryFailureScenario`;
`DeploymentConfiguration.ConfigApiEnvironmentToDefaultForest`.]

Before every generated ConfigAPI send, P
`Generated.Module.SendAsync` calls
`SendAsyncCallback.AddSessionInfoToHttpRequest`. Its target-header step:

1. Discovers the forest when `session.RpsGfsEndpoint` is absent (or always in
   Torus), then stores it using `SetRpsGfsEndpoint`.
2. Takes **only the resulting URI's host**, forces HTTPS, strips any original
   port/path/query and generates its root absolute URI, with trailing slash.
3. Adds that value as `X-MS-Target-Uri` when nonblank. A request containing
   `HostedMigration` gets `/HostedMigration`; the CORE-01 policy/template/batch
   routes do not use that exception.

For the synthetic discovery result above, the header is
`X-MS-Target-Uri: https://admin.example.invalid/`. The **HTTP destination
continues to be the selected ConfigAPI host**; the forest is a routing value,
not a new bearer-token destination or resource audience. The name
`RpsGfsEndpoint` does not imply a live remote session is needed to obtain it.
[P `SendAsyncCallback.AttachTargetUriHeaderToAutoRestCmdlets`; H discovery methods.]

The callback also attaches the bearer, correlation ID, command-name header,
module user-agent and `Accept-Encoding: br, gzip, deflate`, and replaces the
generated URL's **host and port** with the selected ConfigAPI endpoint. The
`X-MS-CmdletName` value comes from `PSInvocationInfo.GetCommandName()` on the
generated operation's invocation, rather than an inferred REST operation name.
Ordinary CORE-01 requests do not unconditionally carry `X-MS-Tenant-Id`:
that header is added for Torus, or for HostedMigration/Teams.Mms with a tenant
ID. Internal support region/ring and RP-base-URI overrides must not be copied
into an ordinary tenant profile. [P `SendAsyncCallback`, `PSInvocationInfo`;
I generic configuration wrappers.]

These are **headers the pinned client supplies**, not a live experiment proving
which individual headers the service rejects when absent. In particular,
omitting target discovery/header has not been qualified. The same callback
also wraps the telemetry bootstrap, so on an uncached ordinary connection
**service discovery precedes the telemetry POST**, then the policy call uses
the cached forest. Discovery itself bypasses this generated callback, avoiding
a recursive prerequisite.

## Necessary origins and audiences

For the modern CORE-01 path, discovery, telemetry bootstrap and the actual
policy/template/deployment requests all use the **same selected ConfigAPI
origin and resource audience**. H `DeploymentConfiguration` maps these public
cloud profiles explicitly:

| Environment | ConfigAPI origin |
| --- | --- |
| Public / commercial / GCC | `https://api.interfaces.records.teams.microsoft.com` |
| GCCH | `https://api.interfaces.records.gov.teams.microsoft.us` |
| DoD | `https://api.interfaces.records.dod.teams.microsoft.us` |
| Gallatin / China | `https://api.interfaces.records.teams.microsoftonline.cn` |
| Bleu / France | `https://api.interfaces.records.communications.svc.sovcloud.fr` |
| Delos / Germany | `https://api.interfaces.records.communications.svc.sovcloud.de` |

`ConfigApiEnvironmentToAppId` uses resource
`48ac35b8-9aa8-4d74-927d-1f4a14a0b239` for all six. The distinct
`EnvironmentToAppId` entries are module **client IDs**, not resource audiences.
P `SessionStateStore.GetLoginInfo` requests `{resource}/user_impersonation`;
K `TPMCmdletHostContract.GetAccessToken` supplies that audience's token.
Native application credentials use the existing audience `/.default` contract.
Cloud identity authorities are recorded in the linked AUTH-06 evidence.
Configured cloud tuples do not prove operation availability in each cloud.

No call to `webdir.online.lync.com/Autodiscover/AutodiscoverService.svc`, Graph,
the discovered forest host, or a separate Skype resource is made by H's modern
discovery routine. Legacy webdir constants still exist in H
`DeploymentConfiguration.SfbOnlineAutoDiscoveryEndpoints`; their presence does
not put them on this call path.

The full PowerShell Connect experience separately uses Graph to resolve the
tenant's default verified domain: K
`Microsoft.TeamsCmdlets.Powershell.Connect.Utilities.GetTenantInformation.GetTenantDomain`
calls `{Graph resource}/v1.0/organization` and reads
`value[0].verifiedDomains[].isDefault/name`. Its domain-to-tenant helper instead
uses the selected authority's `/{domain}/.well-known/openid-configuration`
without a bearer. Those are Connect identity-context helpers, **not additional
audiences required by `Teams.Tenant/serviceDiscovery`**. A native client can
retain explicit tenant/domain configuration and its existing identity
resolution; introducing a mandatory Graph organization read merely to send the
modern policy request would reproduce an unnecessary module-host coupling.

Microsoft's [application authentication guide](https://learn.microsoft.com/en-us/microsoftteams/teams-powershell-application-authentication)
(updated 2025-09-17, checked 2026-11-20) documents app-based support in
commercial/GCC and GCCH/DoD, with exceptions that do **not** name these three
targeting-policy commands. It distinguishes Teams resource tokens, Graph
onboarding permission and directory-role RBAC. This is documentation-level
eligibility; it does not publish the REST flight rule or least-privilege role
for each route. The [Connect reference](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/connect-microsoftteams?view=teams-ps)
documents the independent Teams/Graph resources and delegated
`user_impersonation`. Its statement that modern module versions no longer need
basic-auth enablement does not establish that every retained TRPS branch is
unreachable.

## Native implementation disposition and closure

The modern route is implementable as native Go HTTP: resolve the fixed
cloud/audience tuple, perform bounded/context-aware service discovery at that
origin, validate the returned endpoint shape, derive the target header, and
send the established configuration route/body. Obtaining the flight response
also requires only native HTTP, although using a module-specific telemetry
exchange as the SDK's capability contract needs an explicit compatibility
decision. It does not require executing PowerShell. [H/P call paths above.]

| Question | Closure |
| --- | --- |
| Where does `RemotingCmdletsFlightedForAutoRest` originate? | **Resolved:** Connect telemetry POST response, nested case-sensitive JSON keys, comma-delimited string parsed to a case-insensitive command set. |
| Is `CmdletConfigOverride` another server route assignment? | **Resolved:** initialized empty; inspected setter is local SelfHost session state; not parsed from the telemetry response. |
| How is the ordinary modern policy mapping selected? | **Resolved:** available mapped command plus returned flight membership, or local internal modern override. |
| Does TryModern automatically retry over TRPS? | **Resolved:** no; TryModern and ForceModern choose the same modern mode in this base class; errors propagate. |
| Is a TRPS session needed for modern discovery or modern policy HTTP? | **Resolved:** no; discovery is a direct ConfigAPI GET and the generated modern processor does not create a remote runspace. |
| What destination/audience/header is missing from simple extracted REST routes? | **Resolved:** same-origin `/Teams.Tenant/serviceDiscovery`, same Teams audience, discovery-derived HTTPS forest root in `X-MS-Target-Uri`; ordinary tenant-ID header is not unconditional. |
| Can native REST safely be declared functional for every otherwise allowed unflighted tenant? | **Unresolved:** the package proves a client override can attempt it, not service acceptance; no published backend rule or unflighted-tenant response proves universal support. |
| Does a missing flight prove a tenant capability failure? | **Resolved negative:** it proves only absent client selection, and can also follow a failed bootstrap. Preserve unknown/error states. |
| Does successful discovery/flight imply authorization? | **Resolved negative:** policy route authorization remains independent; current docs describe app support/RBAC but do not qualify exact roles or every cloud. |

The legacy processor is concrete, not a fabricated alternative REST route:
P `TrpsSessionBuilder.CreatePSSessionAndUpdateSessionStore` creates a PowerShell
session through the selected ConfigAPI origin at
`/OcsPowershellOAuth/{forestHost}?AdminDomain={domain}`, passing the same Teams
token as an `oauth` credential and module-version application arguments.
`TrpsPSCmdlet.CreateTrpsPipelineProcessor` executes the original command in
that remote runspace. Simply POSTing the policy JSON there is not evidenced.
Reproducing this protocol natively would require a separate PS remoting wire
contract; the inspection here establishes the boundary, not that protocol.

Implement the modern policy operations and discovery with versioned offline
fixtures, keeping their status distinct from universal tenant parity. Do not
delete the three commands, narrow them to an invented tenant allowlist, mark
them unsupported on absent flight, or use a hidden processing-mode flag as a
public compatibility promise. To close unconditional parity without native
TRPS, obtain first-party evidence that the modern policy backend covers all
allowed tenant/cloud/auth cases, or a settled service contract explaining and
handling the remaining cases. A single successful tenant read alone would not
prove that universal claim.

Required fixture cases include discovery success/missing/malformed endpoint,
no credential-bearing redirects, target-origin normalization, same-audience
selection, empty/missing/malformed flight response, literal comma splitting,
case-insensitive membership, bootstrap error versus empty selection, local
override precedence, and propagated modern failures without mutation replay.
Cache scope must include the native client's cloud, tenant/identity and
endpoint overrides; the module's process-wide domain cache is evidence of its
behavior, not a portable isolation design. Preserve HTTP status/headers/raw
response separately from any capability assessment.

No endpoint manifest, fixture, SDK code, backlog state or coverage status is
changed by this research note. Live authorization, real response fixtures,
per-cloud deployment and universal unflighted-tenant acceptance remain
separate qualification work.
