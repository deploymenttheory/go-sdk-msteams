# Team template permission policy

Use `TeamsAdminAPI.TeamsTemplatePermissionPolicies` for the public compiled
MicrosoftTeams 8.1.0 Get/New/Set/Remove commands and their five parameter sets.
This resource manages team templates independently of meeting-template permissions.
No Grant command is assigned to this public resource surface.

`GetByIdentity`, `List`, `ListFiltered`, `CreateByIdentity` and `DeleteByIdentity`
follow direct PolicyRP. `PatchByIdentity` sends one sparse PATCH; `UpdateByIdentity`
prepares Set with GET, applies local list edits and PATCHes only bound fields.
Preparation and HTTP acceptance remain available independently of persisted state.
Optional Force uses the compiled string form; empty updates/Force alone fail before
credentials. Writes never replay and every operation propagates caller context.

Description is a nullable opaque string. HiddenTemplates is a constructor-initialized
read-only CLR collection: missing/null input retains an empty list; repeated array
properties append in wire order. Each HiddenTemplate owns Identity, nullable/missing
Id and complete raw fields. An explicit null Id setter fails like the pinned class;
`NewHiddenTemplate` is a pure local string constructor.

A nil edit leaves HiddenTemplates unbound. `ReplaceHiddenTemplates`,
`ClearHiddenTemplates` and `ChangeHiddenTemplates` preserve the source modifier
rules: clear for a bound null/empty modifier, replace when supplied, otherwise
remove the first ordinal case-sensitive Id match per Remove entry then append Add
without deduplication. Returned pointers/snapshots are owned. Native PATCH sends a
complete array, not server Add/Remove operators. Explicit array null is rejected;
a bound nil slice sends `[]`. Unknown received fields never become write settings.

Both application auth modes passed empty-list and populated owned CRUD. Populated
IDs were read from the tenant's actual team-template inventory. The test preserves
the list on a description-only update, performs Add/Remove and clear, compares all
public getters with pinned PowerShell four times, lists/deletes and independently
requires GET404/list absence. Journal kind `template-permission-policy` is bound
to its owned resource. Existing templates, Global and assignments are untouched.

Twenty permanent sanitized responses supply unit tests, including full populated
CRUD/list-edit replay. A real populated read exposed an initially ignored JSON
field; that response produced a failing unit regression before the decoder fix.
No separate response store or synthetic successful fixture is used.

Tested permission: dedicated Teams Administrator application role. Additional
binding combinations, effective user visibility, delegated modes, locale-specific
filtering and other clouds remain separate qualification. Sources: pinned Core
TeamsTemplatePermissionPolicy/HiddenTemplate/HiddenTemplates, Get/New/Set/Remove,
PSListModifierHelper, BaseTpm, provider and nested bound-parameter resolver.
[Microsoft reference](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamstemplatepermissionpolicy?view=teams-ps)
requires existing template IDs. See [qualification](../../../docs/acceptance/sip-template-lab-acceptance.md).
