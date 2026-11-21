# CORE-01 template, targeting-policy and batch-deployment evidence

Research revision: 2026-11-20; baseline MicrosoftTeams **8.1.0**. This note covers ten exported proxy commands and two compiled local builders assigned by [CORE-01](../backlog/issues/CORE-01.md). It establishes package-derived contracts, not live qualification. No credentials, tenant operations, or management cmdlet invocations were used. Generated clients, actual pipeline callbacks, custom wrappers, models and base classes were inspected with ILSpy 11.2.0.10050. Decompiled files remain temporary.

## Reproducible sources and confidence

All package paths and type names below refer to [Microsoft's pinned package](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0), package SHA256 `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Sources are abbreviated in the contract sections; a type plus method identifies the evidence independently of decompiler formatting.

| ID | Package source | SHA256 |
| --- | --- | --- |
| P | `bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll` | `8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6` |
| C | `custom/Merged_custom_PsExt.ps1` | `03dbc3a85e12c98dd1807d604782c6a90d0656569541cd9540738825a7fcb255` |
| F | `custom/CmdletConfig.json` | `754d149c1a5904158e9f01eb55ba44d15c17b7e66f7556385ea894f4ff84d906` |
| E | `exports/ProxyCmdletDefinitionsWithHelp.ps1` | `5a8a5a50081b1c6986b93d5b15723e966a2d22b64459a9c300cb2a1bf69bb3b5` |
| H | `netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll` | `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49` |
| B | `netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll` | `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51` |

`G` below means P type `Microsoft.Teams.ConfigAPI.Cmdlets.Generated.ConfigApiBasedCmdlets`; `Models.X` means its sibling `Generated.Models.X`; generated command types are in `Generated.Cmdlets`. Policy dispatch/callback types are in `Microsoft.Teams.ConfigApi.Cmdlets`. Exact public parameters remain pinned by [runtime.json](../../../contracts/powershell/8.1.0/runtime.json); helper parameters are not additional exported commands or public parameter sets.

Confidence is **high for static serialization/dispatch behavior**, **conditional for server-selected policy routing**, and **unobserved for service acceptance, role enforcement, throttling and tenant-specific behavior**. Official Learn pages were checked on 2026-11-20 and are supplementary, unpinned documentation; discrepancies are recorded rather than silently changing the pinned surface.

## Shared transport and authentication boundary

The follow-up [routing investigation](policy-routing-evidence.md) settles the
native discovery request and flight-list bootstrap, including the target-URI
header and local-only override source. It retains the unflighted-tenant service
acceptance question explicitly.

All remote routes in this note use the Config API profile. Generated template/policy requests initially contain `https://api.interfaces.records.teams.microsoft.com`; generated deployment requests initially contain `https://skypeanalyticsapi-dev.trafficmanager.net`. **The latter is not the actual selected production authority.** `Module.CreatePipeline` appends `Module.SendAsync`, which invokes `SendAsyncCallback.AddSessionInfoToHttpRequest`; that callback replaces the host and port with `sessionInfo.ConfigApiEndpoint` before sending. It adds `Authorization: Bearer <session access token>`, `X-MS-CmdletName`, `X-MS-Correlation-Id`, module user-agent, and `Accept-Encoding: br, gzip, deflate`. It discovers the RPS forest when necessary and adds its HTTPS origin as `X-MS-Target-Uri`. Tenant-ID headers are conditional (Torus/internal context, HostedMigration, or Teams.Mms), not unconditional for these families. [P: `Generated.Module.SendAsync/CreatePipeline`; `SendAsyncCallback.AddSessionInfoToHttpRequest/AttachTargetUriHeaderToAutoRestCmdlets/AttachTenantIdHeaderToAutoRestCmdlets`.]

Use the shared audience `48ac35b8-9aa8-4d74-927d-1f4a14a0b239` and evidenced cloud/session origins in [session evidence](../auth/session-evidence.md) and [routing evidence](../http/routing-evidence.md). Do not authenticate these routes with Graph tokens or expose the generated development origin as a service profile. Sovereign profile existence does not demonstrate that every operation is deployed in that cloud. Exact least-privilege roles, application-permission eligibility, and per-cloud availability remain unqualified for these commands.

No inspected family method implements retry, automatic polling, continuation-link traversal or mutation replay. Preserve cancellation and HTTP status/headers/raw body at the shared boundary. A retried batch submission could create a second orchestration: no idempotency token is present in its request. Shared transport retry policy must not turn that uncertainty into automatic replay. [P: G family methods and generated `ProcessRecordAsync` methods.]

## Team templates: routes, public sets and output

`escape(x)` below means .NET `Uri.EscapeDataString(x)` applied to the **whole parameter**, including slashes. Keep the returned template `@odata.id` as opaque data; it is not permission to send to an arbitrary URL. For example `/api/teamtemplates/v1.0/example/Tenant/en-US` becomes `%2Fapi%2Fteamtemplates%2Fv1.0%2Fexample%2FTenant%2Fen-US` after the fixed `/Teams.Templates/` prefix. [P: G `TeamTemplateGet/Delete/Update`.]

| Export | Pinned sets and required application inputs | Method and relative route | Response dispatch |
| --- | --- | --- | --- |
| `Get-CsTeamTemplateList` | One custom set; optional `PublicTemplateLocale` | GET `/Teams.Templates/api/TeamTemplates/v1.0/{escape(locale)}` | 200 `GetTemplatesResponse`; 400/404 template error; default typed error |
| `Get-CsTeamTemplate` | `Get`: `OdataId`; `GetViaIdentity`: `InputObject` | GET `/Teams.Templates/{escape(odataId)}` | 200 `GetTemplateDetailResponse`; 400/404/405 template error; default typed error |
| `New-CsTeamTemplate` | `New`: `Locale`, `Body`; `NewExpanded`: `Locale`, `DisplayName`, `ShortDescription`; `NewViaIdentity`: `InputObject`, `Body`; `NewViaIdentityExpanded`: `InputObject`, `DisplayName`, `ShortDescription` | POST `/Teams.Templates/api/TeamTemplates/v1.0/Tenant/{escape(locale)}` | 201 `CreateTemplateResponse`; 400/500 template error; default typed error |
| `Update-CsTeamTemplate` | `Update`: `OdataId`, `Body`; `UpdateExpanded`: `OdataId`, `DisplayName`, `ShortDescription`; `UpdateViaIdentity`: `InputObject`, `Body`; `UpdateViaIdentityExpanded`: `InputObject`, `DisplayName`, `ShortDescription` | PUT `/Teams.Templates/{escape(odataId)}` | 200/201 `CreateTemplateResponse`; 400/404 template error; default typed error |
| `Remove-CsTeamTemplate` | `Delete`: `OdataId`; `DeleteViaIdentity`: `InputObject` | DELETE `/Teams.Templates/{escape(odataId)}` | 200/204 `Any`; 400 template error; default `Any` error |

Sources: E export definitions; C `Get-CsTeamTemplateList` (line 2825), `New-CsTeamTemplate` (line 2881); P G `TeamTemplateGetList/Get/New/Update/Delete`, corresponding `_Call` and `*ViaIdentity` methods. The exported via-identity Get/Update/Remove commands read `InputObject.OdataId`, reject null, and call the same direct methods as direct input; both New via-identity forms read `InputObject.Locale` similarly. The lower-level G `*ViaIdentity` URI-regex methods exist but are not the calls used by those exported parameter sets. Identity input is not arbitrary URI forwarding. [P: `GetCsTeamTemplate_GetViaIdentity.ProcessRecordAsync`, `RemoveCsTeamTemplate_DeleteViaIdentity.ProcessRecordAsync`, both `UpdateCsTeamTemplate_UpdateViaIdentity*` and `NewCsTeamTemplate_NewViaIdentity*` `ProcessRecordAsync` methods.]

The list wrapper defaults an omitted locale to `en-US`. It checks whitespace as well, but uses `PSBoundParameters.Add`, so an explicitly bound whitespace/null key can encounter duplicate-key behavior: the reliable default is omission. It requests all tenant custom templates regardless of locale, while locale selects Microsoft templates; hidden templates remain listed, according to [Get-CsTeamTemplateList](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/get-csteamtemplatelist?view=teams-ps). That page also advertises `GetViaIdentity`, which the pinned exported custom wrapper does **not** expose (only its internal helper does).

The list response is `{ "value": [summary, ...] }`; the wrapper emits its elements. Summary and create/update responses contain `id`, `name`, `scope`, `description`, `shortDescription`, `iconUri`, integer `channelCount`/`appCount`, string `modifiedOn`/`modifiedBy`/`locale`, and `@odata.id`. Detail is `{ "value": <TeamTemplate> }`, and Get emits that value. `TeamTemplate.BeforeFromJson` also flattens nested `teamDefinition` and `extendedProperties` objects into the model's root before decoding (in that order), then removes those wrapper keys. The merge fills missing keys only: root keys take precedence over `teamDefinition`, which takes precedence over `extendedProperties`; an existing root null also blocks a nested replacement. `JsonObjectExtensions.Merge` and the appended root-collision/null-mask model observations establish that distinction. Request serialization remains flat. Retain fixtures for this response/input compatibility transform. No `nextLink`, cursor or pagination parameter appears in the models or command. A 200 list response can still be incomplete: `X-Teams-TemplateSvc-Diagnostics` containing `TemplatesRetrievalFailure` (case insensitive) causes a warning while returning the partial list. Expose this diagnostic to Go callers. [P: `Models.GetTemplatesResponse/GetTemplateDetailResponse/TeamTemplateSummary/CreateTemplateResponse`; `GetCsTeamTemplateList_Get.overrideOnOk/onOk`; `GetCsTeamTemplate_Get.onOk`.]

Template error payload is `{ "operationId": "...", "errors": [{ "code": "...", "message": "...", "action": "..." }] }`. Named 400/404/405/500 handlers commonly **write error objects to the normal PowerShell output stream**; default handlers construct `RestException<T>` and a PowerShell error. Get's 404 override warns when decoded response is null. Remove's default override does likewise. Remove's 204 path attempts generic JSON decoding before its callback, so a real empty 204 response requires a fixture rather than assuming the parser accepts it. Go should return a status-bearing error with the typed envelope for non-success HTTP responses, preserving the PowerShell compatibility difference explicitly. [P: `Models.TeamTemplateErrorResponse/ErrorObject`; generated template cmdlet `on*` and `overrideOn*` methods.]

Create/update are immediate HTTP responses in the pinned client, not a Graph long-running-operation abstraction. [Update documentation](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/update-csteamtemplate?view=teams-ps) describes submitting an operation, but the inspected implementation does not poll or consume a Location header. GET/DELETE have no request body; POST/PUT serialize UTF-8 JSON with `Content-Type: application/json`.

### Template inputs and omission semantics

Create and update expanded sets share these fields; body sets accept `ITeamTemplate`. The public PowerShell singular spellings do not necessarily match wire keys. [P: `Models.TeamTemplate.ToJson`; `NewCsTeamTemplate_NewExpanded`, `UpdateCsTeamTemplate_UpdateExpanded`; E.]

| Public property | JSON key / type |
| --- | --- |
| `DisplayName`, `ShortDescription` | `displayName`, `shortDescription`: required strings in expanded parameter binding |
| `App`, `Category`, `Channel` | `apps`: app objects; `categories`: strings; `channels`: channel objects |
| `Classification`, `Description`, `Icon` | `classification`, `description`, `icon`: strings |
| `DiscoverySetting`, `FunSetting`, `GuestSetting`, `MemberSetting`, `MessagingSetting` | `discoverySettings`, `funSettings`, `guestSettings`, `memberSettings`, `messagingSettings`: objects |
| `IsMembershipLimitedToOwner` | `isMembershipLimitedToOwners`: nullable Boolean (plural wire name) |
| `OwnerUserObjectId`, `PublishedBy`, `Specialization` | `ownerUserObjectId`, `publishedBy`, `specialization`: strings |
| `TemplateId`, `Uri`, `Visibility` | `templateId`, `uri`, `visibility`: strings |

Top-level null references are omitted; empty strings serialize as empty strings; null arrays are omitted while non-null empty arrays serialize `[]`; null elements are skipped. An omitted membership switch leaves nullable Boolean unset/omitted; explicitly false serializes false. There is no explicit JSON-null emission for these known fields. Model objects contain no additional-properties bag, so unknown body keys are not preserved as a general JSON passthrough. Parameter-level required checks on expanded sets must not be confused with a server schema validation of arbitrary Body JSON. [P: `Models.TeamTemplate` constructors and `ToJson`; generated expanded setters.]

The appended CORE-27 observations also establish that empty serialized model
objects are dropped from app/channel/tab arrays. An explicitly empty tab
configuration can therefore leave an empty tab object, which is dropped while
the channel retains `tabs: []`. Known properties of a different JSON type are
ignored by the generated typed readers; absent/null settings still initialize
the five false-default settings objects. These are actual model executions, not
claims about server validation or tenant update retention.

Creation adds a General channel if no supplied channel has `displayName` equal to `General` under PowerShell's case-insensitive equality. The inserted object has `id: "General"`, `displayName: "General"`, `isFavoriteByDefault: true`. This applies both to Body and expanded inputs, including an empty array. Update does **not** run this custom insertion. Implement create normalization on a copy to avoid mutating caller-owned Go input. [C: `New-CsTeamTemplate`, lines 3100–3140.]

Nested models are not uniformly sparse:

| Object | Wire properties and default behavior |
| --- | --- |
| App | `id` string, omitted when null |
| Channel | optional strings `id`, `displayName`, `description`; nullable Boolean `isFavoriteByDefault`; optional `tabs` array |
| Channel tab | strings `id`, `teamsAppId`, `name`, `sortOrderIndex`, `messageId`, `webUrl`, `key`; optional `configuration` object |
| Discovery settings | `showInTeamsSearchAndSuggestions` always serialized as a Boolean, default false |
| Fun settings | `allowGiphy`, `allowStickersAndMemes`, `allowCustomMemes` always serialized, default false; optional string `giphyContentRating` |
| Guest settings | `allowCreateUpdateChannels`, `allowDeleteChannels` always serialized, default false |
| Member settings | `allowCreateUpdateChannels`, `allowDeleteChannels`, `allowAddRemoveApps`, `uploadCustomApp`, `allowCreateUpdateRemoveTabs`, `allowCreateUpdateRemoveConnectors`, `allowCreatePrivateChannels` always serialized, default false |
| Messaging settings | `allowUserEditMessages`, `allowUserDeleteMessages`, `allowOwnerDeleteMessages`, `allowTeamMentions`, `allowChannelMentions` always serialized, default false |

Sources: P `Models.TeamsAppTemplate`, `ChannelTemplate`, `ChannelTabTemplate`, `TeamDiscoverySettings`, `TeamFunSettings`, `TeamGuestSettings`, `TeamMemberSettings`, `TeamMessagingSettings`, each `ToJson` and fields. Thus supplying an empty settings object is materially different from omitting it. **Construction history also matters:** `TeamTemplate` settings getters lazily initialize objects. Its JSON constructor uses those getters when a corresponding JSON object is absent, so parsing a JSON Body populates all five settings objects with false defaults even though constructing the equivalent sparse expanded input does not. Reading a settings property can likewise change later serialization. Actual [model fixtures](../../../contracts/endpoints/8.1.0/fixtures/core-model-serialization.json) verify both cases. `ChannelTabTemplate.Configuration` has a lazy getter too, but the actual serializer omits the empty configuration object in the captured parsed-tab case; a getter alone does not prove an emitted `{}`. Tab configuration is `TeamsTabConfiguration`, with optional strings `entityId`, `contentUrl`, `removeUrl`, `websiteUrl`; null fields are omitted, empty strings retained. Its model has no arbitrary additional-properties map. [P: `Models.TeamsTabConfiguration.ToJson`.]

The official [template administration guide](https://learn.microsoft.com/en-us/microsoftteams/get-started-with-teams-templates-in-the-admin-console) documents limits of 15 channels, 20 tabs per channel and 50 apps. Those service constraints are not local validation in the inspected wrapper; do not assert live validation or exact failure codes. [New-CsTeamTemplate](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamtemplate?view=teams-ps) and [Get-CsTeamTemplate](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/get-csteamtemplate?view=teams-ps) describe template administration, not creating an actual team.

## Batch deployment

| Export | Parameters | Route / response |
| --- | --- | --- |
| `New-CsBatchTeamsDeployment` | One public set: mandatory strings `TeamsFilePath`, `UsersFilePath`, `UsersToNotify` | POST `/Teams.FLWOrchestrator/orchestration/v1.0/batchflwdeployment`; 202 response object; all other statuses raw-string default error |
| `Get-CsBatchTeamsDeploymentStatus` | `Get`: mandatory `OrchestrationId`; `GetViaIdentity`: mandatory `InputObject` | GET `/Teams.FLWOrchestrator/orchestration/v1.0/scenarios/{escape(orchestrationId)}`; 200/400/404/500 raw strings, undeclared statuses raise undeclared-response error |

Sources: E exports; C `New-CsBatchTeamsDeployment` lines 2304–2365; P G `BatchTeamsDeploymentNew/StatusGet`, `_Call` and `StatusGetViaIdentity`. The exported via-identity status command requires non-null `InputObject.OrchestrationId` and passes it to the same escaped direct method; it does not forward InputObject as a URI. There are no list/page parameters. [P: `GetCsBatchTeamsDeploymentStatus_GetViaIdentity.ProcessRecordAsync`.]

New's custom wrapper copies input files into temporary `Teams.csv` and `Users.csv`, compresses them into `TeamsDeployment.Zip` at Fastest compression using `Compress-Archive -Update`, reads bytes and sends Base64 without line breaks. The generated payload has exact case and spelling:

```json
{"DeploymentCsvs":"<base64 ZIP containing Teams.csv and Users.csv>","UsersToNotify":"admin@example.invalid"}
```

`DeploymentCsvs` is plural even though the internal parameter is `DeploymentCsv`. Null strings are omitted by the generated serializer, but all three public inputs are mandatory. No CSV parsing or row validation occurs in the inspected wrapper. Use isolated/in-memory ZIP construction in Go; do not reproduce the fixed shared temporary filenames or preserve stale archive entries. [C wrapper; P `Models.FlwoServiceModelsFlwosBatchRequest.ToJson`.]

A 202 body is `{ "instanceId": "...", "executionId": "..." }`. The PowerShell override emits a human-readable string containing `instanceId` as OrchestrationID and discards `executionId` from that output. Canonical Go should preserve both typed fields and use `instanceId` for status lookup. Failed submission emits a generic deployment-not-submitted error with status mapped to InvalidArgument for 400, ObjectNotFound for 404, InvalidOperation otherwise; the override does not expose the returned service string as output. Status handlers instead emit the returned strings even for the named 400/404/500 statuses. [P `Models.FlwoServiceModelsFlwosBatchRequestResponse`; `NewCsBatchTeamsDeployment_NewExpanded.overrideOnAccepted/overrideOnDefault`; `GetCsBatchTeamsDeploymentStatus_Get.on*`.]

The server works asynchronously after 202; the client does not wait. [Status documentation](https://learn.microsoft.com/en-gb/powershell/module/microsoftteams/get-csbatchteamsdeploymentstatus?view=teams-ps) names Running and Completed, and Completed can include success, partial success or failure. Therefore completion must not be mapped to success. The exact raw status body grammar and per-team errors need representative evidence; do not invent a structured status object from the help output.

The [frontline deployment guide](https://learn.microsoft.com/en-us/microsoft-365/frontline/deploy-teams-at-scale) documents up to 500 teams, 25 users per team, five email recipients, and Teams licenses for owners. Teams CSV columns in order: Team Name, Existing Team ID, Visibility, Team Template ID. Users CSV columns in order: User Full Name, User UPN or ID, Team Name, ActionType, Owner or Member. Both require headers. Documented values include Public/Private, AddMember/RemoveMember and Owner/Member. These are documentation constraints, not checks observed in local code. [New-CsBatchTeamsDeployment](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csbatchteamsdeployment?view=teams-ps) also describes notification emails. The docs advertise WhatIf/Confirm, but the pinned public custom wrapper lacks SupportsShouldProcess; internal generated New supports it. Preserve the actual exported metadata distinction.

## Teams targeting policies: conditional modern contract

These three exports are tag-setting policies, separate from the hierarchy-upload APIs. Their private cmdlet classes inherit `TrpsPSCmdlet`; F maps them to `Get/Set/Remove-CsConfigurationModern` with `ConfigType=TeamsTargetingPolicy`. `FlightingUtils.GetFlightingCommandInfo` consults session `RemotingCmdletsFlightedForAutoRest`, checks that a modern command exists, and honors session `CmdletConfigOverride`. Legacy/TryModern/ForceModern are internal dispatch overrides stripped from transmitted application parameters. The unflighted path invokes the command over a TRPS runspace. Do not claim that static availability proves every tenant uses modern routing. [P `Get/Set/RemoveCsTeamsTargetingPolicy`, `TrpsPSCmdlet`, `FlightingUtils`; F entries.]

When the pinned modern mapping is selected, the routes are established:

| Public operation | Method and relative path | Body / result |
| --- | --- | --- |
| Get without Identity; Get with Filter | GET `/Skype.Policy/configurations/TeamsTargetingPolicy` | 200 JSON array of arbitrary configuration objects |
| Get with Identity | GET `/Skype.Policy/configurations/TeamsTargetingPolicy/configuration/{escape(identity)}` | 200 configuration object |
| Set | PUT `/Skype.Policy/configurations/TeamsTargetingPolicy/configuration/{escape(identity)}` | JSON configuration object; 204 without public output |
| Remove | DELETE `/Skype.Policy/configurations/TeamsTargetingPolicy/configuration/{escape(identity)}` | No body; 204 without public output |

Source: C `Get-CsConfigurationModern` line 299, `Set-CsConfigurationModern` line 1014, `Remove-CsConfigurationModern` line 958; P generated `GetCsConfiguration_Get/Get1`, `SetCsConfiguration_Update`, `RemoveCsConfiguration_Delete`, G `ConfigurationGetConfigurations/GetConfiguration/UpdateConfiguration/DeleteConfiguration`. There is also a generic PATCH method in G, but the inspected Set wrapper selects **PUT**, not that method. `apiVersion`/`schemaVersion` query parameters exist only on generic helpers; they are omitted by these public wrappers (generated callers check bound-parameter presence). No ETag/If-Match appears.

Get has two optional sets, `Identity` and `Filter`; nonempty Identity addresses a single configuration, omission lists all. Filter is applied **locally** after list retrieval using PowerShell wildcard matching on `Identity` or `Tag:` plus Filter; it is not a server query parameter. Results are converted from raw configuration JSON to PowerShell objects, with Global sorted first. The generic wire model is `XdsConfiguration`: `Identity` plus arbitrary properties, not a nested `properties` envelope. No pagination fields/traversal were found. [C Get wrapper and `Sort-GlobalFirst`; P `Models.XdsConfiguration`, generated Get overrides.]

Set has optional Identity and these optional string properties: `CustomTagsMode`, `Description`, `ManageTagsPermissionMode`, `ShiftBackedTagsMode`, `SuggestedPresetTags`, `TeamOwnersEditWhoCanManageTagsMode`. Omitted Identity becomes `Global` in the wrapper and is also included in the body. Bound properties only enter PropertyBag; null additional-property values are filtered out in `XdsConfiguration.CopyFrom`, empty strings remain. For example the statically derived payload for setting one field is:

```json
{"Identity":"Global","CustomTagsMode":"Enabled"}
```

Do not populate omitted policy settings with guessed defaults. Identity and Filter reject null/empty when supplied; Remove requires Identity. ValidSet **metadata** names Disabled/Enabled for the three toggle fields, and Disabled/EnabledTeamOwner/EnabledTeamOwnerMember for ManageTagsPermissionMode. Those are `CmdletParamMetainfo` attributes, not ordinary PowerShell ValidateSet enforcement. [P policy classes, `TrpsPSCmdlet.GetParameters`, `Models.XdsConfiguration.CopyFrom/ToJson`; C Set wrapper.]

[Set-CsTeamsTargetingPolicy](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/set-csteamstargetingpolicy?view=teams-ps) additionally documents EnabledTeamOwnerMemberGuest and MicrosoftDefault for ManageTagsPermissionMode. Treat that as a documented-but-not-baseline-metadata discrepancy; do not hard-reject server values using an incomplete closed enum. [Get-CsTeamsTargetingPolicy](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/get-csteamstargetingpolicy?view=teams-ps) confirms the all-policies default and wildcard Filter. [Remove-CsTeamsTargetingPolicy](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/remove-csteamstargetingpolicy?view=teams-ps) says removing Global resets defaults and describes fallback assignments, but contains meeting-policy copy/paste text. The service-side reset/default values and assignment effects remain unobserved; the client still sends DELETE.

Generic configuration non-success dispatch decodes `IErrorResponse` with top-level optional strings `code`, `message`, `action` and raises a typed REST error. [P: `Models.ErrorResponse`.] Actual policy validation codes and any unmodeled server fields still need fixtures. The wire response is not the same as PowerShell's formatted/deserialized type decorations. A native Go implementation can use the evidenced modern endpoint and typed known fields plus retained unknown JSON, but **legacy tenant parity remains conditional** until server-discovery eligibility and fallback disposition are explicitly resolved.

## Hidden-template builders: settled local disposition

`New-CsTeamsHiddenTemplate` and `New-CsTeamsHiddenMeetingTemplate` each have one set, mandatory non-null/nonempty string `Id`, no ShouldProcess, and construct `HiddenTemplate`/`HiddenMeetingTemplate` respectively. Their `CreateNew` methods merely call the model constructor. The model sets Id (rejecting null); Identity remains unset. [H `NewHiddenTemplateCmdlet`, `NewHiddenMeetingTemplateCmdlet`, `HiddenTemplate`, `HiddenMeetingTemplate`.]

The base chain was inspected: B `BaseTpmNewElementCmdlet<T>.ProcessRecord` calls CreateNew, a no-op default ValidateProperties, wraps the result as PSObject and writes it. ArgumentException becomes InvalidArgument. B `BaseTpmCmdlet` initializes HTTP-client infrastructure, logging and telemetry, but its BeginProcessing/ProcessRecord perform no management request or connection authorization check. This settles **local object construction, no management endpoint and no management audience**, while acknowledging the module's incidental telemetry infrastructure. The builders do not persist a permission policy.

Official [hidden team template](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamshiddentemplate?view=teams-ps) and [hidden meeting template](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamshiddenmeetingtemplate?view=teams-ps) pages confirm these are objects supplied to the respective template-permission policy commands. Canonical Go mapping: pure constructors/value types `HiddenTemplate{ID: ...}` and `HiddenMeetingTemplate{ID: ...}`. Their eventual policy payload serialization belongs to the permission-policy contract, not a fabricated builder route.

## Canonical Go operations and remaining gates

Proposed mappings fit the shared root-client/service/transport architecture; names are design proposals, not already implemented methods.

| Family | Canonical operations | Important model boundary |
| --- | --- | --- |
| Team templates | `List(ctx, locale)`, `GetByID(ctx, TemplateReference)`, `Create(ctx, locale, TeamTemplate)`, `UpdateByID(ctx, TemplateReference, TeamTemplate)`, `DeleteByID(ctx, TemplateReference)` | Config API; opaque OData identity; optional top-level values and exact nested Boolean behavior; return diagnostics for incomplete list |
| Batch deployment | `Create(ctx, DeploymentArchive, recipients)`, `GetByID(ctx, orchestrationID)`; portable `BuildDeploymentArchive(teamsCSV, usersCSV)` helper | Typed submission handle with instance/execution IDs; raw status until grammar established; explicit polling separate from submission |
| Targeting policies | `List(ctx)`, `GetByIdentity(ctx, identity)`, `UpdateByIdentity(ctx, identity, TargetingPolicyUpdate)`, `DeleteByIdentity(ctx, identity)`; optional compatibility Filter helper | Conditional modern Config API routing; sparse bound-property body; Global default; known fields plus retained unknown fields |
| Hidden builders | `NewHiddenTemplate(id)`, `NewHiddenMeetingTemplate(id)` | Local only; no service request |

Template CRUD/list is evidenced in full. Deployment has create/submit and read-status only: no list, update or delete endpoint is evidenced. Targeting-policy read/list/update/delete are evidenced conditionally; there is no assigned create command or settled create contract for this family. Hidden builders have no remote CRUD lifecycle. Do not invent missing operations to fill an interface.

Each public parameter set above collapses into the same canonical operation; PowerShell Body/expanded and InputObject/direct pairs are input conveniences, not separate backend operations. Runtime debugging, PowerShell proxy credentials, pipeline callbacks, WhatIf/Confirm and OutBuffer do not belong in service request models; context, transport configuration and dry-run facilities belong at their appropriate shared boundaries.

Before marking these commands contract-complete or advancing CORE-01:

1. Add versioned endpoint manifests and executable fixture coverage from this evidence: exact encoded OData paths; missing versus empty arrays; false versus omitted top-level Boolean; non-sparse nested settings; create-only General insertion; locale default; partial-list diagnostic; named and unknown HTTP errors; empty delete response; batch ZIP entries/field casing; policy PUT sparse payload and local wildcard/Global ordering.
2. Fixture both JSON-Body conversion and sparse expanded input, including lazy settings defaults, identity property extraction, tab configuration and policy error envelopes. Preserve unknown status text instead of inventing a deployment status schema.
3. Resolve tenant flighting/session override applicability for the three policies and document native handling for legacy-only tenants. Static endpoint extraction does not settle this conditional capability.
4. Qualify app-only/role restrictions, sovereign deployment, live errors, throttling and eventual-consistency behavior in a sanitized lab when available. Keep this distinct from offline contract tests.

These remaining gates deliberately keep affected implementations blocked; the local-builder disposition and concrete static route/serialization findings can be consumed immediately. No coverage status was changed by this research note.
