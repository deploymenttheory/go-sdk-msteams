# Meeting broadcast policy qualification

Owner: #128 (POL-55). Research owner: POL-05. Baseline: MicrosoftTeams 8.1.0.

## Contract

All five public commands are remoting Functions. Their `custom/CmdletConfig.json`
entries (reconciliation records 157, 282, 417, 527 and 667) select the modern
ConfigAPI helpers with `ConfigType` `TeamsMeetingBroadcastPolicy`
(Get/New/Set/Remove-CsConfigurationModern) and Grant-CsTeamsPolicy with
`PolicyType` `TeamsMeetingBroadcastPolicy`. This is the same dispatch as room
video teleconferencing and network roaming policy, so the SDK uses the shared
identity-bound ConfigAPI adapter (`internal/configapi`) with forest discovery,
not the compiled PolicyRP provider. No shared SDK package changes.

The checked runtime metadata (`runtime.json` commands 134, 251, 368, 476 and
606) and the pinned public proxy agree: Get has Filter/Identity sets; Grant has
GrantToGroup (mandatory Group, nullable Int32 Rank), GrantToTenant and the
Identity (user) set; New/Set/Remove each have one Identity set. New and Set bind
Description, BroadcastAttendeeVisibilityMode and BroadcastRecordingMode as
`System.String` and AllowBroadcastScheduling/AllowBroadcastTranscription as
non-nullable `System.Boolean`, with no validation attributes. The Set helper
defaults an omitted Identity to Global (`UpdateGlobal`).

Wire shape: `/Skype.Policy/configurations/TeamsMeetingBroadcastPolicy` collection
GET, `.../configuration/{identity}` GET, POST/201, sparse PUT/204 and bodyless
DELETE/204. Grants use the same per-scope modern routes as the other ConfigAPI
policies: user PATCH `/Skype.Policy/users/{id}/policies/{type}`, tenant PATCH
`/Skype.Policy/tenants/policies/{type}` and group POST
`/Skype.Policy/groupPolicyAssignments/{id}/policyTypes/{type}`.

## Lab results

On 2026-11-23 both certificate and client-secret application modes passed:

- `TestAccMeetingBroadcastPolicyReadOnly`: Global read, full list (Global first,
  identical raw fields) and the local `Global` filter, plus all five public
  getters compared with pinned PowerShell.
- `TestAccMeetingBroadcastPolicyCRUD`: created an owned, journaled
  `msteams-sdk-<run>-meeting-broadcast-policy` with `AllowBroadcastScheduling=false`
  and `BroadcastRecordingMode=AlwaysDisabled` (Global is `true`/`AlwaysEnabled`),
  read it back, sparsely updated only Description and
  `BroadcastAttendeeVisibilityMode=InvitedUsersInCompany`, required every other
  returned field to stay unchanged, compared with PowerShell after create and
  after update, listed, deleted and independently required GET 404 and list
  absence. Both journal entries ended `cleaned`.

No Global settings, existing policies or assignments were changed. The tests do
not grant the policy, so there are no licensing implications.

Nine sanitized response files from the client-secret run supply the unit tests.
Request shapes for every grant scope, group 202/204 receipts, error statuses,
no-replay mutations, malformed reads, partial create read-back, nil and cancelled
contexts and invalid identities are contract-tested.

## Not yet qualified

- Grant persistence for user, group and tenant scopes (contract-tested only).
- AllowBroadcastTranscription and the remaining mode values on a live write;
  explicit-empty string persistence.
- Effective live-event behavior, delegated credentials and other clouds.

Current Microsoft reference:
[New-CsTeamsMeetingBroadcastPolicy](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamsmeetingbroadcastpolicy?view=teams-ps).
