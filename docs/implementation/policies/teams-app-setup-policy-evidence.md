# POL-32 Teams app setup policy evidence

Issue #105 owns five baseline commands and nine exact parameter sets: Get
(Filter/Identity), New/Set/Remove (Identity), and Grant (GrantToUser/GrantToGroup/
GrantToTenant/Identity). This source/capture checkpoint establishes no native
service, contract pass or successful tenant qualification. The immutable runtime
snapshot and coverage states are unchanged.

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
request/response serialization. Native models must use int64 and avoid float
conversion. Newtonsoft's captured string-to-Order conversion does not by itself
establish Go coercion support; any typed-JSON difference must remain explicit.

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

## Native implementation sequence and qualification

Implement typed snapshots and readonly-list decoding with original response
retention. Keep the element shapes and three equality rules explicit; do not
silently unify a pin's Order with a customization key or add Order to calling pins.
Provide owned pure element/list builders, sparse native writes and explicit Set
read/edit/PATCH composition, then root wiring and three scoped assignments.
Map all nine parameter sets with exact wire/model/edit tests and local invalid
grant selection. Cover Int64 boundaries, duplicate properties, null/absence/false,
omission, future data, partial preparation, redaction, cancellation and no mutation
replay. Add opt-in bounded reads and separately owned write/read-back/cleanup cases.

No lab request, setup-policy write, assignment, app installation, directory role
or consent change has run for this family. Full shell/Newtonsoft conversions,
cloud/permission availability, tenant persistence/cleanup and effective client app
layout remain unqualified under #105. No source observation counts as full parity.
