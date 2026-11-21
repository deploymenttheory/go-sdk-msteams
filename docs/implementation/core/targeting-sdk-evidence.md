# CORE-25 targeting SDK evidence

Implementation candidate for issue #79: six commands/seven observed parameter
sets have native resource/composition mappings. Tenant authorization, modern/TRPS
conditional parity, consequential effects and missing resource contracts remain
open. No tenant request or credential/grant change was made for this work. Ledger
status remains in-progress/metadata-only/not-run; the issue stays open.

| Command / set | Go symbols |
|---|---|
| Get-CsTeamsTargetingPolicy / Identity | targetingpolicy.Service.GetByIdentity, List |
| Get-CsTeamsTargetingPolicy / Filter | targetingpolicy.Service.ListFiltered |
| Set-CsTeamsTargetingPolicy / Identity | targetingpolicy.Service.UpdateByIdentity, UpdateGlobal |
| Remove-CsTeamsTargetingPolicy / Identity | targetingpolicy.Service.DeleteByIdentity |
| Set-TeamTargetingHierarchy / __AllParameterSets | targetinghierarchy.Service.UploadCSV; native v1/v2 ImportCSV, v2 GetHierarchy/CreateHierarchy |
| Get-TeamTargetingHierarchyStatus / __AllParameterSets | targetinghierarchy.Service.GetStatus; native v1/v2 GetOperation |
| Remove-TeamTargetingHierarchy / __AllParameterSets | targetinghierarchy.Service.Delete; native v1/v2 Delete |

Request/profile provenance is the pinned binary and source-derived fixture work
in [policy evidence](template-policy-deployment-evidence.md),
[routing/eligibility](policy-routing-evidence.md),
[hierarchy/profile evidence](hierarchy-migration-contract-evidence.md) and
`special-service-profiles.json`. Immutable assembly/package hashes remain there.
Production Go does not load PowerShell, decompiled code or fixture files.

Policy bodies preserve bound-only settings and explicit empty strings, open modes,
and identity in the sparse PUT. Raw/null/future fields survive reads. Filtering
stays local, Tag-aware and Global-first. The shared matcher extraction retains
existing application-access-policy behavior. Modern routing remains conditional
relative to the exported module's flight-selected TRPS branch; no module-specific
telemetry is introduced as an SDK feature-availability oracle.

Hierarchy native services preserve singleton semantics, bound canonical tenant,
separate API versions, exact audiences and fixed namespaces. Multipart source
bytes/name/type match pinned execution. V2 prepares by Get, Create after 404 only,
then Import; partial outcomes and preparation responses survive later failure.
Mutations are sent once without rollback/polling. HTTP 2xx behavior follows the
pinned helpers; imports/status still require decodable data. Supporting hierarchy
Get/Create responses remain opaque. Missing/non-string/blank v2 id is a native
decode error with response metadata, rather than a fabricated HTTP 400. V1's
optional requestId and native future response fields remain explicit.

Closed profile/auth tests compare the versioned origin/audience/cloud tuples,
deny unsupported clouds/foreign namespaces before credentials and verify four
separate native resource tokens/caches with secret, PFX and PEM flows. Delegated
tests preserve one selected account across resource changes. The new profiles
use /.default; Region.Read.All or other explicit scope/cache behavior is not
silently added. Existing endpoint overrides retain their Graph/ConfigAPI scope.

Six existing shipped-CLR status observations and eight new
`hierarchy-status-projections.json` observations verify presentation independently
of Go implementation. The new hash-guarded capture tool invokes only actual CLR
constructors, without module import/cmdlet/HTTP. Observations include ordered
error arguments, grouped repeated codes, empty error details, numeric/comma enum
parsing, case and whitespace. Native raw statuses/timestamps/errors remain
independent from those presentation quirks. Go tests normalize the fixture's
source line endings to the target platform. The eight new constructor observations
were captured locally; CI Go comparisons do not claim fresh CLR execution there.

The real root client is exercised with synthetic providers/in-memory transports.
The runner now supports optional synthetic tenant UUIDs and structural multipart
comparison without retaining payloads. Tests reject contradictory fixture bodies,
malformed framing and directory-bearing filenames, and report structural paths
only. Resource tests cover all command/set mappings, exact routes/query/headers/
audience/body, source closure, current/unknown states, malformed nested data,
unsupported profiles, canonical tenant, failed preparation, missing import IDs,
403 and unreplayed 429/503. Examples compile without I/O.

CI retains only exact `targeting-policy-service.json` and
`targeting-hierarchy-service.json` reports. The five retained policy and 37 retained
hierarchy cases are scoped evidence with coverage_claim:false/live:not-run;
additional validation/failure tests do not silently become full parity claims.
The PR records actual revision/platform results. Prepared
[lab acceptance](targeting-lab-acceptance.md) is read-only and unrun. Policy
Create/upsert, hierarchy List/full response model, exact authorization, live
propagation and tenant-wide backup/restore remain explicit closure work.

The optional oracle's hierarchy conversion was also checked against a synthetic
actual pinned CLR Result, and its policy conversion against a synthetic property
bag preserving empty/null/absent values. AST parsing and the disabled guard pass
without module import, management command execution or tenant access.
