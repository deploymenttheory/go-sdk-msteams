# App catalog contracts

Owner: #84 (EXT-21). Baseline: MicrosoftTeams 8.1.0.

The seven commands use two backends. Get/Set-CsTeamsSettingsCustomApp are
custom script wrappers over ConfigAPI cmdlets. The TeamsApp and
Get-TeamsArtifacts commands are compiled cmdlets in
`Microsoft.Teams.PowerShell.TeamsCmdlets.dll` that call Graph directly.

## Custom app settings (ConfigAPI)

- The generated `TeamsSettingsCustomAppGet`/`Set` methods both target
  `/Teams.MiddletierService/tenantWideAppsSettingsGlobal`: GET for the read and
  PUT with a JSON body for the write.
- `GetCustomAppSettingResponse` reads `isAppsEnabled`, `isAppsPurchaseEnabled`,
  `isExternalAppsEnabledByDefault`, `isLicenseBasedPinnedAppsEnabled`,
  `isSideloadedAppsInteractionEnabled`, `isTenantWideAutoInstallEnabled`,
  `LobTextColor`, `LobBackground`, `LobLogo`, `LobLogomark` and
  `appAccessRequestConfig` (`adminInstructionMessage`, `approvalPortalUrl`).
- `UpdateCustomAppSettingRequest` writes the same names plus `appSettingsList`.
- The Get wrapper selects only `isSideloadedAppsInteractionEnabled` and
  replaces null with `$false`.
- The Set wrapper calls the internal Get, then the internal Set with
  `isSideloadedAppsInteractionEnabled` and the other properties copied from the
  read, plus `appSettingsList = @()`.
- The service returns camelCase `lobTextColor`… names and an
  `appSettingsList` of `{id,isEnabled,isAppLevelAutoInstallEnabled,context}`
  entries, which the generated model does not declare.
- The generated model reads only the PascalCase `Lob…` keys. AutoRest JSON
  lookup is case-sensitive, so the module reads null and omits the branding
  strings from its PUT. The SDK does the same. This is inferred from the
  serializer and was not seen on the wire.

## Teams apps (Graph)

All three write cmdlets surface Graph errors through
`DeserializeMSGraphApiErrorRoot`. The routes and checks, from the IL:

- `Get-TeamsApp` builds `$filter` clauses `id eq '…'`, `externalId eq '…'`,
  `displayName eq '…'` and `distributionMethod eq '…'`, in that order. It joins
  them with ` and ` and reads every page of GET
  `{graph}/v1.0/appCatalogs/teamsApps` through `HttpUtilities.GetAll`. Values are
  not escaped.
- `New-TeamsApp` throws `Operation not allowed.This call only supports
  DistributionMethod: Organization` unless the method equals `organization`,
  ignoring case. It then POSTs the file bytes as `application/zip` and returns
  the deserialized body.
- `Set-TeamsApp` PUTs the file bytes as `application/zip` to
  `/v1.0/appCatalogs/teamsApps/{Id}`. It accepts any 2xx and writes a success
  string.
- `Remove-TeamsApp` sends DELETE `/v1.0/appCatalogs/teamsApps/{Id}`.

## Teams artifacts (Graph)

`Get-TeamsArtifacts` runs `ExportJob` with its own client-credentials token, built
from the `EXPORT_TOOL.CLIENT_ID`, `EXPORT_TOOL.TENANT_ID` and
`EXPORT_TOOL.CLIENT_SECRET` variables. It writes `artifacts.json` to the process
directory.

- `DoJob` scans SharePoint unless `-OneDrive` is set, then OneDrive unless
  `-SharePoint` is set.
- SharePoint: GET `/v1.0/sites?search=*`, then `/v1.0/sites/{id}/drives`, and
  traverses each `/v1.0/drives/{id}`.
- OneDrive: GET `/v1.0/users` (or `?$filter=mail eq '…'`), and traverses each
  `/v1.0/users/{id}/drive`.
- The traversal reads `{drive}/items/{item}/children` from `root`, following
  `@odata.nextLink`.
- The `ShouldRecurseFolder` and `ShouldOutputArtifact` IL was decoded
  opcode-for-opcode. Comparisons use `OrdinalIgnoreCase`, and the SDK
  reproduces the branches exactly (see the package docs).
- `ArtifactType` values are RecordingTranscript=0 (`Recordings`, `.MP4`),
  Notes=1 (`Meetings`, `.LOOP`) and Whiteboard=2 (`Whiteboards`,
  `.WHITEBOARD`).
- Every non-success is printed and skipped.

## SDK differences

- Filter and mail values are OData-escaped (`'` → `''`). The module sends
  `displayName eq 'o'brien'` and Graph rejects it with HTTP 400, as seen in
  the lab.
- Get-TeamsArtifacts returns `Result{Artifacts, Failures}` instead of writing
  `artifacts.json` and console lines, and authenticates with the client's
  Graph session.

## Lab observations

All runs used client-secret, app-only Teams Administrator credentials with
Graph `Organization.Read.All`.

- The settings GET returns 200 with the full document, including about
  4,000 `appSettingsList` entries whose count varies between reads.
- A PUT of the module-shaped body with the current flag and
  `appSettingsList: []` returned 204. The switches and branding were
  unchanged, and the per-app entries kept the same set and values; only their
  order changed. The original document was still PUT back afterwards.
- `Set-CsTeamsSettingsCustomApp` with the current value also succeeded.
- Graph app catalog GET/POST/PUT/DELETE all return 403 `Forbidden` ("API
  requires one of 'AppCatalog.Read.All, AppCatalog.ReadWrite.All'"), for both
  the SDK and pinned PowerShell.
- The pinned `Get-TeamsApp -DisplayName "o'brien"` gets 400 `BadRequest`
  (invalid filter).
- `/v1.0/sites?search=*` returns 403 `accessDenied`. Get-TeamsArtifacts
  -SharePoint writes an empty `artifacts.json` and reports one error, matching
  the SDK's single `sites` failure.

Acceptance:

- `TestAccAppCatalogReads` (in `tests/integration/lab/custom_app_settings`; read-only, with oracle
  `read-app-catalog.ps1`) checks the settings read, the catalog 403 and the
  SharePoint scan failure, each with pinned PowerShell parity. The artifact
  parity check needs client-secret auth.
- `TestAccAppCatalogWrites` (serial) writes the current sideloading flag back
  and compares the switches and branding. It also sends create, update and
  delete probes with an empty package and the reserved ID
  `00000000-0000-4000-8000-000000000001`, which are denied with 403.

Unit fixtures come from sanitized lab responses (`list_403`, `filter_400` and
`sites_403`), plus a reduced copy of the settings document.

Unqualified, because they need permissions the lab identity lacks:

- Successful app catalog list/create/update/delete (`AppCatalog.*`).
- Artifact discovery in real drives (`Sites.Read.All`, `Files.Read.All`).
- Changing the sideloading flag to a different value. Same-value writes only,
  to avoid changing tenant behaviour.


## Custom-app Get and status checkpoint (2026-11-27)

The [pinned source observation](evidence/custom-app-settings-get-status-source.json)
records the exact public/custom/internal script definitions and file/definition
hashes, generated method identities and selected full status/parser IL bodies.
Get is direct AutoRest, exported to TeamsGA/TeamsPreview/SelfHost, with one
`__AllParameterSets` surface. It performs a no-body GET of
`/Teams.MiddletierService/tenantWideAppsSettingsGlobal`, expects HTTP 200 and an
object, and uses ConfigAPI resource `48ac35b8-9aa8-4d74-927d-1f4a14a0b239`.
Unlike cloud-call-data's TRPS path, no legacy-flight branch is needed for this Get.
The custom wrapper projects only the exact `isSideloadedAppsInteractionEnabled`
property, replacing missing/null with false. The compatibility getter now uses
that exact raw key, including uppercase-only and mixed-case regression cases.
Native full-document branding, app list and raw retention are explicit extensions;
Go typed JSON matching and stricter malformed-type errors are not described as
byte-for-byte generated parser equivalence.

Generated Set dispatch declares only HTTP 204. The prior authored 200 success
vector is corrected to 204 and a separate 200 error case retains its receipt.
The existing real same-value update observation was 204 and remains unchanged.
This checkpoint changes no live selection, authorization or tenant configuration.

Get's complete ordinary set/source, successful fixture, null/absent/false output,
malformed shapes/statuses and typed failures are reviewed separately from Set.
No implementation, contract or live ledger state is promoted without its own
revision-backed execution evidence. Further source-confirmed Set differences
remain open: nested unknown appAccessRequestConfig properties are dropped by the
source and empty nested objects are suppressed; PascalCase Lob* response values
are copied while camelCase branding is ignored. Those behaviors are not claimed
fixed by this read/status checkpoint or by the existing enhanced native document.


### Direct public flag projection

The [two scalar observations](evidence/custom-app-get-scalar-projection.json) and
[bounded pure reproducer](../../../scripts/discovery/inspect-custom-app-get-projection.ps1)
confirm both wrong-case and canonical string-valued flag fields parse successfully
in the pinned model, leave its nullable flag unset, and project false. The complete
native document deliberately retains stricter extension-field decoding, so its
snapshot accessor alone cannot cover every public-command output.

`GetSideloadedAppsInteraction` is the mapped direct projection primitive over the
same no-body GET, ConfigAPI metadata and HTTP 200 status. It requires a non-null
object, then projects only the exact flag key, independent of unrelated full-model
fields. True/false/null/missing and both scalar observations reach this public path;
unsupported statuses, malformed roots and authorization/cancellation errors retain
their responses/errors rather than become successful disabled-setting observations.
Enhanced Get remains unchanged. The new path is replayed against the existing
actual sanitized read fixture as well as explicitly authored source vectors.
Set body/persistence and live states remain separate from this ordinary Get review.


### Ordinary Get decision at revision 4338e4b

[The versioned contract result](evidence/custom-app-get-4338e4b-contract.json)
records the full custom-app package TestUnit execution and example compilation at
`4338e4b452341441d133d91a30d90f297eece0d2`, source/control/set review and file hashes.
All required tests passed with no failures/skips. Its single ordinary set maps to
`GetSideloadedAppsInteraction`; strict enhanced Get is recorded separately. Every
source-accepted flag outcome under review can reach the raw public projection,
including both malformed scalar observations, while errors remain errors.
This independently reviewed Get is implemented/contract-passed in the ledger.
Set and live states remain unchanged. A replayed real read body is evidence of
this revision's parsing, not a newly executed tenant operation or authorization test.


## Source-compatible custom Set preparation

The new pure projection and explicit prepare/apply API close the known nested
config and PascalCase branding differences without changing the existing enhanced
convenience or its authorized lab selection. Twelve pinned pure DTO/serializer
observations are replayed from `testdata/pinned-set-projection.json`; their replies
are controlled source vectors, not tenant observations. The exact child Get header
is source-traced, not an intercepted live request. See
[evidence](evidence/custom-app-set-header-source.json) and
[projection distinctions](evidence/custom-app-set-projection-source.json).

Prepared requests bind to their exact service and own frozen body/snapshot copies.
Tests cover forbidden/malformed preparation, foreign-service rejection, cancellation
before the PUT, retained independent receipts, source 204 success, unsupported 200
acceptance and 403 failure. Session 90802 passed the full custom-app package's
TestUnit/Example selection after explicit expected-status fields were corrected in
the new harness cases. This is offline execution, not proof of tenant persistence.
The Set ledger state remains unchanged pending individual full-source/set review
and revision-bound execution evidence.

The complete affected package passed again at clean frozen revision `4e7df9d`
(including the new ambiguous-action receipt cases and compiled example). The
[revision-bound result](evidence/custom-app-set-4e7df9d-offline.json) records exact
source hashes, raw Go JSON digest and every passing test event. Set/live statuses
remain unchanged; this is a reviewable implementation checkpoint.


## Binding and header correction after the DTO-only checkpoint

Actual inert parameter binding followed by the real pinned SetExpanded setters
and serializer supersedes the earlier DTO-only Set parity assumption. Fourteen
observations (the original twelve plus malformed/casing boolean cases) show that
all five copied nullable booleans bind as present false when missing/null/invalid,
and all four copied null Lob strings bind as present empty strings. Nested typed
config omission remains a separate generated-model rule. Historical DTO captures
and the `4e7df9d` execution record remain unchanged, with their narrower phase
explicitly preserved; they do not establish whole-wrapper Set parity.

The exact shared header trace also corrects the earlier public Get metadata
assumption: public Get and Set's preliminary Get use private child
`Get-CsTeamsSettingsCustomApp_Get`; the source write uses
`Set-CsTeamsSettingsCustomApp_SetExpanded`. Default ProcessArgs hooks signal
execution identity and do not overwrite command headers. The earlier individual
Get qualification at `4338e4b` is withdrawn pending corrected exact-header replay
and a replacement reviewed decision; its historical evidence is not rewritten.
Enhanced native full-document Get/direct Update keep their documented public-name
convention separately. See [correction evidence](evidence/custom-app-binding-header-correction.json).

The parameterized repository repro `scripts/discovery/observe-custom-app-bindings.ps1`
was also executed against the verified package: all fourteen real bodies matched
the preserved binding capture, without errors. Its inert C# command captures
converted parameter values; actual pinned constructor/setters/serializer produce
the separately recorded real body. Bounded [setter IL](evidence/custom-app-binding-setters-source.json)
and exact repro/capture hashes are retained in the correction evidence. Full
corrected custom-app TestUnit/Example execution passed session 64518; both Get
and Set remain unqualified through this source-review checkpoint.

A separate fifteenth real binder/setter observation reads a true target and
requests false; the emitted target is present false. The final portable repro and
Go pure replay include it, and the apply harness exercises requested false with
a true read snapshot. The original fourteen-case capture remains separately
preserved as `pinned-set-binding-projection-14.json`.

The final complete affected package passed at clean frozen revision `b3166ee`: 44
passing test events and no fail/skip, including the requested-false case and
corrected private-header replays. [The result](evidence/custom-app-b3166ee-offline.json)
records exact source/fixture hashes and raw Go JSON digest. Both command states
remain unqualified; this result alone does not restore the withdrawn decision.

The [bounded header closure](evidence/custom-app-header-closure.json) now retains
44 complete resolved method bodies across 20 fully qualified types, exact wrapper
script bodies, the sole command-header literal writer, invocation adapter, default
execution-ID handler/forwarder, terminal send and clone behavior. The complete
private-assembly literal scan resolved 57,736 string operands across 70,462 method
bodies with zero unresolved strings. Nine actual inert private-name/alias/proxy
observations exercise the real pinned PSInvocationInfo adapter: InvocationName can
change, while MyCommand.Name remains the selected private command and supplies
the header. The parameterized `scripts/discovery/inspect-custom-app-header-closure.ps1`
reproduces this bounded source-plus-inert evidence. No actual management lifecycle,
authentication, HTTP request or wire header capture is claimed.
