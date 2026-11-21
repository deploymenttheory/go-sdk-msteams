# Teams Cortana policies

Use `sdk.TeamsAdminAPI.TeamsCortanaPolicies` for the modern configuration and
grant paths selected by the MicrosoftTeams 8.1.0 baseline. SDK runtime calls
require no PowerShell. Retained commands do not prove feature availability or
effective tenant behavior.

| Operation | Native API | Behavior |
|---|---|---|
| Read/list/filter | GetByIdentity / List / ListFiltered | Modern GET; local Identity/Tag filtering; stable case-insensitive Global first |
| Create primitive | CreateByIdentity | One POST requiring 201 |
| Create composition | Create | Explicit POST/Get; retains successful submission on read-back failure |
| Update | UpdateByIdentity / UpdateGlobal | Sparse PUT requiring 204, including identity-only |
| Delete | DeleteByIdentity | Bodyless DELETE requiring 204 |
| Assign | Assign / AssignToUser / AssignTenantDefault / AssignToGroup | Scoped modern grants |

The root supplies its shared identity-bound ConfigAPI backend and forest
discovery. Discovery affects X-MS-Target-Uri; credentials stay on the configured
service origin. Modern command headers are used throughout. Explicit identity
operations reject empty/whitespace identities. ListFiltered whitespace selects
the whole collection as the mapped source helper does.

## Settings and snapshots

All five settings are independent optional typed pointers: Description and
CortanaVoiceInvocationMode strings; AllowCortanaVoiceInvocation,
AllowCortanaAmbientListening and AllowCortanaInContextSuggestions Booleans. Nil
omits a setting; pointers to empty strings and false bind those values. Isolated
public declaration binding converts null strings to empty and numeric Boolean
zero to false, and rejects null Booleans and Boolean strings. The typed API
expresses accepted final values explicitly. It adds no mode enum, Description
length or cross-field rule; server/runtime validation remains separately qualified.

```go
settings := teamscortanapolicy.Settings{
    Description: new(""),
    AllowCortanaVoiceInvocation: new(false),
}
response, err := sdk.TeamsAdminAPI.TeamsCortanaPolicies.UpdateByIdentity(
    ctx, "Tag:Lab", settings,
)
```

Read models retain an owned RawJSON snapshot and RawFields alongside known nullable
projections. Unknown or unprojectable future shapes remain raw with nil projections.
Marshaling a read model returns the original snapshot; use Settings for updates.
Create returns a ReadBackError with CreationResponse when POST succeeds but Get
fails. Submission and HTTP acceptance are separate from persistence/effects.

User/tenant grants require 200; group grants retain 202 acceptance separately from
204, including malformed accepted-body errors and future operation fields. An
actual POST response is required for acceptance; discovery metadata cannot prove
submission. Assign selects exactly one scope and omits absent/zero Rank like the
mapped wrapper. AssignToGroup can send explicit zero priority. Empty PolicyName
sends the captured clear form; effective clear semantics remain unqualified.
Tenant helper Force is Boolean; the public family does not expose Force. The
isolated mapped function rejects both PassThru forms; full runtime behavior is
not inferred from that seam.

All mutations disable replay. There is no implicit polling, restoration, elevation,
ETag or atomicity guarantee. See [unit tests](service_test.go)
and [lab acceptance](../../../docs/implementation/policies/cortana-policy-lab-acceptance.md).
All five commands and eight exact parameter sets have native/test mappings;
issue #113 and full qualification remain open.
