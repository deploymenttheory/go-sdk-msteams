# IP phone policy qualification

Owner: #159 (POL-86). Baseline: MicrosoftTeams 8.1.0.

## SDK behaviour

All five public commands are remoting Functions. Their `custom/CmdletConfig.json`
entries (reconciliation records 152, 278, 412, 524 and 663) select the modern
ConfigAPI helpers with `ConfigType` `TeamsIPPhonePolicy`
(Get/New/Set/Remove-CsConfigurationModern) and Grant-CsTeamsPolicy with
`PolicyType` `TeamsIPPhonePolicy`. This is the same dispatch as room video
teleconferencing and network roaming policy, so the SDK uses the shared
identity-bound ConfigAPI adapter (`internal/configapi`) with forest discovery,
not the compiled PolicyRP provider. No shared SDK package changes.

The checked runtime metadata (`runtime.json`) and the pinned public proxy agree:
Get has Filter/Identity sets; Grant has GrantToGroup (mandatory Group, nullable
Int32 Rank), GrantToTenant and the Identity (user) set; New/Set/Remove each have
one Identity set (mandatory on New, optional on Set). New and Set bind
Description, AllowBetterTogether, AllowHomeScreen, SearchOnCommonAreaPhoneMode
and SignInMode as `System.String`, AllowHotDesking as non-nullable
`System.Boolean` and HotDeskingIdleTimeoutInMinutes as non-nullable
`System.Int64`, with no validation attributes. The Set helper defaults an
omitted Identity to Global (`UpdateGlobal`).

Wire shape: `/Skype.Policy/configurations/TeamsIPPhonePolicy` collection GET,
`.../configuration/{identity}` GET, POST/201, sparse PUT/204 and bodyless
DELETE/204. Grants use the same per-scope modern routes as the other ConfigAPI
policies: user PATCH `/Skype.Policy/users/{id}/policies/{type}`, tenant PATCH
`/Skype.Policy/tenants/policies/{type}` and group POST
`/Skype.Policy/groupPolicyAssignments/{id}/policyTypes/{type}`. The timeout is
returned as a JSON number.

## Lab results

On 2026-11-23 both certificate and client-secret application modes passed:

- `TestAccIPPhonePolicyReadOnly`: Global read, full list (Global first,
  identical raw fields) and the local `Global` filter, plus all seven public
  getters compared with pinned PowerShell.
- `TestAccIPPhonePolicyCRUD`: created an owned, journaled
  `msteams-sdk-<run>-ip-phone-policy` with `SignInMode=CommonAreaPhoneSignIn`,
  `AllowHotDesking=false` and `HotDeskingIdleTimeoutInMinutes=30` (Global is
  `UserSignIn`/`true`/`120`), read it back, sparsely updated only Description,
  `AllowHomeScreen=Disabled` and `AllowBetterTogether=Disabled` (Global is
  `EnabledUserOverride`/`Enabled`), required every other returned field to stay
  unchanged, compared with PowerShell after create and after update, listed,
  deleted and independently required GET 404 and list absence. Both journal
  entries ended `cleaned`.

No Global settings, existing policies or assignments were changed. The tests do
not grant the policy, so there are no licensing implications.

Nine sanitized response files from the client-secret run supply the unit tests.
Request shapes for every grant scope, group 202/204 receipts, error statuses,
no-replay mutations, malformed reads, partial create read-back, nil and cancelled
contexts and invalid identities are unit-tested.

## Not yet qualified

- Grant persistence for user, group and tenant scopes (unit-tested only).
- SearchOnCommonAreaPhoneMode and the remaining mode values on a live write;
  explicit-empty string persistence and out-of-range timeouts.
- Effective device behavior, delegated credentials and other clouds.

Current Microsoft reference:
[New-CsTeamsIPPhonePolicy](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamsipphonepolicy?view=teams-ps).
