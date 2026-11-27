# Only exact pinned model FromJsonString, nullable getter and source custom null->false rule.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference='Stop'
Import-Module (Join-Path $ModulePath 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
foreach($assembly in [AppDomain]::CurrentDomain.GetAssemblies()){try{$null=$assembly.GetTypes()}catch{}}
$t=[AppDomain]::CurrentDomain.GetAssemblies()|ForEach-Object{$_.GetType('Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.GetCustomAppSettingResponse',$false)}|Where-Object{$null -ne $_}|Select-Object -First 1
if((Get-FileHash $t.Assembly.Location).Hash.ToLowerInvariant() -ne '8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'){throw 'Wrong assembly'}
$hostType=[AppDomain]::CurrentDomain.GetAssemblies()|ForEach-Object{$_.GetType('Microsoft.Teams.ConfigAPI.CmdletHostContract.HostSingletionInstance',$false)}|Where-Object{$null -ne $_}|Select-Object -First 1
$hostField=$hostType.GetField('HostContract',[Reflection.BindingFlags]'Public,NonPublic,Static')
if($null -ne $hostField.GetValue($null)){throw 'Pure projection observation requires an absent host contract'}
$parser=$t.GetMethod('FromJsonString');$getter=$t.GetMethod('get_IsSideloadedAppsInteractionEnabled')
$rows=@()
foreach($json in @('{"IsSideloadedAppsInteractionEnabled":"unexpected"}','{"isSideloadedAppsInteractionEnabled":"unexpected"}')){
 $value=$null;$observedFailure=$null;$flag=$null;$projected=$null
 try{$value=$parser.Invoke($null,[object[]]@($json));$flag=$getter.Invoke($value,@());$projected=if($null -eq $flag){$false}else{[bool]$flag}}catch{$e=$_.Exception;while($e.InnerException){$e=$e.InnerException};$observedFailure=@{type=$e.GetType().FullName;message=$e.Message}}
 $rows+=@{json=$json;model_is_null=($null -eq $value);flag_is_null=($null -eq $flag);flag=$flag;projected=$projected;exception=$observedFailure}
}
$methods=@();foreach($m in @($parser,$getter)){$methods+=@{method=[string]$m;il_sha256=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($m.GetMethodBody().GetILAsByteArray())).ToLowerInvariant()}}
if($null -ne $hostField.GetValue($null)){throw 'Host unexpectedly initialized'}
@{scope='Pure parser/getter observation; public null-to-false projection applied as inspected custom source rule, wrapper not invoked';assembly_sha256=(Get-FileHash $t.Assembly.Location).Hash.ToLowerInvariant();reproducer_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant();methods=$methods;observations=$rows}|ConvertTo-Json -Depth 8 -Compress
