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

The complete command review below covers all twelve public settings:
three Nullable<Boolean>, seven String, one nullable Int32 Storyline enum and this
list modifier. Existing retained get/list/nondefault/readback/restore models,
source-authorable bindings and model errors remain distinct. No command status
or live promotion follows from this one field correction.

## Complete concrete setting and control inventory

| All twelve public settings | Own effective parameter and DTO type |
|---|---|
| EnableVideoMessageCaptions, EnableInOrganizationChatControl, CustomEmojis | Nullable<Boolean> |
| Communities, MessagingNotes, FileTypeCheck, UrlReputationCheck, ContentBasedPhishingCheck, ReportIncorrectSecurityDetections, FileTypeCustomizationSetting | String |
| Storyline | Nullable<Core.Storyline>, Int32 enum |
| FileTypeBlockList | PSListModifier<String> / getter-only List<String> |

All eleven scalar parameters have only ParameterAttribute and every actual
scalar DTO setter is plain stfld at IL2 followed by ret7. No per-field enum or
range rule is invented from names of the seven opaque strings. The concrete
AreParametersSpecified has only BoundParameters.ContainsKey calls: setting name
loads at IL11/37/63/89/115/141/167/190/213/236/259/282. It tests bound presence,
not HasValue/current truth or string content. Selected null/false/empty are
settings; Force alone is not. Existing local Update no-settings and shared
source guard evidence are separate from received DTO decoding.

The [actual-type compiled enum binder probe](evidence/pol91-storyline-binding.json)
with [portable reproduction](../../../scripts/discovery/observe-pol91-storyline-binding.ps1)
closes the enum distinction. Source members are exactly Disabled=0 and Enabled=1.
Null and the names bind and the actual DTO/resolver emit null/0/1. Raw unknown
numeric37 and string '-7' are rejected by the PowerShell binder; an already typed
actual Core.Storyline enum37 is accepted and emits37. The inert probe closes a
generic PSCmdlet over the actual enum, avoiding a different enum definition or
Core-reference compiler-warning suppression. It never invokes management code,
authentication or HTTP. Native typed Storyline(37) and received future numeric
values remain representable; native JSON convenience decoding is not mislabeled
identical to raw PowerShell numeric-argument binding. Existing enum/model tests
prove Int32 bounds, known names/case, numeric preservation and invalid shapes.

The actual closed ConfigurationNameNormalization helper returns
TeamsMessagingConfiguration. Own Get has optional Identity and bound nonempty
Filter; Set has mandatory Identity. The same exact shared compiled
[route, audience, result, filter, control and resolver source](pol91-compiled-config-source-review.md#shared-endpoint-audience-response-and-control-evidence)
applies, including bare collections, stable exact Global-first order, local
wildcards/literal whitespace, broad 2xx versus typed decoding, empty acceptance,
no-settings error and ReplayNever. Native encoded-path guards, early errors,
caller confirmation and typed/raw response presentation are explicit Go
boundaries rather than equivalent PowerShell host pipeline behavior.

The full `2b08f57` / `67230` suite executes retained selected/list/filter and all
thirteen response projections (Identity plus twelve settings), selected scalar
null/value serialization, source enum conversion, prepared sparse204, list
clear/remove/replacement, owned raw snapshots, model errors/no later PATCH,
a retained real Privacy403 envelope reused in an authored Messaging request,
non-replay/context/destination guards and the new nullable-entry
prepared edit/readback. HTTP acceptance remains distinct from effective state;
actual captured readback and authored null-entry responses are not conflated.
No new generic-helper permutations are needed solely to repeat established
common source behavior for each scalar name.

To reproduce the independent metadata/control checks, Get-Command the concrete
Get/Set pair, inspect Parameters/ParameterSets and the closed model from
ImplementingType.BaseType.GetGenericArguments()[0]. For each writable declared
model property inspect its actual setter; use the
[portable owning-context IL reader](pol04-primary-source-audit.md#portable-reproduction)
on Core.SetTeamsMessagingConfigurationCmdlet.AreParametersSpecified and
UpdateProperties. Source hashes are verified by the portable probes. This does
not depend on a transient host file, symbol-count inference or server permission
assumptions.

Finite review found that earlier success tests invoked selected Get and Filter
but not the Identity-omitted List wrapper. The nullable-list test now adds one
direct source-authored bare-collection read and asserts typed identity/null
entries/raw receipt before its existing prepared edit/readback. This supplemental
default-set workflow passed at clean `7a902647bec038eccdba651cee717dac2fa910c2`,
session `39584`, selecting only its exact function (31.32 GiB admission,
Go 1.27.4 / GOMAXPROCS 2 / `-p 1 -count=1`, owned scratch removed,
14 GiB floor). Full 2b08f57 proof continues to cover the unchanged cases; this
is not mislabeled a second full-package run.

Own Get Identity/Filter and Set Identity are individually implemented/contract-
passed after complete source/execution review; [versioned records](../../../contracts/coverage/evidence/pol91-nullable-config-commands.json)
keep full and supplemental execution separate. Every live state stays unchanged;
full #185 qualification remains open.
