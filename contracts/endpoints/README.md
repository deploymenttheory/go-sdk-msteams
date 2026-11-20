# Endpoint contract evidence

Version directories contain wire-contract artifacts derived from the pinned module
or dated primary API documentation. They do not replace the immutable PowerShell
inventory or imply implementation, tenant authorization or release readiness.

The 8.1.0 `core.json` inventory maps all 40 CORE-01 commands and their exact runtime
parameter sets to 60 referenced wire primitives. References include branches and
compensation; their order is not an executable workflow. The inventory explicitly
retains open questions and does not claim accepted-contract status.

The initial 8.1.0 fixtures are **offline serializer observations** for CORE-01.
Each artifact records its package/assembly identity and capture method. Synthetic
IDs and content are used throughout. A serializer observation establishes bytes
or JSON values produced by the client; it does not prove server acceptance.

Reproduce the team fixtures using the development oracle after downloading and
verifying the pinned package through the baseline workflow:

```sh
pwsh -NoLogo -NoProfile -NonInteractive \
  -File tools/contracts/capture-team-serialization.ps1 \
  -ModulePath /path/to/verified/MicrosoftTeams/8.1.0 \
  -OutputPath /tmp/core-team-candidate.json
```

Run from the repository root. The capture checks the serializer assembly SHA256
and module version, refuses an existing output, and invokes only local model
construction/serialization. It does not connect or run a management command.
Compare parsed JSON, excluding the capture date if reproduced on a later day;
property order is not part of this JSON contract. Review differences before any
accepted artifact changes. PowerShell is not an SDK runtime dependency.

Resource contracts must separately identify CRUD/list methods, explicit actions,
PowerShell compositions, models, route/audience requirements, partial failures and
unresolved evidence. See [CORE-01 research](../../docs/implementation/core/baseline-inspection.md)
and the [CRUD amendment](../../docs/implementation/plan-amendments.md).

The second capture script, `tools/contracts/capture-core-models.ps1`, covers channel
creation/patch, EDU group defaults, template Body versus expanded construction,
empty arrays/false, sparse policy input and deployment wire-field casing. The
deployment Base64 value is synthetic and is not a valid deployment archive. This
is model serialization evidence, not a tenant request fixture.

Response/error round trips additionally cover template envelopes, deployment IDs,
policy additional fields, hierarchy status projection, known Graph operation
states, lost operation-error information and unknown-enum rejection. See
[response evidence](../../docs/implementation/core/response-model-evidence.md).

Run `python tools/contracts/verify_models.py` for inventory checks, or add
`--module /path/to/verified/module` to reproduce every checked capture. `--acquire`
downloads the pinned archive, verifies its exact SHA512, safely extracts it and
runs the capture scripts in temporary paths with tenant/CI credentials excluded from
the child environment. CI exercises that mode on Linux, macOS and Windows.
Comparison ignores the capture date and JSON object-key order; type, omission,
array order, false, empty values and every other artifact field remain exact.
Captured values and child-process output are excluded from failure diagnostics.

## Generated request and policy helper capture

`capture-generated-requests.ps1` compiles the development adapter in
`tools/contracts/oracle/GeneratedRequestCapture.cs` against the pinned generated
assembly. An injected `ISendAsync` records requests and supplies synthetic responses;
an `IEventListener` supplies ordinary or canceled context. The helper does not use
the production module pipeline or an HTTP client. The adapter rejects unexpected
credential headers and does not invoke cmdlet processors or authentication.

The 24 cases exercise actual generated CRUD/list, user/tenant/group assignment,
template and deployment methods: escaped segments, JSON bodies, status callbacks,
decoder failure, cancellation before send and empty template-delete 204. The
captured URL is **before session host/header rewriting**. In particular, the
generated deployment development host is not an SDK production endpoint. Shared
session/discovery callbacks remain a separate contract and test obligation.
Callbacks await the generated decoder, record null/failure without error text,
and do not emulate PowerShell output/error-stream formatting.

`capture-policy-models.ps1` additionally exercises the pinned AppIds list-modifier
helper, migration DTOs and compiled provider serializers. Omission, null/empty
replacement, Add/Remove, empty operator objects and rejected legacy/numeric inputs
have separate cases. Pure private helpers capture bound-field selection, nested
branding lists and first-slash identity escaping; no HTTP helper is invoked.
Other cases preserve numeric external-communication enums, voicemail TimeSpan
strings and the three compiled assignment DTOs' distinct casing/null behavior.
The current totals are 73 model/helper observations and 24 generated request/dispatch
observations. All are synthetic development evidence; the same verifier recaptures
them after pinned acquisition on the three CI platforms.
