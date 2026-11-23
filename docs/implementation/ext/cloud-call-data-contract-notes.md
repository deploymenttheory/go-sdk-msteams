# Cloud call data connection contracts

Owner: #85 (EXT-22). Baseline: MicrosoftTeams 8.1.0.

CmdletConfig marks Get-CsCloudCallDataConnection and
New-CsCloudCallDataConnection as Remoting cmdlets dispatched to
`Get-CsCloudCallDataConnectionModern` and `New-CsCloudCallDataConnectionModern`.
Neither public nor modern cmdlet declares a parameter beyond the hidden
`MsftInternalProcessingMode` and runtime switches.

The generated ConfigApiBasedCmdlets methods give one route,
`/Teams.InternalSupport/CloudCallDataConnection`: GET for Get and POST for New.
The POST builder sets no request content. Both return `ICloudCallDataConnection`,
whose serializer has the single property `Token`.

Lab observations (client-secret, app-only Teams Administrator):

- GET returns HTTP 403 with a top-level `{"code":"Forbidden","message":"Access
  Denied.",…}` body; pinned PowerShell fails with
  `Forbidden,…GetCsCloudCallDataConnectionModern_Get` and writes a correlation ID
  line straight to the console.
- New was sent live on 2026-11-23 with explicit approval, accepting that a
  success could not be undone (tenant singleton, no remove cmdlet). The POST
  returns HTTP 403 with the same top-level `{"code":"Forbidden",…}` body as
  GET, so no connection was created. Pinned PowerShell
  New-CsCloudCallDataConnection also fails with a `Forbidden,…` error ID. A
  second POST and a GET afterwards return the same 403.

Acceptance:

- `TestAccCloudCallDataConnectionAuthorizationDenied` (read-only) checks the
  GET 403 classification and pinned PowerShell outcome parity.
- `TestAccCloudCallDataConnectionCreateOutcome` (serial step `create-outcome`)
  sends the create and checks the 403 classification and PowerShell New parity.
  It is idempotent: it also accepts a token-bearing connection, in case the lab
  identity is later authorized and the singleton exists. It never needs a
  journal, because the denial changes nothing and a success cannot be removed.

Both sanitized 403 responses feed unit contract tests. Token-bearing success
bodies are covered by synthetic contract cases only. Successful Get/Create and
the already-exists Create outcome need separately authorized (delegated or
higher-privileged) credentials and remain unqualified.
