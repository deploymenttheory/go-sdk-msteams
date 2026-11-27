# Two-stage inert local binder only; no management command/lifecycle/auth/HTTP.
param(
 [Parameter(Mandatory=$true)][string]$ModulePath,
 [string]$VectorPath = (Join-Path $PSScriptRoot '../../msteams/teams_admin_api/custom_app_settings/testdata/pinned-set-projection.json')
)
$ErrorActionPreference='Stop' 
Import-Module $ModulePath -Force -WarningAction SilentlyContinue
foreach($a in [AppDomain]::CurrentDomain.GetAssemblies()){try{$null=$a.GetTypes()}catch{}}
$t=[AppDomain]::CurrentDomain.GetAssemblies()|ForEach-Object{$_.GetType('Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.GetCustomAppSettingResponse',$false)}|Where-Object{$null -ne $_}|Select-Object -First 1
if((Get-FileHash $t.Assembly.Location).Hash.ToLowerInvariant() -ne '8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'){throw 'Wrong assembly'}
$refs=@(Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll'|ForEach-Object{$_.FullName})+@($t.Assembly.Location,[Management.Automation.PSCmdlet].Assembly.Location)
Add-Type -Path (Join-Path $PSScriptRoot 'custom-app-binding-fakes.cs') -ReferencedAssemblies $refs -CompilerOptions '/nowarn:1701'
$fakeInfo=[Management.Automation.CmdletInfo]::new('Test-Ext90CustomAppBinding',[Ext90InertRealReplay.BindingCmdlet])
function Invoke-Ext90InertBridge {
 [CmdletBinding(DefaultParameterSetName='SetExpanded',PositionalBinding=$false)]
 param(
[Parameter(ParameterSetName='SetExpanded')][switch]$IsAppsEnabled,
[Parameter(ParameterSetName='SetExpanded')][switch]$IsAppsPurchaseEnabled,
[Parameter(ParameterSetName='SetExpanded')][switch]$IsExternalAppsEnabledByDefault,
[Parameter(ParameterSetName='SetExpanded')][switch]$IsLicenseBasedPinnedAppsEnabled,
[Parameter(ParameterSetName='SetExpanded')][switch]$IsSideloadedAppsInteractionEnabled,
[Parameter(ParameterSetName='SetExpanded')][switch]$IsTenantWideAutoInstallEnabled,
[Parameter(ParameterSetName='SetExpanded')][string]$LobTextColor,
[Parameter(ParameterSetName='SetExpanded')][string]$LobBackground,
[Parameter(ParameterSetName='SetExpanded')][string]$LobLogo,
[Parameter(ParameterSetName='SetExpanded')][string]$LobLogomark,
[Parameter(ParameterSetName='SetExpanded')][Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.IAppAccessRequestConfig]$AppAccessRequestConfig,
[Parameter(ParameterSetName='SetExpanded')][AllowEmptyCollection()][Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.IApplicationObject[]]$AppSettingsList
 )
 process { & $fakeInfo @PSBoundParameters }
}
$cases=@((Get-Content $VectorPath -Raw -Encoding utf8|ConvertFrom-Json).cases)
$cases+=@{case='malformed-bools';input_json='{"isAppsEnabled":"unexpected","isAppsPurchaseEnabled":null,"isExternalAppsEnabledByDefault":42,"isLicenseBasedPinnedAppsEnabled":[],"isTenantWideAutoInstallEnabled":{}}'}
$cases+=@{case='wrong-case-bools';input_json='{"IsAppsEnabled":true,"IsAppsPurchaseEnabled":true,"IsExternalAppsEnabledByDefault":true,"IsLicenseBasedPinnedAppsEnabled":true,"IsTenantWideAutoInstallEnabled":true}'}
$cases+=@{case='requested-false';input_json='{"isSideloadedAppsInteractionEnabled":true}';requested_enabled=$false}
$realType=$t.Assembly.GetType('Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Cmdlets.SetCsTeamsSettingsCustomApp_SetExpanded',$true)
$instanceFlags=[Reflection.BindingFlags]'Instance,NonPublic,Public'
$rows=@()
foreach($case in $cases){
 $read=$t.GetMethod('FromJsonString').Invoke($null,[object[]]@($case.input_json))
 $requested=$true;if($case.requested_enabled -is [bool]){$requested=$case.requested_enabled}
 $bound=@{IsSideloadedAppsInteractionEnabled=$requested;AppSettingsList=@()}
 foreach($name in @('IsAppsEnabled','IsAppsPurchaseEnabled','IsExternalAppsEnabledByDefault','IsLicenseBasedPinnedAppsEnabled','IsTenantWideAutoInstallEnabled','LobTextColor','LobBackground','LobLogo','LobLogomark','AppAccessRequestConfig')){$bound[$name]=$t.GetProperty($name).GetValue($read)}
 $failure=$null;$observation=$null
 try{$observation=Invoke-Ext90InertBridge @bound}catch{$e=$_.Exception;while($e.InnerException){$e=$e.InnerException};$failure=@{type=$e.GetType().FullName;message=$e.Message}}
 $realBodyJson=$null;$realFailure=$null
 if($null -ne $observation){
  $real=[Activator]::CreateInstance($realType)
  try{
   foreach($entry in $observation.BoundValues.GetEnumerator()){$realType.GetProperty($entry.Key).SetValue($real,$entry.Value)}
   $realBody=$realType.GetField('_body',$instanceFlags).GetValue($real)
   $realBodyJson=$realBody.ToJsonString()
  }catch{$e=$_.Exception;while($e.InnerException){$e=$e.InnerException};$realFailure=@{type=$e.GetType().FullName;message=$e.Message}}
  finally{$realType.GetField('_cancellationTokenSource',$instanceFlags).GetValue($real).Dispose()}
 }
 $rows+=@{case=$case.case;requested_enabled=$requested;input_json=$case.input_json;real_body_json=$realBodyJson;fake_only_body_json=$observation.BodyJson;bound_types=$observation.BoundTypes;bound_text=$observation.BoundText;binding_exception=$failure;real_setter_exception=$realFailure}
}
@{scope='Real local script binder -> inert compiled cmdlet binder captures exact converted parameters -> actual pinned SetExpanded constructor/property setters/private _body.ToJsonString; no management BeginProcessing/ProcessRecord/EndProcessing, auth or HTTP';script_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant();fake_sha256=(Get-FileHash (Join-Path $PSScriptRoot 'custom-app-binding-fakes.cs')).Hash.ToLowerInvariant();assembly_sha256=(Get-FileHash $t.Assembly.Location).Hash.ToLowerInvariant();cases=$rows}|ConvertTo-Json -Depth 10 -Compress
