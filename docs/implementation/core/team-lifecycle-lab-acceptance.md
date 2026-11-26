# Team lifecycle lab acceptance

Status: prepared, not executed. Current dedicated-identity team/group prerequisite
reads returned HTTP 403. Successful CRUD, archive transitions, stored photos and
PowerShell/Go conformance have not been established. This procedure grants no
permissions and creates no accounts.

`TestAccTeamLifecycle` runs the group-first creation and existing-group migration
paths against two new disposable fixtures. Migration creates its own group first;
no caller-supplied existing resource ID can become a mutation target. Each path
checks paired read, complete bounded tenant/user enriched lists, group/settings
update readback, confirmed archive/unarchive, a valid generated JPEG upload and
photo readback, then deletion of the backing group and typed 404s for both group
and team. The photo check requires a nonempty valid JPEG; it does not assume Graph
returns the original encoding.

The selected identity needs effective access for every constituent operation,
and the explicit owner must be an existing same-tenant Teams-enabled lab user.
Application authentication always supplies that owner. Provisioning, licensing,
education templates, delegated flows, restricted clouds and SharePoint completion
are separate prerequisites or scenarios. A directory role alone does not imply
all required Graph application permissions. Nothing here assigns either.

Use an existing private directory outside the repository, with a mode-0600 identity file
and a separate mode-0600 owner file. The identity JSON contains `tenant_id`,
`client_id`, `client_secret`, `certificate_file` and `private_key_file`; the latter
two identify private regular files for the PEM certificate and key. It is the
same format used by the other prepared lab scenarios. The owner file contains:

```json
{"tenant_id":"<selected tenant>","owner_user_id":"<existing licensed lab user>"}
```

All three enablement flags are required before credentials are loaded. A regular
CI run skips these entry points without reading the identity/owner files.
When the access prerequisites and test-owned mutation scope are approved, run:

```sh
MSTEAMS_LAB_READ_ONLY=1 \
MSTEAMS_LAB_MUTATIONS=1 \
MSTEAMS_LAB_CONSEQUENTIAL=1 \
MSTEAMS_LAB_IDENTITY_FILE='/private/lab/identity.json' \
MSTEAMS_LAB_TEAM_OWNER_FILE='/private/lab/team-owner.json' \
MSTEAMS_LAB_JOURNAL_DIRECTORY='/private/lab/journals' \
MSTEAMS_LAB_TIMEOUT_SECONDS=300 \
go test -count=1 -timeout=20m -run '^TestAccTeamLifecycle$' -v ./tests/integration/lab/graph
```

The explicit identity helper defaults to certificate authentication; select
`MSTEAMS_LAB_AUTH_METHOD=client-secret` for the other prepared application path.
It does not use ambient credentials. A journal records the creation intent before
any write and every returned known owned identity before later test stages.
Group-first creation requests retention so the journal owns test cleanup. A crash
inside creation can leave the intent without its ID; recovery must rediscover a
complete unique positive group match with the exact random run name and alias.

Cleanup uses a fresh bounded context while the SDK remains open. It requires
positive paired snapshots, both IDs checked by the management service, the exact
fixture name/alias and a Unified group before deleting. Zero results, denied or
incomplete discovery, a missing snapshot, failed teamification, or ambiguity stay
unresolved. A 404 is never guessed to mean creation did not happen. Unfinished
provisioning or a crash after deletion but before journal persistence can require
operator reconciliation. Keep the private journal; do not delete an arbitrary
existing group to make the test pass.

After stopping the original run, recovery can reopen only the explicitly selected
journal bound to the same tenant/cloud/application:

```sh
# Keep the same three flags, identity, journal directory and timeout as above.
MSTEAMS_LAB_RECOVERY_JOURNAL='/private/lab/journals/<reported basename>' \
go test -count=1 -timeout=10m -run '^TestAccRecoverTeamLifecycle$' -v ./tests/integration/lab/graph
```

The recovery handler recognizes only its own fixture kind. Other scenario entries
remain unresolved. Logs contain static outcomes and a local journal basename;
never publish its contents, tenant IDs, tokens, credentials or response bodies.
The tests currently check direct readbacks strictly; replication-related failures
are failures to investigate, not permission to repeat a mutation or declare parity.

Without the additional oracle flag, these Go-only smoke scenarios do not run
PowerShell comparisons. A successful
future run must retain its actual tested revision and sanitized outcomes; it must
not promote unexecuted education, delegated, cloud or PowerShell scenarios.
Ledger links identify prepared entry points, while live status remains `not-run`.

Offline tests exercise the cleanup handler through the real SDK with no sockets,
including denied/incomplete/ambiguous discovery, wrong names/aliases, non-Unified
groups, local validation, a denied delete, and read-only deletion observation.
They establish cleanup boundaries, not tenant acceptance.

Primary photo [read](https://learn.microsoft.com/en-us/graph/api/profilephoto-get?view=graph-rest-1.0)
and [update](https://learn.microsoft.com/en-us/graph/api/profilephoto-update?view=graph-rest-1.0)
documentation reviewed on 2026-11-21 supports the separate group-photo readback
route and permissions. See [tests](../../../tests/integration/lab) for
native/composed contract scope and observed lab limits.

## Optional pinned PowerShell read comparison

Add `MSTEAMS_LAB_POWERSHELL_ORACLE=1` and an absolute
`MSTEAMS_LAB_MODULE_PATH` pointing to the extracted 8.1.0 module to the same live
command. PowerShell must be installed. The helper validates this optional
configuration before fixture writes; the script verifies module version and the
pinned TeamsCmdlets assembly hash before connecting. Its certificate connection
uses the same explicit tenant/application identity, including when Go uses that
identity's secret. The identity file must therefore retain certificate/key paths.
No ambient signed-in user or browser identity is substituted.

After creation, update and each confirmed archive transition, PowerShell only
runs `Get-Team`: identity lookup, the unique fixture display-name search, and
that owner's filtered joined-team list. It compares selected IDs, names,
description, alias, classification, visibility, archive state and AllowGiphy
against the paired Go read. Nullable values stay distinct. The Go test drives
all writes and journalled cleanup. This is a resulting-state comparison; it does
not execute PowerShell's creation/update/archive/picture/delete commands or prove
their mutation sequences equivalent.

Selectors travel through stdin to a fixed script, never interpolated command
text. Child stderr is discarded, stdout is bounded to 64 KiB, and unexpected,
partial or mixed output fails strict decoding. Only fixed outcome text is logged;
returned projections remain private transient data. No read-oracle output is a CI
artifact. Ordinary Go CI neither launches PowerShell nor reads credentials.
The script was syntax-checked, its disabled guard was executed, and its selected
projection was checked against the pinned CLR model offline. These checks do not
establish successful tenant reads. A denied or mismatched oracle fails the live
scenario; it never becomes a skipped/pass result or a reason to repeat a write.

## CORE-90 cleanup completion ordering

The cleanup handler now verifies both backing-group and team 404s before returning
success to Journal.Recover. A delete receipt alone cannot mark the journal entry
cleaned. A later read denial or cancellation retains recovery work; an interrupted
run whose delete already completed can independently reconcile both absences
without another deletion. Existing positive ownership checks still precede any
new delete. The CORE journal regression covers denial followed by recovery; live
Graph execution still requires the explicitly configured owner and permissions.
