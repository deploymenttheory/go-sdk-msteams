# Online voicemail policies

`Client.TeamsAdminAPI.OnlineVoicemailPolicies` exposes compiled ConfigAPI policy
CRUD/list and separately generated assignment operations.

| SDK operation | MicrosoftTeams 8.1.0 mapping |
|---|---|
| GetByIdentity / List | Get / Identity; omitted Identity selects List |
| ListFiltered | Get / Filter, local Identity/Tag wildcard matching |
| Create / CreateWithGlobalComparison | New / Identity; native POST / explicit optional comparison then POST |
| PatchByIdentity / UpdateByIdentity | Set / Identity; sparse PATCH / explicit read-then-PATCH composition |
| DeleteByIdentity | Remove / Identity; DELETE with bound Identity/optional Force body |
| AssignUser / AssignTenantDefault / AssignGroup | Grant / Identity, GrantToTenant, GrantToGroup |
| AssignWithPowerShellBinding | Explicit mapped helper's PolicyName/Rank coercions |
| GetAssignmentOperation | Additional generated operation read after group acceptance |

CRUD `Settings` contains all thirteen source fields. policy.Set/Null preserve
omitted/value/null, false and empty strings. ShareData remains represented in the
8.1.0 surface. Audio-file fields are references; the SDK never opens/uploads them.
No recording-length range, language-pair restriction or tenant default is inferred
from mutable documentation. Read snapshots preserve received fields and raw future
data. Ordinary formatting is redacted; callers retain HTTP metadata/raw responses.

TimeSpan stores the full signed .NET 100ns tick range. TimeSpanFromTicks and
TimeSpanFromDuration construct typed values; the latter rejects sub-tick precision.
JSON uses Newtonsoft's constant duration string, including signs/days/seven-digit
fractions. ParseTimeSpanConstant accepts that wire form with overflow checks;
broader CLR/PowerShell locale/input binding remains separate qualification. Null
duration is distinct from omitted/zero. No seconds/ISO-8601 substitution is made.

Compiled CRUD uses the shared direct-policy protocol, MPACmdlet, command name and
fresh UUID correlation, with first-slash scope escaping. The generated assignment
adapter uses lazy per-service forest discovery, a target-URI header and separately
escaped full route segments. Both share the injected Teams scope/session, token
lifecycle and destination/audience guards; the discovered forest is only a header.
CRUD calls do not bootstrap generated discovery or look up users through Graph.

Assignment requires exactly one explicit scope and a bound PolicyName field.
policy.Null[string]() omits the generated DTO key; policy.Set("") sends empty.
Native group Priority nil/zero/nonzero remain distinct. AssignWithPowerShellBinding
maps bound null to empty and zero Priority to omission, following the pinned shared
helper. Explicit tenant scope represents the helper's truthy Global selection.
The native API always returns typed outcomes/metadata and exposes no PassThru display
switch; exported wrapper-to-helper PassThru binding remains a recorded source gap.

User/tenant success is HTTP 200. Group submission accepts 202 or 204; 202 retains
acceptance and an optional capitalized OperationId. Missing ID stays missing, and
malformed accepted JSON preserves the result/response with ErrDecode. Operation
reads use a separate route and lower-camel status fields, retaining open status,
timestamp text, count presence and raw user/future fields. No terminal-state or
effective-assignment inference, implicit polling or job deletion is provided.

Update/Create composition results retain preparation and HTTP acceptance on failure.
Writes never replay automatically. Read-back, rollback, atomic/ETag guarantees and
propagation are explicit caller concerns. The [unit tests](service_test.go)
and [lab plan](../../../docs/implementation/policies/online-voicemail-lab-acceptance.md)
separate candidate request correctness from successful tenant qualification.

### Executed owned CRUD

`TestAccOnlineVoicemailPolicyCRUD` passed with certificate and client-secret authentication.
It creates an owned unassigned policy, checks non-default read-back, updates a
setting while verifying omission, checks the list and independently verifies
selected HTTP404 and list absence after deletion. Each run has two public
MicrosoftTeams 8.1.0 read comparisons and one cleaned private journal entry.
The sparse description update preserves disabled transcription. The pinned
comparison covers Description, EnableTranscription and MaximumRecordingLength.

The actual HTTP201/204 replies, created/updated JSON, owned list and deleted404
body supply `TestUnitOwnedCRUDUsesAcceptanceResponses`. Unit checks retain exact
requests, typed settings, complete snapshots and typed errors. CI requires every
CRUD milestone, both comparisons and verified journal cleanup. Global, grants,
media behavior and remaining command/parameter variants are separate scenarios.
