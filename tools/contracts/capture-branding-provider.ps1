# Offline request helpers and local SMA modifiers only, not cmdlet execution.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblies=@(
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'},
 @{path='netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll';sha256='477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af'}
)
foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned assembly.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
Import-Module $manifest -Force
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@((Join-Path $PSHOME 'System.Management.Automation.dll'))+@($assemblies | ForEach-Object {Join-Path $ModulePath $_.path})+@([Newtonsoft.Json.JsonConvert].Assembly.Location)
Add-Type -Path @((Join-Path $PSScriptRoot 'oracle/PolicyProviderCapture.cs'),(Join-Path $PSScriptRoot 'oracle/BrandingProviderCapture.cs')) -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$cases=[Collections.Generic.List[object]]::new()
foreach($op in @('read','list','create','patch','delete','user','tenant','group')) {
 $status=200;$reply='{"Identity":"fixture-response"}'
 if($op -eq 'list'){$reply='[]'}
 if($op -in @('patch','delete')){$status=204;$reply=''}
 $capture=[TeamsContractOracle.BrandingOracle]::Run($op,$status,$reply,$false,$false)
 $cases.Add(@{name=$op;operation=$op;status=$status;response=$reply;force=$false;nested=$false;capture=$capture})
}
foreach($c in @(
 @{name='create-nested';op='create';status=201;reply='{"Identity":"Fixture","Future":false}';force=$false;nested=$true},
 @{name='patch-nested';op='patch';status=204;reply='';force=$false;nested=$true},
 @{name='delete-force-false';op='delete';status=204;reply='';force=$true;nested=$false},
 @{name='read-null-arrays';op='read';status=200;reply='{"Identity":"Global","MeetingBackgroundImages":null,"MeetingBrandingThemes":null,"NdiAssuranceSlateImages":null}';force=$false;nested=$false},
 @{name='read-invalid-asset';op='read';status=200;reply='{"Identity":"Global","MeetingBackgroundImages":[{"Id":"fixture","ImageUri":"http://example.invalid/a"}]}';force=$false;nested=$false},
 @{name='read-null-asset-id';op='read';status=200;reply='{"Identity":"Global","MeetingBackgroundImages":[{"Id":null}]}';force=$false;nested=$false},
 @{name='read-null-asset-order';op='read';status=200;reply='{"Identity":"Global","MeetingBackgroundImages":[{"Order":null}]}';force=$false;nested=$false},
 @{name='read-lowercase-invalid-uri';op='read';status=200;reply='{"Identity":"Global","MeetingBackgroundImages":[{"imageuri":"http://example.invalid/a"}]}';force=$false;nested=$false},
 @{name='read-lowercase-null-theme';op='read';status=200;reply='{"Identity":"Global","MeetingBrandingThemes":[{"identity":null}]}';force=$false;nested=$false},
 @{name='list-envelope';op='list';status=200;reply='{"value":[]}';force=$false;nested=$false},
 @{name='delete-denied';op='delete';status=403;reply='{"Message":"fixture-denied"}';force=$false;nested=$false},
 @{name='read-empty';op='read';status=204;reply='';force=$false;nested=$false},
 @{name='group-opaque';op='group';status=202;reply='{"OperationId":"fixture-operation"}';force=$false;nested=$false}
)) {
 $capture=[TeamsContractOracle.BrandingOracle]::Run($c.op,$c.status,$c.reply,$c.force,$c.nested)
 $cases.Add(@{name=$c.name;operation=$c.op;status=$c.status;response=$c.reply;force=$c.force;nested=$c.nested;capture=$capture})
}
$lists=[Collections.Generic.List[object]]::new()
foreach($kind in @('policy','reaction')) {
 $names=@('null','empty-replace','replace','empty-delta','add','remove')
 if($kind -eq 'policy'){$names+=@('remove-missing','null-item')}else{$names+='remove-different'}
 foreach($empty in @($false,$true)){foreach($name in $names){$observation=[TeamsContractOracle.BrandingOracle]::ApplyList($kind,$name,$empty) | ConvertFrom-Json -AsHashtable;$lists.Add(@{kind=$kind;name=$name;empty=$empty;observation=$observation})}}
}
$artifact=[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-branding-provider-and-local-modifier';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString();sma_assembly_version=[Management.Automation.PSListModifier].Assembly.GetName().Version.ToString()
 stage='Direct pinned HttpRequestHelper calls with synthetic InvocationInfo; fake host refuses tokens; private client/version/host replaced/restored; memory-only sender. MPACmdlet inserted from factory source. Local modifiers call pinned Core.PSListModifierHelper.ApplyTo with host SMA containers and source delegates. No cmdlet constructors/ProcessRecord, binding, auth, discovery or network.'
 cases=$cases.ToArray();lists=$lists.ToArray()
}
$artifact | ConvertTo-Json -Depth 30 | Set-Content -LiteralPath $OutputPath -Encoding utf8
"branding requests: $($cases.Count); list observations: $($lists.Count)"
