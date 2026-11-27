# POL91 TenantAbuse nullable-list boundary

Scope: the two direct compiled Get/Set-CsTeamsTenantAbuseConfiguration commands,
Get Identity/Filter and Set Identity. This is source review and a concrete gap;
no qualification or tenant mutation is claimed.

The pinned Core DTO has seven Nullable<UInt64> scalar fields (LowSeatLimit,
CreateThreadThresholdPerSeat, AddMemberThresholdPerSeat,
SendMessageThresholdPerSeat, CreateThreadThresholdForTrialTenants,
AddMemberThresholdForTrialTenants, SendMessageThresholdForTrialTenants), two
String fields (TrialTenantValidation and SkipValidation), and getter-only
List<String> BlockedTenants. The public corresponding settings are matching
nullable scalars/strings and PSListModifier<String>, with only ParameterAttribute.
No minimum role, opaque-string enum, relationship among thresholds or server
eligibility rule follows from those CLR types. The existing full-width/zero/null
503 vectors prove controlled request bytes, not accepted writes or read decoding.

The exact concrete Set UpdateProperties source checks the bound BlockedTenants
key at IL16. A bound null or actual PSListModifierHelper.IsEmpty at37 clears the
prepared list at50; otherwise the actual PSListModifier.ApplyTo(IList) executes
at69. Individual bound scalar checks then assign only selected properties.
AreParametersSpecified tests bound-key presence for all ten settings; omission,
zero and null are distinct. The shared PolicyRp read/prepare/selected resolver,
route/auth/result/control boundaries are documented in
[the compiled source review](pol91-compiled-config-source-review.md).

## Portable null-member evidence

Run [the hash-checking probe](../../../scripts/discovery/observe-pol91-tenant-abuse-list.ps1)
against the extracted pinned 8.1.0 module:

```powershell
./scripts/discovery/observe-pol91-tenant-abuse-list.ps1 -ModulePath $ModuleRoot
```

[The recorded result](evidence/pol91-tenant-abuse-list.json) includes both exact
assembly hashes and the reproducer hash. It verifies the actual public list
parameter type and lack of extra binder attributes, then uses an inert compiled
PSCmdlet with that effective parameter shape. A literal array containing null
passes this binder; the actual PSListModifier applies it to the actual Core DTO,
and the actual BoundParametersContractResolver emits precisely
`{"BlockedTenants":[null,"entry"]}`. Actual Newtonsoft deserialization into the
same DTO retains the null member. These are primary-authored local vectors,
not captured service responses. The original management lifecycle,
authentication and HTTP never execute.

The actual host modifier additionally proves first-match removals and retained
nulls: start with `["entry",null,"entry",null]`, remove null then `"entry"`, add
null then `"ENTRY"`, producing `["entry",null,null,"ENTRY"]`. List contents are
created through PSListModifier replacement rather than PowerShell's separate
method-call conversion to String. Literal null passed directly to List<String>.Add
can coerce to empty before that method executes and would test a different
boundary. No such coercion is mislabeled a null-preserving result here.

## Native gap and reviewed correction

At the audit baseline, Configuration.BlockedTenants was []string. The Go JSON decoder silently
projects a null array member to an empty string, losing the source distinction
in the typed view even though RawFields retains original bytes. At that baseline,
policy.StringList and policy.ListEdit likewise stored only []string and could not
represent a null member for replacement/add/remove. This is a capability gap,
not a tenant permission blocker.

A correction must preserve null members and ownership, keep omission versus
clear separate, retain first ordinal removal/add ordering, and leave existing
nonnull constructors/call signatures unchanged where possible. Correcting the
canonical response field to []*string changes that unqualified family's public
field type; this tradeoff received explicit review rather than silently
coercing or dropping entries. An additive nullable-list builder/apply capability
can preserve existing shared StringList/ListEdit call signatures and behavior
for their other consumers.

After the API boundary is settled, bounded authored success tests should cover
all scalar types/full UInt64 width/null/zero, typed nullable entries/raw fields,
Global-only list/filter, prepared list delta and empty204 acceptance, retained
Before versus no invented After/persistence, plus a later failure and explicit
caller restoration with null entries. Existing403/503 tests remain denial
proofs. No source symbol, existing denial probe or this local serializer result
is itself a successful service operation or live qualification.

The implementation now uses canonical []*string, calls ApplyNullable on its
prepared snapshot, and extends the shared list constructors/edit application
additively. Existing nonnull signatures delegate through the same nullable
algorithm; constructors and apply copy both slices and pointed-to strings.
The helper test exercises null/empty distinction, constructor/apply ownership,
first removal, duplicate/order retention, empty clears and omitted/invalid edits.

New primary-authored TenantAbuse tests execute typed selected/list/filter
responses with full UInt64 width and null entries, a nullable prepared edit
followed by later403 and caller restoration/readback, and malformed successful
preparation retaining ErrDecode/raw receipt with no PATCH. They are not live
fixtures. No settings/Force-only errors now execute as zero-request cases.
Full policy and TenantAbuse SDK TestUnit/Example plus the lab local recovery
TestUnit passed at clean `4cb4ca15c6adccadb98d95208358b1bb2adcf2dd`,
session `52681`, with an additional exact pre-correction Messaging nonnull
list-edit compatibility case. Go 1.27.4 / GOMAXPROCS 2, `-p 1 -count=1`,
31.97 GiB admission / 14 GiB runtime floor; owned scratch removed. Post-run
selected/shared input diffs were zero. No TestAcc or race ran and no contract or
live status promotion attaches to this correction.

Other shared consumers inspected: Teams external access, ACS federation, external
access policy, shared calling routing, update management policy and messaging.
Their existing public builder/Apply signatures and nonnull wire behavior are
preserved. TenantAbuse and the independently source-proven Messaging field receive
canonical corrections. [Messaging source and execution boundaries](pol91-messaging-source-review.md)
remain separate; no other canonical field changes merely because it uses this helper.

## Complete command review boundary

Fresh own effective metadata and all nine actual scalar setters were checked:
seven Nullable<UInt64> and two String parameters/properties, only
ParameterAttribute, each setter plain stfld2/ret7. There is no HasValue/range/
nonempty branch hidden in those concrete setters. AreParametersSpecified checks
all ten selected setting keys, not their current values. The actual closed
ConfigurationNameNormalization helper returns TeamsTenantAbuseConfiguration.
Own Get Identity/Filter and Set mandatory Identity bindings are the direct
compiled source shapes; inherited nonempty Filter is independently recorded in
the initial sweep. Native path safety and early typed errors remain distinct
from public argument/host ordering. Explicit ReplaceNullableStrings() provides
the source bound-null/empty modifier clear; a nil edit remains omission.

The [shared compiled route/audience/status/list/filter/control/resolver source](pol91-compiled-config-source-review.md#shared-endpoint-audience-response-and-control-evidence)
is the same exact base/helper implementation. Successful raw/type decoding,
empty PATCH acceptance, source no-settings error and non-replay error receipts
are independent of tenant authority. No alternate endpoint, nested discovery,
minimum role, assignment action or paging envelope is invented.

Full `4cb4ca1` / `52681` includes the complete ten-setting runtime comparison,
all nine scalar value/null/full-width controlled503 wire proofs, actual403
error-envelope preservation, typed source-authored selected/list/Global-only
filter success for all getters, explicit zero/null/empty distinctions, nullable
list edits, later403 and caller restoration/readback, overflow200 with raw
receipt and no PATCH, and no-settings/Force/context/destination zero-request
checks. The old denied vectors are not relabeled success; the new successful
replies are clearly authored. Shared list ownership/order/clear/null behavior
also executes in the full policy package, with a legacy nonnull consumer proof
separately scoped. These are contract proofs, not live writes or a claim that
any opaque string/threshold/null-list mutation succeeds in a tenant.

For reproduction, use the own Get/Set Parameters/ParameterSets, closed Core
model property types/setter bodies and the portable owning-context IL reader
linked by the Messaging review on the concrete TenantAbuse overrides. Both
primary probes verify exact pinned assembly hashes. Live authorization,
nondefault threshold/blocked-tenant changes, authoritative oracle and safe
restoration/effective abuse control remain unqualified.
