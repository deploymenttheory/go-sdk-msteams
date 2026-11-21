# Actual generated helpers with a memory sender; no delegated login or tenant I/O.
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
 $cases.Add([ordered]@{name=$name;operation=$operation;response_status=$status;response_body=$reply;method=$r.Method;generated_url=$r.GeneratedUrl;headers=$r.RequestHeaders;content_type=$r.ContentType;sends=$r.Sends;selected_handler=$r.Handler;decoded_type=$r.DecodedType;decoder_failure=$r.DecoderFailure;decoded_null=$r.DecodedNull;body=$body})
}
Observe 'list' 'PolicyPackageGet' @() 200 '[]'
Observe 'get' 'PolicyPackageGet1' @('Fixture/Package:@') 200 '{"Name":"Fixture/Package:@","Policies":{}}'
Observe 'user-get' 'UserPolicyPackageGet' @('fixture-user@example.invalid') 200 '{"Name":"FixturePackage","Policies":{}}'
Observe 'user-recommendations' 'UserPolicyPackageGetRecommendation' @('fixture-user@example.invalid') 200 '[{"Name":"FixturePackage","RecommendationType":"Future","Policies":{}}]'
$create=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.PackageServiceModelsCreateCustomPackageRequest]::FromJsonString('{"PackageName":"Fixture/Package:@","Description":"","PolicyList":[{"PolicyType":"TeamsMeetingPolicy","PolicyName":"FixtureMeeting"}]}')
Observe 'create' 'CustomPolicyPackageNew' @($create) 201 '{"ResponseMessage":"Fixture"}'
$emptyCreate=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.PackageServiceModelsCreateCustomPackageRequest]::new()
Observe 'create-empty-dto' 'CustomPolicyPackageNew' @($emptyCreate) 201 '{}'
$update=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.PackageServiceModelsUpdateCustomPackageRequest]::FromJsonString('{"PackageName":"Fixture/Package:@","Description":"","PolicyList":[]}')
Observe 'update' 'CustomPolicyPackageUpdate' @($update) 200 '{"ResponseMessage":"Fixture"}'
Observe 'delete' 'CustomPolicyPackageRemove' @('Fixture/Package:@') 204 ''
$group=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.PackageServiceModelsApplyPackageGroupRequest]::FromJsonString('{"GroupId":"fixture/group","PackageName":"FixturePackage","PolicyRankings":[{"PolicyType":"TeamsMeetingPolicy","Rank":0}]}')
Observe 'group-assign' 'GroupPolicyPackageGrantAssignment' @($group) 200 '{"ResponseMessage":"Fixture","FailedAssignPolicies":[],"FailedCreatePolicies":[]}'
$groupEmpty=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.PackageServiceModelsApplyPackageGroupRequest]::FromJsonString('{"GroupId":"fixture/group","PackageName":"","PolicyRankings":[]}')
Observe 'group-empty-package' 'GroupPolicyPackageGrantAssignment' @($groupEmpty) 200 '{"ResponseMessage":"Fixture"}'
$groupNull=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.PackageServiceModelsApplyPackageGroupRequest]::FromJsonString('{"GroupId":"fixture/group","PackageName":null,"PolicyRankings":[{"PolicyType":"Fixture"}]}')
Observe 'group-null-package-default-rank' 'GroupPolicyPackageGrantAssignment' @($groupNull) 200 '{}'
$batch=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.PackageServiceModelsPostPackageBatchRequest]::FromJsonString('{"PackageType":"FixturePackage","UserIds":["fixture/a","fixture/a","fixture/b"]}')
Observe 'user-grant' 'UserPolicyPackageGrant' @($batch) 201 '{"ResponseMessage":"Fixture","FailedUserIds":[],"FailedUserErrorMessages":{"LegacyUser":[]}}'
Observe 'batch-grant' 'BatchPolicyPackageNewAssignmentOperation' @($batch) 202 '{"OperationId":"fixture-operation","ResponseMessage":"Fixture","FailedCreatePolicies":[]}'
Observe 'user-partial' 'UserPolicyPackageGrant' @($batch) 201 '{"ResponseMessage":"FixturePartial","FailedUserIds":["fixture/a"],"FailedAssignPolicies":["FixturePolicy"],"FailedUserErrorMessages":{"LegacyUser":["fixture-private"],"Future":["fixture-future"]}}'
Observe 'group-domain-error' 'GroupPolicyPackageGrantAssignment' @($group) 520 '{"ResponseMessage":"FixtureFailure","FailedAssignPolicies":["FixturePolicy"]}'
Observe 'user-domain-error' 'UserPolicyPackageGrant' @($batch) 521 '{"ResponseMessage":"FixtureFailure","FailedUserIds":["fixture/a"]}'
Observe 'create-bad-request' 'CustomPolicyPackageNew' @($create) 400 '{"ResponseMessage":"FixtureFailure"}'
Observe 'get-not-found' 'PolicyPackageGet1' @('Fixture/Package:@') 404 '{"ResponseMessage":"FixtureMissing"}'
Observe 'batch-error' 'BatchPolicyPackageNewAssignmentOperation' @($batch) 500 '{"ResponseMessage":"FixtureFailure","FailedCreatePolicies":["FixturePolicy"]}'
Observe 'create-malformed' 'CustomPolicyPackageNew' @($create) 201 'not-json'
Observe 'get-null' 'PolicyPackageGet1' @('Fixture/Package:@') 200 'null'
Observe 'list-null' 'PolicyPackageGet' @() 200 'null'
Observe 'batch-missing-id' 'BatchPolicyPackageNewAssignmentOperation' @($batch) 202 '{}'
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-policy-package-generated-requests'
 package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assembly=@{path=$relative;sha256=$expected}
 stage='Actual generated helper, DTO serializer and injected ISendAsync memory sender with synthetic callbacks, before host/header rewriting. Retains user header, distinct success/error callback selections and decoder failures. No command construction, credentials, discovery, HttpClient or network. These observations do not qualify documented delegated-mode eligibility or shell binding.'
 cases=$cases.ToArray()
}|ConvertTo-Json -Depth 30|Set-Content -LiteralPath $OutputPath -Encoding utf8
"policy package generated request observations: $($cases.Count)"
