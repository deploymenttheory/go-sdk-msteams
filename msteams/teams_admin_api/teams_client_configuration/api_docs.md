# Teams client configuration

Use `sdk.TeamsAdminAPI.TeamsClientConfigurations` for the compiled
MicrosoftTeams 8.1.0 TeamsClientConfiguration resource. Its baseline exposes
read/list and update operations. SDK runtime calls require no PowerShell.

| Operation | API | Behavior |
|---|---|---|
| Read | GetByIdentity | One compiled GET with explicit full identity |
| List/filter | List / ListFiltered | Bare collection; exact Global first; local Identity/Tag wildcards |
| Sparse update | PatchByIdentity | One compiled PATCH; at least one bound setting |
| Set composition | UpdateByIdentity | Explicit Set-metadata GET then sparse PATCH |

Select Global explicitly when reading/updating tenant settings. Compiled filters
keep whitespace literal. There is no evidenced create/delete/grant operation for
this configuration resource.

## Settings and snapshots

All 17 source settings have typed `policy.Field[T]` values: thirteen nullable
Booleans and four nullable strings. A zero Field omits a setting; `policy.Set`
binds a value and `policy.Null` binds null. Empty strings and false stay explicit.
RestrictedSenderList is opaque text. ContentPin, ResourceAccountContentAccess and
UseUnifiedDomain are also strings. The SDK adds no list splitting, sender/domain
normalization, numeric conversion, string enums, length or cross-field rule.
Future string values and ContentPin leading zeros remain unchanged.

```go
settings := teamsclientconfiguration.Settings{
    AllowEmailIntoChannel: policy.Set(false),
    RestrictedSenderList: policy.Set(""),
    UseUnifiedDomain: policy.Null[string](),
}
result, response, err := sdk.TeamsAdminAPI.TeamsClientConfigurations.UpdateByIdentity(
    ctx, "Global", teamsclientconfiguration.UpdateRequest{Settings: settings},
)
```

The example requires appropriate tenant authorization; server null/empty and
feature behavior remain separately qualified. Read models preserve nullable
known fields, RawFields and an owned complete RawJSON snapshot. Marshaling a read
model returns its received snapshot; use Settings to make changes. Known fields
require actual JSON string/Boolean types. Arbitrary source Newtonsoft primitive
coercions and public PowerShell binder behavior are separate boundaries.
AllowScopedPeopleSearchAndAccess uses the baseline JSON name
AllowScopedPeopleSearchandAccess.

## Partial updates and verification

UpdateResult retains Before, ReadResponse, PatchInvoked and HTTPAccepted even
when the preparation read or later update fails. Set uses the read model Identity,
including explicit null, in the body while retaining the requested route. It
patches only bound fields and never replays received future fields. A bound CRUD
Force switch serializes as a string, including false; Force alone is not a
setting. Mutations never replay. No implicit read-back, restoration, ETag or
atomicity guarantee is added. HTTP acceptance is separate from persistence and
effective client behavior.

See [unit tests](service_test.go)
and [lab acceptance](../../../docs/acceptance/client-configuration-lab-acceptance.md).
Both commands and all three exact parameter sets have native/test mappings.
Issue #112 and full command/tenant qualification remain open.
