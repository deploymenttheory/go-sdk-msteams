# Offline compiled request helpers; no management command or network.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblies=@(
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'},
 @{path='netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll';sha256='477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af'}
)
foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned assembly.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
Import-Module $manifest -Force
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@((Join-Path $PSHOME 'System.Management.Automation.dll'))+@($assemblies | ForEach-Object {Join-Path $ModulePath $_.path})
Add-Type -Path @((Join-Path $PSScriptRoot 'oracle/PolicyProviderCapture.cs'),(Join-Path $PSScriptRoot 'oracle/VoicemailProviderCapture.cs')) -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$cases=[Collections.Generic.List[object]]::new()
foreach($op in @('read','list','create','patch','delete')) {
 $status=200;$reply='{"Identity":"fixture-response","MaximumRecordingLength":"00:01:30","Future":false}'
 if($op -eq 'list'){$reply='[]'}
 if($op -in @('patch','delete')){$status=204;$reply=''}
 $capture=[TeamsContractOracle.VoicemailOracle]::Run($op,$status,$reply,$false,'normal')
 $cases.Add(@{name=$op;operation=$op;status=$status;response=$reply;force=$false;duration='normal';capture=$capture})
}
foreach($c in @(
 @{name='patch-null-duration';op='patch';status=204;reply='';force=$false;duration='null'},
 @{name='create-one-tick';op='create';status=201;reply='{"Identity":"Fixture"}';force=$false;duration='tick'},
 @{name='delete-force-false';op='delete';status=204;reply='';force=$true;duration='normal'},
 @{name='read-empty';op='read';status=204;reply='';force=$false;duration='normal'},
 @{name='read-null-duration';op='read';status=200;reply='{"Identity":"Global","MaximumRecordingLength":null}';force=$false;duration='normal'},
 @{name='read-invalid-duration';op='read';status=200;reply='{"Identity":"Global","MaximumRecordingLength":"PT90S"}';force=$false;duration='normal'},
 @{name='list-envelope';op='list';status=200;reply='{"value":[]}';force=$false;duration='normal'},
 @{name='delete-denied';op='delete';status=403;reply='{"Message":"fixture-denied"}';force=$false;duration='normal'}
)) {
 $capture=[TeamsContractOracle.VoicemailOracle]::Run($c.op,$c.status,$c.reply,$c.force,$c.duration)
 $cases.Add(@{name=$c.name;operation=$c.op;status=$c.status;response=$c.reply;force=$c.force;duration=$c.duration;capture=$capture})
}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-voicemail-compiled-provider';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString()
 stage='Pinned HttpRequestHelper direct CRUD with synthetic InvocationInfo. Host refuses credentials; private client/version/host replaced/restored; memory-only sender. MPACmdlet supplied from factory source. No command constructor/ProcessRecord, auth, discovery, audio-file access/upload or network.'
 cases=$cases.ToArray()
}|ConvertTo-Json -Depth 30|Set-Content -LiteralPath $OutputPath -Encoding utf8
"voicemail provider observations: $($cases.Count)"
