# Offline development oracle only. Does not invoke management cmdlets or connect.
[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$ModulePath,
    [Parameter(Mandatory)][string]$OutputPath
)
$ErrorActionPreference='Stop'
if (Test-Path -LiteralPath $OutputPath) { throw 'Output already exists; choose a new candidate path.' }
$moduleFile = Join-Path $ModulePath 'MicrosoftTeams.psd1'
$assemblyFile = Join-Path $ModulePath 'netcoreapp3.1/Microsoft.Teams.PowerShell.TeamsCmdlets.dll'
$expected = '3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24'
if ((Get-FileHash -LiteralPath $assemblyFile -Algorithm SHA256).Hash.ToLowerInvariant() -ne $expected) {
    throw 'Assembly does not match the pinned 8.1.0 serializer oracle.'
}
$manifest = Import-PowerShellDataFile -LiteralPath $moduleFile
if ($manifest.ModuleVersion -ne '8.1.0') { throw 'Module version is not 8.1.0.' }
Import-Module $moduleFile -Force
$settings=[Microsoft.Teams.PowerShell.TeamsCmdlets.Utils.HttpUtilities].GetField('JsonSerializerSettings',[Reflection.BindingFlags]'NonPublic,Static').GetValue($null)
function Serialize($value) { return [Newtonsoft.Json.JsonConvert]::SerializeObject($(if($null -eq $value){$null}else{$value.PSObject.BaseObject}),$settings)|ConvertFrom-Json -AsHashtable }
$cases=@()
$empty=[Microsoft.Teams.PowerShell.TeamsCmdlets.Model.TeamSettings]::new()
$cases += @{name='team-settings-omitted'; input=@{}; wire=(Serialize $empty.ToTeam())}
$explicit=[Microsoft.Teams.PowerShell.TeamsCmdlets.Model.TeamSettings]::new()
$explicit.AllowGiphy=$false
$explicit.AllowGuestDeleteChannels=$false
$explicit.AllowCreatePrivateChannels=$true
$explicit.AllowUserEditMessages=$false
$explicit.ShowInTeamsSearchAndSuggestions=$false
$explicit.GiphyContentRating=[Microsoft.Teams.PowerShell.TeamsCmdlets.Model.GiphyRatingType]::moderate
$cases += @{name='team-settings-explicit-values';input=@{AllowGiphy=$false;AllowGuestDeleteChannels=$false;AllowCreatePrivateChannels=$true;AllowUserEditMessages=$false;ShowInTeamsSearchAndSuggestions=$false;GiphyContentRating='moderate'};wire=(Serialize $explicit.ToTeam())}
$group=[Microsoft.Teams.PowerShell.TeamsCmdlets.Model.Group]::new()
$group.DisplayName=''
$group.Description=''
$group.Classification=''
$group.MailNickname=''
$group.Visibility=''
$cases += @{name='group-patch-explicit-empty';input=@{DisplayName='';Description='';Classification='';MailNickname='';Visibility=''};wire=(Serialize $group)}
$group=[Microsoft.Teams.PowerShell.TeamsCmdlets.Model.Group]::new()
$cases += @{name='group-patch-omitted';input=@{};wire=(Serialize $group)}
$group.DisplayName='Fixture Team'
$group.GroupTypes=[Collections.Generic.List[string]]@('Unified')
$group.MailEnabled=$true
$group.SecurityEnabled=$false
$group.MailNickname='fixture-team'
$group.Visibility='Private'
$group.Members=[Collections.Generic.List[string]]@('https://graph.microsoft.com/v1.0/users/00000000-0000-0000-0000-000000000001')
$group.Owners=[Collections.Generic.List[string]]@('https://graph.microsoft.com/v1.0/users/00000000-0000-0000-0000-000000000001')
$group.CreationOptions=[Collections.Generic.List[string]]@('ExchangeProvisioningFlags:3552')
$cases += @{name='ordinary-group-create';input=@{DisplayName='Fixture Team';MailNickName='fixture-team';ownerId='00000000-0000-0000-0000-000000000001'};wire=(Serialize $group)}
$cases += @{name='archive-option-omitted';input=@{Archived=$true};wire=(Serialize $null)}
$archive=[Collections.Generic.Dictionary[string,bool]]::new()
$archive.Add('shouldSetSpoSiteReadOnlyForMembers',$false)
$cases += @{name='archive-option-false';input=@{Archived=$true;SetSpoSiteReadOnlyForMembers=$false};wire=(Serialize $archive)}
$cases += @{name='unarchive-body';input=@{Archived=$false};wire=(Serialize '')}
foreach ($case in $cases) {
    $case.purpose = 'Pinned serialization evidence; not a live server acceptance claim.'
    if ($case.name -eq 'archive-option-false') {
        $case.capture = 'Equivalent dictionary serialized with the actual HttpUtilities settings; the cmdlet constructs an anonymous CLR object with the same sole property.'
    } else {
        $case.capture = 'Actual pinned model and HttpUtilities serializer settings, or exact primitive argument passed to HttpUtilities.Post.'
    }
}
$artifact = [ordered]@{
    schema_version = 1
    baseline_version = '8.1.0'
    captured_at = [DateTime]::UtcNow.ToString('yyyy-MM-dd')
    qualification = 'offline-model-serialization'
    package_sha256 = '78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
    assembly = @{path='netcoreapp3.1/Microsoft.Teams.PowerShell.TeamsCmdlets.dll';sha256=$expected}
    methods = @('Model.TeamSettings.ToTeam','Utils.HttpUtilities.JsonSerializerSettings','NewTeam.CreateGroup','SetTeam.ProcessRecord','SetTeamArchivedState.ProcessRecord')
    cases = $cases
}
$artifact|ConvertTo-Json -Depth 20|Set-Content -LiteralPath $OutputPath -Encoding utf8

"model serialization fixtures: $($cases.Count)"
