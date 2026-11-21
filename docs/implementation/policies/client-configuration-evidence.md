# Teams client configuration evidence (POL-39 / #112)

The MicrosoftTeams 8.1.0 baseline exposes compiled Core Get/Set commands for
TeamsClientConfiguration. Get has Filter/Identity sets and Set has Identity, for
three exact command/set pairs. Neither command has a modern CmdletConfig mapping.
Get inherits BaseTpmGetCmdlet and Set inherits BaseTpmSetCmdlet with bound-property
assignment for all 17 settings. No create/delete/grant operation is assigned or
evidenced for this configuration resource.

## Reproducible source capture

`tools/contracts/capture-client-configuration-provider.ps1` verifies the pinned
manifest and Core, PolicyRp and HostContract hashes. Its owned C# adapter invokes
the actual direct HttpRequestHelper with synthetic InvocationInfo, offline fixture
token, credential-refusing host and memory sender. It restores the static client,
module version and host, and never runs a command constructor/lifecycle, public
parameter binder, credentials, discovery or network.

The promoted `client-configuration-provider.json` contains 11 provider captures
and 60 actual model constructor/setter/Newtonsoft observations. Every provider
case dispatches once with zero token requests. All method/path observations are
compiled GET/PATCH on the exact TeamsClientConfiguration resource. The fixture
contains all 17 public settings and no synthesized create/delete/grant calls.

## Selected contract

- Compiled read/list/PATCH uses MPACmdlet, command and correlation headers, with
  no modern forest target header. PATCH serializes model Identity and only bound
  settings. Bound Force serializes as a string, including false. Empty/null reads
  succeed. An empty bare list sets the source success flag false without an HTTP
  exception; an envelope fails. Authorization failure retains its status.
- All 17 settings are nullable: thirteen Booleans and four strings. The strings
  are RestrictedSenderList, ContentPin, ResourceAccountContentAccess and
  UseUnifiedDomain. RestrictedSenderList stays opaque text; no list parser,
  delimiter normalization, mailbox validation or allowlist edit is implied.
  ContentPin is retained as a string, including leading zeros. No setter enum,
  range, length or cross-field constraint is present.
- Model observations cover omission, null, false/true, empty/future strings,
  2000-character sender text, opaque sender delimiters/leading-zero content text,
  case variants and later duplicates. Object-valued Boolean and array-valued
  RestrictedSenderList fail. These observations establish actual model behavior,
  not the public PowerShell binder's conversion of every possible input.
- Native known fields require actual JSON string/Boolean types and retain
  unknown fields separately in owned raw snapshots. Set exposes an explicit
  read-then-sparse-PATCH composition with partial preparation/acceptance results,
  alongside a one-request PATCH primitive. Only bound settings are sent;
  Force alone is not a setting. No implicit replay, restore, ETag or atomicity
  guarantee follows the helper behavior.

## Remaining implementation and qualification

Native read/list/filter and sparse PATCH/Set composition for all 17 settings are present, with root wiring, exact mappings, source comparisons, context/backend/input/failure tests, compiling examples and gated read-only lab support. Provider comparisons use application/delegated synthetic modes; model projections and accepted final settings compose the captured model/helper contracts. This is not public binder or tenant execution evidence. Read models retain owned exact raw snapshots and known typed projections, including later case variants/duplicates. Native JSON requires actual string/Boolean types; arbitrary Newtonsoft primitive coercions remain a declared boundary. Actual lab evidence is recorded in `client-configuration-lab-acceptance.md`.
Tenant/server/cloud/flight rules and effective email/provider/guest/resource/
people-search/domain behavior remain unqualified. Issue #112 and full command/
tenant qualification stay open; these source captures are not live validation.
