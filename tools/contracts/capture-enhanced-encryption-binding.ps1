# Isolated public parameter binding and pure Xds serialization; no command body.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$sources=@(
 @{path='exports/ProxyCmdletDefinitionsWithHelp.ps1';sha256='5a8a5a50081b1c6986b93d5b15723e966a2d22b64459a9c300cb2a1bf69bb3b5'},
 @{path='bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll';sha256='8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'}
)
foreach($s in $sources){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $s.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $s.sha256){throw 'Wrong pinned binding dependency.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
Import-Module $manifest -Force
$tokens=$null;$parseErrors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile((Join-Path $ModulePath $sources[0].path),[ref]$tokens,[ref]$parseErrors)
if($parseErrors.Count){throw 'Pinned export parse failed.'}
$cases=[Collections.Generic.List[object]]::new()
$privateMetadata=[Collections.Generic.List[object]]::new()
$privateIdentityBindings=[Collections.Generic.List[object]]::new()
foreach($verb in @('New','Set')){
 $command="$verb-CsTeamsEnhancedEncryptionPolicy"
 $type=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.XdsConfiguration].Assembly.GetType("Microsoft.Teams.ConfigApi.Cmdlets.${verb}CsTeamsEnhancedEncryptionPolicy",$true)
 $attribute=@($type.GetProperty('CallingEndtoEndEncryptionEnabledType').GetCustomAttributes($true)|Where-Object{$_.GetType().Name -like 'CmdletParamMetainfo*'})
 if($attribute.Count -ne 1){throw 'Private calling mode metadata differs.'}
 $method=$type.GetMethod('ValidateAndUpdatePropertyBagParams')
 $il=$method.GetMethodBody().GetILAsByteArray()
 # Only the verified inherited empty method may be called without constructing
 # the original command. An unexpected override/body aborts this pure seam.
 if($method.DeclaringType.FullName -ne 'Microsoft.Teams.ConfigApi.Cmdlets.TrpsPSCmdlet' -or ($il -join ',') -cne '0,43,0,42'){throw 'Unexpected property bag validation body.'}
 $bag=[Collections.Generic.Dictionary[string,object]]::new();$bag.Add('CallingEndtoEndEncryptionEnabledType','FutureCalling')
 $instance=[Runtime.CompilerServices.RuntimeHelpers]::GetUninitializedObject($type)
 $instance.ValidateAndUpdatePropertyBagParams($bag)
 $privateMetadata.Add(@{command=$command;calling_mode_valid_set=@($attribute[0].ValidSet);property_bag_validator=$method.DeclaringType.FullName;validator_body='verified-empty-il-nop-branch-ret';future_value_unchanged=($bag.Count -eq 1 -and $bag['CallingEndtoEndEncryptionEnabledType'] -ceq 'FutureCalling')})
 $script:identityAttributes=[Collections.ObjectModel.Collection[Attribute]]::new()
 foreach($a in $type.GetProperty('Identity').GetCustomAttributes($true)){
  if($a -is [Management.Automation.ParameterAttribute] -or $a -is [Management.Automation.ValidateArgumentsAttribute]){$script:identityAttributes.Add($a)}
 }
 function Invoke-EncryptionPrivateIdentityFixture {
  [CmdletBinding(DefaultParameterSetName='Identity')]param()
  dynamicparam {
   $parameters=[Management.Automation.RuntimeDefinedParameterDictionary]::new()
   $parameters.Add('Identity',[Management.Automation.RuntimeDefinedParameter]::new('Identity',[string],$script:identityAttributes))
   return $parameters
  }
  process {'accepted'}
 }
 foreach($case in @(@{name='null';value=$null},@{name='empty';value=''},@{name='whitespace';value=' '},@{name='selected';value='Fixture'})){
  $failure=$null
  try{$null=Invoke-EncryptionPrivateIdentityFixture -Identity $case.value}catch{$failure=$_.Exception.GetBaseException().GetType().FullName}
  $privateIdentityBindings.Add(@{command=$command;name=$case.name;failure_type=$failure})
 }
 Remove-Item Function:Invoke-EncryptionPrivateIdentityFixture
 $definition=$ast.Find({param($node)$node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq $command},$true)
 if($null -eq $definition){throw 'Pinned export missing.'}
 $attributes=($definition.Body.ParamBlock.Attributes|ForEach-Object{$_.Extent.Text}) -join "`n"
 $parameters=$definition.Body.ParamBlock.Extent.Text
 # Only the original declaration and attributes are used. None of the export's
 # Begin/Process/End code or custom runtime-attribute processor is invoked.
 . ([scriptblock]::Create("function Invoke-EncryptionBindingFixture {`n$attributes`n$parameters`n process { ConvertTo-Json -InputObject `$PSBoundParameters -Depth 10 -Compress }`n}"))
 $bindingCases=@(
  @{name='omitted-settings';arguments=@{Identity='Fixture'}},
  @{name='empty-strings';arguments=@{Identity='Fixture';Description='';CallingEndtoEndEncryptionEnabledType='';MeetingEndToEndEncryption=''}},
  @{name='future-strings';arguments=@{Identity='Fixture';Description=('x'*2000);CallingEndtoEndEncryptionEnabledType='FutureCalling';MeetingEndToEndEncryption='FutureMeeting'}},
  @{name='bound-null-fields';arguments=@{Identity='Fixture';Description=$null;CallingEndtoEndEncryptionEnabledType=$null;MeetingEndToEndEncryption=$null}},
  @{name='numeric-string';arguments=@{Identity='Fixture';CallingEndtoEndEncryptionEnabledType=0}},
  @{name='boolean-string';arguments=@{Identity='Fixture';MeetingEndToEndEncryption=$false}},
  @{name='empty-identity';arguments=@{Identity='';Description=''}},
  @{name='null-identity';arguments=@{Identity=$null;Description=''}}
 )
 foreach($field in @('Description','CallingEndtoEndEncryptionEnabledType','MeetingEndToEndEncryption')){
  $arguments=@{Identity='Fixture'};$arguments[$field]=$null
  $bindingCases+=@{name=('null-'+$field);arguments=$arguments}
 }
 foreach($c in $bindingCases){
  $failure=$null;$bound=$null;$wire=$null
  try{
   $arguments=$c.arguments
   $bound=Invoke-EncryptionBindingFixture @arguments | ConvertFrom-Json -AsHashtable
   $model=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.XdsConfiguration]::DeserializeFromDictionary($bound)
   $wire=$model.ToJsonString()|ConvertFrom-Json -AsHashtable
  }catch{$failure=$_.Exception.GetBaseException().GetType().FullName}
  $cases.Add(@{command=$command;name=$c.name;bound=$bound;wire=$wire;failure_type=$failure})
 }
 Remove-Item Function:Invoke-EncryptionBindingFixture
}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='isolated-enhanced-encryption-public-parameter-binding';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 sources=$sources;host_powershell_version=$PSVersionTable.PSVersion.ToString()
 stage='Unmodified exported New/Set parameter declarations and attributes in a renamed function with a synthetic body returning bound arguments; actual pure XdsConfiguration dictionary serialization. Private calling-mode ValidSet metadata read separately; actual private Identity parameter/validation attributes are reused in a synthetic RuntimeDefinedParameter binder without original command construction or processing. Inherited property-bag validator IL restricted to the pinned empty nop/br.s-0/ret body, invoked on an uninitialized command solely to observe unchanged future text. No original constructor, export Begin/Process/End, runtime custom-attribute processing, TRPS/flight selection, command lifecycle, credentials or network.'
 cases=$cases.ToArray();private_metadata=$privateMetadata.ToArray();private_identity_bindings=$privateIdentityBindings.ToArray()
}|ConvertTo-Json -Depth 25|Set-Content -LiteralPath $OutputPath -Encoding utf8
"enhanced encryption isolated binding observations: $($cases.Count)"
