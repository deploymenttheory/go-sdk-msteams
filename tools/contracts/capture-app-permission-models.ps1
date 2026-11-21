[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblies=@(
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'},
 @{path='netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll';sha256='477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af'}
)
foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned app permission dependency.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
$env:MSTeamsContextInternal=$null
Import-Module $manifest -Force
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@((Join-Path $PSHOME 'System.Management.Automation.dll'))+@($assemblies|ForEach-Object{Join-Path $ModulePath $_.path})+@((Join-Path $PSHOME 'Newtonsoft.Json.dll'))
Add-Type -Path @((Join-Path $PSScriptRoot 'oracle/PolicyProviderCapture.cs'),(Join-Path $PSScriptRoot 'oracle/AppPermissionProviderCapture.cs')) -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$policies=[Collections.Generic.List[object]]::new()
foreach($modelInput in @('constructor','future-types-long-description','{}','{"DefaultCatalogApps":null,"GlobalCatalogApps":null,"PrivateCatalogApps":null}','{"DefaultCatalogApps":[null]}','{"DefaultCatalogApps":[{"Id":"a"}],"DefaultCatalogApps":[{"Id":"b"}]}','{"DefaultCatalogApps":[{"Id":null}]}','{"DefaultCatalogApps":[{"Id":null,"Id":"good"}]}','{"DefaultCatalogApps":[{"Id":""}],"GlobalCatalogApps":[{"Id":"future-app"}],"PrivateCatalogApps":[{"Id":" "}]}','{"DefaultCatalogApps":{"Id":"wrong-envelope"}}')){
 $policies.Add(@{input=$modelInput;observation=([TeamsContractOracle.AppPermissionOracle]::PolicyModel($modelInput)|ConvertFrom-Json -AsHashtable)})
}
$apps=[Collections.Generic.List[object]]::new();$lists=[Collections.Generic.List[object]]::new()
foreach($catalog in @('default','global','private')){
 foreach($modelInput in @('null','',' ','future-app','json:{}','json:{"Id":null}','json:{"Id":null,"Id":"good"}')){
  $apps.Add(@{catalog=$catalog;input=$modelInput;observation=([TeamsContractOracle.AppPermissionOracle]::AppModel($catalog,$modelInput)|ConvertFrom-Json -AsHashtable)})
 }
 foreach($name in @('null','empty-operators','replace-duplicates','remove-one','remove-two','add-existing','remove-and-add')){
  $lists.Add(@{catalog=$catalog;name=$name;observation=([TeamsContractOracle.AppPermissionOracle]::Lists($catalog,$name)|ConvertFrom-Json -AsHashtable)})
 }
}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-app-permission-models';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString()
 stage='Actual pinned model constructors/setters/Newtonsoft projections and Core PSListModifierHelper with ordinal Id equality. App input null selects a null constructor argument; json: selects deserialization. No command lifecycle, credentials, discovery or request sender is invoked.'
 policies=$policies.ToArray();apps=$apps.ToArray();lists=$lists.ToArray()
}|ConvertTo-Json -Depth 25|Set-Content -LiteralPath $OutputPath -Encoding utf8
"app permission policy observations: $($policies.Count); app observations: $($apps.Count); list observations: $($lists.Count)"
