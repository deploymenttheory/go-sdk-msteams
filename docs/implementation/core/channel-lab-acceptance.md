# Channel lab acceptance

`TestLabChannelCRUD` is prepared and has not run against the tenant. Ordinary CI
skips before reading credentials. It grants no role/permission and creates no user.

Use the private identity, existing licensed owner, journal directory and three
explicit enablement flags documented in [team lifecycle acceptance](team-lifecycle-lab-acceptance.md),
then select `-run '^TestLabChannelCRUD$' -timeout=25m` for the lab package. The
test creates its own uniquely named team first, journals its owned group ID and
uses the existing positively verified parent-team recovery handler. No existing
team/channel ID can be supplied as a mutation target. Interrupted runs retain
the same private parent journal and use `TestLabRecoverTeamLifecycle`.

It creates standard/private/shared channels, waits for shared creation explicitly,
checks ID-based reads, non-default name/empty-description update readback, bounded
all/incoming list calls, name-based deletion and typed 404 after deletion. The
owner is explicit for private/shared. Parent-group deletion cleans any remaining
child channels. Failure/ambiguous parent creation remains unresolved; retain the
journal and never substitute an arbitrary team. No mutation is retried to hide
replication failures. Logging contains only static stage outcomes.

Requires effective group/team creation and cleanup permissions plus the native
channel permissions in [API notes](../../../msteams/graph_api/channel/api_docs.md).
Shared-channel tenant policy, owner licensing and endpoint eligibility must also
be satisfied. A Teams Administrator role does not replace Graph app consent.
Denied prerequisites fail the opted-in test and remain evidence to investigate.
These are Go-only smoke scenarios, with no PowerShell parity, external sharing,
restricted-cloud or Teams-client convergence claim.
