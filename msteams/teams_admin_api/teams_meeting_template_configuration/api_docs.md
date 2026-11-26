# Meeting-template configuration

Use `sdk.TeamsAdminAPI.TeamsMeetingTemplateConfigurations` for the public
MicrosoftTeams 8.1.0 `Get-CsTeamsMeetingTemplateConfiguration` surface.

| Method | Behavior |
|---|---|
| `GetByIdentity(ctx, identity)` | One direct PolicyRP GET with an explicit full identity |
| `List(ctx)` | Bare collection with exact `Global` first |
| `ListFiltered(ctx, pattern)` | Local Identity/Tag wildcard selection after a collection GET |

The baseline exports Get with Identity and Filter parameter sets. A missing
PowerShell Identity maps to explicit `List`; the native SDK requires a caller
context. Empty identities/filters and malformed filters fail before authentication.
Whitespace stays literal. Scoped identities split at the first slash and escape
the remaining name as one segment.

`Configuration`, `Template` and `Option` expose every pinned model property.
Strings and Booleans retain nullability; the source getters materialize absent or
null collections as empty lists. Explicit null nested Name setters are rejected
like the pinned models. No enum, GUID or description/name length restrictions are
inferred. `RawFields` and an owned `RawJSON` preserve unknown response fields.
Marshaling a received model returns its original snapshot.

```go
configuration, response, err := sdk.TeamsAdminAPI.TeamsMeetingTemplateConfigurations.GetByIdentity(ctx, "Global")
```

The lab SPN uses Teams Administrator. Both certificate and client-secret runs
read Global and a filtered collection and compare Description and the complete
nested template/option getter projection with PowerShell 8.1.0. The current lab
returns an empty template collection. Nonempty nested collection behavior has
source-backed typed models; it has not been demonstrated with a live custom
meeting template. No tenant configuration changes are performed by these tests.

Unit tests use the actual sanitized `testdata/get_global.json` and `list.json`
responses. They cover read/list/filter, snapshot ownership, invalid inputs,
context/cancellation, backend isolation, injected authorization failures and
malformed replies. The example compiles without connecting to a tenant.

See [qualification details](../../../docs/implementation/policies/meeting-template-policies-lab-acceptance.md).

Received JSON uses the own pinned DTO scalar profile: numeric/Boolean string
values and nullable Boolean conversions are accepted where the model does,
while explicit null required names and invalid objects still fail. Numeric
string projections preserve their original numeric token text. Getter-only
list arrays accumulate duplicate/case aliases, null aliases retain previous
items, and whole primitive assignments are ignored. These read rules do not
relax typed write construction or the strict canonical PowerShell lab oracle.
Raw snapshots retain the received tokens independently of typed projection.
