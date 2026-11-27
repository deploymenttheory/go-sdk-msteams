[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference='Stop'
$hashes=@{
 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll'='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'
 'Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll'='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'
}
foreach($name in $hashes.Keys){if((Get-FileHash(Join-Path $ModulePath "netcoreapp3.1/$name")).Hash.ToLowerInvariant() -ne $hashes[$name]){throw 'Pinned source differs'}}
Import-Module (Join-Path $ModulePath 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
$declarations=@();$rows=@();$public=@()
$normalizer=[AppDomain]::CurrentDomain.GetAssemblies()|ForEach-Object{$_.GetType('Microsoft.Teams.Policy.Administration.Cmdlets.Providers.Helpers.ConfigurationNameNormalization`1',$false)}|Where-Object{$_}|Select-Object -First 1
foreach($name in @('Get-CsTeamsFirstPartyMeetingTemplateConfiguration','Get-CsTeamsMeetingTemplateConfiguration')){
 $command=Get-Command $name
 if($command.CommandType -ne 'Cmdlet'){throw 'Expected compiled public command'}
 $model=$command.ImplementingType.BaseType.GetGenericArguments()[0]
 $normalized=$normalizer.MakeGenericType($model).GetMethod('GetNormalizedNameForPolicyRP',[Reflection.BindingFlags]'Public,NonPublic,Static').Invoke($null,@())
 $commands=@($command)
 if($name -eq 'Get-CsTeamsFirstPartyMeetingTemplateConfiguration'){$commands+=Get-Command Set-CsTeamsFirstPartyMeetingTemplateConfiguration}
 foreach($c in $commands){
  $public+=[pscustomobject]@{command=$c.Name;kind=$c.CommandType.ToString();implementing_type=$c.ImplementingType.FullName;base_type=$c.ImplementingType.BaseType.ToString();normalized_resource=$normalized;sets=@($c.ParameterSets|ForEach-Object{[pscustomobject]@{name=$_.Name;default=$_.IsDefault;parameters=@($_.Parameters|ForEach-Object{[pscustomobject]@{name=$_.Name;mandatory=$_.IsMandatory;position=$_.Position;type=$_.ParameterType.ToString()}})}});effective_parameters=@($c.Parameters.Values|Sort-Object Name|ForEach-Object{[pscustomobject]@{name=$_.Name;type=$_.ParameterType.ToString();validation_attributes=@($_.Attributes|Where-Object{$_ -isnot [System.Management.Automation.ParameterAttribute]}|ForEach-Object{$_.GetType().FullName})}})}
 }
 $template=$model.GetProperty('TeamsMeetingTemplates').PropertyType.GetGenericArguments()[0]
 $option=$template.GetProperty('TeamsMeetingOptions').PropertyType.GetGenericArguments()[0]
 foreach($entry in @(@('configuration',$model),@('template',$template),@('option',$option))){
  $kind=$entry[0];$type=$entry[1]
  $declarations+=[pscustomobject]@{command=$name;kind=$kind;dto=$type.FullName;properties=@($type.GetProperties()|Sort-Object Name|ForEach-Object{[pscustomobject]@{name=$_.Name;type=$_.PropertyType.ToString();setter=$_.CanWrite}})}
  $inputs=@('{}')
  foreach($p in $type.GetProperties()|Sort-Object Name){
   $base=[Nullable]::GetUnderlyingType($p.PropertyType);if($null -eq $base){$base=$p.PropertyType}
   $tokens=@()
   if($base -eq [string]){$tokens=@('null','true','37','9007199254740993','1.20','1E+03','-0','""','"text"','{}','[]')}
   elseif($base -eq [bool]){$tokens=@('null','true','1','"False"','""','" "','{}','[]')}
   elseif($base.IsGenericType -and $base.GetGenericTypeDefinition() -eq [System.Collections.Generic.List``1]){$tokens=@('null','[]','[null]','[{}]','true','37','"ignored"','{}')}
   foreach($token in $tokens){$inputs+=('{'+[Newtonsoft.Json.JsonConvert]::SerializeObject($p.Name)+':'+$token+'}')}
   if($tokens.Count -gt 0 -and $base.IsGenericType){
    $inputs+=('{"'+$p.Name+'":[{}],"'+$p.Name.ToLowerInvariant()+'":null}')
    $inputs+=('{"'+$p.Name+'":[{}],"'+$p.Name.ToLowerInvariant()+'":[null]}')
    $inputs+=('{"'+$p.Name+'":[{}],"'+$p.Name.ToLowerInvariant()+'":true}')
   }
  }
  foreach($body in $inputs){
   $projection=$null;$failure=$null
   try{$dto=[Newtonsoft.Json.JsonConvert]::DeserializeObject($body,$type);$projection=[Newtonsoft.Json.JsonConvert]::SerializeObject($dto)}catch{$failure=$_.Exception.GetBaseException().GetType().FullName}
   $rows+=[pscustomobject]@{command=$name;kind=$kind;dto=$type.FullName;input_json=$body;projected_json=$projection;error=$failure}
  }
 }
}
[pscustomobject]@{scope='Own compiled meeting-template DTO read deserialization only; actual nested constructor/setter/list behavior. No management lifecycle, authentication or HTTP.';hashes=$hashes;reproducer_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant();public_commands=$public;declarations=$declarations;observations=$rows;management_lifecycle_invoked=$false;authentication_or_http=$false}|ConvertTo-Json -Depth 10
