# Custom app settings

Use `sdk.TeamsAdminAPI.CustomAppSettings` for the tenant-wide Teams apps
settings, including whether users can interact with sideloaded (custom) apps.
Operations use the shared client/auth/transport controls through the ConfigAPI
backend and need no PowerShell runtime. Constructing the root client or this
service performs no network I/O.

| Resource method | PowerShell command | Wire operation | Declared successful statuses |
| --- | --- | --- | --- |
| `Get(ctx)` | enhanced native document read | GET `/Teams.MiddletierService/tenantWideAppsSettingsGlobal` | 200 |
| `GetSideloadedAppsInteraction(ctx)` | `Get-CsTeamsSettingsCustomApp` public projection | GET the same route; exact-key nullable bool → false | 200 |
| `SetSideloadedAppsInteraction(ctx, enabled)` | enhanced convenience | GET, then PUT the same route | 204 |
| `Update(ctx, UpdateRequest)` | (the PUT the cmdlet sends) | PUT `/Teams.MiddletierService/tenantWideAppsSettingsGlobal` | 204 |

Every method returns the actual Resty response alongside its value/error. `Get`
is replay-safe; PUTs are never replayed automatically. Any other HTTP status is
an `*client.APIError`; an undeclared success status is `ErrResponse`.

## Reading

`Get` returns the whole document: the six `is…Enabled` switches, the `lob…`
branding strings, `appAccessRequestConfig` and the per-app `appSettingsList`,
with `RawFields` keeping every property. The cmdlet shows only
`isSideloadedAppsInteractionEnabled`. `GetSideloadedAppsInteraction` reads that
public projection directly, using exact raw-key nullable-bool semantics: missing,
null, wrong-case or malformed scalar values become false as in pinned source.
Unrelated enhanced-document fields cannot prevent that projection. Read/status
errors remain errors and must not be interpreted as a disabled setting.
`Settings.SideloadedAppsInteractionEnabled()` is the corresponding snapshot helper
when an enhanced document has already decoded successfully.
It reads the canonical raw key when a snapshot is present, so an unexpected
uppercase alias cannot change the pinned public output. Native full-document
fields retain Go JSON matching and stricter wrong-type validation; they are an
explicit enhancement beyond the generated DTO/public one-property projection. The list
reflects the service-managed app catalog (thousands of entries in the lab) and
its length can change between reads.

## Writing

`PrepareSideloadedAppsInteraction(ctx, enabled)` reads the source object with
`X-MS-CmdletName: Get-CsTeamsSettingsCustomApp_Get`, the child invocation traced
through the pinned internal proxy and shared header insertion. It freezes a
source-compatible projection without issuing a PUT. `Projection()` and `Before()`
return owned copies; inspecting or editing those copies cannot change the body.
`UpdateSideloadedAppsInteraction(ctx, prepared)` accepts only a request prepared by
that exact service and sends one PUT with Set metadata. The two requests are not
atomic. A caller must account for concurrent changes between preparation and apply.

The pure `ProjectSideloadedAppsInteraction` keeps exact-case nullable booleans and
PascalCase `Lob…` strings, replaces the requested flag, and emits an empty
`appSettingsList`. It copies only the two known exact-case config strings;
null/malformed/unknown fields are omitted, and an empty projected config object
is omitted entirely. Empty strings and false values survive. Omitted/changed paths
show differences from the read snapshot; they do not prove server persistence.
Default formatting of projections, prepared requests and action results is redacted.

Preparation errors retain the read response and produce no applicable request.
The action result retains independent read/action receipts and whether a write
was invoked. `HTTPAccepted` records a 2xx receipt even when the source requires
204 and the operation returns `ErrResponse`. Cancellation before PUT preserves
the read receipt and records no invocation. A transport failure after invocation
can be ambiguous: never blindly replay a prepared write. This API does not claim
that the frozen snapshot is still current or that a write persisted every field.

The existing `SetSideloadedAppsInteraction` remains an enhanced convenience with
strict full-document decoding, preserved nested unknown properties and omitted
branding. Its established live selection and behavior are unchanged. The source
prepare/apply API has offline source/replay coverage only; Set is not yet marked
contract-complete or live-qualified. The earlier authorized lab observation that
an empty app list left managed entries unchanged is not a universal guarantee.

`Update` sends the given document unchanged (nil fields omitted, a nil
`AppSettingsList` sent as `[]`). The `Lob…` keys use the module's PascalCase
request names. Every list entry needs an `ID`.
