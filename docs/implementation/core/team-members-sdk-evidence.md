# CORE-26 SDK progress

Status: in progress. The first slice adds native conversation-membership read
models and separate create/role-patch inputs. No resource methods or root-client
field are exposed yet. Native CRUD/list, group member/owner reference primitives,
PowerShell compositions, service evidence reports and journalled lab acceptance
remain required. This does not close
[CORE-26 / #182](https://github.com/deploymenttheory/go-sdk-msteams/issues/182)
or promote any command coverage.

The models implement the accepted
[membership resource boundary](membership-resource-contract.md): opaque native
membership IDs and directory user IDs are distinct Go types. Read models retain
unknown subtype/role/property information, nullable AAD fields and another
tenant's identity without decoding or substituting the membership ID. Chat-only
history is a read field and does not enter a team role patch.

Role-patch serialization distinguishes omission, explicit null, explicit empty
array and owner promotion. Only the empty array represents demotion. Future
resource methods must reject null/unsupported role writes, keep server-assigned
guest roles immutable, preserve native membership IDs in escaped routes and
return the updated member on HTTP 200. A create input supplies a directory user
ID; the service must construct its binding from the selected Graph cloud.

These are local model/serialization tests, with no HTTP, authentication or tenant
operations. They do not establish native service readiness or parity with the
backing-group behavior of Add/Get/Remove-TeamUser.

Primary sources verified on 2026-11-20:
[AAD member](https://learn.microsoft.com/en-us/graph/api/resources/aaduserconversationmember?view=graph-rest-1.0),
[create](https://learn.microsoft.com/en-us/graph/api/team-post-members?view=graph-rest-1.0),
[update team member](https://learn.microsoft.com/en-us/graph/api/team-update-members?view=graph-rest-1.0),
[list and opaque identities](https://learn.microsoft.com/en-us/graph/api/team-list-members?view=graph-rest-1.0).
