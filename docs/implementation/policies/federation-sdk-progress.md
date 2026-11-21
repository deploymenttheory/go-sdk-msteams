# POL-23 federation implementation progress

Issue #96 has fourteen assigned commands. This checkpoint starts only the three
local domain constructors in `tenant_federation_configuration`; the resource
services, assignments, root wiring and full acceptance remain to be implemented.
No unsupported-operation stubs or native service routes are introduced here.

| Command | Current symbol |
|---|---|
| New-CsEdgeDomainPattern | NewDomainPattern |
| New-CsEdgeAllowList | NewAllowList |
| New-CsEdgeAllowAllKnownDomains | NewAllowAllKnownDomains |

The native allow-list choice marshals the flattened request array, preserving
order, duplicates and explicit empty lists. Allow-all emits the pinned sentinel
array. It does not infer a response default from null or reproduce PowerShell
formatting types. DomainPattern is a local value; no constructor proves actual
DNS ownership, syntax or reachability. Mandatory nonempty domain binding is
represented by NewDomainPattern; list element contents remain unnormalized.

[Static source contracts](federation-branding-voicemail-evidence.md) identify the
exact local helpers and tenant Set normalization. The new
[offline capture](../../../tools/contracts/capture-federation-domains.ps1)
checks 8.1.0/private-assembly SHA256 and executes ConvertToDomainPattern plus
SetCsTenantFederationConfiguration.ValidateAndUpdatePropertyBagParams with
synthetic PSObject choices, then XdsConfiguration serialization. It does not
execute cmdlet processing, connect, acquire a token or attach any HTTP helper.
Its C# adapter suppresses only CS1701 for the pinned SMA 3 reference under this
SMA 7 offline host; the module assembly is unchanged. The
[four retained observations](../../../contracts/endpoints/8.1.0/fixtures/federation-domain-models.json)
are consumed by native tests that assert zero HTTP/token acquisitions, input
ownership and invalid zero-choice behavior. This is local serialization evidence,
not full wrapper/tenant qualification. No tenant call or mutation ran for this
checkpoint. Ledger rows stay in-progress/metadata-only/not-run.

Remaining implementation follows each selected contract independently:

- ExternalAccessPolicy: POST/GET/List/PATCH/DELETE plus explicit assignment actions.
- TeamsAcsFederationConfiguration and TeamsExternalAccessConfiguration: GET/List/PATCH.
- TenantFederationSettings: generated GET/List/PUT and domain-choice/list normalization.

Compiled provider requests require MPACmdlet metadata, preserve first-slash
identity escaping and do not use generated forest discovery. Their Set commands
read current values before applying list changes and emit sparse PATCH. Tenant
federation settings instead use shared generated discovery and full-segment
escaping with sparse PUT. Do not substitute one protocol for the other or infer
an atomic Add/Remove, typed async handle or automatic mutation replay.
