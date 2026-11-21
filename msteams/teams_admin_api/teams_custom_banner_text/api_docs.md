# Teams custom banner text

`Client.TeamsAdminAPI.TeamsCustomBannerTexts` exposes native Go CRUD/list.
PowerShell is used only for development captures and separately enabled lab
comparisons. The compiled 8.1.0 resource is CustomBannerText.

| Go method | Baseline command / parameter set |
|---|---|
| GetByIdentity, List | Get-CsTeamsCustomBannerText / Identity |
| GetByID | Get / ParentAndRelativeKey |
| ListFiltered | Get / Filter |
| CreateByIdentity | New / Identity |
| CreateByID, Create | New / ParentAndRelativeKey; bound or generated Id |
| UpdateByIdentity, PatchByIdentity | Set / Identity; composition or sparse operation |
| UpdateByID, PatchByID | Set / ParentAndRelativeKey; composition or sparse operation |
| DeleteByIdentity | Remove / Identity |
| DeleteByID | Remove / ParentAndRelativeKey |

Relative-ID methods use Global/Id. Full identities preserve first-slash scope/name
escaping with the remainder as one opaque component. Selected IDs are opaque;
the GUID constraint applies to model Id values rather than route selectors.
List consumes a bare array, puts exact Global first, and has no invented paging.
ListFiltered matches local Identity/Tag wildcards, preserving literal whitespace.
Use List for an omitted filter; an explicitly empty pattern rejects locally.

Create requires nonempty Text, as the public New binder does. Description uses
policy.Set/Null to preserve omission, empty and explicit null. Set permits empty
Text but its model setter rejects explicit null. The source model does not apply
the generated text/description length constants. Response models preserve nullable
and omitted properties, owned original JSON and unknown fields. An earlier bad
duplicate Id or explicit null Text fails decoding. Requests send known bound
settings without forwarding received future fields.

CreateByIdentity sends Identity and, when present, its second split component as
Id metadata. CreateByID sends computed Identity and null Id from the new source
DTO. Create generates one UUID, sends Identity/Text without Id, and returns the
generated ID/Identity even if submission fails. It adds no implicit Get/read-back.
All three POST the collection with command metadata. Bound Force serializes as
the compiled string True or False.

UpdateByIdentity/ID explicitly read with Set metadata and then PATCH once. The
body retains the read model's Identity or Id, including null, while the route
keeps the caller's selector. Relative-ID updates omit Identity. UpdateResult
retains Before, ReadResponse, PatchInvoked and HTTPAccepted on failure. Direct
PatchByID requires an explicit read-model BodyID value/null; use UpdateByID for
automatic preparation. Direct PatchByIdentity defaults BodyIdentity to its route,
with an explicit override available. Mixed body selector fields reject locally.
DeleteByID sends Id; DeleteByIdentity sends Identity.

Sparse mutations never replay, read back, wait or roll back implicitly. A 204
returns no fabricated Banner. An accepted malformed reply returns ErrDecode with
its HTTP metadata retained; acceptance does not prove persistence or displayed
behavior. Ordinary model/result formatting is redacted. Raw snapshots/responses
are caller-owned sensitive data.

Calls use the selected Teams audience/origin and compiled policy headers with a
fresh correlation UUID; no Graph lookup or modern forest header is added.
Synthetic application/delegated tests prove routing, not tenant authorization.
Successful tenant CRUD, server/cloud availability, cleanup and effective display
remain open under #114. See [unit tests](service_test.go)
and the [lab acceptance record](../../../docs/implementation/policies/banner-text-lab-acceptance.md).
