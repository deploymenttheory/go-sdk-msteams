# Audio conferencing policy: selected modern contract

POL-34 / issue #107, MicrosoftTeams 8.1.0, checked 2026-11-21. This source
checkpoint establishes selected modern REST behavior and makes no complete
command or live qualification claim.

## Exported implementation and selection

All five public commands are exported proxy functions delegating to private
`Microsoft.Teams.ConfigApi.Cmdlets.*CsTeamsAudioConferencingPolicy` classes,
which inherit `TrpsPSCmdlet`. Their normalized runtime surface has eight sets:
Get Filter/Identity, Grant GrantToGroup/GrantToTenant/Identity, and New/Remove/Set
Identity. Grant Identity selects a user assignment here. The unrelated Core
classes have a different parameter surface and are not the selected export.

`custom/CmdletConfig.json` maps Get/New/Set/Remove to the corresponding
`*-CsConfigurationModern` function with ConfigType
`TeamsAudioConferencingPolicy`; Grant maps to `Grant-CsTeamsPolicy` with the
same PolicyType. All entries are Remoting and advertise TeamsGA/TeamsPreview/
SelfHost/Torus. The private New/Set classes call
`PolicyUtils.ValidateAndUpdatePsListModifier` for MeetingInvitePhoneNumbers.
They expose a bound Boolean AllowTollFreeDialin, not the nullable Core setter.

The ordinary modern-selection boundary, command-specific flights, discovery,
audience and session callback are shared with
[the traced application access chain](application-access-contract-evidence.md)
and [routing evidence](../core/policy-routing-evidence.md). No flight result is
observed for this family. A native REST service must not infer that a Get flight
establishes mutation availability or silently invoke a PowerShell/TRPS fallback.

## Provenance and reproduction

Pinned package SHA256:
`78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`.
Private assembly SHA256:
`8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6`.
Wrapper source SHA256:
`03dbc3a85e12c98dd1807d604782c6a90d0656569541cd9540738825a7fcb255`.
Command configuration SHA256:
`754d149c1a5904158e9f01eb55ba44d15c17b7e66f7556385ea894f4ff84d906`.

These scripts require explicit module/output paths and reject an existing
output. They check the manifest and applicable source/assembly hashes:

- [Wrapper capture](../../../tools/contracts/capture-audio-conferencing-policy-wrappers.ps1)
  executes unmodified mapped function ASTs with synthetic telemetry types and an
  in-memory internal-command module. Formatting adapters return objects
  unchanged. It also checks every mapping against the pinned configuration.
- [List capture](../../../tools/contracts/capture-audio-conferencing-policy-models.ps1)
  invokes actual PolicyUtils normalization and XdsConfiguration serialization
  on synthetic phone-list property bags.
- [Request capture](../../../tools/contracts/capture-audio-conferencing-policy-requests.ps1)
  invokes actual generated ConfigApiBasedCmdlets helpers with an injected memory
  sender that refuses credential headers. There is no HttpClient or production
  pipeline; generated URI construction precedes session host/header rewriting.

No exported command processor, flight lookup, credentials, discovery or tenant
request is executed. No Microsoft binary or decompiled source is redistributed.
The private proxy classes, generic wrapper, SetCsConfiguration_Update generated
dispatch, PolicyUtils and generated CRUD/grant request methods were inspected
in the pinned assembly. Mutable current documentation is not substituted for
the baseline package.

## Captured behavior

The fixtures contain
[22 wrapper cases](../../../contracts/endpoints/8.1.0/fixtures/audio-conferencing-policy-wrappers.json),
[15 normalization cases](../../../contracts/endpoints/8.1.0/fixtures/audio-conferencing-policy-models.json)
and [15 generated cases](../../../contracts/endpoints/8.1.0/fixtures/audio-conferencing-policy-requests.json).
Five wrapper cases fail as expected (create failure, missing PolicyName/scope,
and bound PassThru true/false); four list inputs are rejected. All generated
cases send once, except pre-cancellation, which sends zero times.

| Operation | Selected behavior |
| --- | --- |
| List | GET `/Skype.Policy/configurations/TeamsAudioConferencingPolicy`, bare array |
| Get | GET collection + `/configuration/{identity}`, identity escaped as one complete segment |
| Create | POST collection, 201; mapped New then reads using Get, stopping after failed POST |
| Update | Sparse **PUT** item, 204, including Identity; no preparatory GET |
| UpdateGlobal | Mapped Set supplies Global only when Identity was omitted |
| Delete | DELETE item, 204, **no body** |
| Assign user | PATCH `/Skype.Policy/users/{user}/policies/TeamsAudioConferencingPolicy`, 200 |
| Assign tenant | PATCH `/Skype.Policy/tenants/policies/TeamsAudioConferencingPolicy`, 200 |
| Assign group | POST `/Skype.Policy/groupPolicyAssignments/{group}/policyTypes/TeamsAudioConferencingPolicy`, 202 or 204 |

Optional apiVersion/schemaVersion queries are absent. The ordinary generated
pipeline discovers `/Teams.Tenant/serviceDiscovery`, keeps ConfigAPI as the
credential-bearing destination, and adds the normalized HTTPS forest root as
X-MS-Target-Uri. Command/correlation headers and the Teams audience belong to
the shared client. There is no unconditional MPACmdlet header. All mutations
must remain single-send and must not infer completion from group 202 acceptance.

Get stably partitions case-insensitive exact Global identities first, then
filters locally against Identity or Tag:Filter. Whitespace Filter/Identity in
the mapped helper selects the collection; public nonempty-argument binding is
separate. New's post-success GET is an explicit composition, with successful
submission retained if read-back fails. Sparse PUT omission is evidenced on the
wire; server-side replacement/default rules are not inferred from its verb.

Phone-list omission stays omitted. A scalar becomes a replacement array; null
and empty replacement become `[]`. Add/Remove remain server operators; empty
operators produce `{}`. These are not local compiled-provider list edits.
Replace inside a hashtable and a PSListModifier instance are rejected, as are
numeric/null list items. Empty/whitespace/future string values serialize without
a phone-number regex. Accepted legacy objects lose only the source metadata
keys; acceptance by this helper is not proof that a phone resource accepts an
arbitrary object. The generic model preserves unknown fields and wire presence.

Grant requires a bound PolicyName, but explicit null string binding becomes
empty, which is the captured clear form. Missing user identity requires Global;
an empty/mixed native scope must fail locally. Group Rank nil/zero is omitted
by the wrapper; positive and negative nonzero values are forwarded. The native
group primitive may expose explicit priority zero, which the generated DTO
serializes distinctly. Public PassThru is not accepted by the mapped modern
Grant function, including bound false; this is a recorded baseline failure.

## Outstanding qualification

Native implementation, root wiring, examples, exact ledger mappings and opt-in
lab checks follow this checkpoint. No successful tenant read, mutation,
read-back, cleanup, effective meeting invite change, privilege grant or pinned
PowerShell/Go tenant comparison is claimed. Auth, role/cloud/flight deployment,
server-side operators and omission semantics require separate qualification.
