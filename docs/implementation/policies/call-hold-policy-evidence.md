# Call hold policy: per-operation transport evidence

POL-35 / issue #108, MicrosoftTeams 8.1.0, checked 2026-11-21. The source checkpoint
below establishes contracts for native implementation; it makes no complete
command or tenant qualification claim.

## Export selection

Get/New/Set/Remove are exported Core cmdlets. Grant is an exported proxy function
delegating to private `Microsoft.Teams.ConfigApi.Cmdlets.GrantCsTeamsCallHoldPolicy`,
which inherits TrpsPSCmdlet. `custom/CmdletConfig.json` maps only Grant to
`Grant-CsTeamsPolicy`, PolicyType `TeamsCallHoldPolicy`. The resource therefore
needs compiled CRUD and modern generated assignments within one native service.
Using the compiled generic Grant helper would not mirror this export.

The exact eight baseline sets are Get Filter/Identity, Grant GrantToGroup/
GrantToTenant/Identity, and New/Remove/Set Identity. Runtime metadata, the Core
TeamsCallHoldPolicy and four command classes, private Grant class, command
mapping and unmodified mapped Grant function were checked. Full command binding,
ShouldProcess and tenant execution remain separate from helper observations.

## Provenance and reproduction

Pinned package SHA256:
`78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`.
The fixtures include assembly/source hashes and host version. Scripts check
the manifest and applicable hashes and reject an existing output path.

- [Compiled capture](../../../tools/contracts/capture-call-hold-provider.ps1):
  actual HttpRequestHelper/PolicyRp and model constructors/setters/Newtonsoft.
  Synthetic InvocationInfo, a credential-refusing host and memory sender replace
  the private client/module-version/host and restore them in finally. The factory
  MPACmdlet default is supplied. No command lifecycle or network runs.
- [Grant wrapper capture](../../../tools/contracts/capture-call-hold-grant-wrappers.ps1):
  unmodified mapped Grant AST with synthetic telemetry and in-memory internal
  command stubs. The pinned command configuration is hash-checked and the Grant
  mapping checked explicitly. No exported processor, flights or tenant executes.
- [Grant request capture](../../../tools/contracts/capture-call-hold-grant-requests.ps1):
  actual generated service helpers use a credential-refusing memory sender,
  before production session/header rewriting. No HttpClient, production pipeline,
  credentials or discovery is invoked.

No Microsoft binary or decompiled source is redistributed. The shared compiled
and modern contracts are deliberately distinct. See
[modern routing evidence](../core/policy-routing-evidence.md) and
[audio conferencing modern evidence](audio-conferencing-policy-evidence.md) for
the traced shared generated callback/discovery and flight-selection boundary.

## Observed compiled CRUD

[Provider/model fixture](../../../contracts/endpoints/8.1.0/fixtures/call-hold-provider.json):
15 one-send/zero-token cases and 10 model observations. List envelope and 403
fail as expected; three wrong structured string-field shapes fail decoding.

- Collection `/Skype.Policy/configurations/TeamsCallHoldPolicy`; GET/PATCH/DELETE
  items add `/configuration/{identity}`. The first slash separates scope and the
  remaining identity is escaped as an opaque segment.
- MPACmdlet `true`, the actual CRUD command and a GUID correlation header are
  present; X-MS-Target-Uri is absent. There is no generated forest discovery.
- Create POSTs bound settings and route Identity. Empty settings remain possible.
  The Core New comparison with Global is ShouldProcess presentation; no tenant
  default is silently copied by native creation. Success bodies may be a policy
  or empty; no mandatory generated POST/Get composition is inferred.
- PATCH uses model Identity, independently from the route, and only bound fields.
  Source Base Set requires a bound setting, reads with Set metadata, updates its
  model, then sends the sparse PATCH. A native Update composition must expose
  read preparation/HTTP acceptance and preserve failure before or after dispatch.
- DELETE has an Identity JSON body. Bound Force false/true uses strings
  `False`/`True` in compiled mutation bodies; omission is distinct.
- Empty/JSON-null reads produce null. Lists are bare arrays, not Graph envelopes.

The model has nullable strings Identity, Description, AudioFileId,
StreamingSourceUrl and StreamingSourceAuthType. Source setters accept empty,
null and future values; no URL, GUID or auth-type enum restriction is inferred.
The discarded Description length constants do not enforce a limit: 2,000
characters are accepted by the actual model. Repeated/case-varied scalar
properties use the later setter value. Unknown response properties remain
available to native callers separately from explicit known-field write DTOs.
Arbitrary shell/Newtonsoft coercions and full parameter binding are unqualified.

## Observed modern Grant

[Wrapper fixture](../../../contracts/endpoints/8.1.0/fixtures/call-hold-grant-wrappers.json):
11 cases; missing PolicyName/scope and bound PassThru true/false fail as expected.
[Generated fixture](../../../contracts/endpoints/8.1.0/fixtures/call-hold-grant-requests.json):
seven one-send cases cover user clear, tenant and group 202/204/priority behavior.

The generated path uses ordinary ConfigAPI service discovery and forest routing
through X-MS-Target-Uri, without an unconditional MPACmdlet. The credential
destination remains ConfigAPI and the audience remains Teams administration.
Default modern selection is command/flight-dependent; no flight result is
claimed. Native REST execution must not add a PowerShell/TRPS fallback.

| Scope | Method and route | Declared success |
| --- | --- | --- |
| User | PATCH `/Skype.Policy/users/{user}/policies/TeamsCallHoldPolicy` | 200 |
| Tenant | PATCH `/Skype.Policy/tenants/policies/TeamsCallHoldPolicy` | 200 |
| Group | POST `/Skype.Policy/groupPolicyAssignments/{group}/policyTypes/TeamsCallHoldPolicy` | 202 or 204 |

PolicyName is required to be bound. An explicitly bound null string becomes
empty, the captured clear payload. Missing user identity requires Global.
Group Rank nil/zero is omitted by the wrapper; nonzero values pass through.
The generated native DTO can explicitly send zero priority. Public PassThru is
not accepted by the mapped modern function. Group acceptance must come from the
actual POST response, not discovery metadata, and must remain distinct from
effective completion. All mutations retain no-replay controls.

## Qualification still required

Native per-operation root wiring, CRUD, modern scoped grants, typed settings/raw
models, exact ledger mappings, examples and opt-in lab reads accompany this
checkpoint. Contract cases run both credential modes synthetically, check an
interleaved CRUD/Grant sequence, verify sparse Set preparation/acceptance,
Rank-zero behavior, discovery versus actual POST 202 and single-send mutation
errors. Source model observations and raw snapshot ownership are checked locally.

Actual client-secret and certificate Global reads returned resource 403 on
2026-11-21. Filter/list and optional pinned comparison were not reached. No
tenant state changed. Successful authorization, cloud/flight deployment, CRUD
read-back/cleanup, streaming/audio playback, effective grants and optional pinned
PowerShell/Go comparisons remain open. The ledger stays in-progress/metadata-only;
no privilege grant or successful live qualification is claimed.

The full offline race/shuffle suite, SDK/examples build, vet, ledger validation
and its ten tests passed on 2026-11-21. Capture scripts executed and all four
PowerShell scripts parsed. Review caught an oracle comparison using exact raw
key spelling instead of the Core getter projection; a case-varied/duplicate-key
regression failed before correction and passed after typed canonical comparison.
Focused resource/lab race checks and lab vet passed after that harness correction.
Structural reports and CI evidence are reviewed at their exact revisions;
none of these checks establishes live parity.
