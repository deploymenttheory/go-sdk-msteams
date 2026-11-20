# Implementing a Teams SDK service

Start from the assigned ticket's exact command set and its completed research prerequisite. Read [decision 0001](../decisions/0001-client-and-service-composition.md), [decision 0002](../decisions/0002-portable-authentication.md) and [decision 0003](../decisions/0003-backend-contracts-and-compatibility.md). The agreed [implementation plan](../implementation/plan.md) remains the product baseline.

## Establish the contract

Record the upstream package hash, source paths/documentation revisions and sanitized observations. Resolve each parameter set, default where evidenced, nested model, omission/null rule, validation, output and error envelope. Record auth mode/roles, paging, replay safety, operations, transfers and local-only constructors. If a route or response remains unknown, complete the research ticket before implementing it.

Use the baseline tool to produce a new candidate snapshot. Never overwrite accepted evidence or classify a missing export as an automatic exclusion. BASE-02 owns conditional/public status; BASE-03 owns the command-to-symbol/test coverage ledger.

## Build within the package boundary

Add a resource package beneath the contract-selected API family, following the Jamf resource layout: methods, `models.go`, tests/mocks and `api_docs.md`. The root wires its service through constructor injection. Services depend on the shared client boundary, not root-client lookups, identity libraries or another resource's transport.

Return typed resources and Resty metadata with errors. Put `context.Context` first for remote calls. Separate read and write models; represent explicit null separately from omission when required. Keep actual backend identifiers and normalization rules. Do not use untyped maps as a shortcut for a known model, and do not invent local validation more restrictive than the upstream contract.

Local object builders produce typed values and validation errors without HTTP or credentials. Cross-resource nested types belong in a narrowly named shared model package when reuse is necessary; avoid a generic dumping ground.

## Verify and document

- Assert the exact request method/path/query/body/headers against sourced fixtures, including non-default values and omission/null/zero distinctions.
- Test validation, authorization failures, malformed responses, cancellation and the operation's evidenced paging/async/transfer cases.
- Test local builders without constructing a client and assert no network interaction.
- Compile an example and document auth roles, API version, known live-test prerequisites and the command-to-Go mapping.
- Add opt-in acceptance scenarios using owned fixtures, cleanup journals and restoration for shared settings. Missing lab capabilities remain explicitly contract-tested.
- Update the coverage ledger and run the applicable package tests, example compilation and shared quality gates. Commands are established by FND-01; documentation examples here do not claim those gates already exist.

## Public-interface compatibility checklist

- [ ] Existing exported names/signatures and Resty return conventions remain compatible, or the change explicitly documents its semver impact.
- [ ] Go method versions reflect API contracts, not PowerShell release numbers.
- [ ] New options validate locally and do not introduce authentication/network work at construction.
- [ ] No import cycle, mutable global client, hidden interactive fallback or runtime PowerShell dependency is introduced.
- [ ] Token audiences, tenant identities and backend trust boundaries remain isolated.
- [ ] Read/write models preserve explicit presence/null semantics and existing field meanings.
- [ ] Error wrapping retains context cancellation and typed remote failures.
- [ ] The change preserves retry/continuation/async semantics rather than adding resource-local loops.
- [ ] All changed commands, parameters and builders have evidence, tests and documentation; unavailable live cases are labelled honestly.
