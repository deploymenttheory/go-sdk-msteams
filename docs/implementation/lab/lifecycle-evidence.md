# LAB-01 acceptance lifecycle

Implementation for [LAB-01](https://github.com/deploymenttheory/go-sdk-msteams/issues/59).
The internal `lab` package provides explicit scenario enablement, read-only
capability observations, unique fixture names and private recovery journals.
Resource-specific acceptance suites supply their own CRUD and cleanup handlers.

## Configuration and preflight

Ordinary `go test ./...` discovers the live preflight test and skips it before
reading credentials. Set `MSTEAMS_LAB_READ_ONLY=1` to enable it, and provide
`MSTEAMS_LAB_IDENTITY_FILE` pointing to the existing private lab identity. No
environment is read by SDK construction or the harness until explicitly loaded.

`MSTEAMS_LAB_MUTATIONS=1` additionally permits mutation scenarios and requires
`MSTEAMS_LAB_JOURNAL_DIRECTORY`. `MSTEAMS_LAB_CONSEQUENTIAL=1` adds a separate opt-in
for tenant-wide or telephone-number scenarios and requires both earlier levels.
Entry points call `Config.Require` before their scenario. These are execution
controls; credentials, tenant roles and resource availability are separate.
Flags accept empty/0/1 only. The default timeout is 90 seconds; the explicit
`MSTEAMS_LAB_TIMEOUT_SECONDS` range is 1..300.

The fixed Graph preflight sends GET organization with `$select=id` and GET users
with `$select=id&$top=1`, under caller cancellation and the bounded timeout. It
retains only fixed probe names, HTTP status and available/forbidden/unauthenticated/
unavailable/transient/error categories. It retains no response bodies, token
claims, tenant IDs, request IDs or raw errors. The report is sampled access evidence,
not a complete permission/role/license inventory. A 404 does not infer why a
resource is unavailable. Additional acceptance scenarios must state their own
capability prerequisites rather than treating token acquisition as authorization.

The dedicated commercial lab application is expected to permit organization reads
and deny user reads. The live test checks that deliberately limited access profile;
it never changes grants or tenant configuration.

## Durable mutation intent and cleanup

Create a journal in an operator-selected private directory **outside the checkout
and CI artifacts**. The binding input must identify tenant, cloud and identity
unambiguously; only SHA256 of that input is stored. A different binding cannot
open it for recovery. Files use mode 0600 and directories mode 0700 on POSIX;
existing public directories/files and final symlinks are rejected. The parent
directory must already exist; the harness creates only its private final directory.
On Windows,
the operator must use a directory restricted to their account by ACL; POSIX mode
bits are not an ACL guarantee.

Each run gets a random fixture prefix. `FixtureName` adds a validated label;
use different labels for different resources within one run. Before creating,
call `PlanCreate` with that name. Before changing shared settings, call
`PlanRestore` with the known resource ID and a minimal typed JSON snapshot.
Only after the plan persists may the suite send its mutation. Immediately record
the returned ID with `Applied`, including IDs supplied with partial failures.
Snapshot input is copied so caller changes cannot alter recovery data.

Writes serialize, fsync and close a new private temporary file before renaming it
over the journal. POSIX writes also sync the containing directory and sync the
parent after creating the private directory. Windows has no portable Go directory
flush; file data is synced there. An interrupted process retains the previous complete version
or the new complete version; temporary files are not treated as recovery journals.
This is process-interruption recovery, not a guarantee against filesystem/device
failure. A journal has one owning process. Stop the original run before opening
the file for recovery; concurrent processes must not share it.

`Recover` runs registered handlers in reverse plan order. A creation without a
recorded ID first invokes `Resolve` using its unique fixture name. A handler must
verify ownership and return `ErrUnresolved` for ambiguous matches; an empty ID
with nil error means verified absence. Successful cleanup is persisted and skipped
on later runs. Failed cleanup stays unresolved while independent entries continue.
Restoration uses the saved settings rather than guessed defaults. Handler cleanup
must be idempotent because a crash may happen after the remote cleanup succeeds
but before its local completion record is written.

The harness does not choose a resource's safe delete/restore route or automatically
delete arbitrary IDs. Those exact handlers and equivalence assertions belong to
the resource acceptance ticket. Run cleanup with a fresh bounded context when the
test context is canceled. `Close` releases handles but retains the recovery file;
it performs no network cleanup. Entry/journal formatting is redacted; ordinary
cleanup results contain kind and outcome only. Local files still contain private
IDs and settings and must never be uploaded as test evidence. Authentication
configurations, tokens and raw tenant dumps are never valid snapshots.

## Verification

Unit fixtures cover strict hierarchical enablement, read-only method/path/query,
bounded/canceled contexts, typed-nil clients/handlers, access-outcome distinctions,
unique run names, wrong-tenant recovery rejection, durable intent before a lost
creation response, reverse cleanup order, restoration failure/retry, snapshot
copying, completed-entry skipping, corrupt/closed journal rejection and private
POSIX storage. Fake resource backends exercise creation and settings restoration;
no real mutation is needed for harness validation.

On 2026-11-20 the dedicated commercial lab preflight passed: organization-read
200/available, users-read 403/forbidden. It used the existing client-secret identity
and made only those two GETs. The output contains only probe/status/outcome; no
tenant identifier or response payload is included. The denied read confirms the
app's deliberately limited access, not a general Teams capability failure.

The live preflight is a separate opt-in read-only check. No command ledger row is
promoted to resource implementation or full management parity by this harness.
