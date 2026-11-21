# POL-22 assignment migration implementation progress

Issue #95 remains open for qualification. Native methods now cover its six commands
and exact parameter sets through TeamsAdminAPI.AssignmentMigrations. Implementation
is in-progress and contract metadata-only: no full parity or successful tenant
qualification is claimed. Typed jobs/states, recommendations, nested conflicts and
the lowercase request body preserve received wire presence and future fields.

[Pinned source evidence](assignment-migration-evidence.md) and the existing
policy-model observations establish DTO names/types. Tests consume the one request
casing and eight state observations, preserve omitted/null/zero/future and nested
data, keep original timestamp text, and avoid materializing absent CLR defaults.
Unknown string/Int32 states remain available without an invented terminal-state
machine, automatic completion or tenant-convergence claim. Model race tests,
build/vet and scoped lint pass for the model checkpoint; current live attempts are
recorded below. No mutation ran.

Native operations are Create/List/GetByID/ClearDirectAssignments,
recommendations by policy type/name, and conflict read. Clear is an explicit POST
action, not Delete; no job Update/Delete is established. Threshold omission must
select the string default 500 while preserving an explicitly supplied string.
AllowNonEmptyGroup is always a True/False query Boolean; request keys remain
authority/policytype/policyname/groupid.

These GPA routes use the compiled direct transport, without generated forest
discovery. The federation sender is extracted into shared internal policytransport
with separately encoded query values; policyrp retains its resource whitelist. GPA path
segments and query values must be encoded separately; the source's raw concatenation
does not permit URI forwarding or injection. The six cmdlets' malformed/non-2xx
presentation quirks must become native HTTP/decode errors, not successful empties.

The [new offline capture](../../../tools/contracts/capture-assignment-migration.ps1)
checks pinned Core/PolicyRp/host hashes and calls the actual compiled
CommonHttpHelperRequest.SendGPARequests with source-derived safe suffixes, a
Newtonsoft-serialized request DTO and synthetic InvocationInfo. Fake host refuses
credentials; cached version/client/host are replaced/restored. The only sender is
in memory. MPACmdlet is inserted by the adapter from factory source. CmdletInfo is
metadata only; no command constructor/ProcessRecord, response DTO decoding, auth,
discovery or tenant request runs. Its 15 observations capture requests and raw HTTP
pass-through, not command-level error presentation. Go CI consumes these goldens.

Colocated root-client tests compare both synthetic modes, actual path/query/body/
headers, omitted/empty thresholds, single-attempt submissions, error status/redaction,
malformed accepted bodies, GUID normalization and validation before token acquisition.
Nil service/context, wrong backend and cancellation are covered. The existing header
spy includes GPA and confirms fresh correlations without forest bootstrap.
HTTP errors retain their actual status rather than cmdlet output strings. Empty or
malformed successful bodies fail decoding with response retained; JSON null stays nil.
No automatic wait, terminal-state inference, submission replay or rollback is added.

TestLabAssignmentMigrationReadOnly ran using dedicated-SPN secret and certificate;
both status-list requests returned resource HTTP 403. GetByID and optional pinned
PowerShell comparison were not reached. Other reads need explicit fixtures and
remain unrun. No mode/feature restriction is inferred from denial. Start/cleanup
have not run and effective assignments were not changed. The [lab plan](assignment-migration-lab-acceptance.md)
records consequential prerequisites. App eligibility, feature activation, RBAC/cloud
availability, binding/coercion details and successful live behavior remain unqualified.
