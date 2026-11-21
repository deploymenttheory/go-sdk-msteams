[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblies=@(
 @{path='net472/Microsoft.Teams.Policy.Administration.Cmdlets.OCE.dll';sha256='ea52b14ae69bfeb7ce63497176a81252dd7c0fa17ad93d6b8ae49ef358091f03'},
 @{path='net472/CmdletSettings.json';sha256='c0549eba3a249ef431b0d0c61ee232c815b74fe0cfa1b41b0860a8531f33e6dd'}
)
foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned OCE input.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
$env:MSTeamsContextInternal=$null
Import-Module $manifest -Force
$file=Join-Path $ModulePath $assemblies[0].path
$null=[Reflection.Assembly]::LoadFrom($file)
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@($file,[Management.Automation.PSObject].Assembly.Location,[Newtonsoft.Json.JsonConvert].Assembly.Location)
Add-Type -Path (Join-Path $PSScriptRoot 'oracle/PolicyAssignmentOceCapture.cs') -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$cases=[Collections.Generic.List[object]]::new()
function Observe([string]$name,[string]$service,[string]$operation,[string]$method,[int]$status,[string]$reply,[bool]$all=$false){
 $replacements=[Collections.Generic.Dictionary[string,string]]::new()
 $replacements['operationId']='00000000-0000-4000-8000-000000000004'
 $replacements['groupId']='00000000-0000-4000-8000-000000000005'
 $replacements['userId']='00000000-0000-0000-0000-000000000000'
 $replacements['region']='NOAM';$replacements['policyType']='Fixture';$replacements['all']=$all.ToString()
 $r=[TeamsContractOracle.PolicyAssignmentOceOracle]::Run($service,$operation,$method,$replacements,$status,$reply)
 $cases.Add([ordered]@{name=$name;service=$service;operation=$operation;status=$status;response=$reply;all=$all;capture=$r})
}
Observe 'batch-status' 'LROS' 'GetBatchOperationStatusForRegion' 'GET' 200 '{}'
Observe 'batch-list' 'LROS' 'GetAllBatchOperationStatusForRegion' 'GET' 200 '[]'
Observe 'batch-definition' 'LROS' 'GetBatchOperationDefinitionForRegion' 'GET' 200 '{}'
Observe 'batch-reprocess' 'LROS' 'ReprocessBatchOperationForRegion' 'POST' 202 '"00000000-0000-4000-8000-000000000006"'
Observe 'group-get' 'GPA' 'GetGroupPolicyAssignment' 'GET' 200 '{}'
Observe 'group-list' 'GPA' 'GetGroupPolicyAssignments' 'GET' 200 '[]'
Observe 'group-list-type' 'GPA' 'GetPolicyAssignmentForGroups' 'GET' 200 '[]'
Observe 'group-list-all' 'GPA' 'GetAllGroupPolicyAssignments' 'GET' 200 '[]'
Observe 'group-members' 'GPA' 'GetGPAGroupMembers' 'GET' 200 '[]'
Observe 'group-members-all' 'GPA' 'GetGPAGroupMembers' 'GET' 200 '[]' $true
Observe 'user-membership' 'GPA' 'GetGPAUserMembership' 'GET' 200 '[]'
Observe 'group-refresh' 'GPA' 'RefreshGroupUsers' 'POST' 202 '"fixture-operation"'
Observe 'group-reprocess' 'GPA' 'ReprocessGroupPolicyAssignment' 'POST' 200 '"fixture-operation"'
Observe 'group-null' 'GPA' 'GetGroupPolicyAssignment' 'GET' 200 'null'
Observe 'group-partial-status' 'GPA' 'GetGroupPolicyAssignment' 'GET' 206 '{}'
Observe 'group-forbidden' 'GPA' 'GetGroupPolicyAssignment' 'GET' 403 '{"code":"FixtureDenied","message":"fixture-private"}'
Observe 'group-not-found' 'GPA' 'GetGroupPolicyAssignment' 'GET' 404 '{}'
Observe 'group-unavailable' 'GPA' 'GetGroupPolicyAssignment' 'GET' 503 '{"code":"FixtureUnavailable","message":"fixture-private"}'
$models=[Collections.Generic.List[object]]::new()
foreach($v in @(
 @{name='group-empty';model='GroupAssignment';input='{}'},
 @{name='group-values';model='GroupAssignment';input='{"groupId":"00000000-0000-4000-8000-000000000005","priority":0,"policyName":"","createdBy":null,"createdTime":null,"future":false}'},
 @{name='member-empty';model='GroupMembers';input='{}'},
 @{name='member-values';model='GroupMembers';input='{"id":"00000000-0000-4000-8000-000000000007","@odata.type":"#fixture","future":false}'},
 @{name='membership-empty';model='UserMembership';input='{}'},
 @{name='membership-values';model='UserMembership';input='{"id":"00000000-0000-4000-8000-000000000005","groupTypes":[],"mailEnabled":false,"securityEnabled":true,"visibility":"Future","createdDateTime":null}'},
 @{name='definition-empty';model='BatchDefinition';input='{}'},
 @{name='definition-values';model='BatchDefinition';input='{"OperationId":"00000000-0000-4000-8000-000000000004","OperationName":"","OperationType":"Future","AcceptTime":0,"Ids":[],"Commands":[{"Name":"Fixture","RetryCount":0,"Properties":{"flag":false},"Description":""}]}'},
 @{name='status-empty';model='BatchStatus';input='{}'},
 @{name='status-values';model='BatchStatus';input='{"OperationId":"00000000-0000-4000-8000-000000000004","OverallStatus":"Future","CompletedCount":0,"CreatedTime":"fixture-time","CompletedTime":null}'},
 @{name='extended-values';model='ExtendedBatchStatus';input='{"OperationId":"00000000-0000-4000-8000-000000000004","UserState":[{"Id":"fixture-user","AlternateId":null,"State":"Future","Result":""}]}'},
 @{name='null';model='GroupAssignment';input='null'}
)){
 $rejected=$false;$projection=$null
 try{$projection=[TeamsContractOracle.PolicyAssignmentOceOracle]::Project($v.model,$v.input)|ConvertFrom-Json -AsHashtable}catch{$rejected=$true}
 $models.Add([ordered]@{name=$v.name;model=$v.model;input=($v.input|ConvertFrom-Json -AsHashtable);projection=$projection;rejected=$rejected})
}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-policy-assignment-oce-direct';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;profiles=([TeamsContractOracle.PolicyAssignmentOceOracle]::Profiles()|ConvertFrom-Json -AsHashtable)
 stage='Pinned OCE base constructor reads only its hash-checked packaged settings. Synthetic runtime consumes verbose/warnings. Actual suffix lookup, InitializeHeaders and RestClient sender run with synthetic token and memory HttpClient; private client and global JSON settings are restored. No InitializeServiceParameters, cmdlet lifecycle, host credential acquisition, OCE profile activation or network. Model projections call only Newtonsoft on DTOs.'
 cases=$cases.ToArray();models=$models.ToArray()
}|ConvertTo-Json -Depth 35|Set-Content -LiteralPath $OutputPath -Encoding utf8
"policy assignment OCE observations: $($cases.Count) requests / $($models.Count) models"
