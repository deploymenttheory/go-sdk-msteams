# Policy package tenant tests

The SDK implements package reads, custom create/update/delete, synchronous user
assignment, group assignment and distinct asynchronous batch submission. Ordinary
JSON-backed tests cover request bodies, partial business failures, retained HTTP
receipts, omission/empty values and cancellation. See the [SDK reference](../../msteams/teams_admin_api/policy_packages/api_docs.md).

Application-mode sessions are rejected locally for the commercial, GCC, GCC High
and DoD profiles before discovery or token acquisition. Successful live package
reads and writes need a separately supplied delegated identity. An application
permission grant does not satisfy that prerequisite.

`TestAccPolicyPackagesReadOnly` uses `MSTEAMS_LAB_PACKAGE_DELEGATED_FILE`, a private
JSON file with `tenant_id`, `identity_id`, `access_token` and RFC3339 `expires_at`.
Optional `cloud` selects the environment; optional `user_id` enables selected user
and recommendation reads. The loader rejects a wrong tenant, empty/expired token,
unknown fields and trailing data. Tokens and identifiers are never printed.

The current automated tenant suite has no delegated package fixture, so it does
not claim package CRUD or effective assignments passed. Future writes require
explicit mutation gates, disposable package/user/group subjects and independently
verified cleanup/restoration. An accepted batch ID is not effective-policy success.
