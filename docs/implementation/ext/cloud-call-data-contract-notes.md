# Cloud call data connection contracts

Owner: #85 (EXT-22). Baseline: MicrosoftTeams 8.1.0.

CmdletConfig marks Get-CsCloudCallDataConnection and
New-CsCloudCallDataConnection as Remoting cmdlets dispatched to
`Get-CsCloudCallDataConnectionModern` and `New-CsCloudCallDataConnectionModern`.
Neither public nor Modern cmdlet declares a parameter beyond the hidden
`MsftInternalProcessingMode` and runtime switches.

The generated ConfigApiBasedCmdlets methods give one route,
`/Teams.InternalSupport/CloudCallDataConnection`: GET for Get and POST for New.
The POST builder sets no request content. Both return arrays of
`ICloudCallDataConnection`; each model serializer declares only `Token`.

Lab observations (client-secret, app-only Teams Administrator):

- GET returns HTTP 403 with a top-level `{"code":"Forbidden","message":"Access
  Denied.",…}` body; pinned PowerShell fails with
  `Forbidden,…GetCsCloudCallDataConnectionModern_Get` and writes a correlation ID
  line straight to the console.
- New was sent live on 2026-11-23 with explicit approval, accepting that a
  created connection could not be removed (no remove cmdlet). The POST
  returns HTTP 403 with the same top-level `{"code":"Forbidden",…}` body as
  GET, so no connection was created. Pinned PowerShell
  New-CsCloudCallDataConnection also fails with a `Forbidden,…` error ID. A
  second POST and a GET afterwards return the same 403.

Acceptance:

- `TestAccCloudCallDataConnectionAuthorizationDenied` (read-only) checks the
  GET 403 classification and pinned PowerShell outcome parity.
- `TestAccCloudCallDataConnectionCreateOutcome` (serial step `create-outcome`)
  sends the create and checks the 403 classification and PowerShell New parity.
  It retains the historical explicit create authorization; source inspection does
  not establish server idempotence or ownership of returned models. No remove
  cmdlet or automatic compensation is available. It distinguishes the recorded 403 denial from any future token-bearing 200
  response; the latter has not been qualified.

Both sanitized 403 responses feed unit contract tests. Token-bearing success
bodies are covered by authored Modern array cases only, corrected from earlier
unsupported singleton-object/201 vectors. Successful Modern reads/actions and
server already-exists behavior remain unqualified under the current identity. A
403 does not establish which additional permissions or authentication mode would
authorize them.

## 2026-11-27 finite Modern contract audit

The [full source findings and selected IL](evidence/cloud-call-data-modern-contract.json)
pin `Microsoft.Teams.ConfigAPI.Cmdlets.private.dll` SHA-256
`8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6`.
Generated `CloudCallDataConnectionModernGet_Call` and `...New_Call` state-machine
MoveNext hashes are respectively
`2b74bdba786413d9ff6602d3a7bcdac76be0f636991a9f67d25d5f0286f459df`
and `a27c0d7f6c302eeab109319f898ae14d2ace0bfbaa3d4416f9a99db4dd97b105`.
Both compare status exactly 200 at IL offsets 649/654; 201 goes to default error dispatch.
Their onOk delegates return `Task<ICloudCallDataConnection[]>`, parse JsonArray,
and enumerate models. The earlier single-object/201 success vectors were authored
assumptions, not captured successful service behavior, and are now corrected.

The [pure parser reproducer](../../../scripts/discovery/inspect-cloud-call-data-parser.ps1)
and [observations](evidence/cloud-call-data-parser.json) invoke only audited
response callbacks/model constructors/serializers with synthetic data, with absent
HostContract before/after. Get/New callback IL hashes are
`59bc6c6df654b79aa6288c189b1852e3c945e5d44309a372001203d7d7cdf3c4`
and `daf58dc9429f4d8046d5afa2ebd9b1723b27325f826c3a22397963904fc6f81b`.
Empty text returns a null array; `[]` is nonnull/empty. Null/nonobject items become
null model entries. Object roots throw InvalidCastException, null roots throw
ParserException. Missing/null/numeric/wrong-case Token projects null; an empty string
stays empty. Generated serialization drops unknown fields and null Token. Native
RawFields retention is explicitly extra fidelity, not a claim of source serialization.

List now returns the complete array. Existing Get/Create signatures are retained
as strict unique-model conveniences, with a redacted CardinalityError retaining
all returned models and the HTTP receipt. Accepted Create cardinality errors also
retain CreateOutcome, invocation and 2xx-receipt flags. CreateWithResult exposes the
full action outcome directly. No item is arbitrarily selected or described as
newly created/owned; no compensation deletes returned/pre-existing resources.
An accepted empty/multiple result is not proof of failed mutation or safe replay.

Modern HTTP requests remain no-body GET/POST. Creation never replays; the source
command/default generated pipeline has no built-in retry loop. Native replay-safe
read retry policy is a documented client difference. Public ordinary proxies target
custom TrpsPSCmdlet classes, and CmdletConfig names Modern helpers; live flighting,
override and availability decide Modern versus legacy. Those static mappings do
not establish unconditional Modern selection or settle the legacy transport.
The read/modern action implementation and public-session/tenant qualification are
therefore reviewed separately, with no unsupported scope exclusion or promotion.

Focused SDK and lab compile checks passed in sessions 90029 and 83031 (TestUnit/Example only;
no TestAcc ran). New cases cover cardinality, source nullable parsing, redacted
inspectable accepted outcomes and unsupported 201 versus HTTP receipt. Actual lab
selection remains serialized and unchanged; its full-array, strict 200 handling keeps
authorization denial distinct from a token-bearing source response. Existing live
approval from 2026-11-23 is preserved; no new manual create was run for this audit.


## Selected Modern header handoff correction

Cloud-specific source review independently settles the invocation handoff rather
than assuming the custom-app path applies. The
[bounded packet](evidence/cloud-call-data-header-handoff.json) includes 34 complete
resolved methods across 14 selected types, the four exact proxy definitions,
two exact CmdletConfig entries and both helper property metadata records. Default
cloud descriptors have no AutoRest parameters; null defaults normalize to an
empty dictionary. Sync and async processors pass the selected Modern command
and bound/default dictionary, then add only the execution-ID prepend step.
Both private helpers' InvocationInformation properties are not parameters and
cache their own MyInvocation before Module.CreatePipeline. Relevant reflection
reads warning attributes; it does not inject a different invocation or normalize
request headers. The complete private-assembly scan resolved 449,218 InlineMethod
operands with zero unresolved references and no direct InvocationInformation
setter caller; this negative scan supplements the explicit handoff bodies.

The source-default selected Modern headers are
`Get-CsCloudCallDataConnectionModern_Get` and
`New-CsCloudCallDataConnectionModern_New`. Production now uses those names;
public accepted command names remain the ledger/test mapping. Expected HTTP
headers are source-authored. Existing 403 response captures are response-only
replays, not proof of a tenant request using the corrected names. The shared
header-writer/name-normalization proof is referenced by exact hash from the
custom-app closure, with no duplicate assertion of live management capture.

The portable `scripts/discovery/inspect-cloud-invocation-setters.ps1` reproduces
the direct-call scan; `inspect-ext-il.ps1` can re-read each exact selected type and
method from the verified package. No management lifecycle/authentication/HTTP ran.
Cloud public flighting, explicit mocks/config overrides, legacy transport and live
success remain outside this source-only selected-Modern proof. No cloud
qualification flag or live selection is promoted or changed.

The complete affected cloud SDK TestUnit/Example suite and corresponding lab
package compile passed at clean frozen revision `5ded640` (session 15595). The
[execution record](evidence/cloud-call-data-5ded640-offline.json) retains exact
code/source hashes and the raw Go result digest. The lab had no matching tests;
no TestAcc or tenant lifecycle ran. The portable setter-call scan was independently
rerun and exactly matched the source packet's 449,218 resolved operands and zero
unresolved references. All cloud qualification/live states remain unchanged.
