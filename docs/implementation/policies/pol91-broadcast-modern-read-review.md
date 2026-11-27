# Meeting broadcast: modern received-JSON review

This is a source investigation for the two still-unqualified POL91 commands,
Get/Set-CsTeamsMeetingBroadcastConfiguration. It changes no coverage decision,
Go behavior, live evidence or release claim. The modern received-value pipeline
is materially different from the compiled PolicyRp DTO pipeline.

## Pinned source and executed scope

The [portable observer](../../../scripts/discovery/observe-pol91-broadcast-modern-read.ps1)
checks `custom/Merged_custom_PsExt.ps1` SHA256
`03dbc3a85e12c98dd1807d604782c6a90d0656569541cd9540738825a7fcb255`.
It AST-extracts the actual five pure helpers Custom-ToString,
Convert-PsCustomObjectToPsObject, Get-FormatsForConfig,
Set-FormatOnConfigObject and Sort-GlobalFirst. It executes the received-JSON,
format, optional exposed-property selection and sorting stages used by
Get-CsConfigurationModern, on inert authored inputs. The management helper,
telemetry constructor, private generated HTTP command and flight selection are
not executed. No module import, authentication or HTTP is needed.

The [242 observed outcomes](evidence/pol91-broadcast-modern-read.json) include
each of the ten exposed fields with eleven JSON token shapes, root shapes,
duplicate keys, unknown nested data and Global-first ordering, both with and
without exposed-property selection. There are 234 complete outcomes and eight
exceptions. The record includes the PowerShell runtime version, extracted source
line ranges, producer hash and failure stage. Local execution exited zero;
producer hash, row count and helper count were checked independently when copying
the result into the repository. These are pure local observations, not captures.

An independent reader compared the observer with the actual helper pipeline and
reran the repository producer (session 81919, exit zero). Its parsed output was
identical to the committed evidence: all 242 rows, five helper identities/ranges
and producer hash. The review approved this bounded investigation only.

## Findings that change the next implementation decision

Get-CsConfigurationModern lines 366–368 converts each generated output string
using ConvertFrom-Json and the PSSerializer helper. It does not deserialize a
TeamsMeetingBroadcastConfiguration CLR DTO. Thus compiled scalar coercion rules
must not be reused merely because the configuration endpoint has a similar name.

The observed default result preserves the original JSON value kinds inside an
object. For example, SupportURL:true stays Boolean; SupportURL:{} stays an
object; AllowSdnProviderForBroadcastMeeting:"text" and
SdnAttendeeFallbackCount:"text" stay strings. Unknown nested Future data survives.
With exposure enabled, Select-Object projects exactly the ten named properties:
missing properties become null and unknown properties disappear from that host
projection. Existing properties retain their value kinds. The raw HTTP body is a
separate object from that PowerShell presentation result.

Exact repeated property names are last-wins; case-distinct colliding names raise
InvalidOperationException at the JSON/PSSerializer stage. The existing Go decoder
accepts case-insensitive aliases through encoding/json, while rejecting the
nonmatching known-field token kinds above before callers receive RawJSON. That
is an unresolved behavior difference, not a source-proven malformed response.
Its resolution needs an explicit dynamic/raw result and typed-view design;
adding the compiled ReadString/ReadBool adapters would introduce conversions
that this source does not perform.

Empty-object input survives. In these isolated stages, null, [] and [null] reach
a ParameterBindingValidationException at the mandatory format argument. Primitive
roots survive the default presentation and become ten-null-property projections
when exposed. Those observations do not prove that the preceding generated
response parser emits any of these roots. Inspect that parser before changing
service-level accepted root shapes or lab-oracle expectations.

Sorting places both "global" and "Global" before other identities, preserving
their relative order and the order of the remainder. This confirms the current
case-insensitive stable partition for ordinary string identities. It does not
settle PowerShell comparison/filter semantics for dynamic nonstring identities.

## Remaining contract work

The modern Get helper forwards ConfigName only for a non-whitespace Identity,
applies non-whitespace Filter locally with Identity/Tag wildcard matching, and
adds the exposure query through its pipeline prepend. The Set helper inserts
Identity="Global" only when the property bag lacks the key, then forwards the
bag to private Set-CsConfiguration. These source facts still need to be closed
against this command's actual public binding, generated request/response parser,
headers, status handling and diagnostic/control flow before qualification.

Public default dispatch remains session-flight dependent. The concrete legacy
handoff and unavailable server-side contract are documented in the
[conditional source boundary](pol04-primary-source-audit.md#conditional-dispatch-and-the-precise-legacy-boundary).
Modern observations neither exclude that branch nor establish its models or
authorization. Existing tenant read/update/restoration evidence retains its own
scope; this observer adds no tenant call or live qualification.
