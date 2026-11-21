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
    @{name='scalar';bound=$true;value='future-phone'},
    @{name='replacement';bound=$true;value=@('+440000000001','not-a-phone')},
    @{name='empty-replacement';bound=$true;value=@()},
    @{name='null-replacement';bound=$true;value=$null},
    @{name='operators';bound=$true;value=@{Add=@('+440000000001');Remove=@('not-a-phone')}},
    @{name='empty-operators';bound=$true;value=@{Add=@();Remove=@()}},
    @{name='legacy-object-replacement';bound=$true;value=@([pscustomobject]@{AppId='future-phone';DataSource='fixture-source';Key='fixture-key';Enabled=$false})},
    @{name='legacy-object-operators';bound=$true;value=@{Add=@([pscustomobject]@{AppId='future-phone';DataSource='fixture-source';Key='fixture-key';Enabled=$false})}},
    @{name='replace-operator-rejected';bound=$true;value=@{Replace=@('+440000000001')}},
    @{name='null-item-rejected';bound=$true;value=@('future-phone',$null)},
    @{name='empty-string';bound=$true;value=''},
    @{name='whitespace-string';bound=$true;value=' '},
    @{name='numeric-item-rejected';bound=$true;value=@(123)},
    @{name='ps-list-modifier-rejected';bound=$true;value=[Management.Automation.PSListModifier]::new([object]@('future-phone'))}
)) {
    $bag=[Collections.Generic.Dictionary[string,object]]::new()
    $bag.Add('Identity','tag:Fixture')
    if($inputCase.bound){$bag.Add('MeetingInvitePhoneNumbers',$inputCase.value)}
    $rejected=$false
    try{[Microsoft.Teams.ConfigApi.Cmdlets.PolicyUtils]::ValidateAndUpdatePsListModifier('MeetingInvitePhoneNumbers',$bag,[string])}catch{$rejected=$true}
    $wire=$null
    if(!$rejected){$model=[Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.XdsConfiguration]::DeserializeFromDictionary($bag);$wire=ConvertFrom-Json -InputObject $model.ToJsonString() -AsHashtable}
    $cases.Add(@{name=('phone-list-'+$inputCase.name);source='PolicyUtils.ValidateAndUpdatePsListModifier; XdsConfiguration.DeserializeFromDictionary/ToJsonString';rejected=$rejected;wire=$wire})
}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-audio-conferencing-policy-normalization';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;purpose='Actual pinned PolicyUtils normalization and XdsConfiguration serialization of synthetic MeetingInvitePhoneNumbers bags. No command lifecycle, credentials, discovery or request sender.'
 cases=$cases.ToArray()
}|ConvertTo-Json -Depth 30|Set-Content -LiteralPath $OutputPath -Encoding utf8
"audio conferencing policy normalization observations: $($cases.Count)"
