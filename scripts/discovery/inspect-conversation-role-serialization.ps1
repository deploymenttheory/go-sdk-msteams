#requires -Version 7.0
# Offline pinned local permission builders and request serialization only.
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
$builderBody=[Newtonsoft.Json.JsonConvert]::SerializeObject($permissions)
$builderExtendedProperties=@($permissions.psobject.Properties|Where-Object{$_.MemberType -in @('NoteProperty','ScriptProperty')}|ForEach-Object{$_.Name})
$created=[DateTime]::SpecifyKind([DateTime]::ParseExact('2027-01-01T00:00:00','yyyy-MM-ddTHH:mm:ss',[Globalization.CultureInfo]::InvariantCulture),[DateTimeKind]::Unspecified)
$updated=[DateTime]::SpecifyKind($created.AddDays(1),[DateTimeKind]::Utc)
$httpType=(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Providers.HttpRequestHelper`1').MakeGenericType(@($role))
$helperInstance=$null
$additionalMethod=$httpType.GetMethod('AddAdditionalBodyContent',[Reflection.BindingFlags]'Instance,Static,Public,NonPublic')
$requiredMethod=$httpType.GetMethod('AddRequiredCmdletMetadata',[Reflection.BindingFlags]'Instance,Static,Public,NonPublic')
$nestedMethod=$httpType.GetMethod('GetNestedBoundParameters',[Reflection.BindingFlags]'Instance,Static,Public,NonPublic')
$result=@()
foreach($scenario in @('identity-create','relative-id-create','identity-delete','relative-id-delete','identity-set','relative-id-set','relative-id-set-missing-model-id','permissions-set','sparse-null-name','false-isassignable','model-id-null','model-id-empty','model-id-invalid','model-id-valid','model-id-final-lf','permissions-null','chat-null')){
 try{
  $model=[Activator]::CreateInstance($role,[object[]]@($permissions.psobject.BaseObject,$created,$updated))
  $model.Identity='Global/00000000-0000-4000-8000-000000000001'
  $bound=[Collections.Generic.Dictionary[string,object]]::new();$bound.Add('Identity',$model.Identity)
  switch($scenario){
   'identity-create' {foreach($k in @('Permissions','CreatedAt','UpdatedAt')){$bound.Add($k,$model.$k)}}
   'relative-id-create' {foreach($k in @('Permissions','CreatedAt','UpdatedAt')){$bound.Add($k,$model.$k)};$bound.Add('Id','00000000-0000-4000-8000-000000000001')}
   'identity-delete' {$bound.Add('Force',[System.Management.Automation.SwitchParameter]::new($false))}
   'relative-id-delete' {$bound.Clear();$bound.Add('Id','00000000-0000-4000-8000-000000000001');$bound.Add('Force',[System.Management.Automation.SwitchParameter]::new($false))}
   'identity-set' {$model.Id='00000000-0000-4000-8000-000000000001';$bound.Add('Name',$null)}
   'relative-id-set' {$model.Id='00000000-0000-4000-8000-000000000001';$bound.Add('Id',$model.Id);$bound.Add('Name',$null)}
   'relative-id-set-missing-model-id' {$bound.Add('Id','00000000-0000-4000-8000-000000000001');$bound.Add('Name',$null)}
   'permissions-set' {$bound.Add('Permissions',$model.Permissions)}
   'sparse-null-name' {$role.GetProperty('Name').SetValue($model,$null);$bound.Add('Name',$null)}
   'false-isassignable' {$model.IsAssignable=$false;$bound.Add('IsAssignable',$false)}
   'model-id-null' {$role.GetProperty('Id').SetValue($model,$null);$bound.Add('Id',$null)}
   'model-id-empty' {$model.Id='';$bound.Add('Id','')}
   'model-id-invalid' {$model.Id='fixture-id';$bound.Add('Id','fixture-id')}
   'model-id-valid' {$model.Id='00000000-0000-4000-8000-000000000001';$bound.Add('Id',$model.Id)}
   'model-id-final-lf' {$model.Id="00000000-0000-4000-8000-000000000001`n";$bound.Add('Id',$model.Id)}
   'permissions-null' {$role.GetProperty('Permissions').SetValue($model,$null);$bound.Add('Permissions',$null)}
   'chat-null' {$permissionsType.GetProperty('Chat').SetValue($permissions,$null);$bound.Add('Permissions',$permissions)}
  }
  $expanded=$nestedMethod.Invoke($helperInstance,[object[]]@($bound,$role.psobject.BaseObject,$false));$resolver=[Activator]::CreateInstance($resolverType,[object[]]@($expanded));$settings=[Newtonsoft.Json.JsonSerializerSettings]::new();$settings.ContractResolver=$resolver
  if($scenario -like '*-delete'){$body=[Newtonsoft.Json.Linq.JObject]::new()}else{$body=[Newtonsoft.Json.Linq.JObject]::Parse([Newtonsoft.Json.JsonConvert]::SerializeObject($model,$settings))}
  $null=$additionalMethod.Invoke($null,[object[]]@($bound,$body))
  $metadata=[Collections.Generic.Dictionary[string,Collections.Generic.Dictionary[string,object]]]::new()
  $required=[Collections.Generic.Dictionary[string,object]]::new();$required.Add('Identity',$model.Identity);$required.Add('Id','00000000-0000-4000-8000-000000000001');$metadata.Add('RequiredNotBoundedParameters',$required)
  if($scenario -in @('identity-create','relative-id-create')){$null=$requiredMethod.Invoke($null,[object[]]@($body,$metadata))}
  $result+=@{scenario=$scenario;accepted=$true;expanded_keys=@($expanded.Keys|Sort-Object);body=$body.ToString([Newtonsoft.Json.Formatting]::None)}
 }catch{$result+=@{scenario=$scenario;accepted=$false;failure_type=$_.Exception.GetType().FullName;message=$_.Exception.Message}}
}
@{schema_version=1;baseline_version='8.1.0';extractor_sha256=(Get-FileHash -LiteralPath $PSCommandPath -Algorithm SHA256).Hash.ToLowerInvariant();observation='Actual compiled local permission builders with absent HostContract before/after; role data constructor and exact GetNestedBoundParameters(false), bound resolver, additional body, empty delete body and create-only required metadata. Backend role cmdlets, authentication and HTTP are never invoked. Role bound parameters/required metadata are controlled vectors constrained by inspected declarations and PrepareForScopeCheck IL, not captured backend invocation.';builder_body=$builderBody;builder_extended_properties=$builderExtendedProperties;cases=$result}|ConvertTo-Json -Depth 8 -Compress
