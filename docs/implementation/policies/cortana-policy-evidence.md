# Teams Cortana policy evidence (POL-40 / #113)

The MicrosoftTeams 8.1.0 baseline exports five Functions backed by private
Get/New/Set/Remove/GrantCsTeamsCortanaPolicy TrpsPSCmdlet classes. All five pinned
CmdletConfig mappings select modern configuration/grant functions with
ConfigType/PolicyType TeamsCortanaPolicy. The Core policy class and compiled
provider are not the selected exported path.

Get has Filter/Identity sets; Grant has GrantToGroup, GrantToTenant and Identity;
New/Set/Remove have Identity, for eight exact command/set pairs. The private
New/Set declarations expose Description and CortanaVoiceInvocationMode strings,
and three non-nullable Booleans: AllowCortanaVoiceInvocation,
AllowCortanaAmbientListening and AllowCortanaInContextSuggestions. Feature
availability and effective behavior are not inferred from retained exports.

## Reproducible capture seams

`capture-cortana-wrappers.ps1` verifies the baseline, wrapper hash and all five
command mappings. It executes the unchanged mapped AST functions under synthetic
telemetry and in-memory internal-command seams. Twenty-two observations cover
list/filter ordering, whitespace selection, create/read-back, sparse/default PUT,
delete, grants, required scope/policy name and Rank. Five expected failures cover
create failure, missing scope/name and both PassThru forms. No exported command
processing, runtime attributes, flights, credentials or network runs.

`capture-cortana-requests.ps1` verifies the pinned private assembly, constructs
actual XdsConfiguration/grant DTOs and invokes actual generated service helpers
through a memory sender. Seventeen observations cover read/list, escaped identity,
POST 201, sparse PUT 204 including identity-only and empty/false fields, bodyless
DELETE 204, failure decoding, pre-cancellation and user/tenant/group grants with
202/204 acceptance. Every non-canceled case dispatches once; pre-cancel dispatches
zero times. Host, headers, session discovery and production pipeline are separate.

`capture-cortana-binding.ps1` verifies the export/private hashes, extracts original
New/Set parameter declarations and attributes into a renamed function with a
synthetic bound-argument body, then runs actual pure Xds dictionary serialization.
Twenty-two observations yield twelve successes and ten expected binding failures.
It does not run export Begin/Process/End, runtime custom-attribute processing,
TRPS/flight selection, credentials or networking.

The promoted fixtures are cortana-wrappers.json, cortana-requests.json and
cortana-binding.json under `contracts/endpoints/8.1.0/fixtures/`. Capture tools and
PowerShell are development-only; Microsoft binaries/decompiled source are private.

## Selected behavior

- Modern CRUD uses session forest discovery and generated metadata. Create has a
  one-POST primitive and explicit POST/Get composition with partial read-back
  results. PUT is sparse and allows identity-only requests. Delete is bodyless.
  Whitespace Get/filter selection returns the collection; Global-first ordering
  is case-insensitive and stable. The native API should expose explicit list and
  identity operations rather than silently reinterpret an invalid identity.
- The public declaration binder preserves omitted settings, empty strings and
  false/true. It converts null strings to empty and numeric Boolean zero to false,
  and rejects null Booleans and Boolean strings. Long Description and future mode
  text pass this isolated declaration seam. No private enum, Description length
  or cross-field validation is inferred; runtime/server behavior is unqualified.
- User/tenant grants require 200; group responses distinguish 202 acceptance from
  204. The mapped string binder converts null PolicyName to empty. The wrapper
  omits absent/zero Rank, while the native generated DTO can send explicit zero.
  Tenant helper Force is Boolean; the public family does not expose Force.
  Acceptance needs an actual POST response, not discovery metadata.
- Read values use generic Xds snapshots. Native typed projections should retain
  unknown/unprojectable fields in raw JSON rather than manufacture false/empty
  values. No implicit polling, replay, restoration, elevation, ETag, atomicity or
  effective-clear guarantee follows the captures.

## Remaining implementation and qualification

Implement native modern CRUD/list, explicit create/read-back and grants for all
five settings, root wiring, exact mappings, source comparisons, input/context/
failure tests, documentation/examples and gated read-only lab support. Successful
authorization, server/cloud/flight availability and all effective behavior remain
open. These source captures are not tenant validation; issue #113 and full command
qualification remain open.
