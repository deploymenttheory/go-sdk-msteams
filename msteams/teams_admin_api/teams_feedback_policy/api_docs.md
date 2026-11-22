# Teams feedback policies

Use `sdk.TeamsAdminAPI.TeamsFeedbackPolicies` for MicrosoftTeams 8.1.0 feedback
policy operations. CRUD/list use the observed compiled Core backend; exported
Grant uses the observed modern backend. Both share the injected Teams scope and
session. The SDK runs native Go without a PowerShell runtime.

| Operation | Native API | Behavior |
|---|---|---|
| Read | GetByIdentity | One compiled GET |
| List | List / ListFiltered | Bare collection; exact Global first; local Identity/Tag wildcard matching |
| Create | CreateByIdentity | One compiled POST; empty settings permitted |
| Update | PatchByIdentity | Sparse compiled PATCH; at least one bound setting |
| Set composition | UpdateByIdentity | Set-metadata GET then PATCH; retains preparation and acceptance |
| Delete | DeleteByIdentity | Compiled DELETE with Identity/optional Force JSON body |
| Assign | AssignToUser / AssignTenantDefault / AssignToGroup / Assign | Modern scoped grants with forest-target discovery |

All six settings use `policy.Field[T]`: string UserInitiatedMode and
ReceiveSurveysMode, and nullable Boolean AllowScreenshotCollection,
AllowEmailCollection, AllowLogCollection and EnableFeatureSuggestions. A zero
Field omits a setting, `policy.Set` binds a value, and `policy.Null` binds JSON
null. Empty strings, false and null remain distinct. The pinned setters impose no
enum, length or cross-field rule. Native JSON requires actual string/Boolean
types and does not emulate arbitrary shell coercion.

```go
settings := teamsfeedbackpolicy.Settings{
    AllowEmailCollection: policy.Set(false),
    AllowLogCollection: policy.Null[bool](),
    UserInitiatedMode: policy.Set("FutureMode"),
}
_, response, err := sdk.TeamsAdminAPI.TeamsFeedbackPolicies.CreateByIdentity(
    ctx, "Tag:Lab", teamsfeedbackpolicy.CreateRequest{Settings: settings},
)
```

Select the full configuration identity explicitly. Compiled identities split at
the first slash and escape the remainder as one opaque name; wildcard whitespace
stays literal. Read models preserve nullable getters, RawFields and owned RawJSON.
Marshaling a read model returns its received snapshot; change settings through
request types. Set composition retains the read-model Identity in its body,
including null, and preserves the selected route identity. `UpdateResult` retains
Before, ReadResponse, PatchInvoked and HTTPAccepted on partial failure. It offers
no atomic update or persistence guarantee and does not replay unbound fields.
Compiled bound Force is sent as a string.

Modern grants require a native policy name argument; an empty string sends the
captured clear form. The mapped PowerShell helper coerces null to empty. `Assign`
requires exactly one scope and omits group Rank zero like that helper.
`AssignToGroup` can explicitly send priority zero through a nonnil `*int32`.
Negative signed priorities remain representable. Group 202 acceptance and 204
completion remain distinct; malformed accepted bodies retain acceptance metadata.
Discovery responses cannot establish assignment acceptance. Raw operation fields
are retained without implicit polling or effective-assignment guarantees.

The public Grant wrapper exposes no Force parameter; `AssignTenantDefault` also
offers the Boolean accepted by the captured underlying tenant helper. PassThru and
omitted required helper arguments fail at the source binding boundary and are
not exposed as native operations. Mutations never replay automatically.

See [unit tests](service_test.go)
and [lab acceptance](../../../docs/implementation/policies/feedback-policy-lab-acceptance.md).
All five commands/eight sets have native mappings. Issue #119 remains open for
full command and successful tenant qualification.

### Executed owned policy CRUD

`TestAccFeedbackPolicyCRUD` passed with certificate and client-secret
authentication. Each run creates an unassigned policy, verifies non-default
read-back, applies a sparse update, checks its list entry and verifies deletion
through selected HTTP404 and independent list absence. Two pinned public
MicrosoftTeams 8.1.0 reads agree with the full canonical getter projection.
Creation verifies `UserInitiatedMode=Disabled` and `AllowScreenshotCollection=false`;
update changes `UserInitiatedMode=Enabled` while preserving omitted screenshot collection.

The actual create/update/delete replies, created/updated reads, owned list and
missing-policy JSON supply `TestUnitOwnedCRUDUsesAcceptanceResponses`. The test
checks exact request bodies, omitted settings, update preparation/acceptance,
typed fields and complete raw snapshots. CI requires all four CRUD milestones,
both oracle comparisons and one cleaned journal entry. Global is never edited.
Assignments, effective end-user behavior and remaining parameter variants are
separate acceptance scenarios; these CRUD results do not close the parity ticket.
