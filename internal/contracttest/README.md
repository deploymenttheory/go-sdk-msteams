# Offline service contracts

`contracttest` is shared test infrastructure, consumed from external resource
test packages (`package resource_test`). It injects an in-memory HTTP transport
and synthetic provider into the real root SDK client. It reads no environment
credentials, opens no sockets, and has no tenant mutation or PowerShell runtime
dependency. The callback invokes the resource under test; the runner owns and
closes its SDK client.

Register resource-owned `Case` values in a zero-value `Suite`, then call `Run`.
Cases execute in family/name order. A single case can instead use
`contracttest.Run(ctx, case)`. A setup/fixture error is separate from a completed
comparison: check both the returned error and `Result.Passed()`. An unexecuted
zero-value result never passes. Finish callback goroutines before returning;
registration and fixture input must remain unchanged during execution.

Each case provides:

- Curated name, resource family, repository contract reference and exact
  command/parameter-set mappings. `Scope` is `service`, `primitive` or `local`.
- Ordered `Exchange` attempts, each with method, full escaped HTTPS URL, audience,
  selected headers and JSON body, followed by its synthetic response.
- An `Invoke(context.Context, *msteams.Client) Observation` callback adapting the
  actual service value, Resty response and error.
- An `Expected` result selecting JSON value, raw response bytes, status, headers,
  `errors.Is` classification and/or complete normalized `client.APIError` fields.

The transport checks the token sent at **each attempt**, even when another
audience is valid for a later exchange. The provider separately checks tenant,
cloud, identity and mode binding. `Cloud` selects a built-in SDK environment;
fixtures must supply its matching destinations/audiences. No destination can
escape to the network, including an unexpected call through the session's
borrowed HTTP client. Missing and extra attempts fail. Include discovery,
read-back, paging and explicit retry attempts as separate exchanges when the
service contract calls for them. Replay policy remains the real SDK policy.
`Mode` selects application (the default) or delegated fixture authentication;
the canonical cloud and selected mode are retained in the structural report.

Paths compare their escaped representation, so an opaque ID containing `/`
must remain one segment. Query key order is immaterial; repeated value order and
all values remain exact. Header assertions compare values in order; unspecified
transport-owned headers are allowed. Explicitly include every service-owned
header that the case needs to constrain.

JSON object order is immaterial. Missing and extra fields, explicit `null`, empty
objects/arrays, false, array order, and number spelling remain distinct. Numbers
never pass through a floating-point conversion. Duplicate object keys, malformed
JSON and excessive nesting are rejected. Unknown response fields are checked if
included in the expected JSON; no ignored-field list silently drops them. A nil
request `JSON` requires **no body**; `json.RawMessage("null")` requires a JSON
null body. A nil result `JSON` skips model comparison. A non-nil empty `Body`
requires empty raw response bytes.

Mismatch paths name the attempt/result boundary and structural position, such
as `request[0].audience` or `result.value.object[1]`. Object positions follow the
sorted union of field names. Paths contain neither raw values nor arbitrary
response field names, preventing an unknown key from leaking private content.
Fixture references let reviewers locate the expected field by that position.

`local` cases require no HTTP exchanges and fail if a constructor/helper acquires
credentials. Pre-canceled operation cases can similarly expect no exchanges and
`context.Canceled`. The runner's setup always uses explicit configuration;
ambient `MSTEAMS_*` authentication settings cannot change it.

## Evidence and limits

`WriteEvidence(writer, revision, results)` accepts a full Git revision and
executed results. It produces a stable JSON artifact with version, outcome,
scope, contract/mapping references, attempt/acquisition counts and structural
differences. It omits URLs, headers, payloads, tokens, tenant/application bindings
and raw error strings. Names/mappings/source references are curated metadata;
never derive them from tenant content. Writer failures return a fixed error.

The writer does not update `contracts/coverage/8.1.0.json`. An artifact has
`coverage_claim: false` and `live: "not-run"`. Resource owners must review
complete command/parameter-set coverage and provide implementation/test symbols
before changing the ledger. A passing primitive or local-builder case alone
cannot establish implemented service or full PowerShell parity.

The initial real capture consumer is
[`TestPinnedTemplatePrimitiveFixtures`](../../tests/contract/core/fixtures_test.go).
It runs five captured template wire primitives through the shared ConfigAPI
scope, checks command/set mappings against pinned runtime metadata, and explicitly
does **not** claim a template service implementation. Nine family-registration
checks in the runner unit tests prove reuse, not nine delivered resource suites.
Resource tickets own their exact service fixtures, cloud/auth restrictions,
modern/legacy compositions and tenant acceptance.

Ordinary `go test ./...` runs all these offline checks. CI passes the exact
checked-out revision in `MSTEAMS_CONTRACT_REVISION` during its existing race run,
uses Go's `-artifacts -outputdir` support, and uploads only `core-primitives.json`.
Local artifact generation should use a clean committed implementation revision:

```sh
mkdir -p /tmp/msteams-contract-artifacts
MSTEAMS_CONTRACT_REVISION=<full-tested-commit> \
  go test -race -count=1 -artifacts -outputdir /tmp/msteams-contract-artifacts \
  ./tests/contract/core
```

Omitting the revision still runs all cases and writes no persistent run artifact.
These artifacts never replace a real tenant CRUD and PowerShell parity test.
