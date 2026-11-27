# POL91 Messaging nullable-list correction

Independent effective metadata resolves Set-CsTeamsMessagingConfiguration to the
direct compiled Core command. FileTypeBlockList is PSListModifier<String> with
only ParameterAttribute; its actual DTO property is getter-only List<String>.
Concrete UpdateProperties checks binding at IL16, bound null or IsEmpty37 clears
at50, otherwise actual PSListModifier.ApplyTo(IList) executes at69. This source
was inspected independently, not inferred from the TenantAbuse helper consumer.

The [parameterized primary probe](../../../scripts/discovery/observe-pol91-tenant-abuse-list.ps1)
can run with `-ModulePath $ModuleRoot -Family Messaging`. Its
[own recorded result](evidence/pol91-messaging-nullable-list.json) verifies source
hashes/parameter shape, inert literal-array compiled binding, actual modifier,
actual Core DTO and actual resolver. Replacement emits exactly
`{"FileTypeBlockList":[null,"entry"]}`; actual DTO deserialization retains null.
First-match null/string removal and append ordering also preserve nulls. No
management ProcessRecord, authentication, HTTP or tenant mutation executes.
The TenantAbuse result stays separate under the same updated reproducer hash.

The former native []string field silently converted response null members to
empty strings. FileTypeBlockList is now canonical []*string and prepared edits
use ApplyNullable. The existing retained fixture clear/remove/replace tests are
adapted to the corrected field type; old nonnull builder call signatures remain
available. The field-type change is explicit in the family API reference.

The new primary-authored prepared-edit test asserts distinct null/string entries,
false and typed enum projection, a complete replacement body, Before/raw receipt
ownership and empty204 acceptance without an invented After. A separately
caller-issued authored readback preserves null entries and the earlier snapshot.
It is not a captured null-member write/readback or a claim about server-approved
file extensions. Full Messaging SDK TestUnit/Example passed at clean
`2b08f5705220980c3beed222e8942d39dd3a4b47`, session `67230`. The corresponding
lab package compiled with no matching offline tests; no TestAcc ran. Go 1.27.4 /
GOMAXPROCS 2 / `-p 1 -count=1`, 34.28 GiB admission and 14 GiB runtime guard;
owned scratch removed and post-run selected/shared input diffs zero. No race ran.
The earlier 4cb4ca1 / 52681 existing Messaging list-edit result covered only
nonnull compatibility before this field correction, not this new capability.

The wider command still requires finite review of all twelve public settings:
three Nullable<Boolean>, seven String, one nullable Int32 Storyline enum and this
list modifier. Existing retained get/list/nondefault/readback/restore models,
source-authorable bindings and model errors remain distinct. No command status
or live promotion follows from this one field correction.
