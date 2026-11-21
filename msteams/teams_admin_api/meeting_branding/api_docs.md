# Meeting branding policies

`Client.TeamsAdminAPI.MeetingBrandingPolicies` exposes native CRUD/list and explicit
assignment operations for the compiled `TeamsMeetingBrandingPolicy` ConfigAPI resource.

| SDK operation | MicrosoftTeams 8.1.0 mapping |
|---|---|
| GetByIdentity / List | Get / Identity, including omitted Identity for List |
| ListFiltered | Get / Filter, with local Identity/Tag wildcard matching |
| Create | New / Identity, one sparse POST |
| CreateWithGlobalComparison | Explicit optional Global comparison plus Create |
| PatchByIdentity | Set / Identity, one sparse native PATCH |
| UpdateByIdentity | Set / Identity, Get then bound local edits then PATCH |
| DeleteByIdentity | Remove / Identity, one DELETE with bound Identity body |
| AssignUser / AssignTenantDefault / AssignGroup | GrantToUser / GrantToTenant / GrantToGroup |
| Assign | Exactly one explicit scope; the otherwise unhandled Identity set rejects locally |
| NewBackgroundImage / NewBrandingTheme / NewAssuranceSlate | Three local New builders / Identity |

`Fields` uses policy.Set/Null for omitted/value/null distinctions, including false
and empty strings. Nil list pointers omit a setting. Assets[T]() sends an owned
empty array; populated arrays retain duplicates, null elements and order. Selected
nested values contain all known public properties, including nullable defaults and
inherited Identity. Unknown read snapshot fields are retained on reads and excluded
from writes. URI-valued assets remain local values until a policy write persists them.

ReplaceAssets and ChangeAssets build owned local edits. Empty edits clear lists;
nil edit pointers omit them. To represent a bound null modifier's clearing effect,
use ReplaceAssets[T](). The pinned Core helper invokes a throwing comparator when
removing from a populated policy asset list, including a missing match. The SDK
returns ErrCannotRemoveElements after the read and retains Before without PATCH.
Removal from an empty list needs no comparison. Theme reaction edits compare ID,
image URL, sprite URL, type and branded display name; they remove one match per item,
then append additions. They emit full arrays, with no atomic Add/Remove claim.

Builders perform no token acquisition, URI dereference, image upload or separate
asset persistence. Image/slate strings follow the source's simple ASCII-case HTTPS
regex, including its final-newline behavior. There is no inferred URI length, image
size/dimensions, order range, color or alignment enum validation. Theme URI-like
fields remain open strings. NewReaction is an additional local helper for typed
theme inputs. Snapshots preserve received absence/null/future data and original
Int64 order; explicit nulls on non-null model setters and invalid HTTPS model fields
fail decoding. CLR getter defaults are projected only when forming nested writes.

UpdateResult/CreateResult retain preparation and HTTP acceptance on later failure.
No write retries, automatic read-back, polling, rollback, ETag guarantee or propagation
claim is implicit. User/group/tenant assignment payloads preserve null/empty policy
names, the tenant Force Boolean and Int32 group rank. A group response is an optional
policy snapshot/raw data; an opaque OperationId is not a fabricated async job.

The selected Teams origin/token and shared compiled-policy headers apply, with fresh
UUID correlations and escaped route segments. Generated forest discovery and Graph
lookups are absent. CRUD Force uses source True/False strings. Empty/null successful
bodies remain nil. Ordinary error/snapshot formatting is redacted and raw responses
remain available to callers. Context, destination and audience checks remain shared.

Successful persistence, assignment, Premium feature effects, RBAC, cloud deployment
and full binding/coercion parity remain separately qualified. See
[evidence](../../../docs/implementation/policies/meeting-branding-sdk-progress.md) and
the [lab plan](../../../docs/implementation/policies/meeting-branding-lab-acceptance.md).
