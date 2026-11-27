[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference='Stop'
$hashes=@{
 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll'='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'
 'Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll'='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'
}
foreach($name in $hashes.Keys){if((Get-FileHash(Join-Path $ModulePath "netcoreapp3.1/$name")).Hash.ToLowerInvariant() -ne $hashes[$name]){throw 'Pinned source differs'}}
Import-Module (Join-Path $ModulePath 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
$rows=@();$declarations=@();$public=@()
$normalizer=[AppDomain]::CurrentDomain.GetAssemblies()|ForEach-Object{$_.GetType('Microsoft.Teams.Policy.Administration.Cmdlets.Providers.Helpers.ConfigurationNameNormalization`1',$false)}|Where-Object{$_}|Select-Object -First 1
foreach($name in @('Get-CsTeamsGuestMeetingConfiguration','Get-CsTeamsMeetingConfiguration')){
 $get=Get-Command $name
 if($get.CommandType -ne 'Cmdlet'){throw 'Expected a public compiled Get'}
 $model=$get.ImplementingType.BaseType.GetGenericArguments()[0]
 $normalized=$normalizer.MakeGenericType($model).GetMethod('GetNormalizedNameForPolicyRP',[Reflection.BindingFlags]'Public,NonPublic,Static').Invoke($null,@())
 foreach($cmd in @($get,(Get-Command ($name -replace '^Get-','Set-')))){
  if($cmd.CommandType -ne 'Cmdlet'){throw 'Expected a public compiled Set'}
  $public+=[pscustomobject]@{command=$cmd.Name;kind=$cmd.CommandType.ToString();implementing_type=$cmd.ImplementingType.FullName;base_type=$cmd.ImplementingType.BaseType.ToString();normalized_resource=$normalized;sets=@($cmd.ParameterSets|ForEach-Object{[pscustomobject]@{name=$_.Name;default=$_.IsDefault;parameters=@($_.Parameters|ForEach-Object{[pscustomobject]@{name=$_.Name;mandatory=$_.IsMandatory;type=$_.ParameterType.ToString()}})}});effective_parameters=@($cmd.Parameters.Values|Sort-Object Name|ForEach-Object{[pscustomobject]@{name=$_.Name;type=$_.ParameterType.ToString();validation=@($_.Attributes|Where-Object{$_ -isnot [System.Management.Automation.ParameterAttribute]}|ForEach-Object{$_.GetType().FullName})}})}
 }
 $entries=@([pscustomobject]@{kind='configuration';type=$model})
 if($name -eq 'Get-CsTeamsMeetingConfiguration'){$entries+=[pscustomobject]@{kind='context';type=$model.GetProperty('PublishedEntraAuthenticationContexts').PropertyType.GetGenericArguments()[0]}}
 foreach($entry in $entries){
  $kind=$entry.kind;$type=$entry.type
  $declarations+=[pscustomobject]@{command=$name;kind=$kind;dto=$type.FullName;properties=@($type.GetProperties()|Sort-Object Name|ForEach-Object{[pscustomobject]@{name=$_.Name;type=$_.PropertyType.ToString();setter=$_.CanWrite}})}
  $inputs=@('{}')
  foreach($property in $type.GetProperties()|Sort-Object Name){
   $base=[Nullable]::GetUnderlyingType($property.PropertyType);if($null -eq $base){$base=$property.PropertyType}
   $tokens=@()
   if($base -eq [string]){$tokens=@('null','true','false','0','37','9007199254740993','1.20','1E+03','-0','""','"text"','"c1"','"c99"','"c0"','"c100"','"c1\n"','"c1\r\n"','{}','[]')}
   elseif($base -eq [bool]){$tokens=@('null','true','false','0','1','-1','1.5','1e-400','1e400','""','"true"','" False "','" "','"\u0000true\u0000"','"1"','{}','[]')}
   elseif($base -eq [uint32]){$tokens=@('null','0','1','-1','0.5','1.5','2.5','-0.5','4294967295','4294967296','4294967295.0','1e-400','1e400','true','false','""','"1"','"1.5"','"+1"','"-0"','"1\u0000"','"4294967295"','"4294967296"','{}','[]')}
   elseif($base.IsGenericType -and $base.GetGenericTypeDefinition() -eq [System.Collections.Generic.List``1]){$tokens=@('null','[]','[null]','[{}]','[{"Id":"c1"}]','[{"Id":"invalid"}]','true','37','"ignored"','{}')}
   foreach($token in $tokens){$inputs+=('{'+[Newtonsoft.Json.JsonConvert]::SerializeObject($property.Name)+':'+$token+'}')}
   if($base.IsGenericType -and $base.GetGenericTypeDefinition() -eq [System.Collections.Generic.List``1]){
    foreach($second in @('null','[null]','true')){$inputs+=('{"'+$property.Name+'":[{}],"'+$property.Name.ToLowerInvariant()+'":'+$second+'}')}
   }
  }
  foreach($body in $inputs){
   $projected=$null;$failure=$null
   try{$dto=[Newtonsoft.Json.JsonConvert]::DeserializeObject($body,$type);$projected=[Newtonsoft.Json.JsonConvert]::SerializeObject($dto)}catch{$failure=$_.Exception.GetBaseException().GetType().FullName}
   $rows+=[pscustomobject]@{command=$name;kind=$kind;dto=$type.FullName;input_json=$body;projected_json=$projected;error=$failure}
  }
 }
}
[pscustomobject]@{scope='Own last two compiled POL91 configuration/Entra-context DTO received-JSON profiles. Pure data-model reflection/deserialization only, no management lifecycle, authentication or HTTP.';hashes=$hashes;reproducer_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant();public_commands=$public;declarations=$declarations;observations=$rows;management_lifecycle_invoked=$false;authentication_or_http=$false}|ConvertTo-Json -Depth 10
