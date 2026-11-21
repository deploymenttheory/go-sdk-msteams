# Teams custom banner text evidence (POL-41 / #114)

The MicrosoftTeams 8.1.0 baseline exports compiled Core Get/New/Set/Remove
commands with no modern CmdletConfig mappings. Get has Filter, Identity and
ParentAndRelativeKey sets; each mutation has Identity/ParentAndRelativeKey, for
nine exact command/set pairs. Actual provider normalization selects the
CustomBannerText resource, not TeamsCustomBannerConfiguration or a guessed
modern banner endpoint.

## Capture seam and observations

`capture-banner-text-provider.ps1` verifies the baseline and Core, PolicyRp and
HostContract hashes. The owned BannerTextProviderCapture adapter injects the
memory HTTP client before actual command constructors; the source factory only
initializes a null client, and identity checks ensure it retains the memory
client. It invokes actual pure PrepareForScopeCheck, model constructors/setters
and direct HttpRequestHelper methods with synthetic InvocationInfo. The host
refuses credentials and the sender accepts only the fixture origin/token.
Static client/version/host state is restored.

The promoted banner-text-provider.json contains 24 provider/scope observations
and 17 model observations. No Begin/Process/End, public binder, scope authorization,
token acquisition, discovery or network is executed. Every case has zero token
requests. Missing create Text and explicit null update Text fail locally with
zero dispatches; all other provider cases make one memory dispatch. Empty/null
reads succeed; a collection envelope and resource denial fail separately.
Constructor default GUID shape and actual default-ID POST are observed. Generated
UUIDs are synthetic fixture values, not tenant identifiers.

## Selector preparation and bodies

Actual preparation sets relative-ID requests to Global/Id and adds an unbound
Identity metadata entry. A full Identity containing slash derives Id from its
second split component and adds unbound Id metadata. For an extra slash, metadata
uses only that second component while route escaping preserves the remainder as
one encoded opaque component. The native API must preserve this distinction.

| Operation | Full Identity selector | Relative-ID selector |
|---|---|---|
| Read | GET the selected full identity | GET Global/Id |
| Create | Body Identity plus derived Id and bound settings | Body Identity computed as Global/Id, bound Id projects null from the fresh model |
| Update | Body uses read model Identity and bound settings | Body uses read model Id, including explicit null, without Identity |
| Delete | Identity JSON body | Id JSON body |

Default create without an explicitly bound Id generates a fresh constructor GUID,
prepares Global/Guid and POSTs Text plus that Identity without an Id member.
Bound Force serializes as a string, including false. The resource uses compiled
MPACmdlet, command and correlation headers, with no modern forest target.

GetAll returns a bare CustomBannerText collection. A separate
capture-banner-text-binding.ps1 invokes actual inherited compiled parameter
properties/attributes through an in-process runspace with synthetic lifecycle
overrides. Its banner-text-binding.json contains 24 public binding observations
(19 successes, five expected failures) and five actual protected FilterPolicies
observations. Constructor injection and credential refusal remain in force;
every observation has zero sends and zero token requests. This seam exercises
binding and the pure filter helper, not the original command lifecycle or
authorization.

The mandatory New Text binder rejects empty/null, while optional Set Text binds
empty or null; the latter subsequently fails the model setter. New without Id
selects ParentAndRelativeKey and keeps the constructor GUID. Opaque selected Id
values bind without model GUID validation. Mixed Identity/Id selectors and empty
bound Identity fail. Filters match Identity and Tag-prefixed identities with
case-insensitive wildcard matching; whitespace is literal. Native List additionally
places exact-case Global first, as the compiled Get processing does.

## Model constraints

CustomBannerText exposes nullable Identity/Id/Description and Text. An omitted
Text leaves the default constructor value null, but an explicitly assigned null
Text throws; the text constructor also rejects null and accepts empty. Generated
Text 1/1000 and Description 1/1040 length constants are discarded, so the model
accepts 2000-character text/description. No local length rule is inferred.

Id accepts null or the exact hyphenated GUID regex, including uppercase and the
source-accepted final LF. Empty and brace-wrapped IDs fail when assigning the
model property. These model constraints must not automatically be imposed on
opaque route selectors: selected Id/Identity metadata can bypass that setter.
Later case variants/duplicates apply in Newtonsoft; a later explicit Text null
still fails. Public New Text is mandatory metadata, a distinct binder seam from
the model constructor/provider behavior.

## Native implementation and qualification

TeamsCustomBannerTexts now exposes native CRUD/list for both selector forms,
source-compatible default UUID generation, and explicit Set read/sparse-PATCH
composition. Direct writes perform no implicit read-back, rollback or retry.
Update retains its read snapshot/response, patch invocation and HTTP acceptance
on partial failure; auto-create retains its generated selector after submission
failure. Requests forward only known bound fields. Responses own their raw bytes
and future fields; sequential setter validation rejects an earlier invalid Id or
null Text even if a later duplicate would otherwise replace it. Native JSON
strings are strict; arbitrary PowerShell input-object coercion is not inferred.

Contract tests compare source routes/headers/bodies in synthetic application and
delegated modes, filter behavior, nullable read-derived selectors, accepted empty
responses, malformed replies and non-replayed mutation failures. Gated read-only
lab support compares list/read and optional bounded, tenant-bound pinned source
projections, including a genuine empty list. This does not establish successful
tenant CRUD. See the [lab acceptance record](banner-text-lab-acceptance.md).
Server rules, authorization, updates/read-back/restoration and displayed banner
behavior remain unqualified. Issue #114 and full command/tenant qualification
remain open. Microsoft binaries and decompiled source stay private.
