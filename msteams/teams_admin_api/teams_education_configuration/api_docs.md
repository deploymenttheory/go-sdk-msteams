# Teams education configuration

`Client.TeamsAdminAPI.TeamsEducationConfigurations` exposes native read/list/update
for the public MicrosoftTeams 8.1.0 Get/Set surface. PowerShell is used only by
development captures and separately enabled lab comparisons.

| Go method | Baseline command / parameter set |
|---|---|
| GetByIdentity, List | Get-CsTeamsEducationConfiguration / Identity |
| ListFiltered | Get / Filter |
| PatchByIdentity | Set / Identity; one sparse native PATCH |
| UpdateByIdentity | Set / Identity; explicit preparation read and sparse PATCH |

All three settings are nullable opaque strings: ParentGuardianPreferredContactMethod,
EduGenerativeAIEnhancements and UpdateParentInformation. Use policy.Set/Null to
distinguish omission, empty and explicit null. No enum, parser, length or cross-field
rule is inferred from a property name. Set requires at least one bound setting;
Identity/Force alone rejects locally. Bound Force is the compiled string True/False.

Identities preserve first-slash scope/name escaping with the remainder as one
opaque component. List consumes a bare collection and places exact Global first.
ListFiltered uses local Identity/Tag wildcard matching, with literal whitespace.
Use List for omission; an explicitly empty filter rejects locally. No invented
continuation, modern forest header or Graph lookup is added.

Direct PatchByIdentity sends known bound settings and defaults body Identity to
the selected route, with an explicit nullable BodyIdentity override available.
UpdateByIdentity reads using Set metadata, retains that model's nullable Identity
in the body, and PATCHes the caller's route once. UpdateResult retains Before,
ReadResponse, PatchInvoked and HTTPAccepted after partial failure. A 204 has no
fabricated Configuration; accepted malformed responses return ErrDecode with
HTTP metadata retained. Mutations do not replay, read back, wait or roll back
implicitly. Acceptance does not establish persisted or effective tenant state.

Models retain nullable getters, owned original JSON and unknown fields, with
last valid case-insensitive string values. Native strings/JSON remain typed;
arbitrary shell input coercion is a separate captured behavior. Requests do not
forward received future fields. Ordinary model/settings/result formatting is
redacted; raw snapshots/responses remain caller-owned sensitive configuration.

Calls use the selected Teams audience/origin and compiled policy headers with a
fresh UUID correlation. Synthetic application/delegated tests establish routing,
not successful tenant authorization. Authorized updates/read-back/restoration,
cloud/server availability and effective parent-contact, generative AI and parent
information behavior remain open under #116. See
[source evidence](../../../docs/implementation/policies/education-configuration-evidence.md)
and the [lab acceptance record](../../../docs/implementation/policies/education-configuration-lab-acceptance.md).
