# CORE-20 channel membership SDK evidence

Status: implementation candidate with offline verification. Issue #189 stays
open for tenant and full PowerShell qualification. The dedicated lab SPN exists
and authenticates, but has only Organization.Read.All and no Teams directory
role. Automatic approval review rejected assigning tenant-wide Teams
Administrator without exact privilege authorization. Separate exact-role and
seven-Graph-grant requests remain pending. No grants or tenant membership
mutations were performed by this implementation work.

`GraphAPI.ChannelMembers` delivers native beta CRUD/list, bounded paging,
owner promotion/demotion and complete tenant-aware opaque identity resolution.
`GraphAPI.ChannelUserManagement` composes unique channel lookup, directory/roster
resolution, add/promote/demote/delete selection and complete enriched user
projection for all three pinned command/parameter-set mappings.

The beta route families and intended command effects derive from the pinned
TeamsCmdlets assembly and [CORE-01 source evidence](channel-contract-evidence.md).
Native individual reads and exact HTTP statuses/role-array payloads use current
Microsoft REST documentation, reviewed 2026-11-21. Requests/replies in Go tests
are authored synthetic fixtures informed by those sources, not captured tenant
traffic or execution of the exported PowerShell commands. The pinned private-ID
synthesis, first-page lookup, casing/role-array/binding variants and enrichment
chunk ordering remain explicitly characterized differences in
[native API notes](../../../msteams/graph_api/channel_member/api_docs.md) and
[composition API notes](../../../msteams/graph_api/channel_user_management/api_docs.md).

Native tests exercise the real root client for CRUD/owner actions, omitted versus
empty/null roles, escaped opaque channel/member IDs, external tenant forwarding,
configured cloud/mode bindings, unknown response data, malformed/status failures,
mutation failures without replay, bounded/foreign continuation, cancellation,
closed/wrong backend and ambiguous complete user-to-membership resolution.

Composition tests compare exact request sequence/method/escaped path/query,
selected headers/body/status and selected results through the shared in-memory
contract runner. They cover UUID/UPN add, external shared creation, private and
shared promotion/demotion/deletion, null email, tenant selection, missing and
ambiguous membership, failed/incomplete preflight, 403/429/503 write failures,
standard-channel rejection, local invalid inputs, role precedence/stable order,
filter-before-enrichment, ten-ID chunking, nullable/fallback UPN behavior,
duplicate/unrequested users and denied/malformed/incomplete enrichment. A
deterministic fake-clock test proves the caller deadline spans multiple reads
and prevents the later write. It opens no socket. Examples compile without calls.

CI retains only structural `channel-member-service.json` and
`channel-user-service.json` reports: revision, case outcomes, `coverage_claim:false`
and `live:not-run`. The command ledger remains in-progress/metadata-only/not-run;
no synthetic success promotes tenant or release parity.

The gated [private-channel acceptance scenario](channel-membership-lab-acceptance.md)
is prepared but has not run. It creates its own journalled parent team, uses
three existing same-tenant users, exercises native/composed non-default readback
and cleanup, and can compare Go output with pinned read-only
Get-TeamChannelUser. Ordinary CI skips before credential loading. The oracle
was syntax-checked and its disabled guard executed, not connected to a tenant.

Outstanding qualification includes successful private/shared/external operations,
guests and tenant policies, actual PowerShell state comparison, current service
acceptance of the documented/pinned differences, restricted clouds and Teams
client convergence. No general unsupported error or placeholder substitutes
for any native CRUD operation or pinned command composition.
