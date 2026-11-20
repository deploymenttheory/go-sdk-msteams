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
`--module /path/to/verified/module` to reproduce all 38 model cases. `--acquire`
downloads the pinned archive, verifies its exact SHA512, safely extracts it and
runs both captures in temporary paths with tenant/CI credentials excluded from
the child environment. CI exercises that mode on Linux, macOS and Windows.
Comparison ignores the capture date and JSON object-key order; type, omission,
array order, false, empty values and every other artifact field remain exact.
Captured values and child-process output are excluded from failure diagnostics.
