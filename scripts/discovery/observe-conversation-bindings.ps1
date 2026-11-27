# Controlled offline binding observation. Run in a fresh pwsh -NoProfile process.
# All lifecycle I/O seams are replaced before cmdlet construction; never connects.
# Invoke through a parent process with a 30-second timeout, one scenario per process.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,
 [Parameter(Mandatory)][ValidateSet('role-get-all','role-get-identity','role-get-relative','role-get-filter','role-new-identity','role-new-relative','role-new-comparison-denied','role-set-identity','role-set-relative','role-remove-identity','role-remove-relative','setting-get-identity','setting-get-filter','setting-set-replace')][string]$Scenario)
$ErrorActionPreference='Stop'
$module=$ModulePath
if((Import-PowerShellDataFile (Join-Path $module 'MicrosoftTeams.psd1')).ModuleVersion -ne '8.1.0'){throw 'Wrong module baseline'}
Import-Module (Join-Path $module 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
foreach($assembly in [AppDomain]::CurrentDomain.GetAssemblies()){try{$null=$assembly.GetTypes()}catch{}}
function Find-Type([string]$name){[AppDomain]::CurrentDomain.GetAssemblies()|ForEach-Object{$_.GetType($name,$false)}|Where-Object{$null -ne $_}|Select-Object -First 1}
$provider=Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Providers.HttpClientHelper'
$core=Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewConversationRoleCmdlet'
$contract=Find-Type 'Microsoft.Teams.ConfigAPI.CmdletHostContract.HostSingletionInstance'
$loggerType=Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Providers.Logger'
$utils=Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PowershellUtils'
$oce=Find-Type 'Microsoft.Teams.ConfigAPI.CmdletHostContract.OCEContext'
$flags=[Reflection.BindingFlags]'Public,NonPublic,Static'
if((Get-FileHash $core.Assembly.Location).Hash.ToLowerInvariant() -ne '68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'){throw 'Wrong Core assembly'}
if((Get-FileHash $provider.Assembly.Location).Hash.ToLowerInvariant() -ne 'a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'){throw 'Wrong PolicyRp assembly'}
$hostField=$contract.GetField('HostContract',$flags)
$clientField=$provider.GetField('_client',$flags)
$versionField=$utils.GetField('moduleVersion',$flags)
$loggerProperty=$loggerType.GetProperty('Instance',$flags)
foreach($f in @($hostField,$clientField,$versionField)){if(!$f -or $f.IsInitOnly){throw 'Expected mutable static seam missing'}}
if($null -ne $hostField.GetValue($null) -or $null -ne $clientField.GetValue($null)){throw 'Fresh process required: ambient host or HTTP client exists'}
$oldLogger=$loggerProperty.GetValue($null);$oldVersion=$versionField.GetValue($null)
$oldRelease=[Environment]::GetEnvironmentVariable('MSTeamsReleaseEnvironment')
$oldInternal=[Environment]::GetEnvironmentVariable('MSTeamsContextInternal')
# Reflection metadata only: reject pre-existing nonempty OCE identity/header context.
foreach($name in @('_TenantDomain','_TenantId','_AppId','_ForestName','_IsDMS','_IsOCEContextSet','_Region','_Ring','_Cloud','_Headers','_IsSystemTenant')){
 $p=$oce.GetProperty($name,$flags);$v=$p.GetValue($null)
 if($null -ne $v -and $v -ne $false -and $v -ne ''){if($v -is [Collections.IDictionary] -and $v.Count -eq 0){continue};throw "Ambient OCE value: $name"}
}
$environmentProperty=$oce.GetProperty('_Environment',$flags)
if(!$environmentProperty.GetValue($null).Equals([Activator]::CreateInstance($environmentProperty.PropertyType))){throw 'Ambient nondefault OCE environment'}
$references=@(Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll'|ForEach-Object{$_.FullName})+@($provider.Assembly.Location,$contract.Assembly.Location)
Add-Type -Path (Join-Path $PSScriptRoot 'conversation-offline-fakes.cs') -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$logger=[Ext90Offline.MemoryLogger]::new();$telemetry=[Ext90Offline.MemoryTelemetry]::new()
$fake=[Ext90Offline.FakeHost]::new($logger,$telemetry)
$handler=[Ext90Offline.TerminalHandler]::new();$client=[Net.Http.HttpClient]::new($handler,$true)
$client.Timeout=[TimeSpan]::FromSeconds(5)
$runspace=$null;$pipeline=$null
try{
 [Environment]::SetEnvironmentVariable('MSTeamsReleaseEnvironment',$null)
 [Environment]::SetEnvironmentVariable('MSTeamsContextInternal',$null)
 $hostField.SetValue($null,$fake);$clientField.SetValue($null,$client)
 $loggerProperty.SetValue($null,$logger);$versionField.SetValue($null,'0.0.0-offline')
 if(![Object]::ReferenceEquals($hostField.GetValue($null),$fake) -or ![Object]::ReferenceEquals($clientField.GetValue($null),$client) -or ![Object]::ReferenceEquals($loggerProperty.GetValue($null),$logger) -or $versionField.GetValue($null) -ne '0.0.0-offline'){throw 'Static installation assertion failed'}
 $selfhost=$utils.GetMethod('IsSelfHost',$flags).Invoke($null,@())
 $torus=$utils.GetMethod('IsTorusEnvironment',$flags).Invoke($null,@())
 if($selfhost -or $torus){throw 'Controlled environment assertion failed'}
 # Empty initial session: no profiles, imported modules, inherited functions, scripts or credentials.
 # Register only these exact pinned compiled management types. This reads declarations only.
 $initial=[Management.Automation.Runspaces.InitialSessionState]::Create()
 $commands=@{
 'Get-CsConversationRole'='GetConversationRoleCmdlet';'New-CsConversationRole'='NewConversationRoleCmdlet';
 'Set-CsConversationRole'='SetConversationRoleCmdlet';'Remove-CsConversationRole'='RemoveConversationRoleCmdlet';
 'Get-CsConversationRolesSetting'='GetConversationRolesSettingCmdlet';'Set-CsConversationRolesSetting'='SetConversationRolesSettingCmdlet'}
 $infos=@{}
 foreach($entry in $commands.GetEnumerator()){
  $t=Find-Type ('Microsoft.Teams.Policy.Administration.Cmdlets.Core.'+$entry.Value)
  $initial.Commands.Add([Management.Automation.Runspaces.SessionStateCmdletEntry]::new($entry.Key,$t,$null))
  $infos[$entry.Key]=[Management.Automation.CmdletInfo]::new($entry.Key,$t)
 }
 $runspace=[Management.Automation.Runspaces.RunspaceFactory]::CreateRunspace($initial);$runspace.Open()
 $pipeline=[Management.Automation.PowerShell]::Create();$pipeline.Runspace=$runspace
 $handler.MaxCalls=4;$fake.MaxTokenCalls=2
 switch($Scenario){
  'role-get-all' {
 $reply=[Ext90Offline.PlannedReply]::new();$reply.Method='GET';$reply.AbsoluteUri='https://offline.example.invalid/Skype.Policy/configurations/ConversationRole';$reply.ExpectedBody=$null;$reply.Json='[]'
 $handler.Replies.Enqueue($reply)
 $null=$pipeline.AddCommand($infos['Get-CsConversationRole'])
  }
  'role-get-identity' {
 $handler.MaxCalls=4;$fake.MaxTokenCalls=2
 $r=[Ext90Offline.PlannedReply]::new();$r.Method='GET';$r.AbsoluteUri='https://offline.example.invalid/Skype.Policy/configurations/ConversationRole/configuration/Global/00000000-0000-4000-8000-000000000001';$r.ExpectedBody=$null;$r.Json='{"Identity":"Global/00000000-0000-4000-8000-000000000001","Id":"00000000-0000-4000-8000-000000000001","Name":"Current role","Permissions":{"Chat":{},"Calling":{}},"CreatedAt":"2027-01-01T00:00:00","UpdatedAt":"2027-01-02T00:00:00Z"}';$r.Status=[Net.HttpStatusCode]::OK;$handler.Replies.Enqueue($r)
 $null=$pipeline.AddCommand($infos['Get-CsConversationRole']).AddParameter('Identity','Global/00000000-0000-4000-8000-000000000001')
  }
  'role-get-relative' {
 $handler.MaxCalls=4;$fake.MaxTokenCalls=2
 $r=[Ext90Offline.PlannedReply]::new();$r.Method='GET';$r.AbsoluteUri='https://offline.example.invalid/Skype.Policy/configurations/ConversationRole/configuration/Global/00000000-0000-4000-8000-000000000001';$r.ExpectedBody=$null;$r.Json='{"Identity":"Global/00000000-0000-4000-8000-000000000001","Id":"00000000-0000-4000-8000-000000000001","Name":"Current role","Permissions":{"Chat":{},"Calling":{}},"CreatedAt":"2027-01-01T00:00:00","UpdatedAt":"2027-01-02T00:00:00Z"}';$r.Status=[Net.HttpStatusCode]::OK;$handler.Replies.Enqueue($r)
 $null=$pipeline.AddCommand($infos['Get-CsConversationRole']).AddParameter('Id','00000000-0000-4000-8000-000000000001')
  }
  'role-get-filter' {
 $handler.MaxCalls=4;$fake.MaxTokenCalls=2
 $r=[Ext90Offline.PlannedReply]::new();$r.Method='GET';$r.AbsoluteUri='https://offline.example.invalid/Skype.Policy/configurations/ConversationRole';$r.ExpectedBody=$null;$r.Json='[{"Identity":"Global/00000000-0000-4000-8000-000000000001","Id":"00000000-0000-4000-8000-000000000001","Name":"Current role","Permissions":{"Chat":{},"Calling":{}},"CreatedAt":"2027-01-01T00:00:00","UpdatedAt":"2027-01-02T00:00:00Z"}]';$r.Status=[Net.HttpStatusCode]::OK;$handler.Replies.Enqueue($r)
 $null=$pipeline.AddCommand($infos['Get-CsConversationRole']).AddParameter('Filter','*0001')
  }
  'role-new-identity' {
 $chatInfo=[Management.Automation.CmdletInfo]::new('New-CsTeamsChatPermissions',(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewChatPermissionsCmdlet'))
 $callingInfo=[Management.Automation.CmdletInfo]::new('New-CsTeamsCallingPermissions',(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewCallingPermissionsCmdlet'))
 $permissionsInfo=[Management.Automation.CmdletInfo]::new('New-CsTeamsConversationRolePermissions',(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewConversationRolePermissionsCmdlet'))
 $chat=& $chatInfo -AddParticipants $false -SendMessage $true
 $calling=& $callingInfo -AddParticipants $true -SendAudio $false
 $permissions=& $permissionsInfo -Chat $chat -Calling $calling
 $created=[DateTime]::SpecifyKind([DateTime]::new(2027,1,1),[DateTimeKind]::Unspecified)
 $updated=[DateTime]::SpecifyKind($created.AddDays(1),[DateTimeKind]::Utc)
 $reply=[Ext90Offline.PlannedReply]::new();$reply.Method='POST';$reply.AbsoluteUri='https://offline.example.invalid/Skype.Policy/configurations/ConversationRole';$reply.ExpectedBody='{"Identity":"Global/00000000-0000-4000-8000-000000000001","Permissions":{"Identity":null},"CreatedAt":"2027-01-01T00:00:00","UpdatedAt":"2027-01-02T00:00:00Z","Id":"00000000-0000-4000-8000-000000000001"}';$reply.Status=[Net.HttpStatusCode]::Created
 $reply.Json='{"Identity":"Global/00000000-0000-4000-8000-000000000001","Id":"00000000-0000-4000-8000-000000000001","Permissions":{"Chat":{},"Calling":{}},"CreatedAt":"2027-01-01T00:00:00","UpdatedAt":"2027-01-02T00:00:00Z"}'
 $reply.ExpectedBody=[Newtonsoft.Json.Linq.JObject]::Parse($reply.ExpectedBody).ToString([Newtonsoft.Json.Formatting]::Indented)
 $preflight=[Ext90Offline.PlannedReply]::new();$preflight.Method='GET';$preflight.AbsoluteUri='https://offline.example.invalid/Skype.Policy/configurations/ConversationRole/configuration/Global';$preflight.ExpectedBody='';$preflight.Json=$reply.Json.Replace('Global/00000000-0000-4000-8000-000000000001','Global')
 $handler.Replies.Enqueue($preflight)
 $handler.Replies.Enqueue($reply)
 $null=$pipeline.AddCommand($infos['New-CsConversationRole']).AddParameter('Identity','Global/00000000-0000-4000-8000-000000000001').AddParameter('Permissions',$permissions).AddParameter('CreatedAt',$created).AddParameter('UpdatedAt',$updated).AddParameter('Confirm',$false)
  }
  'role-new-relative' {
 $handler.MaxCalls=4;$fake.MaxTokenCalls=2
 $chatInfo=[Management.Automation.CmdletInfo]::new('New-CsTeamsChatPermissions',(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewChatPermissionsCmdlet'))
 $callingInfo=[Management.Automation.CmdletInfo]::new('New-CsTeamsCallingPermissions',(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewCallingPermissionsCmdlet'))
 $permissionsInfo=[Management.Automation.CmdletInfo]::new('New-CsTeamsConversationRolePermissions',(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewConversationRolePermissionsCmdlet'))
 $chat=& $chatInfo -AddParticipants $false -SendMessage $true
 $calling=& $callingInfo -AddParticipants $true -SendAudio $false
 $permissions=& $permissionsInfo -Chat $chat -Calling $calling
 $created=[DateTime]::SpecifyKind([DateTime]::new(2027,1,1),[DateTimeKind]::Unspecified)
 $updated=[DateTime]::SpecifyKind($created.AddDays(1),[DateTimeKind]::Utc)
 $r=[Ext90Offline.PlannedReply]::new();$r.Method='GET';$r.AbsoluteUri='https://offline.example.invalid/Skype.Policy/configurations/ConversationRole/configuration/Global';$r.ExpectedBody=$null;$r.Json='{"Identity":"Global","Id":"00000000-0000-4000-8000-000000000001","Name":"Current role","Permissions":{"Chat":{},"Calling":{}},"CreatedAt":"2027-01-01T00:00:00","UpdatedAt":"2027-01-02T00:00:00Z"}';$r.Status=[Net.HttpStatusCode]::OK;$handler.Replies.Enqueue($r)
 $r=[Ext90Offline.PlannedReply]::new();$r.Method='POST';$r.AbsoluteUri='https://offline.example.invalid/Skype.Policy/configurations/ConversationRole';$r.ExpectedBody=[Newtonsoft.Json.Linq.JObject]::Parse('{"Permissions":{},"Id":null,"CreatedAt":"2027-01-01T00:00:00","UpdatedAt":"2027-01-02T00:00:00Z","Identity":"Global/00000000-0000-4000-8000-000000000001"}').ToString([Newtonsoft.Json.Formatting]::Indented);$r.Json='{"Identity":"Global/00000000-0000-4000-8000-000000000001","Id":"00000000-0000-4000-8000-000000000001","Name":"Current role","Permissions":{"Chat":{},"Calling":{}},"CreatedAt":"2027-01-01T00:00:00","UpdatedAt":"2027-01-02T00:00:00Z"}';$r.Status=[Net.HttpStatusCode]::Created;$handler.Replies.Enqueue($r)
 $null=$pipeline.AddCommand($infos['New-CsConversationRole']).AddParameter('Id','00000000-0000-4000-8000-000000000001').AddParameter('Permissions',$permissions).AddParameter('CreatedAt',$created).AddParameter('UpdatedAt',$updated).AddParameter('Confirm',$false)
  }
  'role-new-comparison-denied' {
 $chatInfo=[Management.Automation.CmdletInfo]::new('New-CsTeamsChatPermissions',(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewChatPermissionsCmdlet'))
 $callingInfo=[Management.Automation.CmdletInfo]::new('New-CsTeamsCallingPermissions',(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewCallingPermissionsCmdlet'))
 $permissionsInfo=[Management.Automation.CmdletInfo]::new('New-CsTeamsConversationRolePermissions',(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewConversationRolePermissionsCmdlet'))
 $chat=& $chatInfo -AddParticipants $false -SendMessage $true
 $calling=& $callingInfo -AddParticipants $true -SendAudio $false
 $permissions=& $permissionsInfo -Chat $chat -Calling $calling
 $created=[DateTime]::SpecifyKind([DateTime]::new(2027,1,1),[DateTimeKind]::Unspecified)
 $updated=[DateTime]::SpecifyKind($created.AddDays(1),[DateTimeKind]::Utc)
 $reply=[Ext90Offline.PlannedReply]::new();$reply.Method='POST';$reply.AbsoluteUri='https://offline.example.invalid/Skype.Policy/configurations/ConversationRole';$reply.ExpectedBody='{"Identity":"Global/00000000-0000-4000-8000-000000000001","Permissions":{"Identity":null},"CreatedAt":"2027-01-01T00:00:00","UpdatedAt":"2027-01-02T00:00:00Z","Id":"00000000-0000-4000-8000-000000000001"}';$reply.Status=[Net.HttpStatusCode]::Created
 $reply.Json='{"Identity":"Global/00000000-0000-4000-8000-000000000001","Id":"00000000-0000-4000-8000-000000000001","Permissions":{"Chat":{},"Calling":{}},"CreatedAt":"2027-01-01T00:00:00","UpdatedAt":"2027-01-02T00:00:00Z"}'
 $reply.ExpectedBody=[Newtonsoft.Json.Linq.JObject]::Parse($reply.ExpectedBody).ToString([Newtonsoft.Json.Formatting]::Indented)
 $preflight=[Ext90Offline.PlannedReply]::new();$preflight.Method='GET';$preflight.AbsoluteUri='https://offline.example.invalid/Skype.Policy/configurations/ConversationRole/configuration/Global';$preflight.ExpectedBody='';$preflight.Status=[Net.HttpStatusCode]::Forbidden;$preflight.Json='{"code":"Forbidden"}'
 $handler.Replies.Enqueue($preflight)
 $handler.Replies.Enqueue($reply)
 $null=$pipeline.AddCommand($infos['New-CsConversationRole']).AddParameter('Identity','Global/00000000-0000-4000-8000-000000000001').AddParameter('Permissions',$permissions).AddParameter('CreatedAt',$created).AddParameter('UpdatedAt',$updated).AddParameter('Confirm',$false)
  }
  'role-set-identity' {
 $chatInfo=[Management.Automation.CmdletInfo]::new('New-CsTeamsChatPermissions',(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewChatPermissionsCmdlet'))
 $callingInfo=[Management.Automation.CmdletInfo]::new('New-CsTeamsCallingPermissions',(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewCallingPermissionsCmdlet'))
 $permissionsInfo=[Management.Automation.CmdletInfo]::new('New-CsTeamsConversationRolePermissions',(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewConversationRolePermissionsCmdlet'))
 $chat=& $chatInfo -AddParticipants $false -SendMessage $true
 $calling=& $callingInfo -AddParticipants $true -SendAudio $false
 $permissions=& $permissionsInfo -Chat $chat -Calling $calling
 $created=[DateTime]::SpecifyKind([DateTime]::new(2027,1,1),[DateTimeKind]::Unspecified)
 $updated=[DateTime]::SpecifyKind($created.AddDays(1),[DateTimeKind]::Utc)
 $reply=[Ext90Offline.PlannedReply]::new();$reply.Method='PATCH';$reply.AbsoluteUri='https://offline.example.invalid/Skype.Policy/configurations/ConversationRole/configuration/Global/00000000-0000-4000-8000-000000000001';$reply.ExpectedBody='{"Identity":"Global/00000000-0000-4000-8000-000000000001","Name":null}';$reply.Status=[Net.HttpStatusCode]::NoContent;$reply.Json=''
 $reply.ExpectedBody=[Newtonsoft.Json.Linq.JObject]::Parse($reply.ExpectedBody).ToString([Newtonsoft.Json.Formatting]::Indented)
 $preflight=[Ext90Offline.PlannedReply]::new();$preflight.Method='GET';$preflight.AbsoluteUri=$reply.AbsoluteUri;$preflight.ExpectedBody='';$preflight.Json='{"Identity":"Global/00000000-0000-4000-8000-000000000001","Id":"00000000-0000-4000-8000-000000000001","Name":"Current role","Permissions":{"Chat":{},"Calling":{}},"CreatedAt":"2027-01-01T00:00:00","UpdatedAt":"2027-01-02T00:00:00Z"}'
 $handler.Replies.Enqueue($preflight);$handler.Replies.Enqueue($reply)
 $null=$pipeline.AddCommand($infos['Set-CsConversationRole']).AddParameter('Identity','Global/00000000-0000-4000-8000-000000000001').AddParameter('Name',$null).AddParameter('Confirm',$false)
  }
  'role-set-relative' {
 $handler.MaxCalls=4;$fake.MaxTokenCalls=2
 $r=[Ext90Offline.PlannedReply]::new();$r.Method='GET';$r.AbsoluteUri='https://offline.example.invalid/Skype.Policy/configurations/ConversationRole/configuration/Global/00000000-0000-4000-8000-000000000001';$r.ExpectedBody=$null;$r.Json='{"Identity":"Global/00000000-0000-4000-8000-000000000001","Id":"00000000-0000-4000-8000-000000000009","Name":"Current role","Permissions":{"Chat":{},"Calling":{}},"CreatedAt":"2027-01-01T00:00:00","UpdatedAt":"2027-01-02T00:00:00Z"}';$r.Status=[Net.HttpStatusCode]::OK;$handler.Replies.Enqueue($r)
 $r=[Ext90Offline.PlannedReply]::new();$r.Method='PATCH';$r.AbsoluteUri='https://offline.example.invalid/Skype.Policy/configurations/ConversationRole/configuration/Global/00000000-0000-4000-8000-000000000001';$r.ExpectedBody=[Newtonsoft.Json.Linq.JObject]::Parse('{"Id":"00000000-0000-4000-8000-000000000009","Name":null}').ToString([Newtonsoft.Json.Formatting]::Indented);$r.Json='';$r.Status=[Net.HttpStatusCode]::NoContent;$handler.Replies.Enqueue($r)
 $null=$pipeline.AddCommand($infos['Set-CsConversationRole']).AddParameter('Id','00000000-0000-4000-8000-000000000001').AddParameter('Name',$null).AddParameter('Confirm',$false)
  }
  'role-remove-identity' {
 $handler.MaxCalls=4;$fake.MaxTokenCalls=2
 $r=[Ext90Offline.PlannedReply]::new();$r.Method='DELETE';$r.AbsoluteUri='https://offline.example.invalid/Skype.Policy/configurations/ConversationRole/configuration/Global/00000000-0000-4000-8000-000000000001';$r.ExpectedBody=[Newtonsoft.Json.Linq.JObject]::Parse('{"Identity":"Global/00000000-0000-4000-8000-000000000001","Force":"False"}').ToString([Newtonsoft.Json.Formatting]::Indented);$r.Json='';$r.Status=[Net.HttpStatusCode]::NoContent;$handler.Replies.Enqueue($r)
 $null=$pipeline.AddCommand($infos['Remove-CsConversationRole']).AddParameter('Identity','Global/00000000-0000-4000-8000-000000000001').AddParameter('Force',$false).AddParameter('Confirm',$false)
  }
  'role-remove-relative' {
 $handler.MaxCalls=4;$fake.MaxTokenCalls=2
 $r=[Ext90Offline.PlannedReply]::new();$r.Method='DELETE';$r.AbsoluteUri='https://offline.example.invalid/Skype.Policy/configurations/ConversationRole/configuration/Global/00000000-0000-4000-8000-000000000001';$r.ExpectedBody=[Newtonsoft.Json.Linq.JObject]::Parse('{"Id":"00000000-0000-4000-8000-000000000001","Force":"False"}').ToString([Newtonsoft.Json.Formatting]::Indented);$r.Json='';$r.Status=[Net.HttpStatusCode]::NoContent;$handler.Replies.Enqueue($r)
 $null=$pipeline.AddCommand($infos['Remove-CsConversationRole']).AddParameter('Id','00000000-0000-4000-8000-000000000001').AddParameter('Force',$false).AddParameter('Confirm',$false)
  }
  'setting-get-identity' {
 $handler.MaxCalls=4;$fake.MaxTokenCalls=2
 $r=[Ext90Offline.PlannedReply]::new();$r.Method='GET';$r.AbsoluteUri='https://offline.example.invalid/Skype.Policy/configurations/ConversationRolesSetting/configuration/Global';$r.ExpectedBody=$null;$r.Json='{"Identity":"Global","ConversationRoles":[]}';$r.Status=[Net.HttpStatusCode]::OK;$handler.Replies.Enqueue($r)
 $null=$pipeline.AddCommand($infos['Get-CsConversationRolesSetting']).AddParameter('Identity','Global')
  }
  'setting-get-filter' {
 $handler.MaxCalls=4;$fake.MaxTokenCalls=2
 $r=[Ext90Offline.PlannedReply]::new();$r.Method='GET';$r.AbsoluteUri='https://offline.example.invalid/Skype.Policy/configurations/ConversationRolesSetting';$r.ExpectedBody=$null;$r.Json='[{"Identity":"Global","ConversationRoles":[]}]';$r.Status=[Net.HttpStatusCode]::OK;$handler.Replies.Enqueue($r)
 $null=$pipeline.AddCommand($infos['Get-CsConversationRolesSetting']).AddParameter('Filter','Glo*')
  }
  'setting-set-replace' {
 $handler.MaxCalls=4;$fake.MaxTokenCalls=2
 $chatInfo=[Management.Automation.CmdletInfo]::new('New-CsTeamsChatPermissions',(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewChatPermissionsCmdlet'))
 $callingInfo=[Management.Automation.CmdletInfo]::new('New-CsTeamsCallingPermissions',(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewCallingPermissionsCmdlet'))
 $permissionsInfo=[Management.Automation.CmdletInfo]::new('New-CsTeamsConversationRolePermissions',(Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewConversationRolePermissionsCmdlet'))
 $chat=& $chatInfo -AddParticipants $false -SendMessage $true
 $calling=& $callingInfo -AddParticipants $true -SendAudio $false
 $permissions=& $permissionsInfo -Chat $chat -Calling $calling
 $created=[DateTime]::SpecifyKind([DateTime]::new(2027,1,1),[DateTimeKind]::Unspecified)
 $updated=[DateTime]::SpecifyKind($created.AddDays(1),[DateTimeKind]::Utc)
 $roleType=Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.ConversationRole'
 $replacement=[Activator]::CreateInstance($roleType,[object[]]@($permissions.psobject.BaseObject,$created,$updated));$replacement.Id='00000000-0000-4000-8000-000000000001';$replacement.Identity='Global/00000000-0000-4000-8000-000000000001';$replacement.Name='Controlled role'
 $modifierType=([Management.Automation.PSListModifier[int]]).GetGenericTypeDefinition().MakeGenericType(@($roleType));$modifier=[Activator]::CreateInstance($modifierType);$modifier.Replace.Add($replacement)
 $r=[Ext90Offline.PlannedReply]::new();$r.Method='GET';$r.AbsoluteUri='https://offline.example.invalid/Skype.Policy/configurations/ConversationRolesSetting/configuration/Global';$r.ExpectedBody=$null;$r.Json='{"Identity":"Global","ConversationRoles":[]}';$r.Status=[Net.HttpStatusCode]::OK;$handler.Replies.Enqueue($r)
 $r=[Ext90Offline.PlannedReply]::new();$r.Method='PATCH';$r.AbsoluteUri='https://offline.example.invalid/Skype.Policy/configurations/ConversationRolesSetting/configuration/Global';$r.ExpectedBody=[Newtonsoft.Json.Linq.JObject]::Parse('{"Identity":"Global","ConversationRoles":[{"Identity":"Global/00000000-0000-4000-8000-000000000001","Permissions":{"Identity":null},"Id":"00000000-0000-4000-8000-000000000001","Name":"Controlled role","IsAssignable":null,"CreatedAt":"2027-01-01T00:00:00","UpdatedAt":"2027-01-02T00:00:00Z"}]}').ToString([Newtonsoft.Json.Formatting]::Indented);$r.Json='';$r.Status=[Net.HttpStatusCode]::NoContent;$handler.Replies.Enqueue($r)
 $null=$pipeline.AddCommand($infos['Set-CsConversationRolesSetting']).AddParameter('Identity','Global').AddParameter('ConversationRoles',$modifier).AddParameter('Confirm',$false)
  }

 }
 $planned=$handler.Replies.Count
 $invokeException=$null
 try{$output=$pipeline.Invoke()}catch{$invokeException=$_.Exception.GetType().FullName}
 if(![Object]::ReferenceEquals($hostField.GetValue($null),$fake) -or ![Object]::ReferenceEquals($clientField.GetValue($null),$client) -or ![Object]::ReferenceEquals($loggerProperty.GetValue($null),$logger) -or $versionField.GetValue($null) -ne '0.0.0-offline'){throw 'Fake seams replaced during invocation'}
 $captured=@($handler.Captures|ForEach-Object{
  $lines=$_ -split "`n";$first=$lines[0] -split ' ',2
  $boundary=$_.LastIndexOf("`n`n");$rawBody=$_.Substring($boundary+2)
  @{method=$first[0];url=$first[1];cmdlet_header=($lines|Where-Object{$_.StartsWith('X-MS-CmdletName: ')}|ForEach-Object{$_.Substring(17)});body=$rawBody;body_sha256=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData([Text.Encoding]::UTF8.GetBytes($rawBody))).ToLowerInvariant()}
 })
 $pass=$null -eq $invokeException -and $pipeline.Streams.Error.Count -eq 0 -and $handler.Replies.Count -eq 0 -and $handler.Calls -eq $planned -and $fake.TokenCalls -eq 1
 @{schema_version=1;observation='Actual pinned compiled CmdletInfo binding/lifecycle with synthetic host/auth/logger/telemetry and terminal no-inner HTTP handler. Request construction only, not tenant/persistence/public-export proof. Synthetic replies are fixed in this reproducer; authorization/correlation headers omitted from capture.';scenario=$Scenario;outcome=$(if($pass){'passed'}else{'failed'});http_calls=$handler.Calls;token_calls=$fake.TokenCalls;module_calls=$fake.ModuleCalls;telemetry_calls=$telemetry.Events.Count;remaining_replies=$handler.Replies.Count;captures=$captured;output_count=$output.Count;output_json=@($output|ForEach-Object{[Newtonsoft.Json.JsonConvert]::SerializeObject($_.BaseObject)});warnings=@($pipeline.Streams.Warning|ForEach-Object{$_.Message});error_ids=@($pipeline.Streams.Error|ForEach-Object{$_.FullyQualifiedErrorId});invoke_exception=$invokeException;core_sha256=(Get-FileHash $core.Assembly.Location).Hash.ToLowerInvariant();policyrp_sha256=(Get-FileHash $provider.Assembly.Location).Hash.ToLowerInvariant();script_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant();fake_types_sha256=(Get-FileHash (Join-Path $PSScriptRoot 'conversation-offline-fakes.cs')).Hash.ToLowerInvariant()}|ConvertTo-Json -Depth 12 -Compress

}finally{
 if($pipeline){$pipeline.Dispose()};if($runspace){$runspace.Dispose()}
 $hostField.SetValue($null,$null);$clientField.SetValue($null,$null);$loggerProperty.SetValue($null,$oldLogger);$versionField.SetValue($null,$oldVersion)
 [Environment]::SetEnvironmentVariable('MSTeamsReleaseEnvironment',$oldRelease);[Environment]::SetEnvironmentVariable('MSTeamsContextInternal',$oldInternal)
 $client.Dispose()
 if($null -ne $hostField.GetValue($null) -or $null -ne $clientField.GetValue($null)){throw 'Static cleanup assertion failed'}
}
