# Network roaming and notification/feeds qualification

Owners: #138 (POL-65) and #139 (POL-66). Baseline: MicrosoftTeams 8.1.0.

Network roaming's public runtime entries are Functions. CmdletConfig selects
Remoting and Get/New/Set/Remove-CsConfigurationModern with ConfigType
TeamsNetworkRoamingPolicy. Generated HTTP code uses bare-array GET, POST/201,
PUT/204 and bodyless DELETE/204. ConfigName is escaped whole. New's wrapper
composes submission plus Get; Set defaults an omitted Identity to Global. Public
Boolean and Int64 settings are non-nullable and there is no Force parameter.
The older compiled classes are not substituted for the public proxy contract.

Notification/feeds' public entries are compiled Get/Set/Remove cmdlets. The pinned
manifest/runtime does not export New, although its compiled implementation exists.
The SDK's additional native CreateByIdentity follows that actual POST provider
and has live tenant proof; it is not recorded as an additional baseline export.
All three source settings are nullable opaque strings. The shared compiled
provider supplies Set preparation, sparse PATCH and DELETE body/Force behavior.

Primary local sources: public proxy definitions, CmdletConfig, merged modern
helpers and generated ConfigAPI create/read/list/update/delete API methods;
TeamsNotificationAndFeedsPolicy, its compiled cmdlets, BaseTpm* and provider
HttpRequestHelper. PowerShell oracles verify the selected pinned module files:

- ConfigAPI: 8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6
- Core: 68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49
- PolicyRP: a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51
- Merged helpers: 03dbc3a85e12c98dd1807d604782c6a90d0656569541cd9540738825a7fcb255
- CmdletConfig: 754d149c1a5904158e9f01eb55ba44d15c17b7e66f7556385ea894f4ff84d906
- Public proxies: 5a8a5a50081b1c6986b93d5b15723e966a2d22b64459a9c300cb2a1bf69bb3b5

On 2026-11-22 both certificate and client-secret runs with the dedicated SDK SPN
(Teams Administrator) read/listed Global and matched the pinned public getters.
Both modes also passed owned native CRUD. Network roaming creates a non-default
description, disabled video and explicit 50000 bitrate; its description-only PUT
preserves both omitted settings. Notification creates a non-default description
and explicit suggested-feed setting; its description-only PATCH preserves that
setting. Both compare created/updated getters with PowerShell, list their unique
policy, delete it and require independent typed GET404 plus collection absence.

Notification additionally PATCHes Description as JSON null, verifies the getter
is null while the other setting survives, and compares the pinned module a third
time. Both application modes passed. Its dedicated CI suite requires this null
assertion and all three comparisons, as well as ordinary CRUD and a cleaned
recovery entry. Network roaming joins the ordinary owned-policy CRUD suite.
Every create intent is durably journaled before the request; generic bound recovery
registers both policy kinds. No existing/Global policy, network site, user/group
assignment, license, role or organization is changed.

Twenty permanent sanitized service response files come directly from these
acceptance tests. Unit tests load them for typed fields, full snapshots, routes,
headers, filters, mutation receipts, null read-back and actual deletion errors.
Empty failing and deliberately malformed replies exercise negative contracts;
no synthetic successful response store is introduced. Examples compile without
tenant access. Each public command/parameter set has symbol and test mappings.

Remaining live qualification: effective network/feed behavior, empty-string
persistence, modern numeric boundary rejection, delegated modes, locale-specific
filter behavior and other clouds. Merged code and selected-operation tests do not
by themselves promote whole-ticket/release coverage states.
