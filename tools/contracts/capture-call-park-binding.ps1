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
foreach($verb in @('New','Set')){
 $command="$verb-CsTeamsCallParkPolicy"
 $definition=$ast.Find({param($node)$node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq $command},$true)
 if($null -eq $definition){throw 'Pinned export missing.'}
 $attributes=($definition.Body.ParamBlock.Attributes|ForEach-Object{$_.Extent.Text}) -join "`n"
 $parameters=$definition.Body.ParamBlock.Extent.Text
 # Only the original declaration and attributes are used. None of the export's
 # Begin/Process/End code or custom runtime-attribute processor is invoked.
 . ([scriptblock]::Create("function Invoke-CallParkBindingFixture {`n$attributes`n$parameters`n process { ConvertTo-Json -InputObject `$PSBoundParameters -Depth 10 -Compress }`n}"))
 $bindingCases=@(
  @{name='omitted-settings';arguments=@{Identity='Fixture'}},
  @{name='false-zero-empty';arguments=@{Identity='Fixture';Description='';AllowCallPark=$false;ParkTimeoutSeconds=[long]0;PickupRangeStart=[long]0;PickupRangeEnd=[long]0}},
  @{name='signed-and-large';arguments=@{Identity='Fixture';AllowCallPark=$true;ParkTimeoutSeconds=[long]::MinValue;PickupRangeStart=[long]9007199254740993;PickupRangeEnd=[long]::MaxValue}},
  @{name='bound-null-fields';arguments=@{Identity='Fixture';Description=$null;AllowCallPark=$null;ParkTimeoutSeconds=$null;PickupRangeStart=$null;PickupRangeEnd=$null}},
  @{name='numeric-string';arguments=@{Identity='Fixture';PickupRangeStart='9007199254740993'}},
  @{name='overflow';arguments=@{Identity='Fixture';PickupRangeEnd='9223372036854775808'}}
 )
 foreach($field in @('Description','AllowCallPark','ParkTimeoutSeconds','PickupRangeStart','PickupRangeEnd')){
  $arguments=@{Identity='Fixture'};$arguments[$field]=$null
  $bindingCases+=@{name=('null-'+$field);arguments=$arguments}
 }
 foreach($c in $bindingCases){
  $failure=$null;$bound=$null;$wire=$null
  try{
   $arguments=$c.arguments
   $bound=Invoke-CallParkBindingFixture @arguments | ConvertFrom-Json -AsHashtable
   $model=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.XdsConfiguration]::DeserializeFromDictionary($bound)
   $wire=$model.ToJsonString()|ConvertFrom-Json -AsHashtable
  }catch{$failure=$_.Exception.GetBaseException().GetType().FullName}
  $cases.Add(@{command=$command;name=$c.name;bound=$bound;wire=$wire;failure_type=$failure})
 }
 Remove-Item Function:Invoke-CallParkBindingFixture
}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='isolated-call-park-public-parameter-binding';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 sources=$sources;host_powershell_version=$PSVersionTable.PSVersion.ToString()
 stage='Unmodified exported New/Set parameter declarations and attributes in a renamed function with a synthetic body returning bound arguments; actual pure XdsConfiguration dictionary serialization. No export Begin/Process/End, runtime custom-attribute processing, TRPS/flight selection, command lifecycle, credentials or network.'
 cases=$cases.ToArray()
}|ConvertTo-Json -Depth 25|Set-Content -LiteralPath $OutputPath -Encoding utf8
"call park isolated binding observations: $($cases.Count)"
