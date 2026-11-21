[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblies=@(
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'},
 @{path='netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll';sha256='477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af'}
)
foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned app setup dependency.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
$env:MSTeamsContextInternal=$null;Import-Module $manifest -Force
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@((Join-Path $PSHOME 'System.Management.Automation.dll'))+@($assemblies|ForEach-Object{Join-Path $ModulePath $_.path})+@((Join-Path $PSHOME 'Newtonsoft.Json.dll'))
Add-Type -Path @((Join-Path $PSScriptRoot 'oracle/PolicyProviderCapture.cs'),(Join-Path $PSScriptRoot 'oracle/AppSetupProviderCapture.cs')) -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$policies=[Collections.Generic.List[object]]::new()
foreach($modelInput in @('constructor','long-description','{}','{"AppPresetList":null,"PinnedAppBarApps":null,"PinnedMessageBarApps":null,"AppPresetMeetingList":null,"AdditionalCustomizationApps":null,"PinnedCallingBarApps":null}','{"PinnedAppBarApps":[{"Id":"a","Order":1}],"PinnedAppBarApps":[{"Id":"b","Order":2}]}','{"AllowSideLoading":false,"AllowUserPinning":true}','{"PinnedMessageBarApps":[null]}','{"PinnedAppBarApps":[{"Id":"a","Order":9007199254740993}]}','{"PinnedMessageBarApps":[{"Id":"a","Order":null}]}','{"AdditionalCustomizationApps":[{"Id":"a","AdditionalCustomizationId":null}]}','{"PinnedCallingBarApps":[{"Id":null}]}','{"AppPresetList":{"Id":"bad-shape"}}')){
 $policies.Add(@{input=$modelInput;observation=([TeamsContractOracle.AppSetupOracle]::PolicyModel($modelInput)|ConvertFrom-Json -AsHashtable)})
}
$elements=[Collections.Generic.List[object]]::new();$lists=[Collections.Generic.List[object]]::new()
foreach($field in [TeamsContractOracle.AppSetupOracle]::Collections){
 $inputs=[Collections.Generic.List[string]]::new()
 foreach($modelInput in @('null','','future-app','json:{}','json:{"Id":null}','json:{"Id":""}','json:{"Id":null,"Id":"good"}')){$inputs.Add($modelInput)}
 if($field -in @('PinnedAppBarApps','PinnedMessageBarApps')){
  foreach($order in @('0','-1','9223372036854775807','null','9223372036854775808','"2"')){$inputs.Add('json:{"Id":"a","Order":'+$order+'}')}
 }
 if($field -eq 'AdditionalCustomizationApps'){$inputs.Add('json:{"Id":"a","AdditionalCustomizationId":null}');$inputs.Add('json:{"Id":"a","AdditionalCustomizationId":""}')}
 foreach($modelInput in $inputs){$elements.Add(@{field=$field;input=$modelInput;observation=([TeamsContractOracle.AppSetupOracle]::ElementModel($field,$modelInput)|ConvertFrom-Json -AsHashtable)})}
 foreach($name in @('null','empty-operators','remove-one','remove-two','remove-three','replace-duplicates','add-existing','remove-and-add')){
  $lists.Add(@{field=$field;name=$name;observation=([TeamsContractOracle.AppSetupOracle]::Lists($field,$name)|ConvertFrom-Json -AsHashtable)})
 }
}
[ordered]@{schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd');qualification='offline-app-setup-models';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7';assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString();stage='Pinned model constructors/setters/Newtonsoft and Core PSListModifierHelper with source-specific Id, Id+Order or Id+AdditionalCustomizationId equality. Element input null selects null constructor Id; json: selects deserialization. No command lifecycle, credentials, discovery or sender invoked.';policies=$policies.ToArray();elements=$elements.ToArray();lists=$lists.ToArray()}|ConvertTo-Json -Depth 25|Set-Content -LiteralPath $OutputPath -Encoding utf8
"app setup policy observations: $($policies.Count); element observations: $($elements.Count); list observations: $($lists.Count)"
