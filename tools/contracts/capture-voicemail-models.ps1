# Pure pinned voicemail model/duration serialization, with no cmdlet or I/O helper.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblies=@(
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'}
)
foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned assembly.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
Import-Module $manifest -Force
$cases=[Collections.Generic.List[object]]::new()
$ticks=@([long]0,[long]1,[long]-1,[long]9,[long]10,[long]11,[long]9999,[long]10000,[long]9999999,[long]10000000,[long]900000000,[long]864000000000,[long]864000000001,[long]::MaxValue,[long]::MinValue)
foreach($tick in $ticks) {
 $model=[Microsoft.Teams.Policy.Administration.Cmdlets.Core.OnlineVoicemailPolicy]::new()
 $model.Identity='fixture-policy';$model.Description='';$model.EnableTranscription=$false
 $model.ShareData='FutureValue';$model.PreambleAudioFile='fixture-audio-reference'
 $model.MaximumRecordingLength=[TimeSpan]::FromTicks($tick)
 $wire=[Newtonsoft.Json.JsonConvert]::SerializeObject($model.PSObject.BaseObject) | ConvertFrom-Json -AsHashtable
 $cases.Add(@{name=('ticks-'+$cases.Count);ticks=$tick;wire=$wire})
}
$model=[Microsoft.Teams.Policy.Administration.Cmdlets.Core.OnlineVoicemailPolicy]::new()
$model.Identity='fixture-policy';$model.MaximumRecordingLength=$null
$wire=[Newtonsoft.Json.JsonConvert]::SerializeObject($model.PSObject.BaseObject) | ConvertFrom-Json -AsHashtable
$cases.Add(@{name='null-duration';ticks=$null;wire=$wire})
$artifact=[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-voicemail-model-and-timespan';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString()
 stage='Pinned OnlineVoicemailPolicy model setters and Newtonsoft serialization; native TimeSpan ticks cover precision, signs, days and Int64 extrema. No cmdlet processing, file read/upload, credential access or network. No backend recording limit/default is inferred.'
 cases=$cases.ToArray()
}
$artifact | ConvertTo-Json -Depth 20 | Set-Content -LiteralPath $OutputPath -Encoding utf8
"voicemail model observations: $($cases.Count)"
