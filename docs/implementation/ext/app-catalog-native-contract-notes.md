# AppCatalog native contract corrections

This checkpoint repairs source-confirmed behavior in the four accepted public
Get/New/Set/Remove-TeamsApp commands. Each has one __AllParameterSets set. These
compiled TeamsCmdlets use Graph directly; no Cloud TRPS/Modern or custom-app
private-header handoff is inferred for this family. No app qualification or live
state is promoted by this implementation checkpoint.

The [portable primary](evidence/app-catalog-primary.json) retains all four exact
processing/status/route bodies, DTO/helper/error metadata and IL, direct root
export dispatch, actual parser captures, inert public WriteObject observations,
reproducer sources and their setup. Source assembly SHA256 is
`3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24`.
The [Get supplement](evidence/app-catalog-get-supplement.json) pins numeric text,
continuation behavior and empty-body flow. The
[New input proof](evidence/app-catalog-new-input-binding.json) closes mandatory
DistributionMethod/default binding with exact constructor/accessor/metadata and
inert binding captures. No management lifecycle, authentication or HTTP ran in
these source probes.

## Response shape and source status

All four active HTTP paths use IsSuccessStatusCode rather than the SDK's prior
narrow status subsets. Reads still require a valid nonnull value array; empty,
whitespace, literal null and missing/null value cause source errors. Create can
instead emit a null scalar for an accepted empty/null body. Native responses and
decode errors retain receipts without inventing identities or persistence.

Actual typed List<TeamsTenantApp> output through the exact WriteObject(object,true)
overload preserves null entries. This requires the unreleased public API change
from value to pointer page entries and ID string to *string. All response models
accept missing/null/empty/whitespace IDs separately from safe required mutation
IDs. The own SDK consumers, examples and the external custom-app lab consumer
were audited; the lab only checks denied-list length/status, not row IDs.

Actual JsonConvert string reads preserve large integer/fraction/exponent lexemes
without rounding, and produce lowercase boolean text. All four known fields and
the continuation use the corresponding nullable projection. Source RawFields is
not used to choose aliases: ordered typed decoding determines scalar last-wins.
A separate typed collection accumulator reproduces the actual default list reuse:
value/VALUE arrays append in wire order, null resets, later arrays restart. The
seven-case actual duplicate observation is retained unchanged.

## Inputs, transport and native extensions

IsNullOrWhiteSpace controls filter omission. Pure framework URI construction shows
mixed tabs/newlines are percent-encoded; the earlier blanket native control
rejection was stronger than this source behavior and is removed for query data.
Mutation ID/path and continuation trust guards remain unchanged. Quote doubling,
bounded/cycle-protected traversal and retained partial read results remain explicit
native extensions; source emits only after every page succeeds.

New DistributionMethod is mandatory and its constructor defaults to null, not
organization. Actual inert binding rejects omission/null/empty before processing;
nonempty values bind, and processing accepts only OrdinalIgnoreCase organization.
Native required explicit organization remains unchanged, with early local rejection
instead of acquiring credentials for known invalid input.

Both New and Set call File.ReadAllBytes then ByteArrayContent then HTTP without
any nonempty-byte check. Native nil bytes retain omitted-input rejection; an
explicit loaded empty slice sends zero bytes with a positive stream budget of one.
No shared upload behavior is changed. The SDK accepts source 2xx receipts without
claiming operation completion; writes are ReplayNever and require reconciliation
rather than blind replay after accepted/uncertain or undecodable outcomes.

The existing serialized custom-app lab selection remains unchanged; no valid app
package is published, app installed, role granted or permission purchased by this
work. Historical 403 response bodies remain denial evidence only. Source reply
vectors and mock HTTP receipts must not be called successful tenant operations.

## Frozen affected validation

The complete AppCatalog SDK TestUnit/Example suite and audited external
custom-app lab consumer compiled/passed at clean `772483b` (session 90098, exit 0):
32 passing test events, zero failure/skip, and both package terminals passed.
The lab had no matching tests and no TestAcc ran. The
[execution record](evidence/app-catalog-772483b-offline.json) retains exact code,
consumer, source/fixture hashes and raw result digest. The preceding `8678086`
run failed only two invalid uppercase fixture labels; its raw artifact is preserved.
The passing revision changes those labels only, with no production/harness change.
All four AppCatalog qualification/live states remain unchanged pending individual
full-set/common-control/auth/error/native-boundary review.

## Authorization evidence and current API discrepancy

The [current official permission/protocol review](app-catalog-auth-and-current-api.md)
distinguishes documented application reads from delegated-only published writes.
The source Set PUT remains exact 8.1.0 behavior; current update documentation's
appDefinitions POST is a separate route and cannot prove the legacy PUT's current
auth requirements or server compatibility. Historical app-only 403 results are
not operation success or a promise that granting write permissions would fix them.
No role/consent or tenant operation is introduced by this documentation review.
