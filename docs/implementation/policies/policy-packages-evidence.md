# POL-28 policy package contract investigation

Issue #101 owns nine commands and twelve parameter sets in the accepted 8.1.0
runtime snapshot. This is a source investigation checkpoint, not a native service,
contract pass or tenant qualification. The default baseline and coverage states
are unchanged. No policy-package request or mutation has run in the lab.

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

All inspected helpers initially target the standard ConfigAPI origin. Generated
host/header rewriting still needs explicit inspection/capture before implementation.
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
header, not a path/query parameter. A request capture must retain this header;
the existing shared sender currently captures only method/URL/body. User-specific
recommendations otherwise share the general package-list route.

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

## Authentication evidence

The existing capability catalog records a separately dated application-authentication
exception for all PolicyPackage commands in commercial/GCC/GCC High/DoD profiles.
[Microsoft's application-authentication guidance](https://learn.microsoft.com/en-us/microsoftteams/teams-powershell-application-authentication),
rechecked 2026-11-21, retains that exception. This is documentation evidence, not a
tenant probe or proof of other clouds' support. Preserve the distinction between
this documented restriction and binary wire evidence; do not infer additional
delegated-only restrictions from a denial or missing feature.

The existing dedicated lab identity uses application credentials, so successful
package tenant qualification needs an appropriate delegated context. It is not
resolved by granting that SPN more Graph permissions or by interpreting 403 as
evidence that application mode works. Independent implementation/capture work can
continue without changing roles, consent or interactive sessions.

## Next implementation steps

1. Capture generated request headers/body/status callbacks and model projections;
   verify wrapper pair parsing and rank/null binding with isolated helpers.
2. Establish the documented mode guard and its cloud bounds before token acquisition.
3. Add native custom-package CRUD/list and explicit user/group/batch operations,
   preserving composition, dictionary presence and typed failure results.
4. Wire the root collection, all twelve set mappings, compiling examples, offline
   contracts and opt-in delegated tenant acceptance. Leave unavailable qualification
   visible without publishing unsupported method stubs or dropping commands.
