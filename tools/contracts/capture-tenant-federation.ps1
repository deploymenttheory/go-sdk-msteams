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
$references = @((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName) + @($assemblyFile,(Join-Path $PSHOME 'System.Management.Automation.dll'))
Add-Type -Path (Join-Path $PSScriptRoot 'oracle/GeneratedRequestCapture.cs') -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
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

Add-Type -Path (Join-Path $PSScriptRoot 'oracle/FederationDomainCapture.cs') -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$model=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.XdsConfiguration]::DeserializeFromDictionary(@{Identity='Global';AllowFederatedUsers=$false;AllowedDomains=@('fixture.example.invalid')})
Observe 'tenant-federation-list' 'ConfigurationGetConfigurations' @('TenantFederationSettings',$null,$null) 200 '[]'
Observe 'tenant-federation-read' 'ConfigurationGetConfiguration' @('TenantFederationSettings','Global',$null,$null) 200 '{"Identity":"Global","AllowedDomains":null,"AllowFederatedUsers":false}'
Observe 'tenant-federation-update' 'ConfigurationUpdateConfiguration' @('TenantFederationSettings','Global',$null,$null,$model) 204 ''
$models=[Collections.Generic.List[object]]::new()
foreach($name in @('omitted','bound-scalars','allow-all','allow-list','empty-allow-list','allowed-replacement','allowed-operators','allowed-empty-operators','blocked-mixed','blocked-operators','trial-mixed','trial-null','bound-null-scalar','conflicting-choices','null-choice','replace-operator')) {
    $rejected=$false;$wire=$null
    try{$wire=[TeamsContractOracle.FederationDomains]::NormalizeCase($name) | ConvertFrom-Json -AsHashtable}catch{$rejected=$true}
    $models.Add(@{name=$name;rejected=$rejected;wire=$wire})
}
$artifact=[ordered]@{
    schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
    qualification='offline-tenant-federation-helper-and-dispatch';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
    assembly=@{path=$assemblyPath;sha256=$expected}
    stage='Generated memory-only request capture before host/header rewrite; pure custom property-bag normalization/model serialization. No command processing, authentication, discovery or HTTP client.'
    cases=$cases.ToArray();models=$models.ToArray()
}
$artifact | ConvertTo-Json -Depth 30 | Set-Content -LiteralPath $OutputPath -Encoding utf8
"tenant federation observations: $($cases.Count) requests; $($models.Count) models"
