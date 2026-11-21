# Offline model oracle. No connection or management cmdlet is invoked.
[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$ModulePath,
    [Parameter(Mandatory)][string]$OutputPath
)
$ErrorActionPreference = 'Stop'
if (Test-Path -LiteralPath $OutputPath) { throw 'Output already exists; choose a new candidate path.' }
$manifestPath = Join-Path $ModulePath 'MicrosoftTeams.psd1'
$assemblyPath = Join-Path $ModulePath 'netcoreapp3.1/Microsoft.Teams.PowerShell.TeamsCmdlets.dll'
$expectedHash = '3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24'
if ((Get-FileHash -LiteralPath $assemblyPath -Algorithm SHA256).Hash.ToLowerInvariant() -ne $expectedHash) { throw 'Wrong baseline assembly.' }
if ((Import-PowerShellDataFile -LiteralPath $manifestPath).ModuleVersion -ne '8.1.0') { throw 'Wrong baseline version.' }
Import-Module $manifestPath -Force
$settings = [Microsoft.Teams.PowerShell.TeamsCmdlets.Utils.HttpUtilities].GetField('JsonSerializerSettings', [Reflection.BindingFlags]'NonPublic,Static').GetValue($null)
$fields = @(
    'AllowCreateUpdateChannels', 'AllowDeleteChannels', 'AllowAddRemoveApps',
    'AllowCreateUpdateRemoveTabs', 'AllowCreateUpdateRemoveConnectors', 'AllowCreatePrivateChannels',
    'AllowGuestCreateUpdateChannels', 'AllowGuestDeleteChannels',
    'AllowUserEditMessages', 'AllowUserDeleteMessages', 'AllowOwnerDeleteMessages',
    'AllowTeamMentions', 'AllowChannelMentions', 'AllowGiphy',
    'AllowStickersAndMemes', 'AllowCustomMemes', 'ShowInTeamsSearchAndSuggestions'
)
$cases = @()
foreach ($field in $fields) {
    foreach ($value in @($false, $true)) {
        $model = [Microsoft.Teams.PowerShell.TeamsCmdlets.Model.TeamSettings]::new()
        $model.$field = $value
        $wire = [Newtonsoft.Json.JsonConvert]::SerializeObject($model.ToTeam(), $settings) | ConvertFrom-Json -AsHashtable
        $cases += [ordered]@{parameter=$field; value=$value; wire=$wire}
    }
}
foreach ($value in @('moderate', 'strict')) {
    $model = [Microsoft.Teams.PowerShell.TeamsCmdlets.Model.TeamSettings]::new()
    $model.GiphyContentRating = [Enum]::Parse([Microsoft.Teams.PowerShell.TeamsCmdlets.Model.GiphyRatingType], $value)
    $wire = [Newtonsoft.Json.JsonConvert]::SerializeObject($model.ToTeam(), $settings) | ConvertFrom-Json -AsHashtable
    $cases += [ordered]@{parameter='GiphyContentRating'; value=$value; wire=$wire}
}
$artifact = [ordered]@{
    schema_version=1
    baseline_version='8.1.0'
    qualification='offline-model-serialization'
    captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
    package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
    assembly=@{path='netcoreapp3.1/Microsoft.Teams.PowerShell.TeamsCmdlets.dll'; sha256=$expectedHash}
    methods=@('Model.TeamSettings.ToTeam', 'Utils.HttpUtilities.JsonSerializerSettings')
    purpose='Actual isolated model-field serialization; not live server acceptance or command execution.'
    cases=$cases
}
$artifact | ConvertTo-Json -Depth 20 | Set-Content -LiteralPath $OutputPath -Encoding utf8
"model field observations: $($cases.Count)"
