#requires -Version 7.0
# Offline data-model construction and serialization only; no cmdlet invocation.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference='Stop'
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong module baseline'}
Import-Module $manifest -Force -WarningAction SilentlyContinue
foreach($assembly in [AppDomain]::CurrentDomain.GetAssemblies()){try{$null=$assembly.GetTypes()}catch{}}
function Find-Type([string]$Name){[AppDomain]::CurrentDomain.GetAssemblies()|ForEach-Object{$_.GetType($Name,$false)}|Where-Object{$null -ne $_}|Select-Object -First 1}
foreach($pin in @(@{file='Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';hash='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},@{file='Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';hash='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'})){
 if((Get-FileHash -LiteralPath (Join-Path (Join-Path $ModulePath 'netcoreapp3.1') $pin.file) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $pin.hash){throw 'Wrong pinned model or resolver assembly'}
}
$resolverType=Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Providers.Helpers.BoundParametersContractResolver'
$result=@()
foreach($scenario in @(
 @{name='uint32-full-range';model='TenantConfiguration';values=@{MaxAllowedDomains=[uint32]::MaxValue;MaxBlockedDomains=[uint32]1};bound=@('MaxAllowedDomains','MaxBlockedDomains')},
 @{name='uint32-sparse-zero';model='TenantConfiguration';values=@{MaxAllowedDomains=[uint32]0;MaxBlockedDomains=[uint32]42};bound=@('MaxAllowedDomains')},
 @{name='uint32-sparse-one';model='TenantConfiguration';values=@{MaxAllowedDomains=[uint32]1;MaxBlockedDomains=[uint32]42};bound=@('MaxAllowedDomains')},
 @{name='uint32-explicit-null';model='TenantConfiguration';values=@{MaxAllowedDomains=$null;MaxBlockedDomains=[uint32]42};bound=@('MaxAllowedDomains')},
 @{name='uri-empty';model='TeamsAppPolicyConfiguration';values=@{AppCatalogUri=''};bound=@('AppCatalogUri')},
 @{name='uri-explicit-null';model='TeamsAppPolicyConfiguration';values=@{AppCatalogUri=$null;ResourceUri=$null};bound=@('AppCatalogUri','ResourceUri')},
 @{name='uri-terminal-lf';model='TeamsAppPolicyConfiguration';values=@{AppCatalogUri="https://catalog.example.invalid/`n"};bound=@('AppCatalogUri')},
 @{name='uri-http';model='TeamsAppPolicyConfiguration';values=@{AppCatalogUri='http://catalog.example.invalid/'};bound=@('AppCatalogUri')},
 @{name='uri-sparse';model='TeamsAppPolicyConfiguration';values=@{AppCatalogUri='https://catalog.example.invalid/';ResourceUri='https://resource.example.invalid/unbound'};bound=@('AppCatalogUri')}
)){
 $type=Find-Type ('Microsoft.Teams.Policy.Administration.Cmdlets.Core.'+$scenario.model)
 try {
 $model=[Activator]::CreateInstance($type)
 $model.Identity='Global'
 foreach($key in $scenario.values.Keys){$type.GetProperty($key).SetValue($model,$scenario.values[$key])}
 $bound=[Collections.Generic.Dictionary[string,object]]::new()
 $bound.Add('Identity','Global')
 foreach($key in $scenario.bound){$bound.Add($key,$scenario.values[$key])}
 $resolver=[Activator]::CreateInstance($resolverType,[object[]]@($bound))
 $settings=[Newtonsoft.Json.JsonSerializerSettings]::new()
 $settings.ContractResolver=$resolver
 $result+=@{name=$scenario.name;model=$type.FullName;model_assembly=[IO.Path]::GetFileName($type.Assembly.Location);model_sha256=(Get-FileHash $type.Assembly.Location -Algorithm SHA256).Hash.ToLowerInvariant();resolver=$resolverType.FullName;resolver_assembly=[IO.Path]::GetFileName($resolverType.Assembly.Location);resolver_sha256=(Get-FileHash $resolverType.Assembly.Location -Algorithm SHA256).Hash.ToLowerInvariant();null_handling=[string]$settings.NullValueHandling;accepted=$true;body=[Newtonsoft.Json.JsonConvert]::SerializeObject($model,$settings)}
 }catch{
 $result+=@{name=$scenario.name;model=$type.FullName;accepted=$false;property=$key;failure_type=$_.Exception.GetType().FullName}
 }
}
@{schema_version=1;baseline_version='8.1.0';extractor_sha256=(Get-FileHash -LiteralPath $PSCommandPath -Algorithm SHA256).Hash.ToLowerInvariant();observation='Offline data-model setters and BoundParametersContractResolver serialization. PropertyInfo.SetValue preserves CLR null without PowerShell string coercion. No cmdlet instance, sign-in or tenant operation.';cases=$result}|ConvertTo-Json -Depth 8 -Compress
