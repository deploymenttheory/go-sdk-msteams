# Teams events policies

Use `sdk.TeamsAdminAPI.TeamsEventsPolicies` for the compiled MicrosoftTeams 8.1.0
TeamsEventsPolicy resource. The SDK runs native Go without a PowerShell runtime.

| Operation | Native API | Behavior |
|---|---|---|
| Read | GetByIdentity | One compiled GET |
| List | List / ListFiltered | Bare collection; exact Global first; local Identity/Tag wildcard matching |
| Create | CreateByIdentity | One POST; empty settings allowed |
| Update | PatchByIdentity | One sparse PATCH; at least one bound setting required |
| Set composition | UpdateByIdentity | Set-metadata GET then sparse PATCH; preserves preparation and acceptance |
| Delete | DeleteByIdentity | One DELETE with Identity/optional Force JSON body |
| Assign | AssignUser / AssignTenantDefault / AssignGroup | Compiled scoped grants; exactly one scope via Assign |

Pass the configuration identity explicitly, for example `Global` or `Tag:Lab`.
Scoped identities split at the first slash and escape the remainder as one opaque
name. Whitespace in wildcard filters remains literal. Read models retain nullable
typed getters, RawFields and an owned RawJSON snapshot. Marshaling a read model
returns its received snapshot; use request settings for changes.

All 25 settings use `policy.Field[T]`: 23 strings and nullable Boolean
AllowEventIntegrations/UseMicrosoftECDN. The zero Field omits a setting,
`policy.Set` binds a value and `policy.Null` binds explicit JSON null. Empty
strings, false and null remain distinct. Source strings are opaque: no enum,
Description length or cross-field restriction is inferred. Native JSON requires
typed strings/Booleans and does not emulate arbitrary shell coercion.

```go
settings := teamseventspolicy.Settings{
    UseMicrosoftECDN: policy.Set(false),
    AllowEventIntegrations: policy.Null[bool](),
    Description: policy.Null[string](),
}
_, response, err := sdk.TeamsAdminAPI.TeamsEventsPolicies.CreateByIdentity(
    ctx, "Tag:Lab", teamseventspolicy.CreateRequest{Settings: settings},
)
```

An empty create matches compiled New processing. Identity-only/Force-only Set
does not satisfy its setting predicate. `UpdateResult` retains Before,
ReadResponse, PatchInvoked and HTTPAccepted on partial failure. The composition
uses the read model Identity in its body, including null, while retaining the
selected route identity. It does not replay unbound properties or provide an
atomic update guarantee. HTTP acceptance does not prove persistence or effects.

Compiled grants preserve nil versus empty PolicyName. User and tenant PATCH
bodies include additionalParameters:null; tenant PolicyName is capitalized and
forceSwitchPresent is Boolean. Required group Rank is signed Int32 and always
sent, including zero/negative values. Bound CRUD Force is a string. Empty/mixed
scope rejects before credentials/HTTP; the public parameterless Grant set has an
observed source error with no request. Group 202 returns raw model/HTTP metadata
without inferred polling or effective assignment semantics. Mutations never
replay automatically.

The pinned module exports compiled Core cmdlets for this family even though
modern configuration mappings also exist. Captures verify the actual exports;
modern mapping presence does not select modern dispatch.

See [source evidence](../../../docs/implementation/policies/events-policy-evidence.md)
and [lab acceptance](../../../docs/implementation/policies/events-policy-lab-acceptance.md).
All five commands/nine sets are mapped. Issue #118 remains open for full command
and tenant qualification.
