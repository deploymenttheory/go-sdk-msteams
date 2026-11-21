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
$models=[Collections.Generic.List[object]]::new()
foreach($name in @('language-null','language-empty','language-space','language-future','localization-absent','localization-null-language','message-absent','message-null-id','message-empty-id','message-invalid-id','message-id-final-lf','message-null-collection','message-long-description')) {
 $models.Add(@{name=$name;observation=([TeamsContractOracle.RecordingMessageOracle]::Model($name)|ConvertFrom-Json -AsHashtable)})
}
$lists=[Collections.Generic.List[object]]::new()
foreach($name in @('null','empty-operators','replace-empty','replace-duplicates','replace-null','remove-one','remove-two','add-existing','remove-and-add')) {
 $lists.Add(@{name=$name;observation=([TeamsContractOracle.RecordingMessageOracle]::List($name)|ConvertFrom-Json -AsHashtable)})
}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-recording-message-models';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString()
 stage='Actual pinned model constructors/setters/Newtonsoft projections and Core PSListModifierHelper.ApplyTo using the seven-field source comparator. No command lifecycle, credentials, discovery or request sender is invoked.'
 models=$models.ToArray();lists=$lists.ToArray()
}|ConvertTo-Json -Depth 25|Set-Content -LiteralPath $OutputPath -Encoding utf8
"recording model observations: $($models.Count); list observations: $($lists.Count)"
