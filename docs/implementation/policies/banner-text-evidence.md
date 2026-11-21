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

GetAll returns a bare CustomBannerText collection. Full Get filtering/sorting and
public binder semantics require their own comparisons; provider snapshots alone
do not prove all command processing.

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

## Remaining implementation and qualification

Implement full native CRUD/list, both selector forms, source-compatible default
ID generation and explicit Set read/modify/PATCH composition with partial results.
Preserve source body Identity/Id distinctions, omission/null/empty/Force, owned raw
snapshots, exact parameter-set mappings and appropriate local model validation.
Add public binding observations, filter/identity preparation coverage, successful
and failure wire comparisons, documentation/examples and gated read-only lab
support. Server rules, authorization, updates/read-back/restoration and feature
behavior remain unqualified. Issue #114 and full command/tenant qualification
remain open. Microsoft binaries and decompiled source stay private.
