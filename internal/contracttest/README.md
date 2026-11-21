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

- Curated name, resource family, repository fixture reference and exact
  command/parameter-set mappings. Sources accept JSON under `contracts/` or
  a service's `msteams/**/testdata/`, with an optional JSON pointer. Ordinary
  input vectors can reference their `*_test.go` file under `msteams/`, `tests/`
  or `internal/`. Absolute
  paths, traversal and backslashes are rejected. `Scope` is `service`, `primitive`
  or `local`.
- Ordered `Exchange` attempts, each with method, full escaped HTTPS URL, audience,
  selected headers and either a JSON body or exact binary bytes, followed by its
  synthetic response. `WireRequest.JSON` and `WireRequest.Binary` are mutually
  exclusive. Both nil require no body; a non-nil empty binary slice requires zero
  bytes. Binary assertions preserve all bytes and never parse or normalize JSON.
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
the canonical cloud and selected mode are retained in the returned result.

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

## Scope and limits

The runner asserts the real SDK request and response behavior against permanent
unit fixtures. It does not write per-run contract artifacts or promote coverage
ledger states. Resource tests own their command and parameter-set coverage checks.

A passing unit test proves the selected assertions against its supplied fixture.
Tenant CRUD, PowerShell comparison and cleanup require separately executed
`TestAcc` tests and their recorded acceptance results. Unit fixtures derived from
acceptance runs must be sanitized before they enter `testdata`.
