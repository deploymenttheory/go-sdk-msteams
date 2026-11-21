# POL-27 policy assignment contract investigation

Issue #100 covers fourteen commands. Six appear in the default, unauthenticated
8.1.0 runtime snapshot; eight are conditional OCE declarations. Neither a missing
default export nor an OCE declaration removes a command from SDK scope. This is a
source investigation checkpoint, not a completed service, contract pass or live
qualification. The immutable baseline and coverage states are unchanged.

## Pinned inputs

The package is the accepted 8.1.0 archive, SHA256
`78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`.
Relevant packaged file SHA256 values:

| File | SHA256 |
|---|---|
| MicrosoftTeams.psm1 | `f2fc41164c72da25065dc28543eec89afda0a761d58ab44eb1fdc8a3f82cf990` |
| MicrosoftTeams.psd1 | `3a563674a6482c4fa2f8cb982085574e01be726949a8058178564193f11ffa2e` |
| netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.OCE.dll | `de6da2b13b4b675d7956a0c57af53fccc6cf457a3de2e8464273f95103e3dfc5` |
| net472/Microsoft.Teams.Policy.Administration.Cmdlets.OCE.dll | `ea52b14ae69bfeb7ce63497176a81252dd7c0fa17ad93d6b8ae49ef358091f03` |
| netcoreapp3.1/CmdletSettings.json and net472/CmdletSettings.json | `c0549eba3a249ef431b0d0c61ee232c815b74fe0cfa1b41b0860a8531f33e6dd` |

Assembly inspection uses ILSpy locally, without importing a privileged module,
constructing commands, calling ProcessRecord, acquiring tokens or making requests.
Source facts below need executable metadata/request captures before ledger mapping.
Microsoft binaries and decompiled source are not redistributed here.

## Ordinary generated commands

The accepted runtime snapshot establishes these exact command/set identities:

| Command | Sets |
|---|---|
| Get-CsBatchPolicyAssignmentOperation | Get, Get1 |
| Get-CsGroupPolicyAssignment | Get, Get1, Get2 |
| Get-CsUserPolicyAssignment | Get, GetViaIdentity |
| New-CsBatchPolicyAssignmentOperation | NewExpanded |
| New-CsGroupPolicyAssignment | NewExpanded |
| Remove-CsGroupPolicyAssignment | Remove |

Group lookup distinguishes group plus policy type, group-only and policy-type-only.
User lookup has a string identity and an InputObject variant. Batch creation carries
an identity array, policy type/name, operation name and additional parameters. Group
creation carries rank and PassThru; removal also carries PassThru. Generated DTO
presence, callbacks, wrapper coercions and response projections must be inspected
separately. Existing voicemail operation reads do not qualify these owned commands.

## Conditional OCE commands

The root manifest groups all eight names under its MPA OCE comment. The root loader
selects OCE when MSTeamsContextInternal is IsOCEModule; even its Core branch points
to the packaged net472 OCE assembly. The inspected netcoreapp3.1 sibling contains
the following Cmdlet/Parameter attributes. Cross-target equivalence remains to be
checked before treating these as the loader's executable signatures.

| Command | Declared command-specific parameters / sets |
|---|---|
| Get-CsBatchOperationDefinition | OperationId Guid required; sets GetBatchOperationDefinition and GetBatchOperationDefinitionForRegion, with Region required in the latter |
| Get-CsBatchOperationStatus | GetAllBatchOperationStatus default; GetBatchOperationStatus requires OperationId; GetBatchOperationStatusForRegion requires OperationId and Region; GetAllBatchOperationStatusForRegion requires Region |
| Get-CsGroupPolicyAssignments | GetAllGroupPolicyAssignments default; GetGroupPolicyAssignment requires GroupId and PolicyType; GetGroupPolicyAssignments requires GroupId; GetPolicyAssignmentForGroups has optional PolicyType |
| Get-GPAGroupMembers | Required GroupId Guid, optional All switch |
| Get-GPAUserMembership | Required Identity Guid |
| Invoke-CsRefreshGroupUsers | Required GroupId Guid and TenantRegion; optional UserId Guid and PolicyType |
| Invoke-CsReprocessBatchOperation | Required OperationId Guid and Region |
| Invoke-CsReprocessGroupPolicyAssignment | Required GroupId Guid, PolicyType and TenantRegion |

All inherit a required Tenant Guid (aliases TenantId, Domain, TenantDomain) and
optional Environment from IC3DirectoryTorusCmdletBase. The attribute alias Domain
does not make the underlying Guid a DNS-domain input. Region and Environment enum
values and default all-parameter set names still need executable metadata capture.

The base obtains a ConfigAPI token from the host and initializes X-MS-Tenant-Id and
a fresh X-MS-Correlation-Id. RestClient sends a Bearer token directly to the selected
configured ConfigAPI origin. No generated forest bootstrap or compiled policy-RP
MPACmdlet default appears in this sender. Its 403 presentation refers to Contributor
elevation for the GPA/LROS subscription; that message establishes neither ordinary
tenant role sufficiency nor application eligibility. No elevation is attempted.

## OCE route and response findings

CmdletSettings.json supplies distinct GPA and LROS route templates. These are
source evidence, not permission to send requests or evidence of feature availability.

| Operation | Method / configured relative route |
|---|---|
| Batch status by ID | GET Skype.Policy/assignments/operations/status/{operationId}?region={region} |
| Batch statuses in region | GET Skype.Policy/assignments/operations/status?region={region} |
| Batch definition | GET Skype.Policy/assignments/operations/definition/{operationId}?region={region} |
| Batch reprocess | POST Skype.Policy/assignments/operations/reprocess/{operationId}?region={region} |
| Group assignment | GET Skype.Policy/groupPolicyAssignments/oce/{groupId}/policyTypes/{policyType} |
| Group assignments | GET Skype.Policy/groupPolicyAssignments/oce/{groupId}/policyTypes |
| Assignments by policy type | GET Skype.Policy/groupPolicyAssignments/oce?policyType={policyType} |
| All group assignments | GET Skype.Policy/groupPolicyAssignments/oce |
| Group members | GET Teams.InternalSupport/groups/{groupId}/users?all={all} |
| User membership | GET Teams.InternalSupport/users/{userId}/memberOf |
| Refresh group users | POST Skype.Policy/groupPolicyAssignments/{groupId}/refresh/users?tenantRegion={region}&userId={userId}&policyType={policyType} |
| Reprocess group assignment | POST Skype.Policy/groupPolicyAssignments/{groupId}/policyTypes/{policyType}/reprocess?tenantRegion={region} |

LROS region-specific reads require HTTP 200. Reads without a region search the
environment's configured regions: single-operation lookup stops at the first 200;
list collects regional results. Failures remain per-region observations. The command
adds Region to decoded status/definition results. Reprocessing requires HTTP 202 and
parses the response as a quoted Guid; it does not wait for completion. These are
distinct from generated policy-assignment operations and their operation-ID DTO.

GPA uses any successful HTTP status, then Newtonsoft JSON decoding into a single
assignment, assignment list, member list, membership list or string action result.
Refresh has no request body and substitutes an omitted UserId with Guid.Empty.
The source interpolates string policy types directly; native code must encode path
and query components separately while preserving routing and omission semantics.

## Remaining implementation sequence

1. Capture both OCE assembly metadata and compare command signatures across targets;
   preserve conditional provenance separately from the immutable default observation.
2. Inspect generated assignment DTOs and send/callback methods, plus wrapper bindings;
   capture requests and model projections using isolated, in-memory senders.
3. Implement typed native services for ordinary and explicit OCE operations using
   existing audience/destination controls. Preserve per-region failures, rank, direct
   and effective assignment provenance, and per-user batch outcomes.
4. Wire the root client, exact command/set mappings, compiling examples, offline
   contracts and opt-in tenant tests. Keep submission acceptance separate from
   completed processing or effective policy state; do not replay mutations implicitly.

No tenant request or mutation ran during this investigation. OCE contributor
eligibility, mode support, cloud availability, successful reads and consequential
operations remain unqualified.
