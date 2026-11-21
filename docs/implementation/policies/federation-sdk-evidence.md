# POL-23 federation SDK candidate

Native methods cover issue #96's fourteen baseline commands and twenty-one
parameter sets across four resource packages and local domain values. The issue
remains open: implementation is in-progress, contract metadata-only, and successful
tenant qualification is not claimed. No assignments or tenant configuration
changes were made for this candidate.

| Command group | Native mapping |
|---|---|
| Get/Set-CsExternalAccessPolicy; New/Remove-CsExternalAccessPolicy | ExternalAccessPolicies Get/List/ListFiltered, Create, UpdateByIdentity/PatchByIdentity, DeleteByIdentity |
| Grant-CsExternalAccessPolicy | Explicit AssignUser/AssignTenantDefault/AssignGroup; zero/mixed Assign scope rejects locally for the otherwise unhandled Identity set |
| Get/Set-CsTeamsAcsFederationConfiguration | ACSFederationConfigurations Get/List/ListFiltered, UpdateByIdentity/PatchByIdentity |
| Get/Set-CsTeamsExternalAccessConfiguration | TeamsExternalAccessConfigurations Get/List/ListFiltered, UpdateByIdentity/PatchByIdentity |
| Get/Set-CsTenantFederationConfiguration | TenantFederationConfigurations Get/List/ListFiltered, sparse UpdateByIdentity/UpdateGlobal |
| New-CsEdgeDomainPattern/AllowList/AllowAllKnownDomains | Local NewDomainPattern/NewAllowList/NewAllowAllKnownDomains |

## Provenance and protocol boundary

[POL-01 source research](federation-branding-voicemail-evidence.md) records package
SHA256 `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`,
assembly hashes, exact helpers/models and signature metadata. Production Go never
loads PowerShell, source fixtures or .NET.

The [new provider oracle](../../../tools/contracts/capture-policy-provider.ps1)
executes direct pinned HttpRequestHelper methods with synthetic InvocationInfo and
bound parameters. A fake host refuses token acquisition; the private HTTP client
and cached version are replaced/restored. Its only sender is an in-memory handler.
MPACmdlet is inserted by the adapter from factory source; normal factory/bootstrap,
authentication, ShouldProcess and base command execution are not run. The
[24 observations](../../../contracts/endpoints/8.1.0/fixtures/policy-provider-dispatch.json)
settle direct routes/headers, full first-slash escaping, sparse POST/PATCH,
assignment casing, optional any-2xx dispatch, errors and known null-list failures.
Seven additional list observations use the local host's PSListModifier and the
pinned cmdlet's null/IsEmpty branch, confirming ordinal first removal, appended
duplicates and empty-edit clearing. Host-specific binding remains separately
qualified. Captures ran locally; Go CI compares these goldens without rerunning
the new oracles.

The capture corrected an earlier static note: compiled DELETE includes bound
Identity in its JSON body; bound Force adds `"False"`/`"True"`. The earlier `{}`
description omitted AddAdditionalBodyContent's non-common Identity parameter.
Native deletion and the corrected research follow the executed observation.

The [tenant oracle](../../../tools/contracts/capture-tenant-federation.ps1) captures
three generated requests and sixteen pure normalization cases. Thirteen accepted
property bags cover scalars, allow-all/list, aliases, mixed objects and empty/null
replacements; conflicting choices, a null choice and Replace-operator input reject.
Its [fixture](../../../contracts/endpoints/8.1.0/fixtures/tenant-federation-dispatch.json)
is separate from provider PATCH. The existing four-case domain normalization capture
remains consumed by zero-HTTP/token tests.

Compiled resources use a direct Teams scope with MPACmdlet and fresh correlation
metadata, without generated forest discovery. The selected origin is trusted caller
configuration and still passes shared destination guards. Uri.EscapeDataString's
unreserved encoding is preserved, including colon/plus/at and the first scope/name
slash. Tenant settings instead use generated discovery, full-segment identities
and sparse PUT. No Graph lookup, cross-cloud fallback or new auth audience is added.

## Native behavior and verification

The public policy value package represents omitted/null/value scalars and owned
full arrays/local list edits. Native Patch is one request; Update performs the
pinned read/modify/write composition and retains preparation/HTTP acceptance on
failure. Create is one native POST; explicit CreateWithGlobalComparison retains a
failed comparison without cloning Global defaults or preventing later creation.
No read-back, ETag/atomicity, compensation, polling or mutation replay is implicit.

Assignment PolicyName null/empty, tenant Force Boolean, CRUD Force string and
Int32 group rank remain distinct. An accepted group response is an optional policy
snapshot/raw future fields, not a fabricated async job. Empty/null successful bodies
remain nil. Read snapshots retain wire presence/future data rather than silently
materializing CLR constructor defaults. Known collection nulls rejected by pinned
setters fail decoding with raw response retained. Native reads preserve an ACS null
without asserting a service default. Native list sorting uses exact Global; modern
tenant sorting follows the separate helper. Locale-specific matching and uncommon
binding/coercion behaviors remain qualification gaps.

Colocated root-client tests consume pinned wire observations in both synthetic mode
bindings, assert request shape/audience/headers, and exercise local filtering, all
four request field inventories, copied inputs, null/false/empty values, non-replay,
denials, cancellation and partial failure. A transport spy verifies fresh UUID
correlations and absence of forest bootstrap across all compiled families. The
shared modern list serializer now serves application-access, application-meeting
and tenant federation wrappers, preserving existing public types/error identities
and prior fixture tests. Reports contain structural evidence only and retain
coverage_claim:false/live:not-run; they do not promote complete baseline parity.

Local full race/shuffle tests, build, vet and full lint passed. Coverage/model/
resource validators and all sixteen Python validator tests passed. New PowerShell
scripts parse; the disabled read oracle rejects before module/credential access.
The provider/list capture host was PowerShell 7.6.6, SMA assembly 7.6.0.500.
The PR records the candidate revision and platform CI evidence separately.

## Actual lab status

On 2026-11-21, TestAccFederationReadOnly was attempted with the existing dedicated
SPN using client-secret and certificate authentication. All four selected Global
resource reads returned HTTP 403 in both runs. Only fixed-stage status metadata was
logged; no payload, token, identifier or principal data was retained. The later
filtered-list and PowerShell comparisons did not run. The four Get ledger entries
record live unavailable, distinct from unsupported mode/feature or successful
acceptance. A 403 does not identify its sole cause. The separately requested role
approval remains pending; no role, consent or browser changes occurred.

[Prepared acceptance](federation-lab-acceptance.md) includes optional pinned
PowerShell read comparison and journaled CRUD of one random unassigned external
policy. The mutation test has not run. Tenant Global changes and assignment effects
are consequential and have no automatic live execution. Successful RBAC/consent,
effective policy, propagation/clear semantics, cloud deployment, legacy flight and
exported wrapper behavior remain open in #96.
