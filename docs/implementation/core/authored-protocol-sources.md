# Primary sources for authored CORE admin protocol vectors

Reviewed 2026-11-26 against the verified MicrosoftTeams **8.1.0** package at `/private/tmp/msteams-core90-pinned/module`. This investigation imported the local module, reflected IL/attributes, and serialized/deserialized controlled objects. It did not connect, authenticate, invoke a tenant cmdlet, build Go, or capture a service response. All example IDs and values below are authored offline values. HTTP acceptance and a returned handle are separate from successful background work.

## Reproducible primary provenance

Package: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0), [versioned package download](https://www.powershellgallery.com/api/v2/package/MicrosoftTeams/8.1.0). Package SHA256 recorded in [baseline inspection](baseline-inspection.md): `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`.

| Label | Exact package member | SHA256 independently read in this investigation |
|---|---|---|
| C | `bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll` | `8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6` |
| T | `netcoreapp3.1/Microsoft.Teams.PowerShell.TeamsCmdlets.dll` | `3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24` |

Offsets below are decimal IL byte offsets, as reported by the existing read-only `/private/tmp/msteams-core90-inspect.ps1` helper. Prefix **G** abbreviates `Microsoft.Teams.ConfigAPI.Cmdlets.Generated`; prefix **P** abbreviates `Microsoft.Teams.PowerShell.TeamsCmdlets`. An async name such as `G.ConfigApiBasedCmdlets+<BatchTeamsDeploymentNew_Call>d__147` means its `MoveNext` method. Method/token offsets are evidence of the pinned implementation, not public REST documentation. Assembly version `9.1116.2.0` belongs to this package's generated private DLL; it is not the PowerShell package version.

For response casing, merely finding CLR properties was insufficient. Evidence below uses actual pinned `ToJsonString`/`FromJsonString`, JsonProperty attributes, and default Newtonsoft serialization on controlled instances. For region/migration/hierarchy generic reads, `P.Utils.HttpUtilities.Get<T>(HttpClient, Uri, Dictionary<string,string>)` checks `IsSuccessStatusCode` at16, reads content at24/29, and calls `Newtonsoft.Json.JsonConvert.DeserializeObject<T>` at49 (generic method token resolved with region type). This connects the models to wire deserialization. Its two-argument overload delegates to that implementation. These broad success checks justify a selected authored200 reply, not a claim that the live service emits every2xx code.

## Batch deployment

**C / generated response dispatch:** `G.ConfigApiBasedCmdlets+<BatchTeamsDeploymentNew_Call>d__147.MoveNext` reads status at629, compares202 at649, dispatches `onAccepted` with `Task<IFlwoServiceModelsFlwosBatchRequestResponse>` at810 and reads the body at832. `G.Models.FlwoServiceModelsFlwosBatchRequestResponse.ToJson` emits `instanceId` at67 and `executionId` at130. Actual controlled serialization produced:

```json
{"instanceId":"offline-instance","executionId":"offline-execution"}
```

Calling its static `FromJsonString` on `{"instanceId":"offline-only","future":false}` then `ToJsonString` returned only `{"instanceId":"offline-only"}`. Thus either identity is an optional modeled string; unknown-field roundtrip retention is an SDK enhancement, not pinned model behavior. This source does not establish any terminal state in a submission receipt.

**C / status:** `G.ConfigApiBasedCmdlets+<BatchTeamsDeploymentStatusGet_Call>d__151.MoveNext` compares200 at701, passes `Task<string>` to `onOk` at907 and `ReadAsStringAsync` at929.400/404/500 handlers are also declared. The successful body is opaque text, not a generated JSON state model. A code-ready authored reply is status200 with bytes `offline opaque status\r\n`; assert exact text/body and response metadata, and that ID and Identity routes escape the orchestration ID as one path segment. Do not parse the text or claim completion from it. Microsoft's [batch create](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csbatchteamsdeployment?view=teams-ps) and [batch status](https://learn.microsoft.com/en-gb/powershell/module/microsoftteams/get-csbatchteamsdeploymentstatus?view=teams-ps) docs separately describe submission and querying its generated orchestration ID.

Existing pinned request implementation is represented by [batch service](../../../msteams/teams_admin_api/batch_deployment/service.go) and [archive](../../../msteams/teams_admin_api/batch_deployment/archive.go): discovery precedes POST `/Teams.FLWOrchestrator/orchestration/v1.0/batchflwdeployment` with JSON `DeploymentCsvs` and `UsersToNotify`; status GET uses `/scenarios/{orchestrationId}`. The package custom wrapper is `custom/Merged_custom_PsExt.ps1` function `New-CsBatchTeamsDeployment` (line2304; internal invocation2357). Keep the reviewed archive request expectation; authoring response202 does not establish production archive eligibility.

## Multi-Geo region

**T / caller:** `P.GetMultiGeoRegion.ExecuteMultiGeoRegionForEntity` builds user `/api/tms/prod/users/` at34 with `/region` at45 and calls `HttpUtilities.Get<GetMultiGeoRegionResponse>` at80; the group equivalent begins123 and calls169. Both use the same model/default serializer; neither is a Graph region schema.

**T / wire serialization:** Actual `JsonConvert.SerializeObject` on a populated `P.Model.GetMultiGeoRegionResponse` yielded:

```json
{"Region":"offline-region","LastMovementTimeStamp":"2026-01-01T00:00:00+00:00","PreferredDataLocation":"offline-location"}
```

The default instance serializes all three fields as null. `Region` and `PreferredDataLocation` are strings; `LastMovementTimeStamp` is Nullable<DateTime>. There are no JsonProperty overrides on these properties. Selected authored status200 can therefore exercise string/date/null fields using these exact names. Synthetic `"future":{"flag":false}` may additionally test SDK raw-field preservation, explicitly as an unknown field and without claiming the pinned type retains it. A separate null reply should assert nullable members, not a guessed default region. The [Microsoft Multi-Geo documentation](https://learn.microsoft.com/en-us/microsoft-365/enterprise/multi-geo-capabilities-in-teams-in-microsoft-365?view=o365-worldwide) describes current location, preferred data location and last movement time; exact wire casing above comes from the pinned serializer.

Use the existing [region service](../../../msteams/teams_admin_api/region/service.go) request contract and exact delegated `https://migrationsvc.teams.microsoft.com/Region.Read.All` scope. A successful synthetic HTTP response does not establish that scope has been consented or granted to the SDK identity.

## Private channel migration

**T / caller:** `P.GetTenantPrivateChannelMigrationStatus.ProcessRecord` builds `/channels/privateChannelMigrationStatus` at130 and calls `HttpUtilities.Get<PrivateChannelMigrationStatusResponse>` at165. Keep its existing cloud origin/prefix/audience request vectors; a different API is not implied.

**T / actual serialization** of populated `P.PrivateChannelMigrationStatusResponse`:

```json
{"TenantId":"offline-tenant","MigrationStatus":"offline-open-status","LastChecked":"2026-01-01T00:00:00+00:00","Details":"{\"offline\":false}","MigrationStartTimeStamp":null,"MigrationCompletionTimeStamp":null}
```

All six fields are emitted by the default instance as null. TenantId/MigrationStatus/Details are strings; the other three are Nullable<DateTime>. No JsonProperty overrides exist. This proves the exact `TenantId` and `TimeStamp` spelling. `Details` is JSON **text inside a string**; it is not a nested object on the wire. The deliberately named offline key is caller-controlled nested text in this vector, not a claimed service Details schema. A selected authored200 exercises explicit `DecodeDetails`, nullable movement timestamps and raw future fields; malformed Details text should remain readable until explicitly decoded.

The [first-party migration cmdlet page](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/get-tenantprivatechannelmigrationstatus) independently documents the typed object, string Details, example timestamps, and status values. This investigation uses the pinned model to retain LastChecked as well; docs may show a different subset. Neither the synthetic open string nor null timestamps prove migration has started/completed in a tenant.

## Targeting hierarchy v1/v2

**T / status routes and models:** `P.GetTeamTargetingHierarchyStatus.ProcessRecord` selects v2 `/api/v1/hierarchies/{tenant}/operations/{id-or-latest}` using strings at175/192, and v1 `/api/rest/v1/nodes/upload/status` at274 with optional requestId query at243. Generic GET calls are `RetailServiceOperation` at432 and `HierarchyServiceOperation` at447, then command projection constructors at437/452.

Pinned **JsonProperty attributes** establish:

| Native model | Exact wire fields and CLR types |
|---|---|
| `P.Model.RetailServiceOperation` | `id` string; `status` string; `errorDetails` ErrorDetails; `lastUpdatedAt`, `lastModifiedBy`, `filename` strings |
| `P.Model.HierarchyServiceOperation` | `id` string; `status` string; `error` OperationError; `lastActionDateTime`, `lastModifiedBy`, `fileName` strings |
| `P.Model.ErrorDetails` | `errorMessage` string field; `lastKnownStatus` string property |
| `P.Model.OperationError` | `message` string; `details` ErrorDetail[] |
| `P.Model.ErrorDetail` | `code`, `message` strings; `arguments` Dictionary<string,string> |

Default serialization was performed and produced the exact names above with nulls. These timestamps are strings in the pinned models, so use an authored opaque timestamp to verify no invented date conversion. Enum inspection supplies v1 `Starting`, `Ingesting`, `Successful`, `Failed`, `Unknown` and v2 `NotStarted`, `Running`, `Succeeded`, `Failed`, `UnknownFutureValue`. Selected200 examples:

```json
{"id":"offline-v1","status":"Successful","errorDetails":null,"lastUpdatedAt":"offline-time","lastModifiedBy":null,"filename":"fixture.csv"}
```

```json
{"id":"offline-v2","status":"Succeeded","error":null,"lastActionDateTime":"offline-time","lastModifiedBy":null,"fileName":"fixture.csv"}
```

Assert native status plus projected status independently (`Succeeded` becomes command presentation `Successful`), ID, filename, null modifier and raw fields. For a failed v2 operation, the attributes support `{"id":"offline-v2","status":"Failed","error":{"message":"offline top-level","details":[{"code":"OfflineCode","message":"offline {x}","arguments":{"{x}":"value"}}]}}`; the error code/text are authored, not real service diagnostic claims. Existing local projection vectors have further pinned projection coverage; wire success tests must actually invoke the service and projection.

**T / imports and preparation:** `P.SetTeamTargetingHierarchy.ProcessRecord` sets multipart content type `application/json` at365/375, basename at393, field name `csvFile` at404 and multipart Add at411. `UploadCsvTargetingV1` builds `/api/rest/v1/nodes/upload` at6, POSTs at77, checks IsSuccessStatusCode at89, reads/deserializes `Dictionary<string,string>` at168/173/183 and writes it unmodified at191. It imposes **no required response key**. A code-ready authored200 may use `{}` to prove valid optional-handle absence. For `{"requestId":"offline-v1"}`, provenance is the unmodified dictionary implementation plus the [first-party Set example](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/set-teamtargetinghierarchy?view=teams-ps), which displays requestId. Inferring that displayed key is the v1 wire key is justified by the unmodified dictionary path; the method itself does not constrain it. Do not make requestId mandatory for v1.

`UploadCsvTargetingV2` performs GET `/api/v1/hierarchies/{tenant}` at71; compares404 at88; constructs `P.Model.TeamHierarchy` at121; sets bound tenant Id at133 and literal Operational Hierarchy at139/144; serializes at153; POSTs `/api/v1/hierarchies` at184. Its response-success checks are198/283. It then POSTs `/api/v1/hierarchies/{tenant}/importCsv` at408, checks success422, reads content505/510/515, deserializes Dictionary<string,string> at520, requires wire **id** at529/534 and projects that value into **requestId** at588/595/600/605. Missing id throws400 at563/573. `TeamHierarchy` JsonProperty attributes emit `id`/`displayName` exactly.

Selected authored chain:

| Exchange | Reply | Strong assertion |
|---|---|---|
| v2 hierarchy GET |200, empty body | LookupHTTPAccepted true; no create; body is opaque and never decoded |
| alternate v2 hierarchy GET |404 | absence permits create, not successful lookup |
| v2 hierarchy POST after404 |201, empty body | CreateAttempted/CreationHTTPAccepted true; no invented response entity |
| v2 import POST |202, `{"id":"offline-import"}` | ImportAttempted/ImportHTTPAccepted true; returned RequestID equals wire id, raw native id preserved |
| import after successful preparation |403 | retain successful lookup/create responses and flags, report failure; no replay/rollback |

Those selected2xx statuses are permitted by the pinned `IsSuccessStatusCode` checks, not captured service statuses. Preparation GET/create bodies are never read or interpreted; using an empty opaque body is stronger provenance than inventing hierarchy response fields. Ingestion completion requires a separate status observation. A malformed/missing v2 handle after202 must not erase preparation/HTTP acceptance or fabricate usable identity. The [first-party status page](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/get-teamtargetinghierarchystatus?view=teams-ps) also distinguishes upload initiation from querying its outcome.

`P.RemoveTeamTargetingHierarchy.ProcessRecord` checks IsSuccessStatusCode at293 and does not decode a success entity. Selected204 empty reply can assert response metadata and one attempt for both versions; live deletion or restoration is outside this investigation.

## Targeting policy Config API

The pinned `custom/CmdletConfig.json` maps `Get-CsTeamsTargetingPolicy` to `Get-CsConfigurationModern` with ConfigType TeamsTargetingPolicy (lines1822–1834); corresponding Remove/Set entries begin3804/4566. Public wrapper types inherit TrpsPSCmdlet and their property declarations alone do not prove modern wire schemas. The modern generated configuration handlers do:

| C / `G.ConfigApiBasedCmdlets` state machine | Status and payload dispatch evidence |
|---|---|
| `<ConfigurationGetConfiguration_Call>d__228.MoveNext` |200 at649; `onOk` Task<IXdsConfiguration> at810; reads body832 |
| `<ConfigurationGetConfigurations_Call>d__232.MoveNext` |200 at649; `onOk` Task<IXdsConfiguration[]> at810; reads body832: bare array |
| `<ConfigurationUpdateConfiguration>d__234.MoveNext` | route `/Skype.Policy/configurations/`91 + `/configuration/`113; HTTP **Put**432 |
| `<ConfigurationUpdateConfiguration_Call>d__240.MoveNext` |204 at649; bodyless `onNoContent`810 |
| `<ConfigurationDeleteConfiguration_Call>d__224.MoveNext` |204 at649; bodyless `onNoContent`810 |

Do not cite `<ConfigurationUpdateConfiguration1...>` for PUT: its HTTP method is Patch (432), despite a similar route and response dispatcher.

Actual controlled `G.Models.XdsConfiguration.FromJsonString` and `ToJsonString` on `{"Identity":"Global","CustomTagsMode":"FutureServerMode","Description":null,"future":{"flag":false}}` returned:

```json
{"CustomTagsMode":"FutureServerMode","future":{"flag":false},"Identity":"Global"}
```

Thus exact Identity spelling and an extensible configuration property bag are established; null Description is discarded by that pinned roundtrip. The SDK may preserve original null/raw fields from the response, but that is an explicit SDK behavior, not pinned reserialization parity. Selected authored200 entity can use the object above and selected list200 a bare array containing it. To assert Global-first stable sorting and case-insensitive tag filtering, add controlled tag identities and actual known fields from the existing reviewed Global fixture; label synthetic future/nullable values separately. An SDK accepted arbitrary property bag is not evidence every arbitrary server setting is supported.

Use the existing [policy service](../../../msteams/teams_admin_api/targeting_policy/service.go) sparse PUT request and selected204 empty response; DELETE204 likewise has no result entity. Successful update/delete assertions should include response metadata, one mutation and no fabricated readback, rather than inventing a policy receipt body.403/429/503 and malformed/undeclared read responses remain negative tests. Permanent Global fixture reads are separate from writing/deleting Global; these authored contracts do not authorize tenant mutations.

## Template Update201 / Delete200 variants

**C / exact handlers:** `G.ConfigApiBasedCmdlets+<TeamTemplateUpdate_Call>d__960.MoveNext` compares200 at719 and201 at730 (earlier ordering comparison201 at710); `onOk` at925 and `onCreated` at1240 both receive `Task<ICreateTemplateResponse>`, with content reads947/1262. This directly proves201 is a declared update response using the same generated receipt type as200. It does not prove a live tenant will return201 for a selected update.

Actual controlled `G.Models.CreateTemplateResponse.ToJsonString` emitted:

```json
{"id":"offline-id","name":"Offline","scope":"Tenant","channelCount":0,"appCount":0,"locale":"en-US","@odata.id":"/api/teamtemplates/v1.0/offline/Tenant/en-US"}
```

All values are authored. Use this with201 for a reviewed variant; assert ID/name/counts/locale/scope/opaque OData ID. Prefer the already captured receipt for known200/201 create paths; do not label this serialization a capture. Default serializer emits the two zero counts, so those numeric fields have generated defaults.

`G.ConfigApiBasedCmdlets+<TeamTemplateDelete_Call>d__944.MoveNext` compares200 at692 and204 at703; `onOk` at881 and `onNoContent` at1196 receive `Task<IAny>`, with reads903/1218. Calling `G.Models.Any.FromJsonString` on `{"futureReceipt":false}` succeeded; its `ToJsonString` returned `{}`. A status200 `{}` reply therefore has direct pinned parser provenance. A raw futureReceipt field can test SDK preservation separately, but the pinned Any model discards unknown content. Existing captured204 remains the captured delete contract; authored200 is a declared variant, not a new capture.

## How to attach provenance to authored replies

For each vector, cite this note plus package version, DLL hash label, exact class/method and offsets above. Use a label such as “authored from pinned8.1.0 serializer/dispatcher” and retain the primary package/doc URL. Give each materially distinct reply its own provenance: discovery source does not prove the subsequent resource reply; a status selector does not prove an invented body; a model serializer does not prove a terminal workflow. Unknown fields should be named and asserted as forward-compatibility probes, never described as documented service fields. No live status or ledger qualification is promoted by this note.


## Published CSV inputs and update-only policy scope

Rechecked the current first-party [frontline deployment guide](https://learn.microsoft.com/en-us/microsoft-365/frontline/deploy-teams-at-scale) and [hierarchy upload reference](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/set-teamtargetinghierarchy?view=teams-ps) on 2026-11-26. Both publish caller column layouts. The positive protocol vectors use those layouts with explicitly authored identity/name rows. Batch archive construction and hierarchy multipart streaming continue to preserve bytes; they do not locally validate tenant resources, licensing or service ingestion. The batch guide requires licensed team owners, and hierarchy import replaces the singleton: these are concrete live prerequisites, not missing column documentation.

The current [targeting policy setter](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/set-csteamstargetingpolicy?view=teams-ps) describes updating an existing instance, including its example. Its public command contract does not establish creation/upsert. The pinned PUT/204 evidence supports the SDK update operation; manufacturing a create result from the HTTP verb would add an unsupported claim.
