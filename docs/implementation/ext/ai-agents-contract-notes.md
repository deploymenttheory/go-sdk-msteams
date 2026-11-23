# AI agent contracts

Owner: #83 (EXT-20). Baseline: MicrosoftTeams 8.1.0.

## Agent configurations and catalog (ConfigAPI)

CmdletConfig marks Get/New/Set/Remove-CsAgent and Get-CsAiAgents as AutoRest
cmdlets. The generated ConfigApiBasedCmdlets methods give the wire routes:

- AgentGet: GET `/Teams.VoiceApps/aiagents` with optional `First`, `Skip`,
  `ExcludeContent`, `NameFilter`, `TypeFilter`, `SortBy`, `Descending`.
- AgentGet1: GET `/Teams.VoiceApps/aiagents/{identity}`.
- AgentNew: POST `/Teams.VoiceApps/aiagents` with a JSON body.
- AgentSet: PUT `/Teams.VoiceApps/aiagents/{identity}` with a JSON body.
- AgentRemove: DELETE `/Teams.VoiceApps/aiagents/{identity}`.
- AiAgentGet: GET `/Teams.PlatformService/AiAgents` with `providerId`,
  `isTeamsIvrEnabled`, `displayNamePrefix`, `displayNameContains`,
  `maxResults`, `continuationToken`, `agentId`, `agentIds`, `showCount`,
  `channel`. Unbound values are omitted and switches use .NET `True`/`False`.

Model serializers give the property names:

- `AiAgentConfigurationDtoModel`: `ConfigurationId`, `Name`, `AgentId`,
  `AgentType`, `AgentTargetTagTemplateId`.
- Envelopes: `AIAgent` for create, get and update; `AIAgents` and
  `AgentCount` for the list.
- `AiAgentQueryResult`: `agents`, `continuationToken`, `agentCount`.
- `AiAgent`: `msaAppId`, `msaAppTenantId`, `botHandle`, `isTeamsIvrEnabled`,
  `providerId`, `publishState`, `cid` (Int64), `callingApiVersion`,
  `messagingApiVersion`, `channels` (string[]), `isTeamsVoiceEnabled`,
  `description`.
- `AiAgentDescription`: catalog fields including `displayName`, `agentId`,
  `agentType`, `starRating` (Double) and string arrays.

Custom wrappers (Merged_custom_PsExt.ps1):

- Get-CsAgent maps `-Id` to Identity and sends no list query. `-Force` is
  removed.
- New-CsAgent builds the DTO from `Name`, `AIAgentId` and `AIAgentType`. It adds
  `AgentTargetTagTemplateId` only when that parameter is bound.
- Set-CsAgent sends Identity = `Instance.Id` plus every instance field, so the
  PUT is a full replacement.
- Remove-CsAgent maps `-Id` to Identity.
- The PowerShell projection (`AIAgentConfiguration`) renames the wire fields
  to `Id`, `AIAgentId`, `AIAgentType` and `AIAgentTargetTagTemplateId`. The SDK
  keeps the wire names.

## AI-generated knowledge (MCPS)

Get-AIGeneratedKnowledgeContainer and Remove-AIGeneratedKnowledge have no
CmdletConfig entry. They are TeamsCmdlets (`Microsoft.Teams.PowerShell.TeamsCmdlets.dll`),
with mandatory `ThreadId` and optional `CallId`.

`TeamsCmdlet.PerformAuthorization` sets `McpsResource` to the fixed string
`https://teams.microsoft.com/api/mcps/df` and requests
`6bc3b958-689b-49f5-9006-36d165f30e00/.default` for every session and cloud.
The routes are:

- Get: `HttpUtilities.Get` of `{McpsResource}/mkb/crux/container-urls?threadId=…[&callId=…]`,
  which returns a list of strings. The cmdlet prints the strings in batches
  sized by the Purview import limit, joined with `;`.
- Remove: `HttpUtilities.Delete` of `{McpsResource}/mkb/crux?threadId=…[&callId=…]`.
  On success it prints a fixed message.

The SDK adds a closed `AIGeneratedKnowledge` backend and service profile with
this origin, path prefix and application-ID audience. Only the commercial cloud
is mapped, because no other cloud's resource is evidenced.

## Lab observations

These used client-secret, app-only Teams Administrator credentials.

- The agent list was `{"AIAgents":[],"AgentCount":0,"Diagnostics":[]}` before
  the lifecycle test ran.
- An unknown configuration on get, update or delete is a 404 with top-level
  `code` 100004.
- Create with an empty name is 400 `AIAgent_NameNullOrEmpty`. An unknown agent
  type is 400 `AIAgent_AgentTypeNotRecognized` (case-sensitive). An unknown tag
  template is 400 `AIAgent_InvalidTagsTemplateId`.
- Create with `MicrosoftCopilotStudio` and a reserved agent ID returns 201 with
  the configuration. The agent ID is not checked against the provider.
- Get, list and PUT return 200 with the `AIAgent`/`AIAgents` envelopes, and a
  rename is read back. Delete returns 200 `{"Diagnostics":[]}` and a later get
  is 404.
- Get-CsAiAgents with a non-GUID provider is 400 `BadArgument`
  (`error.code`). An unknown GUID provider gives
  `{"agents":[],"agentCount":0}`.
- MCPS issues an application token for its resource, but GET container-urls
  and DELETE crux both answer an empty 401. In the pinned module, the app-only
  certificate session fails before sending: `KeyNotFoundException` for the MCPS
  resource. Both probes used a fabricated thread ID
  (`19:msteams-sdk-nonexistent-0000@thread.v2`), so no meeting data could be
  read or removed.

## Acceptance

- `TestAccAIAgentReads` (read-only): list, a paged list, a missing get, and the
  catalog query and non-GUID rejection.
- `TestAccAIAgentRejectedWrites` (serial):
  - unknown-type create and update;
  - a missing delete;
  - the two app-only MCPS 401s.
- `TestAccAIAgentLifecycle` (mutation, proof `ai-agent-lifecycle`):
  - creates one journaled `msteams-sdk-…-agent`;
  - reads, lists, rejects an unknown tag template, renames and deletes it;
  - observes the 404;
  - compares the agent and the catalog count with pinned PowerShell
    (`read-ai-agent.ps1`).

  Recovery kind `ai-agent` resolves by the unique name and deletes only owned
  names.

Sanitized responses from these runs feed the unit contract tests.

## Unqualified

- Platform catalog entries. The lab has no published agents for a known
  provider, so the `AiAgent`/`AiAgentDescription` field types are covered by
  synthetic cases only.
- Successful knowledge reads and deletes. They need delegated credentials and a
  real meeting. `Remove-AIGeneratedKnowledge` is deliberately never run against
  real data.
- Associating an agent with a resource account (resource accounts cannot be
  created with app-only credentials; see CALL-27).
