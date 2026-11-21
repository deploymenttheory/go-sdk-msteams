# POL-28 policy package contract investigation

Issue #101 owns nine commands and twelve parameter sets in the accepted 8.1.0
runtime snapshot. Native service methods and tests now map that surface, with
implementation in-progress and contract metadata-only. No full parity or tenant
qualification is claimed. The default baseline is unchanged; no package request
or mutation has run in the lab.

The pinned package SHA256 is
`78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`.
Generated service/cmdlet/model facts come from
bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll, SHA256
`8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6`.
Custom wrappers come from custom/Merged_custom_PsExt.ps1, SHA256
`03dbc3a85e12c98dd1807d604782c6a90d0656569541cd9540738825a7fcb255`.
Local ILSpy inspection does not construct commands or acquire credentials.
Microsoft binaries and decompiled source are not redistributed here.

| Baseline command | Sets |
|---|---|
| Get-CsPolicyPackage | Get, Get1, GetViaIdentity |
| Get-CsUserPolicyPackage | Get |
| Get-CsUserPolicyPackageRecommendation | Get |
| Grant-CsGroupPolicyPackageAssignment | RequiredPolicyList |
| Grant-CsUserPolicyPackage | GrantExpanded |
| New-CsBatchPolicyPackageAssignmentOperation | NewExpanded |
| New-CsCustomPolicyPackage | RequiredPolicyList |
| Remove-CsCustomPolicyPackage | Remove, RemoveViaIdentity |
| Update-CsCustomPolicyPackage | RequiredPolicyList |

## Generated routes and protocol distinctions

All inspected helpers initially target the standard ConfigAPI origin. The native
service uses the established generated ConfigAPI adapter for forest-header routing.
These routes are distinct from compiled policy-RP and conditional OCE protocols.

| Helper | Method / relative route |
|---|---|
| PolicyPackageGet | GET /Teams.VerticalPackaging/package |
| PolicyPackageGet1 | GET /Teams.VerticalPackaging/package/{escapedPackageName} |
| CustomPolicyPackageNew | POST /Teams.VerticalPackaging/package/custompackage |
| CustomPolicyPackageUpdate | PUT /Teams.VerticalPackaging/package/custompackage |
| CustomPolicyPackageRemove | DELETE /Teams.VerticalPackaging/package/custompackage/{escapedPackageName} |
| GroupPolicyPackageGrantAssignment | POST /Teams.VerticalPackaging/package/group |
| UserPolicyPackageGet | GET /Teams.VerticalPackaging/package/assigned |
| UserPolicyPackageGetRecommendation | GET /Teams.VerticalPackaging/package |
| UserPolicyPackageGrant | POST /Teams.VerticalPackaging/package/batch |
| BatchPolicyPackageNewAssignmentOperation | POST /Teams.VerticalPackaging/package/operation/batchpackage |

User reads and recommendations carry their user identifier in the `user` request
header, not a path/query parameter. The request capture now retains this header;
user-specific recommendations otherwise share the general package-list route.

Generated callback signatures distinguish ordinary 400/404/500 error models from
success data. Group and synchronous user grants additionally expose 520/521 handlers.
Batch package submission uses 202; synchronous user grant uses 201. HTTP acceptance,
response-message text, failed policy creation/assignment, failed users and per-user
error details must remain separate observations, with typed data retained on errors.
No terminal-state or effective propagation guarantee follows from a returned job ID.

## Composition and binding findings

Create/update DTOs have Description, PackageName and PolicyList. Policy entries are
typed PolicyType/PolicyName pairs. Update is a PUT to the collection with PackageName
inside the body, not a guessed PATCH-by-name route. Group assignment uses GroupId,
PackageName and PolicyRankings. Ranking entries carry a non-nullable Int32 Rank;
the generated serializer emits Rank=0 for its untouched default.

The custom create/update wrappers accept string PolicyList entries and split each
once at comma, dot, colon, semicolon, space or tab, then trim both pieces. Missing
delimiters throw locally. Group rankings use the same split, then PowerShell's
`-as [int]` conversion; failed conversion produces null before downstream binding.
Executable wrapper/DTO captures must resolve that coercion rather than guessing a
range validator or silently discarding rank zero. Native typed lists should be the
primary API, with explicit local compatibility helpers for wrapper string inputs.

Package responses use capitalized wire fields Name, Description, PackageType,
Policies and RecommendationType. Their policy dictionaries and future fields need
lossless snapshots. User grant responses expose FailedCreatePolicies,
FailedAssignPolicies, FailedUserIds and FailedUserErrorMessages. The generated
error container has a known LegacyUser array, rather than an arbitrary user-key
dictionary. Expose that typed field while retaining unknown future object members;
do not infer that current generated properties exhaust possible server data.

## Executable request/model observations

[Request capture](../../../tools/contracts/capture-policy-package-requests.ps1)
now records twenty-three actual generated calls with synthetic callbacks and an
in-memory sender, before host/header rewriting. It retains user headers, DTO casing,
collection PUT updates, nullable package/default rank, distinct accepted/created
responses, partial business failures, malformed success and missing operation IDs.
The 521 user-grant response selects the generated on520 handler; native error handling
must retain the original HTTP status rather than deriving it from the callback name.

[Model capture](../../../tools/contracts/capture-policy-package-models.ps1) adds
eighteen pinned projections. Rank is serialized as zero for missing or null input;
the known LegacyUser error container drops unknown generated properties. Native
read models must preserve raw future data in addition to typed known fields.

The shared development sender gained a credential-refusing request-header capture.
Re-running the existing voicemail assignment capture preserved all fourteen accepted
observations exactly. Neither new oracle performs authentication, command processing,
discovery or network I/O. Production SDK code is not implemented by these captures.

## Authentication evidence

The existing capability catalog records a separately dated application-authentication
exception for all PolicyPackage commands in commercial/GCC/GCC High/DoD profiles.
[Microsoft's application-authentication guidance](https://learn.microsoft.com/en-us/microsoftteams/teams-powershell-application-authentication),
rechecked 2026-11-21, retains that exception. This is documentation evidence, not a
tenant probe or proof of other clouds' support. Preserve the distinction between
this documented restriction and binary wire evidence; do not infer additional
delegated-only restrictions from a denial or missing feature.

The native guard uses the catalog per command and resolved documentation profile
before discovery/token acquisition. Unknown documentation profiles stay unqualified.
Both generated headers and complete routes are validated before discovery; typed
business-error decoding excludes discovery replies. Native group rankings use typed
Int32 values rather than silently reproducing arbitrary shell numeric string casts.

The existing dedicated lab identity uses application credentials, so successful
package tenant qualification needs an appropriate delegated context. It is not
resolved by granting that SPN more Graph permissions or by interpreting 403 as
evidence that application mode works. Independent implementation/capture work can
continue without changing roles, consent or interactive sessions.
The opt-in TestAccPolicyPackagesReadOnly run confirmed the configured identity's
local mode rejection and skipped delegated qualification. No package request or
token acquisition ran. A separately supplied, tenant-bound delegated-token file
can enable the bounded read path later; absence remains visible as unavailable.

## Remaining qualification

1. Complete shell/CLR coercion qualification, especially numeric string ranks;
   the primary SDK API uses typed entries and preserves rank zero explicitly.
2. Run delegated list/read projections against the pinned module with tenant-bound
   fixtures. The existing application identity is not an eligible substitute.
3. Exercise custom CRUD and group/user/batch assignment on owned disposable fixtures
   with non-default read-back and cleanup under mutation/consequential gates. Keep
   partial outcomes and HTTP acceptance separate from effective convergence.
