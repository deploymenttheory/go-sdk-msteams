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
$cases=[Collections.Generic.List[object]]::new()
foreach($c in @(
 @{name='read';op='read';fields='none';status=200;reply='{"Identity":"fixture-response","AllowSideLoading":false,"AllowUserPinning":null,"AppPresetList":[]}'},
 @{name='list';op='list';fields='none';status=200;reply='[]'},
 @{name='create';op='create';fields='all';status=201;reply='{}'},
 @{name='create-empty-settings';op='create';fields='none';status=201;reply='{}'},
 @{name='patch';op='patch';fields='all';status=204;reply=''},
 @{name='patch-empty-lists';op='patch';fields='empty-lists';status=204;reply=''},
 @{name='patch-scalars';op='patch';fields='scalars';status=204;reply=''},
 @{name='delete';op='delete';fields='none';status=204;reply=''},
 @{name='grant-user';op='user';fields='none';status=204;reply=''},
 @{name='grant-tenant';op='tenant';fields='none';status=204;reply=''},
 @{name='grant-group';op='group';fields='none';status=204;reply=''},
 @{name='grant-user-empty';op='user';fields='none';status=204;reply=''},
 @{name='grant-tenant-empty';op='tenant';fields='none';status=204;reply=''},
 @{name='grant-group-empty';op='group';fields='none';status=204;reply=''},
 @{name='read-null-lists';op='read';fields='none';status=200;reply='{"AppPresetList":null,"PinnedAppBarApps":null,"PinnedMessageBarApps":null,"AppPresetMeetingList":null,"AdditionalCustomizationApps":null,"PinnedCallingBarApps":null}'},
 @{name='read-null-id';op='read';fields='none';status=200;reply='{"AppPresetList":[{"Id":null}]}'},
 @{name='read-int64-order';op='read';fields='none';status=200;reply='{"PinnedAppBarApps":[{"Id":"fixture","Order":9007199254740993}]}'},
 @{name='read-order-overflow';op='read';fields='none';status=200;reply='{"PinnedAppBarApps":[{"Id":"fixture","Order":9223372036854775808}]}'},
 @{name='read-null-customization';op='read';fields='none';status=200;reply='{"AdditionalCustomizationApps":[{"Id":"fixture","AdditionalCustomizationId":null}]}'},
 @{name='list-envelope';op='list';fields='none';status=200;reply='{"value":[]}'},
 @{name='read-denied';op='read';fields='none';status=403;reply='{"Message":"fixture-denied"}'}
)){
 $force=$c.op -in @('create','patch','delete');$nullPolicyName=$c.name -notlike '*-empty'
 $capture=[TeamsContractOracle.AppSetupOracle]::Run($c.op,$c.fields,$c.status,$c.reply,$force,$nullPolicyName)
 $name=if($nullPolicyName){$null}else{''}
 $cases.Add(@{name=$c.name;operation=$c.op;fields=$c.fields;status=$c.status;response=$c.reply;force=$force;policy_name=$name;capture=$capture})
}
[ordered]@{schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd');qualification='offline-app-setup-provider';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7';assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString();stage='Pinned direct HttpRequestHelper, synthetic InvocationInfo, credential-refusing host and memory sender. Private client/version/host replaced/restored. No command constructor/lifecycle, credentials, discovery or network.';cases=$cases.ToArray()}|ConvertTo-Json -Depth 25|Set-Content -LiteralPath $OutputPath -Encoding utf8
"app setup provider observations: $($cases.Count)"
