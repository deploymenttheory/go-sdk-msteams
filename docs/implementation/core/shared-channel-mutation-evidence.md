# Shared-channel relationship mutation evidence

Research date: 2026-11-21. CORE-24 / issue #190. Pinned management baseline
remains MicrosoftTeams 8.1.0. This note extends the earlier
[channel contract assessment](channel-contract-evidence.md#sharing-creation-and-update-assessed-not-declared-unavailable)
with first-party generated request construction. No tenant requests, permission
changes or service mutation experiments were performed.

## What the first-party clients establish

Microsoft publishes
[New-MgBetaTeamChannelSharedWithTeam](https://learn.microsoft.com/en-us/powershell/module/microsoft.graph.beta.teams/new-mgbetateamchannelsharedwithteam?view=graph-powershell-beta)
and [Update-MgBetaTeamChannelSharedWithTeam](https://learn.microsoft.com/en-us/powershell/module/microsoft.graph.beta.teams/update-mgbetateamchannelsharedwithteam?view=graph-powershell-beta).
Their generated descriptions and inputs establish client commands for navigation
creation/update, accepting a sharedWithChannelTeamInfo body. They belong to the
Graph module; the four pinned MicrosoftTeams shared-channel commands remain
Get-AssociatedTeam, Get-SharedWithTeam, Get-SharedWithTeamUser and Remove-SharedWithTeam.

The current Microsoft Graph beta Go SDK at commit
`088169b5c08cd4ce7295646c149a7d87fb62c039` also contains these builders:

| Builder | Verified construction |
| --- | --- |
| [Collection](https://github.com/microsoftgraph/msgraph-beta-sdk-go/blob/088169b5c08cd4ce7295646c149a7d87fb62c039/teams/item_channels_item_shared_with_teams_request_builder.go) | Post/ToPostRequestInformation select POST on `/teams/{team-id}/channels/{channel-id}/sharedWithTeams`, JSON body and Accept JSON; Send requests a sharedWithChannelTeamInfo result |
| [Item](https://github.com/microsoftgraph/msgraph-beta-sdk-go/blob/088169b5c08cd4ce7295646c149a7d87fb62c039/teams/item_channels_item_shared_with_teams_shared_with_channel_team_info_item_request_builder.go) | Patch/ToPatchRequestInformation select PATCH on the addressed sharedWithTeams item, JSON body and Accept JSON; Send requests the same result model |

Collection Git blob: `6929027d6f832e519c77d406aaca1c12aeb717eb`.
Item Git blob: `9187e767515003b8ea136a471f278899243edab0`.
The [base client](https://github.com/microsoftgraph/msgraph-beta-sdk-go/blob/088169b5c08cd4ce7295646c149a7d87fb62c039/graph_base_service_client.go)
selects beta (blob `88db42849f7113f3e22159b51f45dd5755f1a4fa`).
Construction accepts the broad generated model;
the inspected methods do not enforce service-specific input requirements or a
particular success status. Get/Delete have REST operation links in their comments;
the mutation comments give generic navigation-operation descriptions.

These facts strengthen the candidate wire contract. They establish executable
first-party request builders, beyond the PowerShell command names alone.

## Service contract that remains to be established

The current [REST resource reference](https://learn.microsoft.com/en-us/graph/api/resources/sharedwithchannelteaminfo?view=graph-rest-beta)
lists relationship read/list, unshare and allowed-member listing. It identifies
the resource ID as read-only. The read/unshare operation pages supply permissions
and service behavior for those methods. The inspected resource/index pages do
not supply a matching creation/update contract.

Consequently these details still need authoritative service evidence or a
controlled permitted lab observation:

- Fields required/accepted for linking an existing channel to another team,
  including team/tenant identity and read-only-looking model properties.
- The invitation, target-owner acceptance and cross-tenant policy prerequisites.
- Effective delegated/application permissions and supported cloud/access scope.
- Actual POST/PATCH success statuses, returned body/identity and error behavior.
- Which properties can change on an existing sharing relationship and how to
  observe or reverse that change without substituting a team rename.

Inference: generated navigation methods are insufficient to settle these service
requirements. The mismatch remains an evidence gap; it does not establish that
the operations are unsupported. Keep candidate Create/Update unresolved rather
than exposing guessed requests or generic unsupported stubs. The agreed true
CRUD SDK requirement remains open for this relationship until the contract is
settled. Channel creation and sharing an already existing channel remain distinct
operations.

CORE-24 can implement and test its four evidenced pinned commands and native
relationship reads/unsharing while this investigation continues. Its issue and
qualification state must retain the missing mutation evidence and unrun tenant
scenarios. Successful token acquisition or synthetic builder execution would not
close that gap.
