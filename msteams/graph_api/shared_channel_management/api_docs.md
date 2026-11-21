# Shared-channel command compositions

`sdk.GraphAPI.SharedChannelManagement` composes independently usable resource
services for four MicrosoftTeams 8.1.0 commands, all `__AllParameterSets`:

| Command | Go operation | Composition |
|---|---|---|
| Get-AssociatedTeam | ListAssociatedTeams | Empty user selects explicit delegated view; UUID selects user view; UPN resolves before user view |
| Get-SharedWithTeam | ListSharedWithTeams | Empty SharedTeamID collects all sharing pages; supplied UUID performs one Get and preserves collection output |
| Get-SharedWithTeamUser | ListAllowedUsers | Complete allowedMembers traversal, local role projection/filtering and bounded directory enrichment |
| Remove-SharedWithTeam | UnshareWithTeam | One native relationship Delete |

Every operation accepts context and shares the client's lifetime and configured
Graph cloud. Construction is local. Returned responses stay available for
diagnostics. The selected-sharing Get branch ignores traversal bounds because
it performs no traversal. Native relationship operations stay available for
callers needing unprojected records or their own compositions.

Associated user resolution uses the existing shared resolver: a UUID requires no
directory lookup; UPN tries the canonical v1.0 direct read, and only an HTTP 404
allows complete unique mail lookup. Ambiguity, bounded traversal or 202 cannot
establish a usable identity. This deliberately retains the established SDK policy
instead of the pinned historical `myorganization` alias/first-match behavior.
An empty user in application mode fails locally, without another identity/prompt.

Allowed-user filtering accepts Owner/Member case-insensitively, matching the
pinned parameter set. Unfiltered output preserves native future/guest data;
owner takes precedence over guest, and stable output groups owners, members,
then guests. Filtering precedes UPN lookup. Distinct selected UUIDs are queried
in groups of at most ten through the shared supporting directory helper, with
bounded opaque paging per group. No `$batch` or duplicate auth/retry policy is
introduced. A returned null/absent UPN clears the initial projection; a missing
directory record retains the original UPN/email. Foreign tenant IDs remain
unchanged. Duplicate/unrequested directory results fail rather than inventing
identity correspondence. On enrichment failure, native records, projected data
and names from earlier fully validated groups remain available with
`Complete=false`; partial data never claims complete parity.

Permissions are the union of each native endpoint and any directory lookup:
associated-team Team.ReadBasic.All, sharing reads ChannelMember.Read.All (or
application resource-specific .Group), and directory application User.Read.All
for enrichment. Unshare uses ChannelMember.ReadWrite.All (or application .Group
with resource-specific consent). Lookup roles/cloud availability are not inferred
from a successful token. Detailed sources/restrictions are in
[unit tests](service_test.go).
Offline cases are source-informed synthetic traffic; tenant and pinned-command
comparisons remain unrun. Relationship Create/Update contracts remain open.
