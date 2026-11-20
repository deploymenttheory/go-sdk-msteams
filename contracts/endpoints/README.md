# Endpoint contract evidence

Version directories contain wire-contract artifacts derived from the pinned module
or dated primary API documentation. They do not replace the immutable PowerShell
inventory or imply implementation, tenant authorization or release readiness.

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
