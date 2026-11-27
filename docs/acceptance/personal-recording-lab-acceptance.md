# Personal attendant and recording rollout qualification

Owners: #140 (POL-67) and #141 (POL-68). Baseline: MicrosoftTeams 8.1.0.

Both resources are public compiled Core cmdlets. Each assigns Get Identity/Filter,
New/Set/Remove Identity, and GrantToUser/GrantToGroup/GrantToTenant/Identity. The
shared BaseTpm* and PolicyRP provider establish direct GET/POST/PATCH/DELETE,
first-slash scoped routes, sparse field binding, optional Force string and compiled
grant bodies. The empty Grant Identity set has no source operation; an explicit
Go grant scope is required.

Personal attendant has eight nullable opaque string setters/getters; recording
rollout has one (MeetingRecordingOwnership). There are no source enum validations.
Local primary evidence is the pinned models, all implementing cmdlets and provider
methods. Read oracles verify Core SHA256
68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49 and PolicyRP
SHA256 a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51.

On 2026-11-22 the dedicated Teams Administrator SDK SPN read/listed Global and
matched the pinned getters. The personal-attendant Global values were all
EnabledUserOverride. Its owned CRUD scenario creates AutomaticRecording=Disabled
and AutomaticTranscription=EnabledUserOverride, updates only transcription to
Disabled, checks recording is preserved and compares all eight getters twice.
The actual response uses an unprefixed identity; the SDK and unit fixtures retain
that identity rather than inserting Tag:. Its explicit Tag: filter therefore
correctly selects no owned resource in this fixture.

Recording rollout creates an owned policy with MeetingInitiator, updates it to
MeetingOrganizer, verifies both persisted states with SDK reads and two pinned
PowerShell comparisons, and checks actual sparse request bytes contain exactly
Identity and MeetingRecordingOwnership. Both scenarios list, delete and require
independent typed GET404 plus complete collection absence. Both application auth
modes passed CRUD. Creates are durably journaled beforehand and bound recovery
handlers register both owned policy kinds. No Global policy, grant, user/group,
license or privilege is changed.

Eighteen permanent sanitized response files come directly from these tests.
Typed fields, complete snapshots, filters, wire dispatch, actual error envelopes
and mutation receipts are exercised through those unit fixtures. Request-only
vectors and controlled failures test null/value binding, context, invalid input,
grant selection and no-replay behavior separately from successful tenant results.
Examples compile; each public parameter set maps to symbols and tests.

[Microsoft's personal-attendant reference](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/set-csteamspersonalattendantpolicy?view=teams-ps)
states that only recording/transcription controls remain configurable during the
documented temporary restriction; the other six were left at defaults in the test.
This does not introduce artificial SDK validation. Current recording-rollout
reference URLs redirect; successful API read-back is not a claim about effective
meeting storage behavior.

Remaining: personal-attendant's other six live setting changes, grant/clear
persistence and propagation, null/empty persistence, effective attendant/recording
behavior, delegated modes, locale-specific filters and other clouds. Merged code
and selected-operation checks remain separate from whole-ticket/release coverage.
