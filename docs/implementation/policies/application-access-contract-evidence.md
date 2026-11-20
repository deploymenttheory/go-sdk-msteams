# POL-01 application access policy contract evidence

Research revision: **2026-11-20**, MicrosoftTeams **8.1.0**. This bounded note
covers `Get/Grant/New/Remove/Set-CsApplicationAccessPolicy`, the first five
commands in [POL-01](../backlog/issues/POL-01.md). It establishes their selected
modern client contracts, not ticket completion, production qualification or
unconditional REST availability for every tenant. No module import, management
cmdlet execution, credentials, connection or tenant request was used. A local
PowerShell language-only experiment is identified separately below.

**The resource needs Create, Get, List, Update and Delete, plus explicit user,
tenant-default and group assignment actions.** Granting a policy is separate
from updating its AppIds. New is a POST followed by a GET; Set selects PUT,
despite a separate PATCH method existing in the assembly. These distinctions
come from the complete mapping chain below, not command-name inference.

## Sources and reproducibility

The package is [MicrosoftTeams 8.1.0 on PowerShell Gallery](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0),
SHA256 `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`.
Original files were hashed with `shasum -a 256`; temporary decompilations are
outside the repository. These source IDs identify package paths throughout.

| ID | Package or repository source | SHA256 |
| --- | --- | --- |
| P | `bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll` | `8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6` |
| F | `custom/CmdletConfig.json` | `754d149c1a5904158e9f01eb55ba44d15c17b7e66f7556385ea894f4ff84d906` |
| C | `custom/Merged_custom_PsExt.ps1` | `03dbc3a85e12c98dd1807d604782c6a90d0656569541cd9540738825a7fcb255` |
| I | `internal/Merged_internal.ps1` | `6ff95158174bab031dc1fc24112a2bc6ae81cf1e2f59db4ecb58afbefc8c2bd9` |
| E | `exports/ProxyCmdletDefinitionsWithHelp.ps1` | `5a8a5a50081b1c6986b93d5b15723e966a2d22b64459a9c300cb2a1bf69bb3b5` |
| M | `en-US/MicrosoftTeams-help.xml` | `6829e6302742ebe986572c1e25b5d4796d2cdde0d87b67217ecac908906e1b44` |
| H | `netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll` | `477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af` |
| K | `netcoreapp3.1/Microsoft.TeamsCmdlets.PowerShell.Connect.dll` | `c1c2b2d1c4a5306fb70f373df7ed5bfca4f9db58e90be358afd8a1d48ed46aea` |
| R | [normalized runtime metadata](../../../contracts/powershell/8.1.0/runtime.json) | `7acfc9d635da928a2cf3deb4ff6f660f239edb218581613708a27a74f48aa1d8` |

ILSpy CLI **11.2.0.10050** inspected methods offline, for example:

```sh
env DOTNET_ROOT=/opt/homebrew/opt/dotnet/libexec \
  DOTNET_CLI_HOME=/private/tmp/msteams-dotnet-home \
  /private/tmp/msteams-dotnet-tools/ilspycmd \
  -t Microsoft.Teams.ConfigApi.Cmdlets.SetCsApplicationAccessPolicy \
  /private/tmp/msteams-baseline-810/module/bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll
```

P namespaces are abbreviated: `Custom` means
`Microsoft.Teams.ConfigApi.Cmdlets`; `Generated` means
`Microsoft.Teams.ConfigAPI.Cmdlets.Generated`, including its `.Cmdlets`,
`.Models` and `.Runtime` children. Type/method names are reproducible citations
independent of decompiler line formatting. Package source is the authority for
8.1.0 behavior; current Microsoft documentation supplements its meaning.

Official references checked **2026-11-20** are
[Get](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/get-csapplicationaccesspolicy?view=teams-ps),
[Grant](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/grant-csapplicationaccesspolicy?view=teams-ps),
[New](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csapplicationaccesspolicy?view=teams-ps),
[Remove](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/remove-csapplicationaccesspolicy?view=teams-ps)
and [Set](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/set-csapplicationaccesspolicy?view=teams-ps).
These are mutable `teams-ps` references, not an 8.1.0 REST specification; their
rendered parameter tables disagree with pinned metadata in places. The Graph
[application access policy guide](https://learn.microsoft.com/en-us/graph/cloud-communication-online-meeting-application-access-policy)
shows update date **2024-11-07**; the Teams
[application authentication guide](https://learn.microsoft.com/en-us/microsoftteams/teams-powershell-application-authentication)
shows **2025-09-17**. Both were checked on the research date.

## Verified mapping and selection boundary

E exports functions that dispatch into P's five `Custom.*CsApplicationAccessPolicy`
classes, all inheriting `Custom.TrpsPSCmdlet`. F explicitly contains:

| Public command | Modern command | Injected default |
| --- | --- | --- |
| Get | `Get-CsConfigurationModern` | `ConfigType=ApplicationAccessPolicy` |
| New | `New-CsConfigurationModern` | `ConfigType=ApplicationAccessPolicy` |
| Set | `Set-CsConfigurationModern` | `ConfigType=ApplicationAccessPolicy` |
| Remove | `Remove-CsConfigurationModern` | `ConfigType=ApplicationAccessPolicy` |
| Grant | `Grant-CsTeamsPolicy` | `PolicyType=ApplicationAccessPolicy` |

Every entry has `CmdletType=Remoting`, module
`Microsoft.Teams.ConfigAPI.Cmdlets`, and TeamsGA/TeamsPreview/SelfHost/Torus
exports. F entries start at lines 906, 2052, 2504, 3322 and 3930. C defines the
mapped functions at lines 299, 773, 893, 958 and 1014. I maps their internal
commands to the generated P cmdlets. [F/E/C/I; P
`Custom.TrpsCmdletAutoRestCommandInfo` constructor,
`Custom.TrpsPSCmdlet.GetParameters/ProcessRecord`.]

All five use the selection algorithm traced in
[shared policy routing evidence](../core/policy-routing-evidence.md): ordinary
modern selection needs an available mapping and command-specific returned
flight membership, or a local internal modern override. A Get flight does not
establish New/Set/Remove/Grant selection. P's `Custom.FlightingUtils.GetFlightingCommandInfo`
and these five inherited `TrpsPSCmdlet.TryGetFlightedApi` paths were checked.
`TryModern` and `ForceModern` both select the same modern mode; the processor
does not automatically retry a failed modern request through TRPS. Without
modern selection it retains the original command's remoting path. Flight
absence proves neither REST rejection nor service authorization. No flight
result or actual service acceptance was observed here.

## Origin, audience, discovery and headers

Every generated operation below uses `Generated.Module.CreatePipeline` and its
`SendAsync` callback. This was checked in Get/Set generated cmdlets and each
Grant generated branch; C routes New/Remove through the same generated module.
Thus the shared routing evidence applies to these particular mappings, unlike
unrelated compiled policy-provider paths. [P `Generated.Module.BeforeCreatePipeline/AfterCreatePipeline/SendAsync`;
`Generated.Cmdlets.GetCsConfiguration_Get/Get1`,
`SetCsConfiguration_Update`, `GrantCsUserPolicy_GrantExpanded`,
`GrantCsTenantPolicy_GrantExpanded`,
`GrantCsGroupPolicyAssignment_GrantExpanded.ProcessRecordAsync`.]

The ordinary generated request goes to the selected ConfigAPI origin. The
public cloud origin is `https://api.interfaces.records.teams.microsoft.com`;
GCCH, DoD and China use respectively
`https://api.interfaces.records.gov.teams.microsoft.us`,
`https://api.interfaces.records.dod.teams.microsoft.us` and
`https://api.interfaces.records.teams.microsoftonline.cn`. H also maps Bleu
and Delos origins as recorded in the shared note. The resource audience is
`48ac35b8-9aa8-4d74-927d-1f4a14a0b239`, not Graph or the discovered forest.
Delegated host acquisition uses `/user_impersonation`; application credentials
use the resource's `/.default` scope. Cloud identity authorities remain owned by
[AUTH-06](../auth/session-evidence.md). Configured cloud tuples do not prove
these operations are deployed in every cloud. [H `DeploymentConfiguration`;
P `Custom.SessionStateStore.GetLoginInfo`; K `TPMCmdletHostContract.GetAccessToken`;
shared routing evidence and application authentication guide.]

Before an uncached ordinary generated call, H
`ConfigApiAutoDForestProvider.ModernAutodiscoverInternal` performs GET
`{ConfigApiOrigin}/Teams.Tenant/serviceDiscovery` with the same bearer and
`Accept: application/json`. It consumes `Endpoints.AdminServiceEndpoint` as a
bare host, prefixes HTTPS, and the callback derives its HTTPS root for
`X-MS-Target-Uri`. Missing/malformed discovery is an error, not permission to
guess a forest. The destination remains ConfigAPI. [H method above; P
`Custom.SendAsyncCallback.AttachTargetUriHeaderToAutoRestCmdlets`.]

P `Custom.SendAsyncCallback.AddSessionInfoToHttpRequest` adds the bearer,
`X-MS-Target-Uri`, `X-MS-Correlation-Id`, `X-MS-CmdletName`, module User-Agent and
`Accept-Encoding: br, gzip, deflate`, replacing the generated URI's host/port
with the selected endpoint. `X-MS-CmdletName` comes from invocation information,
not a guessed public-command string. Ordinary `/Skype.Policy` calls do not get
an unconditional `X-MS-Tenant-Id`. Internal region/ring/RP overrides are
conditional host behavior. JSON mutations explicitly set
`Content-Type: application/json` and UTF-8 content. The route methods add no
extra family-specific authorization header. These are supplied-client headers;
individual server-required-header tests remain unobserved.

## Resource routes and command composition

Let `T=ApplicationAccessPolicy`, `N` be the supplied policy Identity and `B` the
bound property bag. Path segments use `Uri.EscapeDataString` independently;
do not interpolate unescaped names. The outer C wrappers pass neither
`apiVersion` nor `schemaVersion`, so their generated optional queries are
absent. No version value should be invented. [P `Generated.ConfigApiBasedCmdlets`
methods cited in the table; I internal Get/New/Set/Remove mappings; C wrappers.]

| Native operation / PowerShell composition | Method and relative route | Request and declared success |
| --- | --- | --- |
| `ApplicationAccessPolicies.List` / Get with no Identity | GET `/Skype.Policy/configurations/{T}` | No body; 200 bare JSON array of configuration objects |
| `ApplicationAccessPolicies.GetByIdentity(N)` / Get with Identity | GET `/Skype.Policy/configurations/{T}/configuration/{N}` | No body; 200 one configuration object |
| `ApplicationAccessPolicies.Create(B)` / first part of New | POST `/Skype.Policy/configurations/{T}` | JSON B including Identity; 201, no success body consumed |
| New output composition | GET item route after successful POST | Reads the new Identity; returns the read object |
| `ApplicationAccessPolicies.UpdateByIdentity(N,B)` / Set | **PUT** `/Skype.Policy/configurations/{T}/configuration/{N}` | JSON B including Identity; 204 |
| `ApplicationAccessPolicies.DeleteByIdentity(N)` / Remove | DELETE `/Skype.Policy/configurations/{T}/configuration/{N}` | No body; 204 |

Method citations in P `Generated.ConfigApiBasedCmdlets` are respectively
`ConfigurationGetConfigurations`, `ConfigurationGetConfiguration`,
`ConfigurationCreateConfiguration`, `ConfigurationUpdateConfiguration`,
`ConfigurationDeleteConfiguration` and each corresponding `_Call` method.
C `New-CsConfigurationModern` submits, stops when its error variable is set,
then reads; a successful mutation followed by a failed GET must remain a
successful submission with a readback failure, not trigger another POST.
C Set inserts `Identity="Global"` only when that key is absent, then passes
that Identity in both route and body. That generic fallback does **not** prove
the backend accepts a Global ApplicationAccessPolicy resource or documents its
defaults. The current Set parameter table calling Identity mandatory conflicts
with the pinned optional parameter and this fallback. [C; E; R; Set reference.]

`ConfigurationUpdateConfiguration1` emits PATCH to the same item path, but
I `Set-CsConfiguration` with `-Body` selects `SetCsConfiguration_Update`, whose
`ProcessRecordAsync` calls **`ConfigurationUpdateConfiguration` (PUT)**.
Do not replace this with PATCH or assume ordinary full-object replacement
semantics merely because the selected verb is PUT: the sent body is the sparse
bound bag plus Identity. Server-side treatment of omitted properties remains
separately unobserved. [I `Set-CsConfiguration`; P generated methods above.]

Get's nonblank `Filter` performs the collection GET and filters **locally** with
PowerShell `-Like` against `Identity` or `Tag:{Filter}`. The help's description
of a nonstandard filter is not the implementation. A blank Identity in the
generic helper would select collection, but public compiled Identity/Filter
parameters have `ValidateNotNullOrEmpty`; public omission is the reliable List
case. The wrapper orders exact `Global` identities first and otherwise retains
enumeration order. [C `Get-CsConfigurationModern/Sort-GlobalFirst`; P
`Custom.GetCsApplicationAccessPolicy`; Get reference.]

## Parameter sets, identity and update presence

R records 2/3/1/1/1 parameter sets. E owns public function binding; compiled
classes add validation through the runtime custom-attribute processing. Below
are domain parameters; standard diagnostic and ShouldProcess parameters are
omitted from this table. Their complete wrapper propagation is not qualified
by parameter metadata alone.

| Command and set | Domain parameters and binding |
| --- | --- |
| Get `Identity` (default) | Optional string Identity at position 1 |
| Get `Filter` | Optional string Filter, mutually exclusive with Identity |
| New `Identity` | Mandatory string Identity at position 1; optional object AppIds and string Description |
| Set `Identity` | Optional string Identity at position 1; optional object AppIds and string Description |
| Remove `Identity` | Mandatory string Identity at position 1; input by property name |
| Grant `GrantToTenant` (public default) | Optional Global switch and PolicyName string at position 1 |
| Grant `Identity` | Optional Identity string at position 0, input by value/property; optional PolicyName string |
| Grant `GrantToGroup` | Mandatory Group string at position 0; optional nullable Int32 Rank and PolicyName string |

All mutations expose Confirm/WhatIf; Grant also exposes PassThru. All five
expose the hidden internal processing-mode string. Compiled Get Identity/Filter,
New/Set/Remove Identity and Grant Identity/Group reject explicit null/empty
values. Grant PolicyName instead explicitly permits null/empty. The public
snapshot's empty validation arrays do not erase these compiled attributes.
[E; R; P `Custom.Get/New/Set/Remove/GrantCsApplicationAccessPolicy`.]

Identity denotes a policy for CRUD, a user for user Grant and Group a group
identifier for group Grant. The pinned help M and current Grant reference
describe PolicyName as the policy name without `tag:`. The wrapper does not
strip prefixes from CRUD route identities or normalize user identifiers;
preserve their values. Published examples use the user's Entra object ID.
UPN/SIP/name acceptance, name length/case rules, policy count limits and AppId
existence/GUID validation are not established by these paths. The client
validates collection elements as strings; it does not validate them as GUIDs.
[M command parameter nodes; C `Grant-CsTeamsPolicy`; P `Custom.PolicyUtils`;
Grant reference.]

`TrpsPSCmdlet.GetParameters` copies **bound parameters**, injects the F default,
then moves parameters absent from the target helper's signature into
`PropertyBag`. For New/Set, this preserves omitted versus explicitly supplied
AppIds/Description; it does not serialize every CLR property default. New/Set
override `ValidateAndUpdatePropertyBagParams` to call
`Custom.PolicyUtils.ValidateAndUpdatePsListModifier("AppIds", ..., typeof(string))`.
Its exact modern representations are:

| AppIds input intent | Property-bag/wire representation |
| --- | --- |
| Omitted | No AppIds key; no injected empty-list default |
| One string or direct collection | `"AppIds":["app-a", "app-b"]` replacement list |
| Empty direct collection | `"AppIds":[]` |
| Hashtable with Add/Remove | `"AppIds":{"Add":["app-a"],"Remove":["app-b"]}`; omit an operator whose collection is empty |
| Hashtable with both operator collections empty | `"AppIds":{}` |
| Hashtable containing Replace | Client-side ArgumentException; use the direct replacement list |
| A PSListModifier object itself | Client-side ArgumentException; legacy type metadata is not permission to pass that object |

These representations are evidenced client output, not proof of backend
atomicity, duplicate handling, add-existing/remove-missing behavior, operator
precedence or accepted empty-list policy semantics. Preserve operators rather
than emulating them with an unqualified read/modify/write race. [P
`Custom.PolicyUtils.ValidateAndUpdatePsListModifier/ConvertToJsonConversionFriendly`;
`Generated.Models.XdsConfiguration.CopyFrom/ToJson`;
`Generated.Runtime.JsonSerializable.ToJson/ToJsonValue`.]

The [offline policy observations](../../../contracts/endpoints/8.1.0/fixtures/policy-model-serialization.json)
reproduce these modern input shapes with the pinned helper and model. The
[generated request observations](../../../contracts/endpoints/8.1.0/fixtures/generated-request-dispatch.json)
separately exercise CRUD/list, escaped Identity, assignment status dispatch,
malformed errors and pre-send cancellation. These isolated primitives do not
execute the exported New POST/readback or flight/discovery compositions.

`XdsConfiguration` owns only string `Identity` plus arbitrary additional
properties. Its dictionary conversion drops null-valued properties, and
`ToJson` omits null Identity. Empty strings and empty arrays survive. In public
PowerShell function binding, an explicit null Description becomes an empty
string before model conversion; omission stays absent. An explicit null AppIds
passes the list-modifier replacement conversion, producing an empty array in
the language-only check described next. There is no evidenced JSON-null clear
operation for these properties. A native presence model must not equate absent,
empty string and empty list, or invent a fourth wire null behavior.

An isolated `/opt/homebrew/bin/pwsh -NoLogo -NoProfile -NonInteractive` script
on **7.6.6** used only local functions and
`[System.Management.Automation.PSListModifier]::new([object]$value)`: bound
`[string]$null` became non-null length-zero with the bound key present;
PSListModifier of null and of an empty array both had empty Replace/Add/Remove
collections. No MicrosoftTeams import or cmdlet was performed. This confirms
language coercion on the installed oracle runtime, not service behavior; the
package method citations remain the contract evidence.

## Assignment actions and null policy removal

C `Grant-CsTeamsPolicy` **requires the PolicyName key to be bound**, despite
optional metadata; omission throws a ParameterBindingException. Group selects
group Grant first. Otherwise nonblank Identity selects user Grant; without it,
Global must be true or the helper throws another ParameterBindingException.
The public default set therefore does not make bare Grant a valid tenant
mutation. [C lines 832–863; E/R.]

| Native assignment action | Method and relative route | Body; declared success |
| --- | --- | --- |
| `AssignToUser(user, policy)` / Grant Identity | PATCH `/Skype.Policy/users/{user}/policies/ApplicationAccessPolicy` | `{"policyName":"Example"}`; 200, no success body consumed |
| `AssignTenantDefault(policy)` / Grant Global | PATCH `/Skype.Policy/tenants/policies/ApplicationAccessPolicy` | `{"policyName":"Example","forceSwitchPresent":false}`; 200, no success body consumed |
| `AssignToGroup(group, policy, rank)` / Grant Group | POST `/Skype.Policy/groupPolicyAssignments/{group}/policyTypes/ApplicationAccessPolicy` | `{"policyName":"Example","priority":2}`; 202 OperationId or 204 |

Sources: P `Generated.ConfigApiBasedCmdlets.UserGrantPolicy`,
`TenantGrantPolicy`, `GroupGrantPolicyAssignment` and corresponding `_Call`;
`Generated.Models.SingleGrantUserRequest/SingleGrantTenantRequest/GroupGrantPayload.ToJson`;
generated `GrantCs*Policy*_GrantExpanded` setters; C `Grant-CsTeamsPolicy`.
Bodies above are synthetic source-derived examples, not captures. The public
command has no Force argument; the generic helper passes its false default as
`forceSwitchPresent` for tenant Grant. No policy-family additional parameters
are supplied. Empty additional-parameter objects are omitted by the serializer.
Group Rank becomes JSON `priority`; C forwards it only when PowerShell truthy,
so zero/null/omission do not send priority, while a supplied negative nonzero
value is forwarded without a client range validator. Backend rank limits and
default placement are unobserved.

The Grant reference documents one direct policy per user, replacement when a
new policy is assigned, explicit null PolicyName to unassign, and tenant Global
as the fallback for users without explicit assignments. In this pinned modern
chain, the public/string helper binding converts null PolicyName to `""`, and
model serializers preserve that empty string: the user-clear body is
`{"policyName":""}`, not a JSON-null field or invented DELETE. Tenant-clear
and group-clear requests likewise preserve the empty PolicyName string, but
their exact removal/fallback service effects were not observed and are not
established by the user-clear documentation. [E/C; P generated models above;
local binding experiment; Grant reference.]

**PassThru discrepancy:** E and `Custom.GrantCsApplicationAccessPolicy` expose
PassThru, but mapped C `Grant-CsTeamsPolicy` has no such parameter and no
PropertyBag. `TrpsPSCmdlet.GetParameters` retains it and `SyncProcessor`
passes all parameters. Static inspection therefore identifies an unresolved
modern binding mismatch when PassThru is explicitly bound; do not promise that
the generated internal cmdlets' PassThru behavior is reachable through this
public wrapper. A credential-free isolated binding test or sanitized lab
observation can settle the complete wrapper behavior. This discrepancy does
not remove the three assignment actions from native parity scope.

## Responses, errors, paging, batching and retry

Get list consumes a **bare array**, not Graph's `value` envelope. An item is an
object with Identity and open properties; AppIds/Description are known inputs,
not proof that every response always contains them. `XdsConfiguration.BeforeJson`
retains the original object. The generated Get success override emits that
JSON string, then C parses it into the displayed policy object and adds format
metadata. A native typed model should retain unknown fields/raw response and
not reproduce PowerShell formatting wrappers. Both Get overrides surface
`X-Ms-ConfigApi-PowerShell-WarningMessage` as a warning. [P
`Generated.Cmdlets.GetCsConfiguration_Get/Get1.overrideOnOk`;
`Generated.Models.XdsConfiguration.BeforeFromJson`;
`generated.custom.Helpers.HttpResponseMessageHelper.TryGetWarning`; C Get.]

The `_Call` methods use the success status table above; all other statuses go
to `ErrorResponse` with optional lower-case `code`, `message`, `action` string
keys. Generated `onDefault` builds PowerShell ErrorRecords, preserving the
action as RecommendedAction, and has a RestException branch for missing
code/message. Invalid/non-JSON errors can fail decoding. The package does not
prove status-specific 400/401/403/404/409/429 bodies or family-specific codes.
Preserve status, headers and raw bytes even when structured decoding fails;
do not replace every failure with a policy-not-found error. [P
`Generated.Models.ErrorResponse.FromJson`; relevant generated `onDefault`,
especially `SetCsConfiguration_Update.onDefault`.]

There are no First/Skip/page-size/continuation arguments, `nextLink` parsing or
paging loops in these collection methods. List performs one request and local
Filter cannot be sent as an OData query. This is a client contract, not evidence
that the service has no hidden maximum or truncation. [C Get; P
`Generated.ConfigApiBasedCmdlets.ConfigurationGetConfigurations/_Call`.]

Group Grant's 202 payload is `{"OperationId":"opaque-job-id"}` with capital
O/I. Its generated `onAccepted` returns that identifier without polling; 204
has no job body. The package separately contains GET
`/Skype.Policy/assignments/operations/{operationId}` returning a BatchJobStatus
on 200, with lower-case `operationId`, `overallStatus`, progress counts and
`userState`. This supports a distinct operation handle/read composition; it
does not establish which jobs the group route actually creates or their
terminal states/retention/poll interval. Do not confuse accepted assignment
with completed propagation. [P `GroupGrantPolicyAssignment_Call`;
`Generated.Models.BatchOperationId`; generated group `onAccepted/onNoContent`;
`ConfigApiBasedCmdlets.BatchGetPolicyAssignmentOperation1/_Call`;
`Generated.Models.BatchJobStatus`.]

Pipeline user inputs use `Custom.CustomRunspacePool` and separate helper
invocations. Its batch optimization dictionary contains only
`Get-CsUserSearch`, not application-access-policy Grant. The documented
`Get-CsOnlineUser | Grant-...` composition therefore means per-input user
assignment, not an evidenced batch-create route or transaction. There is no
whole-pipeline rollback or batch-level replay contract. [P
`Custom.TrpsPSCmdlet.getProcessor`,
`Custom.CustomRunspacePool.BatchPropertiesForCmdlet/BatchPipelineInput`; Grant reference.]

The normal generated pipeline terminal
`Generated.Runtime.CustomHttpClientFactory.SendAsync` calls HttpClient once
with the cancellation token. No policy-specific retry, Retry-After polling,
idempotency token or mutation replay loop appears in these methods. Keep SDK
transport retry policy separate; ambiguous POST/PUT/PATCH/DELETE failures must
not silently replay. The documented **up to 30-minute** delay concerns changes
taking effect in Graph online meeting/virtual event calls; it is not a
ConfigAPI retry interval or proof that Create failed. [P `Generated.Module` and
terminal above; Graph application access policy guide.]

## Authorization and remaining qualification

The application authentication guide includes these five commands under its
general supported-command statement: none is in the published exception list.
It documents app support in commercial/GCC and GCCH/DoD and directory-role RBAC.
It does not settle the least-privilege role per route, every sovereign cloud,
user administrative-unit scope or universal modern-flight acceptance. The
Graph guide's downstream meeting application permissions are separate from
authorization to administer these policies through ConfigAPI. A downstream
Graph Forbidden response is not the management API's error shape. [Application
authentication guide; Graph application access policy guide; shared routing evidence.]

| Contract aspect | Disposition |
| --- | --- |
| Five command mappings, CRUD/list, three Grant scopes and selected verbs | High-confidence pinned static evidence; keep all native operations |
| Request presence, AppIds arrays/operators, string-null coercion and Set default insertion | Source-established client behavior; backend empty/null effects require qualification |
| Discovery/audience/header path | Established shared generated path, verified for these mappings |
| Real response examples, missing/duplicate policies, invalid AppIds, auth/role failures, rank bounds | Unobserved; source-derived examples are not captures |
| Name normalization, Global policy resource support/defaults, removal of an assigned policy | Unresolved; no extrapolation from generic policy rules |
| Modern selection for all eligible tenants/auth modes/clouds | Unresolved; retain the shared flight/legacy boundary |
| Group clear behavior, async job completion and propagation | Unresolved beyond emitted request, 202/204 decoding and separate job-read route |
| Explicit public Grant PassThru in modern mode | Evidenced signature mismatch; complete runtime binding behavior unobserved |
| Confirm/WhatIf through public wrappers | Metadata advertises ShouldProcess; do not treat that alone as an observed dry-run guarantee. New/Set's generic C helpers lack SupportsShouldProcess, and their unmatched bound parameters can enter the property bag; complete propagation needs an isolated binding test before lab mutation work |
| Pagination limits, consistency, concurrent operator atomicity, idempotency and retries | Not published by the inspected contract; do not invent guarantees |

Required downstream fixture work is: distinct Get/List envelopes; escaped
policy/user/group names; each success and raw/non-JSON error path; sparse
Update and AppIds replace/add/remove/empty cases; missing PolicyName and
missing Global validation; null-to-empty user clear; omitted/zero/nonzero Rank;
New's POST/readback split; group 202 versus 204; discovery/flight error
separation; and no mutation replay after ambiguous submission. Keep source
contract tests and live qualification states separate. This note introduces no
endpoint manifest, fixtures, SDK code, coverage claims or backlog closure;
the rest of POL-01 remains outstanding.
