# Education assignments app policy evidence (POL-42 / #115)

The actual public MicrosoftTeams 8.1.0 exports are compiled Core
Get-CsTeamsEducationAssignmentsAppPolicy and Set-CsTeamsEducationAssignmentsAppPolicy.
Get exposes Filter/Identity and Set exposes Identity, for three exact command/set
pairs. No modern CmdletConfig mapping exists. Core also contains New/Remove/Grant
classes, but those classes are not public exports in this baseline and do not
establish additional public management operations.

## Offline source captures

capture-education-assignments-provider.ps1 pins Core/PolicyRp/HostContract hashes,
verifies actual public export types, and invokes direct HttpRequestHelper with
synthetic InvocationInfo under the shared credential-refusing host and memory
sender. It restores client/version/host state. The promoted provider fixture has
14 request observations and 25 actual model constructor/setter/Newtonsoft
observations. Every provider case has zero token requests; invalid URL input
fails before dispatch. No original command constructor/lifecycle, binder, scope
authorization, credentials, discovery or network is executed in this seam.

capture-education-assignments-binding.ps1 separately exercises inherited compiled
parameter properties/attributes/constructors through an in-process runspace with
synthetic lifecycle overrides. Its fixture has 21 binding observations (18
successes and three expected failures), plus five actual protected FilterPolicies
observations. The Set subclass invokes the actual protected AreParametersSpecified.
All observations have zero sends and zero token requests. This seam replaces the
original lifecycle and does not prove command execution or authorization.

## Wire and model behavior

Actual helper normalization selects TeamsEducationAssignmentsAppPolicy. Reads use
the compiled configuration identity route; list uses the bare collection. Scoped
identity escaping splits the first slash and escapes the remaining opaque name.
PATCH includes the read model's Identity and only bound settings. Bound Force is
the string True/False. Compiled MPACmdlet, command and UUID correlation headers
apply, with no modern forest target. The source Get path reads the bare list and
filters locally; exact Global records are emitted first. An empty helper list
leaves its auxiliary Success flag false, but original Get consumes PolicyResult
without checking that flag, so the empty collection is a valid list result.

All five settings are nullable strings: ParentDigestEnabledType,
MakeCodeEnabledType, TurnItInEnabledType, TurnItInApiUrl and TurnItInApiKey. The
first three remain opaque strings; no local enum rule is inferred. The URL setter
allows null, otherwise applies the case-insensitive HTTPS-prefix regex followed
by at least one non-LF character. Source observations accept opaque hosts, a space
after the prefix and one final LF. Empty, HTTP, leading space, empty suffix and a
middle LF fail. This is the source regex rather than URI or network validation.
An earlier invalid URL fails even if a later case variant supplies a valid value.
The key accepts empty/null and long text; fixture keys are synthetic only.

The public binder preserves null/empty optional strings and converts numeric
string inputs, but does not invoke the URL model setter. Set's actual parameter
predicate rejects Identity-only and Force-only calls and accepts a bound null
setting. Get bound empty Identity/Filter and Set empty Identity reject. These
binder observations are separate from model validation and server constraints.

## Implementation and qualification still required

Implement native Get/List/Filter and sparse PATCH plus explicit Set read/PATCH
composition, typed presence-aware settings and owned snapshots, sequential URL
setter validation, failure/replay/context guards, compiling examples and API docs.
Keep any native strict JSON/string boundary explicit instead of claiming arbitrary
PowerShell object coercion. Add exact command/set ledger mappings, sanitized
contract artifacts, and gated read-only lab/source projection comparison.

No tenant read or write has run for this family. Successful authorization, server
validation, updates/read-back/restoration, cloud coverage and effective assignment
app behavior remain unqualified. Issue #115 stays open. Microsoft binaries and
decompiled source remain private.
