# Conditional conversation roles

`TeamsAdminAPI.ConversationRoles` uses the compiled `ConversationRole` resource
at `/Skype.Policy/configurations/ConversationRole`. It adds no exports, roles,
consents or licenses. Constructors and local projections do not authenticate.
These commands remain conditional and unqualified for live operation.

## Reads and local values

`GetByIdentity`, `GetByID`, `List` and `ListFiltered` map the Identity,
ParentAndRelativeKey and Filter declarations. `GetByID(id)` targets `Global/id`;
it does not treat a response's `Id` property as an interchangeable full identity.
Omitted Filter uses List; explicitly empty Filter fails locally as declared,
while whitespace is a literal wildcard pattern.

Role/permission DTOs retain complete owned raw snapshots and nullable typed
flags: eighteen chat and thirteen calling properties. `NewChatPermissions`,
`NewCallingPermissions` and `NewPermissions` construct local data only. The last
requires both nested objects, matching its observed CLR constructor. These Go
constructors do not claim public export or complete binding parity for the
three separately discovered compiled PowerShell builder types.

Dates use the existing SDK CLR convention: explicit unspecified/UTC/local kind,
years 1–9999 and 100-nanosecond precision. Omitted response dates project to the
unspecified CLR minimum. Explicit null dates/Permissions and explicit null
nested Chat/Calling objects fail model decoding; null nullable flags are valid.
Unknown properties and omitted/null/empty distinctions remain in RawFields and
RawJSON. Model `Id` follows the pinned GUID expression, including its terminal-LF
scalar behavior; URL/identity validation remains a separate transport boundary.

## Inspectable compiled writes

The pinned serializer drops nested Chat/Calling properties even when its actual
local builders supplied populated flags. Full effective permission writes remain
an unresolved management-wire contract. This SDK does not silently replace that
behavior with an assumed native JSON payload or reject accepted flag inputs.

1. `ProjectCreateByIdentity` or `ProjectCreateByID` accepts `CreateParameters` and
   returns an opaque `CreateRequest`. Inspect `request.Projection()` first.
2. `Create(ctx, request)` sends the frozen observed request. Full Identity creation
   adds derived Id metadata; relative Id creation retains the source's `Id:null`
   representation. Source acceptance and actual persistence of either variant
   still need live proof.
3. `PrepareUpdateByIdentity`/`PrepareUpdateByID` performs only the Set-metadata GET
   and local projection. Inspect its Projection, Before and ReadResponse.
4. `UpdateByIdentity`/`UpdateByID` applies that frozen PATCH without another GET or
   projection. The service/client, exact selector and parameter set must match
   preparation. Changing exposed snapshots, projection bytes or original input
   cannot retarget or modify the prepared request.

`WriteProjection.Body` is an owned inspectable JSON copy. OmittedPaths describes
source projection loss, and ChangedPaths exposes parameter-set changes such as
relative creation's null Id. String/GoString redact requests, projections, DTOs
and results. Explicitly inspecting raw bodies can expose resource data; they are
not automatically logged. There is no additional approval flag or consent flow.

Bound Name/IsAssignable null and false remain distinct from omission. Dates are
non-null CLR values when selected. Selecting no mutable settings, including
Force alone, produces an explicit NoOp without credentials or HTTP. Selected
null Permissions fails local CLR projection after preparation, retaining the
read receipt. A bound Force false uses the observed `"False"` string.

`DeleteByIdentity` and `DeleteByID` retain the original bound-parameter body:
Identity or Id respectively, plus an optional Force string. RemoveAsync does not
apply creation's required-metadata body merge. These are exact target operations,
not bulk cleanup or assignment APIs.

## Results and limits

Preparation denial returns its response and an incomplete request that cannot be
applied. Apply results retain Before, ReadResponse, Projection, PatchInvoked and
HTTPAccepted independently of decoding or lost-response/cancellation errors.
Writes use ReplayNever. A 2xx receipt does not prove stored/effective permissions;
ambiguous writes need independent target reconciliation. Preparation plus apply
is not atomic and has no invented ETag guarantee.

Request vectors come from the pinned local pipeline; reply vectors are authored
model cases. No new role or Global setting mutation is enabled in live CI. See
[the full source notes](../../../docs/implementation/ext/conversation-role-contract-notes.md)
for exact hashes, the twelve declared management sets and three additional
builder declarations. No command is marked contract-complete from these tests.
