[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblies=@(
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'},
 @{path='netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll';sha256='477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af'}
)
foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned recording dependency.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
$env:MSTeamsContextInternal=$null
Import-Module $manifest -Force
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@((Join-Path $PSHOME 'System.Management.Automation.dll'))+@($assemblies|ForEach-Object{Join-Path $ModulePath $_.path})+@((Join-Path $PSHOME 'Newtonsoft.Json.dll'))
Add-Type -Path @((Join-Path $PSScriptRoot 'oracle/PolicyProviderCapture.cs'),(Join-Path $PSScriptRoot 'oracle/RecordingMessageProviderCapture.cs')) -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$cases=[Collections.Generic.List[object]]::new()
foreach($c in @(
 @{name='read-identity';op='read';selector='identity';fields='none';status=200;reply='{"Identity":"fixture-response","Id":"00000000-0000-4000-8000-000000000031","Description":null,"RecordingAndTranscriptionLocalizationCustomMessage":[]}'},
 @{name='read-relative-id';op='read';selector='id';fields='none';status=204;reply=''},
 @{name='list';op='list';selector='identity';fields='none';status=200;reply='[]'},
 @{name='create-identity';op='create';selector='identity';fields='all';status=201;reply='{}'},
 @{name='create-relative-id';op='create';selector='id';fields='all';status=201;reply='{}'},
 @{name='create-generated-id';op='create';selector='auto';fields='all';status=201;reply='{}'},
 @{name='create-empty-settings';op='create';selector='id';fields='none';status=201;reply='{}'},
 @{name='create-non-guid-id';op='create';selector='invalid-id';fields='description';status=201;reply='{}'},
 @{name='patch-identity';op='patch';selector='identity';fields='all';status=204;reply=''},
 @{name='patch-relative-id';op='patch';selector='id';fields='all';status=204;reply=''},
 @{name='patch-id-description';op='patch';selector='id';fields='description';status=204;reply=''},
 @{name='delete-identity';op='delete';selector='identity';fields='none';status=204;reply=''},
 @{name='delete-relative-id';op='delete';selector='id';fields='none';status=204;reply=''},
 @{name='read-invalid-model-id';op='read';selector='id';fields='none';status=200;reply='{"Id":"fixture-invalid-guid"}'},
 @{name='read-null-collection';op='read';selector='id';fields='none';status=200;reply='{"RecordingAndTranscriptionLocalizationCustomMessage":null}'},
 @{name='read-denied';op='read';selector='id';fields='none';status=403;reply='{"Message":"fixture-denied"}'}
)){
 $force=$c.op -in @('create','patch','delete')
 $capture=[TeamsContractOracle.RecordingMessageOracle]::Run($c.op,$c.selector,$c.fields,$c.status,$c.reply,$force)
 $cases.Add(@{name=$c.name;operation=$c.op;selector=$c.selector;fields=$c.fields;status=$c.status;response=$c.reply;force=$force;capture=$capture})
}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-recording-message-provider';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString()
 stage='Actual scope-preparation methods and pinned direct HttpRequestHelper. Memory sender installed before command constructors; no lifecycle, credentials, discovery or network. Constructor GUID is checked then replaced by a fixed fixture ID before unbound-auto preparation.'
 cases=$cases.ToArray()
}|ConvertTo-Json -Depth 25|Set-Content -LiteralPath $OutputPath -Encoding utf8
"recording provider observations: $($cases.Count)"
