# Upgrade configuration

`TeamsAdminAPI.TeamsUpgradeConfigurations` implements the three public pinned
8.1.0 Get/Set parameter sets through compiled PolicyRP. GetByIdentity, List and
ListFiltered own complete response snapshots. UpdateByIdentity composes Set's
preparation GET and sparse PATCH, retaining the read receipt and write acceptance;
PatchByIdentity submits only selected settings directly. No write replays.

DownloadTeams and BlockLegacyAuthorization are nullable booleans;
SfBMeetingJoinUx is a nullable opaque string. `policy.Field` preserves omitted,
null, false and empty values. The inspected source imposes no local enums.
Controls use the shared compiled Identity/Force encoding. There are no public
create, delete or assignment commands for this configuration.

The existing lab Teams Administrator application passed certificate/client-secret
Global Get/List and pinned comparisons. A journaled DownloadTeams toggle verifies
all getters, omitted settings, three pinned comparisons and independent original
restoration. Bound recovery restores only Global DownloadTeams. Permanent actual
response fixtures supply unit read, model, filter and sparse update replay; fault
injection remains separate from live outcomes. Other field persistence, delegated
flows and other clouds have not been qualified. Source notes are in
`docs/implementation/policies/upgrade-contract-notes.md`.
