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
- New was not sent live. The connection is a tenant singleton with no remove
  cmdlet, so a successful call cannot be undone, and the live probe was not
  authorized in this session.

Acceptance: `TestAccCloudCallDataConnectionAuthorizationDenied` (read-only)
checks the 403 classification and pinned PowerShell outcome parity; its
sanitized response feeds the unit contract test. Token-bearing success bodies
and every Create outcome are covered by synthetic contract cases only.
Successful Get/Create need separately authorized credentials and remain
unqualified.
