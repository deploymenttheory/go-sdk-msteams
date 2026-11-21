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
foreach($verb in @('Grant')){
 $entry=$configuration.CmdletConfiguration["$verb-CsTeamsFeedbackPolicy"]
 $modern=if($verb -eq 'Grant'){'Grant-CsTeamsPolicy'}else{"$verb-CsConfigurationModern"}
 $key=if($verb -eq 'Grant'){'PolicyType'}else{'ConfigType'}
 if($entry.CmdletType -ne 'Remoting' -or $entry.ModernCmdlet -ne $modern -or $entry.DefaultAutoRestParameters[$key] -ne 'TeamsFeedbackPolicy'){throw 'Feedback policy Grant mapping differs.'}
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
$state=@{calls=[Collections.Generic.List[object]]::new()}
$stub=New-Module -Name 'Microsoft.Teams.ConfigAPI.Cmdlets.internal' -ArgumentList $state -ScriptBlock {
 param($state)
 $script:state=$state
 function Record($operation,$parameters){
  $copy=$parameters|ConvertTo-Json -Depth 12|ConvertFrom-Json -AsHashtable
  $script:state.calls.Add(@{operation=$operation;parameters=$copy})
 }
 function Grant-CsUserPolicy {[CmdletBinding()]param($Identity,$PolicyType,[string]$PolicyName,$AdditionalParameters) Record 'assign-user' $PSBoundParameters}
 function Grant-CsTenantPolicy {[CmdletBinding()]param($PolicyType,[string]$PolicyName,$AdditionalParameters,[bool]$forceSwitchPresent) Record 'assign-tenant' $PSBoundParameters}
 function Grant-CsGroupPolicyAssignment {[CmdletBinding()]param($GroupId,$PolicyType,[string]$PolicyName,$Rank) Record 'assign-group' $PSBoundParameters}
 Export-ModuleMember -Function *-Cs*
}
Import-Module $stub
function Convert-PsCustomObjectToPsObject($value){return $value}
function Set-FormatOnConfigObject($ConfigObject,$ConfigType){return $ConfigObject}
$names=@('Grant-CsTeamsPolicy')
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
 $state.calls.Clear()

 $values=@();$failure=$null
 try{$values=@(& $invoke)}catch{$failure=$_.Exception.GetBaseException().GetType().FullName}
 $cases.Add(@{name=$name;calls=@($state.calls.ToArray());output_identities=@($values|ForEach-Object{$_.Identity});failure_type=$failure})
}
Observe 'grant-user' {Grant-CsTeamsPolicy -PolicyType TeamsFeedbackPolicy -Identity 'fixture/user' -PolicyName 'Fixture'}
Observe 'clear-user' {Grant-CsTeamsPolicy -PolicyType TeamsFeedbackPolicy -Identity 'fixture/user' -PolicyName $null}
Observe 'grant-tenant' {Grant-CsTeamsPolicy -PolicyType TeamsFeedbackPolicy -Global -PolicyName 'Fixture'}
Observe 'grant-group-rank-omitted' {Grant-CsTeamsPolicy -PolicyType TeamsFeedbackPolicy -Group 'fixture/group' -PolicyName 'Fixture'}
Observe 'grant-group-rank-zero' {Grant-CsTeamsPolicy -PolicyType TeamsFeedbackPolicy -Group 'fixture/group' -PolicyName 'Fixture' -Rank 0}
Observe 'grant-group-rank-two' {Grant-CsTeamsPolicy -PolicyType TeamsFeedbackPolicy -Group 'fixture/group' -PolicyName 'Fixture' -Rank 2}
Observe 'grant-group-rank-negative' {Grant-CsTeamsPolicy -PolicyType TeamsFeedbackPolicy -Group 'fixture/group' -PolicyName 'Fixture' -Rank -1}
Observe 'grant-missing-policy-name' {Grant-CsTeamsPolicy -PolicyType TeamsFeedbackPolicy -Global}
Observe 'grant-missing-scope' {Grant-CsTeamsPolicy -PolicyType TeamsFeedbackPolicy -PolicyName 'Fixture'}
Observe 'grant-passthru-true-rejected' {Grant-CsTeamsPolicy -PolicyType TeamsFeedbackPolicy -Global -PolicyName 'Fixture' -PassThru}
Observe 'grant-passthru-false-rejected' {Grant-CsTeamsPolicy -PolicyType TeamsFeedbackPolicy -Global -PolicyName 'Fixture' -PassThru:$false}
[ordered]@{schema_version=1;baseline_version='8.1.0';qualification='isolated-pinned-feedback-policy-grant-wrapper-execution';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd');package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7';source=@{path='custom/Merged_custom_PsExt.ps1';sha256=$hash};methods=$names;seams=@('synthetic CustomCmdletUtils and SendAsyncStep types','in-memory internal command module','formatting adapters return objects unchanged');limitations=@('No exported TrpsPSCmdlet, flight selection, compiled generated HTTP pipeline or tenant was executed.','Parameter and composition observations apply to these unmodified mapped functions under the declared stubs.');cases=@($cases.ToArray())}|ConvertTo-Json -Depth 20|Set-Content -LiteralPath $OutputPath -Encoding utf8
"feedback policy grant wrapper observations: $($cases.Count)"
