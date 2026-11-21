[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$relative='bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll'
$expected='8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'
$file=Join-Path $ModulePath $relative
if((Get-FileHash -LiteralPath $file -Algorithm SHA256).Hash.ToLowerInvariant() -ne $expected){throw 'Wrong pinned generated assembly.'}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
$env:MSTeamsContextInternal=$null
Import-Module $manifest -Force
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@($file)
Add-Type -Path (Join-Path $PSScriptRoot 'oracle/PolicyPackageModelCapture.cs') -ReferencedAssemblies $references
$cases=[Collections.Generic.List[object]]::new()
foreach($v in @(
 @{name='package-empty';model='Package';input='{}'},
 @{name='package-values';model='Package';input='{"Name":"Fixture","Description":"","PackageType":"Future","RecommendationType":null,"Policies":{"TeamsMeetingPolicy":"FixtureMeeting","Future":{"flag":false},"nullable":null},"future":0}'},
 @{name='summary-values';model='Summary';input='{"Name":"Fixture","Policies":{},"Description":null,"PackageType":"Future"}'},
 @{name='recommendation-values';model='Recommendation';input='{"Name":"Fixture","RecommendationType":"Future","Policies":{"TeamsMeetingPolicy":"FixtureMeeting"}}'},
 @{name='custom-empty';model='CustomResponse';input='{}'},
 @{name='custom-values';model='CustomResponse';input='{"ResponseMessage":"","future":false}'},
 @{name='group-empty';model='GroupResponse';input='{}'},
 @{name='group-partial';model='GroupResponse';input='{"ResponseMessage":"FixturePartial","FailedAssignPolicies":["FixturePolicy"],"FailedCreatePolicies":[],"future":false}'},
 @{name='user-empty';model='UserResponse';input='{}'},
 @{name='user-partial';model='UserResponse';input='{"ResponseMessage":"FixturePartial","FailedUserIds":["fixture/a"],"FailedUserErrorMessages":{"LegacyUser":["fixture-private"],"Future":["fixture-future"]},"FailedCreatePolicies":[],"FailedAssignPolicies":null,"future":false}'},
 @{name='batch-empty';model='BatchResponse';input='{}'},
 @{name='batch-values';model='BatchResponse';input='{"OperationId":"fixture-operation","ResponseMessage":"","FailedCreatePolicies":[],"future":false}'},
 @{name='pair-empty';model='PolicyPair';input='{}'},
 @{name='pair-values';model='PolicyPair';input='{"PolicyType":"FixtureType","PolicyName":"","future":false}'},
 @{name='rank-default';model='Ranking';input='{}'},
 @{name='rank-zero';model='Ranking';input='{"PolicyType":"FixtureType","Rank":0}'},
 @{name='rank-negative';model='Ranking';input='{"PolicyType":"FixtureType","Rank":-1}'},
 @{name='rank-null';model='Ranking';input='{"PolicyType":"FixtureType","Rank":null}'}
)){
 $rejected=$false;$projection=$null
 try{$projection=[TeamsContractOracle.PolicyPackageModelOracle]::Project($v.model,$v.input)|ConvertFrom-Json -AsHashtable}catch{$rejected=$true}
 $cases.Add([ordered]@{name=$v.name;model=$v.model;input=($v.input|ConvertFrom-Json -AsHashtable);rejected=$rejected;projection=$projection})
}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-policy-package-generated-models'
 package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assembly=@{path=$relative;sha256=$expected}
 stage='Pinned generated FromJsonString/ToJsonString projections only. Records serializer casing, policy dictionaries, partial-failure containers and Int32 rank defaults. No command construction, credentials, sender or network. Native snapshots must separately preserve received absence/null/future fields rather than fabricating model defaults.'
 cases=$cases.ToArray()
}|ConvertTo-Json -Depth 30|Set-Content -LiteralPath $OutputPath -Encoding utf8
"policy package generated model observations: $($cases.Count)"
