# CORE-26 tenant acceptance

Status: prepared, not run. This does not qualify live membership CRUD or close
[CORE-26](https://github.com/deploymenttheory/go-sdk-msteams/issues/182).
The dedicated SPN currently lacks the required management access. No permission
or role is granted by these tests, and credentials remain outside the repository.

The test creates a fresh disposable native team with two existing lab owners,
then uses a third existing lab user for membership CRUD and role changes. It
cannot accept an existing team as a mutation target. Team creation provisions
backing resources; all three independent lab opt-ins are required. Guest-specific
behavior, interactive authentication and Teams-client convergence remain separate
qualification scenarios.

Provide a private regular JSON file, selected by
`MSTEAMS_LAB_MEMBERSHIP_USERS_FILE`, with these keys:

```json
{
  "tenant_id": "<lab tenant ID>",
  "owner_user_id": "<existing licensed lab user ID>",
  "second_owner_user_id": "<different existing licensed lab user ID>",
  "member_user_id": "<third existing lab user ID>"
}
```

The tenant must match the explicitly selected credential identity. All user IDs
must be distinct directory UUIDs. This file and the existing lab identity,
certificate/key or secret files must have private permissions (0600 on Unix).
The journal directory must be private (0700 on Unix); the journal binds recovery
to the original tenant/cloud/app. The existing client helper defaults to
certificate authentication; `MSTEAMS_LAB_AUTH_METHOD=client-secret` selects the
other supported unattended lab method. Both are application authentication.

Only after appropriate consent and lab fixtures have been provisioned, configure
the explicit private paths and run:

```sh
MSTEAMS_LAB_READ_ONLY=1 \
MSTEAMS_LAB_MUTATIONS=1 \
MSTEAMS_LAB_CONSEQUENTIAL=1 \
MSTEAMS_LAB_TIMEOUT_SECONDS=300 \
go test -count=1 -run '^TestLabTeamMembershipCRUD$' ./tests/integration/lab
```

The command does not create a user, license an account, grant Graph consent or
assign any directory role. Normal CI skips before loading credentials. Access
must cover native team create/read/status, group read/delete, membership
create/read/update/delete/list, and backing member/owner references. Read-denial
probes are not successful access qualification.

The scenario persists creation intent before POST and any known team ID before
waiting. It checks native create/get/list, promotion/demotion through read and
list snapshots, and deletion through received 404 plus complete list absence.
It then checks composed owner add/list, owner-only removal preserving membership,
and member removal. Test-owned read polling is bounded by the overall timeout,
never retries mutations and stops immediately on received read errors. It
qualifies only the observations made during a successful run.

Cleanup uses a fresh bounded context. It requires positive reads of both team
and backing group with matching IDs, exact run-specific names and Unified group
type before deleting that owned backing group. Unknown creation IDs can be
resolved only through a complete, unique positive name match. Empty, denied,
incomplete or ambiguous discovery is unresolved; it is not proof of absence.
Consequently a failed/incomplete provisioning or a crash after deletion can
leave a journal requiring operator reconciliation. Group deletion uses the
normal Graph lifecycle, not permanent deleted-item purging.

To recover an interrupted run, stop the original process, select its private
file in `MSTEAMS_LAB_RECOVERY_JOURNAL`, retain the original credential binding and
enable the same three lab flags, then run only
`TestLabRecoverTeamMembership`. It will not mutate unrelated journal kinds or
delete a resource after failed ownership checks. Logs contain stage outcomes and
the private journal basename, never tenant/user IDs, tokens or response bodies.

Credential-free root-client tests exercise positive, denied, missing, ambiguous,
incomplete and mismatched ownership paths and ensure no delete occurs without
both ownership reads. Private user-file tests check binding and distinct IDs.
These offline checks do not substitute for a successful lab run.

## Optional pinned PowerShell comparison

Add `MSTEAMS_LAB_POWERSHELL_ORACLE=1` and an absolute
`MSTEAMS_LAB_MODULE_PATH` for the extracted pinned 8.1.0 module. The shared
[read-oracle procedure](team-lifecycle-lab-acceptance.md#optional-pinned-powershell-read-comparison)
describes certificate authentication, module guards and private bounded process
output. Configuration is checked before fixture writes. The Go SDK may use its
selected secret/certificate method; PowerShell uses the same identity's certificate.

After composed owner addition, owner-only removal and member removal, the Go test
requires a complete team-user list, then compares it with `Get-TeamUser` on the
same disposable team. The oracle first positively checks its ID and run-specific
name through `Get-Team`. Directory IDs and roles compare without case, nullable
user/name fields remain distinct, and collection order is ignored. Duplicate,
missing, unexpected or invalid identities fail comparison. No tenant/user IDs or
raw output are logged or uploaded.

PowerShell performs reads only. This resulting-state check does not execute
Add-TeamUser/Remove-TeamUser mutations or claim identical mutation traffic. Guest,
service-principal, delegated and Teams-client convergence scenarios remain
separate. The collection comparison was tested offline; no live oracle has run.
