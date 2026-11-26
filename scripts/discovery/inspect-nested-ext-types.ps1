#requires -Version 7.0
# Read-only metadata observation. The caller supplies an extracted, verified
# MicrosoftTeams 8.1.0 package; this script never connects to a tenant.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference='Stop'
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong module baseline'}
$pins=@{
 'Microsoft.Teams.PowerShell.TeamsCmdlets.dll'='3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24'
 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll'='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'
}
foreach($file in $pins.Keys){
 if((Get-FileHash -LiteralPath (Join-Path (Join-Path $ModulePath 'netcoreapp3.1') $file) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $pins[$file]){throw 'Wrong pinned assembly'}
}
Import-Module $manifest -Force -WarningAction SilentlyContinue
$names=@('Add-TeamsAppInstallation','Get-CsConversationRole','Get-CsConversationRolesSetting','Get-CsTeamsAppPolicyConfiguration','Get-CsTenantConfiguration','Get-LicenseReportForChangeNotificationSubscription','Get-TeamsAppInstallation','New-CsConversationRole','Remove-CsConversationRole','Remove-TeamsAppInstallation','Set-CsConversationRole','Set-CsConversationRolesSetting','Set-CsTeamsAppPolicyConfiguration','Set-CsTenantConfiguration','Update-TeamsAppInstallation')
$types=@()
foreach($assembly in [AppDomain]::CurrentDomain.GetAssemblies()){
 try{$types+=$assembly.GetTypes()}catch{}
}
$result=@()
foreach($name in $names){
 $matched=@()
 foreach($type in $types){
  foreach($attribute in $type.GetCustomAttributes([System.Management.Automation.CmdletAttribute],$false)){
   if(($attribute.VerbName+'-'+$attribute.NounName) -eq $name){
    $parameters=@($type.GetProperties()|Where-Object{$_.GetCustomAttributes([System.Management.Automation.ParameterAttribute],$true).Count -gt 0}|ForEach-Object{@{name=$_.Name;type=$_.PropertyType.FullName}})
    $matched+=@{type=$type.FullName;base=$type.BaseType.FullName;assembly=$type.Assembly.FullName;assembly_file=[IO.Path]::GetFileName($type.Assembly.Location);assembly_sha256=(Get-FileHash -LiteralPath $type.Assembly.Location -Algorithm SHA256).Hash.ToLowerInvariant();parameters=$parameters}
   }
  }
 }
 $result+=@{name=$name;types=$matched;root_exported=($null -ne (Get-Command -Name $name -Module MicrosoftTeams -ErrorAction SilentlyContinue))}
}
$result|ConvertTo-Json -Depth 8 -Compress
