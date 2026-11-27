# AI agents

Use `sdk.TeamsAdminAPI.AIAgents` for Teams AI agents:

- the voice apps agent configurations that let a resource account hand calls to
  an external conversational provider such as Microsoft Copilot Studio;
- the Teams platform catalog of AI agents;
- the AI-generated meeting knowledge kept by the MCPS service.

Operations use the shared client/auth/transport controls and need no PowerShell
runtime. Constructing the root client or this service performs no network I/O.

| Resource method | PowerShell command | Wire operation | Declared successful statuses |
| --- | --- | --- | --- |
| `List(ctx, options)` | `Get-CsAgent` | GET `/Teams.VoiceApps/aiagents[?First=&Skip=&ExcludeContent=&NameFilter=&TypeFilter=&SortBy=&Descending=]` | 200 |
| `GetByID(ctx, id)` | `Get-CsAgent -Id` | GET `/Teams.VoiceApps/aiagents/{id}` | 200 |
| `Create(ctx, request)` | `New-CsAgent` | POST `/Teams.VoiceApps/aiagents` | 201 |
| `UpdateByID(ctx, id, request)` | `Set-CsAgent` | PUT `/Teams.VoiceApps/aiagents/{id}` | 200 |
| `DeleteByID(ctx, id)` | `Remove-CsAgent` | DELETE `/Teams.VoiceApps/aiagents/{id}` | 200 |
| `QueryPlatformAgents(ctx, request)` | `Get-CsAiAgents` | GET `/Teams.PlatformService/AiAgents?providerId=…` | 200 |
| `ListKnowledgeContainers(ctx, request)` | `Get-AIGeneratedKnowledgeContainer` | GET `https://teams.microsoft.com/api/mcps/df/mkb/crux/container-urls?threadId=[&callId=]` | any 2xx |
| `DeleteKnowledge(ctx, request)` | `Remove-AIGeneratedKnowledge` | DELETE `https://teams.microsoft.com/api/mcps/df/mkb/crux?threadId=[&callId=]` | any 2xx |

Every method takes a context and returns the actual Resty response alongside
its value/error. Reads are replay-safe; writes are never replayed automatically.
Any other HTTP status is an `*client.APIError`; an undeclared success status is
`ErrResponse`. Invalid input fails with `ErrInput` before discovery or
credentials are used.

## Agent configurations

`Agent` keeps the wire names (`ConfigurationId`, `AgentId`, `AgentType`,
`AgentTargetTagTemplateId`). The PowerShell projection renames them to `Id`,
`AIAgentId`, `AIAgentType` and `AIAgentTargetTagTemplateId`.

- `AgentType` is compared case-sensitively. The only documented value is
  `AgentTypeMicrosoftCopilotStudio`; others are rejected with HTTP 400
  (`AIAgent_AgentTypeNotRecognized`).
- A non-nil `AgentTargetTagTemplateID` must name an existing tag template, or
  the service answers 400 (`AIAgent_InvalidTagsTemplateId`).
- `UpdateByID` replaces the whole configuration, so send every field back.
  Set-CsAgent does the same with the instance it is given.
- A missing configuration is an HTTP 404 with top-level `code` 100004 on get,
  update and delete.
- `List` without options sends no query, as the module does. The paging and
  filter options come from the module's generated internal cmdlet.
  `AgentList.AgentCount` is the service's total.

Agent configurations carry no licence or billing. The agent ID is not checked
against the provider when the configuration is created. To make an agent
reachable, associate it with a resource account through
`ResourceAccounts.CreateAssociations`.

## Platform catalog

`QueryPlatformAgents` needs a `ProviderID` GUID; the service rejects other
values with HTTP 400 `BadArgument`. Nil filters are omitted, and switches are
sent as .NET `True`/`False`. Pass `QueryResult.ContinuationToken` back to read
the next page. `PlatformAgent` and `PlatformAgentDescription` type the common
catalog fields and keep the rest in `RawFields`.

## AI-generated knowledge

The knowledge methods use the MCPS service. The module hardcodes its origin
(`https://teams.microsoft.com/api/mcps/df`) and resource application ID
(`6bc3b958-689b-49f5-9006-36d165f30e00`), and requests that resource's
`/.default` scope. Only the commercial cloud is mapped; other clouds fail
locally with `client.ErrDestination`.

- `ListKnowledgeContainers` returns every container URL. The module prints them
  in batches sized for Purview import; the SDK returns them all.
- `DeleteKnowledge` permanently removes a meeting's or call's AI-generated
  knowledge and cannot be undone.

MCPS issues application tokens for its resource but rejects them with an empty
HTTP 401, so these methods need delegated credentials.

## Pinned source parsing and API correction

The continuation source packet is linked from
[the contract notes](../../../docs/implementation/ext/ai-agents-contract-notes.md).
ConfigAPI requests use the private default command names
`Get-CsAgent_Get`, `Get-CsAgent_Get1`, `New-CsAgent_New`,
`Set-CsAgent_SetExpanded`, `Remove-CsAgent_Remove`, and `Get-CsAiAgents_Get`.
MCPS independently uses the two public knowledge command names. Historical lab
JSON captures prove response bytes, not these corrected request headers.

This unreleased API corrects configuration `AgentList.AgentCount` to `*int64`;
catalog `QueryResult.AgentCount` remains `*int32`. `Channels`, `Capabilities`,
`Categories`, `Keywords`, `Disclosures` and `SupportedLocales` contain `*string` entries. All 25 pinned catalog description properties are typed, including developer privacy/terms, extra, five tile URLs and webpage. Generated parsing is
case-sensitive, ignores wrong-type fields, preserves nullable array positions,
and rejects exact duplicate keys recursively. It does not use the legacy
PolicyRp/Newtonsoft coercion rules. Catalog descriptions lazily default to an
empty non-nil description. A null configuration-list row fails public domain
projection with `client.ErrDecode` and preserves the response; catalog null rows
remain nil entries. Unknown original fields remain separately in `RawFields`.

Create omits an unbound nil target. Update sends all four fields, converting a
nil target to an empty string as the actual Set wrapper binding does. Empty
replacement strings are sent. Optional query strings omit only nil/empty values;
whitespace and control characters are safely encoded, and signed query integers
are not given an invented nonnegative constraint. Mutation IDs still use native
safe path validation. Public Get-CsAgent exposes none of `ListOptions`.

A successful object response with absent/null `AIAgent` projects an empty Agent.
Empty/whitespace or array roots produce a nil generated response, with the HTTP
receipt retained. Literal scalar roots fail. Remove accepts these same null
roots at exactly 200; no fabricated object is required. Consumers must inspect
nil values and responses independently, especially after writes, and must never
blindly replay a write on missing output. Diagnostics remain in raw responses;
the SDK does not reproduce PowerShell warning/verbose/output streams.

`ListKnowledgeContainers` now returns `[]*string`, retaining nulls and exact
numeric lexemes coerced to strings by its distinct legacy reader. This is a
native raw-list API; PowerShell prints semicolon-joined batches of 500 with
headings/warnings for larger results. An actual empty array succeeds, whereas an
empty body or null list fails the source's subsequent Count access and is
reported as `client.ErrDecode` with its receipt. MCPS accepts any 2xx before
these parse rules. Its observed app-only 401 and pinned pre-request certificate
failure are unavailable evidence, not successful operation. This repair makes
no tenant, authorization or seven-command qualification claim.
