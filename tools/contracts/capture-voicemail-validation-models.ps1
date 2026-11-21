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
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@($assemblies|ForEach-Object{Join-Path $ModulePath $_.path})+@([Newtonsoft.Json.JsonConvert].Assembly.Location)
Add-Type -Path (Join-Path $PSScriptRoot 'oracle/VoicemailValidationModelCapture.cs') -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$cases=[Collections.Generic.List[object]]::new()
foreach($v in @(
 @{name='null';uri='';is_null=$true},
 @{name='https';uri='https://example.invalid/validate';is_null=$false},
 @{name='uppercase';uri='HTTPS://example.invalid/validate';is_null=$false},
 @{name='space';uri='https:// ';is_null=$false},
 @{name='over-1024';uri=('https://example.invalid/'+('a'*2048));is_null=$false},
 @{name='final-newline';uri="https://example.invalid/validate`n";is_null=$false},
 @{name='http';uri='http://example.invalid/validate';is_null=$false},
 @{name='relative';uri='validate';is_null=$false},
 @{name='empty';uri='';is_null=$false},
 @{name='empty-https';uri='https://';is_null=$false},
 @{name='interior-newline';uri="https://example.invalid/val`nidate";is_null=$false}
)) {
 $rejected=$false;$wire=$null
 try{$wire=[TeamsContractOracle.VoicemailValidationModelOracle]::Run($v.uri,$v.is_null)|ConvertFrom-Json -AsHashtable}catch{$rejected=$true}
 $cases.Add(@{name=$v.name;uri=$v.uri;is_null=$v.is_null;rejected=$rejected;wire=$wire})
}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-voicemail-validation-setters';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString()
 stage='Pinned model setters and Newtonsoft serialization, with true null supplied in C# to separate PowerShell property string coercion. No command, credential access, request helper, URI dereference or network.'
 cases=$cases.ToArray()
}|ConvertTo-Json -Depth 20|Set-Content -LiteralPath $OutputPath -Encoding utf8
"voicemail validation setter observations: $($cases.Count)"
