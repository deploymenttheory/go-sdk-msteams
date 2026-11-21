[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblies=@(
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'},
 @{path='netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll';sha256='477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af'}
)
foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned banner text dependency.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
$env:MSTeamsContextInternal=$null
Import-Module $manifest -Force
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@((Join-Path $PSHOME 'System.Management.Automation.dll'))+@($assemblies|ForEach-Object{Join-Path $ModulePath $_.path})+@((Join-Path $PSHOME 'Newtonsoft.Json.dll'))
Add-Type -Path @((Join-Path $PSScriptRoot 'oracle/PolicyProviderCapture.cs'),(Join-Path $PSScriptRoot 'oracle/BannerTextBindingCapture.cs')) -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$cases=[Collections.Generic.List[object]]::new()
$id='01234567-89ab-cdef-0123-456789abcdef'
foreach($c in @(
 @{verb='New';name='identity-text';arguments=@{Identity="Global/$id";Text='fixture'}},
 @{verb='New';name='relative-text';arguments=@{Id=$id;Text='fixture'}},
 @{verb='New';name='default-id';arguments=@{Text='fixture'}},
 @{verb='New';name='empty-text';arguments=@{Id=$id;Text=''}},
 @{verb='New';name='null-text';arguments=@{Id=$id;Text=$null}},
 @{verb='New';name='null-description';arguments=@{Id=$id;Text='fixture';Description=$null}},
 @{verb='New';name='empty-identity';arguments=@{Identity='';Text='fixture'}},
 @{verb='New';name='opaque-relative';arguments=@{Id='opaque';Text='fixture'}},
 @{verb='New';name='force-false';arguments=@{Id=$id;Text='fixture';Force=$false}},
 @{verb='Set';name='identity-text';arguments=@{Identity="Global/$id";Text='fixture'}},
 @{verb='Set';name='relative-text';arguments=@{Id=$id;Text='fixture'}},
 @{verb='Set';name='null-text';arguments=@{Id=$id;Text=$null}},
 @{verb='Set';name='empty-text';arguments=@{Id=$id;Text=''}},
 @{verb='Set';name='null-description';arguments=@{Id=$id;Description=$null}},
 @{verb='Set';name='opaque-relative';arguments=@{Id='opaque';Text='fixture'}},
 @{verb='Set';name='mixed-selectors';arguments=@{Identity="Global/$id";Id=$id;Text='fixture'}},
 @{verb='Get';name='list';arguments=@{}},
 @{verb='Get';name='filter';arguments=@{Filter='Fixture*'}},
 @{verb='Get';name='relative';arguments=@{Id=$id}},
 @{verb='Get';name='empty-identity';arguments=@{Identity=''}},
 @{verb='Get';name='opaque-relative';arguments=@{Id='opaque'}},
 @{verb='Remove';name='identity';arguments=@{Identity="Global/$id"}},
 @{verb='Remove';name='relative';arguments=@{Id=$id}},
 @{verb='Remove';name='force-false';arguments=@{Id=$id;Force=$false}}
)){
 $argsJson=$c.arguments|ConvertTo-Json -Compress
 $observed=[TeamsContractOracle.BannerBindingOracle]::Run($c.verb,$argsJson)|ConvertFrom-Json -AsHashtable
 if($null -ne $observed.output){$observed.output=$observed.output|ConvertFrom-Json -AsHashtable}
 $cases.Add(@{command=($c.verb+'-CsTeamsCustomBannerText');name=$c.name;arguments=$c.arguments;observed=$observed})
}
$filters=@('','Fixture*',' ','Global','Global/*')|ForEach-Object{@{pattern=$_;observed=([TeamsContractOracle.BannerBindingOracle]::Filter($_)|ConvertFrom-Json -AsHashtable)}}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='inherited-compiled-banner-binding-and-pure-filter';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString()
 stage='Owned subclasses inherit original compiled properties/attributes/constructors. Synthetic Begin/Process/End replace source lifecycle solely to return actual compiled binder values. In-process runspace/fixture commands, memory client and credential-refusing host. Actual protected FilterPolicies invoked separately. No original lifecycle, scope authorization, token acquisition, discovery or network. Source preparation/provider are separate fixtures.'
 cases=$cases.ToArray();filters=$filters
}|ConvertTo-Json -Depth 30|Set-Content -LiteralPath $OutputPath -Encoding utf8
"banner inherited binder cases: $($cases.Count); filters: $($filters.Count)"
