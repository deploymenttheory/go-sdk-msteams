# Shared-channel relationships

`sdk.GraphAPI.ChannelShares` addresses the beta Graph relationship between an
explicit host team/channel and a shared-with team. `GetByID`, one-page `List`,
lazy `NewPager`, complete `ListAll`, and `DeleteByID` use
`/teams/{host}/channels/{channel}/sharedWithTeams[/{shared-with}]`.
`NewAllowedMembersPager`/`ListAllowedMembers` read that relationship's
`allowedMembers` collection. They do not substitute the direct channel roster
or recipient team's roster. A returned access-record ID is not evidence that
it can address a direct channel membership mutation.

Host/recipient selectors are explicit UUIDs. The channel ID is opaque and
escaped as one segment. Reads require HTTP 200, unshare requires 204. Unshare
deletes the selected relationship; it does not delete either team or channel.
It is sent once, including after 429/503; a lost response requires explicit
reconciliation. Reads use the shared bounded replay/paging controls. Incomplete
traversals retain prior items/last response and a nonnil error. Opaque
continuations remain subject to the configured Graph origin checks.

Snapshots retain nullable fields, explicit false, subtype and raw future fields;
native member records retain opaque membership IDs, optional directory IDs,
tenant IDs and roles. A recipient tenant is never rewritten as the host tenant.
Models describe returned data rather than write requests.

Current documentation lists ChannelMember.Read.All for delegated reads and
ChannelMember.Read.Group (resource-specific consent), or ChannelMember.Read.All,
for application reads. Unshare needs ChannelMember.ReadWrite.All, or application
ChannelMember.ReadWrite.Group with resource-specific consent. Team/channel
visibility, shared-channel prerequisites and cloud availability also apply.
See [endpoint evidence](../../../docs/implementation/core/channel-contract-evidence.md)
for exact sources and restrictions. Construction or token acquisition proves
neither permissions nor cloud/service availability.

Relationship Create/Update service contracts remain open. First-party generated
POST/PATCH builders exist, but accepted fields, exact success responses and
authorization/workflow have not been established. No guessed mutation or generic
unsupported stub is exposed. See the [mutation investigation](../../../docs/implementation/core/shared-channel-mutation-evidence.md).
The four pinned commands compose these native services through
`GraphAPI.SharedChannelManagement`. Live qualification has not run.
