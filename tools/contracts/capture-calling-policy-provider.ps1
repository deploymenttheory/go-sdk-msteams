[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblies=@(
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'},
 @{path='netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll';sha256='477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af'}
)
foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned calling policy dependency.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
$env:MSTeamsContextInternal=$null
Import-Module $manifest -Force
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@((Join-Path $PSHOME 'System.Management.Automation.dll'))+@($assemblies|ForEach-Object{Join-Path $ModulePath $_.path})+@((Join-Path $PSHOME 'Newtonsoft.Json.dll'))
Add-Type -Path @((Join-Path $PSScriptRoot 'oracle/PolicyProviderCapture.cs'),(Join-Path $PSScriptRoot 'oracle/CallingPolicyProviderCapture.cs')) -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$cases=[Collections.Generic.List[object]]::new()
$properties=[Microsoft.Teams.Policy.Administration.Cmdlets.Core.TeamsCallingPolicy].GetProperties()|Where-Object Name -ne 'Identity'
$all=[ordered]@{}
foreach($p in $properties){
 $all[$p.Name]=if($p.PropertyType -eq [string]){if($p.Name -eq 'RecordingAndTranscriptionCustomMessageIdentifier'){'01234567-89AB-CDEF-0123-456789abcdef'}else{'Future-'+$p.Name}}elseif($p.PropertyType -eq [Nullable[bool]]){$false}elseif($p.Name -eq 'CallRecordingExpirationDays'){9999L}else{9007199254740993L}
}
$allJson=$all|ConvertTo-Json -Compress
foreach($c in @(
 @{name='read';op='read';input='{}';status=200;reply=('{"Identity":"Global","Future":false,'+$allJson.Substring(1));force=-1},
 @{name='list-empty';op='list';input='{}';status=200;reply='[]';force=-1},
 @{name='create-all';op='create';input=$allJson;status=201;reply='{}';force=-1},
 @{name='create-empty';op='create';input='{}';status=204;reply='';force=-1},
 @{name='patch-all';op='patch';input=$allJson;status=204;reply='';force=-1},
 @{name='patch-null';op='patch';input='{"Description":null,"AllowPrivateCalling":null,"CallingSpendUserLimit":null,"CallRecordingExpirationDays":null,"RecordingAndTranscriptionCustomMessageIdentifier":null}';status=204;reply='';force=-1},
 @{name='patch-empty';op='patch';input='{"Description":""}';status=204;reply='';force=-1},
 @{name='patch-force-false';op='patch';input='{"Description":""}';status=204;reply='';force=0},
 @{name='patch-force-true';op='patch';input='{"Description":""}';status=204;reply='';force=1},
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
 @{name='grant-group-empty';op='group';input='{}';status=204;reply='';force=-1}
)){
 $nullPolicyName=$c.name -notlike '*-empty'
 $capture=[TeamsContractOracle.CallingPolicyOracle]::Run($c.op,$c.input,$c.status,$c.reply,$c.force,$nullPolicyName)
 $cases.Add(@{name=$c.name;operation=$c.op;input=$c.input;status=$c.status;response=$c.reply;force_mode=$c.force;policy_name=$(if($nullPolicyName){$null}else{''});capture=$capture})
}
$models=[Collections.Generic.List[object]]::new()
function Add-Model([string]$name,[string]$inputJson){
 $models.Add(@{name=$name;input=$inputJson;observation=([TeamsContractOracle.CallingPolicyOracle]::Model($inputJson)|ConvertFrom-Json -AsHashtable)})
}
Add-Model 'constructor' 'constructor'
Add-Model 'empty-object' '{}'
Add-Model 'all-settings' $allJson
foreach($p in $properties){
 Add-Model ($p.Name+'-null') (@{$p.Name=$null}|ConvertTo-Json -Compress)
 if($p.PropertyType -eq [string]){
  Add-Model ($p.Name+'-empty') (@{$p.Name=''}|ConvertTo-Json -Compress)
  Add-Model ($p.Name+'-future') (@{$p.Name=('Future-'+$p.Name)}|ConvertTo-Json -Compress)
 }elseif($p.PropertyType -eq [Nullable[bool]]){
  Add-Model ($p.Name+'-false') (@{$p.Name=$false}|ConvertTo-Json -Compress)
  Add-Model ($p.Name+'-true') (@{$p.Name=$true}|ConvertTo-Json -Compress)
 }
}
foreach($n in @([long]::MinValue,-1L,0L,1L,9999L,10000L,9007199254740993L,[long]::MaxValue)){
 foreach($p in @('CallRecordingExpirationDays','CallingSpendUserLimit')){Add-Model ($p+'-'+$n) (@{$p=$n}|ConvertTo-Json -Compress)}
}
foreach($text in @('9223372036854775808','1.5','"9007199254740993"')){
 foreach($p in @('CallRecordingExpirationDays','CallingSpendUserLimit')){Add-Model ($p+'-json-'+$text) ('{"'+$p+'":'+$text+'}')}
}
Add-Model 'guid-uppercase' '{"RecordingAndTranscriptionCustomMessageIdentifier":"01234567-89AB-CDEF-0123-456789ABCDEF"}'
Add-Model 'guid-trailing-lf' '{"RecordingAndTranscriptionCustomMessageIdentifier":"01234567-89ab-cdef-0123-456789abcdef\n"}'
Add-Model 'guid-braces' '{"RecordingAndTranscriptionCustomMessageIdentifier":"{01234567-89ab-cdef-0123-456789abcdef}"}'
Add-Model 'long-description' (@{Description=('x'*2000)}|ConvertTo-Json -Compress)
Add-Model 'case-duplicate' '{"Description":"first","description":"second"}'
Add-Model 'duplicate-null' '{"AllowPrivateCalling":true,"AllowPrivateCalling":null}'
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-calling-policy-provider-and-models';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString()
 stage='Actual pinned direct HttpRequestHelper with synthetic InvocationInfo, credential-refusing host and memory sender. Client/version/host restored. MPACmdlet factory default supplied. Actual model constructors/setters/Newtonsoft projections for all 39 settings. No command constructor/lifecycle, public parameter binder, credentials, discovery or network. Direct compiled CRUD and grants only.'
 cases=$cases.ToArray();models=$models.ToArray()
}|ConvertTo-Json -Depth 25|Set-Content -LiteralPath $OutputPath -Encoding utf8
"calling policy provider cases: $($cases.Count); model observations: $($models.Count); settings: $(@($properties).Count)"
