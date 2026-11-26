# Tenant model foundation and remaining public projection

This is an implementation foundation for the same Get-CsTenant work in #191,
not a contract qualification decision. The fourteen approved EXT contracts,
including Find-CsGroup, remain unchanged. Full public helper/date/associative
conversion and own TRPS/legacy dispatch mapping remain required before moving
to another family or claiming Get-CsTenant complete.

The [portable own source/data packet](evidence/tenant-model-foundation-source.json)
binds the default public proxy/private type, Modern helper path, generated
Tenant and seven nested types, actual data getters and nine own ordered/type
profiles. The selected public postprocessing probe executes only the three
Select-Object/Add-Member/Write-Output statements against a controlled DTO; no
Teams management cmdlet lifecycle, auth or HTTP is invoked.

## Native full-wire view

Tenant now types all 35 known wire fields, adding announcements, partnerships,
company tags, country, pool FQDN, directory/name-recording flags, pools, two
Int32 counts, state, sync info, TNM account ID and whenChanged. The nested
partnership's four fields and sync object's six fields are typed, along with
the previously omitted assigned-plan serviceInstance. Opaque tenant/plan/context
IDs remain strings; the Group public GUID conversion is not generalized.

This unreleased API widens SIPDomains to []*string and the two timestamp maps
to map[string]json.RawMessage. Null/nonstring string-array positions and
null/nonobject nested-model positions survive. Formally object-valued maps
retain original numbers, booleans, nulls, arrays and objects rather than reject
non-string values or lose large integer lexemes. Typed fields and RawFields
have independent owned storage. Known fields use their own exact generated
case/type/duplicate/range profile; unrelated directory/configuration decoders
are unchanged. The common JSON-node scanner is reused only after own source
proves the same parser path, without altering Find's code.

Existing timestamp string fields and createdDateTime/isInGracePeriod wire
enrichments remain explicit native views. They preserve wire text, not CLR
DateTime or public PowerShell output. This distinction is operation-critical:
source InvariantCulture TryParse with AdjustToUniversal (16) turns offset ISO
into UTC, preserves naive clocks as Unspecified, and maps malformed strings to
null. A raw wire string cannot silently stand in for that projection.

## Same-command next phase

The actual public helper selects all 35 CLR properties and flattens only its two
associative timestamp objects through AdditionalProperties. Data probes prove
that those source dictionaries omit null object entries, convert numeric values
through CLR scalar types (1.20 becomes Single), and convert a nested array null
to the string "null". Those are source observations, not permissions to coerce
or discard the native lossless wire view. A separately inspectable public
projection must map its own source date/associative conversion before being
called complete. The prior Json.NET DateTime helper cannot be copied blindly:
its offset-to-local rule differs from this AdjustToUniversal path.

The public proxy wraps a private GetCsTenant/TrpsPSCmdlet. Its descriptor selects
Get-CsTenantPoint for the Modern path, which calls private Get-CsTenantObou_Get
with Extended and then performs the postprocessing above. Own dispatch/header
closure remains separate from common TRPS implementation evidence; no universal
Modern selection, retired legacy transport or public-parameter behavior is
inferred from the descriptor alone. No Get-CsTenant flags or live state change
in this foundation.

## Executed foundation checks

At source checkpoint `863d41c`, the five selected SDK/root/lab packages passed
131 test/subtest events with no test failures or skips; the administration
composition package has no test files. The same package selection passed
configured lint with zero issues using Go 1.27.4. The root review independently
reproduced the exact nine-case wire-profile packet and checked all fourteen
embedded foundation/supplement producer hashes and two shared handoff references.
See the [revision-bound execution record](../../../contracts/coverage/evidence/ext90-tenant-wire-execution.json).

This validates the bounded wire foundation. Public projection, dispatch and
full-command qualification retain the separate limitations above.
