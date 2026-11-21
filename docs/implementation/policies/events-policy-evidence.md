# Teams events policy evidence (POL-45 / #118)

The pinned 8.1.0 public Get/New/Set/Remove/Grant exports load actual Core cmdlets.
The provider capture verifies every CommandType and implementing assembly/type.
CmdletConfig contains modern entries for this name, but those entries do not
override the observed loaded Core path. Routes must follow the compiled provider
rather than an inferred modern configuration template.

Get has Filter/Identity; Grant has GrantToUser, GrantToTenant, GrantToGroup and an
empty default Identity set; New/Set/Remove have Identity. These are nine exact
command/set pairs. All 25 settings are preserved: 23 strings and nullable
AllowEventIntegrations/UseMicrosoftECDN Booleans. No model enum/cross-field rules
are inferred; the source Description length constants are discarded.

## Executed source captures

capture-events-policy-provider.ps1 pins Core/PolicyRp/HostContract, verifies all
five actual public exports, and invokes direct HttpRequestHelper with synthetic
InvocationInfo under the shared credential-refusing host and memory sender.
Client/version/host state is restored. Its fixture contains 24 provider requests
and 83 actual model constructor/setter/Newtonsoft observations. Every provider
case makes one memory send and zero token requests. No original command
constructor/lifecycle, binder, scope authorization, discovery or network runs.

capture-events-policy-binding.ps1 uses owned subclasses inheriting original
properties/attributes/constructors, synthetic lifecycle overrides, a memory
client and credential-refusing host. It captures 121 binder/predicate observations
(116 successes, four binding failures and one expected no-scope source error)
plus five protected FilterPolicies selections. New/Set predicates are actual
protected methods. Only the parameterless Grant case invokes actual private
ExecuteGrant with synthetic token/origin; its default branch emits the source
error and returns PolicyResult=false without sending HTTP. All other grant cases
exercise binding only. Every binding/filter case has zero sends/token requests.
Original auth/scope/processing is not executed by this seam.

## Observed behavior

The compiled resource normalizes to TeamsEventsPolicy. List uses a bare array and
local Identity/Tag wildcard selection, with exact Global emitted first. Scoped
identities split on the first slash and escape the remainder as one opaque name.
CRUD uses GET/collection POST/sparse PATCH/body-bearing DELETE. Create supplies
the selected Identity; Set's body retains the read model Identity. Bound Force
serializes as a string. MPACmdlet, command and GUID correlation headers apply;
there is no modern forest target. The source's optional New Global comparison
supports ShouldProcess presentation, so native create needs no implicit read.

Native settings should distinguish omitted/null/empty/false/true. Compiled
optional string/nullable Boolean binding preserves null. New accepts an empty
settings body: its predicate can be false, but BaseTpmNew does not reject it. Set
rejects Identity-only and Force-only via its predicate. Long Description, unknown
string values and case-insensitive last values pass actual model comparisons;
wrong Boolean objects/string arrays fail. Native typed JSON need not emulate
arbitrary shell input coercion.

Grant user/tenant PATCH bodies retain additionalParameters:null. Tenant uses
capitalized PolicyName and Boolean forceSwitchPresent. Group POST always includes
required int32 Rank as priority, including zero and negative values. Nil/empty
PolicyName remain distinct, unlike the modern mapped string binder. A bound
Global:false still selects GrantToTenant. The parameterless Identity set has no
valid grant operation and emits its source error rather than sending a request.
Group 202 is accepted by the compiled helper and decoded as the policy model;
raw HTTP/result metadata must remain available without inventing polling or
effective assignment semantics.

## Native implementation checkpoint and qualification remaining

TeamsEventsPolicies now exposes native compiled CRUD/list, all 25 typed settings,
explicit Set read/PATCH composition with partial results, user/tenant/group grants
and a credential-free missing-scope guard mapped to the parameterless set. Root
and allowlist wiring and owned snapshots are included. Initial race tests compare
all 24 provider requests in synthetic application/delegated modes, all 83 model
observations, five actual source filter selections and the parameterless grant
failure. All nine command/set pairs have native symbols and initial comparisons.

Before the implementation PR, complete the remaining public binder comparisons,
partial update and mutation failure/no-replay/context guards, compiling examples,
API docs, gated lab/source reads, full repository checks and reviewed sanitized
artifacts. Ledger rows remain in-progress/metadata-only. Modern mapping metadata
remains distinct from actual compiled dispatch; this checkpoint is not full
command or tenant qualification.

No tenant read/write has run for this family. Successful authorization,
server/cloud availability, updates/read-back/cleanup, grants and effective webinar,
townhall, recording/transcription/chat/ECDN/integration behavior remain unqualified.
Issue #118 stays open. Microsoft binaries and decompiled source remain private.
