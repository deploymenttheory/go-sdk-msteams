# Custom-app Get and Set replacement decision review

Status: approved after independent source, header, code/fixture-hash and execution
review. This decision records implementation and offline contract completion for
these two public command sets only; live states remain unchanged. It does not
claim tenant qualification or close EXT-90.

| Public command | Accepted public set | Native source path | Reviewed decision |
| --- | --- | --- | --- |
| Get-CsTeamsSettingsCustomApp | __AllParameterSets | Service.GetSideloadedAppsInteraction | implemented / contract passed |
| Set-CsTeamsSettingsCustomApp | __AllParameterSets | Service.PrepareSideloadedAppsInteraction, ProjectSideloadedAppsInteraction, Service.UpdateSideloadedAppsInteraction | implemented / contract passed |

The accepted immutable runtime metadata is `contracts/powershell/8.1.0/runtime.json`:
Get command index 152 and Set index 622, parameter-set index 0. Internal Get and
SetExpanded names describe source dispatch, not additional public parameter sets.
Configured source exports include TeamsGA, TeamsPreview and SelfHost; this is not
execution proof for every cloud/profile/authentication mode.

## Source and request review

The public proxies select the custom wrappers and direct internal AutoRest
helpers; this family has no legacy TRPS flighting branch. Both requests use the
ConfigAPI route `/Teams.MiddletierService/tenantWideAppsSettingsGlobal`, its bound
ConfigAPI audience and shared forest discovery. Get is a bodyless GET requiring
200. Set reads first with the private child Get header, then sends a single PUT
with private SetExpanded metadata, requiring 204. Neither command silently
selects Graph or a different authentication method.

The full resolved header chain and exact scripts are committed in
[evidence](evidence/custom-app-header-closure.json), with a
[parameterized source/inert reproducer](../../../scripts/discovery/inspect-custom-app-header-closure.ps1).
The source adapter reads MyCommand.Name, not InvocationName. Nine actual inert
private-name/alias/steppable observations use the real pinned adapter. The default
execution-ID step/handler does not rewrite headers; the committed sole writer,
terminal and clone bodies close that path. This is source plus inert evidence,
not an actual management request capture.

Get uses the exact lower-camel flag key. True/false survive; absent, null,
unrecognized casing and malformed nullable scalars project false. Unrelated
strict enhanced-document decoding cannot block this public projection. Root,
HTTP/status and cancellation failures remain errors with their available receipt,
so false plus an error is not a successful disabled-setting observation.

Set's public input is one mandatory Boolean. It always performs the source
read/replace flow rather than inventing a no-op for equal values. Five copied
nullable booleans pass through actual source binding: missing/null/unrecognized
values become present false. Four copied PascalCase branding strings bind null
as present empty strings; recognized values, including empty strings, survive.
The requested flag overrides the read, including explicit false. The typed nested
config parser drops unknown/wrong-case/invalid nullable members; only its two
known strings survive. An empty typed config is omitted; known empty strings
preserve its presence. The source list is always empty.

[Fourteen preserved observations](../../../msteams/teams_admin_api/custom_app_settings/testdata/pinned-set-binding-projection-14.json)
and the [fifteen-case follow-up](../../../msteams/teams_admin_api/custom_app_settings/testdata/pinned-set-binding-projection.json)
exercise inert binding followed by the actual pinned SetExpanded constructor,
property setters and serializer. The fifteenth reads true and requests false.
The [portable binding repro](../../../scripts/discovery/observe-custom-app-bindings.ps1)
reproduces all fifteen bodies without management lifecycle/authentication/HTTP.
The earlier DTO-only captures remain unchanged and explicitly do not establish
whole-wrapper binding parity.

## Native controls and failure behavior

Native Go Boolean arguments preserve explicit false without a separate presence
flag. Source Set with WhatIf still performs its unhooked preliminary GET and withholds
the PUT. A native caller can Prepare, inspect and withhold Apply. Confirm maps to
the caller's decision whether to apply; no PowerShell prompt or extra approval
switch is introduced. PowerShell prepend
callbacks are not executed: native shared transport/middleware injection is the
separate common-control mapping, and can apply to both read and write. Source
Set's preliminary read does not receive the outer wrapper's prepend hook. Logging,
output and error preferences use native client controls and Go return values.

Preparation failure produces no applicable request and retains its read receipt.
An opaque request binds to its preparing service and owns frozen body/snapshot
copies. Inspection cannot change the wire or retarget the tenant. Cancellation
before PUT retains the read receipt with no action invocation. Action results
retain separate read/action receipts, invocation and 2xx acceptance flags even
when status, transport or cancellation errors prevent successful completion.
Unsupported 200 is an accepted receipt plus ErrResponse; only 204 is the source
success. Lost response is ambiguous and must never prompt blind replay. PUT has
ReplayNever; source snapshot concurrency is not hidden by an automatic reread.
Projection/body formatting is redacted. Explicit projection copies show omitted
and changed fields without claiming persistence.

The existing enhanced full-document Get, direct Update and older convenience
retain their separate public-header/native document conventions. The source
command mapping does not rely on those extensions. Their established live
selection and wire behavior remain unchanged.

## Execution and limits

Clean frozen code revision `b3166ee` passed the complete affected package's
TestUnit/Example selection under the disk guard: 44 passing test events, zero
failure/skip, exit 0. The [execution record](evidence/custom-app-b3166ee-offline.json)
contains exact code/fixture hashes and the raw Go JSON digest. It includes all
fifteen real binding replays, false apply with a true read snapshot, private
headers, 204/200/403, denied/malformed preparation, frozen copies/service binding,
pre-write cancellation, lost response and accepted/canceled action receipts.
Examples compile without executing a tenant workflow. Source vectors and HTTP
replies remain separately labeled; authored 204/200 outcomes are not live proof.

Historical Get revision `4338e4b` assumed the wrong public header and its earlier
qualification is withdrawn; historical DTO-only Set revision `4e7df9d` omitted
binding coercions. Their evidence is preserved, with
[explicit corrections](evidence/custom-app-binding-header-correction.json).
This replacement requires the corrected source chain and `b3166ee`
execution together. Live flags stay unchanged: no new tenant request, mutation,
authorization grant, field-persistence proof or release-readiness claim follows.
