# Pure model constructors/setters only. No cmdlet, request helper or credentials.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblies=@(
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'}
)
foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned assembly.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
Import-Module $manifest -Force
$cases=[Collections.Generic.List[object]]::new()
$vectors=@(
 @{name='https';uri='https://example.invalid/image.png'},
 @{name='uppercase';uri='HTTPS://example.invalid/image.png'},
 @{name='http';uri='http://example.invalid/image.png'},
 @{name='relative';uri='image.png'},
 @{name='empty';uri=''},
 @{name='empty-https';uri='https://'},
 @{name='space';uri='https:// '},
 @{name='over-1024';uri=('https://example.invalid/'+('a'*2048))},
 @{name='final-newline';uri="https://example.invalid/image.png`n"},
 @{name='interior-newline';uri="https://example.invalid/image`n.png"}
)
foreach($kind in @('image','slate','reaction')) {
 foreach($v in $vectors) {
  $wire=$null;$rejected=$false
  try {
   $value=switch($kind){
    'image' {[Microsoft.Teams.Policy.Administration.Cmdlets.Core.MeetingBackgroundImage]::new('fixture-image',[long]-1,'Fixture',$v.uri)}
    'slate' {[Microsoft.Teams.Policy.Administration.Cmdlets.Core.NdiAssuranceSlate]::new('fixture-slate','Fixture',$v.uri)}
    'reaction' {[Microsoft.Teams.Policy.Administration.Cmdlets.Core.MeetingReactionItemModel]::new('fixture-reaction',$v.uri,'FutureReaction')}
   }
   $wire=[Newtonsoft.Json.JsonConvert]::SerializeObject($value.PSObject.BaseObject) | ConvertFrom-Json -AsHashtable
  } catch {$rejected=$true}
  $cases.Add(@{name=($kind+'-'+$v.name);kind=$kind;uri=$v.uri;rejected=$rejected;wire=$wire})
 }
}
$theme=[Microsoft.Teams.Policy.Administration.Cmdlets.Core.MeetingBrandingTheme]::new('fixture-theme')
$theme.LogoImageLightUri='relative-or-future';$theme.BrandAccentColor='future-color'
$theme.ChatBackgroundImageAlignment='FutureAlignment';$theme.EnableChatBackground=$false
$theme.Enabled=$false;$theme.DisableOriginalReactions=$false
$wire=[Newtonsoft.Json.JsonConvert]::SerializeObject($theme.PSObject.BaseObject) | ConvertFrom-Json -AsHashtable
$cases.Add(@{name='theme-open-fields';kind='theme';rejected=$false;wire=$wire})
$artifact=[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-branding-model-constructors-and-setters';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString()
 stage='Pure pinned model constructors/setters and Newtonsoft serialization. Synthetic strings. No cmdlet binding/CreateNew/UpdateProperties/ProcessRecord, credential access, request helper, URI dereference or network.'
 cases=$cases.ToArray()
}
$artifact | ConvertTo-Json -Depth 20 | Set-Content -LiteralPath $OutputPath -Encoding utf8
"branding model observations: $($cases.Count)"
