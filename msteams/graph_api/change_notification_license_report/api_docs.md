# Change-notification license report composition

Access through `sdk.GraphAPI.ChangeNotificationLicenseReports`. This service is
read-only; it does not create a subscription, assign licenses or change roles.

| Method | Result |
| --- | --- |
| `ListActivity(ctx, period, options)` | Beta JSON activity rows and partial paging receipts |
| `ListUsers(ctx, options)` | Directory ID/UPN/display-name projection |
| `ListLicenseDetailsByUserID(ctx, directoryID, options)` | v1.0 license details for one directory ID |
| `Get(ctx, period, options)` | Bounded cross-resource rows and per-user license outcomes |
| `ProjectPinnedRows(activity, users, licenses)` | Pure, separate pinned PowerShell output projection |

Periods are 7, 30, 90 or 180 days. Activity uses beta JSON, `$top=2000` and
`$format=application/json`; users use v1.0 `$top=999`. License requests use the
first-party `/users/{id}/licenseDetails` route, not concatenated UPN syntax.
Reported ID, reported UPN and resolved directory ID are separate fields. A nonempty
reported directory ID is authoritative; otherwise an exact unique UPN match is required from
a complete user traversal. No UPN normalization or reused-name retargeting occurs.

`Get` completes activity and user traversal before resolving subjects, then batches
at most twenty license reads with correlated string IDs. Default collection budgets
are 1,000 pages/100,000 items and the composition permits 10,000 distinct users;
callers may lower these through `ListOptions`/`GetOptions`. Continuations retain
same-cloud HTTPS origin, exact resource and license subject. Limits, cycles,
cancellation and later-page errors return partial data and receipts. A failed
outer batch stops later chunks; denied individual replies never imply false.

Graph documents `Reports.Read.All` for application/delegated activity reads and
separate directory-user permissions. LicenseDetails currently documents delegated
`LicenseAssignment.Read.All` plus a supported Entra/custom role, with application
permissions unsupported. `Get` and the license primitive return `ErrLicenseMode`
locally in application mode, without acquiring a different identity. Report/user
primitives remain separately usable. No scope acquisition, role change, delegated
fallback or permission denial is treated as successful qualification.

Native `HasMatchingSourcePlan` is a nullable predicate for the pinned plan ID,
not entitlement/billing/effective authorization. Missing, null, malformed or denied
plan data stays unknown/error; only complete valid nonmatching data yields false.
A known match alongside malformed data remains positive with an error and
`PlanDataComplete=false`. Inspect `Complete`, stage flags and errors, rather than
assuming outer HTTP 200 establishes every subrequest or row.

Counters preserve JSON numeric precision and raw omission/null/unknown fields;
native derived totals use `big.Int` with a bounded JSON-number computation.
`ProjectPinnedRows` separately preserves source Int32-wrap behavior, missing-license
row omission, present-null empty-string verdict, ordinal UPN/plan comparisons,
display-name fallbacks and malformed nested-data errors. It performs no network
operation and does not replace native rows with source diagnostic strings.

The conditional compiled cmdlet has one Period set; it is commented preview in the
root manifest. Native primitives and pure projection do not establish default-session
export, supported app license access or successful tenant qualification. See the
[primary/source evidence](../../../docs/implementation/ext/license-report-contract-notes.md).


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
