# POL-32 Teams app setup policy evidence

Issue #105 owns five baseline commands and nine exact parameter sets: Get
(Filter/Identity), New/Set/Remove (Identity), and Grant (GrantToUser/GrantToGroup/
GrantToTenant/Identity). Native CRUD, explicit Set composition, local builders
and three scoped assignment methods now cover all nine mapped sets. The ledger
remains in-progress and metadata-only without a full command/tenant parity claim.
The immutable runtime snapshot is unchanged.

The accepted MicrosoftTeams 8.1.0 package SHA256 is
`78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`.
Both scripts verify Core, PolicyRp and host assembly hashes and retain them in
the fixtures. Local source inspection covers TeamsAppSetupPolicy, its five command
classes, six element/container types and the compiled base/provider/list helpers.

## Protocol and request evidence

The direct compiled resource is TeamsAppSetupPolicy. Collection GET/POST use
`/Skype.Policy/configurations/TeamsAppSetupPolicy`; item GET/PATCH/DELETE append
the source's first-slash escaped configuration identity. The selected Teams
origin/audience and compiled MPA/command/fresh correlation headers apply, without
generated forest discovery or Graph lookup.

[Twenty-one provider observations](../../../contracts/endpoints/8.1.0/fixtures/app-setup-provider.json)
cover read/list/create/patch/delete, empty/scalar/full-list writes, all three grant
scopes with true-null/empty policy names, null lists, null Id/customization setters,
Int64 Order, overflow, an invalid list envelope and denial. Every case has one
memory send, zero token requests, GUID correlation, MPA true and no target URI.
CRUD bound-false Force is a string; tenant assignment Force is Boolean.

[The development adapter](../../../tools/contracts/oracle/AppSetupProviderCapture.cs)
uses actual pinned HttpRequestHelper methods with synthetic InvocationInfo,
credential-refusing host and memory sender, restoring private client/version/host.
No command constructor/lifecycle, credentials, discovery or network runs. Bound
list membership is supplied alongside the typed DTO's actual contents; request
bodies contain complete known element fields and arrays, not server Add/Remove
operators. True null PolicyName is selected in C# to avoid method-binder coercion.

The compiled base grant switch has user/group/tenant operations. Its fourth
parameterless Identity set cannot select an operation. Native validation should
reject a missing/mixed scope before credentials/HTTP rather than invent an endpoint.
New permits empty settings; Set requires a bound setting, with false/null counted
as bound and Force alone insufficient. The optional New Global comparison read
serves ShouldProcess presentation and is separate from native resource preparation.

## Models, decoding and list equality

[Model evidence](../../../contracts/endpoints/8.1.0/fixtures/app-setup-models.json)
contains twelve policy, fifty-six element and forty-eight list observations.
The policy has Identity, Description, nullable AllowSideLoading/AllowUserPinning,
and six readonly lists: AppPresetList, PinnedAppBarApps, PinnedMessageBarApps,
AppPresetMeetingList, AdditionalCustomizationApps and PinnedCallingBarApps.
Some containers initialize in the constructor and others lazily; serialization
observes all six as empty arrays. Null list properties are ignored and repeated
arrays append. Ordered decoding and original-response retention are required;
plain last-property replacement would lose source behavior. Description's discarded
length constants do not enforce a limit; 2,000 characters are accepted.

All elements have Identity and Id. Their Id setters reject null, while a missing
setter can leave Id absent; empty and future non-GUID strings are accepted.
AdditionalCustomizationApp also has AdditionalCustomizationId with the same
null-setter rule. The two bar pin elements have signed Int64 Order, default zero,
with no source range check. Negative values and Int64 maximum are accepted;
null/overflow setters fail. Captures preserve 9,007,199,254,740,993 exactly through
request/response serialization. Native models use int64, never float conversion,
and implement the captured numeric-string Order form with checked decimal parsing.
Other arbitrary Newtonsoft conversions remain separately unqualified.

Each list uses its command-specific ordinal equality predicate:

| List | Comparison key |
|---|---|
| AppPresetList | Id |
| AppPresetMeetingList | Id |
| PinnedCallingBarApps | Id |
| PinnedAppBarApps | Id and Order |
| PinnedMessageBarApps | Id and Order |
| AdditionalCustomizationApps | Id and AdditionalCustomizationId |

Identity is not a key. Null/empty modifiers clear; Remove deletes the first match
per entry, then Add appends without deduplication. Replacement and addition retain
order and duplicates. The three-removal cases distinguish same Id with a different
Order/customization from Id-only lists. Future received fields must stay in
snapshots and be excluded from source-compatible request DTO projection.

## Native implementation and qualification

TeamsAdminAPI.TeamsAppSetupPolicies exposes native CRUD, sparse PATCH, explicit
read/edit/PATCH composition and user/group/tenant assignments. Typed App, PinnedApp
and CustomizationApp models and edits keep all six named list fields and the three
equality predicates explicit. Constructors/edits own their pointers, list contents
and raw data; nil/zero edit receivers reject safely. Only the named compiled
resource is added to the existing adapter; transport/auth behavior is unchanged.

All nine parameter sets map to compiling methods and tests. All twenty-one
provider fixtures run in application/delegated synthetic modes. Canonical model
projections compare all twelve policy, fifty-six element and forty-eight list
observations. Exact JSON-number comparisons prevent float precision from hiding
incorrect Int64 handling. The captured string Order first failed and passed after
checked native parsing. Tests cover every list/key, readonly duplicate append and
null behavior, owned data, known DTO projection, body identity, false/omission,
partial preparation, no-HTTP builders/invalid grant selection, local guards,
status/redaction and single-send mutations. Structural reports retain
coverage_claim false and live not-run.

Actual Global reads on 2026-11-21 returned resource HTTP 403 with the existing
dedicated lab SPN in both client-secret and certificate modes. Filter/list and
optional pinned module comparison were not reached. No payload was retained and
no policy/layout, assignment, app installation, role or consent changed. Get's
ledger live state is unavailable; writes remain not-run. The opt-in harness can
compare canonical nullable fields and six lists with a tenant-bound hash-checked
PowerShell module, preserving Int64 numbers without float conversion. It rejects
incomplete properties rather than manufacturing empty arrays.

Full shell/Newtonsoft conversions, cloud/permission availability, successful
tenant writes/read-back/cleanup and effective client app layout remain unqualified
under #105. Native JSON fields require their Go types except the captured checked
Order string form. Raw presence differs from CLR default presentation; conservative
identity validation does not claim every shell input. Future write acceptance
requires owned policy/principal fixtures and authorized journals/restoration.

Local verification on 2026-11-21 passed the full offline race/shuffle suite,
SDK/examples build, vet, scoped lint with zero issues, ledger validation and its
ten validator tests. All three PowerShell scripts parse, including the tightened
collection/element completeness checks. Revision-bound structural artifacts are
reviewed independently of the unchanged qualification states.
