# Team-template SDK evidence

Work date: 2026-11-20. Status: **in progress, pending live CRUD acceptance**.
This work does not close [CORE-27 / #80](https://github.com/deploymenttheory/go-sdk-msteams/issues/80).

The root collection now exposes `sdk.TeamsAdminAPI.TeamTemplates` with real
Create/GetByID/UpdateByID/DeleteByID/List methods. Typed definitions, summaries,
detail/error snapshots and local hidden-template values cover the assigned
surface. All seven commands and 15 observed parameter sets have Go symbol/test
mappings; their ledger statuses remain in-progress/metadata-only/not-run.

## Resource behavior

The service composes the existing ConfigAPI backend with one discovery adapter
shared by the root collection's policy/template resources. Constructors are
local. The returned forest is a normalized routing header; the bearer remains
bound to the configured ConfigAPI origin/audience. Request metadata includes
the operation's public command name, using the same X-MS-CmdletName purpose as
`SendAsyncCallback.AddUseragentAndCmdletNameHeaders` and its invocation name.

The opaque @odata.id is escaped as one segment and never followed as a URL.
Create adds General on a copy when needed; update sends PUT without that
insertion. Mutation replay is disabled. Create/update return immediate typed
receipts; there is no invented polling/read-back operation. List exposes the
partial-retrieval diagnostic and makes no undocumented paging guarantee.

Body input reproduces known generated-model behavior, including sparse expanded
construction versus false-default JSON settings. Read snapshots preserve unknown
fields, explicit nulls/empties and original envelopes. HTTP errors retain typed
details and shared classification without printing private diagnostics.
Native Go errors replace PowerShell's named-handler console output behavior.
See [api_docs.md](../../../msteams/teams_admin_api/team_template/api_docs.md).

## Pinned observations and contract execution

The existing five captured template request/dispatch cases are exercised through
the **actual service** using CORE-15's root-client runner, together with model
captures and the source-established General-channel wrapper composition.
Additional declared-status/error/cancellation/cloud cases are synthetic contract
tests; they are not represented as new PowerShell executions or tenant success.

Seven new model observations append to
[`core-model-serialization.json`](../../../contracts/endpoints/8.1.0/fixtures/core-model-serialization.json):
root/nested collision priority, root null masking, null array items, empty tab
configuration, other known-property JSON types, null settings and the full known
definition. The original 30 observations and their indexes remain unchanged.
All 12 template-definition cases match Go serialization. The actual captures
caught empty model objects being dropped from arrays while [] remains retained;
the Go implementation was corrected to match.

The capture helper checks the pinned module version and assembly hashes. Source
review of TeamTemplate.BeforeFromJson, JsonObjectExtensions.Merge and the custom
New wrapper agrees with these observations. No decompiled implementation code
is committed.

The actual pinned local builder cmdlets were also executed with empty,
whitespace and synthetic opaque IDs. Both rejected empty IDs with
ParameterBindingValidationException, accepted whitespace/opaque IDs unchanged,
and left Identity unset. Their Go constructors are local values only; permission
policy persistence/serialization belongs to the relevant policy service.

Colocated tests cover every observed set, bound/default locale, exact backend
route/audience/headers/body, declared response statuses, source/Body composition,
caller ownership, empty deletion, partial-list diagnostics, unknown fields,
authorization/malformed/error cases, pre-cancellation and no mutation replay.
Cloud/mode cases prove configured routing/binding, not live sovereign eligibility.
Six schema-named service/local report files retain exact mappings and tested
revision in CI; no payloads, identities, URLs or raw errors enter those reports.

## Actual tenant observations and remaining acceptance

The new SDK List method was executed with the existing dedicated lab SPN:

| Actual operation | Observed outcome |
| --- | --- |
| Team-template List, client-secret | HTTP 403, classified forbidden; typed envelope retained |
| Team-template List, certificate | HTTP 403, classified forbidden; typed envelope retained |

These are actual authorization observations, not successful management CRUD.
No template, team, policy, Graph grant or directory role was changed by these
checks. The requested tenant-scoped Teams Administrator role remains pending
after automatic review rejected the earlier broader privilege change.

`TestLabTeamTemplateCRUD` prepares expanded and JSON Body variants. Each journals
creation intent, creates an isolated named tenant template, resolves and verifies
its opaque reference, checks non-default create/update read-back and General,
lists, deletes and verifies absence. Optional pinned PowerShell reads compare
fields/channel order after create and update. Cleanup verifies ownership and
retains unresolved state, including an unidentified creation with an empty list.
An explicit bound recovery entry point is included. No real team or permission
policy is created, assigned or edited by those scenarios.

Those live mutation/parity scenarios are **prepared, not run**. Tenant PUT
omission behavior, persistence/propagation and available cleanup must be observed
before delivery/live coverage is promoted. The SDK exposes the evidenced PUT
wire behavior rather than claiming undocumented PATCH retention semantics.
