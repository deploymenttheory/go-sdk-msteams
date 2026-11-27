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
foreach($case in @(
 @{model='ConversationRole';body='{}'},
 @{model='ConversationRole';body='{"Permissions":null}'},
 @{model='ConversationRole';body='{"Permissions":{"Chat":null}}'},
 @{model='ConversationRole';body='{"CreatedAt":null}'},
 @{model='ConversationRolesSetting';body='{}'},
 @{model='ConversationRolesSetting';body='{"ConversationRoles":null}'},
 @{model='ConversationRolesSetting';body='{"ConversationRoles":[]}'},
 @{model='ConversationRolesSetting';body='{"ConversationRoles":[null]}'}
)){
 $type=Find-Type ('Microsoft.Teams.Policy.Administration.Cmdlets.Core.'+$case.model)
 try{$value=[Newtonsoft.Json.JsonConvert]::DeserializeObject($case.body,$type);$result+=@{model=$case.model;input=$case.body;accepted=$true;output=[Newtonsoft.Json.JsonConvert]::SerializeObject($value)}}
 catch{$result+=@{model=$case.model;input=$case.body;accepted=$false;failure=$_.Exception.GetType().FullName}}
}
@{schema_version=1;baseline_version='8.1.0';extractor_sha256=(Get-FileHash -LiteralPath $PSCommandPath -Algorithm SHA256).Hash.ToLowerInvariant();observation='Offline Newtonsoft JSON deserialization into pinned CLR data models only; no cmdlet, authentication or HTTP. Captures are model vectors, not tenant replies.';cases=$result}|ConvertTo-Json -Depth 8 -Compress
