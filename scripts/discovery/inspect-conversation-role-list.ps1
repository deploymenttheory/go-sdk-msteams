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
$role=Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.ConversationRole'
$modifierType=([System.Management.Automation.PSListModifier[int]]).GetGenericTypeDefinition().MakeGenericType(@($role))
$listType=([Collections.Generic.List[int]]).GetGenericTypeDefinition().MakeGenericType(@($role))
$lambda=Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.SetConversationRolesSettingCmdlet+<>c'
$lambdaMethod=$lambda.GetMethod('<UpdateProperties>b__2_0',[Reflection.BindingFlags]'Instance,Static,Public,NonPublic')
$predicateType=([Func[int,int,bool]]).GetGenericTypeDefinition().MakeGenericType(@($role,$role,[bool]))
$predicate=$lambdaMethod.CreateDelegate($predicateType,[Activator]::CreateInstance($lambda,$true))
$helper=Find-Type 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.PSListModifierHelper'
$method=$helper.GetMethods()|Where-Object{$_.Name -eq 'ApplyTo' -and $_.IsGenericMethodDefinition -and $_.GetParameters().Count -eq 3 -and $_.GetParameters()[2].ParameterType.Name -like 'Func*'}|Select-Object -First 1
$method=$method.MakeGenericMethod(@($role))
function New-Role([string]$id){$value=[Activator]::CreateInstance($role);$value.Id=$id;return $value}
$result=@()
foreach($current in @('nonempty','empty','null')){
foreach($scenario in @('null','empty','add','replace','remove-existing','remove-missing','add-and-remove','replace-and-remove','replace-and-add','all')){
 $list=[Activator]::CreateInstance($listType)
 if($current -eq 'nonempty'){$list.Add((New-Role '00000000-0000-4000-8000-000000000001'))}
 if($current -eq 'null'){$list=$null}
 $modifier=[Activator]::CreateInstance($modifierType)
 switch($scenario){
  'null' {$modifier=$null}
  'replace-and-remove' {$modifier.Replace.Add((New-Role '00000000-0000-4000-8000-000000000002'));$modifier.Remove.Add((New-Role '00000000-0000-4000-8000-000000000001'))}
  'replace-and-add' {$modifier.Replace.Add((New-Role '00000000-0000-4000-8000-000000000002'));$modifier.Add.Add((New-Role '00000000-0000-4000-8000-000000000003'))}
  'all' {$modifier.Replace.Add((New-Role '00000000-0000-4000-8000-000000000002'));$modifier.Add.Add((New-Role '00000000-0000-4000-8000-000000000003'));$modifier.Remove.Add((New-Role '00000000-0000-4000-8000-000000000001'))}
  'add' {$modifier.Add.Add((New-Role '00000000-0000-4000-8000-000000000002'))}
  'replace' {$modifier.Replace.Add((New-Role '00000000-0000-4000-8000-000000000002'))}
  'remove-existing' {$modifier.Remove.Add((New-Role '00000000-0000-4000-8000-000000000001'))}
  'remove-missing' {$modifier.Remove.Add((New-Role '00000000-0000-4000-8000-000000000002'))}
  'add-and-remove' {$modifier.Add.Add((New-Role '00000000-0000-4000-8000-000000000002'));$modifier.Remove.Add((New-Role '00000000-0000-4000-8000-000000000001'))}
 }
 try{$null=$method.Invoke($null,[object[]]@($modifier,$list,$predicate));$result+=@{current=$current;scenario=$scenario;accepted=$true;count=$list.Count;ids=@($list|ForEach-Object{$_.Id})}}
 catch{$result+=@{current=$current;scenario=$scenario;accepted=$false;count=$list.Count;failure_type=$_.Exception.InnerException.GetType().FullName}}
}
}
@{schema_version=1;baseline_version='8.1.0';extractor_sha256=(Get-FileHash -LiteralPath $PSCommandPath -Algorithm SHA256).Hash.ToLowerInvariant();observation='Offline data models and PSListModifierHelper with the exact compiled ConversationRolesSetting removal predicate; no cmdlet instance, sign-in or tenant operation. Failure types reflect the offline exception/resource path, not a live server error.';cases=$result}|ConvertTo-Json -Depth 8 -Compress
