# POL-31 Teams app permission policy evidence

Issue #104 owns five commands and nine exact parameter sets: Get (Filter,
Identity), New/Set/Remove (Identity), and Grant (GrantToUser, GrantToGroup,
GrantToTenant, Identity). Native CRUD, explicit Set composition and three scoped
assignment methods now cover all nine mapped sets. Coverage remains in-progress
and metadata-only; full command/tenant qualification is not claimed. The immutable
runtime snapshot remains unchanged.

The accepted package is MicrosoftTeams 8.1.0, SHA256
`78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`.
Both scripts check the packaged Core, PolicyRp and host hashes and retain them
in their outputs. Source inspection concerns TeamsAppPermissionPolicy and its
Get/New/Set/Remove/Grant classes, the three catalog app/container types, compiled
base classes and Core PSListModifierHelper.

## Native contract

The family uses the direct compiled policy protocol, with resource name
TeamsAppPermissionPolicy. Collection GET/POST use
`/Skype.Policy/configurations/TeamsAppPermissionPolicy`; item GET/PATCH/DELETE
append the first-slash escaped configuration identity. The selected Teams
origin/audience and compiled MPA/command/correlation headers apply, without
generated forest discovery or Graph identity resolution.

The [eighteen provider cases](../../../contracts/endpoints/8.1.0/fixtures/app-permission-provider.json)
cover read/list, native create (including empty settings), sparse PATCH with
full/scalar/empty-list values, DELETE, user/group/tenant assignment, null catalogs,
invalid app Id setters, an invalid list envelope and resource denial. Each case
has one memory send, zero token requests, GUID correlation, MPA true and no target
URI. Bound-false CRUD Force is the True/False string convention. Grant uses its
own per-scope body shape and a Boolean tenant force flag.

The [development adapter](../../../tools/contracts/oracle/AppPermissionProviderCapture.cs)
invokes the pinned helpers with synthetic InvocationInfo and a memory sender.
Private client/version/host state is replaced/restored; the host refuses token
acquisition. No command constructor/lifecycle, telemetry, credentials, discovery
or network runs. A first capture exposed PowerShell's method binder converting a
null string argument to empty; the accepted capture selects true null inside C#
with an explicit flag and separately records empty-name grants. This establishes
the helper distinction without claiming full shell binding equivalence.

User and group grants use lower-case policyName; tenant uses PolicyName and
forceSwitchPresent. Null and empty PolicyName remain distinct. Group Rank is an
Int32 and the captured negative value is sent as priority; no backend range is
inferred. User and group identities are escaped as whole path segments. The
fourth Grant/Identity parameter set exposes no management parameters. The base
switch cannot select a grant for it and reports an error; the native selection
validator rejects an empty/mixed scope before credentials or HTTP, rather
than inventing a fourth assignment endpoint.

## Model and list behavior

[Model captures](../../../contracts/endpoints/8.1.0/fixtures/app-permission-models.json)
contain ten policy, twenty-one app and twenty-one list observations. The policy
has Identity, Description, DefaultCatalogAppsType, GlobalCatalogAppsType,
PrivateCatalogAppsType and three catalog lists. All three type setters accept
free strings, including unknown, empty and null values. Description's discarded
length constants do not enforce a limit; the actual setter accepts 2,000 characters.

Each public catalog list is a read-only property backed by a lazily initialized
container/list. A fresh policy presents all three as empty arrays. Deserialization
ignores null catalog properties and retains the lists. Repeated DefaultCatalogApps
arrays append to the same list: two occurrences with a and b produce [a,b], rather
than replacing the first. A plain Go struct unmarshal would lose this behavior;
the native decoder must process list properties in wire order and retain the
original response representation so duplicate properties are not silently erased.
An object-shaped list envelope fails decoding. Null list elements are retained.

DefaultCatalogApp, GlobalCatalogApp and PrivateCatalogApp each have Identity and
Id. Their Id setters reject null, but a parameterless deserializer may leave Id
absent. Constructors accept empty, whitespace and arbitrary non-GUID Id strings.
An earlier null Id setter still fails if a later duplicate supplies a valid value.
Do not infer GUID, length or catalog membership validation from the property name.

The three list modifiers use ordinal Id equality, ignoring Identity. Null and
empty modifiers clear; replacement/addition preserve order and duplicates. Each
Remove deletes only the first matching Id, then Add appends without deduplication.
Case matters: a and A differ. Source bound-property serialization includes all
known child DTO fields, not future received properties. Native snapshots and
mutation inputs must preserve that distinction.

## Native implementation and remaining qualification

TeamsAdminAPI.TeamsAppPermissionPolicies exposes Create/Get/List/Update/Delete,
local filtering, sparse PATCH, explicit Set read/edit/PATCH composition and
user/group/tenant assignment. Typed snapshots process readonly catalog properties
in order; RawJSON preserves original duplicates and future data while RawFields
provides last-occurrence inspection. Separate requests project known DTO fields.
Owned catalog-entry/list builders preserve source Id equality and list semantics.
Only the named resource was added to the compiled adapter; transport/auth behavior
is unchanged. Catalog mode strings remain open for future values.

Every one of the nine sets maps to compiling symbols and source-backed tests.
All eighteen request fixtures run in application/delegated synthetic modes.
Model tests compare all ten policy, twenty-one app and twenty-one list observations,
including canonical CLR projections separately from raw received presence. Local
grant selection, input/context/backend guards, all three catalogs' duplicate
append/null/setter rules, DTO projection, update preparation, status/redaction and
single-send mutation failures are covered. Structural reports keep coverage_claim
false and live not-run; these are not full parity or live passes.

Actual Global reads on 2026-11-21 returned resource HTTP 403 using the existing
dedicated lab SPN in both client-secret and certificate modes. Filter/list and
optional pinned PowerShell comparison were not reached. No payload was retained
and no policy, assignment, app, role or consent changed. Get's ledger live state is
unavailable; writes remain not-run. The opt-in read harness can compare known
nullable fields and three catalogs with a tenant-bound hash-checked module when
access is available.

Successful tenant roles, migration/feature availability, cloud support, full
shell/Newtonsoft conversions, writes/read-back/cleanup and effective app access
remain unqualified under #104. Native response fields require their Go types;
raw presence differs from CLR presentation, and conservative identity validation
does not claim every shell string. Future write acceptance requires owned policy
and principal fixtures with an authorized journal and restoration/cleanup plan.

Local verification on 2026-11-21 passed the full offline race/shuffle suite,
SDK/examples build, vet, lint with zero issues, ledger validation and its ten
validator tests. All three PowerShell scripts parse. The lab comparison selects
Global by identity and permits extra Tag:Global filter matches, with a regression
test for ambiguous/missing Global selections. Revision-bound structural artifacts
are reviewed independently of the unchanged qualification states.
