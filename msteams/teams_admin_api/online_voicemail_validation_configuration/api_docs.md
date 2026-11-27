# Online voicemail validation configuration

`Client.TeamsAdminAPI.OnlineVoicemailValidationConfigurations` manages the compiled
`OnlineVoicemailValidationConfiguration` ConfigAPI resource.

| Method | Baseline command / set | Behavior |
|---|---|---|
| GetByIdentity | Get / Identity | GET one explicit scoped identity |
| List | Get / omitted Identity | Bare array with exact Global identities first |
| ListFiltered | Get / Filter | List plus local Identity/Tag wildcard matching |
| PatchByIdentity | Set / Identity | One sparse native PATCH |
| UpdateByIdentity | Set / Identity | GET current, apply bound values, PATCH once |

The two settings are AudioFileValidationEnabled and AudioFileValidationURI, using
policy.Set/Null to preserve omission/value/null, false and empty distinctions.
Non-null URI strings follow the pinned model's simple ASCII-case HTTPS regex.
Null is accepted; empty and HTTP values reject locally before token acquisition.
The regex allows final LF, spaces and long values as captured. It establishes no
URL parser, reachability, length limit or audio upload. No setting URI is dereferenced.

Snapshots retain raw future data and received absence/null/false rather than client
defaults. Invalid non-null URI setters fail decoding, including case variants and
earlier duplicate keys. Raw response metadata stays available; ordinary formatting
is redacted. Native Patch uses route identity by default; Update binds the preceding
read model's body identity while keeping the requested route.

UpdateResult retains Before, ReadResponse, PatchInvoked and HTTPAccepted on partial
failure. Acceptance does not prove persisted/effective behavior. No mutation replay,
automatic read-back, wait, rollback or ETag/atomicity guarantee is implicit.
CRUD Force follows the compiled True/False string body convention.

The selected Teams origin/audience and shared compiled-policy headers apply, with
fresh UUID correlation and escaped scope/name segments. This resource uses direct
calls without generated forest discovery or a Graph identity lookup. No Create,
Delete or assignment operation is established by the two baseline commands.

App/delegated token-mode fixtures establish request routing; successful tenant/RBAC/
cloud behavior remains separate qualification. See [unit tests](service_test.go)
and the [lab plan](../../../docs/acceptance/voicemail-validation-lab-acceptance.md).

An explicitly bound empty Filter is invalid and returns ErrInput before credentials
or HTTP, matching the inherited ValidateNotNullOrEmpty binder. Use List for an
omitted filter. Literal whitespace/nonempty patterns retain local wildcard behavior.
