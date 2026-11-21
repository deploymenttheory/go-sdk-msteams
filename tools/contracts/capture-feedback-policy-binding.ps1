[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblies=@(
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'},
 @{path='netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll';sha256='477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af'}
)
foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned feedback policy dependency.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
$env:MSTeamsContextInternal=$null
Import-Module $manifest -Force
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@((Join-Path $PSHOME 'System.Management.Automation.dll'))+@($assemblies|ForEach-Object{Join-Path $ModulePath $_.path})+@((Join-Path $PSHOME 'Newtonsoft.Json.dll'))
Add-Type -Path @((Join-Path $PSScriptRoot 'oracle/PolicyProviderCapture.cs'),(Join-Path $PSScriptRoot 'oracle/FeedbackPolicyBindingCapture.cs')) -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$requests=@(
 @{verb='Get';name='list';arguments=@{}},
 @{verb='Get';name='filter';arguments=@{Filter='Fixture*'}},
 @{verb='Get';name='empty-filter';arguments=@{Filter=''}},
 @{verb='Get';name='empty-identity';arguments=@{Identity=''}},
 @{verb='New';name='identity-only';arguments=@{Identity='Fixture'}},
 @{verb='New';name='empty-identity';arguments=@{Identity=''}},
 @{verb='Set';name='identity-only';arguments=@{Identity='Fixture'}},
 @{verb='Set';name='force-only';arguments=@{Identity='Fixture';Force=$false}},
 @{verb='Set';name='empty-identity';arguments=@{Identity='';UserInitiatedMode=''}},
 @{verb='Remove';name='force-false';arguments=@{Identity='Fixture';Force=$false}}
)
$properties=@([Microsoft.Teams.Policy.Administration.Cmdlets.Core.TeamsFeedbackPolicy].GetProperties()|Where-Object Name -ne 'Identity')
foreach($verb in @('New','Set')){
 $all=@{Identity='Fixture'}
 foreach($p in $properties){
  $nullArgs=@{Identity='Fixture'};$nullArgs[$p.Name]=$null
  $requests+=@{verb=$verb;name=($p.Name+'-null');arguments=$nullArgs}
  if($p.PropertyType -eq [string]){
   $value=@{Identity='Fixture'};$value[$p.Name]=''
   $requests+=@{verb=$verb;name=($p.Name+'-empty');arguments=$value}
   $all[$p.Name]='Future-'+$p.Name
  }else{
   $value=@{Identity='Fixture'};$value[$p.Name]=$false
   $requests+=@{verb=$verb;name=($p.Name+'-false');arguments=$value}
   $all[$p.Name]=$true
  }
 }
 $requests+=@{verb=$verb;name='all-settings';arguments=$all}
}
$cases=[Collections.Generic.List[object]]::new()
foreach($c in $requests){
 $arguments=$c.arguments|ConvertTo-Json -Compress
 $observed=[TeamsContractOracle.FeedbackBindingOracle]::Run($c.verb,$arguments)|ConvertFrom-Json -AsHashtable
 if($null -ne $observed.output){$observed.output=$observed.output|ConvertFrom-Json -AsHashtable}
 if($observed.sends -ne 0 -or $observed.tokens -ne 0){throw 'Feedback binder crossed offline boundary.'}
 $cases.Add(@{command=($c.verb+'-CsTeamsFeedbackPolicy');name=$c.name;arguments=$c.arguments;observed=$observed})
}
$filters=@('','Fixture*',' ','Global','Global/*')|ForEach-Object{@{pattern=$_;observed=([TeamsContractOracle.FeedbackBindingOracle]::Filter($_)|ConvertFrom-Json -AsHashtable)}}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='inherited-compiled-feedback-binding-and-filter';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString()
 stage='Owned subclasses inherit actual Core properties/attributes/constructors. Synthetic Begin/Process/End replace original lifecycle solely to capture binder values and protected New/Set predicates. Compiled CRUD only; modern public Grant is captured separately. Actual protected Get filter invoked separately. Memory client/credential-refusing host restored; no original auth/scope/processing, credentials, discovery or network.'
 cases=$cases.ToArray();filters=$filters
}|ConvertTo-Json -Depth 30|Set-Content -LiteralPath $OutputPath -Encoding utf8
"feedback binder cases: $($cases.Count); filters: $($filters.Count)"
