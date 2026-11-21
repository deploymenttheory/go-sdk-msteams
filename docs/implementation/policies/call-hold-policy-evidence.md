# Call hold policy: per-operation transport evidence

POL-35 / issue #108, MicrosoftTeams 8.1.0, checked 2026-11-21. This checkpoint
establishes source contracts before native implementation; it makes no complete
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

Native root wiring, tests, examples, exact ledger mappings and opt-in lab reads
follow this checkpoint. Successful authorization, cloud/flight deployment,
CRUD read-back/cleanup, streaming/audio playback, effective grants and optional
pinned PowerShell/Go tenant comparisons remain open. No tenant mutation,
privilege grant or successful live read is claimed here.
