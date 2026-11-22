# VDI and virtual appointments contracts

Owners: #152 (POL-79), #153 (POL-80). Baseline: MicrosoftTeams 8.1.0.

Checked runtime metadata resolves all ten assigned public commands as compiled
Cmdlets. Each family has Get (Identity/Filter), New/Set/Remove (Identity) and Grant
(Identity error without resource HTTP, GrantToUser, GrantToTenant and GrantToGroup): eighteen parameter
sets total. The existence of proxy functions elsewhere in the module must not
override the actual loaded command resolution. CmdletConfig has no modern mapping
for these two resource families.

Inspected Core models expose nullable DisableCallsAndMeetings and
DisableAudioVideoInCallsAndMeetings booleans plus opaque nullable VDI2Optimization
for VDI, and a single nullable EnableSmsNotifications boolean for virtual
appointments. The inspected property setters execute no enums, ranges or
cross-field validation. The New/Set classes copy only bound properties. Shared New reads Global for
ShouldProcess display without copying its settings; native Create sends only the
selected fields. Shared Set rejects empty settings, prepares the existing object
and sparsely PATCHes bound properties. Force is a compiled request control.

The inspected BaseTpmGrantCmdlet default Identity path reports Cannot determine
which grant to execute; it is not a successful no-op. The SDK rejects missing or
mixed scopes before token acquisition and preserves explicit user/tenant/group
scopes. Shared compiled request shapes include null additionalParameters and
required signed group rank, with no optional-rank modern dispatch here.

Both certificate and client-secret lab modes passed Global Get/List and owned
unassigned CRUD for both services. VDI created both disable flags true with
VDI2Optimization Disabled, then sparsely changed optimization to Enabled while
preserving both flags. Virtual appointments created SMS false and updated true.
Each test checks exact sparse request fields, all selected getters, two pinned
PowerShell comparisons, list membership and independent deletion/GET404/list
absence. Eighteen permanent actual sanitized responses, including empty accepted
bodies, directly feed unit replay. No existing policy, Global setting or assignment
was changed. Recovery has bound vdi-policy and virtual-appointments-policy handlers.

Grant wire/error behavior has unit coverage; successful user/tenant/group
persistence and clear remain unqualified. Delegated tenant flows and other clouds
also remain separate. The existing lab Teams Administrator application and repo
secrets suffice for the completed reads/CRUD; no privileges were added.

Sources: pinned runtime.json; Core model/New/Set/Get/Remove/Grant classes for both
families; Providers BaseTpmNew/Set/GrantCmdlet and existing compiled HTTP helpers,
all from checksum-verified 8.1.0 assemblies. Current Microsoft references support
the chosen VDI optimization values and SMS example, while validation follows the
pinned source rather than newer documentation:
- https://learn.microsoft.com/en-us/powershell/module/microsoftteams/set-csteamsvdipolicy?view=teams-ps
- https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamsvirtualappointmentspolicy?view=teams-ps
