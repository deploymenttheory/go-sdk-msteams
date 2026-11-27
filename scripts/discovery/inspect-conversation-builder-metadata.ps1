#requires -Version 7.0
# Reflection only. CmdletInfo construction reads declarations; no cmdlet,
# authentication, dynamic parameter provider or tenant operation is invoked.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference='Stop'
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong module baseline'}
Import-Module $manifest -Force -WarningAction SilentlyContinue
foreach($assembly in [AppDomain]::CurrentDomain.GetAssemblies()){try{$null=$assembly.GetTypes()}catch{}}
$corePath=Join-Path (Join-Path $ModulePath 'netcoreapp3.1') 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll'
$coreHash=(Get-FileHash -LiteralPath $corePath -Algorithm SHA256).Hash.ToLowerInvariant()
if($coreHash -ne '68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'){throw 'Wrong pinned core assembly'}
$observed=@()
foreach($item in @(@{name='New-CsTeamsChatPermissions';type='NewChatPermissionsCmdlet'},@{name='New-CsTeamsCallingPermissions';type='NewCallingPermissionsCmdlet'},@{name='New-CsTeamsConversationRolePermissions';type='NewConversationRolePermissionsCmdlet'})){
 $observed+=@{name=$item.name;root_exported=(@(Get-Command -Name $item.name -ErrorAction SilentlyContinue).Count -gt 0);types=@(@{type='Microsoft.Teams.Policy.Administration.Cmdlets.Core.'+$item.type;assembly_file='Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';assembly_sha256=$coreHash})}
}
$manifests=@()
foreach($file in @('MicrosoftTeams.psd1','Microsoft.Teams.Policy.Administration.Cmdlets.Core.oce.psd1','Microsoft.Teams.Policy.Administration.Cmdlets.Core.preview.psd1')){
 $path=Join-Path $ModulePath $file;$data=Import-PowerShellDataFile -LiteralPath $path
 $manifests+=@{file=$file;sha256=(Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash.ToLowerInvariant();cmdlet_wildcard_patterns=@($data.CmdletsToExport|Where-Object{[System.Management.Automation.WildcardPattern]::ContainsWildcardCharacters($_)});builders=@($observed|ForEach-Object{@{name=$_.name;explicit_cmdlet_export=($data.CmdletsToExport -contains $_.name)}})}
}
$records=@()
foreach($record in $observed){
 foreach($match in $record.types){
  $type=[AppDomain]::CurrentDomain.GetAssemblies()|ForEach-Object{$_.GetType($match.type,$false)}|Where-Object{$null -ne $_}|Select-Object -First 1
  $metadata=[System.Management.Automation.CommandMetadata]::new($type)
  $cmdlet=[System.Management.Automation.CmdletInfo]::new($record.name,$type)
  $declarations=@()
  foreach($property in $type.GetProperties()){
   $attributes=@($property.GetCustomAttributes([System.Management.Automation.ParameterAttribute],$true))
   if($attributes.Count -eq 0){continue}
   $parameters=@($attributes|ForEach-Object{@{set=$_.ParameterSetName;mandatory=$_.Mandatory;position=$_.Position;pipeline=$_.ValueFromPipeline;pipeline_property=$_.ValueFromPipelineByPropertyName;remaining_arguments=$_.ValueFromRemainingArguments}})
   $validates=@($property.GetCustomAttributes([System.Management.Automation.ValidateSetAttribute],$true)|ForEach-Object{@{values=@($_.ValidValues);ignore_case=$_.IgnoreCase}})
   $declarations+=@{name=$property.Name;type=$property.PropertyType.FullName;parameter_attributes=$parameters;validate_sets=$validates;declared_attributes=@($property.CustomAttributes|ForEach-Object{[string]$_})}
  }
  $sets=@($cmdlet.ParameterSets|ForEach-Object{@{name=$_.Name;default=$_.IsDefault;parameters=@($_.Parameters|ForEach-Object{@{name=$_.Name;type=$_.ParameterType.FullName;mandatory=$_.IsMandatory;position=$_.Position;pipeline=$_.ValueFromPipeline;pipeline_property=$_.ValueFromPipelineByPropertyName;remaining_arguments=$_.ValueFromRemainingArguments;aliases=@($_.Aliases)}})}})
  $records+=@{name=$record.name;type=$type.FullName;assembly_file=$match.assembly_file;assembly_sha256=$match.assembly_sha256;root_exported=$record.root_exported;default_parameter_set=$metadata.DefaultParameterSetName;parameter_sets=$sets;property_declarations=$declarations}
 }
}
@{schema_version=1;baseline_version='8.1.0';extractor_sha256=(Get-FileHash -LiteralPath $PSCommandPath -Algorithm SHA256).Hash.ToLowerInvariant();package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7';observation='Supplemental compiled local permission-builder declaration metadata; CmdletInfo and CommandMetadata construction only. No invocation, dynamic parameter evaluation, sign-in or tenant operation. Separate from immutable default-session runtime evidence.';manifest_observations=$manifests;records=$records}|ConvertTo-Json -Depth 12 -Compress
