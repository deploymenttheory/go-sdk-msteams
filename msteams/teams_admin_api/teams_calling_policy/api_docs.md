# Teams calling policies

Use `sdk.TeamsAdminAPI.TeamsCallingPolicies` for the compiled MicrosoftTeams 8.1.0
TeamsCallingPolicy resource. No PowerShell runtime is required by the SDK.

| Operation | Native API | Behavior |
|---|---|---|
| Read | GetByIdentity | One compiled GET |
| List | List / ListFiltered | Bare collection; exact Global first; local Identity/Tag wildcard filter |
| Create | CreateByIdentity | One POST; no implicit Global comparison/read-back |
| Update | PatchByIdentity | One sparse PATCH; at least one setting required |
| Set composition | UpdateByIdentity | Explicit Set-metadata GET then sparse PATCH; retains preparation/acceptance |
| Delete | DeleteByIdentity | One DELETE with Identity/optional Force JSON body |
| Assign | AssignUser / AssignTenantDefault / AssignGroup | Compiled scoped grants; exactly one scope via Assign |

Pass the full configuration identity explicitly, such as `Global` or `Tag:Lab`.
Compiled wildcard filtering keeps whitespace literal. Scoped identities encode
opaque segments independently. Read models keep nullable typed values, RawFields
and an owned RawJSON snapshot. Marshaling a read model returns its received
snapshot; use request settings to make changes.

## Settings and validation

All 39 baseline settings are exposed as `policy.Field[T]`. A zero Field omits the
setting, `policy.Set` binds a value, and `policy.Null` binds explicit JSON null.
Empty strings, false, zero and full signed Int64 values remain distinct. The
source model allows null settings; a public PowerShell binder is a separate seam.
CallingSpendUserLimit uses signed Int64 without an inferred local spend limit.
CallRecordingExpirationDays accepts null or 1..9999. The custom recording message
identifier accepts null or a hyphenated GUID, including the source-accepted final
LF. Other string choices, paths, lengths and cross-field rules are server
validated; future string values are retained.

Native JSON decoding requires actual string, Boolean and Int64 JSON types. The
source Newtonsoft model additionally coerces some numeric strings/fractions;
those coercions are observed in the fixture and are not performed by the SDK.
Typed settings can express their final accepted values without precision loss.

```go
settings := teamscallingpolicy.Settings{
    AllowPrivateCalling: policy.Set(false),
    CallingSpendUserLimit: policy.Set(int64(0)),
    CallRecordingExpirationDays: policy.Set(int64(120)),
    Description: policy.Null[string](),
}
_, response, err := sdk.TeamsAdminAPI.TeamsCallingPolicies.CreateByIdentity(
    ctx, "Tag:Lab", teamscallingpolicy.CreateRequest{Settings: settings},
)
```

The API example requires appropriate tenant authorization. Creation and HTTP
acceptance do not prove persistence or effective calling behavior. UpdateResult
retains Before, ReadResponse, PatchInvoked and HTTPAccepted on partial failure.
Set composition uses the read model Identity in the body, including explicit
null, while retaining the requested route identity. It does not merge/replay
unbound fields or provide ETag/atomicity guarantees.

Compiled grant PolicyName is `*string`: nil and empty remain distinct on the
wire. Group Rank is a required signed Int32, including zero. Tenant Force is a
Boolean; compiled CRUD bound Force serializes as a string. Empty/mixed scope
selection rejects before credentials/HTTP. Mutations never replay; there is no
implicit polling, restoration or elevation. Backend clear/effect semantics need
separate qualification.

## Testing

[Fixture-backed CRUD unit tests](service_crud_response_test.go) load the actual
responses from the owned unassigned-policy acceptance workflow. Certificate and
client-secret runs verify non-default creation, sparse updates, omitted-field
preservation, listing, deletion and independent absence, with PowerShell reads.
All five commands and nine parameter sets retain their mappings. Issue #110 stays
open for assignment, effective-policy and remaining mode/parameter qualification.
