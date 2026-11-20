# CORE-27 SDK progress

Status: **in progress**. Only the two local hidden-template builders are
implemented so far. Remote template Create/Get/Update/Delete/List, typed models,
collection wiring, service fixtures and tenant acceptance remain required. This
work does not close [CORE-27 / #80](https://github.com/deploymenttheory/go-sdk-msteams/issues/80).

`teamtemplate.NewHiddenTemplate` and `NewHiddenMeetingTemplate` construct local
ID values, reject an empty ID with `ErrInput`, and perform no management or
authentication operation. They do not persist a permission policy. Colocated
tests use CORE-15's `scope: "local"` runner and exact Identity-set mappings;
ordinary constructor tests cover ID retention and invalid input.

The actual pinned 8.1.0 local PowerShell cmdlets were also executed with empty,
whitespace and synthetic opaque IDs. Both rejected empty IDs with
`ParameterBindingValidationException`, accepted whitespace/opaque IDs unchanged,
and left Identity unset. No connection or tenant management call was requested.
This confirms the limited validation boundary; the Go constructors do not invent
a whitespace or server-existence check. Policy serialization remains owned by
the relevant permission-policy service.

Before building typed template models, three additional observations were
captured through the shipped generated serializers:

- Existing root fields win over nested `teamDefinition`, then
  `extendedProperties` fills fields still absent.
- An existing root null masks a nested string; the known string is omitted on
  model serialization rather than replaced by the nested value.
- Null entries in app/category/channel/tab arrays are skipped on serialization.

These append to `core-model-serialization.json`; the original 30 observations
and their indexes remain unchanged. The capture helper verifies pinned module
version and assembly hashes before local model operations. The source review of
`TeamTemplate.BeforeFromJson` and `JsonObjectExtensions.Merge` agrees with the
new runtime observations. No decompiled implementation code is committed.

Local builder race/lint checks pass. No command/parameter-set ledger row is
promoted, and no remote resource or live tenant success is claimed by this
partial implementation. The policy PR remains a dependency for the shared
ConfigAPI discovery adapter; its tenant validation awaits the separately
requested Teams Administrator role assignment.
