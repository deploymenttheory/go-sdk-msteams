#requires -Version 7.0
# Reflection only. CmdletInfo construction reads declarations; no cmdlet,
# authentication, dynamic parameter provider or tenant operation is invoked.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference='Stop'
$observed=& (Join-Path $PSScriptRoot 'inspect-nested-ext-types.ps1') -ModulePath $ModulePath | ConvertFrom-Json
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
@{schema_version=1;baseline_version='8.1.0';extractor_sha256=(Get-FileHash -LiteralPath $PSCommandPath -Algorithm SHA256).Hash.ToLowerInvariant();package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7';observation='Pinned compiled-type declaration metadata; CmdletInfo and CommandMetadata construction only. No invocation, dynamic parameter evaluation, sign-in or tenant operation. Separate from immutable default-session runtime evidence.';records=$records}|ConvertTo-Json -Depth 12 -Compress
