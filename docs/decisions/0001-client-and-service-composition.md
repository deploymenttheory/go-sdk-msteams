# Decision 0001 Client and service composition

Status: accepted architecture from the agreed implementation baseline; production SDK code is delivered by the foundation and resource tickets. Implements [ARCH-01](https://github.com/deploymenttheory/go-sdk-msteams/issues/21).

## Context

MicrosoftTeams 8.1.0 exposes management operations through multiple backend families and also exports local object constructors. A single Graph wrapper cannot account for the full management surface. The agreed reference is the [Jamf SDK at 25714da5](https://github.com/deploymenttheory/go-sdk-jamfpro-v2/tree/25714da557956da9e317f725ca3d22e8818ec77e); its composition and response conventions apply, while its endpoints, RSQL, token keep-alive and server-version assumptions do not.

## Decision

The module is `github.com/deploymenttheory/go-sdk-msteams`. The public `msteams` package lives in `msteams/` (import `github.com/deploymenttheory/go-sdk-msteams/msteams`) and owns construction and service wiring. Go source and its package tests belong under that directory, not the repository root. It exposes `NewClient(authConfig, options...)` and `NewClientFromEnv(options...)`, returning a concrete client and error. Options validate configuration; construction does not authenticate or perform network discovery. Environment loading occurs only in the explicit environment constructor. Explicit options are applied after loaded configuration and validated together.

The root has `GraphAPI` and `TeamsAdminAPI` service collections. The collections group typed resource services; they do not contain transport implementations. Resource packages follow `msteams/graph_api/<resource>` and `msteams/teams_admin_api/<resource>`, with models, methods, tests, mocks and `api_docs.md`. Package identifiers use ordinary Go naming; established directory names preserve the requested Jamf layout.

The root injects a shared `client.Client` request-building interface into services. It retains one transport/authentication session per client. Service constructors perform local wiring only. No `init` authentication, mutable global client, service locator or DI framework is permitted. Shared wire primitives belong in narrowly named packages beneath `shared`, not in resource packages imported by the transport.

Allowed dependency direction:

```text
msteams root -> resource services -> client interface and request types
             -> client transport -> auth provider contract
             -> config           -> auth/configuration types
resource services -> narrowly scoped shared model primitives
local builders -> local model and validation types
```

`client`, `auth`, `config` and `shared` must not import the root or a resource service. A resource must not import the root to reach another service. Cross-service operations receive explicit dependencies or share pure model types; the root wires them. This makes the package graph acyclic without a runtime registry.

## Public interface conventions

The user clarified on 2026-11-20 that this is a **resource CRUD SDK** (recorded in
the [plan amendments](../implementation/plan-amendments.md)). Resource
services expose `Create`, `GetByID`, `UpdateByID`, `DeleteByID` and `List` where
those operations are supported by the evidenced service. Identity-based resources
use `GetByIdentity`, `UpdateByIdentity` and `DeleteByIdentity`; singleton resources
use `Get` and `Update`. A list returns a collection/page and remains distinct from
reading one identified resource. Methods take typed inputs and return typed results,
Resty response metadata and errors under the conventions below.

Specific actions remain explicit methods: for example `Archive`, `Unarchive`,
`CreateFromGroup`, `UploadPicture`, `PromoteOwner` or `DemoteOwner`. Model membership
as a resource with its own supported create/read/update/delete/list operations.
An action is not automatically a CRUD update merely because it uses POST or PATCH.
Local builders remain pure constructors rather than remote create operations.

PowerShell parity is a **mapping to one or more SDK operations**, not a requirement
to expose one SDK method per cmdlet. Each contract records primitive CRUD/action
operations separately from the compositions needed to reproduce a command's
behavior. For example, `New-Team` may compose group creation, team creation from
that group and reads; `Set-Team` may compose group and team updates. These sequences
are not atomic: compositions retain partial results, affected resource/operation
identities and errors so callers can recover. Any compensation has an explicit
ownership boundary and must not delete an existing caller-supplied resource.

Resource contracts and implementation tickets must include a CRUD/list/action
matrix with evidence, typed models, identities, success/error shapes and command
composition mappings. Supporting operations needed for these compositions are in
scope even when no same-named PowerShell command exists. Missing command names do
not prove an API operation is unavailable. Conversely, no route, mutation or
unsupported stub is invented merely to fill every CRUD cell: document the actual
read-only, singleton, action-only or other evidenced service constraint.

The coverage ledger remains command/parameter-set based and links every constituent
SDK symbol and test. Contract tests cover both individual operations and composed
workflows, including partial failure, explicit false/empty values, list traversal,
cancellation and mutation replay safety. PowerShell method naming and pipeline
shape do not determine the Go service API.

Illustrative shapes below establish signatures, not unverified endpoint routes or complete resource schemas:

```go
// Root constructors.
func NewClient(authConfig *AuthConfig, options ...ClientOption) (*Client, error)
func NewClientFromEnv(options ...ClientOption) (*Client, error)

// A Graph resource service.
func (s *Teams) GetByID(ctx context.Context, id string) (*ResourceTeam, *resty.Response, error)
func (s *Teams) Create(ctx context.Context, request *RequestTeam) (*Operation, *resty.Response, error)
func (s *Teams) UpdateByID(ctx context.Context, id string, request *RequestTeamUpdate) (*resty.Response, error)
func (s *Teams) DeleteByID(ctx context.Context, id string) (*resty.Response, error)
func (s *Teams) List(ctx context.Context, request *RequestTeamList) (*ResourceTeamPage, *resty.Response, error)
func (s *Teams) Archive(ctx context.Context, id string, request *RequestTeamArchive) (*Operation, *resty.Response, error)

// A Teams administration resource service.
func (s *MessagingPolicies) GetByIdentity(ctx context.Context, identity string) (*ResourcePolicy, *resty.Response, error)
func (s *MessagingPolicies) UpdateByIdentity(ctx context.Context, identity string, request *RequestPolicy) (*ResourcePolicy, *resty.Response, error)
```

Actual research contracts decide whether creation is synchronous or returns an operation and whether a mutation returns a resource or only response metadata. Do not force an invented response into every method. Deletes with no resource return `(*resty.Response, error)`. Lists expose typed page retrieval and explicit aggregate traversal, preserving partial failures. All remote calls take `context.Context` first; local validation returns errors rather than panicking.

Root type aliases can expose configuration conveniences without moving ownership into the root. Request and response models are distinct. Where PATCH semantics distinguish missing from null, use a field representation with explicit presence and null state; a lone `*T` cannot represent all three states. False, zero and empty collections must survive when explicitly supplied. Unknown response fields should not break ordinary decoding, while contract observation records them as possible drift.

## Three representative call paths

1. **Graph team read:** `Client.GraphAPI.Teams.GetByID` validates identity, builds a Graph backend request, then shared transport selects the Graph audience, sends HTTP and decodes the typed result/error. It never requests a Teams token simply because both services share a root client.
2. **Teams policy update:** `Client.TeamsAdminAPI.MessagingPolicies.UpdateByIdentity` encodes the evidenced policy contract. The Teams backend adapter resolves its approved route/envelope and the transport obtains the Teams audience token. Policy grant/assignment remains a separate operation where the baseline exposes it.
3. **Local attendant builder:** a resource-package constructor such as `NewPrompt(...)` returns a validated typed value for use in a later request. It does not accept a client, require authentication, allocate a background worker or send HTTP. Runtime SDK consumers never invoke PowerShell.

## Consequences and verification

Consumers get the requested Jamf-style typed services and Resty response metadata. Resty is therefore a public compatibility dependency; changing its major version requires API review. Service tests replace the injected request boundary, and transport tests use controlled HTTP servers.

FND-01/FND-02 must compile the module and representative examples and verify the import graph. Resource tickets must show command-to-method/model mappings, meaningful validation and error cases, and no HTTP call for local constructors. Until those implementations exist, these snippets are design contracts, not runnable examples or claims of Go test coverage.
