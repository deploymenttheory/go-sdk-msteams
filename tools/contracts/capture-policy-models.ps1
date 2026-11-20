# Offline policy helper/model oracle. No command processor, token or HTTP helper.
[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$ModulePath,
    [Parameter(Mandatory)][string]$OutputPath
)
$ErrorActionPreference='Stop'
if (Test-Path -LiteralPath $OutputPath) { throw 'Output already exists; choose a new candidate path.' }
$assemblies=@(
    @{path='bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll';sha256='8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'},
    @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
    @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'}
)
foreach($assembly in $assemblies) {
    if((Get-FileHash -LiteralPath (Join-Path $ModulePath $assembly.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $assembly.sha256){throw 'Policy assembly differs from pinned baseline.'}
}
$moduleFile=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $moduleFile).ModuleVersion -ne '8.1.0'){throw 'Module version is not 8.1.0.'}
Import-Module $moduleFile -Force
$cases=[Collections.Generic.List[object]]::new()
foreach($inputCase in @(
    @{name='omitted';bound=$false;value=$null},
    @{name='scalar';bound=$true;value='fixture-app'},
    @{name='replacement';bound=$true;value=@('fixture-a','fixture-b')},
    @{name='empty-replacement';bound=$true;value=@()},
    @{name='null-replacement';bound=$true;value=$null},
    @{name='operators';bound=$true;value=@{Add=@('fixture-a');Remove=@('fixture-b')}},
    @{name='empty-operators';bound=$true;value=@{Add=@();Remove=@()}},
    @{name='legacy-object-replacement';bound=$true;value=@([pscustomobject]@{AppId='fixture-app';DataSource='fixture-source';Key='fixture-key';Enabled=$false})},
    @{name='legacy-object-operators';bound=$true;value=@{Add=@([pscustomobject]@{AppId='fixture-app';DataSource='fixture-source';Key='fixture-key';Enabled=$false})}},
    @{name='replace-operator-rejected';bound=$true;value=@{Replace=@('fixture-a')}},
    @{name='numeric-item-rejected';bound=$true;value=@(123)},
    @{name='ps-list-modifier-rejected';bound=$true;value=[Management.Automation.PSListModifier]::new([object]@('fixture-app'))}
)) {
    $bag=[Collections.Generic.Dictionary[string,object]]::new()
    $bag.Add('Identity','tag:Fixture')
    if($inputCase.bound){$bag.Add('AppIds',$inputCase.value)}
    $rejected=$false
    try{[Microsoft.Teams.ConfigApi.Cmdlets.PolicyUtils]::ValidateAndUpdatePsListModifier('AppIds',$bag,[string])}catch{$rejected=$true}
    $wire=$null
    if(!$rejected){$model=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.XdsConfiguration]::DeserializeFromDictionary($bag);$wire=ConvertFrom-Json -InputObject $model.ToJsonString() -AsHashtable}
    $cases.Add(@{name=('appids-'+$inputCase.name);source='PolicyUtils.ValidateAndUpdatePsListModifier; XdsConfiguration.DeserializeFromDictionary/ToJsonString';rejected=$rejected;wire=$wire})
}
$request=[Microsoft.Teams.Policy.Administration.Cmdlets.Core.Custom.GPA.Models.DirectToGroupAssignmentMigration]::new()
$request.Authority='Tenant';$request.PolicyType='FixturePolicy';$request.PolicyName='Fixture';$request.GroupId=[Guid]'00000000-0000-4000-8000-000000000001'
$wire=[Newtonsoft.Json.JsonConvert]::SerializeObject($request.PSObject.BaseObject)
$cases.Add(@{name='migration-request-field-casing';source='GPA.Models.DirectToGroupAssignmentMigration.JsonProperty';wire=(ConvertFrom-Json -InputObject $wire -AsHashtable)})
foreach($state in @('NotStarted','PopulationInProgress','GpaCreationInProgress','WaitingForGpaProcessed','CleanUpInProgress','Completed','Failed','ConnectionLost')){
    $json='{"migrationEventId":"00000000-0000-4000-8000-000000000002","groupId":"00000000-0000-4000-8000-000000000001","state":"'+$state+'","directAssignmentsProcessed":0,"startTime":"2026-01-01T00:00:00Z"}'
    $model=[Newtonsoft.Json.JsonConvert]::DeserializeObject($json,[Microsoft.Ic3.AdminConfig.RP.Policy.Group.Models.DirectToGroupAssignmentMigrationResponseModel])
    $wire=[Newtonsoft.Json.JsonConvert]::SerializeObject($model.PSObject.BaseObject)
    $cases.Add(@{name=('migration-state-'+$state);source='DirectToGroupAssignmentMigrationResponseModel; DirectToGroupAssignmentMigrationState';input=$json;wire=(ConvertFrom-Json -InputObject $wire -AsHashtable)})
}
# Reflect only pure private serialization/escaping helpers, never request helpers.
$providerAssembly=[Microsoft.Teams.Policy.Administration.Cmdlets.Providers.BasePolicy].Assembly
$helperType=$providerAssembly.GetType('Microsoft.Teams.Policy.Administration.Cmdlets.Providers.HttpRequestHelper`1')
$resolverType=$providerAssembly.GetType('Microsoft.Teams.Policy.Administration.Cmdlets.Providers.Helpers.BoundParametersContractResolver')
$privateStatic=[Reflection.BindingFlags]'NonPublic,Static'
function ObserveProvider([string]$name,[object]$model,[hashtable]$bound){
    $bag=[Collections.Generic.Dictionary[string,object]]::new()
    foreach($key in $bound.Keys){$bag.Add($key,$bound[$key])}
    $closedHelper=$helperType.MakeGenericType([type[]]@($model.GetType()))
    $nested=$closedHelper.GetMethod('GetNestedBoundParameters',$privateStatic).Invoke($null,[object[]]@($bag,$model.GetType(),$false))
    $resolver=[Activator]::CreateInstance($resolverType,[Reflection.BindingFlags]'Instance,Public,NonPublic',$null,[object[]]@($nested),$null)
    $settings=[Newtonsoft.Json.JsonSerializerSettings]::new()
    $settings.ContractResolver=$resolver;$settings.MaxDepth=128
    $wire=[Newtonsoft.Json.JsonConvert]::SerializeObject($model.PSObject.BaseObject,$settings)
    $cases.Add(@{name=$name;source='HttpRequestHelper.GetNestedBoundParameters; BoundParametersContractResolver';wire=(ConvertFrom-Json -InputObject $wire -AsHashtable)})
}
$external=[Microsoft.Teams.Policy.Administration.Cmdlets.Core.ExternalAccessPolicy]::new()
$external.Identity='tag:Fixture';$external.Description='';$external.EnableFederationAccess=$false
ObserveProvider 'provider-bound-omitted' $external @{Identity='tag:Fixture'}
ObserveProvider 'provider-bound-empty-and-false' $external @{Identity='tag:Fixture';Description='';EnableFederationAccess=$false}
foreach($enum in [Enum]::GetValues([Microsoft.Teams.Policy.Administration.Cmdlets.Core.CommunicationWithExternalOrgs])){
    $external.CommunicationWithExternalOrgs=$enum
    ObserveProvider ('provider-external-enum-'+$enum) $external @{Identity='tag:Fixture';CommunicationWithExternalOrgs=$enum}
}
$voicemail=[Microsoft.Teams.Policy.Administration.Cmdlets.Core.OnlineVoicemailPolicy]::new()
$voicemail.Identity='tag:Fixture';$voicemail.MaximumRecordingLength=[TimeSpan]::FromSeconds(90)
ObserveProvider 'provider-voicemail-timespan' $voicemail @{Identity='tag:Fixture';MaximumRecordingLength=$voicemail.MaximumRecordingLength}
$voicemail.MaximumRecordingLength=$null
ObserveProvider 'provider-voicemail-explicit-null' $voicemail @{Identity='tag:Fixture';MaximumRecordingLength=$null}
$branding=[Microsoft.Teams.Policy.Administration.Cmdlets.Core.TeamsMeetingBrandingPolicy]::new()
$branding.Identity='tag:Fixture'
ObserveProvider 'provider-branding-empty-lists' $branding @{Identity='tag:Fixture';MeetingBrandingThemes=@();MeetingBackgroundImages=@();NdiAssuranceSlateImages=@()}
$image=[Microsoft.Teams.Policy.Administration.Cmdlets.Core.MeetingBackgroundImage]::new('fixture-image',0,'Fixture','https://example.invalid/background.png')
$image.IsRequired=$false;$branding.MeetingBackgroundImages.Add($image)
$theme=[Microsoft.Teams.Policy.Administration.Cmdlets.Core.MeetingBrandingTheme]::new('fixture-theme')
$theme.Enabled=$false;$branding.MeetingBrandingThemes.Add($theme)
$slate=[Microsoft.Teams.Policy.Administration.Cmdlets.Core.NdiAssuranceSlate]::new('fixture-slate','Fixture','https://example.invalid/slate.png')
$slate.IsDefault=$false;$branding.NdiAssuranceSlateImages.Add($slate)
ObserveProvider 'provider-branding-nested-lists' $branding @{Identity='tag:Fixture';MeetingBrandingThemes=@($theme);MeetingBackgroundImages=@($image);NdiAssuranceSlateImages=@($slate)}
foreach($grant in @(
    @{name='user-clear';model=[Microsoft.Teams.Policy.Administration.Cmdlets.Providers.SingleGrantPayload]::new()},
    @{name='tenant-clear';model=[Microsoft.Teams.Policy.Administration.Cmdlets.Providers.GrantToTenantPayload]::new()},
    @{name='group-clear-zero-rank';model=[Microsoft.Teams.Policy.Administration.Cmdlets.Providers.GroupGrantPayload]::new()}
)){
    if($grant.name -eq 'tenant-clear'){$grant.model.IsForcePresent=$false}
    if($grant.name -eq 'group-clear-zero-rank'){$grant.model.Priority=0}
    $wire=[Newtonsoft.Json.JsonConvert]::SerializeObject($grant.model.PSObject.BaseObject)
    $cases.Add(@{name=('provider-'+$grant.name);source='Provider grant DTO; JsonConvert.SerializeObject';wire=(ConvertFrom-Json -InputObject $wire -AsHashtable)})
}
$closedHelper=$helperType.MakeGenericType([type[]]@($external.GetType()))
$escape=$closedHelper.GetMethod('EscapeIdentityForUri',$privateStatic)
foreach($identity in @(@{name='tagged';value='tag:Fixture / Policy'},@{name='nested';value='parent/Fixture / Child'})){
    $cases.Add(@{name=('provider-identity-'+$identity.name);source='HttpRequestHelper.EscapeIdentityForUri';input=$identity.value;wire=$escape.Invoke($null,[object[]]@($identity.value))})
}
$artifact=[ordered]@{
    schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
    qualification='offline-policy-helper-and-model';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
    assemblies=$assemblies;purpose='Synthetic pure helper/model observations, not tenant response captures or server acceptance.';cases=$cases.ToArray()
}
$artifact|ConvertTo-Json -Depth 30|Set-Content -LiteralPath $OutputPath -Encoding utf8
"policy model observations: $($cases.Count)"
