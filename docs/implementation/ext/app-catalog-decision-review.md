# AppCatalog four-command closure review

Status: independently reviewed and approved. This decision records implementation
and pinned-source offline contract completion for the four records below, each
with one accepted __AllParameterSets set. It claims no successful tenant execution
and does not close EXT-90.

| Command | Native constituents | Required/optional public inputs |
| --- | --- | --- |
| Get-TeamsApp | Service.NewPager, Service.List; nullable model/page decoder | Optional Id, ExternalId, DisplayName, DistributionMethod strings |
| New-TeamsApp | Service.Create; nullable model decoder | Mandatory DistributionMethod and Path |
| Set-TeamsApp | Service.UpdateByID; binary upload | Mandatory Id and Path |
| Remove-TeamsApp | Service.DeleteByID | Mandatory Id |

Immutable public sets come from `contracts/powershell/8.1.0/runtime.json` and
compile to the directly exported TeamsCmdlets classes. This family has no
TRPS/Modern/custom-app suffix dispatch. The
[primary source packet](evidence/app-catalog-primary.json),
[Get supplement](evidence/app-catalog-get-supplement.json), and
[New binding proof](evidence/app-catalog-new-input-binding.json) retain exact
methods, metadata, actual pinned parser observations and inert output/binding
probes. No source probe invokes a management lifecycle/authentication/HTTP.

## Per-command behavior and execution

Get constructs the four filters in source order, omitting IsNullOrWhiteSpace
values and preserving encoded mixed controls. It follows source continuation
metadata and accepts any 2xx before decoding. Null rows remain nil pointers;
nonnull objects can have nil/empty/whitespace IDs. All four strings preserve
exact numeric lexemes and lowercase booleans. Ordered typed decoding implements
scalar alias last-wins and collection alias append/reset semantics, independently
of the native RawFields map. Empty/null/missing collections fail, and nonblank
continuations retain native origin/operation/cycle/budget guards. Actual source
fixtures plus HTTP replays cover aliases, nullable rows, large IDs, pages,
continuation termination, decode failure and historical read denial.

New's mandatory DistributionMethod has no organization default. Actual inert
binding rejects omitted/null/empty before processing; nonempty values bind, and
processing accepts only OrdinalIgnoreCase organization. The native explicit
organization requirement rejects invalid values locally before token acquisition.
Caller-owned file I/O maps Path to Package bytes. Nil means omitted input;
explicit empty bytes map an existing readable empty file, reaching HTTP as the
full source bodies permit. POST uses application/zip, accepts any 2xx and decodes
one nullable model. Empty/null output keeps a nil model and HTTP receipt, not a
fabricated ID or presumed failure. Tests execute organization casing, missing-ID
objects, scalar-null/empty responses, zero-byte uploads, accepted decode failure,
local errors and a response-only 403 write replay.

Set preserves the pinned PUT /v1.0/appCatalogs/teamsApps/{id}, exact package
bytes/application/zip and any-2xx acceptance without body decoding. Empty loaded
packages reach PUT; nil package remains omitted-input failure. Tests exercise
200/201/202/204/206/299, required safe ID/package errors and a response-only 403
receipt/no-replay path. This is the pinned 8.1.0 operation; current update docs
specify a distinct appDefinitions POST and do not prove legacy PUT server/auth
compatibility. No protocol substitution is hidden in this proposal.

Remove preserves DELETE to the catalog ID and any-2xx acceptance without a model.
Tests exercise 200/202/204/206/299, required safe ID errors and a response-only 403
receipt/no-replay path. Catalog ID remains distinct from the manifest/external ID.
A 2xx receipt alone does not prove completed deletion.

## Binding, common controls and deliberate native boundaries

All four accepted public sets expose logging/error/output preferences but no
WhatIf/Confirm parameters. Their processing bodies contain no ShouldProcess or
ShouldContinue call. The SDK adds no invented ShouldProcess behavior: callers
choose whether to invoke a primitive. Native context, logging and Go return
values/errors replace PowerShell stream/preference controls; no PowerShell
pipeline or preference object is executed by the SDK.

Required source Id/Path strings use mandatory binding. Native mutation IDs keep
the separate blank/control/path-navigation guards and safe segment escaping;
read ID nullability does not authorize mutation. Path's filesystem failures
remain caller-owned I/O, as compiled examples demonstrate. Native nil versus
explicit empty package presence is intentional and source-backed. Model/page
pointer corrections are explicit unreleased public API changes, not dropping
null rows, coercing them into objects or adding compatibility sentinels.

Quote doubling, trusted bounded/cycle-protected traversal, partial read results,
raw-field retention and native transport/retry policy are documented extensions.
Source buffers all pages before output; native partial results always retain the
error and last receipt. Writes are ReplayNever even where upstream transport
helpers retry. Accepted malformed responses retain HTTP receipts. They do not
establish failure/absence, ownership or permission to replay/compensate blindly.

## Authorization disposition and external limits

The [current official API/permission review](app-catalog-auth-and-current-api.md)
is separate from pinned-source request parity. It documents application catalog
read permission support, delegated-only published create/delete/update support,
admin prerequisites and the current update-route discrepancy. Current permission
tables for the appDefinitions POST are not evidence for legacy PUT auth support.
No identity, consent, permission or role is changed and no implicit fallback or
local unsupported-mode stub is introduced. Historical application 403 results
remain denials, not positive operation evidence or a promise that an application
write grant would enable those published mutations.

Successful live catalog reads, delegated publication/deletion with an approved
owned package and current legacy-PUT server compatibility remain unqualified.
These are explicit live/backend limitations; this contract decision is
bounded to the fully traced pinned 8.1.0 request/model/control behavior and the
native boundaries above. Source/actual local captures, mock 2xx receipts and
historical tenant 403 bodies remain separately labeled.

## Exact affected validation

Clean source revision b5b26e2 passed the complete AppCatalog SDK TestUnit/Example
selection and audited custom-app lab consumer compile: exit 0, both packages
passed, 36 passing test events, no failure/skip and no TestAcc. This includes
command-specific binary-write/delete denial receipts and no replay, beyond the
prior 772483b 32-pass implementation checkpoint. The raw result and exact source,
consumer, fixture and primary-evidence hashes accompany this reviewed packet.
The earlier 8678086 failed fixture-label artifact is preserved unchanged. No
production or harness change was used to conceal that failure.
