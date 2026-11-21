# Team-operation read-only lab acceptance

Prepared, not run. `TestAccTeamOperationReadOnly` reads an explicitly selected
existing team operation. It starts no mutation, performs no remote cancellation
and never follows the returned target location. The originating operation may
already be complete; its lifetime and access remain server decisions.

Set MSTEAMS_LAB_READ_ONLY=1, the private identity file and an explicit mode-0600
MSTEAMS_LAB_TEAM_OPERATION_FILE containing tenant_id, team_id and operation_id.
The fixture tenant must match the SDK identity's tenant. Team ID is a UUID;
operation ID is the original opaque value from an observed receipt. Use the
existing lab timeout setting and select
`-run '^TestAccTeamOperationReadOnly$' -timeout=10m` for the lab package. Ordinary
CI skips before loading credentials or fixtures. No mutation flags are required.

The helper defaults to certificate authentication; client-secret selects the
other prepared application path. The SDK reads the explicit beta status route
and requires a matching identity with an available status. Denied, expired,
missing and malformed reads fail; they do not imply an unsupported feature or
automatically change permissions. Exact team-operation permissions remain an
unresolved prerequisite; the current chat permission table is not applied.

With MSTEAMS_LAB_POWERSHELL_ORACLE=1 and an absolute pinned 8.1.0 module path, the
bounded child process checks module version/assembly hash, authenticates using
the same identity's certificate, then invokes Get-Operation with those explicit
IDs. It compares operation ID/type/status, timestamp instants, attempts and
target ID/location, preserving nullable Go fields. The operation branch is a
read-only selected-ID probe and does not require the owned-team name guard used
by the other fixture scenarios. It never creates or changes that selected team.

Comparison is a point-in-time observation. A running operation can change
between reads; a mismatch fails visibly rather than being relabelled equivalent.
The script emits only a bounded private projection; stderr is discarded and no
raw output, IDs or credentials become CI artifacts. Static outcomes are logged.
Successful read comparison would not qualify creation, failure induction,
active polling cadence, other auth modes, clouds or operation-record mutations.
