[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblies=@(
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'},
 @{path='netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll';sha256='477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af'}
)
foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned privacy dependency.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
$env:MSTeamsContextInternal=$null
Import-Module $manifest -Force
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@((Join-Path $PSHOME 'System.Management.Automation.dll'))+@($assemblies|ForEach-Object{Join-Path $ModulePath $_.path})
Add-Type -Path @((Join-Path $PSScriptRoot 'oracle/PolicyProviderCapture.cs'),(Join-Path $PSScriptRoot 'oracle/PrivacyProviderCapture.cs')) -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$cases=[Collections.Generic.List[object]]::new()
foreach($c in @(
 @{name='read';op='read';status=200;reply='{"Identity":"fixture-response","EnablePrivacyMode":false,"AutoInitiateContacts":true,"PublishLocationDataDefault":null,"DisplayPublishedPhotoDefault":false,"Future":false}';force=$false;null_location=$true},
 @{name='list';op='list';status=200;reply='[]';force=$false;null_location=$true},
 @{name='patch-null';op='patch';status=204;reply='';force=$false;null_location=$true},
 @{name='patch-false';op='patch';status=204;reply='';force=$false;null_location=$false},
 @{name='patch-force-false';op='patch';status=204;reply='';force=$true;null_location=$true},
 @{name='read-empty';op='read';status=204;reply='';force=$false;null_location=$true},
 @{name='list-envelope';op='list';status=200;reply='{"value":[]}';force=$false;null_location=$true},
 @{name='read-denied';op='read';status=403;reply='{"Message":"fixture-denied"}';force=$false;null_location=$true}
)){
 $capture=[TeamsContractOracle.PrivacyOracle]::Run($c.op,$c.status,$c.reply,$c.force,$c.null_location)
 $cases.Add(@{name=$c.name;operation=$c.op;status=$c.status;response=$c.reply;force=$c.force;null_location=$c.null_location;capture=$capture})
}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-privacy-provider';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString()
 stage='Pinned direct HttpRequestHelper with synthetic InvocationInfo, credential-refusing host and memory sender. Private client/version/host are replaced/restored; MPACmdlet default comes from factory source. No command constructor/ProcessRecord, credentials, discovery or network.'
 cases=$cases.ToArray()
}|ConvertTo-Json -Depth 25|Set-Content -LiteralPath $OutputPath -Encoding utf8
"privacy provider observations: $($cases.Count)"
