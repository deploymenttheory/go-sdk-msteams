# Media logging policy

`TeamsAdminAPI.TeamsMediaLoggingPolicies` implements the five public pinned 8.1.0
`Get-CsTeamsMediaLoggingPolicy` and `Grant-CsTeamsMediaLoggingPolicy` parameter sets
through modern ConfigAPI discovery. The baseline has no New/Set/Remove for this
policy. The service provides `Global` (AllowMediaLogging=false) and `Tag:Enabled`
(AllowMediaLogging=true).

| Method | PowerShell set | Behavior |
|---|---|---|
| `GetByIdentity` | Get/Identity | GET one identity, escaped as a whole segment |
| `List` | Get/Identity (omitted) | Bare collection, stable case-insensitive Global first |
| `ListFiltered` | Get/Filter | Local Identity/Tag wildcards; blank selects all |
| `AssignToUser` | Grant/Identity | PATCH user policy; empty name clears the direct assignment |
| `AssignTenantDefault` | Grant/GrantToTenant | PATCH tenant policy with explicit forceSwitchPresent |
| `AssignToGroup` | Grant/GrantToGroup | POST group assignment, optional signed priority, 202 vs 204 receipt |
| `Assign` | all Grant sets | Exactly one scope; omits zero Rank; tenant Force false |

`Policy` projects Identity, Description and AllowMediaLogging. It keeps the
complete owned response in RawJSON/RawFields, including unknown or unprojectable
fields. PolicyName is a required string even though PowerShell metadata marks it
optional: the mapped helper rejects an unbound name. PassThru only affects
PowerShell output. Native methods return response metadata, and group grants also
return a receipt. A 202 or 204 confirms submission, not propagation. Writes never
replay and share the caller's context and identity-bound transport. Invalid
identities, ambiguous scopes, malformed filters and cancelled contexts fail before
any HTTP request.

Qualification (client-secret, Teams Administrator): the lab read Global and
Tag:Enabled, listed and filtered them and confirmed an HTTP404. It compared the
results with pinned PowerShell. User and group grants of `Enabled` to the dedicated
SDK fixtures were read back, compared with PowerShell and reversed through journaled
restoration. GrantToTenant, delegated modes, other clouds and propagation are
unqualified. See
[qualification](../../../docs/implementation/policies/media-logging-lab-acceptance.md).
