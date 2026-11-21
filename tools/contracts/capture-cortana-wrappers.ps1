# Offline pinned-source wrapper oracle. No module import, authentication or HTTP.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$source=Join-Path $ModulePath 'custom/Merged_custom_PsExt.ps1'
$hash='03dbc3a85e12c98dd1807d604782c6a90d0656569541cd9540738825a7fcb255'
if((Get-FileHash -LiteralPath $source -Algorithm SHA256).Hash.ToLowerInvariant() -ne $hash){throw 'Wrong pinned wrapper source.'}
if((Import-PowerShellDataFile -LiteralPath (Join-Path $ModulePath 'MicrosoftTeams.psd1')).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
$configurationPath='custom/CmdletConfig.json'
$configurationHash='754d149c1a5904158e9f01eb55ba44d15c17b7e66f7556385ea894f4ff84d906'
if((Get-FileHash -LiteralPath (Join-Path $ModulePath $configurationPath) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $configurationHash){throw 'Wrong pinned command mappings.'}
$configuration=Get-Content -LiteralPath (Join-Path $ModulePath $configurationPath) -Raw|ConvertFrom-Json -AsHashtable
foreach($verb in @('Get','New','Set','Remove','Grant')){
 $entry=$configuration.CmdletConfiguration["$verb-CsTeamsCortanaPolicy"]
 $modern=if($verb -eq 'Grant'){'Grant-CsTeamsPolicy'}else{"$verb-CsConfigurationModern"}
 $key=if($verb -eq 'Grant'){'PolicyType'}else{'ConfigType'}
 if($entry.CmdletType -ne 'Remoting' -or $entry.ModernCmdlet -ne $modern -or $entry.DefaultAutoRestParameters[$key] -ne 'TeamsCortanaPolicy'){throw 'Cortana mapping differs.'}
}
# Synthetic host types make the original functions executable without the module.
# The telemetry and internal-command seams have no networking implementation.
Add-Type -TypeDefinition @'
using System.Collections;
namespace Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Runtime { public class SendAsyncStep {} }
namespace Microsoft.Teams.ConfigAPI.Cmdlets.Telemetry {
 public class CustomCmdletUtils {
  public CustomCmdletUtils(object invocation) {}
  public Hashtable ProcessArgs() { return new Hashtable(); }
  public void PutHttpPipelineSteps(object arguments) {}
  public void SendTelemetry() {}
 }
}
'@
$state=@{calls=[Collections.Generic.List[object]]::new();policies=@();failCreate=$false}
$stub=New-Module -Name 'Microsoft.Teams.ConfigAPI.Cmdlets.internal' -ArgumentList $state -ScriptBlock {
 param($state)
 $script:state=$state
 function Record($operation,$parameters){
  $copy=$parameters|ConvertTo-Json -Depth 12|ConvertFrom-Json -AsHashtable
  $script:state.calls.Add(@{operation=$operation;parameters=$copy})
 }
 function Get-CsConfiguration {
  [CmdletBinding()]param($ConfigType,$ConfigName)
  Record 'get' $PSBoundParameters
  $values=$script:state.policies
  if($PSBoundParameters.ContainsKey('ConfigName')){$values=@($values|Where-Object{$_.Identity -eq $ConfigName})}
  foreach($value in $values){$value|ConvertTo-Json -Depth 12 -Compress}
 }
 function New-CsConfiguration {
  [CmdletBinding()]param($ConfigType,$Body)
  Record 'create' $PSBoundParameters
  if($script:state.failCreate){throw 'synthetic-create-failure'}
 }
 function Set-CsConfiguration {[CmdletBinding()]param($ConfigType,$ConfigName,$Body) Record 'update' $PSBoundParameters}
 function Remove-CsConfiguration {[CmdletBinding()]param($ConfigType,$ConfigName) Record 'delete' $PSBoundParameters}
 function Grant-CsUserPolicy {[CmdletBinding()]param($Identity,$PolicyType,[string]$PolicyName,$AdditionalParameters) Record 'assign-user' $PSBoundParameters}
 function Grant-CsTenantPolicy {[CmdletBinding()]param($PolicyType,[string]$PolicyName,$AdditionalParameters,[bool]$forceSwitchPresent) Record 'assign-tenant' $PSBoundParameters}
 function Grant-CsGroupPolicyAssignment {[CmdletBinding()]param($GroupId,$PolicyType,[string]$PolicyName,$Rank) Record 'assign-group' $PSBoundParameters}
 Export-ModuleMember -Function *-Cs*
}
Import-Module $stub
function Convert-PsCustomObjectToPsObject($value){return $value}
function Set-FormatOnConfigObject($ConfigObject,$ConfigType){return $ConfigObject}
$names=@('Sort-GlobalFirst','Get-CsConfigurationModern','New-CsConfigurationModern','Set-CsConfigurationModern','Remove-CsConfigurationModern','Grant-CsTeamsPolicy')
$tokens=$null;$parseErrors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile($source,[ref]$tokens,[ref]$parseErrors)
if($parseErrors.Count){throw 'Pinned source parse failed.'}
foreach($name in $names){
 $definition=$ast.Find({param($node) $node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq $name},$true)
 if($null -eq $definition){throw 'Pinned wrapper missing.'}
 . ([scriptblock]::Create($definition.Extent.Text))
}
$cases=[Collections.Generic.List[object]]::new()
function Observe($name,$invoke){
 $state.calls.Clear();$state.failCreate=$false
 $state.policies=@(@{Identity='Other';AllowCortanaVoiceInvocation=$false},@{Identity='gLoBaL';AllowCortanaVoiceInvocation=$false},@{Identity='Global';AllowCortanaVoiceInvocation=$false},@{Identity='Tag:Global';AllowCortanaVoiceInvocation=$false},@{Identity='Fixture';Description='';CortanaVoiceInvocationMode='FutureMode';AllowCortanaVoiceInvocation=$false;AllowCortanaAmbientListening=$true;AllowCortanaInContextSuggestions=$false})
 $values=@();$failure=$null
 try{$values=@(& $invoke)}catch{$failure=$_.Exception.GetBaseException().GetType().FullName}
 $cases.Add(@{name=$name;calls=@($state.calls.ToArray());output_identities=@($values|ForEach-Object{$_.Identity});failure_type=$failure})
}
Observe 'list-global-first' {Get-CsConfigurationModern -ConfigType TeamsCortanaPolicy}
Observe 'filter-global-first' {Get-CsConfigurationModern -ConfigType TeamsCortanaPolicy -Filter '*'}
Observe 'filter-tag-prefix' {Get-CsConfigurationModern -ConfigType TeamsCortanaPolicy -Filter 'Global'}
Observe 'create-readback' {New-CsConfigurationModern -ConfigType TeamsCortanaPolicy -PropertyBag @{Identity='Fixture';Description='';CortanaVoiceInvocationMode='FutureMode';AllowCortanaVoiceInvocation=$false;AllowCortanaAmbientListening=$true;AllowCortanaInContextSuggestions=$false}}
Observe 'create-failure-stops-readback' {$state.failCreate=$true;New-CsConfigurationModern -ConfigType TeamsCortanaPolicy -PropertyBag @{Identity='Fixture'}}
Observe 'update-omitted-identity-global' {Set-CsConfigurationModern -ConfigType TeamsCortanaPolicy -PropertyBag @{AllowCortanaVoiceInvocation=$false}}
Observe 'update-cortana-fields' {Set-CsConfigurationModern -ConfigType TeamsCortanaPolicy -PropertyBag @{Identity='Fixture';Description='';CortanaVoiceInvocationMode='FutureMode';AllowCortanaVoiceInvocation=$false;AllowCortanaAmbientListening=$true;AllowCortanaInContextSuggestions=$false}}
Observe 'filter-whitespace-returns-all' {Get-CsConfigurationModern -ConfigType TeamsCortanaPolicy -Filter ' '}
Observe 'identity-whitespace-selects-collection' {Get-CsConfigurationModern -ConfigType TeamsCortanaPolicy -Identity ' '}
Observe 'update-explicit-identity' {Set-CsConfigurationModern -ConfigType TeamsCortanaPolicy -PropertyBag @{Identity='Fixture';AllowCortanaVoiceInvocation=$false}}
Observe 'delete' {Remove-CsConfigurationModern -ConfigType TeamsCortanaPolicy -Identity 'Fixture'}
Observe 'grant-user' {Grant-CsTeamsPolicy -PolicyType TeamsCortanaPolicy -Identity 'fixture/user' -PolicyName 'Fixture'}
Observe 'clear-user' {Grant-CsTeamsPolicy -PolicyType TeamsCortanaPolicy -Identity 'fixture/user' -PolicyName $null}
Observe 'grant-tenant' {Grant-CsTeamsPolicy -PolicyType TeamsCortanaPolicy -Global -PolicyName 'Fixture'}
Observe 'grant-group-rank-omitted' {Grant-CsTeamsPolicy -PolicyType TeamsCortanaPolicy -Group 'fixture/group' -PolicyName 'Fixture'}
Observe 'grant-group-rank-zero' {Grant-CsTeamsPolicy -PolicyType TeamsCortanaPolicy -Group 'fixture/group' -PolicyName 'Fixture' -Rank 0}
Observe 'grant-group-rank-two' {Grant-CsTeamsPolicy -PolicyType TeamsCortanaPolicy -Group 'fixture/group' -PolicyName 'Fixture' -Rank 2}
Observe 'grant-group-rank-negative' {Grant-CsTeamsPolicy -PolicyType TeamsCortanaPolicy -Group 'fixture/group' -PolicyName 'Fixture' -Rank -1}
Observe 'grant-missing-policy-name' {Grant-CsTeamsPolicy -PolicyType TeamsCortanaPolicy -Global}
Observe 'grant-missing-scope' {Grant-CsTeamsPolicy -PolicyType TeamsCortanaPolicy -PolicyName 'Fixture'}
Observe 'grant-passthru-true-rejected' {Grant-CsTeamsPolicy -PolicyType TeamsCortanaPolicy -Global -PolicyName 'Fixture' -PassThru}
Observe 'grant-passthru-false-rejected' {Grant-CsTeamsPolicy -PolicyType TeamsCortanaPolicy -Global -PolicyName 'Fixture' -PassThru:$false}
[ordered]@{schema_version=1;baseline_version='8.1.0';qualification='isolated-pinned-cortana-policy-wrapper-execution';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd');package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7';source=@{path='custom/Merged_custom_PsExt.ps1';sha256=$hash};methods=$names;seams=@('synthetic CustomCmdletUtils and SendAsyncStep types','in-memory internal command module','formatting adapters return objects unchanged');limitations=@('No exported TrpsPSCmdlet, flight selection, compiled generated HTTP pipeline or tenant was executed.','Parameter and composition observations apply to these unmodified mapped functions under the declared stubs.');cases=@($cases.ToArray())}|ConvertTo-Json -Depth 20|Set-Content -LiteralPath $OutputPath -Encoding utf8
"cortana policy wrapper observations: $($cases.Count)"
