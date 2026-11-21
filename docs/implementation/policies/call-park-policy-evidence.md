# Call park policy: selected modern contract

POL-36 / issue #109, MicrosoftTeams 8.1.0, checked 2026-11-21. This source
checkpoint precedes native implementation and makes no completed command or
tenant qualification claim.

## Export and transport selection

All five exports are functions delegating to private
`Microsoft.Teams.ConfigApi.Cmdlets.*CsTeamsCallParkPolicy` classes inheriting
TrpsPSCmdlet. The exact eight sets are Get Filter/Identity, Grant GrantToGroup/
GrantToTenant/Identity, and New/Remove/Set Identity. The pinned command map selects
the corresponding Get/New/Set/Remove-CsConfigurationModern function with
ConfigType TeamsCallParkPolicy, and Grant-CsTeamsPolicy with the same PolicyType.

The five private classes and normalized/exported parameter declarations were
checked. New/Set expose Boolean AllowCallPark, string Description, and signed
Int64 ParkTimeoutSeconds/PickupRangeStart/PickupRangeEnd. No local range/length
attributes or resource-specific property-bag validators appear on those classes.
That does not prove the service accepts every value or a nonsensical range.

The selected modern routing, command-specific flights, discovery, audience and
callback are traced in [shared routing](../core/policy-routing-evidence.md) and
[the modern configuration chain](audio-conferencing-policy-evidence.md).
No flight result is observed for this family, and a native REST service must
not silently invoke a PowerShell/TRPS fallback.

## Reproducibility and captured stages

Pinned package SHA256:
`78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`.
Scripts require explicit module/output paths, reject existing output and check
the manifest and applicable source/private-assembly hashes.

- [Mapped wrappers](../../../tools/contracts/capture-call-park-wrappers.ps1):
  unmodified mapped function ASTs with synthetic telemetry and memory-only
  internal-command stubs. It validates all five pinned command mappings.
  Formatting adapters return values unchanged. No exported processor or network.
- [Generated requests](../../../tools/contracts/capture-call-park-requests.ps1):
  actual ConfigApiBasedCmdlets helper/DTO request construction and response
  dispatch, before session host/header rewriting. The injected memory sender
  refuses credential headers; no HttpClient/production pipeline is invoked.
- [Public binding](../../../tools/contracts/capture-call-park-binding.ps1):
  original New/Set parameter declarations and attributes in renamed functions
  whose synthetic bodies return bound arguments, followed by actual pure
  XdsConfiguration dictionary serialization. No original Begin/Process/End,
  custom runtime-attribute processing, flight/TRPS selection or tenant runs.

The stages are complementary. The binding seam does not exercise private
runtime-added validation or the real command body. No credentials, discovery,
tenant mutation or Microsoft binary/decompiled source is retained.

Fixtures contain
[22 wrapper cases](../../../contracts/endpoints/8.1.0/fixtures/call-park-wrappers.json),
[17 generated cases](../../../contracts/endpoints/8.1.0/fixtures/call-park-requests.json)
and [22 isolated binding cases](../../../contracts/endpoints/8.1.0/fixtures/call-park-binding.json).
Five wrapper cases fail as expected; six binding cases fail. Every generated
case sends once except pre-cancellation, which sends zero times.

## Selected operations and presence

| Operation | Selected modern behavior |
| --- | --- |
| List/Get | GET `/Skype.Policy/configurations/TeamsCallParkPolicy` and `/configuration/{identity}` |
| Create | POST collection, 201; mapped New then performs a separate Get |
| Update | Sparse PUT item, 204, including Identity; no preparatory Get |
| UpdateGlobal | Mapped Set supplies Global only when Identity was omitted |
| Delete | Bodyless DELETE item, 204 |
| Grant user/tenant/group | Generated PATCH user/tenant, 200; POST group, 202 or 204 |

Identities are whole escaped segments. Optional apiVersion/schemaVersion queries
are absent. The ordinary generated path discovers Teams.Tenant/serviceDiscovery
and sends the normalized forest HTTPS root in X-MS-Target-Uri while retaining
ConfigAPI as the credential destination and Teams as the audience. It has no
unconditional MPACmdlet. Source Get sorts Global first and applies local
Identity/Tag wildcards; whitespace Filter/Identity selects the collection in the
mapped helper. Server-side PUT omission/replacement rules remain unobserved.

Wire generation and isolated declaration binding preserve false, empty string,
zero, negative values, signed Int64 limits and 9,007,199,254,740,993 exactly.
There is no floating-point intermediate. PickupRangeStart and PickupRangeEnd
are distinct fields. Client-side validation must not invent an ordering rule or
documented service range based only on these permissive helper observations.

Omitted settings remain absent; identity-only New/Set bags serialize. Isolated
binding converts bound null Description to empty and null Int64 fields to zero.
Null Boolean input fails, as does Int64 overflow. A decimal numeric string binds
to Int64 exactly. The native API can use typed optional pointers: nil means
unbound, while explicit empty/false/zero values reproduce the corresponding wire
forms. Arbitrary PowerShell coercions are not a native runtime dependency.

Grant requires bound PolicyName; explicit null string binding becomes empty,
the captured clear form. Group Rank nil/zero is omitted by the mapped wrapper;
the generated DTO can send explicit priority zero. Missing scope/PolicyName and
bound PassThru true/false are recorded baseline failures. Group 202 acceptance
must refer to the actual POST, not discovery, and does not establish completion.
Mutations must not be replayed. New submission must survive a failed read-back.

## Outstanding qualification

Native service/root wiring, typed models, examples, ledger mappings and opt-in
lab support follow this checkpoint. Successful authorization, cloud/flight
deployment, pickup/timeout rules, CRUD read-back/cleanup, grant completion and
effective call parking remain open. No successful tenant read, mutation,
privilege change or pinned PowerShell/Go comparison is claimed.
