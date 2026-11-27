[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference = 'Stop'
$hashes = @{
 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll' = '68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'
 'Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll' = 'a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'
}
foreach ($name in $hashes.Keys) {
 if ((Get-FileHash (Join-Path $ModulePath "netcoreapp3.1/$name")).Hash.ToLowerInvariant() -ne $hashes[$name]) { throw 'Pinned source differs' }
}
Import-Module (Join-Path $ModulePath 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
$names = @('Get-CsPrivacyConfiguration','Get-CsTeamsClientConfiguration','Get-CsTeamsEducationConfiguration','Get-CsOnlineVoicemailValidationConfiguration','Get-CsTeamsUpgradeConfiguration','Get-CsTeamsSipDevicesConfiguration','Get-CsTeamsMultiTenantOrganizationConfiguration','Get-CsTeamsTenantAbuseConfiguration','Get-CsTeamsMessagingConfiguration')
$declarations = @()
$rows = foreach ($name in $names) {
 $command = Get-Command $name
 if ($command.CommandType -ne 'Cmdlet' -or $null -eq $command.ImplementingType) { throw 'Expected a compiled public command' }
 $model = $command.ImplementingType.BaseType.GetGenericArguments()[0]
 $declarations += [pscustomobject]@{ command=$name; public_type=$command.ImplementingType.FullName; public_base=$command.ImplementingType.BaseType.ToString(); dto=$model.FullName }
 foreach ($property in $model.GetProperties() | Sort-Object Name) {
  $base = [Nullable]::GetUnderlyingType($property.PropertyType)
  if ($null -eq $base) { $base = $property.PropertyType }
  if ($base -eq [string]) {
   $kind = 'string'
   $tokens = @('null','true','false','0','37','9007199254740993','1.20','1E+03','-0','""','"text"','{}','[]')
  } elseif ($base -eq [bool]) {
   $kind = 'boolean'
   $tokens = @('null','true','false','0','1','-1','1.5','1e-400','1e400','""','"true"','" False "','" "','"\u0000true\u0000"','"1"','"yes"','{}','[]')
  } elseif ($base -eq [uint64]) {
   $kind = 'uint64'
   $tokens = @('null','0','1','-1','1.5','"1"','"1.5"','"18446744073709551615"','"18446744073709551616"','true','""','{}','[]')
  } elseif ($base.IsEnum -or $base -eq [int32]) {
   $kind = 'enum-or-int32'
   $tokens = @('null','0','37','-7','1.5','"37"','"Disabled"','"future"','true','""','2147483648','{}','[]')
  } else { continue }
  foreach ($token in $tokens) {
   $inputJSON = '{' + [Newtonsoft.Json.JsonConvert]::SerializeObject($property.Name) + ':' + $token + '}'
   $outputJSON = $null
   $failure = $null
   try {
    # Pure own DTO read path used by the PolicyRp provider. Never instantiate
    # a management cmdlet or invoke its lifecycle, auth or HTTP machinery.
    $dto = [Newtonsoft.Json.JsonConvert]::DeserializeObject($inputJSON,$model)
    $outputJSON = [Newtonsoft.Json.JsonConvert]::SerializeObject($property.GetValue($dto))
   } catch { $failure = $_.Exception.GetBaseException().GetType().FullName }
   [pscustomobject]@{ command=$name; dto=$model.FullName; field=$property.Name; declared_type=$property.PropertyType.ToString(); kind=$kind; input_json=$inputJSON; projected_json=$outputJSON; error=$failure }
  }
 }
}
[pscustomobject]@{
 scope='Actual pinned own DTO scalar deserialization. Mirrors the plain JsonConvert.DeserializeObject<TPolicy> call at PolicyRp HttpClientHelper SendAsync IL452; no management lifecycle, authentication or HTTP. String/bool/uint/enum projections are observed, not inferred across model setters.'
 hashes=$hashes
 reproducer_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant()
 declarations=$declarations
 observations=@($rows)
 management_lifecycle_invoked=$false
 authentication_or_http=$false
} | ConvertTo-Json -Depth 6
