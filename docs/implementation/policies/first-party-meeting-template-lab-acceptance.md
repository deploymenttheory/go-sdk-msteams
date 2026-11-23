# First-party meeting-template configuration qualification

Owner: #121 (POL-48). Baseline: MicrosoftTeams 8.1.0.

The pinned runtime metadata has Get Identity/Filter and Set Identity. Both are
compiled Core cmdlets (`GetTeamsFirstPartyMeetingTemplateConfigurationCmdlet`,
`SetTeamsFirstPartyMeetingTemplateConfigurationCmdlet`) deriving from the
shared `BaseTpmGetCmdlet` / `BaseTpmSetCmdlet`, so they use the direct compiled
PolicyRP protocol: `GET`/`PATCH`
`/Skype.Policy/configurations/TeamsFirstPartyMeetingTemplateConfiguration[/configuration/{identity}]`.
The type name is not in `ConfigurationNameNormalization`'s rename list. There are
no public create/delete commands for this singleton.

Local primary sources: Core `TeamsFirstPartyMeetingTemplateConfiguration`,
`TeamsFirstPartyMeetingTemplateConfigurationTeamsMeetingTemplates`,
`TeamsMeetingTemplateType`, `TeamsMeetingOption`, both cmdlets and
`PSListModifierHelper`; PolicyRP `BaseTpmSetCmdlet`, `HttpRequestHelper`
(`SetAsync`, `GetNestedBoundParameters`) and `BoundParametersContractResolver`.
Pinned assembly SHA256 values:

- Core: `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49`
- PolicyRP: `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51`

Set reads the selected identity, applies bound `Description` and the
`TeamsMeetingTemplates` list modifier (keyed by ordinal `Name` equality), then
PATCHes. The serializer only writes bound property names, but binding
`TeamsMeetingTemplates` expands the nested template/option property names, which
include `Description`; the prepared root Description is therefore serialized
with a template edit even when it was not bound. This was confirmed offline by
invoking the pinned `GetNestedBoundParameters` and contract resolver against
the pinned types: a templates-only bind produced
`{"Identity":"Global","TeamsMeetingTemplates":[{...all typed fields...}],"Description":"prepared"}`
and a Description-only bind produced `{"Identity":"Global","Description":...}`.
The Description setter's compiled length checks are no-ops; no length or enum
validation is invented. Template and option `Name` setters reject null.

## Live results

On 2026-11-23 the dedicated SDK service principal (Teams Administrator, client
secret) read Global by identity and via the local filter, and compared the
complete Description and nested template/option projection with PowerShell
8.1.0. The tenant has a null Description and eight populated first-party
templates.

`TestAccFirstPartyMeetingTemplateConfigurationUpdateRestore` journals the
original (null) Description, sets a non-default Description, requires persisted
read-back with the template list unchanged, compares with the pinned module,
then restores the original null Description, independently re-reads and
compares again. The journal entry was cleaned and a separate PowerShell read
confirmed the restored null Description. No licence, role, permission or
template list was changed.

Five permanent sanitized response files come from these runs (`get_global`,
`list`, `update_204`, `update_readback`, `restored_readback`). Sanitization
replaces template/option names and descriptions; the read fixture still carries
the actual populated template structure. Unit tests use them for routing,
headers, typed getters, snapshot ownership, filters, sparse Description PATCH,
and full-array template replace/clear/Remove/Add bodies. Common PolicyRP denial
propagation uses the actual privacy-service 403 envelope and is not a live
denial claim for this resource.

## Still unqualified

Live template-list writes are deliberately not exercised: the list holds
Microsoft-managed first-party templates and rewriting it could detach the
tenant's copy from service updates, which cannot be reliably restored. Their
binding contract is source-backed and unit-tested only. Effective meeting
behaviour, delegated flows, certificate authentication for this service
partition (proven once by the auth partition), and other clouds also remain
unqualified.
