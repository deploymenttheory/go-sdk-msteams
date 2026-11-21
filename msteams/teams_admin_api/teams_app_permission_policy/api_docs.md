# Teams app permission policies

`Client.TeamsAdminAPI.TeamsAppPermissionPolicies` provides native CRUD and scoped
assignments for the compiled TeamsAppPermissionPolicy resource.

| Go method | Baseline command / parameter set |
|---|---|
| GetByIdentity, List | Get-CsTeamsAppPermissionPolicy / Identity; explicit or omitted identity |
| ListFiltered | Get / Filter; local Identity/Tag wildcard matching |
| CreateByIdentity | New / Identity; local edits on fresh lists then POST |
| PatchByIdentity | Set / Identity; one sparse native PATCH |
| UpdateByIdentity | Set / Identity; explicit read/edit/PATCH composition |
| DeleteByIdentity | Remove / Identity |
| AssignUser | Grant / GrantToUser |
| AssignGroup | Grant / GrantToGroup |
| AssignTenantDefault | Grant / GrantToTenant |
| Assign with zero/mixed scopes | Grant / Identity; local input error, no invented endpoint |

The policy has Description, DefaultCatalogAppsType, GlobalCatalogAppsType,
PrivateCatalogAppsType and three independent catalog lists. Mode strings remain
open for future values, including empty/null. No description length maximum or
app-ID GUID/catalog-membership validation is invented. NewCatalogApp is a pure
constructor accepting empty, whitespace and arbitrary non-GUID strings.

The three source catalog types share Identity/Id fields and setter rules; the SDK
uses CatalogApp in each distinctly named list. Explicit null Id setters reject
decoding, including earlier bad duplicates; an absent Id can remain unset. Null
list elements are retained. ReplaceCatalogApps, ChangeCatalogApps and
ClearCatalogApps own their data. A nil request edit is unbound; a bound null/empty
modifier is ClearCatalogApps. Empty operators clear. Removes delete the first
ordinal Id match, ignoring Identity and future data; additions preserve order and
duplicates. Null comparator operands fail locally.

Readonly source lists have special decoding behavior. Null catalog properties
are ignored; repeated arrays append in wire order. Policy.RawJSON preserves the
original representation, including duplicates and future fields. RawFields shows
the last occurrence of each exact key for inspection. Received snapshots marshal
their original representation rather than reflecting edits to typed members.
Mutation inputs are separate; request DTO projection emits known fields only.

ScalarSettings uses policy.Set/Null to preserve binding, empty and null values.
Settings contains full replacement arrays: a nil bound slice becomes []; an
explicit null array field rejects locally because the source DTO getter emits an
array. UpdateRequest instead accepts local catalog edits. Set requires a setting;
Force alone is invalid. New permits empty settings and adds no implicit Global
comparison read, which is source ShouldProcess presentation. CRUD Force uses the
compiled True/False string body convention.

UpdateResult retains Before, ReadResponse, PatchInvoked and HTTPAccepted on partial
failure. Update takes body Identity from the read model while preserving the caller's
route. Native Patch defaults body Identity to the caller's identity, with an explicit
override available. A 204 has no fabricated Policy. HTTP acceptance does not prove
persisted/effective app access; there is no implicit read-back, wait, rollback,
mutation replay, ETag or atomicity guarantee.

Assign requires exactly one scope. Null and empty PolicyName remain distinct;
clear/default/effective semantics need tenant qualification. Group Rank is Int32
and serialized as priority, without a guessed backend range. User/group identities
are escaped as whole path segments. Tenant assignment has a Boolean Force flag.
Assignments are single native requests, not invented async jobs.

Requests use the selected Teams audience/origin and direct compiled-policy headers,
with fresh correlation and no generated forest or Graph lookup. List returns a
bare array with exact Global first; ListFiltered rejects bound empty patterns and
keeps literal whitespace. Models/edits/results redact ordinary formatting; raw
snapshots and responses remain caller-owned sensitive data.

Application/delegated synthetic fixtures establish routing, not successful tenant
RBAC, migration/feature availability or cloud support. Full shell/Newtonsoft
coercion, writes/read-back/cleanup and effective access remain unqualified under
#104. See [evidence](../../../docs/implementation/policies/teams-app-permission-policy-evidence.md)
and the [lab plan](../../../docs/implementation/policies/teams-app-permission-policy-lab-acceptance.md).
