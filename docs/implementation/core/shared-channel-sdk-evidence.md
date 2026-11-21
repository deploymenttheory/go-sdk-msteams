# CORE-24 shared-channel SDK evidence

Implementation candidate for issue #190. Four pinned commands have native
resource/composition mappings, with offline checks. This is not full CRUD or
tenant qualification: sharing relationship Create/Update contracts, live
authorization, prerequisites and PowerShell comparison remain open. The ledger
stays in-progress/metadata-only/not-run; the issue stays open.

| Baseline command / set | Native supporting services | Composition |
|---|---|---|
| Get-AssociatedTeam / __AllParameterSets | AssociatedTeams.ListMine/ListForUser; UserResolver.Resolve | SharedChannelManagement.ListAssociatedTeams |
| Get-SharedWithTeam / __AllParameterSets | ChannelShares.ListAll/GetByID | SharedChannelManagement.ListSharedWithTeams |
| Get-SharedWithTeamUser / __AllParameterSets | ChannelShares.ListAllowedMembers; UserResolver.LookupPrincipalNames | SharedChannelManagement.ListAllowedUsers |
| Remove-SharedWithTeam / __AllParameterSets | ChannelShares.DeleteByID | SharedChannelManagement.UnshareWithTeam |

Route/projection sources are the pinned MicrosoftTeams 8.1.0 binary and synthetic
serialization, identified by immutable package/assembly hashes and member names
in [CORE-01 evidence](channel-contract-evidence.md). Current documented status,
permissions, cloud availability and shared-channel restrictions remain separately
labelled there. The [mutation investigation](shared-channel-mutation-evidence.md)
traces first-party generated POST/PATCH request builders without treating their
presence as proof of accepted service mutations. No reverse-engineered source is
vendored. A new-channel create is not a substitute for sharing an existing one.

Native resources retain host, shared-with and returned tenant identities,
nullable/false/future fields and opaque channel/access-record IDs. Allowed members
are read from the exact sharing relationship, independently of the direct channel
or target team's roster. Unshare sends one relationship Delete and never deletes
a channel/team or retries an ambiguous mutation.

Compositions preserve optional single versus collection output, delegated-only
current-user selection, complete associated views, role precedence/order and
filter-before-enrichment. The established directory resolver deliberately retains
canonical unique mail fallback instead of the pinned historical alias/first match.
The shared UPN helper retains ten-distinct-ID grouping, complete bounded paging,
returned-null versus missing-user semantics, and only fully validated earlier
groups on later failure. Existing channel-user tests pass through that extraction.
No generic batch, new dependency, transport/auth loop or root Go file is added.

Tests use the real root client, synthetic credentials and an in-memory transport.
They assert method, escaped beta path, query, Graph audience, Accept header and
body absence, plus response/projection/status/error behavior. They cover native
read/list/unshare/allowed paths, opaque continuation/partial traversal, invalid
selectors and options, absent/null/future fields, 403 and unreplayed 429/503,
explicit-user versus delegated-me selection, canonical UPN fallback, stable role
projection/filtering, enrichment failures, nil/canceled context, nil/closed
services and wrong backend. Compiling examples perform no HTTP.

CI retains only the exact structural `shared-channel-service.json` and
`shared-channel-management-service.json` report names, with revision and
coverage_claim:false/live:not-run. Native retained cases exercise seven scenarios;
composition reports exercise fourteen scenarios and all four command/set mappings.
Additional negative/local/paging cases run without claiming those reports cover
the whole command contract. None is captured tenant traffic or exported-cmdlet
execution. Revision/platform results belong to the implementation PR.

The prepared [read-only lab acceptance](shared-channel-lab-acceptance.md) requires
explicit private selectors and may compare pinned PowerShell observations. It
does not establish sharing, delete relationships or alter directory grants. The
SPN exists; exact privilege requests remain pending. No live scenario has run.
The oracle's team/user conversion functions were extracted from its AST and
checked against synthetic instances of the actual hash-pinned CLR models.
Parsing and the disabled guard also passed without importing the module,
executing a management cmdlet or contacting a tenant. This validates projection
code, not an exported-command or live PowerShell/Go comparison.
