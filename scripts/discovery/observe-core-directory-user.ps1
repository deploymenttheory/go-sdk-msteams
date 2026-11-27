param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference='Stop'
$dll=Join-Path $ModulePath 'netcoreapp3.1/Microsoft.Teams.PowerShell.TeamsCmdlets.dll'
$expected='3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24'
if((Get-FileHash $dll -Algorithm SHA256).Hash.ToLowerInvariant() -ne $expected){throw 'Pinned TeamsCmdlets hash mismatch'}
Import-Module (Join-Path $ModulePath 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
$t=[Microsoft.Teams.PowerShell.TeamsCmdlets.Model.AadUser]
$fields=[ordered]@{id='Id';displayName='DisplayName';userPrincipalName='UserPrincipalName';userType='UserType';mail='Mail'}
$metadata=@($fields.Values|ForEach-Object{$p=$t.GetProperty($_);if($p.PropertyType -ne [string]){throw 'Unexpected property type'};[ordered]@{name=$p.Name;type=$p.PropertyType.FullName;setter_il=[Convert]::ToHexString($p.SetMethod.GetMethodBody().GetILAsByteArray());getter_il=[Convert]::ToHexString($p.GetMethod.GetMethodBody().GetILAsByteArray())}})
$rows=@()
foreach($field in $fields.Keys){
 foreach($token in @('null','""','"opaque-user"','" \u00e9 "','true','false','37','9007199254740993','1.5','1e2','1e400','-0','[]','{}')){
  $inputJson='{"'+$field+'":'+$token+'}'
  try{$value=[Newtonsoft.Json.JsonConvert]::DeserializeObject($inputJson,$t);$rows+= [ordered]@{field=$field;input_json=$inputJson;value=$value.($fields[$field]);exception=$null}}
  catch{$rows+=[ordered]@{field=$field;input_json=$inputJson;value=$null;exception=$_.Exception.GetBaseException().GetType().FullName}}
 }
}
[ordered]@{schema_version=1;scope='Pure actual AadUser JsonConvert data operations; no management lifecycle, auth or HTTP';powershell_version=$PSVersionTable.PSVersion.ToString();assembly_sha256=$expected;producer_sha256=(Get-FileHash $PSCommandPath -Algorithm SHA256).Hash.ToLowerInvariant();type=$t.FullName;properties=$metadata;observations=$rows}|ConvertTo-Json -Depth 12
