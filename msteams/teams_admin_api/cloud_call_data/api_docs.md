# Cloud call data connections

`TeamsAdminAPI.CloudCallData` implements the source-backed Modern ConfigAPI
resource at `/Teams.InternalSupport/CloudCallDataConnection`. Constructors perform
no I/O; methods use the existing ConfigAPI audience/forest/session controls.
The public PowerShell proxies choose Modern versus legacy through session flighting;
this native resource does not emulate legacy remoting or claim universal selection.

| Method | Wire/result |
| --- | --- |
| `List(ctx)` | GET; full returned connection array, HTTP 200 only |
| `Get(ctx)` | Unique-result convenience over List |
| `CreateWithResult(ctx)` | POST with no body; full typed action outcome, HTTP 200 only |
| `Create(ctx)` | Unique-result convenience over CreateWithResult |

Existing Get/Create signatures remain available. Their previously authored object
response/201 assumptions are corrected: both Modern parser callbacks require an
array and both dispatch success only at 200. Object/null roots and unsupported
2xx statuses fail with the response retained; 201 is not a declared success.

List preserves empty text as a nil array, `[]` as an allocated empty array, and
null/nonobject entries as nil models. Each object preserves exact raw fields.
The pinned projection recognizes uppercase `Token` only; absent/null/non-string
or wrong-case values yield nil typed Token, while empty string remains nonnil.
RawFields retains data ignored by that projection, an explicit native extension.
Tokens and response bodies are credentials: inspect deliberately and never log them.
Connection/result/cardinality-error String/GoString values are redacted.

Get/Create require exactly one non-null returned model. Zero, multiple or a lone
null entry returns `ErrCardinality`, not an arbitrary element. Use
`errors.As(err, &cardinality)` with `*CardinalityError` to inspect every Returned
model and Response; for Create it also retains the full CreateOutcome, invocation
and HTTP-acceptance flags. This error does not establish that creation failed or
that no connection exists. Never replay blindly after an accepted/uncertain action.

CreateWithResult records attempted invocation and a 2xx receipt separately from
parsed/source-declared success. Its returned connections are not assumed newly
created or owned. There is no remove operation or automatic compensation.
List/Get are replay-safe under the native shared policy; creation is ReplayNever.
The pinned default pipeline has no built-in retry loop, so native read retries
are an explicit client-policy difference. No extra grants, roles or fallback
identity are acquired.

The app-only dedicated lab historically returned 403 for both operations; the
2026-11-23 irreversible-create authorization is preserved. Denials and authored
array replies do not qualify successful creation, token usability, ownership or
persistence. See [primary/source evidence](../../../docs/implementation/ext/cloud-call-data-contract-notes.md).


The selected Modern helpers use `X-MS-CmdletName` values
`Get-CsCloudCallDataConnectionModern_Get` and
`New-CsCloudCallDataConnectionModern_New`. The native resource now follows that
source metadata instead of the earlier public-name convention. The bounded
[handoff proof](../../../docs/implementation/ext/evidence/cloud-call-data-header-handoff.json)
independently traces Modern command resolution, default parameters, Sync/async
processors, helper invocation properties and the shared header writer. There is
no default hidden invocation override in that path. This does not prove the
public session selected Modern or settle legacy request metadata.

Historical 403 captures supply response bytes for controlled offline error replay;
they do not prove that a tenant request with the corrected private headers was
rejected. No new live request or creation follows from the source/replay checks.
The existing serialized lab selection and historical authorization are unchanged.
