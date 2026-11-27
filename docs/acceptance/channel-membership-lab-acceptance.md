# Channel membership lab acceptance

Prepared, not run. `TestAccChannelMembershipCRUD` uses the private identity and
three distinct existing licensed lab users in the schema from
[team membership acceptance](membership-lab-acceptance.md). Set all three explicit
read/mutation/consequential flags, identity file, membership-user file, timeout
and private journal directory described there. Select
`-run '^TestAccChannelMembershipCRUD$' -timeout=25m` for the lab package.
Regular CI skips before private-file loading.

The test creates its own unique group-first parent team and journals the owned
group ID before further operations. Additional parent-team memberships are
prerequisites inside that owned fixture. It creates one private channel with an
explicit owner; no caller-supplied team/channel/membership ID can be mutated.
The shared lifecycle cleanup handler positively identifies the owned parent
team and Unified group before deletion. Interrupted runs retain the same private
journal and use TestAccRecoverTeamLifecycle. Incomplete/ambiguous creation is
not inferred absent, and no unrelated parent team is substituted.

A second existing user exercises native channel member create/read/list,
promotion, demotion and deletion. The third exercises name/user add, promotion,
demotion, list/enrichment and deletion. Updated identity/role observations must
converge across individual and complete-list reads. Deletion requires both a
typed 404 and absence from a complete list. Only reads poll; mutations are never
repeated to hide propagation. Parent cleanup owns residual memberships/channels.
Directory users, other teams and permission grants are unchanged by this test.

With `MSTEAMS_LAB_POWERSHELL_ORACLE=1` and the absolute extracted 8.1.0 module path,
the existing bounded read-oracle process validates module version/assembly hash,
authenticates with the same identity's certificate, verifies the unique parent
fixture through Get-Team, then reads Get-TeamChannelUser. It compares unique
directory IDs, selected nullable user/name fields and roles after each composed
mutation, without assuming output order. Go performs all writes. No raw oracle
output or tenant identifiers become CI artifacts. This compares resulting state;
it does not execute the PowerShell mutations or prove their wire sequences.

Effective Graph group/team creation, team/member administration, channel creation
and channel-member read/write permissions are prerequisites. Directory enrichment
needs User.Read.All for application tokens. Teams policy, owner licensing and
actual endpoint access must be satisfied; a Teams directory role alone is not
Graph consent. No grant is made by this procedure. Optional-oracle configuration
is validated before fixture writes. Static logs identify stages and journal
basename only. Shared/external, guest, cloud and Teams-client scenarios remain
separate and are not promoted by this private-channel test.
