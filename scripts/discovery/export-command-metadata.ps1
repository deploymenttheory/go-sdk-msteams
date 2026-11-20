param(
    [Parameter(Mandatory = $true)][string]$ModulePath,
    [Parameter(Mandatory = $true)][string]$OutputPath,
    [string]$ExpectedVersion = '8.1.0'
)
$ErrorActionPreference = 'Stop'
Import-Module $ModulePath -ErrorAction Stop
$module = Get-Module MicrosoftTeams
$observedVersion = $module.Version.ToString()
$prerelease = $module.PrivateData.PSData.Prerelease
if ($prerelease) { $observedVersion += '-' + $prerelease.TrimStart('-') }
if ($observedVersion -ne $ExpectedVersion) { throw "Expected MicrosoftTeams $ExpectedVersion" }
$commands = @(Get-Command -Module MicrosoftTeams | Sort-Object Name | ForEach-Object {
    $command = $_
    [ordered]@{
        name = $command.Name
        command_type = $command.CommandType.ToString()
        output_types = @($command.OutputType | ForEach-Object { $_.Name } | Sort-Object -Unique)
        parameter_sets = @($command.ParameterSets | Sort-Object Name | ForEach-Object {
            [ordered]@{
                name = $_.Name
                is_default = $_.IsDefault
                parameters = @($_.Parameters | Sort-Object Name | ForEach-Object {
                    [ordered]@{
                        name = $_.Name
                        type = $_.ParameterType.FullName
                        mandatory = $_.IsMandatory
                        position = $_.Position
                        aliases = @($_.Aliases | Sort-Object)
                        pipeline_input = $_.ValueFromPipeline
                        pipeline_by_property = $_.ValueFromPipelineByPropertyName
                        validation = @($_.Attributes | Where-Object { $_ -is [System.Management.Automation.ValidateArgumentsAttribute] } | ForEach-Object {
                            $attribute = $_
                            $properties = [ordered]@{}
                            foreach ($property in @('ValidValues','MinRange','MaxRange','MinLength','MaxLength','RegexPattern','Options','IgnoreCase','ErrorMessage')) {
                                if ($attribute.PSObject.Properties[$property]) {
                                    $properties[$property] = $attribute.$property
                                }
                            }
                            [ordered]@{ type = $attribute.GetType().FullName; properties = $properties }
                        })
                    }
                })
            }
        })
    }
})
if ($commands.Count -eq 0) { throw 'Module import exposed no commands' }
[ordered]@{
    module_version = $observedVersion
    powershell_version = $PSVersionTable.PSVersion.ToString()
    platform = [System.Runtime.InteropServices.RuntimeInformation]::OSDescription
    authenticated = $false
    commands = $commands
} | ConvertTo-Json -Depth 30 | Set-Content -Encoding utf8 $OutputPath
Write-Output "Exported $($commands.Count) commands"
