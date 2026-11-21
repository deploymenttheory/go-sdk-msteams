# POL-22 assignment migration implementation progress

Issue #95 has six assigned commands. This checkpoint introduces typed migration
jobs/states, policy-type/name recommendations, nested conflicts and the lowercase
request body. Native HTTP services, root wiring and command ledger mappings are
still pending; no service stubs or unsupported-operation methods are exposed.

[Pinned source evidence](assignment-migration-evidence.md) and the existing
policy-model observations establish DTO names/types. Tests consume the one request
casing and eight state observations, preserve omitted/null/zero/future and nested
data, keep original timestamp text, and avoid materializing absent CLR defaults.
Unknown string/Int32 states remain available without an invented terminal-state
machine, automatic completion or tenant-convergence claim. Model race tests,
build/vet and scoped lint pass. No tenant call or mutation ran for this checkpoint.

Remaining native operations are Create/List/GetByID/ClearDirectAssignments,
recommendations by policy type/name, and conflict read. Clear is an explicit POST
action, not Delete; no job Update/Delete is established. Threshold omission must
select the string default 500 while preserving an explicitly supplied string.
AllowNonEmptyGroup is always a True/False query Boolean; request keys remain
authority/policytype/policyname/groupid.

These GPA routes use the compiled direct transport, without generated forest
discovery. The now-merged federation protocol can be factored into a small shared
internal direct-policy transport while retaining its resource whitelist. GPA path
segments and query values must be encoded separately; the source's raw concatenation
does not permit URI forwarding or injection. The six cmdlets' malformed/non-2xx
presentation quirks must become native HTTP/decode errors, not successful empties.

Add pinned in-memory request/dispatch observations, service positive/negative
tests, compiling examples, docs, read-only opt-in acceptance and exact ledger/report
mappings. Start/cleanup alter effective assignments and remain consequential lab
work with explicit fixtures/restore decisions. App eligibility, feature activation,
RBAC/cloud availability and successful live behavior remain unqualified.
