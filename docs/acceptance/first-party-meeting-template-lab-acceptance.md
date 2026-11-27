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

- Core: `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49`
- PolicyRP: `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51`

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
binding behaviour is unit-tested only. Effective meeting
behaviour, delegated flows, certificate authentication for this service
partition (proven once by the auth partition), and other clouds also remain
unqualified.
