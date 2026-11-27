[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference='Stop'
$hashes=@{
 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll'='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'
 'Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll'='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'
}
foreach($name in $hashes.Keys){if((Get-FileHash(Join-Path $ModulePath "netcoreapp3.1/$name")).Hash.ToLowerInvariant() -ne $hashes[$name]){throw 'Pinned source differs'}}
Import-Module (Join-Path $ModulePath 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
$cases=@(
 @{command='Get-CsTeamsTenantAbuseConfiguration';field='LowSeatLimit';kind='uint64-boundaries';tokens=@('0.5','2.5','3.5','-0.5','-0.1','-0.5001','1e-400','1e400','18446744073709551615','18446744073709551616','9007199254740993','18446744073709551615.0','false','" 1 "','"+1"','"-0"','"0.0"','"1e3"','" "','"++1"','"+-0"','"1\u0000"','"\u00001"','"\u00a01"')},
 @{command='Get-CsTeamsMessagingConfiguration';field='Storyline';kind='enum-boundaries';tokens=@('1.0','1e3','-2147483649','4294967296','9223372036854775807','18446744073709551615','18446744073709551616','"2147483648"','"-2147483649"','false','" "','" enabled "','"1,0"','"Enabled,Disabled"')}
)
foreach($entry in @(@('Get-CsTeamsTenantAbuseConfiguration','BlockedTenants'),@('Get-CsTeamsMessagingConfiguration','FileTypeBlockList'))){
 $cases+=@{command=$entry[0];field=$entry[1];kind='nullable-string-list';tokens=@('null','[]','[null,7,false,9007199254740993,1.20,1E+03,-0]','[{}]','[[]]','true','"entry"')}
}
$uris=@('https://n.invalid/','https://abc/',('https://abc/'+"`n"),('https://abc/'+"`r`n"),'https://abc/\','https://abc/\n',('https://abc/'+"`n"+'rest'),'HTTPS://n.invalid/','','http://n.invalid/')
$cases+=@{command='Get-CsOnlineVoicemailValidationConfiguration';field='AudioFileValidationUri';kind='uri-setter';tokens=@($uris|ForEach-Object{[Newtonsoft.Json.JsonConvert]::SerializeObject($_)})}
$rows=@(foreach($c in $cases){
 $command=Get-Command $c.command
 $model=$command.ImplementingType.BaseType.GetGenericArguments()[0]
 $property=$model.GetProperty($c.field)
 foreach($token in $c.tokens){
  $body='{'+[Newtonsoft.Json.JsonConvert]::SerializeObject($c.field)+':'+$token+'}'
  $value=$null;$failure=$null
  try{$dto=[Newtonsoft.Json.JsonConvert]::DeserializeObject($body,$model);$value=[Newtonsoft.Json.JsonConvert]::SerializeObject($property.GetValue($dto))}catch{$failure=$_.Exception.GetBaseException().GetType().FullName}
  [pscustomobject]@{command=$c.command;public_type=$command.ImplementingType.FullName;dto=$model.FullName;field=$c.field;kind=$c.kind;input_json=$body;projected_json=$value;error=$failure}
 }
 if($c.kind -eq 'nullable-string-list'){
  foreach($body in @('{}',('{'+ '"'+$c.field+'":["a"],"'+$c.field.ToLowerInvariant()+'":null}'),('{'+ '"'+$c.field+'":["a"],"'+$c.field.ToLowerInvariant()+'":[null,7,false]}'))){
   $value=$null;$failure=$null
   try{$dto=[Newtonsoft.Json.JsonConvert]::DeserializeObject($body,$model);$value=[Newtonsoft.Json.JsonConvert]::SerializeObject($property.GetValue($dto))}catch{$failure=$_.Exception.GetBaseException().GetType().FullName}
   [pscustomobject]@{command=$c.command;public_type=$command.ImplementingType.FullName;dto=$model.FullName;field=$c.field;kind=$c.kind;input_json=$body;projected_json=$value;error=$failure}
  }
 }
})
[pscustomobject]@{scope='Actual own pinned DTO numeric boundaries, nullable string-list population and URI setter behavior. Pure data-model/JsonConvert execution only; no management lifecycle, authentication or HTTP.';hashes=$hashes;reproducer_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant();observations=$rows;management_lifecycle_invoked=$false;authentication_or_http=$false}|ConvertTo-Json -Depth 6
