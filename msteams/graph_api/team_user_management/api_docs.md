# Team-user command compositions

`sdk.GraphAPI.TeamUserManagement` implements the backing-group workflows used by
the pinned 8.1.0 Add/Get/Remove-TeamUser commands. `TeamMembers` remains the native
conversation-membership CRUD/list service; its membership IDs are not directory
user IDs.

| Method | Behavior |
| --- | --- |
| `Add` | Resolve the explicit user, traverse members, add missing membership, then ensure ownership for role Owner. |
| `List` | Traverse owners/members as required by role; project owners, ordinary members, then guests. |
| `Remove` | Resolve the explicit user, traverse owners, guard the last returned owner, remove ownership, then membership unless role Owner was selected. |

Group IDs must be directory UUIDs. An explicit user UUID needs no resolver read;
UPNs use direct lookup with direct-404-only canonical mail fallback. Empty user
selectors do not infer the current user. Role comparison ignores case.
Add/Remove accept omitted/Member or Owner; List also accepts Guest.

Each relationship read must finish within its own `Traversal` bounds before
establishing absence. `ResolveMaxPages` independently bounds mail resolution.
Denied, malformed or incomplete reads stop the next write. A successful empty
list permits creation, explicitly correcting the module's status-only no-op.
Duplicate-looking mutation errors remain errors.

`ChangeResult` retains the resolved ID, resolution source, ordered stages and
latest received response on failure. Completed writes remain completed. An
unconfirmed write might have taken effect remotely. The composition does not
replay, roll back or wait for synchronization. Cancellation and client close
apply to the entire workflow.

Remove with role Owner only demotes; membership remains. `ErrLastOwner` is a
local precheck error retaining the actual read response, not a fabricated HTTP
412 or concurrency guarantee. Graph still enforces its restrictions. Reference
deletes always use the fixed `/$ref` suffix and cannot delete a directory user.

List deduplicates IDs and gives owners precedence. Role Owner skips member
reads; Guest skips owner reads. Nullable directory fields and full native
objects remain available. The pinned projection categorizes missing `userType`
as member without claiming a known directory type. `Complete` requires every
selected read to succeed; partial projections survive failures.

Permissions must cover all selected endpoint operations. Limited-data responses,
guest restrictions and v1.0 service-principal omissions still apply: complete
cursor traversal is not proof of visibility of every object type. See the
[native reference notes](../group_relationship/api_docs.md) and [resolver notes](../directory_user/api_docs.md).
No role or permission is granted. Offline reports execute the root SDK with
synthetic responses; live qualification and client convergence remain untested.
