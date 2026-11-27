# Conditional license report contract notes

Research date: 2026-11-26. Scope: credential-free static inspection of pinned public 8.1.0 plus Microsoft Graph documentation. No report cmdlet invocation, sign-in or tenant request. Pure-data observations and offline Go tests are distinguished below. This supports an implementable composition design; it does not qualify the command as a default-session export or a tenant-tested operation.

## Evidence identity

[Conditional declaration metadata](evidence/conditional-command-metadata.json), record `Get-LicenseReportForChangeNotificationSubscription`, pins package SHA-256 `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7` and `Microsoft.Teams.PowerShell.TeamsCmdlets.dll` SHA-256 `3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24`; extractor SHA-256 `b4e39467e853aabd4e9409b7d0060a8d0e8cc1a435061adf8402060ce4346a70`. It records `root_exported: false` and mandatory `Int32 Period`, validated against 7, 30, 90, 180, with no pipeline binding.

The [full IL observation](evidence/license-report-il.json) records declaring
types, assembly hashes, exact method IL SHA-256, instruction offsets/operands,
exception regions and literal constants. These settle the control-flow findings
below. The bounded [read-only decoder](../../../scripts/discovery/inspect-ext-il.ps1)
accepts the pinned ModulePath, TypeName and MethodPattern; it does not invoke
inspected methods. Source JSON model attributes were inspected separately;
those declarations are not operation qualification.

## Evidenced composition and result

1. `ProcessRecord` authorizes with `Reports.Read.All`, then fetches Teams activity, directory users, and licenses, in that order. Helpers construct Graph-resource-relative `beta/reports/getTeamsUserActivityUserDetail(period='D{Period}')?$format=application/json&top={MaximumGetTeamsUserActivityUserDetailPageSize}` and `v1.0/users?$top={MaximumGetUsersPageSize}`; both call `HttpUtilities.GetAll<T>`. Preserve the evidenced report query's literal `top` when describing source behavior; the documented parameter is `$top`. Do not silently call this discrepancy verified equivalent.
2. License candidates are activity-report UPNs, partitioned using `MicrosoftGraphConstants.BatchLimit` with `Take`/`Skip`; helper constructs `/v1.0/$batch`, with indexed GET subrequests `/users('` + `Uri.EscapeDataString(UPN)` + `')/licenseDetails`. No dedicated license-report HTTP route is evidenced. Full IL resolves batch limit **20**, users page size **999**, report page size **10000**. `BatchRequest.Flush` produces UTF-8 `application/json` `{"requests":[...]}`; each request has numeric `id` (zero-based), `method: "GET"`, and `url`. `BatchCall` POSTs this body. JSON response attributes are `responses`, `id` (Int32), `status`, `body.value`, and `body.error`.
3. `GenerateDlpUserLicenseReports` iterates the activity report, joins directory users and license results by UPN, and constructs `DlpUserLicenseReport` with `UserPrincipalName`, `DisplayName` (fallback `Not Found`), `NumberOfMessagesSent` (`Int64`, derived from private-chat and team-chat counts), and `HasRequiredLicensePlan` (string). Lambda predicates test whether **any** license contains **any** service plan whose `ServicePlanId` equals `6dc145d6-95dd-4191-b9c3-185575ee6f6b`. No provisioning-status, SKU-name, or assignment-state predicate appears in that predicate. Full IL establishes a critical distinction: a missing license dictionary key **skips the row** (offset 147 branches to loop advancement 313). For a present key, non-null license collections yield exact strings `True`/`False`; a present null collection yields the empty string through `Nullable<bool>.ToString()`. The null-coalescing `Not Found` license fallback does not replace that empty string. Missing directory users or null display names yield `Not Found`; empty display names remain empty. Counts use unchecked **Int32 addition then Int64 conversion** (offsets 212–214), so overflow wraps before widening. These source behaviors are mechanically implementable; they must not be replaced with a false verdict for missing data.
4. Cross-batch license dictionaries are grouped by UPN and retain the first value, rather than merging every duplicate license collection. Directory `ToDictionary` uses the UPN selector with no explicit comparer shown. Do not introduce case-folding, duplicate tolerance, or normalization as source parity without further evidence.


### Nested null/default behavior (pure-data observations)

The [bounded reproducer](../../../scripts/discovery/inspect-license-report-nulls.ps1) (SHA-256 `1c3518823b9795a7df5ee8e75ced0ba0c0ef5bc789fdf70336324e32fd60b1bb`) and [observations](evidence/license-report-nulls.json) (SHA-256 `f5f167aca7bd31ac3c12876b3d223bd028a763b44675271fce48baab1f564977`) deserialize synthetic JSON into actual pinned models and invoke only the previously IL-audited pure predicates, never the report command. Constructor IL confirms no initializers: `LicenseDetails.ServicePlans` is nullable `IEnumerable<ServicePlanInfo>` defaulting to null; `ServicePlanInfo.ServicePlanId` is **string**, default null, not a CLR Guid. Missing or explicit null `servicePlans` throws `ArgumentNullException(source)` in the nested `Any`; `[]` returns false. Null license items or null plan items throw `NullReferenceException`. Missing/null/empty/all-zero/nonmatching `servicePlanId` returns false; the exact lowercase target string returns true and uppercase spelling returns false. Both nested and outer `Any` short-circuit: a matching item before a null item succeeds, while null before match throws. Thus malformed nested data must not silently become an unlicensed verdict in native composition. A pinned projection can reproduce these exact throws/defaults while retaining a separate unknown/error outcome in native output. Outer null license collection remains the projector's explicit nullable-Boolean bypass (empty string), while outer `[]` returns `False`; this differs from a non-null license object with null servicePlans, which throws. The observation JSON hashes every input and the reproducer, records the pinned assembly hash, and retains source assembly identity. The [license model IL](evidence/license-details-model-il.json) and [plan model IL](evidence/license-plan-model-il.json) establish the field defaults; these are source observations, not tenant replies.

## Pagination and partial failures

Full `GetAll<T>` IL repeatedly GETs `PageResponse<T>`, appends `value`, and follows `@odata.nextLink` until null/empty/whitespace (offsets 6–52). It constructs the next URI directly and uses the same HttpClient. There is no local origin validation, cycle detection, page cap, partial-result return, or per-page recovery. `Get<T>` throws `ApiException` for non-success HTTP responses, preserving status and parsed Graph error or raw response content; empty successful content returns default(T), which `GetAll` then dereferences. An SDK may apply its existing same-origin continuation guard as an explicit safety constraint; it must not claim the PowerShell source supplied that guard. No licenseDetails subresponse pagination traversal exists: only `body.value` is consumed.

The license helper maps each numeric response ID back to the requested UPN. Status 200 adds `body.value` to the dictionary; all other statuses append `Error retrieving license details for user with UPN '…'` plus `body.error.message` to a string list. Individual failed subrequests are not retried. `ProcessRecord` enumerates report objects with `WriteObject(object, true)`, then writes the strings with `WriteObject(object)`: these are success-stream values, not `WriteError` records. Missing users/licenses produce debug diagnostics, and missing license keys omit report rows as above. Duplicate successful UPNs within a batch, malformed/out-of-range IDs, null bodies/error objects, or null response lists receive no defensive checks in this helper and can throw. Across batches, grouping retains the first successful UPN value.

The command batch wrapper catches only the **first** `ApiException` with code 401 using an exception filter (offsets 86–123), calls `PerformAuthorization` with endpoint enum 3 and null permission list, and retries that batch once (125–147). A second 401, other exception, or reauthorization failure propagates. Do not interpret the enum's numeric value or the null permission argument as proof of additional scopes. Whole-batch non-success responses become `ApiException` with parsed Graph error text/code, or raw content if parsing fails; empty successful batch content creates a default `BatchResponse`.

The shared synchronous transport uses Polly `WaitAndRetry` with **five retries** after the initial request, for status **408, 429, 500, 502, 503, 504**. It uses `Retry-After.Delta` when present; the source's date branch reads **DateTimeOffset.Offset**, not date minus current time (retry-lambda offsets 69–94). With neither value it uses decorrelated jitter V2, median first delay 1 second. This retry applies to outer HTTP requests; a 200 batch envelope containing failing subresponses does not trigger it. This date-branch quirk is source evidence, not a recommended interpretation of Retry-After. No cancellation token supplied by the caller reaches this synchronous helper (`CancellationToken.None`).

## Official Graph constraints and implementation decision

The [beta activity report documentation](https://learn.microsoft.com/en-us/graph/api/reportroot-getteamsuseractivityuserdetail?view=graph-rest-beta) documents D7/D30/D90/D180, `$format=application/json`, `$top`/`$skipToken`, JSON `200` with a `value` collection and default 2,000-item page; CSV instead uses a short-lived redirect. Delegated and application access both support `Reports.Read.All`; delegated users additionally need an appropriate role. Beta remains changeable and unsupported for production use. Implement the JSON path explicitly and do not substitute CSV parsing or infer a stable beta contract.

The [v1.0 licenseDetails documentation](https://learn.microsoft.com/en-us/graph/api/user-list-licensedetails?view=graph-rest-1.0) documents GET `/users/{id}/licenseDetails`, a collection response, and `$select`. Its permission table explicitly says **application permissions are not supported**. For work/school delegated access the least privilege is `LicenseAssignment.Read.All`, with a supported Entra role or suitable custom role. Therefore report application permission does **not** establish application support for the full composition. The command's `Reports.Read.All` authorization call does not prove it acquires all downstream permissions; directory-user permission requirements also need separate verification. The source's parenthesized UPN batch syntax is observed source evidence, not independently confirmed by the documented path form.

The composition, numeric constants, page traversal, batch serialization, row inclusion/fallbacks, arithmetic, partial diagnostics and bounded retry rules are now source-backed and implementable with synthetic fixtures; they do not justify blanket unavailability. An SDK representation should retain typed rows and per-user diagnostics explicitly, while documenting any deliberate differences from mixed PowerShell success-stream output and malformed-response exceptions. Remaining external gaps: supported application authentication for the full composition, downstream delegated/directory-user authorization, parenthesized UPN syntax acceptance, the `top`/`$top` discrepancy, and tenant behavior. Preserve conditional/default-session qualification until independent evidence changes it.


## Native implementation and offline validation

`GraphAPI.ChangeNotificationLicenseReports` exposes separate bounded JSON activity,
directory-user and directory-ID license-detail reads, plus `Get` composition.
The native activity request uses documented `$top=2000`, not the source's literal
`top=10000`. Native license requests use the documented `/users/{id}/licenseDetails`
path only after complete directory resolution. Reported user ID, reported UPN and
resolved directory ID remain distinct. A nonempty reported directory ID is authoritative:
missing ID resolution cannot retarget a reused UPN. UPN-only rows require one exact
ordinal match; duplicate/missing identities produce a partial result.

Native batching uses documented unique string IDs, at most twenty subrequests, and
correlates reordered replies. Each stage retains receipts, partial items and errors;
a failed outer envelope stops later batches, while individual failures remain
explicit. Continuations are same-cloud HTTPS, same resource and same license
subject, with finite page/item/user limits and cycle checks. Context cancellation
propagates through the existing client; no reauthorization or identity fallback is
added. Read-only primitives remain available separately when full composition's
license mode cannot be used. The full `Get` and license primitive reject application
mode locally because the current first-party licenseDetails table explicitly does
not support application permission; this is a documented native constraint, not
proof that a pinned app attempt succeeded or that a delegated role is configured.

Activity count values remain precision-preserving `json.Number`. The API examples
and [commercial Graph CSDL observation](evidence/license-report-graph-metadata.json)
do not declare a JSON count width (the function returns `Edm.Stream`), so native
counts do not inherit the pinned Int32 cap. Derived totals accept JSON-grammar
integer-valued decimal/exponent numbers with bounded computation (4,096 characters
and absolute decimal exponent); arbitrary wire values remain in the raw snapshot
when a derived total cannot be computed. `ProjectPinnedRows` separately requires
Int32 representability and reproduces unchecked addition followed by widening.

Missing/null/empty license collections, servicePlans and plan IDs remain distinct.
Native unknown, denied, unparseable or partial data never yields a negative match;
only a complete empty or valid nonmatching collection yields false. A known match
with other malformed plan data retains true with an error and incomplete plan-data
status. The fixed plan test is solely the source's ordinal presence predicate,
not current subscription entitlement, billing or effective authorization.

Focused offline `TestUnit|Example` checks passed in sessions 36483 and 98712 (including compiled examples) on 2026-11-27,
including all four periods, twenty-plus-one batching with reversed response order,
subject-bound continuations and page budgets, partial directory/denied license
outcomes, identity handoff and reused UPN rejection, arbitrary-precision totals,
JSON-number grammar, source Int32 wrap and actual nested source defaults. Replies
are authored primary-contract vectors, not captured tenant responses. This work
owns the remaining conditional command/set without promoting implementation,
contract or live qualification. Downstream delegated credentials/roles and actual
successful read evidence remain external prerequisites; no new license, role or
subscription has been created.


Graph [user](https://learn.microsoft.com/en-us/graph/api/resources/user?view=graph-rest-1.0)
and [directoryObject](https://learn.microsoft.com/en-us/graph/api/resources/directoryobject?view=graph-rest-1.0)
primary schemas define ID as an opaque String; directoryObject explicitly permits
non-GUID IDs. This resource therefore uses a dedicated minimal `DirectoryUser`
snapshot, exact ID comparison and safely escaped ID path segments. A reported
nonempty ID is authoritative even when non-GUID, so a missing match cannot retarget
through UPN. Missing/empty report IDs may use unique exact UPN resolution from a
complete directory read. Continuations compare the decoded single subject segment
exactly, including case, while retaining the fixed resource and trusted origin.
The shared UUID-only directory resolver is unchanged. GUID format remains a
validation rule for documented `servicePlanInfo.servicePlanId`, not user identity.


The opaque-ID correction passed the focused report package in session 63983 on
2026-11-27. The new regression uses a non-GUID ID containing `#`, verifies its
single-segment escaping through both batch and direct reads, follows only its exact
subject continuation, and refuses case-folded ID/UPN fallback. This is an authored
schema-valid vector, not a claim that a particular tenant issues that identifier.
