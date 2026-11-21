# Targeting policies

`sdk.TeamsAdminAPI.TargetingPolicies` manages the native modern ConfigAPI tag
settings, independently from hierarchy uploads. Operations are `GetByIdentity`,
bare-array `List`, local `ListFiltered`, sparse `UpdateByIdentity` and
`DeleteByIdentity`. `UpdateGlobal` names the pinned omitted-identity default
explicitly. The native identity-addressed methods require an identity.

Identity is escaped as one path segment. PUT includes it in the body and requires
204; the SDK does not substitute generic PATCH. Nil settings are omitted, matching
the pinned mapper's null filtering; explicit empty strings survive. No setting
defaults are guessed. Mode strings remain open because pinned metadata does not
establish exhaustive service enums. Read snapshots retain unknown/null/empty
fields and serialize their received object. They are not mutation requests.

List requires the evidenced HTTP-200 bare array and stably puts Global first.
Filter is applied locally to Identity or Tag:Identity, with no server query or
invented pagination. The shared internal matcher preserves the existing SDK's
UTF-16 wildcard positions, bracket/range/backtick behavior and portable case
folding; locale-specific PowerShell collation is not inferred. Blank native filter
input selects the full list. Get requires HTTP 200; Delete requires 204.
Writes are sent once and never rolled back or automatically read back.

The per-session ConfigAPI adapter performs fixed-origin discovery once, normalizes
the returned forest into X-MS-Target-Uri and sends X-MS-CmdletName. The discovered
forest remains a header, never a credential destination. This uses the existing
ConfigAPI cloud/audience tuple and transport controls. It does not require an
unrelated Graph organization request just to discover policy routing.

This native service explicitly uses the modern REST contract. The pinned module
can select TRPS for unflighted tenants; the SDK does not invoke PowerShell or
silently claim that legacy parity is complete. Missing flight data or available
generated methods alone prove neither acceptance nor rejection of native REST.
Application authentication documentation names no exception for these three
commands, but exact route eligibility/RBAC and least privilege remain unqualified.
Graph consent is not a substitute for Teams resource authorization.

Policy Create/upsert is not established by the assigned commands or source.
No guessed POST/assignment or generic unsupported stub is exposed. The documented
Global delete reset/default-assignment effects are service behavior, not a guessed
reset payload. Confirm those effects in a disposable tenant before consequential
acceptance. See [routing and eligibility evidence](../../../docs/implementation/core/policy-routing-evidence.md)
and [pinned policy contract](../../../docs/implementation/core/template-policy-deployment-evidence.md).
