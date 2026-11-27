# Voice applications policy

`TeamsAdminAPI.TeamsVoiceApplicationsPolicies` implements the eight public pinned 8.1.0
parameter sets with caller context, complete owned RawFields/RawJSON snapshots,
local Identity/Tag wildcard selection and Global-first lists. Construction has no
network or auth side effects. 34 nullable settings: 25 booleans and nine strings for description, queue monitoring and metrics permissions. Source strings stay opaque; the inspected setters do not impose ranges or enums.

Get/New/Set/Remove use compiled PolicyRP; Grant uses modern ConfigAPI discovery
and its distinct headers. Both providers share one injected identity-bound scope.
CreateByIdentity sends bound fields once. PatchByIdentity sends one sparse PATCH;
UpdateByIdentity composes Set's preparation GET/sparse PATCH and retains partial
receipts. DeleteByIdentity sends the compiled Identity/optional Force body. Scalar
policy.Field values distinguish omission/null/false/empty. Force alone is not an
update. New's display-only Global read does not become inherited settings.

Assign selects exactly one user/tenant/group; individual native grant methods
remain available. Empty PolicyName clears, and the argument is required despite
optional public metadata. Group rank is signed optional Int32: the wrapper omits
zero; native AssignToGroup can explicitly send zero. Public Force is absent; the
lower tenant helper exposes its evidenced native force field. HTTP202 acceptance
is separate from HTTP204 and is not propagation proof. PassThru concerns
PowerShell output. No write replays or polls implicitly.

Both app-auth modes passed Global Get/List and owned unassigned CRUD with exact
sparse fields, all typed getter/omission checks, two pinned PowerShell comparisons,
list/delete and independent GET404/list-absence cleanup. Nine permanent sanitized
actual responses supply unit model/read/CRUD replay. Source/fault vectors are
separate from live server successes. All raw metadata is retained per response;
internal XML Key differences are not immutable state across separate reads.

The existing lab Teams Administrator application and secrets suffice. Successful
Grant persistence/clear, other setting persistence, delegated tenant flows and
other clouds remain unqualified. No existing policy or assignment changed.
See the [testing guide](../../../docs/guides/testing.md).
