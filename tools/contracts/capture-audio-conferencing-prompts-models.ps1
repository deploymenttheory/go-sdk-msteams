[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblies=@(
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'},
 @{path='netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll';sha256='477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af'}
)
foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned audio conferencing prompt dependency.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
$env:MSTeamsContextInternal=$null
Import-Module $manifest -Force
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@((Join-Path $PSHOME 'System.Management.Automation.dll'))+@($assemblies|ForEach-Object{Join-Path $ModulePath $_.path})+@((Join-Path $PSHOME 'Newtonsoft.Json.dll'))
Add-Type -Path @((Join-Path $PSScriptRoot 'oracle/PolicyProviderCapture.cs'),(Join-Path $PSScriptRoot 'oracle/AudioConferencingPromptsCapture.cs')) -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$models=[Collections.Generic.List[object]]::new()
$inputs=@{
 configuration=@('constructor','{}','{"Identity":null,"Prompts":null,"Packages":null}','{"Prompts":[null],"Packages":[null]}','{"Prompts":[{}],"prompts":[{}],"Packages":[{}],"Packages":[{}]}','{"Prompts":{"Id":"bad-envelope"}}','{"Prompts":[{"Id":null,"Id":"11111111-1111-1111-1111-111111111111"}]}','{"Future":false}')
 prompt=@('constructor','{}','{"Identity":null,"Name":null,"AudioPrompt":null,"TextPrompt":null}','{"Id":null}','{"Type":null}','{"Locale":null}','{"Id":""}','{"Id":"future-id"}','{"Id":"11111111-1111-1111-1111-111111111111"}','{"Id":"ABCDEF12-1234-1234-1234-123456789ABC"}','{"Id":"11111111-1111-1111-1111-111111111111\n"}','{"Id":"11111111-1111-1111-1111-111111111111\r\n"}','{"Id":null,"Id":"11111111-1111-1111-1111-111111111111"}','{"Id":"11111111-1111-1111-1111-111111111111","Type":"","Locale":"future-locale","Name":"","AudioPrompt":"","TextPrompt":""}')
 package=@('constructor','{}','{"Id":null}','{"Id":""}','{"Id":"future-id"}','{"Id":"11111111-1111-1111-1111-111111111111"}','{"Id":"11111111-1111-1111-1111-111111111111\n"}','{"DefaultLocale":null}','{"DefaultLocale":""}','{"DefaultLocale":"future-locale","Name":null}','{"InboundStartRecordingPrompt":null,"InboundEndRecordingPrompt":null,"OutboundStartRecordingPrompt":null,"OutboundEndRecordingPrompt":null,"MeetingStartRecordingPrompt":null,"MeetingEndRecordingPrompt":null}','{"InboundStartRecordingPrompt":["a","a",null,""],"inboundStartRecordingPrompt":["b"]}','{"InboundStartRecordingPrompt":{}}','{"Id":null,"Id":"11111111-1111-1111-1111-111111111111"}')
}
foreach($kind in @('configuration','prompt','package')){
 foreach($modelInput in $inputs[$kind]){
  $models.Add(@{kind=$kind;input=$modelInput;observation=([TeamsContractOracle.AudioConferencingPromptsOracle]::Model($kind,$modelInput)|ConvertFrom-Json -AsHashtable)})
 }
}
$constructors=[Collections.Generic.List[object]]::new()
foreach($kind in @('prompt','package')){
 foreach($c in @(
  @{name='valid';id='11111111-1111-1111-1111-111111111111';type='FutureType';locale='xx-future';null_id=$false;null_type=$false;null_locale=$false},
  @{name='empty-fields';id='11111111-1111-1111-1111-111111111111';type='';locale='';null_id=$false;null_type=$false;null_locale=$false},
  @{name='null-id';id='';type='FutureType';locale='xx-future';null_id=$true;null_type=$false;null_locale=$false},
  @{name='invalid-id';id='future-id';type='FutureType';locale='xx-future';null_id=$false;null_type=$false;null_locale=$false},
  @{name='null-type-or-default-locale';id='11111111-1111-1111-1111-111111111111';type='';locale='xx-future';null_id=$false;null_type=$true;null_locale=$false},
  @{name='null-locale';id='11111111-1111-1111-1111-111111111111';type='FutureType';locale='';null_id=$false;null_type=$false;null_locale=$true}
 )){
  $observation=[TeamsContractOracle.AudioConferencingPromptsOracle]::Constructor($kind,$c.id,$c.type,$c.locale,$c.null_id,$c.null_type,$c.null_locale)|ConvertFrom-Json -AsHashtable
  $constructors.Add(@{kind=$kind;name=$c.name;id=$c.id;type_or_locale=$c.type;locale=$c.locale;null_id=$c.null_id;null_type_or_locale=$c.null_type;null_locale=$c.null_locale;observation=$observation})
 }
}
$lists=[Collections.Generic.List[object]]::new()
foreach($kind in @('prompt','package')){
 $names=@('null','empty-operators','replace-duplicates','replace-null','remove-one','remove-two','add-existing','add-null','remove-null','remove-and-add','remove-from-empty','null-existing')
 if($kind -eq 'prompt'){$names+=@('identity-ignored','different-Id','different-Name','different-AudioPrompt','different-TextPrompt','different-Type','different-Locale')}
 foreach($name in $names){
  $lists.Add(@{kind=$kind;name=$name;observation=([TeamsContractOracle.AudioConferencingPromptsOracle]::Lists($kind,$name)|ConvertFrom-Json -AsHashtable)})
 }
}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-audio-conferencing-prompts-models';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString()
 stage='Actual pinned constructors/setters/Newtonsoft models and Core PSListModifierHelper. Prompt comparator copies the six-field Set source. Package comparator throws like Set source. Null constructor inputs selected in C#. No command lifecycle, credentials, discovery or sender.'
 models=$models.ToArray();constructors=$constructors.ToArray();lists=$lists.ToArray()
}|ConvertTo-Json -Depth 35|Set-Content -LiteralPath $OutputPath -Encoding utf8
"audio conferencing prompt models: $($models.Count); constructors: $($constructors.Count); lists: $($lists.Count)"
