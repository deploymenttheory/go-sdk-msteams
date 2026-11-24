# Teams first-party meeting-template configuration

Use `sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations` for the
compiled public MicrosoftTeams 8.1.0 Get and Set commands.

| Native operation | PowerShell behavior |
|---|---|
| `GetByIdentity` | Get Identity; explicit Global or full scoped identity |
| `List` | Get with omitted Identity; bare collection, exact Global first |
| `ListFiltered` | Get Filter; local Identity/Tag wildcard matching |
| `PatchByIdentity` | One sparse PATCH, with at least one bound setting |
| `UpdateByIdentity` | Set's preparation GET, local template edits and PATCH |

The direct compiled provider sends `MPACmdlet: true` and command metadata to
`/Skype.Policy/configurations/TeamsFirstPartyMeetingTemplateConfiguration`
without modern ConfigAPI discovery. Scoped identities split at the first slash.
Empty identity/filter and malformed filters fail before credentials.

`ScalarSettings.Description` distinguishes omission, explicit null and empty.
`UpdateRequest.TeamsMeetingTemplates` takes a `TemplateEdit`:
`ReplaceTemplates`, `ClearTemplates`, or `ChangeTemplates(add, remove)`, which
removes the first ordinal (case-sensitive) `Name` match per Remove entry and
then appends Add entries without deduplication. An empty delta clears the list,
like the source's empty modifier. Results own their pointers and snapshots.
`NewTemplate` and `NewOption` are pure local constructors.

Template writes send the complete replacement array with every typed template
and option property; unknown response fields are not written back. As in the
pinned serializer, binding the template list also sends the prepared root
Description. A bound nil slice on `PatchByIdentity` sends `[]`; a null array is
not a source DTO value and is rejected. Bound Force retains the compiled string
form. Read/modify/write has no concurrency or ETag guarantee, and mutations
never replay automatically. `UpdateResult` preserves preparation, mutation
invocation and HTTP acceptance separately; a 204 is not a persisted-state claim.

Getters are typed nullable fields with materialized template/option lists;
`RawFields` and owned `RawJSON` retain unknown fields, and marshaling a received
model returns the original snapshot.

The lab identity uses Teams Administrator; no additional permissions are needed.
Global read/list/filter and a journaled Description update/read-back/restore to
the original null value passed against the pinned module. Live template-list
writes, effective meeting behavior, delegated flows and other clouds remain
unqualified.

```go
_, _, err := sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations.UpdateByIdentity(
    ctx, "Global", teamsfirstpartymeetingtemplateconfiguration.UpdateRequest{
        ScalarSettings: teamsfirstpartymeetingtemplateconfiguration.ScalarSettings{
            Description: policy.Set("Lab first-party templates"),
        },
    },
)
```

See [qualification](../../../docs/implementation/policies/first-party-meeting-template-lab-acceptance.md).
