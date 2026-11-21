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
$model=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.XdsConfiguration]::DeserializeFromDictionary(@{Identity='Global';AllowRemoveParticipantAppIds=@('fixture-a','fixture-b')})
Observe 'application-meeting-list' 'ConfigurationGetConfigurations' @('ApplicationMeetingConfiguration',$null,$null) 200 '[]'
Observe 'application-meeting-read' 'ConfigurationGetConfiguration' @('ApplicationMeetingConfiguration','Global',$null,$null) 200 '{"Identity":"Global","AllowRemoveParticipantAppIds":[]}'
Observe 'application-meeting-update' 'ConfigurationUpdateConfiguration' @('ApplicationMeetingConfiguration','Global',$null,$null,$model) 204 ''
$models=[Collections.Generic.List[object]]::new()
foreach($inputCase in @(
    @{name='omitted';bound=$false;value=$null},
    @{name='scalar';bound=$true;value='fixture-app'},
    @{name='replacement';bound=$true;value=@('fixture-a','fixture-b')},
    @{name='empty-replacement';bound=$true;value=@()},
    @{name='null-replacement';bound=$true;value=$null},
    @{name='operators';bound=$true;value=@{Add=@('fixture-a');Remove=@('fixture-b')}},
    @{name='empty-operators';bound=$true;value=@{Add=@();Remove=@()}},
    @{name='legacy-object';bound=$true;value=@([pscustomobject]@{AppId='fixture-app';DataSource='fixture-source';Key='fixture-key';Enabled=$false})},
    @{name='replace-operator-rejected';bound=$true;value=@{Replace=@('fixture-a')}},
    @{name='numeric-item-rejected';bound=$true;value=@(123)},
    @{name='ps-list-modifier-rejected';bound=$true;value=[Management.Automation.PSListModifier]::new([object]@('fixture-app'))}
)) {
    $bag=[Collections.Generic.Dictionary[string,object]]::new()
    $bag.Add('Identity','Global')
    if($inputCase.bound){$bag.Add('AllowRemoveParticipantAppIds',$inputCase.value)}
    $rejected=$false
    try{[Microsoft.Teams.ConfigApi.Cmdlets.PolicyUtils]::ValidateAndUpdatePsListModifier('AllowRemoveParticipantAppIds',$bag,[string])}catch{$rejected=$true}
    $wire=$null
    if(!$rejected){$value=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.XdsConfiguration]::DeserializeFromDictionary($bag);$wire=ConvertFrom-Json -InputObject $value.ToJsonString() -AsHashtable}
    $models.Add(@{name=$inputCase.name;rejected=$rejected;wire=$wire})
}
$artifact=[ordered]@{
    schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
    qualification='offline-application-meeting-helper-and-dispatch';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
    assembly=@{path=$assemblyPath;sha256=$expected}
    stage='Generated methods captured before session host/header rewrite; pure list normalization and model serialization. No authentication, discovery, command processing or HTTP client.'
    cases=$cases.ToArray();models=$models.ToArray()
}
$artifact | ConvertTo-Json -Depth 30 | Set-Content -LiteralPath $OutputPath -Encoding utf8
"application meeting observations: $($cases.Count) requests; $($models.Count) models"
