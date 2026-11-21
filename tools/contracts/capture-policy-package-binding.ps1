# Execute unmodified custom functions with telemetry/internal-command stubs only.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$source=Join-Path $ModulePath 'custom/Merged_custom_PsExt.ps1'
$hash='03dbc3a85e12c98dd1807d604782c6a90d0656569541cd9540738825a7fcb255'
if((Get-FileHash -LiteralPath $source -Algorithm SHA256).Hash.ToLowerInvariant() -ne $hash){throw 'Wrong pinned wrapper.'}
if((Import-PowerShellDataFile -LiteralPath (Join-Path $ModulePath 'MicrosoftTeams.psd1')).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
Add-Type -TypeDefinition @'
using System.Collections;
namespace Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Runtime { public class SendAsyncStep {} }
namespace Microsoft.Teams.ConfigAPI.Cmdlets.Telemetry {
 public class CustomCmdletUtils {
  public CustomCmdletUtils(object invocation) {}
  public Hashtable ProcessArgs() { return new Hashtable(); }
  public void SendTelemetry() {}
 }
}
'@
$state=@{calls=[Collections.Generic.List[object]]::new()}
$stub=New-Module -Name 'Microsoft.Teams.ConfigAPI.Cmdlets.internal' -ArgumentList $state -ScriptBlock {
 param($state)
 $script:state=$state
 function Record($operation,$parameters){$script:state.calls.Add(@{operation=$operation;parameters=($parameters|ConvertTo-Json -Depth 15|ConvertFrom-Json -AsHashtable)})}
 function New-CsCustomPolicyPackage {[CmdletBinding(SupportsShouldProcess)]param($Identity,$PolicyList,$Description) Record 'create' $PSBoundParameters}
 function Update-CsCustomPolicyPackage {[CmdletBinding(SupportsShouldProcess)]param($Identity,$PolicyList,$Description) Record 'update' $PSBoundParameters}
 function Grant-CsGroupPolicyPackageAssignment {[CmdletBinding(SupportsShouldProcess)]param($GroupId,$PackageName,$PolicyRankings) Record 'group' $PSBoundParameters}
 Export-ModuleMember -Function *-Cs*
}
Import-Module $stub
$tokens=$null;$errors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile($source,[ref]$tokens,[ref]$errors)
if($errors.Count){throw 'Pinned source parse failed.'}
foreach($name in @('New-CsCustomPolicyPackage','Update-CsCustomPolicyPackage','Grant-CsGroupPolicyPackageAssignment')){
 $definition=$ast.Find({param($node)$node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq $name},$true)
 if($null -eq $definition){throw 'Missing pinned wrapper.'}
 . ([scriptblock]::Create($definition.Extent.Text))
}
$cases=[Collections.Generic.List[object]]::new()
function Observe([string]$name,[string]$command,[string[]]$values){
 $state.calls.Clear();$failure=$null
 try{
  switch($command){
   'create' {New-CsCustomPolicyPackage -Identity FixturePackage -PolicyList $values -Description ''}
   'update' {Update-CsCustomPolicyPackage -Identity FixturePackage -PolicyList $values}
   'group' {Grant-CsGroupPolicyPackageAssignment -GroupId 'fixture/group' -PackageName '' -PolicyRankings $values}
  }
 }catch{$failure=$_.Exception.GetBaseException().GetType().FullName}
 $cases.Add([ordered]@{name=$name;command=$command;input=$values;calls=@($state.calls.ToArray());failure_type=$failure})
}
Observe 'create-delimiters' 'create' @(' A , B ','A.B','A:B','A;B','A B',"A`tB")
Observe 'update-duplicates' 'update' @('Type,Name','Type,Name')
Observe 'create-first-split-only' 'create' @('Type,,Name','Type:Name:Child')
Observe 'create-empty-pieces' 'create' @(',Name','Type,',',')
Observe 'create-missing-delimiter' 'create' @('TypeName')
Observe 'group-integer-ranks' 'group' @('Type,0','Type;-1','Type:2147483647','Type -2147483648')
Observe 'group-rank-conversions' 'group' @('Type,garbage','Type,','Type,2147483648','Type,0x10','Type,1.5','Type,1e2','Type,08')
Observe 'group-missing-delimiter' 'group' @('TypeRank')
Observe 'group-empty-rankings' 'group' @()
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-policy-package-wrapper-binding'
 source=@{path='custom/Merged_custom_PsExt.ps1';sha256=$hash}
 stage='Unmodified AST-extracted custom functions under telemetry/internal-command stubs. Captures pair parsing and intermediate PowerShell rank values before generated DTO conversion. No real module import, credentials, sender, command lifecycle or network.'
 cases=$cases.ToArray()
}|ConvertTo-Json -Depth 25|Set-Content -LiteralPath $OutputPath -Encoding utf8
"policy package wrapper observations: $($cases.Count)"
