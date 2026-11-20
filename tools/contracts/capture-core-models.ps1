# Offline development oracle: local model construction and serialization only.
[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$ModulePath,
    [Parameter(Mandatory)][string]$OutputPath
)
$ErrorActionPreference = 'Stop'
if (Test-Path -LiteralPath $OutputPath) { throw 'Output already exists; choose a new candidate path.' }
$assemblies = @(
    @{path='netcoreapp3.1/Microsoft.Teams.PowerShell.TeamsCmdlets.dll';sha256='3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24'},
    @{path='bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll';sha256='8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'}
)
foreach ($assembly in $assemblies) {
    if ((Get-FileHash -LiteralPath (Join-Path $ModulePath $assembly.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $assembly.sha256) {
        throw 'Assembly does not match the pinned 8.1.0 model oracle.'
    }
}
$moduleFile = Join-Path $ModulePath 'MicrosoftTeams.psd1'
if ((Import-PowerShellDataFile -LiteralPath $moduleFile).ModuleVersion -ne '8.1.0') { throw 'Module version is not 8.1.0.' }
Import-Module $moduleFile -Force
$cases = [Collections.Generic.List[object]]::new()
$settings = [Microsoft.Teams.PowerShell.TeamsCmdlets.Utils.HttpUtilities].GetField('JsonSerializerSettings',[Reflection.BindingFlags]'NonPublic,Static').GetValue($null)
function Add-Observed($name, $source, $inputValue, [string]$json) {
    $cases.Add(@{name=$name; source=$source; input=$inputValue; wire=(ConvertFrom-Json -InputObject $json -AsHashtable)})
}
function Serialize-Graph($value) {
    return [Newtonsoft.Json.JsonConvert]::SerializeObject($value.PSObject.BaseObject,$settings)
}

# Channel creation defaults, explicit values, private owner and sparse patch.
$graphAssembly = [Microsoft.Teams.PowerShell.TeamsCmdlets.NewTeamChannel].Assembly
$channel = [Activator]::CreateInstance($graphAssembly.GetType('Microsoft.Teams.PowerShell.TeamsCmdlets.NewTeamChannel+ChannelModelForCreation'),$true)
$channel.DisplayName = 'Fixture'
Add-Observed 'channel-create-default' 'NewTeamChannel.ChannelModelForCreation; HttpUtilities.JsonSerializerSettings' @{DisplayName='Fixture'} (Serialize-Graph $channel)
$channel.Description = ''
$channel.MembershipType = [Microsoft.Teams.PowerShell.TeamsCmdlets.Constants.ChannelMembershipType]::Private
$member = [Activator]::CreateInstance($graphAssembly.GetType('Microsoft.Teams.PowerShell.TeamsCmdlets.NewTeamChannel+AadUserModelForAdd'),$true)
$member.GetType().GetProperty('UserId',[Reflection.BindingFlags]'Public,Instance,DeclaredOnly').SetValue($member,'00000000-0000-4000-8000-000000000001')
$member.GetType().GetProperty('Roles',[Reflection.BindingFlags]'Public,Instance,DeclaredOnly').SetValue($member,[string[]]@('owner'))
$channel.Members = [Microsoft.Teams.PowerShell.TeamsCmdlets.Model.GraphAadUserConversationMember[]]@($member)
Add-Observed 'channel-create-private-owner' 'NewTeamChannel.ChannelModelForCreation; AadUserModelForAdd' @{DisplayName='Fixture';Description='';MembershipType='Private';Owner='00000000-0000-4000-8000-000000000001'} (Serialize-Graph $channel)
$patch = [Activator]::CreateInstance($graphAssembly.GetType('Microsoft.Teams.PowerShell.TeamsCmdlets.SetTeamChannel+TeamChannelModelForPatch'),$true)
Add-Observed 'channel-patch-omitted' 'SetTeamChannel.TeamChannelModelForPatch' @{} (Serialize-Graph $patch)
$patch.Description = ''
Add-Observed 'channel-patch-empty-description' 'SetTeamChannel.TeamChannelModelForPatch' @{Description=''} (Serialize-Graph $patch)

# EDU group mutations are a pure helper, never ProcessRecord/CreateGroup.
$newTeam = [Microsoft.Teams.PowerShell.TeamsCmdlets.NewTeam]::new()
foreach ($template in @('EDU_Class','EDU_PLC')) {
    $group = [Microsoft.Teams.PowerShell.TeamsCmdlets.Model.Group]::new()
    $group.Visibility = 'Private'
    $kind = [Enum]::Parse([Microsoft.Teams.PowerShell.TeamsCmdlets.Model.TemplateType],$template)
    $result = $newTeam.SetGroupPropertiesBasedOnTemplate($group,$kind)
    Add-Observed ('group-template-' + $template) 'NewTeam.SetGroupPropertiesBasedOnTemplate; Model.Group' @{Template=$template;Visibility='Private'} (Serialize-Graph $result)
}

# Generated Config API models have their own serializer, not Newtonsoft's settings.
$template = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.TeamTemplate]::new()
$template.DisplayName = 'Fixture'
$template.ShortDescription = 'Fixture summary'
Add-Observed 'template-expanded-sparse' 'Models.TeamTemplate.ToJson' @{DisplayName='Fixture';ShortDescription='Fixture summary'} $template.ToJsonString()
$template.IsMembershipLimitedToOwner = $false
$template.Channel = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.IChannelTemplate[]]@()
$template.App = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.ITeamsAppTemplate[]]@()
$template.Category = [string[]]@()
Add-Observed 'template-expanded-empty-arrays-false' 'Models.TeamTemplate.ToJson' @{DisplayName='Fixture';ShortDescription='Fixture summary';IsMembershipLimitedToOwner=$false;Channel=@();App=@();Category=@()} $template.ToJsonString()
$jsonInput = '{"displayName":"Fixture","shortDescription":"Fixture summary"}'
$parsed = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.TeamTemplate]::FromJsonString($jsonInput)
Add-Observed 'template-json-body-default-settings' 'Models.TeamTemplate.FromJsonString/ToJson' $jsonInput $parsed.ToJsonString()
$jsonInput = '{"teamDefinition":{"displayName":"Definition","channels":[]},"extendedProperties":{"displayName":"Extended","shortDescription":"Summary"}}'
$parsed = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.TeamTemplate]::FromJsonString($jsonInput)
Add-Observed 'template-json-flattened-precedence' 'Models.TeamTemplate.BeforeFromJson/ToJson' $jsonInput $parsed.ToJsonString()
$jsonInput = '{"displayName":"Fixture","shortDescription":"Summary","channels":[{"displayName":"General","tabs":[{"name":"Fixture tab"}]}]}'
$parsed = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.TeamTemplate]::FromJsonString($jsonInput)
Add-Observed 'template-json-tab-configuration-omitted' 'Models.TeamTemplate.FromJsonString; ChannelTabTemplate' $jsonInput $parsed.ToJsonString()
$policy = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.XdsConfiguration]::DeserializeFromDictionary(@{Identity='Global';CustomTagsMode='Enabled';Description='';SuggestedPresetTags=$null})
Add-Observed 'targeting-policy-sparse-empty-null' 'Models.XdsConfiguration.DeserializeFromDictionary/ToJson' @{Identity='Global';CustomTagsMode='Enabled';Description='';SuggestedPresetTags=$null} $policy.ToJsonString()
$batch = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.FlwoServiceModelsFlwosBatchRequest]::new()
$batch.DeploymentCsv = 'Rml4dHVyZQ=='
$batch.UsersToNotify = 'fixture@example.invalid'
Add-Observed 'deployment-wire-field-casing' 'Models.FlwoServiceModelsFlwosBatchRequest.ToJson' @{DeploymentCsvs='Rml4dHVyZQ==';UsersToNotify='fixture@example.invalid'} $batch.ToJsonString()

# Synthetic responses are decoded and reserialized through the shipped models.
# This records model behavior, not a server response or HTTP error classification.
$generated = [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.TeamTemplate].Assembly
foreach ($response in @(
    @{name='template-error-envelope';type='TeamTemplateErrorResponse';json='{"operationId":"fixture-operation","errors":[{"code":"FixtureFailure","message":"Fixture message","action":"Fixture action"}],"futureField":true}'},
    @{name='template-error-empty-errors';type='TeamTemplateErrorResponse';json='{"errors":[]}'},
    @{name='template-list-response';type='GetTemplatesResponse';json='{"value":[{"id":"fixture","name":"Fixture","channelCount":1,"appCount":0,"locale":"en-US","@odata.id":"/api/teamtemplates/v1.0/fixture/Tenant/en-US"}]}'},
    @{name='template-detail-response';type='GetTemplateDetailResponse';json='{"value":{"teamDefinition":{"displayName":"Fixture","channels":[]},"extendedProperties":{"shortDescription":"Summary"}}}'},
    @{name='deployment-accepted-response';type='FlwoServiceModelsFlwosBatchRequestResponse';json='{"instanceId":"fixture-instance","executionId":"fixture-execution"}'},
    @{name='policy-error-envelope';type='ErrorResponse';json='{"code":"FixtureFailure","message":"Fixture message","action":"Fixture action","futureField":true}'},
    @{name='policy-unknown-fields-preserved';type='XdsConfiguration';json='{"Identity":"Global","FutureFlag":false,"FutureValues":[]}'},
    @{name='template-error-empty-object';type='TeamTemplateErrorResponse';json='{}'}
)) {
    $type = $generated.GetType('Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.' + $response.type, $true)
    $decoded = $type.GetMethod('FromJsonString').Invoke($null,[object[]]@($response.json))
    Add-Observed $response.name ('Models.' + $response.type + '.FromJsonString/ToJsonString') $response.json $decoded.ToJsonString()
}
foreach ($status in @('NotStarted','Running','Succeeded','Failed','UnknownFutureValue','FutureStatus')) {
    $jsonInput = '{"id":"fixture-operation","status":"' + $status + '","lastActionDateTime":"2026-01-01T00:00:00Z","fileName":"fixture.csv"}'
    $decoded = [Newtonsoft.Json.JsonConvert]::DeserializeObject($jsonInput,[Microsoft.Teams.PowerShell.TeamsCmdlets.Model.HierarchyServiceOperation])
    $projection = [Microsoft.Teams.PowerShell.TeamsCmdlets.GetTeamTargetingHierarchyStatus+Result]::new($decoded)
    Add-Observed ('hierarchy-status-projection-' + $status) 'GetTeamTargetingHierarchyStatus.Result(HierarchyServiceOperation)' $jsonInput (Serialize-Graph $projection)
}
$jsonInput = '{"id":"fixture-operation","status":"failed","operationType":"createTeam","attemptsCount":1,"createdDateTime":"2026-01-01T00:00:00Z","lastActionDateTime":"2026-01-01T00:00:01Z","targetResourceId":"fixture-team","error":{"code":"FixtureFailure","message":"Fixture message"}}'
$decoded = [Newtonsoft.Json.JsonConvert]::DeserializeObject($jsonInput,[Microsoft.Teams.PowerShell.TeamsCmdlets.Model.TeamsAsyncOperation])
Add-Observed 'team-operation-known-status-error-projection' 'Model.TeamsAsyncOperation; HttpUtilities.JsonSerializerSettings' $jsonInput (Serialize-Graph $decoded)
$jsonInput = '{"id":"fixture-operation","status":"futureStatus","operationType":"createTeam"}'
$rejected = $false
try { $null = [Newtonsoft.Json.JsonConvert]::DeserializeObject($jsonInput,[Microsoft.Teams.PowerShell.TeamsCmdlets.Model.TeamsAsyncOperation]) } catch { $rejected = $true }
$cases.Add(@{name='team-operation-unknown-status';source='Model.TeamsAsyncOperation.StringEnumConverter';input=$jsonInput;wire=@{rejected=$rejected}})
$rejected, $returnedNull = $false, $false
try { $returnedNull = $null -eq [Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.TeamTemplateErrorResponse]::FromJsonString('') } catch { $rejected = $true }
$cases.Add(@{name='template-empty-text-parser';source='Models.TeamTemplateErrorResponse.FromJsonString';input='';wire=@{rejected=$rejected;returnedNull=$returnedNull}})

$artifact = [ordered]@{
    schema_version=1
    baseline_version='8.1.0'
    captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
    qualification='offline-model-serialization'
    package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
    assemblies=$assemblies
    purpose='Synthetic model observations; deployment Base64 is a field-casing fixture, not a valid ZIP or service acceptance claim.'
    cases=$cases.ToArray()
}
$artifact | ConvertTo-Json -Depth 30 | Set-Content -LiteralPath $OutputPath -Encoding utf8
"core model observations: $($cases.Count)"
