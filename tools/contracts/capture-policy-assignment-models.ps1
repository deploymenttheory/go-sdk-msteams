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
Add-Type -Path (Join-Path $PSScriptRoot 'oracle/PolicyAssignmentModelCapture.cs') -ReferencedAssemblies $references
$cases=[Collections.Generic.List[object]]::new()
foreach($v in @(
 @{name='group-empty';model='GroupAssignment';input='{}'},
 @{name='group-zero';model='GroupAssignment';input='{"groupId":"fixture-group","priority":0,"policyType":"Fixture","policyName":"","createdBy":null,"future":false}'},
 @{name='group-negative';model='GroupAssignment';input='{"priority":-1,"createdTime":"2026-11-21T10:00:00.1234567Z"}'},
 @{name='effective-empty';model='EffectivePolicy';input='{}'},
 @{name='effective-sources';model='EffectivePolicy';input='{"policyType":"Fixture","policyName":"","policySource":[{"assignmentType":"Direct","policyName":"","groupId":null},{"assignmentType":"Group","policyName":"FixtureGroup","groupId":"fixture-group"},{"assignmentType":"Future","future":false}],"future":0}'},
 @{name='source-empty';model='SourceEntry';input='{}'},
 @{name='source-values';model='SourceEntry';input='{"assignmentType":"Group","policyName":"Fixture","policyType":"FixtureType","reference":"fixture-group","groupId":"future-group-field","future":null}'},
 @{name='batch-empty';model='BatchJobStatus';input='{}'},
 @{name='batch-values';model='BatchJobStatus';input='{"operationId":"fixture-operation","operationName":"","overallStatus":"Future","notStartedCount":0,"inProgressCount":1,"completedCount":2,"errorCount":3,"createdBy":"fixture-user","createdTime":"2026-11-21T10:00:00.1234567Z","completedTime":null,"userState":[{"id":"fixture-user","state":"Future","result":""}],"future":false}'},
 @{name='batch-null-count';model='BatchJobStatus';input='{"completedCount":null,"userState":[]}'},
 @{name='simple-empty';model='SimpleBatchJobStatus';input='{}'},
 @{name='simple-zero';model='SimpleBatchJobStatus';input='{"operationId":"fixture-operation","overallStatus":"Future","completedCount":0}'},
 @{name='user-state-empty';model='GrantActionStatus';input='{}'},
 @{name='user-state-values';model='GrantActionStatus';input='{"id":"fixture-user","state":"Future","result":"","future":false}'}
)){
 $rejected=$false;$projection=$null
 try{$projection=[TeamsContractOracle.PolicyAssignmentModelOracle]::Project($v.model,$v.input)|ConvertFrom-Json -AsHashtable}catch{$rejected=$true}
 $cases.Add([ordered]@{name=$v.name;model=$v.model;input=($v.input|ConvertFrom-Json -AsHashtable);rejected=$rejected;projection=$projection})
}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-policy-assignment-generated-models'
 package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assembly=@{path=$relative;sha256=$expected}
 stage='Pinned generated FromJsonString/ToJsonString model projections. No command construction, credentials, request helper or network. Native snapshots additionally retain received timestamps, missing/null presence and future raw fields rather than fabricating CLR defaults.'
 cases=$cases.ToArray()
}|ConvertTo-Json -Depth 25|Set-Content -LiteralPath $OutputPath -Encoding utf8
"policy assignment generated model observations: $($cases.Count)"
