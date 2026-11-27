# Conditional Teams app policy configuration

Use `sdk.TeamsAdminAPI.TeamsAppPolicyConfigurations`. These Get/Set commands appear in the
pinned OCE manifest; a default-session root import does not export them. Routing
is not proof of tenant availability, entitlement or successful live operation.

| Method | Compiled behavior |
| --- | --- |
| `GetByIdentity(ctx, identity)` | GET one explicit configuration identity |
| `List(ctx)` | GET the bare collection, ordering Global first |
| `ListFiltered(ctx, pattern)` | Apply pinned local Identity/Tag wildcard filtering; explicit empty is invalid |
| `PatchByIdentity(ctx, identity, PatchRequest)` | One sparse PATCH, no preparation read |
| `UpdateByIdentity(ctx, identity, UpdateRequest)` | Set-metadata GET followed by sparse PATCH |

`AppCatalogURI` and `ResourceURI` retain the wire keys `AppCatalogUri` and `ResourceUri`. Explicit null is supported; empty or non-HTTPS-prefix values fail before credentials. Validation mirrors the pinned case-insensitive HTTPS-prefix expression, not full URL parsing. Its .NET end anchor admits one terminal LF; this is scalar data, never an HTTP destination.

Requests use the direct compiled PolicyRP protocol with `MPACmdlet` and
`X-MS-CmdletName`, without generated-client forest discovery. The resource path
is `/Skype.Policy/configurations/TeamsAppPolicyConfiguration`.
The API exposes no unobserved create/delete/grant operation.

`Settings` uses `policy.Field` to distinguish omission, selected values and null.
`policy.Controls.Force` preserves an explicitly bound false value; Force alone
is not a setting. `UpdateByIdentity` retains the preparation snapshot/response,
`PatchInvoked` and `HTTPAccepted` after later errors or cancellation. Acceptance
does not prove persistence or effect; the caller must independently read back.

`Configuration` keeps typed nullable fields plus owned `RawFields` and `RawJSON`
for complete snapshot round trips. Known invalid model values fail decoding;
unknown fields and their null/empty presence remain intact.

See [unit tests](service_test.go)
for exact assembly hashes, model-setter/serializer observations and the separate
compiled declaration sets. No live tenant mutation is configured for this
resource; Global settings require explicit caller authorization.

An omitted Filter corresponds to `List`. The actual compiled Filter property has
`ValidateNotNullOrEmpty`; `ListFiltered("")` therefore fails locally. Whitespace
is a literal pattern, not omission.
