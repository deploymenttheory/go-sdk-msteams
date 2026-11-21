[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblies=@(
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'},
 @{path='netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll';sha256='477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af'}
)
foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned education assignments dependency.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
$env:MSTeamsContextInternal=$null
Import-Module $manifest -Force
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@((Join-Path $PSHOME 'System.Management.Automation.dll'))+@($assemblies|ForEach-Object{Join-Path $ModulePath $_.path})+@((Join-Path $PSHOME 'Newtonsoft.Json.dll'))
Add-Type -Path @((Join-Path $PSScriptRoot 'oracle/PolicyProviderCapture.cs'),(Join-Path $PSScriptRoot 'oracle/EducationAssignmentsProviderCapture.cs')) -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$exports=@('Get-CsTeamsEducationAssignmentsAppPolicy','Set-CsTeamsEducationAssignmentsAppPolicy')|ForEach-Object{
 $command=Get-Command $_ -Module MicrosoftTeams -ErrorAction Stop
 if($command.CommandType -ne 'Cmdlet' -or $command.ImplementingType.Assembly -ne [Microsoft.Teams.Policy.Administration.Cmdlets.Core.TeamsEducationAssignmentsAppPolicy].Assembly){throw 'Unexpected actual export type.'}
 @{name=$command.Name;type=$command.ImplementingType.FullName}
}
$cases=[Collections.Generic.List[object]]::new()
$properties=@([Microsoft.Teams.Policy.Administration.Cmdlets.Core.TeamsEducationAssignmentsAppPolicy].GetProperties()|Where-Object Name -ne 'Identity')
if($properties.Count -ne 5 -or @($properties|Where-Object PropertyType -ne ([string])).Count){throw 'Baseline setting shape differs.'}
$allJson='{"ParentDigestEnabledType":"FutureParent","MakeCodeEnabledType":"FutureCode","TurnItInEnabledType":"FutureTurnItIn","TurnItInApiUrl":"HTTPS://fixture.invalid","TurnItInApiKey":"offline-fixture-key"}'
foreach($c in @(
 @{name='read';op='read';input='{}';status=200;reply=('{"Identity":"Global","Future":false,'+$allJson.Substring(1));force=-1},
 @{name='list-empty';op='list';input='{}';status=200;reply='[]';force=-1},
 @{name='patch-all';op='patch';input=$allJson;status=204;reply='';force=-1},
 @{name='patch-null';op='patch';input='{"ParentDigestEnabledType":null,"MakeCodeEnabledType":null,"TurnItInEnabledType":null,"TurnItInApiUrl":null,"TurnItInApiKey":null}';status=204;reply='';force=-1},
 @{name='patch-empty';op='patch';input='{"ParentDigestEnabledType":"","MakeCodeEnabledType":"","TurnItInEnabledType":"","TurnItInApiKey":""}';status=204;reply='';force=-1},
 @{name='patch-force-false';op='patch';input='{"TurnItInApiKey":""}';status=204;reply='';force=0},
 @{name='patch-force-true';op='patch';input='{"TurnItInApiKey":""}';status=204;reply='';force=1},
 @{name='patch-url-final-lf';op='patch';input='{"TurnItInApiUrl":"https://opaque\n"}';status=204;reply='';force=-1},
 @{name='patch-url-empty';op='patch';input='{"TurnItInApiUrl":""}';status=204;reply='';force=-1},
 @{name='patch-url-http';op='patch';input='{"TurnItInApiUrl":"http://fixture.invalid"}';status=204;reply='';force=-1},
 @{name='read-empty';op='read';input='{}';status=204;reply='';force=-1},
 @{name='read-null';op='read';input='{}';status=200;reply='null';force=-1},
 @{name='list-envelope';op='list';input='{}';status=200;reply='{"value":[]}';force=-1},
 @{name='read-denied';op='read';input='{}';status=403;reply='{"Message":"fixture-denied"}';force=-1}
)){
 $capture=[TeamsContractOracle.EducationAssignmentsOracle]::Run($c.op,$c.input,$c.status,$c.reply,$c.force)
 if($capture.TokenRequests -ne 0 -or $capture.Sends -gt 1){throw 'Capture crossed the offline boundary.'}
 $cases.Add(@{name=$c.name;operation=$c.op;input=$c.input;status=$c.status;response=$c.reply;force_mode=$c.force;capture=$capture})
}
$models=[Collections.Generic.List[object]]::new()
function Add-Model([string]$name,[string]$inputJson){
 $models.Add(@{name=$name;input=$inputJson;observation=([TeamsContractOracle.EducationAssignmentsOracle]::Model($inputJson)|ConvertFrom-Json -AsHashtable)})
}
Add-Model 'constructor' 'constructor'
Add-Model 'empty-object' '{}'
Add-Model 'all-settings' $allJson
foreach($p in $properties){
 Add-Model ($p.Name+'-null') (@{$p.Name=$null}|ConvertTo-Json -Compress)
 Add-Model ($p.Name+'-empty') (@{$p.Name=''}|ConvertTo-Json -Compress)
}
foreach($entry in @(
 @{name='url-http';value='http://fixture.invalid'},
 @{name='url-upper';value='HTTPS://fixture.invalid'},
 @{name='url-opaque';value='https://opaque'},
 @{name='url-space';value='https:// '},
 @{name='url-prefix-space';value=' https://fixture.invalid'},
 @{name='url-final-lf';value="https://fixture.invalid`n"},
 @{name='url-middle-lf';value="https://fixture`n.invalid"},
 @{name='url-empty-host';value='https://'}
)){Add-Model $entry.name (@{TurnItInApiUrl=$entry.value}|ConvertTo-Json -Compress)}
Add-Model 'long-api-key' (@{TurnItInApiKey=('x'*2000)}|ConvertTo-Json -Compress)
Add-Model 'case-duplicate' '{"ParentDigestEnabledType":"first","parentdigestenabledtype":"second"}'
Add-Model 'invalid-then-valid-url' '{"TurnItInApiUrl":"bad","turnitinapiurl":"https://fixture.invalid"}'
Add-Model 'invalid-string-array' '{"TurnItInApiKey":[]}'
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-education-assignments-provider-and-models';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString();exports=$exports
 stage='Actual pinned public export types and direct HttpRequestHelper with synthetic InvocationInfo, credential-refusing host and memory sender. Client/version/host restored. MPACmdlet factory default supplied. Actual model constructors/setters/Newtonsoft projections for all five settings. No command constructor/lifecycle, public parameter binder, credentials, discovery or network. Compiled read/list/PATCH only; unexported Core New/Remove/Grant classes do not establish public operations.'
 cases=$cases.ToArray();models=$models.ToArray()
}|ConvertTo-Json -Depth 25|Set-Content -LiteralPath $OutputPath -Encoding utf8
"education assignments provider cases: $($cases.Count); models: $($models.Count); settings: $($properties.Count)"
