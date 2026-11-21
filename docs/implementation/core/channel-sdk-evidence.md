# CORE-21 channel SDK evidence

Status: implemented candidate with offline contracts; tenant qualification and
full MicrosoftTeams parity remain open in issue #181. The existing dedicated
lab identity authenticates, but has only Organization.Read.All and no directory
role. Teams policy reads are 403; that does not establish Graph channel access.
Automatic approval review rejected adding Teams Administrator without exact-scope
authorization. No role/consent was changed and no channel lab write was attempted.

The service follows the pinned beta CRUD/list route families in
[CORE-01 evidence](channel-contract-evidence.md) and
[endpoint manifest](../../../contracts/endpoints/8.1.0/core.json).
Individual channel reads, native lowercase creation values and exact 201/202/204
statuses use current Microsoft documentation, separately from the pinned client's
uppercase values, broad-success helper and first-page name lookup. Review date:
2026-11-21. Synthetic HTTP replies do not establish actual service acceptance.

The root service owns native CRUD, bounded one-page/complete hosted/all/incoming
listing, explicit owner resolution and complete unique-name compositions. A
shared-channel receipt retains a safe beta operation handle and the original
response, including when handle extraction fails. Server URLs are never inferred
to be the channel itself. Future read fields and null/absent distinctions survive.
Name and owner compositions use one overall session/caller deadline.

`TestUnitNativeCRUDContracts`, `TestUnitCollectionsAndPaging`,
`TestUnitFailuresPreserveStatusAndDoNotReplayWrites`,
`TestUnitSharedCreationAndOperation`, `TestUnitOwnerResolutionComposition` and
`TestUnitNameCompositionsUseCompleteUniqueLookup` exercise the real root client with
synthetic credentials and an in-memory HTTP transport. Exact method/escaped
path/query/selected headers/body/status and selected results are compared by
the shared contract runner. Cases cover all six command mappings, genuine
ID-based CRUD, omission/empty values, list bounds and untrusted origins,
403/429/503 mutation failures without replay, malformed reads, beta shared-create
polling after 31 seconds, and unique later-page name resolution. Additional
local tests check zero-I/O validation, cancellation, unsafe operation locations
and passive host-team projection. Compiling examples perform no tenant calls.

Only `channel-service.json` structural reports are retained by CI; they contain
the tested revision and case outcomes, `coverage_claim:false` and `live:not-run`.
The command ledger remains `in-progress`/`metadata-only`/`not-run`. No unknown
runtime input or blocked lab case becomes a fabricated pass or generic stub.

The gated `TestAccChannelCRUD` is prepared for an owned parent team and all
three channel types, with private journal recovery; it has not run. See
[lab procedure](channel-lab-acceptance.md).

Remaining qualification includes real standard/private/shared CRUD and cleanup,
private owner and external prerequisites, actual shared operation Location/error
shapes, incoming/host projection, restricted clouds and PowerShell resulting-state
comparison. The deliberate name-lookup/payload differences are documented in
[API notes](../../../msteams/graph_api/channel/api_docs.md).

Primary sources for current supporting behavior:
[channel read](https://learn.microsoft.com/en-us/graph/api/channel-get?view=graph-rest-beta),
[creation](https://learn.microsoft.com/en-us/graph/api/channel-post?view=graph-rest-beta),
[update](https://learn.microsoft.com/en-us/graph/api/channel-patch?view=graph-rest-beta),
[deletion](https://learn.microsoft.com/en-us/graph/api/channel-delete?view=graph-rest-beta),
[channel model](https://learn.microsoft.com/en-us/graph/api/resources/channel?view=graph-rest-beta).
