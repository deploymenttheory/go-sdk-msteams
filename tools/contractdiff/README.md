# Semantic contract comparison

Compare two snapshots produced by `tools/baseline/baseline.py`. The comparison retains
old/new evidence and coverage ownership; it never adopts a version or changes the ledger.

```sh
python tools/contractdiff/compare.py --before /tmp/8.1.0.json \
  --after /tmp/candidate.json --output /tmp/impact.json
python -m unittest discover -s tools/contractdiff/tests -v
```

JSON and `.json.gz` inputs are supported (64 MiB uncompressed limit). Output creation
is immutable: an identical existing report is accepted, a different report is not replaced.
No tenant connection, module install or network access occurs during comparison.

| Exit | Status | Meaning |
|---|---|---|
| 0 | `no-delta` | No compared evidence changed and no schema/provenance warnings |
| 1 | `needs-review` | Changes, unknown fields, changed package bytes or observation environments need review; report written |
| 2 | `incomplete` | Missing/malformed extraction or output error; never interpret a previous report as this run's result |

## Compared evidence

Case-insensitive command identities, command types and alias targets; parameter-set
membership/default selection; parameter type, mandatory flag, position, aliases,
pipeline flags, validators and explicit defaults; output type names; Gallery and nested
manifest candidates; package command configuration; and packaged help/release-note changes.
Case-only names and source/command enumeration order do not create semantic deltas.
Missing values remain distinct from explicit null, false and zero.

Defaults are compared when explicitly present; the current PowerShell extractor cannot
establish every runtime default. Similarly, output type names do not describe all returned
fields. Do not infer a stable default or full output schema from missing metadata.
An alias without an extracted target fails visibly. Unknown fields are retained or flagged
for review even if unchanged on both sides. Empty/incomplete inventories and missing
help/configuration/provenance cannot yield `no-delta`.

Documentation changes remain `documentation`. Ordinary inventory/configuration changes
are `metadata`. Script files with only hash evidence are `unclassified`; a changed script
hash does not prove a changed route or authentication rule. A changed package hash also
prevents a clean result because DLL behaviour is not established by signature metadata.
The report's `no-delta` is scoped to the supplied evidence, never a promise that Microsoft
server behaviour is unchanged.

A rename appears as removal plus addition. Preview and stable channels are recorded
separately; neither comparison direction implies promotion. Compatibility hints
(`additive`, `potentially-breaking`, `review`) are review inputs, not automatic SDK
semantic-version decisions.

## Observed or documented contracts

Use `--contracts-before` / `--contracts-after` for sanitized, reviewed contract facts
outside discoverable signatures. Each file is an array:

```json
[
  {
    "command": "Get-CsCallQueue",
    "classification": "documentation",
    "evidence": "path/to/reviewed-evidence.json#/claims/0",
    "facts": {"output_fields": {"Statistics": false}}
  }
]
```

`observed-behaviour` is reserved for supplied observation evidence. It is not inferred
from metadata or a release note. Facts may describe output schemas, defaults, routing,
authentication or validation. Evidence references and both values are preserved; changing
between documentation and observation provenance is itself flagged. The comparator
requires provenance fields but does not verify remote URLs or attest that a supplied
observation was actually executed. Capture/review and LAB qualification own that proof.
Only sanitized schema facts belong in these inputs; raw credential-bearing captures do not.

Every change includes the resource family, stable research/implementation ticket, GitHub
issue link, implementation/model symbols and mapped tests where the coverage ledger has
them. Unknown commands are retained with DELTA-01 triage ownership and an explicit
assignment-required flag. Global/documentation changes also retain that triage owner;
missing resource mappings never disappear from the report. No invented Go symbol paths
are emitted for services not yet implemented.

Snapshot canonical hashes identify the precise input objects. `normalized-snapshot:`
paths describe the semantic comparison model; full old/new values are embedded in the
report. They are not claimed to be JSON pointers into the raw runtime-array format.

## Historical regression evidence

The offline fixtures preserve complete command/source metadata acquired from verified
8.0.0 and pinned 8.1.0 packages. See [fixture provenance](tests/fixtures/provenance.json)
and [fixture notes](tests/fixtures/README.md). Regression tests detect the real new phone
usage-order command, `InviteWithReplaces` and messaging parameter additions. A separate
**documentation** projection tests the release-note removal of CallQueue `Statistics`;
it is not represented as a live observation or a complete CallQueue output schema.
