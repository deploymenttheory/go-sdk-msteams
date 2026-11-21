# Teams feedback policy evidence (POL-46 / #119)

Actual 8.1.0 public exports load compiled Core Get/New/Set/Remove cmdlets and a
modern Grant proxy Function. Grant maps to Grant-CsTeamsPolicy with
PolicyType=TeamsFeedbackPolicy. The remaining compiled Grant class is not the
selected public export and is not used as evidence for native assignments.

## Executed source captures

The provider capture pins Core/PolicyRp/HostContract, verifies all four compiled
exports and the actual Grant Function, and executes the pinned direct compiled
HttpRequestHelper with synthetic InvocationInfo, credential-refusing host and
memory sender. It contains 15 CRUD provider cases and 26 actual model
constructor/setter/Newtonsoft observations. Each provider case makes one memory
send and zero token requests. Static client/version/host state is restored.
Original authentication, scope authorization, lifecycle and network do not run.

All six settings are exposed by the source model: UserInitiatedMode and
ReceiveSurveysMode strings; nullable Boolean AllowScreenshotCollection,
AllowEmailCollection, AllowLogCollection and EnableFeatureSuggestions. Null,
empty, false, true and future strings are retained; no source enum/cross-field
validator appears in these setters. Source model comparisons accept a long mode
string and case-insensitive duplicate names. Wrong Boolean objects/string arrays
fail. Public command binding remains a separate seam.

The inherited compiled binding capture records 36 CRUD binder/predicate cases
(32 successes and four validation failures), plus five actual protected Get
filter selections. Owned subclasses inherit original properties, attributes and
constructors, replace Begin/Process/End with synthetic output, and invoke actual
protected New/Set setting predicates. Every observation has zero sends/token
requests under the restored memory client/credential-refusing host. New's
Identity-only predicate is false but its base processing permits create; Set
requires a bound setting. No compiled Grant class is invoked.

The isolated pinned Grant-CsTeamsPolicy wrapper capture verifies the command
mapping and records 11 observations through a synthetic telemetry host and
in-memory internal commands. User/tenant/group forwarding, null-to-empty policy
name coercion, omitted/zero/nonzero/negative Rank and missing-scope/name/PassThru
failures are observed. It does not run the original proxy/Trps lifecycle,
flighting, generated HTTP pipeline or tenant authorization.

Seven generated modern service helper observations pin the private assembly and
execute request construction/response dispatch through the shared memory sender.
User/tenant PATCH uses lower-case policyName; group POST uses optional priority.
Explicit native priority zero is representable, while the mapped wrapper omits
Rank zero. Group 202 and 204 select distinct handlers. This seam runs before the
production session host/header rewrite, discovery and credential pipeline.

## Native implementation and qualification remaining

Compiled CRUD uses bare collection reads, scoped identity splitting, POST/sparse
PATCH/body-bearing DELETE and string Force. Set retains read-model Identity.
Modern Grant uses forest-target discovery and modern assignment payloads. These
observations require separate backend wiring in the native service.

This source checkpoint does not yet provide a native feedback policy service.
Remaining work includes native typed models,
CRUD/list/Set/grants and root wiring; comparisons and failure/context/no-replay
tests; API docs/examples; gated lab/source reads; command/set ledger mappings;
full checks and sanitized CI review. Issue #119 remains open. No tenant read/write
has run for this family, and no successful CRUD, persistence, assignment or
effective feedback/survey/collection behavior is claimed. Microsoft binaries and
decompiled source remain private.
