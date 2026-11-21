# POL-29 privacy configuration evidence

Issue #102 owns Get-CsPrivacyConfiguration (Filter and Identity sets) and
Set-CsPrivacyConfiguration (Identity set). The native service is implemented with
all three exact parameter-set mappings. The coverage ledger remains in-progress
and metadata-only: these provider observations and synthetic tests do not establish
full command parity or successful tenant qualification. The immutable default
baseline is unchanged.

The accepted 8.1.0 package SHA256 is
`78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`.
The relevant packaged Core/PolicyRp/host SHA256 values are pinned in
[the capture script](../../../tools/contracts/capture-privacy-provider.ps1).
Local ILSpy inspection and the accepted runtime snapshot agree on all four nullable
Boolean settings: EnablePrivacyMode, AutoInitiateContacts,
PublishLocationDataDefault and DisplayPublishedPhotoDefault. The model has a
separate Identity property and simple nullable setters, with no extra validation
or manufactured defaults.

GetPrivacyConfigurationCmdlet derives from BaseTpmGetCmdlet of PrivacyConfiguration.
SetPrivacyConfigurationCmdlet derives from the corresponding BaseTpmSetCmdlet.
It considers a setting specified by BoundParameters membership, preserving false
and null, then updates only those bound properties. Force alone is not a setting.
This establishes the compiled direct protocol, not generated forest discovery or
Graph. The resource name is PrivacyConfiguration.

The native service exposes explicit Get/List/local Filter and sparse PATCH,
plus Set's explicit read/modify/PATCH composition. Scope/identity escaping,
Global-first ordering, Identity/Tag wildcard selection, body identity override,
Force binding and compiled headers belong to the existing policyrp adapter.
It retains preparation metadata if a composed write fails, without atomicity,
ETags, automatic replay or effective-state guarantees. No Create/Delete/Grant
operation is established by these two commands.

[The development oracle](../../../tools/contracts/oracle/PrivacyProviderCapture.cs)
uses actual pinned HttpRequestHelper methods with synthetic InvocationInfo,
credential-refusing host and a memory-only sender. It replaces/restores the private
client, version and host. MPACmdlet's default is supplied from factory source.
No management command constructor/ProcessRecord, credentials, generated discovery
or network runs. The script captures reads/list, nullable/false sparse settings,
bound-false Force, empty success, an invalid collection envelope and a denial.
This is a provider/helper observation, not full shell binding or live qualification.

The eight accepted cases each observed one send, zero credential requests, a valid
correlation ID, MPACmdlet true and no target-URI header. The source helper marks an
empty successful list's internal success flag false, but returns an empty list;
BaseTpmGetCmdlet consumes PolicyResult without checking that flag. The native
service therefore returns the empty collection without an HTTP error. Empty/null
reads stay empty/null; an object-shaped list envelope fails decoding.

The shared compiled adapter admits the named PrivacyConfiguration resource and
preserves a literal whitespace filter. Source BaseTpmGetCmdlet uses IsNullOrEmpty;
the adapter incorrectly treated whitespace as omission. A failing source-backed
regression demonstrated this, and the one-line fix leaves the other families'
empty-pattern behavior intact. The new service rejects an explicitly empty bound
filter locally, while List represents omission. Root wiring, typed models,
examples, sparse/composed mutation
tests, all-field presence tests, input/context checks, denial/status redaction and
single-send mutation failures are resource-owned. Reports map Get/Identity,
Get/Filter and Set/Identity explicitly, with coverage_claim false and live not-run.
False/null values are bound, and omitted read settings are not accidentally sent
back by the composed update. Body identity comes from the read model while route
identity remains caller-owned.

The read-only lab harness optionally compares all four nullable settings against
the hash-checked pinned PowerShell module. It never writes privacy settings or
changes role/consent. Tenant writes require separately owned fixtures, non-default
read-back and an authorized restoration plan. Full shell binding/coercion,
warning/verbose presentation, effective privacy behavior, permission/cloud support
and tenant write qualification remain open under #102. Go responses intentionally
require typed Boolean data; arbitrary Newtonsoft/shell coercions are not claimed.

Actual read-only Global attempts on 2026-11-21 returned resource HTTP 403 with the
existing dedicated SPN in both client-secret and certificate modes. List/filter
and the optional pinned module comparison were not reached. The ledger records
Get's live status as unavailable; Set remains not-run. No payload, credentials,
privacy mutation, permission or role change is retained or performed by these tests.

Local verification on 2026-11-21 passed the full offline race/shuffle suite,
SDK/examples build, vet, scoped lint with zero issues, coverage-ledger validation
and its ten validator tests. Both added PowerShell scripts parse successfully.
The whitespace/empty-filter regression first failed against the original behavior
and passed after the source-backed correction. Exact revision-bound structural
reports are retained by local verification/CI, without promoting coverage states.
