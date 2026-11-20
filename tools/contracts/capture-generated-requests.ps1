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
$policy = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.XdsConfiguration]::DeserializeFromDictionary(@{Identity='tag:Fixture / Policy';Description='';AppIds=@('fixture-app')})
Observe 'policy-list' 'ConfigurationGetConfigurations' @('ApplicationAccessPolicy',$null,$null) 200 '[]'
Observe 'policy-read-escaped' 'ConfigurationGetConfiguration' @('ApplicationAccessPolicy','tag:Fixture / Policy',$null,$null) 200 '{"Identity":"tag:Fixture / Policy","AppIds":[]}'
Observe 'policy-create' 'ConfigurationCreateConfiguration' @('ApplicationAccessPolicy',$null,$null,$policy) 201 ''
Observe 'policy-update-put' 'ConfigurationUpdateConfiguration' @('ApplicationAccessPolicy','tag:Fixture / Policy',$null,$null,$policy) 204 ''
Observe 'policy-delete' 'ConfigurationDeleteConfiguration' @('ApplicationAccessPolicy','tag:Fixture / Policy') 204 ''
Observe 'policy-forbidden' 'ConfigurationGetConfiguration' @('ApplicationAccessPolicy','Fixture',$null,$null) 403 '{"code":"FixtureDenied","message":"Fixture message","action":"Fixture action"}'
Observe 'policy-malformed-error' 'ConfigurationGetConfiguration' @('ApplicationAccessPolicy','Fixture',$null,$null) 500 'not-json'
Observe 'policy-pre-canceled' 'ConfigurationGetConfigurations' @('ApplicationAccessPolicy',$null,$null) 200 '[]' $true
$targeting = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.XdsConfiguration]::DeserializeFromDictionary(@{Identity='Global';CustomTagsMode='Enabled';Description=''})
Observe 'targeting-policy-list' 'ConfigurationGetConfigurations' @('TeamsTargetingPolicy',$null,$null) 200 '[]'
Observe 'targeting-policy-read' 'ConfigurationGetConfiguration' @('TeamsTargetingPolicy','Global',$null,$null) 200 '{"Identity":"Global","CustomTagsMode":"Enabled"}'
Observe 'targeting-policy-update' 'ConfigurationUpdateConfiguration' @('TeamsTargetingPolicy','Global',$null,$null,$targeting) 204 ''
Observe 'targeting-policy-delete' 'ConfigurationDeleteConfiguration' @('TeamsTargetingPolicy','Global') 204 ''
$userGrant = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.SingleGrantUserRequest]::new()
$userGrant.PolicyName = 'Fixture'
Observe 'user-assignment' 'UserGrantPolicy' @('fixture/user','ApplicationAccessPolicy',$userGrant) 200 ''
$userGrant.PolicyName = ''
Observe 'user-clear-empty-policy' 'UserGrantPolicy' @('fixture/user','ApplicationAccessPolicy',$userGrant) 200 ''
$tenantGrant = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.SingleGrantTenantRequest]::new()
$tenantGrant.PolicyName = 'Fixture'; $tenantGrant.ForceSwitchPresent = $false
Observe 'tenant-assignment' 'TenantGrantPolicy' @('ApplicationAccessPolicy',$tenantGrant) 200 ''
$groupGrant = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.GroupGrantPayload]::new()
$groupGrant.PolicyName = 'Fixture'; $groupGrant.Priority = 2
Observe 'group-assignment-accepted' 'GroupGrantPolicyAssignment' @('fixture/group','ApplicationAccessPolicy',$groupGrant) 202 '{"OperationId":"fixture-operation"}'
Observe 'group-assignment-no-content' 'GroupGrantPolicyAssignment' @('fixture/group','ApplicationAccessPolicy',$groupGrant) 204 ''
Observe 'template-read-opaque-id' 'TeamTemplateGet' @('/api/teamtemplates/v1.0/fixture/Tenant/en-US') 200 '{"value":{"displayName":"Fixture"}}'
Observe 'template-list' 'TeamTemplateGetList' @('en-US') 200 '{"value":[]}'
Observe 'template-delete-empty' 'TeamTemplateDelete' @('/api/teamtemplates/v1.0/fixture/Tenant/en-US') 204 ''
$template = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.TeamTemplate]::new()
$template.DisplayName='Fixture'; $template.ShortDescription='Summary'
Observe 'template-create' 'TeamTemplateNew' @('en-US',$template) 201 '{"id":"fixture-template"}'
Observe 'template-update' 'TeamTemplateUpdate' @('/api/teamtemplates/v1.0/fixture/Tenant/en-US',$template) 200 '{"id":"fixture-template"}'
$deployment = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.FlwoServiceModelsFlwosBatchRequest]::new()
$deployment.DeploymentCsv='Rml4dHVyZQ=='; $deployment.UsersToNotify='fixture@example.invalid'
Observe 'deployment-create' 'BatchTeamsDeploymentNew' @($deployment) 202 '{"instanceId":"fixture-instance","executionId":"fixture-execution"}'
Observe 'deployment-read' 'BatchTeamsDeploymentStatusGet' @('fixture/instance') 200 'Fixture raw status'
$artifact = [ordered]@{
    schema_version=1; baseline_version='8.1.0'; captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
    qualification='offline-generated-request-and-dispatch'; package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
    assembly=@{path=$assemblyPath;sha256=$expected}
    stage='Generated service helper before session host/header rewrite; the injected sender captures in memory and is not attached to an HTTP client or the production module pipeline.'
    cases=$cases.ToArray()
}
$artifact | ConvertTo-Json -Depth 30 | Set-Content -LiteralPath $OutputPath -Encoding utf8
"generated request observations: $($cases.Count)"
