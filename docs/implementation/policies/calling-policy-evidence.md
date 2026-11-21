# Teams calling policy evidence (POL-37 / #110)

The MicrosoftTeams 8.1.0 baseline exports five compiled Core cmdlets. Get has
Filter/Identity sets; Grant has GrantToGroup, GrantToTenant, GrantToUser and an
unhandled default Identity set; New, Set and Remove have Identity sets. These
select the compiled policy provider, rather than the modern configuration route.

## Reproducible offline observations

`tools/contracts/capture-calling-policy-provider.ps1` verifies the baseline
manifest and the Core, PolicyRp and HostContract assembly SHA-256 hashes. Its
owned C# adapter invokes the actual generic direct request helper with synthetic
InvocationInfo, an offline fixture token, a credential-refusing host and a memory
sender. It restores the static client, module version and host. No command
constructor/lifecycle, public parameter binder, discovery, credentials or network
is exercised. This is a development-only capture, not an SDK runtime dependency.

The promoted `calling-policy-provider.json` contains 21 provider cases and 144
actual model constructor/setter/Newtonsoft observations. Every provider case made
one memory dispatch and zero token requests. All 39 public settings are present
in the all-settings create/PATCH cases; the capture script checks model types.
The native implementation and contract comparisons are still pending.

## Selected contract

- CRUD uses the compiled Skype.Policy configuration resource and command headers.
  POST uses the requested Identity; PATCH uses the model Identity. Remove sends
  an Identity JSON body. A bound Force switch serializes as a string, including
  false. Empty and JSON-null reads succeed; an empty collection has a false
  source result flag without an HTTP exception. A collection envelope fails.
- User, tenant and group grants use the compiled routes. Null/empty PolicyName
  stay distinct; tenant Force is Boolean and group priority is signed Int32.
  Selecting no scope has no source operation. No native operation has been
  qualified yet; persistence, effective behavior and clearing remain open.
- The settings comprise 22 nullable strings, 15 nullable Booleans and two nullable
  signed Int64 values. Model nulls remain null, false/zero remain explicit, and
  CallingSpendUserLimit preserves values above 2^53 and the full signed range.
- CallRecordingExpirationDays accepts null or 1 through 9999 inclusive. The
  RecordingAndTranscriptionCustomMessageIdentifier accepts null or the exact
  hyphenated GUID regex, including uppercase and a final LF accepted by .NET's
  dollar anchor. Empty, future text and braces fail. Other string setters accept
  future values; Description accepts 2000 characters despite discarded generated
  length constants. No extra enum, path, spend, length or cross-field constraint
  is inferred. Server-side rules remain unqualified.
- Newtonsoft additionally converts numeric strings to Int64 and rounds JSON 1.5
  to 2; overflow fails. These are observed source decoding conversions, not
  public PowerShell binder evidence. A native typed API must express accepted
  final values explicitly and document any stricter JSON decoding boundary.
- JSON model projection accepts case variants and later duplicate members.
  Future fields need raw preservation separately from typed known fields.

## Remaining qualification

Implement native CRUD/list, explicit Set Get/PATCH composition, grants, typed
presence-aware settings, local model validation and offline comparisons. Record
opt-in lab evidence separately. The dedicated lab identity still lacks Teams
resource authorization; no role/consent or tenant mutation is part of this capture.
Keep issue #110 and full qualification open until the remaining evidence exists.
