[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblies=@(
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'},
 @{path='netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll';sha256='477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af'}
)
foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned files policy dependency.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
$env:MSTeamsContextInternal=$null
Import-Module $manifest -Force
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@((Join-Path $PSHOME 'System.Management.Automation.dll'))+@($assemblies|ForEach-Object{Join-Path $ModulePath $_.path})+@((Join-Path $PSHOME 'Newtonsoft.Json.dll'))
Add-Type -Path @((Join-Path $PSScriptRoot 'oracle/PolicyProviderCapture.cs'),(Join-Path $PSScriptRoot 'oracle/FilesPolicyProviderCapture.cs')) -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$exports=@('Get','New','Set','Remove','Grant')|ForEach-Object{
 $command=Get-Command "$_-CsTeamsFilesPolicy" -Module MicrosoftTeams -ErrorAction Stop
 if($command.CommandType -ne 'Cmdlet' -or $command.ImplementingType.Assembly -ne [Microsoft.Teams.Policy.Administration.Cmdlets.Core.TeamsFilesPolicy].Assembly){throw 'Unexpected actual public files export.'}
 @{name=$command.Name;type=$command.ImplementingType.FullName}
}
$properties=@([Microsoft.Teams.Policy.Administration.Cmdlets.Core.TeamsFilesPolicy].GetProperties()|Where-Object Name -ne 'Identity')
if($properties.Count -ne 4 -or @($properties|Where-Object{$_.PropertyType -ne [string]}).Count){throw 'Files setting shape differs.'}
$all=[ordered]@{};$nulls=[ordered]@{};$empty=[ordered]@{}
foreach($p in $properties){
 $nulls[$p.Name]=$null
 if($p.PropertyType -eq [string]){$all[$p.Name]='Future-'+$p.Name;$empty[$p.Name]=''}else{$all[$p.Name]=$false;$empty[$p.Name]=$true}
}
$allJson=$all|ConvertTo-Json -Compress
$nullJson=$nulls|ConvertTo-Json -Compress
$emptyJson=$empty|ConvertTo-Json -Compress
$cases=[Collections.Generic.List[object]]::new()
foreach($c in @(
 @{name='read';op='read';input='{}';status=200;reply=('{"Identity":"Global","Future":false,'+$allJson.Substring(1));force=-1},
 @{name='list-empty';op='list';input='{}';status=200;reply='[]';force=-1},
 @{name='create-all';op='create';input=$allJson;status=201;reply='{}';force=-1},
 @{name='create-empty';op='create';input='{}';status=204;reply='';force=-1},
 @{name='patch-all';op='patch';input=$allJson;status=204;reply='';force=-1},
 @{name='patch-null';op='patch';input=$nullJson;status=204;reply='';force=-1},
 @{name='patch-empty';op='patch';input=$emptyJson;status=204;reply='';force=-1},
 @{name='patch-force-false';op='patch';input='{"NativeFileEntryPoints":""}';status=204;reply='';force=0},
 @{name='patch-force-true';op='patch';input='{"NativeFileEntryPoints":""}';status=204;reply='';force=1},
 @{name='delete';op='delete';input='{}';status=204;reply='';force=-1},
 @{name='delete-force-false';op='delete';input='{}';status=204;reply='';force=0},
 @{name='read-empty';op='read';input='{}';status=204;reply='';force=-1},
 @{name='read-null';op='read';input='{}';status=200;reply='null';force=-1},
 @{name='list-envelope';op='list';input='{}';status=200;reply='{"value":[]}';force=-1},
 @{name='read-denied';op='read';input='{}';status=403;reply='{"Message":"fixture-denied"}';force=-1},
 @{name='grant-user-null';op='user';input='{}';status=204;reply='';force=-1},
 @{name='grant-tenant-null';op='tenant';input='{}';status=204;reply='';force=0},
 @{name='grant-group-null';op='group';input='{}';status=204;reply='';force=-1},
 @{name='grant-user-empty';op='user';input='{}';status=204;reply='';force=-1},
 @{name='grant-tenant-empty';op='tenant';input='{}';status=204;reply='';force=1},
 @{name='grant-group-empty';op='group';input='{}';status=204;reply='';force=-1},
 @{name='grant-group-rank-zero';op='group';input='{}';status=204;reply='';force=-1;rank=0},
 @{name='grant-group-rank-two';op='group';input='{}';status=204;reply='';force=-1;rank=2},
 @{name='grant-group-accepted';op='group';input='{}';status=202;reply='{"OperationId":"fixture-operation"}';force=-1;rank=2}
)){
 $nullName=$c.name -notlike '*-empty';$rank=if($c.ContainsKey('rank')){$c.rank}else{-1}
 $capture=[TeamsContractOracle.FilesPolicyOracle]::Run($c.op,$c.input,$c.status,$c.reply,$c.force,$nullName,$rank)
 if($capture.TokenRequests -ne 0 -or $capture.Sends -ne 1){throw 'Files provider crossed the offline boundary.'}
 $cases.Add(@{name=$c.name;operation=$c.op;input=$c.input;status=$c.status;response=$c.reply;force_mode=$c.force;policy_name=$(if($nullName){$null}else{''});rank=$rank;capture=$capture})
}
$models=[Collections.Generic.List[object]]::new()
function Add-Model([string]$name,[string]$inputJson){
 $models.Add(@{name=$name;input=$inputJson;observation=([TeamsContractOracle.FilesPolicyOracle]::Model($inputJson)|ConvertFrom-Json -AsHashtable)})
}
Add-Model 'constructor' 'constructor'
Add-Model 'empty-object' '{}'
Add-Model 'all-settings' $allJson
foreach($p in $properties){
 Add-Model ($p.Name+'-null') (@{$p.Name=$null}|ConvertTo-Json -Compress)
 if($p.PropertyType -eq [string]){
  Add-Model ($p.Name+'-empty') (@{$p.Name=''}|ConvertTo-Json -Compress)
  Add-Model ($p.Name+'-future') (@{$p.Name=('Future-'+$p.Name)}|ConvertTo-Json -Compress)
 }else{
  Add-Model ($p.Name+'-false') (@{$p.Name=$false}|ConvertTo-Json -Compress)
  Add-Model ($p.Name+'-true') (@{$p.Name=$true}|ConvertTo-Json -Compress)
 }
}
Add-Model 'long-native-entry-points' (@{NativeFileEntryPoints=('x'*2000)}|ConvertTo-Json -Compress)
Add-Model 'case-duplicate' '{"NativeFileEntryPoints":"first","nativefileentrypoints":"second"}'
Add-Model 'duplicate-null' '{"SPChannelFilesTab":"first","spchannelfilestab":null}'
Add-Model 'invalid-string-object' '{"DefaultFileUploadAppId":{}}'
Add-Model 'invalid-string-array' '{"FileSharingInChatswithExternalUsers":[]}'
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-files-policy-provider-and-models';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString();exports=$exports
 stage='Actual public exports verified as Core cmdlets. Actual pinned direct HttpRequestHelper with synthetic InvocationInfo, credential-refusing host and memory sender. Client/version/host restored. MPACmdlet factory default supplied. Actual model constructors/setters/Newtonsoft projections for all four string settings. No original command constructor/lifecycle, public binder, scope authorization, credentials, discovery or network. Direct compiled CRUD and grants only.'
 cases=$cases.ToArray();models=$models.ToArray()
}|ConvertTo-Json -Depth 25|Set-Content -LiteralPath $OutputPath -Encoding utf8
"files provider cases: $($cases.Count); models: $($models.Count); settings: $($properties.Count)"
