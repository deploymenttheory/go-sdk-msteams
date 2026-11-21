# Memory-only generated service calls, before the module host/header pipeline.
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
Add-Type -Path (Join-Path $PSScriptRoot 'oracle/GeneratedRequestCapture.cs') -ReferencedAssemblies $references
$cases=[Collections.Generic.List[object]]::new()
function Observe([string]$name,[string]$operation,[object[]]$inputs,[int]$status,[string]$reply){
 $r=[TeamsContractOracle.Oracle]::Run($operation,$inputs,$status,$reply,$false)
 $body=$null;if($null -ne $r.RequestBody){$body=$r.RequestBody|ConvertFrom-Json -AsHashtable}
 $cases.Add([ordered]@{name=$name;operation=$operation;response_status=$status;response_body=$reply;method=$r.Method;generated_url=$r.GeneratedUrl;content_type=$r.ContentType;sends=$r.Sends;selected_handler=$r.Handler;decoded_type=$r.DecodedType;decoder_failure=$r.DecoderFailure;decoded_null=$r.DecodedNull;body=$body})
}
Observe 'batch-list' 'BatchGetPolicyAssignmentOperation' @('') 200 '[]'
Observe 'batch-list-filter' 'BatchGetPolicyAssignmentOperation' @('Future /?&:@') 200 '[{"operationId":"fixture-op","overallStatus":"Future","completedCount":0}]'
Observe 'batch-get' 'BatchGetPolicyAssignmentOperation1' @('fixture/operation:@') 200 '{"operationId":"fixture/operation:@","overallStatus":"Future","userState":[{"id":"fixture-user","state":"Future","result":""}]}'
Observe 'group-get' 'GroupGetPolicyAssignment' @('fixture/group:@','Fixture/Policy:@') 200 '{"groupId":"fixture/group:@","priority":0,"policyType":"Fixture/Policy:@","policyName":""}'
Observe 'group-list' 'GroupGetPolicyAssignment1' @('fixture/group:@') 200 '[{"groupId":"fixture/group:@","priority":-1}]'
Observe 'group-list-type' 'GroupGetPolicyAssignment2' @('Fixture /?&:@') 200 '[]'
Observe 'group-list-all' 'GroupGetPolicyAssignment2' @('') 200 '[]'
Observe 'user-effective' 'UserGetPolicyAssignment' @('fixture/user:@','Fixture /?&:@') 200 '[{"policyType":"Fixture","policyName":"","policySource":[{"assignmentType":"Direct","policyName":""}]}]'
Observe 'user-effective-all' 'UserGetPolicyAssignment' @('fixture/user:@','') 200 '[]'
$group=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.GroupAssignPayload]::FromJsonString('{"policyName":"Fixture","priority":0}')
Observe 'group-create' 'GroupNewPolicyAssignment' @('fixture/group:@','Fixture/Policy:@',$group) 201 '{"OperationId":"fixture-operation"}'
$group=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.GroupAssignPayload]::new()
Observe 'group-create-empty' 'GroupNewPolicyAssignment' @('fixture/group:@','Fixture/Policy:@',$group) 204 ''
$group=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.GroupAssignPayload]::FromJsonString('{"policyName":"","priority":-1}')
Observe 'group-create-empty-name-negative-rank' 'GroupNewPolicyAssignment' @('fixture/group:@','Fixture/Policy:@',$group) 201 '{"OperationId":"fixture-operation","future":false}'
Observe 'group-delete' 'GroupRemovePolicyAssignment' @('fixture/group:@','Fixture/Policy:@') 202 '{"OperationId":"fixture-operation"}'
Observe 'group-delete-no-content' 'GroupRemovePolicyAssignment' @('fixture/group:@','Fixture/Policy:@') 204 ''
$update=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.GroupUpdatePayload]::FromJsonString('{"policyName":"UpdatedFixture","priority":0}')
Observe 'group-update' 'GroupSetPolicyAssignment' @('fixture/group:@','Fixture/Policy:@',$update) 202 '{"OperationId":"fixture-operation"}'
$update=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.GroupUpdatePayload]::new()
Observe 'group-update-empty' 'GroupSetPolicyAssignment' @('fixture/group:@','Fixture/Policy:@',$update) 204 ''
$batch=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.BatchAssignBody]::FromJsonString('{"policy":{"policyType":"Fixture","policyName":""},"identities":["fixture/a","fixture/a","fixture/b"],"additionalParameters":{"flag":false,"n":0,"empty":"","null":null}}')
Observe 'batch-create' 'BatchNewPolicyAssignmentOperation' @('Fixture /?&:@',$batch) 202 '{"OperationId":"fixture-operation"}'
$batch=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.BatchAssignBody]::new()
$batch.Identity=[string[]]@()
Observe 'batch-create-empty' 'BatchNewPolicyAssignmentOperation' @('',$batch) 202 '{}'
$nested=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.BatchAssignBody]::FromJsonString('{"identities":[],"additionalParameters":{"nested":{"flag":false,"null":null},"array":[0,false,"",null,{"null":null}]}}')
Observe 'batch-create-nested-additional' 'BatchNewPolicyAssignmentOperation' @('',$nested) 202 '{"OperationId":"fixture-operation"}'
Observe 'batch-list-null' 'BatchGetPolicyAssignmentOperation' @('') 200 'null'
Observe 'group-get-null' 'GroupGetPolicyAssignment' @('fixture/group','Fixture') 200 'null'
Observe 'group-get-malformed' 'GroupGetPolicyAssignment' @('fixture/group','Fixture') 200 'not-json'
Observe 'group-create-malformed' 'GroupNewPolicyAssignment' @('fixture/group','Fixture',$group) 201 'not-json'
Observe 'batch-create-malformed' 'BatchNewPolicyAssignmentOperation' @('',$batch) 202 'not-json'
Observe 'group-forbidden' 'GroupGetPolicyAssignment' @('fixture/group','Fixture') 403 '{"code":"FixtureDenied","message":"fixture-private"}'
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-policy-assignment-generated-requests'
 package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assembly=@{path=$relative;sha256=$expected}
 stage='Actual generated helper and DTO serializers with injected ISendAsync memory sender and synthetic callbacks, before host/header rewrite. No command construction, credential pipeline, discovery, HttpClient or network. Native DTO presence is separate from PowerShell parameter binding.'
 cases=$cases.ToArray()
}|ConvertTo-Json -Depth 30|Set-Content -LiteralPath $OutputPath -Encoding utf8
"policy assignment generated request observations: $($cases.Count)"
