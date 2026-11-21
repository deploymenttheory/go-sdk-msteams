# Teams education assignments app policy

`Client.TeamsAdminAPI.TeamsEducationAssignmentsAppPolicies` exposes native Go
read/list/update operations for the public MicrosoftTeams 8.1.0 surface.
PowerShell is used only by development captures and separately enabled lab
comparisons. The baseline exports Get and Set. Unexported Core New/Remove/Grant
classes do not establish additional public operations.

| Go method | Baseline command / parameter set |
|---|---|
| GetByIdentity, List | Get-CsTeamsEducationAssignmentsAppPolicy / Identity |
| ListFiltered | Get / Filter |
| PatchByIdentity | Set / Identity; one sparse native PATCH |
| UpdateByIdentity | Set / Identity; explicit preparation read and sparse PATCH |

Identities preserve the compiled first-slash scope/name escaping, with the
remainder treated as one opaque component. List consumes the bare collection,
puts exact Global identities first and adds no invented continuation. ListFiltered
performs local Identity/Tag wildcard selection. Whitespace is literal; use List
for omission, since a bound empty filter rejects locally.

Settings exposes all five nullable source strings: ParentDigestEnabledType,
MakeCodeEnabledType, TurnItInEnabledType, TurnItInAPIURL and TurnItInAPIKey. Use
policy.Set/Null to preserve omission, empty and explicit null. Enabled-type
strings remain opaque; the source model does not impose an enum or cross-field
rule. TurnItInAPIURL allows null, otherwise applies the source's explicit ASCII
HTTPS scheme alternatives with a nonempty non-LF suffix. It accepts an opaque
host, a space suffix and one final LF, and rejects empty, HTTP and a middle LF.
This validates the source string pattern; no TurnItIn network request occurs.

Set requires at least one bound setting; Identity/Force alone rejects locally.
Force serializes as the compiled string True or False. Direct PatchByIdentity
sends known bound settings and defaults body Identity to the selected route;
BodyIdentity can explicitly override it, including null. UpdateByIdentity reads
using Set metadata and retains that read model's nullable Identity in the PATCH
body while keeping the caller's route. Invalid settings reject before the read.

UpdateResult retains Before, ReadResponse, PatchInvoked and HTTPAccepted on
partial failure. A 204 has no fabricated Policy. An accepted malformed reply
returns ErrDecode with HTTP metadata retained. Neither method implicitly reads
back, waits, rolls back or replays mutations; HTTP acceptance does not establish
persistence or effective assignment app behavior.

Response models own raw snapshots/future fields, preserve nullable values and
apply URL setter validation in input order. An earlier invalid URL fails even
if a later duplicate supplies a valid value. Typed strings/JSON remain strict;
arbitrary PowerShell object coercion is a separate shell behavior. Requests send
known settings without forwarding received unknown fields. Ordinary policy,
settings and result formatting is redacted. Raw snapshots/responses and the
TurnItIn API key remain caller-owned sensitive configuration.

Requests use the selected Teams audience/origin and compiled policy headers,
with a fresh UUID correlation and no modern forest target or Graph lookup.
Synthetic application/delegated tests establish wire behavior rather than live
authorization. Successful tenant updates/read-back/restoration, cloud/server
availability and effective assignment app behavior remain open under #115. See
[source evidence](../../../docs/implementation/policies/education-assignments-evidence.md)
and the [lab acceptance record](../../../docs/implementation/policies/education-assignments-lab-acceptance.md).
