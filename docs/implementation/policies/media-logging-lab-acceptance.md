# Media logging policy qualification

Owner: #126 (POL-53). Baseline: MicrosoftTeams 8.1.0.

The public `Get-CsTeamsMediaLoggingPolicy` and `Grant-CsTeamsMediaLoggingPolicy`
Functions are ConfigAPI remoting proxies (`CmdletConfig.json`: ModernCmdlet
`Get-CsConfigurationModern` with ConfigType `TeamsMediaLoggingPolicy`, and
`Grant-CsTeamsPolicy` with PolicyType `TeamsMediaLoggingPolicy`;
`reconciliation.json` records 154 and 280). The baseline exports no New, Set or
Remove command for this policy, so the service-provided `Global`
(AllowMediaLogging=false) and `Tag:Enabled` (AllowMediaLogging=true) instances are
the only policies.

## Wire contract

- Get, Identity and Filter sets (`runtime.json` command 131): `GET
  /Skype.Policy/configurations/TeamsMediaLoggingPolicy[/configuration/{id}]`
  after Teams service discovery. The SDK escapes the identity as one whole segment,
  orders a case-insensitive `Global` first and applies Identity/Tag wildcard
  filtering locally. A blank filter selects the full collection, as in the mapped
  modern helper. Unknown and unprojectable fields stay in RawFields/RawJSON.
- Grant (`runtime.json` command 249) dispatches through the merged
  `Grant-CsTeamsPolicy` helper:
  - Identity: `PATCH /Skype.Policy/users/{id}/policies/TeamsMediaLoggingPolicy`
    with `{"policyName":...}`. The helper rejects an unbound PolicyName, so the SDK
    takes a required string; an empty name clears the direct assignment.
  - GrantToTenant (`-Global`): `PATCH /Skype.Policy/tenants/policies/TeamsMediaLoggingPolicy`
    with `policyName` and `forceSwitchPresent`. The public proxy does not expose
    Force: `Assign` sends false and `AssignTenantDefault` keeps the low-level flag.
  - GrantToGroup: `POST /Skype.Policy/groupPolicyAssignments/{group}/policyTypes/TeamsMediaLoggingPolicy`
    with `policyName` and optional signed `priority`. The helper drops a zero Rank,
    which `Assign` mirrors; `AssignToGroup` can send zero explicitly. A 202 receipt
    is kept separate from 204, and neither is treated as member propagation.
- PassThru and MsftInternalProcessingMode concern PowerShell output and dispatch
  only. Mutations never replay.

These routes match the qualified ConfigAPI grant plumbing used by the
application-access, room-video, Cortana and upgrade policy services.

## Live results (2026-11-23, client-secret, Teams Administrator)

- Reads: Global and Tag:Enabled by identity, the full list, the `Enabled` Tag
  filter and an HTTP404 for an absent identity. Selected and listed raw fields
  agreed, and both instances' typed getters matched the pinned PowerShell function.
- User grant: the dedicated SDK user had no direct media logging assignment. The
  empty baseline was journaled before granting `Enabled`. Read-back through
  `PolicyAssignments.GetUser` converged, then the clear form restored the empty
  state, verified independently. The pinned module compared before, assigned and
  restored states. User grant and clear both returned 200 with no body.
- Group grant: the dedicated empty security group had no assignments. Its empty
  baseline was journaled, then `Enabled` was granted at rank 1, which returned 204.
  Get/list read-back agreed. The assignment was removed with
  `PolicyAssignments.DeleteGroup`, then absence (GET404 and list) was verified. The
  pinned module compared all three states.
- Recovery kinds `media-logging-user-assignment` and `media-logging-group-assignment`
  are bound to the private fixtures and registered for journal recovery. Every
  journal entry was cleaned. No licence or other tenant setting was touched.

Sanitized actual responses (`get_global`, `get_tag_enabled`, `list`,
`get_missing_404`, `assign_user`, `clear_user`, `assign_group_204`) feed the unit
tests. The built-in `Tag:Enabled` identity is a public service name and is kept
verbatim; tenant-derived values are redacted.

## Not qualified

GrantToTenant was not run live because it changes the default for every user
without a direct assignment. It is contract-tested only. Delegated modes, other
clouds, group-to-member propagation, effective client media logging and the 202
accepted-operation path are also unqualified.
