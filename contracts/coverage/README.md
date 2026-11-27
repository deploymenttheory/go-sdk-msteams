# Capability and evidence ledger

[`8.1.0.json`](8.1.0.json) is the current SDK coverage ledger. Unlike the immutable
PowerShell observations and approved publication preview, it changes as implementation
and qualification progress. [`schema.json`](schema.json) defines its versioned structure.
`tools/coverage/ledger.py` validates the schema subset used here and the cross-file rules.

The ledger retains all **745** reconciled candidates: 730 original candidates and 15
nested declarations. All original management/session/builder entries retain their
approved implementation owner. The 15 additional nested candidates have EXT-01 research
ownership because their public support and configuration/app contracts remain unresolved;
they have no implementation owner or implementation claim. They block release readiness.
The exact original export gap `Set-CsNotifyCache` keeps its EXT-23 implementation owner.

Recorded exclusions include module telemetry, the module-local token helper and
NotifyCache without an evidenced wire contract; each ledger row links its rationale.
A missing feature in the lab is never an exclusion. Session and common-parameter mappings describe their
native equivalents; these mappings alone do not count as implemented behaviour.

## Fields and states

Each command has an identity, disposition, resource family, research owner, implementation
owner and JSON pointer into the reconciliation evidence. Every observed parameter set
has its own exact runtime pointer and lists of Go implementation, contract-test and
live-test references. Go references contain a repository-relative file and a function,
type or `Receiver.Method` name. Resource-specific parameters remain in the referenced
PowerShell metadata; `common_parameters` records the SDK mapping for shell mechanics.

Three independent dimensions prevent discovery from being mistaken for delivery:

| Dimension | States | Evidence required for success |
|---|---|---|
| Implementation | not-started, in-progress, implemented, unsupported | Existing Go symbols for the command and each parameter set; no unsupported-stub flag |
| Contract | metadata-only, passed, failed | Implemented command; tests for every set; successful versioned run evidence |
| Live | not-run, passed, failed, skipped, unavailable | Passed contract; live tests for every set; successful versioned live run evidence |

A non-passing live state needs a reason. `unavailable` can coexist with completed contract
coverage; it remains visible in the report and does not turn into a live pass. The QUAL
release gate additionally evaluates available tenant conformance; this checker cannot
infer tenant availability.

A successful run artifact is a repository JSON file (or JSON pointer) containing at
least `outcome: "passed"` and `revision` identifying the tested implementation revision.
Store sanitized execution summaries, not tokens, raw tenant identifiers or credentials.
Evidence with a skipped/failed outcome cannot back a passed state. Missing metadata,
foreign/missing files, stale names and command/parameter-set mismatches fail validation.

The checker performs structural Go symbol lookup and verifies declared evidence. It
does not prove arbitrary Go semantics, authenticate to a tenant or execute referenced
tests. A contributor must supply actual CI/tenant run evidence and review whether the
implementation is complete; a symbol's existence is not behavioural proof. Unsupported
implementations must retain `status: "unsupported"` or `unsupported_stub: true`, and
cannot claim implementation or contract success.

## Updating coverage

1. Resolve the command's contract and scope in its family research ticket. Keep newly
   discovered candidates represented; assign a known implementation task once in scope.
2. Implement the resource/builder/session behaviour and add meaningful tests.
3. Add command and parameter-set symbol/test mappings and sanitized run evidence. For
   conditional commands without runtime signatures, first add reviewed contract metadata
   with explicit parameter-set names and evidence references; absence cannot count as an
   empty supported surface.
4. Update implementation, contract and live states independently. Check the report and
   include the mapping/evidence change in the implementation PR.

```sh
python -m unittest discover -s tools/coverage/tests -v
python tools/coverage/ledger.py
# Full implementation/contract gate, expected to fail during development:
python tools/coverage/ledger.py --require-complete
```

Output is machine-readable JSON; failures exit nonzero. Normal CI checks integrity while
allowing incomplete work. The release mode rejects unresolved scope and any non-excluded
entry without complete implementation and contract evidence. `complete` includes approved exclusions;
`implemented`, `contract_tested` and `excluded` are reported separately. Implementation, contract and live counts are reported from the current ledger.
Existing service code and passing individual scenarios do not by themselves prove
complete command/parameter-set qualification. The publication preview remains historical; it is not regenerated for each
implementation change.

The checker runs on every PR, including Go-only changes, so deleting a mapped symbol or
test fails CI. Unknown additional ledger fields are rejected rather than silently
creating an alternate meaning for coverage status. Schema changes require an explicit
validator and fixture update.
