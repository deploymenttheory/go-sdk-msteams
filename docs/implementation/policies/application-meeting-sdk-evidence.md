# POL-21 application meeting configuration candidate

Issue #94 remains open. Native methods cover the two assigned commands and three
parameter sets. The ledger stays in-progress/metadata-only/not-run: successful
offline wire execution does not establish full exported-command or tenant parity.

| Command/set | Go mapping |
|---|---|
| Get-CsApplicationMeetingConfiguration / Identity | Service.GetByIdentity; Service.List for omitted Identity |
| Get-CsApplicationMeetingConfiguration / Filter | Service.ListFiltered |
| Set-CsApplicationMeetingConfiguration / Identity | Service.UpdateByIdentity; Service.UpdateGlobal for omitted Identity |

Pinned sources are [POL-01 research](federation-branding-voicemail-evidence.md#application-meeting-configuration-and-tenant-federation-settings),
the immutable package SHA256 `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`,
`Merged_custom_PsExt.ps1` Get/Set-CsConfigurationModern, generated
ConfigurationGetConfiguration/GetConfigurations/UpdateConfiguration, custom
Get/SetCsApplicationMeetingConfiguration and PolicyUtils.ValidateAndUpdatePsListModifier.
The selected path is the modern ConfigAPI adapter; no legacy TRPS path is claimed.

The new [offline capture](../../../tools/contracts/capture-application-meeting.ps1)
checks module 8.1.0 and private assembly SHA256
`8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6`.
It invokes three generated methods with an injected in-memory sender and eleven
pure list/model normalization cases. It does not process management commands,
connect, acquire tokens, discover endpoints or attach a network transport.
The [retained observations](../../../contracts/endpoints/8.1.0/fixtures/application-meeting-dispatch.json)
confirm GET item/list and sparse PUT/204; omission versus null/empty replacement,
Add/Remove operators, empty operators and legacy object cleanup. Replace hashtable,
numeric items and explicit PSListModifier inputs reject in the pinned helper;
the native typed API exposes replacement/delta constructors instead. Capture is
locally executed; Go CI compares this golden without rerunning this new oracle.

Resource tests exercise the root client with both authentication mode bindings,
one same-origin discovery, exact ConfigAPI audience/forest header, method/path,
body/status, identity escaping and raw nullable/future fields. Mutation denial and
429/503 cases use the default transport policy and assert a single attempt; no
write read-back or automatic compensation occurs. Context, invalid-input,
malformed-success, stable Global-first/local filtering and copied-input behavior
are covered. Shared wildcard tests qualify portable UTF-16/case behavior without
claiming all locale-specific PowerShell collation. Retained resource reports
contain structural data only and keep coverage_claim:false/live:not-run.

The [API documentation](../../../msteams/teams_admin_api/application_meeting_configuration/api_docs.md)
records the operations and absent Create/Delete contract. The documented
Global-only instance and the helper's open identity forwarding are distinguished;
the Go primitive requires an explicit identity or named UpdateGlobal operation.

`TestLabApplicationMeetingConfigurationReadOnly` is prepared and unrun. It checks
Global item/filtered-list observations and optionally compares Identity/string
AppIDs with the pinned certificate-authenticated PowerShell read. Its opt-in gate
runs before credential access; child output is bounded, transient and never
retained. The dedicated lab identity still needs the separately requested role
approval. Existing 403 observations for other ConfigAPI resources do not prove
the sole cause or this resource's authorization. No role/grant/browser changes
or tenant configuration writes were performed for this candidate.

Successful read/write permissions, modern flight/legacy eligibility, cloud
deployment, exported wrapper confirmation semantics and consequential update
read-back remain unqualified. Updating Global affects tenant meeting behavior;
live mutation acceptance requires an explicit snapshot/restore plan and owned
application IDs. This candidate prepares only read acceptance. The common lab
flags, private identity/module paths and optional PowerShell oracle flag from
[lab lifecycle documentation](../lab/lifecycle-evidence.md) apply.
