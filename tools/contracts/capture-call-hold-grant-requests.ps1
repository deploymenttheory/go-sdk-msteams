# Credential-free request construction/response-dispatch oracle. Sender is memory only.
[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$ModulePath,
    [Parameter(Mandatory)][string]$OutputPath
)
$ErrorActionPreference = 'Stop'
if (Test-Path -LiteralPath $OutputPath) { throw 'Output already exists; choose a new candidate path.' }
$assemblyPath = 'bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll'
$expected = '8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'
$assemblyFile = Join-Path $ModulePath $assemblyPath
if ((Get-FileHash -LiteralPath $assemblyFile -Algorithm SHA256).Hash.ToLowerInvariant() -ne $expected) { throw 'Generated assembly differs from the pinned baseline.' }
$moduleFile = Join-Path $ModulePath 'MicrosoftTeams.psd1'
if ((Import-PowerShellDataFile -LiteralPath $moduleFile).ModuleVersion -ne '8.1.0') { throw 'Module version is not 8.1.0.' }
Import-Module $moduleFile -Force
$references = @((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName) + @($assemblyFile)
Add-Type -Path (Join-Path $PSScriptRoot 'oracle/GeneratedRequestCapture.cs') -ReferencedAssemblies $references
$cases = [Collections.Generic.List[object]]::new()
function Observe([string]$name, [string]$operation, [object[]]$inputs, [int]$status, [string]$response, [bool]$canceled=$false) {
    $result = [TeamsContractOracle.Oracle]::Run($operation, $inputs, $status, $response, $canceled)
    $record = @{
        name=$name; operation=$operation; response_status=$status; response_body=$response
        method=$result.Method; generated_url=$result.GeneratedUrl; content_type=$result.ContentType
        sends=$result.Sends; selected_handler=$result.Handler; decoded_type=$result.DecodedType
        decoder_failure=$result.DecoderFailure; decoded_null=$result.DecodedNull; canceled=$canceled; body=$null
    }
    if ($null -ne $result.RequestBody) { $record.body = ConvertFrom-Json -InputObject $result.RequestBody -AsHashtable }
    $cases.Add($record)
}
$userGrant = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.SingleGrantUserRequest]::new()
$userGrant.PolicyName = 'Fixture'
Observe 'user-assignment' 'UserGrantPolicy' @('fixture/user','TeamsCallHoldPolicy',$userGrant) 200 ''
$userGrant.PolicyName = ''
Observe 'user-clear-empty-policy' 'UserGrantPolicy' @('fixture/user','TeamsCallHoldPolicy',$userGrant) 200 ''
$tenantGrant = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.SingleGrantTenantRequest]::new()
$tenantGrant.PolicyName = 'Fixture'; $tenantGrant.ForceSwitchPresent = $false
Observe 'tenant-assignment' 'TenantGrantPolicy' @('TeamsCallHoldPolicy',$tenantGrant) 200 ''
$groupGrant = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.GroupGrantPayload]::new()
$groupGrant.PolicyName = 'Fixture'; $groupGrant.Priority = 2
Observe 'group-assignment-accepted' 'GroupGrantPolicyAssignment' @('fixture/group','TeamsCallHoldPolicy',$groupGrant) 202 '{"OperationId":"fixture-operation"}'
Observe 'group-assignment-no-content' 'GroupGrantPolicyAssignment' @('fixture/group','TeamsCallHoldPolicy',$groupGrant) 204 ''
$groupGrant.Priority = 0
Observe 'group-assignment-zero-priority' 'GroupGrantPolicyAssignment' @('fixture/group','TeamsCallHoldPolicy',$groupGrant) 202 '{"OperationId":"fixture-operation"}'
$groupGrant = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.GroupGrantPayload]::new()
$groupGrant.PolicyName = ''
Observe 'group-clear-without-priority' 'GroupGrantPolicyAssignment' @('fixture/group','TeamsCallHoldPolicy',$groupGrant) 204 ''
$artifact = [ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-call-hold-grant-generated-requests';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assembly=@{path=$assemblyPath;sha256=$expected}
 stage='Actual pinned generated service helper before session host/header rewrite; injected memory sender, no HttpClient/credentials, flighting, discovery or production pipeline.'
 cases=$cases.ToArray()
}
$artifact|ConvertTo-Json -Depth 30|Set-Content -LiteralPath $OutputPath -Encoding utf8
"call hold grant generated observations: $($cases.Count)"
