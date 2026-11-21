# Teams education configuration evidence (POL-43 / #116)

Actual MicrosoftTeams 8.1.0 public exports are compiled Core Get/Set commands.
Get-CsTeamsEducationConfiguration has Filter/Identity sets; Set has Identity,
for three exact command/set pairs. No modern CmdletConfig mapping exists. The
captured provider normalizes the resource to TeamsEducationConfiguration.

## Source capture boundaries

capture-education-configuration-provider.ps1 pins Core/PolicyRp/HostContract hashes,
checks actual public export types, and invokes direct HttpRequestHelper with
synthetic InvocationInfo and the shared credential-refusing host/memory sender.
Client/version/host state is restored. The fixture contains 11 provider request
observations and 16 actual constructor/setter/Newtonsoft model observations.
Each provider case makes one memory send and zero token requests. No original
command constructor/lifecycle, binder, authorization, discovery or network runs.

capture-education-configuration-binding.ps1 separately exercises inherited compiled
properties/attributes/constructors through an in-process runspace and synthetic
lifecycle overrides. It captures 17 binding observations (14 successes, three
expected validation failures) and five actual protected FilterPolicies selections.
The Set subclass invokes actual protected AreParametersSpecified. Every binding
and filter observation has zero sends and token requests. This replaces the
original lifecycle and does not prove command execution or tenant authorization.

## Observed behavior

All three settings are nullable opaque strings: ParentGuardianPreferredContactMethod,
EduGenerativeAIEnhancements and UpdateParentInformation. Actual setters accept
null, empty, future and long text without enum/length/cross-field validation.
Canonical case-insensitive duplicates retain the last valid value, including
null. JSON arrays assigned to string properties fail. Arbitrary PowerShell object
coercion is distinct from a typed native string/JSON boundary.

The compiled public binder preserves null/empty optional strings and converts a
numeric string argument. Set's actual predicate rejects Identity-only and
Force-only calls, while a bound null setting counts as supplied. Empty bound
Identity/Filter rejects before processing. No create/delete/grant public operation
is inferred beyond the assigned baseline exports.

Reads use the compiled configuration identity route. Scoped identities split on
the first slash, with the remainder escaped as one opaque name. List consumes a
bare collection and applies Identity/Tag wildcard filtering locally, including
literal whitespace. Original Get emits exact Global identities first. PATCH
includes the read model's Identity plus only bound settings; bound Force is the
string True/False. MPACmdlet, command and GUID correlation headers apply, without
a modern forest target. A collection envelope fails separately from resource
denial; empty/null reads and accepted empty PATCH responses retain their actual
source result shapes.

## Native implementation and qualification remaining

TeamsEducationConfigurations now exposes native read/list/filter and sparse PATCH
plus explicit Set read/PATCH composition. Presence-aware fields, owned snapshots,
input/context and no-replay guards, an example and API docs accompany source
provider/binding/model/filter comparisons and all three command/set mappings.
Mutations retain partial read preparation and HTTP acceptance without fabricated
read-back or replay. Ordinary model/settings/result formatting is redacted; raw
snapshots/responses remain caller-owned sensitive configuration. Native strings
and JSON remain typed; the source binder's numeric coercion is tested through
its captured typed bound value rather than added to the Go API.

The read-only lab/source harness gates credentials, bounds execution/output,
requires complete typed getter projections and binds the optional pinned source
oracle to the selected tenant. Ledger rows remain in-progress/metadata-only;
offline artifacts do not establish successful tenant coverage or complete public
command lifecycle parity.

Actual opt-in Global lab reads returned HTTP 403 in client-secret and certificate
modes on 2026-11-21. No tenant writes, roles or consent changes ran. Successful
authorization, server validation, setting updates/read-back/restoration, cloud
coverage and effective parent contact, generative AI and parent information
behavior remain unqualified. See the
[lab acceptance record](education-configuration-lab-acceptance.md). Issue #116
stays open. Microsoft binaries and decompiled source remain private.
