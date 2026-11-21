[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblies=@(
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'},
 @{path='netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll';sha256='477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af'}
)
foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned call hold dependency.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
$env:MSTeamsContextInternal=$null
Import-Module $manifest -Force
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@((Join-Path $PSHOME 'System.Management.Automation.dll'))+@($assemblies|ForEach-Object{Join-Path $ModulePath $_.path})+@((Join-Path $PSHOME 'Newtonsoft.Json.dll'))
Add-Type -Path @((Join-Path $PSScriptRoot 'oracle/PolicyProviderCapture.cs'),(Join-Path $PSScriptRoot 'oracle/CallHoldProviderCapture.cs')) -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$cases=[Collections.Generic.List[object]]::new()
foreach($c in @(
 @{name='read';op='read';fields='all';status=200;reply='{"Identity":"Global","Description":"","AudioFileId":"future-file","StreamingSourceUrl":"future://fixture","StreamingSourceAuthType":null,"Future":false}';force=-1},
 @{name='list-empty';op='list';fields='all';status=200;reply='[]';force=-1},
 @{name='create-all';op='create';fields='all';status=201;reply='{"Identity":"Global"}';force=-1},
 @{name='create-empty';op='create';fields='empty';status=204;reply='';force=-1},
 @{name='patch-all';op='patch';fields='all';status=204;reply='';force=-1},
 @{name='patch-null';op='patch';fields='description-null';status=204;reply='';force=-1},
 @{name='patch-empty';op='patch';fields='description-empty';status=204;reply='';force=-1},
 @{name='patch-force-false';op='patch';fields='all';status=204;reply='';force=0},
 @{name='patch-force-true';op='patch';fields='all';status=204;reply='';force=1},
 @{name='delete';op='delete';fields='empty';status=204;reply='';force=-1},
 @{name='delete-force-false';op='delete';fields='empty';status=204;reply='';force=0},
 @{name='read-empty';op='read';fields='all';status=204;reply='';force=-1},
 @{name='read-null';op='read';fields='all';status=200;reply='null';force=-1},
 @{name='list-envelope';op='list';fields='all';status=200;reply='{"value":[]}';force=-1},
 @{name='read-denied';op='read';fields='all';status=403;reply='{"Message":"fixture-denied"}';force=-1}
)){
 $capture=[TeamsContractOracle.CallHoldOracle]::Run($c.op,$c.fields,$c.status,$c.reply,$c.force)
 $cases.Add(@{name=$c.name;operation=$c.op;fields=$c.fields;status=$c.status;response=$c.reply;force_mode=$c.force;capture=$capture})
}
$models=[Collections.Generic.List[object]]::new()
foreach($inputJson in @('constructor','long-description','{}','{"Identity":null,"Description":null,"AudioFileId":null,"StreamingSourceUrl":null,"StreamingSourceAuthType":null}','{"Identity":"Global","Description":"","AudioFileId":"future-file","StreamingSourceUrl":"future://fixture","StreamingSourceAuthType":"FutureAuth","Future":false}','{"Description":"first","description":"second"}','{"Description":"first","Description":null}','{"AudioFileId":{}}','{"StreamingSourceUrl":[]}','{"StreamingSourceAuthType":{}}')){
 $models.Add(@{input=$inputJson;observation=([TeamsContractOracle.CallHoldOracle]::Model($inputJson)|ConvertFrom-Json -AsHashtable)})
}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-call-hold-provider-and-models';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString()
 stage='Actual pinned direct HttpRequestHelper with synthetic InvocationInfo, credential-refusing host and memory sender. Client/version/host restored. MPACmdlet factory default supplied. Actual model constructors/setters/Newtonsoft projections. No command lifecycle, credentials, discovery or network. This captures compiled CRUD only, not exported remoting Grant.'
 cases=$cases.ToArray();models=$models.ToArray()
}|ConvertTo-Json -Depth 25|Set-Content -LiteralPath $OutputPath -Encoding utf8
"call hold provider cases: $($cases.Count); model observations: $($models.Count)"
