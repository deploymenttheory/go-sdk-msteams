#requires -Version 7.0
# Offline pinned local permission builders and Set-list serialization only.
# Requires an absent host contract; never invokes a backend role cmdlet.
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
$role=Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.ConversationRole'
$permissionsType=Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.ConversationRolePermissions'
$chatType=Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.ChatPermissions'
$hostType=Find-Type 'Microsoft.Teams.ConfigAPI.CmdletHostContract.HostSingletionInstance'
$hostField=$hostType.GetField('HostContract',[Reflection.BindingFlags]'Public,NonPublic,Static')
if($null -ne $hostField.GetValue($null)){throw 'Offline builder invocation refused: host contract already initialized'}
$chatInfo=[System.Management.Automation.CmdletInfo]::new('New-CsTeamsChatPermissions',(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewChatPermissionsCmdlet'))
$callingInfo=[System.Management.Automation.CmdletInfo]::new('New-CsTeamsCallingPermissions',(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewCallingPermissionsCmdlet'))
$permissionsInfo=[System.Management.Automation.CmdletInfo]::new('New-CsTeamsConversationRolePermissions',(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewConversationRolePermissionsCmdlet'))
$chat=& $chatInfo -AddParticipants $false -SendMessage $true
$calling=& $callingInfo -AddParticipants $true -SendAudio $false
$permissions=& $permissionsInfo -Chat $chat -Calling $calling
if($null -ne $hostField.GetValue($null)){throw 'Host contract unexpectedly changed'}
$created=[DateTime]::SpecifyKind([DateTime]::ParseExact('2027-01-01T00:00:00','yyyy-MM-ddTHH:mm:ss',[Globalization.CultureInfo]::InvariantCulture),[DateTimeKind]::Unspecified)
$updated=[DateTime]::SpecifyKind($created.AddDays(1),[DateTimeKind]::Utc)
$httpType=(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Providers.HttpRequestHelper`1').MakeGenericType(@($role))
$helperInstance=$null
$additionalMethod=$httpType.GetMethod('AddAdditionalBodyContent',[Reflection.BindingFlags]'Instance,Static,Public,NonPublic')
$requiredMethod=$httpType.GetMethod('AddRequiredCmdletMetadata',[Reflection.BindingFlags]'Instance,Static,Public,NonPublic')
$nestedMethod=$httpType.GetMethod('GetNestedBoundParameters',[Reflection.BindingFlags]'Instance,Static,Public,NonPublic')
$roleSettings=Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.ConversationRolesSetting'
$settingsHttp=(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Providers.HttpRequestHelper`1').MakeGenericType(@($roleSettings))
$nestedMethod=$settingsHttp.GetMethod('GetNestedBoundParameters',[Reflection.BindingFlags]'Instance,Static,Public,NonPublic')
$additionalMethod=$settingsHttp.GetMethod('AddAdditionalBodyContent',[Reflection.BindingFlags]'Instance,Static,Public,NonPublic')
$modifierType=([System.Management.Automation.PSListModifier[int]]).GetGenericTypeDefinition().MakeGenericType(@($role))
$lambda=Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.SetConversationRolesSettingCmdlet+<>c'
$lambdaMethod=$lambda.GetMethod('<UpdateProperties>b__2_0',[Reflection.BindingFlags]'Instance,Static,Public,NonPublic')
$predicateType=([Func[int,int,bool]]).GetGenericTypeDefinition().MakeGenericType(@($role,$role,[bool]))
$predicate=$lambdaMethod.CreateDelegate($predicateType,[Activator]::CreateInstance($lambda,$true))
$helper=Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.PSListModifierHelper'
$apply=$helper.GetMethods()|Where-Object{$_.Name -eq 'ApplyTo' -and $_.IsGenericMethodDefinition -and $_.GetParameters().Count -eq 3 -and $_.GetParameters()[2].ParameterType.Name -like 'Func*'}|Select-Object -First 1
$apply=$apply.MakeGenericMethod(@($role))
function New-Role([string]$id){
 $value=[Activator]::CreateInstance($role,[object[]]@($permissions.psobject.BaseObject,$created,$updated))
 $value.Id=$id;$value.Identity='Global/'+$id;$value.Name='Controlled role'
 return $value
}
$result=@()
foreach($current in @('empty','nonempty')){
 foreach($scenario in @('null','empty','add','replace','remove','replace-and-remove')){
  $model=[Activator]::CreateInstance($roleSettings);$model.Identity='Global'
  if($current -eq 'nonempty'){$model.ConversationRoles.Add((New-Role '00000000-0000-4000-8000-000000000001'))}
  $modifier=[Activator]::CreateInstance($modifierType)
  switch($scenario){
   'null' {$modifier=$null}
   'add' {$modifier.Add.Add((New-Role '00000000-0000-4000-8000-000000000002'))}
   'replace' {$modifier.Replace.Add((New-Role '00000000-0000-4000-8000-000000000002'))}
   'remove' {$modifier.Remove.Add((New-Role '00000000-0000-4000-8000-000000000001'))}
   'replace-and-remove' {$modifier.Replace.Add((New-Role '00000000-0000-4000-8000-000000000002'));$modifier.Remove.Add((New-Role '00000000-0000-4000-8000-000000000001'))}
  }
  try{
   $null=$apply.Invoke($null,[object[]]@($modifier,$model.ConversationRoles,$predicate))
   $bound=[Collections.Generic.Dictionary[string,object]]::new();$bound.Add('Identity','Global');$bound.Add('ConversationRoles',$modifier)
   $expanded=$nestedMethod.Invoke($null,[object[]]@($bound,$roleSettings.psobject.BaseObject,$false))
   $resolver=[Activator]::CreateInstance($resolverType,[object[]]@($expanded));$settings=[Newtonsoft.Json.JsonSerializerSettings]::new();$settings.ContractResolver=$resolver
   $body=[Newtonsoft.Json.Linq.JObject]::Parse([Newtonsoft.Json.JsonConvert]::SerializeObject($model,$settings))
   $null=$additionalMethod.Invoke($null,[object[]]@($bound,$body))
   $result+=@{current=$current;scenario=$scenario;accepted=$true;expanded_keys=@($expanded.Keys|Sort-Object);body=$body.ToString([Newtonsoft.Json.Formatting]::None)}
  }catch{$result+=@{current=$current;scenario=$scenario;accepted=$false;failure_type=$_.Exception.GetType().FullName}}
 }
}
@{schema_version=1;baseline_version='8.1.0';extractor_sha256=(Get-FileHash -LiteralPath $PSCommandPath -Algorithm SHA256).Hash.ToLowerInvariant();observation='Actual local permission-builder output, exact compiled list helper/removal predicate, GetNestedBoundParameters(false), bound resolver and Set additional body. Host contract absent before/after local builders; no management command, authentication or HTTP. Controlled current model vectors are not successful tenant replies.';cases=$result}|ConvertTo-Json -Depth 8 -Compress
