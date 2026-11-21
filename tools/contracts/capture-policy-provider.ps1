# Offline pure helper/model oracle; never processes a management command.
[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$ModulePath,
    [Parameter(Mandatory)][string]$OutputPath
)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblyPath='bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll'
$expected='8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'
$assemblyFile=Join-Path $ModulePath $assemblyPath
if((Get-FileHash -LiteralPath $assemblyFile -Algorithm SHA256).Hash.ToLowerInvariant() -ne $expected){throw 'Pinned assembly differs.'}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
Import-Module $manifest -Force
$otherAssemblies=@(
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'},
 @{path='netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll';sha256='477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af'}
)
foreach($item in $otherAssemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $item.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $item.sha256){throw 'Pinned provider assembly differs.'}}
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@((Join-Path $PSHOME 'System.Management.Automation.dll'))+@($otherAssemblies | ForEach-Object {Join-Path $ModulePath $_.path})
Add-Type -Path (Join-Path $PSScriptRoot 'oracle/PolicyProviderCapture.cs') -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$cases=[Collections.Generic.List[object]]::new()
foreach($operation in @('read','list','create','update','delete','user','tenant','group','acs-read','acs-list','acs-update','external-read','external-list','external-update')) {
    $status=200;$body='{"Identity":"fixture-response"}'
    if($operation -eq 'list' -or $operation.EndsWith('-list')){$body='[]'}
    if($operation -in @('update','delete','acs-update','external-update')){$status=204;$body=''}
    $value=[TeamsContractOracle.ProviderOracle]::Run($operation,$status,$body,$false)
    $cases.Add(@{name=$operation;status=$status;response=$body;capture=$value})
}
foreach($c in @(
 @{name='delete-force-false';op='delete';status=204;body='';force=$true},
 @{name='read-no-content';op='read';status=204;body='';force=$false},
 @{name='group-accepted-opaque';op='group';status=202;body='{"OperationId":"fixture-operation"}';force=$false},
 @{name='list-envelope';op='list';status=200;body='{"value":[]}';force=$false},
 @{name='read-tagged';op='read-tagged';status=200;body='{"Identity":"tag:Fixture / Policy"}';force=$false},
 @{name='read-future-enum';op='read';status=200;body='{"Identity":"Global","CommunicationWithExternalOrgs":7}';force=$false},
 @{name='read-null-domain-list';op='read';status=200;body='{"Identity":"Global","AllowedExternalDomains":null}';force=$false},
 @{name='acs-null-list';op='acs-read';status=200;body='{"Identity":"Global","AllowedAcsResources":null}';force=$false},
 @{name='external-null-list';op='external-read';status=200;body='{"Identity":"Global","BlockedUsers":null}';force=$false},
 @{name='delete-denied';op='delete';status=403;body='{"Message":"fixture-denied"}';force=$false}
)) {
    $value=[TeamsContractOracle.ProviderOracle]::Run($c.op,$c.status,$c.body,$c.force)
    $cases.Add(@{name=$c.name;status=$c.status;response=$c.body;capture=$value})
}
$lists=[Collections.Generic.List[object]]::new()
foreach($name in @('null','empty-operators','replacement','remove-one','remove-missing','add-existing','remove-and-add')) {
    $value=[TeamsContractOracle.ProviderOracle]::ApplyListCase($name) | ConvertFrom-Json -AsHashtable
    $lists.Add(@{name=$name;observation=$value})
}
$artifact=[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-compiled-provider-request-and-dispatch';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 host_powershell_version=$PSVersionTable.PSVersion.ToString();sma_assembly_version=[Management.Automation.PSListModifier].Assembly.GetName().Version.ToString()
 assemblies=$otherAssemblies
 stage='Direct HttpRequestHelper methods; synthetic InvocationInfo/bound parameters; fake host refuses tokens; cached module version and private HTTP client replaced/restored; only in-memory handler. MPACmdlet default header is inserted by adapter from factory source. No command processing, connection, discovery or network.'
 cases=$cases.ToArray();lists=$lists.ToArray()
}
$artifact | ConvertTo-Json -Depth 30 | Set-Content -LiteralPath $OutputPath -Encoding utf8
"provider request observations: $($cases.Count)"
