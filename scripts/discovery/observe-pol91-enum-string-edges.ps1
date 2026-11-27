[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference='Stop'
$hashes=@{
 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll'='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'
 'Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll'='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'
}
foreach($name in $hashes.Keys){if((Get-FileHash(Join-Path $ModulePath "netcoreapp3.1/$name")).Hash.ToLowerInvariant() -ne $hashes[$name]){throw 'Pinned source differs'}}
Import-Module (Join-Path $ModulePath 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
$command=Get-Command Get-CsTeamsMessagingConfiguration
$model=$command.ImplementingType.BaseType.GetGenericArguments()[0]
$tokens=@('"1\u0000"','"37\u0000"','"\u00a01"','"1\u00a0"','"Enabled\u0000"','"\u0000Enabled"','"Enabled, Disabled"','"Enabled,0"','"1,Enabled"','"37,Disabled"')
$rows=@(foreach($token in $tokens){
 $body='{"Storyline":'+$token+'}'
 $value=$null;$errorType=$null
 try{$dto=[Newtonsoft.Json.JsonConvert]::DeserializeObject($body,$model);$value=[Newtonsoft.Json.JsonConvert]::SerializeObject($dto.Storyline)}catch{$errorType=$_.Exception.GetBaseException().GetType().FullName}
 [pscustomobject]@{command=$command.Name;dto=$model.FullName;field='Storyline';kind='enum-string-edges';input_json=$body;projected_json=$value;error=$errorType}
})
[pscustomobject]@{scope='Actual own pinned Messaging DTO enum-string deserialization. Numeric and named tokens retain distinct grammar. No management lifecycle, authentication or HTTP.';hashes=$hashes;public_type=$command.ImplementingType.FullName;dto=$model.FullName;reproducer_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant();observations=$rows;management_lifecycle_invoked=$false;authentication_or_http=$false}|ConvertTo-Json -Depth 6
