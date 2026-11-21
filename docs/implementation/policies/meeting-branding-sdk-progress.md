# POL-24 meeting branding SDK candidate

Native methods cover issue #97's eight commands and twelve observed parameter
sets. The issue remains open: implementation is in-progress, contract metadata-only
and successful tenant qualification is outstanding. No policy/assignment or role
changes were made for this candidate.

## Native surface

TeamsAdminAPI.MeetingBrandingPolicies exposes GetByIdentity/List/ListFiltered,
Create/DeleteByIdentity, one-request PatchByIdentity, explicit read/modify/PATCH
UpdateByIdentity, optional CreateWithGlobalComparison, and per-scope assignment
methods. Assign requires exactly one scope; its zero/mixed input rejects locally
for the otherwise unhandled Identity parameter set. Local NewBackgroundImage,
NewBrandingTheme and NewAssuranceSlate construct values without persistence, token
acquisition or URI dereference. NewReaction is an additional typed input helper.

Writes preserve omitted/null/false/empty fields, explicit Force casing and all
8.1.0 theme/chat/reaction fields. Selected nested assets serialize known public
properties and inherited Identity, while read snapshots retain raw future data.
Typed reads preserve missing wire fields without manufacturing CLR defaults;
write projections reflect the pinned public getters. Nulls on known non-null
setters and invalid HTTPS asset fields fail decoding. Parent null lists remain
received nulls without claiming a service default. Context, audience/destination
checks, UUID correlations and normalized errors use the shared compiled transport;
no generated forest discovery, Graph lookup or new audience is added.

## Pinned executable evidence

[Source research](federation-branding-voicemail-evidence.md) records 8.1.0 package
and assembly hashes, methods and parameter metadata. Production Go never loads
PowerShell, .NET or fixture files. The initial two nested serialization cases
remain consumed by snapshot tests.

The [pure asset capture](../../../tools/contracts/capture-branding-assets.ps1)
checks Core/PolicyRp hashes and executes pinned model constructors/setters with
Newtonsoft serialization. Its 31 observations include 16 accepted and 15 rejected
URI cases. Go builders and snapshots consume those goldens. Uppercase HTTPS,
final newline, space and strings beyond 1024 characters follow the simple model
regex. Theme URI-like/color/alignment strings remain open. No image size/order
limit, upload API or URI reachability is invented. Required CLI strings additionally
follow source ValidateNotNullOrEmpty; parameter binding/coercion is not executed.

The [provider capture](../../../tools/contracts/capture-branding-provider.ps1)
checks Core/PolicyRp/host hashes and executes direct HttpRequestHelper methods with
synthetic InvocationInfo. Fake host refuses tokens; version/host/private HTTP client
are replaced/restored and the only sender is memory-only. MPACmdlet default comes
from factory source. Its 21 requests cover CRUD/list, scope payloads, nested selected
fields, bound false Force, empty/denied/wrong-envelope results, raw group acceptance
and five typed asset decoder failures. Captures do not run cmdlet constructors,
ProcessRecord, auth, discovery or network. Go CI consumes goldens, not PowerShell.

Thirty local list observations call the actual pinned Core.PSListModifierHelper
with host SMA containers and source equality delegates. They corrected an earlier
research note: bound null/empty clears the list. Policy-list removal invokes a
throwing comparator whenever an existing element is compared, even for a missing
match; no comparison occurs on an empty target. Reaction removal compares five
fields and removes the first match, preserving additions/duplicates. Nil Go edit
pointers omit settings; ReplaceAssets[T]() represents explicit clearing. Null list
elements remain serializable. Host-specific conversion/binding remains unqualified.

## Verification and qualification

Root-client request tests run both synthetic token modes, proving selected routing
rather than application/delegated tenant eligibility. Local cases prove no HTTP or
credential requests; constructor capture goldens, list rules, nested ownership,
known-field projection, sparse write/composed update, partial failure, error status/
redaction, non-replay and invalid scope/input are exercised. Header spy includes
branding. Request-field inventories compare New/Set and theme options with pinned
signatures. Examples compile; sanitized CI evidence maps exact command/set pairs.

TestLabMeetingBrandingReadOnly ran with the existing dedicated SPN using client
secret and certificate. Both Global requests returned resource HTTP 403. List and
optional tenant-bound pinned PowerShell scalar comparison were not reached.
No successful live behavior, mode restriction or feature/RBAC cause is inferred.
The [lab plan](meeting-branding-lab-acceptance.md) separates reads from owned policy
CRUD and consequential assignment/Global scenarios. Those writes and Premium
rendering/effective behavior remain unrun. HTTP acceptance is separate from
propagation; no write replay, automatic polling, rollback, ETag or job handle is
invented. Evidence/report success is not full command qualification.
