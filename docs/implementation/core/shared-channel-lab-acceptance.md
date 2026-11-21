# Shared-channel lab acceptance

Status: prepared, not run. No relationship was created/read/deleted during this
implementation. Exact tenant role/Graph approval requests remain pending; token
acquisition alone is not authorization evidence. Existing user-provided Teams
browser login does not authorize substituting an identity to bypass these gates.

`TestLabSharedChannelReadOnly` requires `MSTEAMS_LAB_READ_ONLY=1`, the existing
private `MSTEAMS_LAB_IDENTITY_FILE`, and a 0600 non-symlink
`MSTEAMS_LAB_SHARED_CHANNEL_FILE`. With the flag absent it skips before reading
credentials or fixture selectors. The fixture JSON has `tenant_id`,
`host_team_id`, `channel_id`, `shared_team_id`, and optional
`associated_user_id`. UUID team/user selectors are explicit, channel ID stays
opaque, and fixture tenant must match the selected identity. Keep this file in
the private local lab directory; do not commit or attach its values to CI.

Prepare an existing shared channel and recipient-team relationship using an
authorized, independently verified workflow. Cross-tenant policy, membership,
cloud availability and sharing consent are environmental prerequisites, not
inferred from IDs. No automatic browser setup or speculative POST/PATCH exists.

The case reads the selected native relationship, a complete sharing list,
composed allowed-user output and optional explicit-user associated view, bounded
by the lab timeout and page/item limits. It asserts selected relationship
presence rather than treating an empty result as a successful visibility test.
It performs no write and needs no cleanup. An empty allowed-user/associated
result still does not qualify positive membership visibility; use a known
populated fixture for that scenario.

Optional `MSTEAMS_LAB_POWERSHELL_ORACLE=1` plus absolute
`MSTEAMS_LAB_MODULE_PATH` enables the shared read-only child oracle. It verifies
the pinned 8.1.0 assembly hash and uses certificate app auth without an interactive
prompt. It executes only Get-SharedWithTeam (list and selected),
Get-SharedWithTeamUser and optional explicit-user Get-AssociatedTeam. Bounded
stdout stays transient, stderr is discarded, and no IDs/tokens/server text are
retained in test logs/artifacts. Reports reject mixed/incomplete output. Comparison
retains tenant, nullable fields, host false versus unknown and unique identities;
it does not assume server ordering. The child never calls Remove-SharedWithTeam.

Ordinary checks run with all lab flags unset. To run only this qualified probe:

```sh
go test -count=1 -run '^TestLabSharedChannelReadOnly$' ./tests/integration/lab
```

Remaining qualification: positive/denied visibility for host and recipient,
populated allowed-user and cross-tenant projection/enrichment, delegated `/me`,
real continuation/propagation, and one unshare of an explicitly owned disposable
fixture with journaled known parent identities and complete absence read-back.
Unshare may remove access and has no settled inverse write; never use a production
relationship as cleanup. Establish the setup/restore contract before automating
that destructive acceptance. Sharing Create/Update must have evidenced fields,
responses, permissions and workflow before claiming true relationship CRUD.
