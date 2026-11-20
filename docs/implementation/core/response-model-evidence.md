# CORE-01 response and error model observations

Research revision: 2026-11-20. This extends the pinned model oracle from 21 to
38 observations (eight team cases and 30 core cases). All inputs are synthetic
and all output comes from local shipped model constructors/serializers. These
are not tenant response captures or successful service-authorization claims.

Sources are the same SHA256-pinned TeamsCmdlets and generated Config API
assemblies recorded in [the fixture artifact](../../../contracts/endpoints/8.1.0/fixtures/core-model-serialization.json).
The [capture script](../../../tools/contracts/capture-core-models.ps1) verifies
those hashes and module version before importing. No ProcessRecord, connection,
request helper, token provider or management operation is invoked. The script
records exception occurrence only for explicit negative parser cases; it retains
no exception text or environment credentials.

## Templates and policy envelopes

The generated `Models.TeamTemplateErrorResponse` round trip retains operationId
and ordered errors containing code/message/action, while discarding the synthetic
futureField. Explicit errors `[]` remains an empty array. An empty JSON object
produces `{}`. Separately, `FromJsonString("")` returns null without throwing;
this settles that model parser's empty-text behavior, not every HTTP status
handler's subsequent null handling.

`GetTemplatesResponse` retains its value array and summary counts, including
appCount zero. A detail response decodes the nested template and populates its
default false settings, as the existing request-input observations also showed.
An absent settings object in the original response must not be mistaken for
evidence that the server explicitly sent every false setting. The raw response
and typed projection have different information content.

`FlwoServiceModelsFlwosBatchRequestResponse` retains both instanceId and executionId.
The SDK's creation result must preserve both, although the PowerShell formatted
output emphasizes orchestration identity. This fixture does not establish raw
deployment-status grammar or turn Completed into terminal success.

Generic `Models.ErrorResponse` retains code/message/action and drops an unknown
field, while `Models.XdsConfiguration` preserves arbitrary additional properties
including false and empty arrays. This confirms that forward-compatible policy
fields need a retained additional-properties representation, whereas the pinned
template/error models themselves do not provide one. API errors still need safe
printable diagnostics; explicit raw responses remain caller-controlled.

Sources: generated types `TeamTemplateErrorResponse`, `GetTemplatesResponse`,
`GetTemplateDetailResponse`, `FlwoServiceModelsFlwosBatchRequestResponse`,
`ErrorResponse`, `XdsConfiguration`, each FromJsonString/ToJsonString. Routes,
statuses and handler behavior remain in [the family evidence](template-policy-deployment-evidence.md).

## Hierarchy status projection

The exact `GetTeamTargetingHierarchyStatus.Result(HierarchyServiceOperation)`
constructor produces these public status strings from the synthetic wire model:

| Wire status | PowerShell projection |
| --- | --- |
| NotStarted | Starting |
| Running | Ingesting |
| Succeeded | Successful |
| Failed | Failed |
| UnknownFutureValue | Unknown |
| FutureStatus (unknown synthetic value) | FutureStatus |

The projection also renames lastActionDateTime to LastUpdatedAt and retains the
operation ID/file name. These are distinct from Graph teamsAsyncOperation status
names. A native operation model should retain original status alongside any
normalized state. An unknown status remains incomplete/unknown rather than
success. See [the hierarchy contract](hierarchy-migration-contract-evidence.md)
for the separate resource audiences, version selection and error structures.

## Pinned team-operation model limitations

The shipped `Model.TeamsAsyncOperation` parses known string enums and reserializes
their CLR names (`failed` becomes `Failed`, `createTeam` becomes `CreateTeam`). It
retains attempts/timestamps/resource identity. Its class has no error property:
the synthetic input error is dropped during the model round trip. An unknown
status string is rejected by its StringEnumConverter.

These observations describe limitations of the pinned client model. The native
SDK already exposes the [Graph operation adapter](../../../msteams/graph_api/operation.go)
and must preserve operation errors and unknown states under the accepted
[backend contract](../../decisions/0003-backend-contracts-and-compatibility.md).
PowerShell behavioral parity does not require discarding useful error information
or interpreting a decoder limitation as a successful operation. The pinned
observation and native API behavior must remain separately documented in the
command composition mapping. Team-operation permission qualification is still
open; chat-operation permissions are not transferable evidence.

## Verification and remaining work

`python tools/contracts/verify_models.py --module /path/to/verified/module`
reproduced all 38 cases as parsed JSON. Comparison preserves null/omission,
zero/false, array order and every field except capture date. Existing CI runs the
same scripts after SHA512-verified acquisition on all three target platforms.

These model observations strengthen request/response fixture prerequisites.
CORE-01 still requires complete per-command wire fixtures, presence/validation
mapping, the documented compatibility decisions and conditional service-routing
qualification. The capability ledger remains unchanged.
