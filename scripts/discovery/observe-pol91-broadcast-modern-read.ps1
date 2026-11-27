[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference = 'Stop'
$source = Join-Path $ModulePath 'custom/Merged_custom_PsExt.ps1'
$sourceHash = '03dbc3a85e12c98dd1807d604782c6a90d0656569541cd9540738825a7fcb255'
if ((Get-FileHash $source).Hash.ToLowerInvariant() -ne $sourceHash) { throw 'Pinned helper source differs' }
$tokens = $null; $errors = $null
$ast = [System.Management.Automation.Language.Parser]::ParseFile($source, [ref]$tokens, [ref]$errors)
if ($errors.Count) { throw 'Source parse failed' }
$names = @('Custom-ToString', 'Convert-PsCustomObjectToPsObject', 'Get-FormatsForConfig', 'Set-FormatOnConfigObject', 'Sort-GlobalFirst')
$functions = foreach ($name in $names) {
    $nodes = @($ast.FindAll({param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq $name}, $false))
    if ($nodes.Count -ne 1) { throw "Expected exactly one own helper: $name" }
    . ([scriptblock]::Create($nodes[0].Extent.Text))
    [pscustomobject]@{ name=$name; start_line=$nodes[0].Extent.StartLineNumber; end_line=$nodes[0].Extent.EndLineNumber }
}
# Execute the pure received-JSON/presentation stages from Get-CsConfigurationModern.
# This deliberately does not instantiate telemetry, invoke the management helper,
# choose a session flight, bind the generated HTTP command or contact a tenant.
$inputs = @('{}', 'null', '[]', '[null]', 'true', '37', '"text"',
    '{"Identity":"Global","Future":{"x":1}}',
    '[{"Identity":"other"},{"Identity":"global"},{"Identity":"Global"},{"Identity":"Tag:x"}]',
    '{"Identity":"first","identity":"second"}', '{"Identity":"first","Identity":"second"}')
foreach ($field in @('Identity','SupportURL','AllowSdnProviderForBroadcastMeeting','SdnName','SdnLicenseId','SdnAzureSubscriptionId','SdnApiTemplateUrl','SdnApiToken','SdnRuntimeconfiguration','SdnAttendeeFallbackCount')) {
    foreach ($token in @('null','true','false','0','1','9007199254740993','1.20','""','"text"','{}','[]')) {
        $inputs += '{"'+$field+'":'+$token+'}'
    }
}
$rows = foreach ($raw in $inputs) {
    foreach ($expose in @($false,$true)) {
        $failure=$null; $projected=$null; $count=$null; $stage='json-and-psserializer'
        try {
            $output = Convert-PsCustomObjectToPsObject (ConvertFrom-Json -InputObject $raw)
            $stage='format'
            $output = $output | ForEach-Object { Set-FormatOnConfigObject -ConfigObject $_ -ConfigType 'TeamsMeetingBroadcastConfiguration' }
            if ($expose) {
                $output = ($output | Select-Object Identity, SupportURL, AllowSdnProviderForBroadcastMeeting, SdnName, SdnLicenseId, SdnAzureSubscriptionId, SdnApiTemplateUrl, SdnApiToken, SdnRuntimeconfiguration, SdnAttendeeFallbackCount)
            }
            $stage='sort'
            $output = @(Sort-GlobalFirst $output)
            $count = $output.Count
            $projected = ConvertTo-Json -InputObject $output -Depth 12 -Compress
            $stage='complete'
        } catch { $failure=$_.Exception.GetBaseException().GetType().FullName }
        [pscustomobject]@{input_json=$raw;expose=$expose;stage=$stage;output_count=$count;projected_json=$projected;error=$failure}
    }
}
[pscustomobject]@{
    scope='Pure pinned modern received-JSON/presentation stages only. Not whole-command, generated response parser, flight dispatch, legacy or tenant execution.'
    helper_source_sha256=$sourceHash
    reproducer_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant()
    powershell_version=$PSVersionTable.PSVersion.ToString()
    extracted_functions=@($functions)
    observations=@($rows)
    management_lifecycle_invoked=$false
    authentication_or_http=$false
} | ConvertTo-Json -Depth 8
