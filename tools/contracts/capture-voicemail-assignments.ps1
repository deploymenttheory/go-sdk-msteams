# Memory-only generated requests before host/header rewrite; no credential pipeline.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblyPath='bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll'
$expected='8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'
$assemblyFile=Join-Path $ModulePath $assemblyPath
if((Get-FileHash -LiteralPath $assemblyFile -Algorithm SHA256).Hash.ToLowerInvariant() -ne $expected){throw 'Wrong pinned assembly.'}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
Import-Module $manifest -Force
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@($assemblyFile)
Add-Type -Path (Join-Path $PSScriptRoot 'oracle/GeneratedRequestCapture.cs') -ReferencedAssemblies $references
$cases=[Collections.Generic.List[object]]::new()
function Observe([string]$name,[string]$operation,[object[]]$inputs,[int]$status,[string]$reply) {
 $r=[TeamsContractOracle.Oracle]::Run($operation,$inputs,$status,$reply,$false)
 $body=$null;if($null -ne $r.RequestBody){$body=$r.RequestBody | ConvertFrom-Json -AsHashtable}
 $cases.Add(@{name=$name;operation=$operation;response_status=$status;response_body=$reply;method=$r.Method;generated_url=$r.GeneratedUrl;content_type=$r.ContentType;sends=$r.Sends;selected_handler=$r.Handler;decoded_type=$r.DecodedType;decoder_failure=$r.DecoderFailure;decoded_null=$r.DecodedNull;body=$body})
}
$user=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.SingleGrantUserRequest]::new()
$user.PolicyName='Fixture'
Observe 'user' 'UserGrantPolicy' @('fixture/user','OnlineVoicemailPolicy',$user) 200 ''
$user.PolicyName=''
Observe 'user-empty' 'UserGrantPolicy' @('fixture/user','OnlineVoicemailPolicy',$user) 200 ''
$user=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.SingleGrantUserRequest]::new()
Observe 'user-null' 'UserGrantPolicy' @('fixture/user','OnlineVoicemailPolicy',$user) 200 ''
$tenant=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.SingleGrantTenantRequest]::new()
$tenant.PolicyName='Fixture';$tenant.ForceSwitchPresent=$false
Observe 'tenant' 'TenantGrantPolicy' @('OnlineVoicemailPolicy',$tenant) 200 ''
$tenant=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.SingleGrantTenantRequest]::new();$tenant.ForceSwitchPresent=$true
Observe 'tenant-null-force' 'TenantGrantPolicy' @('OnlineVoicemailPolicy',$tenant) 200 ''
$group=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.GroupGrantPayload]::new()
$group.PolicyName='Fixture'
Observe 'group-accepted' 'GroupGrantPolicyAssignment' @('fixture/group','OnlineVoicemailPolicy',$group) 202 '{"OperationId":"fixture-operation"}'
$group.Priority=0
Observe 'group-zero-priority' 'GroupGrantPolicyAssignment' @('fixture/group','OnlineVoicemailPolicy',$group) 204 ''
$group=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.GroupGrantPayload]::new();$group.Priority=-1
Observe 'group-null-negative' 'GroupGrantPolicyAssignment' @('fixture/group','OnlineVoicemailPolicy',$group) 202 '{"OperationId":"fixture-operation","Future":false}'
Observe 'group-malformed-accepted' 'GroupGrantPolicyAssignment' @('fixture/group','OnlineVoicemailPolicy',$group) 202 'not-json'
Observe 'group-empty-id' 'GroupGrantPolicyAssignment' @('fixture/group','OnlineVoicemailPolicy',$group) 202 '{}'
Observe 'group-forbidden' 'GroupGrantPolicyAssignment' @('fixture/group','OnlineVoicemailPolicy',$group) 403 '{"code":"FixtureDenied","message":"fixture-private"}'
Observe 'operation-read' 'BatchGetPolicyAssignmentOperation1' @('fixture/operation') 200 '{"operationId":"fixture/operation","overallStatus":"Future","completedCount":0,"errorCount":null,"Future":false}'
Observe 'operation-malformed' 'BatchGetPolicyAssignmentOperation1' @('fixture/operation') 200 'not-json'
Observe 'operation-unknown-fields' 'BatchGetPolicyAssignmentOperation1' @('fixture/operation') 200 '{"operationId":"fixture/operation","overallStatus":"Future","userState":[]}'
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-voicemail-generated-assignment';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assembly=@{path=$assemblyPath;sha256=$expected}
 stage='Actual generated service helper with injected ISendAsync memory sender and synthetic callbacks, before host/header rewrite. No HttpClient, credentials, command processing, flight/discovery or network. DTO-null and native zero-priority behavior are distinct from wrapper string/rank binding.'
 cases=$cases.ToArray()
}|ConvertTo-Json -Depth 30|Set-Content -LiteralPath $OutputPath -Encoding utf8
"voicemail generated assignment observations: $($cases.Count)"
