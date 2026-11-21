[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblies=@(
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'},
 @{path='netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll';sha256='477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af'}
)
foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned education configuration dependency.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
$env:MSTeamsContextInternal=$null
Import-Module $manifest -Force
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@((Join-Path $PSHOME 'System.Management.Automation.dll'))+@($assemblies|ForEach-Object{Join-Path $ModulePath $_.path})+@((Join-Path $PSHOME 'Newtonsoft.Json.dll'))
Add-Type -Path @((Join-Path $PSScriptRoot 'oracle/PolicyProviderCapture.cs'),(Join-Path $PSScriptRoot 'oracle/EducationConfigurationBindingCapture.cs')) -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$cases=[Collections.Generic.List[object]]::new()
$requests=@(
 @{verb='Get';name='list';arguments=@{}},
 @{verb='Get';name='identity';arguments=@{Identity='Global'}},
 @{verb='Get';name='empty-identity';arguments=@{Identity=''}},
 @{verb='Get';name='filter';arguments=@{Filter='Fixture*'}},
 @{verb='Get';name='empty-filter';arguments=@{Filter=''}},
 @{verb='Set';name='identity-only';arguments=@{Identity='Global'}},
 @{verb='Set';name='force-only';arguments=@{Identity='Global';Force=$false}},
 @{verb='Set';name='empty-identity';arguments=@{Identity='';UpdateParentInformation='Future'}},
 @{verb='Set';name='numeric-string';arguments=@{Identity='Global';UpdateParentInformation=0}},
 @{verb='Set';name='all-settings';arguments=@{Identity='Global';ParentGuardianPreferredContactMethod='FutureContact';EduGenerativeAIEnhancements='FutureEnhancement';UpdateParentInformation='FutureUpdate'}},
 @{verb='Set';name='force-true';arguments=@{Identity='Global';UpdateParentInformation='';Force=$true}}
)
foreach($name in @('ParentGuardianPreferredContactMethod','EduGenerativeAIEnhancements','UpdateParentInformation')){
 $nullArgs=@{Identity='Global'};$nullArgs[$name]=$null
 $emptyArgs=@{Identity='Global'};$emptyArgs[$name]=''
 $requests+=@{verb='Set';name=($name+'-null');arguments=$nullArgs}
 $requests+=@{verb='Set';name=($name+'-empty');arguments=$emptyArgs}
}
foreach($c in $requests){
 $argsJson=$c.arguments|ConvertTo-Json -Compress
 $observed=[TeamsContractOracle.EducationConfigurationBindingOracle]::Run($c.verb,$argsJson)|ConvertFrom-Json -AsHashtable
 if($null -ne $observed.output){$observed.output=$observed.output|ConvertFrom-Json -AsHashtable}
 if($observed.sends -ne 0 -or $observed.tokens -ne 0){throw 'Binding crossed the offline boundary.'}
 $cases.Add(@{command=($c.verb+'-CsTeamsEducationConfiguration');name=$c.name;arguments=$c.arguments;observed=$observed})
}
$filters=@('','Fixture*',' ','Global','Global/*')|ForEach-Object{@{pattern=$_;observed=([TeamsContractOracle.EducationConfigurationBindingOracle]::Filter($_)|ConvertFrom-Json -AsHashtable)}}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='inherited-compiled-education-configuration-binding-and-filter';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString()
 stage='Owned subclasses inherit original compiled properties/attributes/constructors. Synthetic Begin/Process/End replace source lifecycle solely to return compiled binder values and actual protected AreParametersSpecified. In-process runspace/fixture commands, memory client and credential-refusing host. Actual protected FilterPolicies invoked separately. No original lifecycle, scope authorization, token acquisition, discovery or network. Source model/provider are a separate fixture.'
 cases=$cases.ToArray();filters=$filters
}|ConvertTo-Json -Depth 30|Set-Content -LiteralPath $OutputPath -Encoding utf8
"education configuration binder cases: $($cases.Count); filters: $($filters.Count)"
