# Offline direct request helpers; the sender/host cannot acquire credentials or network.
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
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@((Join-Path $PSHOME 'System.Management.Automation.dll'))+@($assemblies|ForEach-Object{Join-Path $ModulePath $_.path})
Add-Type -Path @((Join-Path $PSScriptRoot 'oracle/PolicyProviderCapture.cs'),(Join-Path $PSScriptRoot 'oracle/VoicemailValidationProviderCapture.cs')) -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$cases=[Collections.Generic.List[object]]::new()
foreach($c in @(
 @{name='read';op='read';status=200;reply='{"Identity":"fixture-response","AudioFileValidationEnabled":false,"AudioFileValidationUri":null,"Future":false}';force=$false;null_uri=$true},
 @{name='list';op='list';status=200;reply='[]';force=$false;null_uri=$true},
 @{name='patch-null';op='patch';status=204;reply='';force=$false;null_uri=$true},
 @{name='patch-uri';op='patch';status=204;reply='';force=$false;null_uri=$false},
 @{name='patch-force-false';op='patch';status=204;reply='';force=$true;null_uri=$true},
 @{name='read-empty';op='read';status=204;reply='';force=$false;null_uri=$true},
 @{name='read-invalid-uri';op='read';status=200;reply='{"Identity":"Global","AudioFileValidationUri":"http://example.invalid/a"}';force=$false;null_uri=$true},
 @{name='read-case-invalid-uri';op='read';status=200;reply='{"Identity":"Global","audiofilevalidationuri":"http://example.invalid/a"}';force=$false;null_uri=$true},
 @{name='list-envelope';op='list';status=200;reply='{"value":[]}';force=$false;null_uri=$true},
 @{name='read-denied';op='read';status=403;reply='{"Message":"fixture-denied"}';force=$false;null_uri=$true}
)) {
 $capture=[TeamsContractOracle.VoicemailValidationOracle]::Run($c.op,$c.status,$c.reply,$c.force,$c.null_uri)
 $cases.Add(@{name=$c.name;operation=$c.op;status=$c.status;response=$c.reply;force=$c.force;null_uri=$c.null_uri;capture=$capture})
}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-voicemail-validation-provider';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString()
 stage='Pinned direct HttpRequestHelper methods with synthetic InvocationInfo; credential-refusing host; private client/version/host replaced/restored; memory-only sender. MPACmdlet default inserted from factory source. No command constructor/ProcessRecord, auth, discovery, URI dereference or network.'
 cases=$cases.ToArray()
}|ConvertTo-Json -Depth 30|Set-Content -LiteralPath $OutputPath -Encoding utf8
"voicemail validation request observations: $($cases.Count)"
