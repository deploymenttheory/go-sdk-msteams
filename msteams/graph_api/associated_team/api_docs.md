# Associated teams

`sdk.GraphAPI.AssociatedTeams` exposes a membership-derived beta Graph view,
including teams associated through shared channels. `ListForUser`/`NewPager`
require an explicit directory UUID. `ListMine`/`NewMinePager` explicitly select
the delegated `/me/teamwork/associatedTeams` route. Application sessions fail
locally before acquiring credentials for a current-user request. Blank explicit
IDs never silently select `/me`.

Lists are complete bounded traversals of HTTP-200 pages; failures retain partial
items and last response with an error. Snapshots preserve host tenant,
nullable display name and raw future fields. The view does not provide team
lifecycle mutations; those belong to `GraphAPI.Teams`/`TeamManagement`.

Current associated-team documentation lists Team.ReadBasic.All and restricts
delegated access to the signed-in user; application access requires an explicit
user. A pinned explicit-user branch is not a promise that delegated requests can
read another user. Cloud availability and authorization require separate lab
qualification. See [unit tests](service_test.go).

`SharedChannelManagement.ListAssociatedTeams` adds the pinned UUID/UPN/empty-user
selection, using the shared directory resolver. The native service stays ID-based.
