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

## EXT-90 continuation: superseding source review

The portable [primary packet](evidence/ai-agents-source/primary.json) and
[SHA manifest](evidence/ai-agents-source/manifest.json) bind the actual pinned
8.1.0 assemblies, fully qualified types/method bodies, inert observations and
embedded reproducers. They are source/data evidence, not a management lifecycle
or tenant result. The primary packet remains immutable. Its initial prose
requiring an object for Remove is superseded by the
[ordering supplement](evidence/ai-agents-source/ordering-supplement.json):
empty/whitespace/array roots yield a null generated result; literal null and
numeric scalars fail. Remove still declares exactly 200.

The same supplement records 81 ordering cases (56 exact duplicate exceptions,
25 case-distinct projections), lazy default/null-row distinctions, own private
command headers and safe query formatting. The
[array supplement](evidence/ai-agents-source/final-arrays-supplement.json)
separates four generated nullable string arrays from the MCPS JsonConvert
string reader. The [numeric supplement](evidence/ai-agents-source/numeric-supplement.json)
binds Int64 configuration counts, Int32 catalog counts and raw integer lexeme
parsing; [Double edges](evidence/ai-agents-source/double-edge-supplement.json)
record source overflow to infinity and underflow to zero.

The unreleased canonical API corrections are configuration count `*int64`,
nullable generated string-array entries, and raw MCPS `[]*string`. Public
configuration projection fails on null/nonmodel rows; it must not inherit the
catalog's nullable row behavior. Generated DTOs have exact-case/type semantics,
recursive exact-duplicate rejection and lazy description defaults. They do not
share PolicyRp coercion. New omits an unbound target, whereas the actual Set
wrapper binder sends all four strings and converts null instance fields to
empty strings. Source-success statuses are New 201, Get/Set/catalog/Remove 200;
MCPS helpers accept any 2xx before their independent parse/presentation path.

The source-generated config headers are private default names, resolved through
each own proxy/wrapper/pipeline chain; MCPS uses a different factory and public
names. Existing captured tenant JSON remains response-only evidence for the
corrected requests. Source diagnostic/output formatting, trusted paths, native
replay/context controls, commercial-only MCPS mapping and raw-field enrichment
are explicitly distinct boundaries. No seven-command ledger or live status is
promoted by this first repair batch. Actual Go execution is recorded separately
when the frozen affected packages have completed.

The first frozen repair revision `b838d0f` completed the full affected SDK,
AI lab and EXT integration `TestUnit|Example` selection: all three package
terminals passed with 158 test-pass events and no failures/skips. No TestAcc
executed. The [bounded execution record](evidence/ai-agents-b838d0f-check.json)
contains the exact command, file hashes and raw Go JSON SHA. This is an
implementation check, not a seven-command closure decision or a live claim.
