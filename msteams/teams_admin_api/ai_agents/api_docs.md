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
| `Create(ctx, request)` | `New-CsAgent` | POST `/Teams.VoiceApps/aiagents` | 200, 201 |
| `UpdateByID(ctx, id, request)` | `Set-CsAgent` | PUT `/Teams.VoiceApps/aiagents/{id}` | 200 |
| `DeleteByID(ctx, id)` | `Remove-CsAgent` | DELETE `/Teams.VoiceApps/aiagents/{id}` | 200, 204 |
| `QueryPlatformAgents(ctx, request)` | `Get-CsAiAgents` | GET `/Teams.PlatformService/AiAgents?providerId=…` | 200 |
| `ListKnowledgeContainers(ctx, request)` | `Get-AIGeneratedKnowledgeContainer` | GET `https://teams.microsoft.com/api/mcps/df/mkb/crux/container-urls?threadId=[&callId=]` | 200 |
| `DeleteKnowledge(ctx, request)` | `Remove-AIGeneratedKnowledge` | DELETE `https://teams.microsoft.com/api/mcps/df/mkb/crux?threadId=[&callId=]` | 200, 202, 204 |

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
