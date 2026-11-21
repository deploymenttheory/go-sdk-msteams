# POL-31 Teams app permission policy evidence

Issue #104 owns five commands and nine exact parameter sets: Get (Filter,
Identity), New/Set/Remove (Identity), and Grant (GrantToUser, GrantToGroup,
GrantToTenant, Identity). This source/capture checkpoint establishes no native
service, contract pass or successful tenant qualification. Coverage states and
the immutable runtime snapshot remain unchanged.

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
validator should reject an empty/mixed scope before credentials or HTTP, rather
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

## Implementation sequence and remaining qualification

Implement typed snapshots and ordered readonly-list decoding first, with pure
owned catalog-entry/list builders and explicit omission/null/empty semantics.
Then expose native Create/Get/List/Update/Delete, local filtering, sparse PATCH,
explicit Set read/edit/PATCH composition, and user/group/tenant assignment. Add
only the named resource to the compiled adapter and wire its root collection.
Keep source free-form catalog modes open for future values.

Map every one of the nine sets to compiling symbols and source-backed tests.
Cover no-HTTP invalid grant selection, input/context/backend guards, the three
catalogs' setter and duplicate-property semantics, DTO projection, partial update
preparation, status/redaction and no mutation replay. Add bounded opt-in reads
and separately owned write/read-back/cleanup scenarios. These observations do
not establish tenant roles, migration/feature availability, cloud support, full
shell/Newtonsoft conversions or effective app access. No lab request, policy or
assignment write, role or consent change has run for this family.
