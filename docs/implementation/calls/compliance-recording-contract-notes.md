# Compliance recording contracts

Owner: #73 (CALL-25). Baseline: MicrosoftTeams 8.1.0.

The policy and application cmdlets are compiled `Microsoft.Teams.Policy.
Administration.Cmdlets.Core` cmdlets on the PolicyRp provider. They use the
shared compiled transport (`internal/policyrp`): MPACmdlet, X-MS-CmdletName
and correlation headers, no forest discovery and no write replay.

- Policies: `Skype.Policy/configurations/TeamsComplianceRecordingPolicy`
  `[/configuration/{identity}]`. GET, POST to the collection, PATCH, DELETE.
- Applications: `Skype.Policy/configurations/TeamsComplianceRecordingApplication`
  `[/configuration/{policy}/{id}]`. The resource name was settled live: the
  provider literal `ComplianceRecordingApplication` and the plural are 400
  40001 ("Provided type … is not valid"); `TeamsComplianceRecordingApplication`
  returns the collection.
- Grant uses the modern ConfigAPI grant shared with other policies
  (`policyType` `TeamsComplianceRecordingPolicy`).
- `New-CsTeamsComplianceRecordingPairedApplication` derives from
  `BaseTpmNewElementCmdlet`: it builds an element locally and sends nothing.

Model facts from the module:

- Policy: Identity, ComplianceRecordingApplications, Enabled,
  WarnUserOnRemoval, DisableComplianceRecordingAudioNotificationForCalls,
  Description, RecordReroutedCalls, CustomPromptsEnabled,
  CustomPromptsPackageId, CustomBanner.
- Application: Identity, ComplianceRecordingPairedApplications, Id,
  RequiredBeforeMeetingJoin, RequiredBeforeCallEstablishment,
  RequiredDuringMeeting, RequiredDuringCall, ConcurrentInvitationCount (uint).
  The Id setter requires a GUID.
- `-Priority` is a declared application parameter but the cmdlet never copies
  it into the model, so it has no wire effect and no SDK field.
- The list parameters are PSListModifiers. The SDK applies Replace/Add/Remove
  to the value read before the PATCH (`ListEdit`).

Lab observations (client-secret, app-only Teams Administrator):

- Global reads with every setting at its default and no applications. The
  application collection is empty.
- Policy create/read/update/list/delete of an owned `msteams-sdk-*` policy
  succeeds; the deleted policy is 404 / 40407. PowerShell 8.1.0 reads agree
  with the SDK projection after create and after update.
- Application create with an Id that is not a resource account is 400 / 40013
  `ErrorComplianceRecordingApplicationIdentityNotAValidAppInstance`, with or
  without paired applications. Paired applications together with
  ConcurrentInvitationCount 2 is 400 / 40013
  `ErrorComplianceRecordingPairedApplicationAndConcurrentInvitationBothSet`.
- Get, Set and Remove of a missing application are 404 / 40407 in both
  PowerShell and the SDK.

Not qualified:

- A successful application create, update or remove. It needs a resource
  account, and app-only credentials cannot create one (see
  resource-accounts-contract-notes.md). The SDK request shapes are covered by
  contract tests; the success responses are not.
- Live grants. A compliance recording grant changes how a user's calls are
  handled, so the lab never grants these policies; the grant wire shapes are
  covered by contract tests against the shared modern grant.
- Custom prompt packages and custom banners with real content.
