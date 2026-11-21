# CORE-22 deployment and migration SDK evidence

Candidate implementation for issue #77: four commands/five parameter sets have
native symbols. The ledger remains in-progress/metadata-only/not-run and the issue
stays open. Exact authorization, deployment grammar/workflow outcomes, cloud
deployment and consequential lab behavior remain unqualified. No tenant request,
notification email, directory grant or migration mutation was made for this work.

| Command / set | Go mapping |
|---|---|
| New-CsBatchTeamsDeployment / __AllParameterSets | batchdeployment.BuildArchive; Service.Create |
| Get-CsBatchTeamsDeploymentStatus / Get | batchdeployment.Service.GetByID |
| Get-CsBatchTeamsDeploymentStatus / GetViaIdentity | batchdeployment.Identity; Service.GetByIdentity |
| Get-MultiGeoRegion / Identity | region.Service.GetForUser/GetForGroup |
| Get-TenantPrivateChannelMigrationStatus / __AllParameterSets | privatechannelmigration.Service.Get |

Provenance remains the immutable package/assembly hashes, pinned methods and
offline model/dispatch observations in [batch evidence](template-policy-deployment-evidence.md),
[ConfigAPI routing](policy-routing-evidence.md) and
[migration/profile evidence](hierarchy-migration-contract-evidence.md). Production
Go does not load PowerShell or source fixtures. The generated deployment request's
development hostname is pre-callback data; the selected ConfigAPI origin and
existing same-origin discovery/header adapter are used before transmission.

The archive builder writes exactly Teams.csv/Users.csv with fast Deflate, preserves
original bytes, bounds combined input and full ZIP framing, retains reader ownership
and uses no shared temporary files. Create sends plural DeploymentCsvs plus supplied
UsersToNotify once and expects 202. Both identities/future fields and accepted partial
metadata survive decoding failure. Raw status and HTTP errors stay available; named
400/404/500 strings are failures in Go. Completed never implies success. A structured
per-team decoder/waiter, arbitrary resource CRUD or automatic compensation is not
invented. Submission can send notification email; live destructive acceptance needs
explicit disposable fixtures/recipients and cleanup decisions.

Private migration retains all known/null/future properties and original timestamp
text. Details remains JSON text with optional explicit object decoding. Native 2xx
handling follows the pinned helper, without assuming an unproven full state machine.
Exact cloud origins/prefixes/templates audiences are shared by auth/transport;
unsettled clouds do not fall back. Same-host cross-cloud prefixes fail before tokens.

Multi-Geo reads escape caller IDs without Graph lookup and preserve rich nullable
observations, even when the module projects only Region. Application mode is
documented unsupported and locally rejected. The commercial/GCC profile requests
exact delegated Region.Read.All. TokenRequest.Scope is comparable and participates
in acquisition/coalescing/cache/rejection isolation. Existing empty-scope requests
retain resource /.default behavior. Supplied tokens use TokenRequest.TokenKey for
explicit scope inventory, without reusing audience-default entries or decoding JWT
claims. Native MSAL/borrowed Azure requests preserve exact scope; arbitrary native
scope widening is rejected. Tests prove one scope's rejection does not invalidate
another and a transport 401 refresh keeps the exact scope and binding.

Root-client in-memory tests assert method/path/escaping/query/body/header/audience/
scope and selected outputs. Existing Graph/ConfigAPI/default-token regressions remain
covered. Retained report scope is four batch, four region and ten migration cases,
including all five command/set mappings; additional error, context, archive budget,
coalescing, mode and namespace cases run separately. CI retains only the exact three
resource report filenames. Reports keep coverage_claim:false/live:not-run and do not
contain credentials, identifiers, URLs, CSV/ZIP payloads or server text.

[Lab acceptance](deployment-migration-lab-acceptance.md) is prepared and unrun.
Read-only batch/private migration probes may compare selected pinned PowerShell
observations; region needs a separately provisioned delegated token and a qualified
delegated PowerShell session for later comparison. Application credentials are not
substituted. Native read/date/detail preservation and granted roles/scopes remain
distinct from successful live or exported-command qualification.
