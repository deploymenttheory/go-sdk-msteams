# Get-CsTenant public projection design and source boundaries

This is a proposed second phase of the same Get-CsTenant work in #191. The
lossless wire foundation is frozen at local `2cc064c`; no Go implementation in
this phase has been changed, compiled or published. The fourteen approved EXT
contracts and all live states remain unchanged. A complete Get-CsTenant decision
requires its public helper projection and its own dispatch/header closure.

The [new portable conversion supplement](evidence/tenant-public-projection-source.json)
contains four complete reflection/data producers and their outputs, separately
from the immutable foundation packet. These execute no management lifecycle,
authentication or HTTP. Results describe the pinned private assembly and this
PowerShell/CoreCLR host, not a successful tenant operation.

## Proposed public boundary

Keep `Tenant` as the native lossless wire model with timestamp strings and
`map[string]json.RawMessage`. It must not silently discard an input null, round
a number or reinterpret a date merely to resemble the PowerShell presentation.

Add a separate `TenantPublicView` with every one of the 35 source CLR properties
mapped explicitly to Go names, including its nested assigned-plan and sync-info
date fields. Exclude native createdDateTime/isInGracePeriod enrichments from that
source presentation. The two top-level timestamp dictionaries are flattened,
while sync-info retains its typed six properties and separately converted
additional properties. Construction owns all returned slices, maps and values.

Use an independent `TenantDateTime` representation with a `time.Time` value,
explicit UTC or Unspecified kind, and 100 ns precision. Offset input under the
source's AdjustToUniversal rule becomes UTC; timezone-free input remains a wall
clock. The existing conversation DateTime parser converts offsets to Local and
accepts an ISO-only profile, so reusing it would assert the wrong contract.

Represent associative values with a typed discriminated value rather than
`float64`-only JSON decoding: null, string, bool, Int32, Int64, Single, Double,
Decimal, array or object. Single must retain float32 rounding, signed zero and
infinity. Decimal, if reachable under an explicitly supported culture profile,
needs a decimal representation rather than conversion through float64. Native
raw dictionaries remain available independently.

The proposed pure `ProjectTenant` returns an owned view plus an inspectable
projection report. An unresolved date/culture path must produce an explicit
incomplete-projection error and path report, not a fabricated source null or a
successful complete view. Report formatting must not print tenant data. A later
resource composition can retain the original HTTP receipt alongside that result;
it must never fetch again because projection failed. This intermediate boundary
is not sufficient by itself to qualify Get-CsTenant.

## Actual conversion rules already settled

`JsonNumber.IsInteger` checks whether the raw lexeme contains a decimal point.
The integer branch tries Int32 and then Int64. Thus `1e2` and an integer greater
than Int64 become null; they do not fall through to floating-point parsing.
The non-integer branch tries Single, Double and Decimal using overloads without
an explicit culture. This host accepts Single infinity for `1.0e100` and
`1.0e1000`. Culture is observable: `1.20` becomes Single 1.2 under en-GB and
Single 120 under de-DE. A culture-free projector cannot claim both behaviors.

`JsonSerializable.FromJson` skips a JSON null object entry before calling
ToValue, but retains a null result from numeric conversion as a dictionary key
whose value is null. Nested objects recursively apply that rule. Arrays call
ToValue on every element; their JSON null sentinel uses the base ToString path
and becomes the string `"null"`. Empty object and array retain their distinct
collection types. The supplemental IL includes the recursive lambdas, generic
dictionary conversion and scalar methods; the data probes check these outputs.

## Date and culture closure still required

The generated date constructor calls InvariantCulture DateTime.TryParse with
AdjustToUniversal. Bounded own probes show that it accepts US numeric dates,
month names, date-only and time-only values in addition to ISO. Missing date or
year uses the host clock. Fractions beyond seven places are rounded to ticks.
At the lower bound, `0001-01-01T00:00:00+02:00` becomes UTC Jan 1 at 22:00;
the corresponding upper-bound overflow returns null. Invalid dates and empty
strings return null.

Before freezing a public projection implementation, resolve whether its API
explicitly supplies a reference clock and culture profile, and which date parser
source can support the full promised profile. The evidence does not justify a
short list of Go layouts being called complete CLR TryParse parity. An explicit
incomplete result is the honest intermediate behavior while that work continues;
it cannot be substituted for the command's full public contract end state.

## Own dispatch work

The public Get-CsTenant proxy selects the private GetCsTenant TrpsPSCmdlet.
Its descriptor names Get-CsTenantPoint for Modern execution. That helper sets
Extended, invokes internal Get-CsTenantObou and applies the three public
Select-Object/Add-Member/Write-Output statements. The compiled Modern default
command is Get-CsTenantObou_Get. Shared TRPS machinery is useful primary
implementation evidence, but this command still needs its exact parameter,
module-resolution and invocation/header handoff mapped. Runtime flighting or an
explicit override is not universal Modern or legacy qualification.
