# Teams files policies

Use `sdk.TeamsAdminAPI.TeamsFilesPolicies` for the compiled MicrosoftTeams 8.1.0
files policy resource. The SDK runs native Go without a PowerShell runtime.

| Operation | Native API | Behavior |
|---|---|---|
| Read | GetByIdentity | One compiled GET |
| List | List / ListFiltered | Bare collection; exact Global first; local Identity/Tag wildcard matching |
| Create | CreateByIdentity | One POST; empty settings permitted |
| Update | PatchByIdentity | One sparse PATCH; at least one bound setting |
| Set composition | UpdateByIdentity | Set-metadata GET then PATCH; retains preparation and acceptance |
| Delete | DeleteByIdentity | DELETE with Identity/optional Force JSON body |
| Assign | AssignUser / AssignTenantDefault / AssignGroup / Assign | Compiled scoped grants |

All four settings are `policy.Field[string]`: NativeFileEntryPoints,
SPChannelFilesTab, DefaultFileUploadAppID and FileSharingInChatsWithExternalUsers.
The latter two Go names preserve source JSON tags DefaultFileUploadAppId and
FileSharingInChatswithExternalUsers. A zero Field omits a setting, `policy.Set`
binds a value and `policy.Null` binds JSON null. Empty and null remain distinct.
The source setters impose no enum, GUID, length or cross-field rule; server
constraints remain separately qualified. Native JSON requires actual strings and
does not emulate arbitrary shell coercion.

```go
settings := teamsfilespolicy.Settings{
    NativeFileEntryPoints: policy.Set(""),
    SPChannelFilesTab: policy.Null[string](),
}
_, response, err := sdk.TeamsAdminAPI.TeamsFilesPolicies.PatchByIdentity(
    ctx, "Tag:Lab", teamsfilespolicy.PatchRequest{Settings: settings},
)
```

Pass the full identity explicitly. Compiled identities split at the first slash
and escape the remainder as one opaque name. Wildcard whitespace stays literal.
Read models preserve nullable getters, RawFields and owned RawJSON. Marshaling a
read model returns its received snapshot; use request settings for changes.
`UpdateResult` retains Before, ReadResponse, PatchInvoked and HTTPAccepted on
partial failure. Set composition retains the read-model Identity in the body,
including null, while retaining the selected route identity. It provides no
atomicity or persistence guarantee and does not replay unbound fields. Bound
CRUD Force serializes as a string.

Compiled grants preserve nil versus empty PolicyName. User and tenant PATCH
include additionalParameters:null; tenant uses capitalized PolicyName and Boolean
forceSwitchPresent. Required group Rank is signed Int32 and always sent,
including zero/negative values. Empty/mixed scope rejects before credentials/HTTP;
a parameterless Grant is rejected locally without HTTP. Group 202 is decoded as the compiled policy model; raw future fields/HTTP
metadata remain available without inferred polling or effective assignments.
Mutations never replay automatically.

See [unit tests](service_test.go)
and [lab acceptance](../../../docs/acceptance/files-policy-lab-acceptance.md).
All five commands/nine sets have native mappings. Issue #120 stays open for full
command and successful tenant qualification.

### Executed owned policy CRUD

`TestAccFilesPolicyCRUD` passed with certificate and client-secret
authentication. Each run creates an unassigned policy, verifies non-default
read-back, applies a sparse update, checks its list entry and verifies deletion
through selected HTTP404 and independent list absence. Two pinned public
MicrosoftTeams 8.1.0 reads agree with the full canonical getter projection.
Creation verifies `NativeFileEntryPoints=Disabled` and `SPChannelFilesTab=Enabled`;
update changes `SPChannelFilesTab=Disabled` while preserving omitted native entry points.

The actual create/update/delete replies, created/updated reads, owned list and
missing-policy JSON supply `TestUnitOwnedCRUDUsesAcceptanceResponses`. The test
checks exact request bodies, omitted settings, update preparation/acceptance,
typed fields and complete raw snapshots. CI requires all four CRUD milestones,
both oracle comparisons and one cleaned journal entry. Global is never edited.
Assignments, effective end-user behavior and remaining parameter variants are
separate acceptance scenarios; these CRUD results do not close the parity ticket.
