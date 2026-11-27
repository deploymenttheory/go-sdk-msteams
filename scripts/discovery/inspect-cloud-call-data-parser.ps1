# Pure pinned model/response parser observation only. No management lifecycle, host, auth or HTTP invocation.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference='Stop'
Import-Module (Join-Path $ModulePath 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
foreach($a in [AppDomain]::CurrentDomain.GetAssemblies()){try{$null=$a.GetTypes()}catch{}}
function Find-Type([string]$name){[AppDomain]::CurrentDomain.GetAssemblies()|ForEach-Object{$_.GetType($name,$false)}|Where-Object{$null -ne $_}|Select-Object -First 1}
$flags=[Reflection.BindingFlags]'Instance,Static,Public,NonPublic'
$hostType=Find-Type 'Microsoft.Teams.ConfigAPI.CmdletHostContract.HostSingletionInstance'
$hostField=$hostType.GetField('HostContract',[Reflection.BindingFlags]'Public,NonPublic,Static')
if($null -ne $hostField.GetValue($null)){throw 'Pure parser observation requires an absent host contract'}
$model=Find-Type 'Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.CloudCallDataConnection'
if((Get-FileHash $model.Assembly.Location).Hash.ToLowerInvariant() -ne '8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'){throw 'Wrong private assembly'}
$lambda=Find-Type 'Microsoft.Teams.ConfigAPI.Cmdlets.Generated.ConfigApiBasedCmdlets+<>c'
$instance=[Activator]::CreateInstance($lambda,$true)
$cases=@('[]','[{}]','[{"Token":null}]','[{"Token":""}]','[{"Token":"synthetic-offline"}]','null','{}','{"Token":"synthetic-offline"}','[null]','[null,{}, {"Token":"synthetic-offline"}]','[{"Token":"a","Unknown":42}]','[{"token":"lowercase"}]','[{"Token":123}]','[false,1,"x",[]]','')
$observations=@()
foreach($methodName in @('<CloudCallDataConnectionModernGet_Call>b__184_0','<CloudCallDataConnectionModernNew_Call>b__187_0')){
 $method=$lambda.GetMethod($methodName,$flags)
 foreach($json in $cases){
  $value=$null;$observedFailure=$null;$items=@()
  try{
   $task=[Threading.Tasks.Task]::FromResult[string]($json)
   $value=$method.Invoke($instance,[object[]]@(,$task))
   if($null -ne $value){foreach($item in $value){if($null -eq $item){$items+=@{is_null=$true}}else{$items+=@{is_null=$false;token_is_null=($null -eq $item.Token);token=$item.Token;reserialized=$item.ToJsonString()}}}}
  }catch{$e=$_.Exception;while($e.InnerException){$e=$e.InnerException};$observedFailure=@{type=$e.GetType().FullName;message=$e.Message}}
  $observations+=@{parser=$methodName;json=$json;result_is_null=($null -eq $value);count=$(if($null -eq $value){$null}else{$value.Length});items=$items;error=$observedFailure}
 }
}
$modelObservations=@()
foreach($json in @('{}','{"Token":null}','{"Token":""}','{"Token":"synthetic-offline","Unknown":42}','{"Token":123}','{"token":"lowercase"}','null','[]')){
 $value=$null;$observedFailure=$null
 try{$value=$model.GetMethod('FromJsonString',$flags).Invoke($null,[object[]]@($json))}catch{$e=$_.Exception;while($e.InnerException){$e=$e.InnerException};$observedFailure=@{type=$e.GetType().FullName;message=$e.Message}}
 $modelObservations+=@{json=$json;result_is_null=($null -eq $value);token_is_null=($null -eq $value -or $null -eq $value.Token);token=$(if($null -ne $value){$value.Token}else{$null});reserialized=$(if($null -ne $value){$value.ToJsonString()}else{$null});error=$observedFailure}
}
if($null -ne $hostField.GetValue($null)){throw 'Host unexpectedly initialized'}
@{assembly_sha256=(Get-FileHash $model.Assembly.Location).Hash.ToLowerInvariant();reproducer_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant();default_token_is_null=($null -eq ([Activator]::CreateInstance($model)).Token);parser_observations=$observations;model_observations=$modelObservations}|ConvertTo-Json -Depth 10 -Compress
