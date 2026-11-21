# Offline compiled request helper only; no connection, cmdlet processing or tenant.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblies=@(
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'},
 @{path='netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll';sha256='477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af'}
)
foreach($item in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $item.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $item.sha256){throw 'Pinned GPA assembly differs.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
Import-Module $manifest -Force
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@((Join-Path $PSHOME 'System.Management.Automation.dll'))+@($assemblies | ForEach-Object {Join-Path $ModulePath $_.path})+@([Newtonsoft.Json.JsonConvert].Assembly.Location)
Add-Type -Path @((Join-Path $PSScriptRoot 'oracle/PolicyProviderCapture.cs'),(Join-Path $PSScriptRoot 'oracle/AssignmentMigrationCapture.cs')) -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$cases=[Collections.Generic.List[object]]::new()
foreach($operation in @('list','read','types','names','conflicts','create','clear')) {
 $body=switch($operation){'list'{'[]'} 'types'{'[{"policyType":"FixturePolicy","daCount":0}]'} 'names'{'[{"policyName":"Fixture","daCount":0}]'} 'conflicts'{'{"totalCountUsersWithConflict":0}'} default{'{"migrationEventId":"00000000-0000-4000-8000-000000000002","state":"NotStarted","directAssignmentsProcessed":0}'}}
 $value=[TeamsContractOracle.AssignmentMigrationOracle]::Run($operation,200,$body,$false)
 $cases.Add(@{name=$operation;operation=$operation;status=200;response=$body;allow=$false;capture=$value})
}
foreach($c in @(
 @{name='create-allow';op='create';status=202;body='{"state":"PopulationInProgress","Future":false}';allow=$true},
 @{name='clear-allow';op='clear';status=202;body='{"state":42}';allow=$true},
 @{name='read-denied';op='read';status=403;body='{"Message":"fixture-denied"}';allow=$false},
 @{name='create-denied';op='create';status=403;body='{"Message":"fixture-denied"}';allow=$false},
 @{name='read-malformed';op='read';status=200;body='not-json';allow=$false},
 @{name='read-empty';op='read';status=204;body='';allow=$false},
 @{name='list-null';op='list';status=200;body='null';allow=$false},
 @{name='list-envelope';op='list';status=200;body='{"value":[]}';allow=$false}
)) {
 $value=[TeamsContractOracle.AssignmentMigrationOracle]::Run($c.op,$c.status,$c.body,$c.allow)
 $cases.Add(@{name=$c.name;operation=$c.op;status=$c.status;response=$c.body;allow=$c.allow;capture=$value})
}
if(@($cases | Where-Object {$_.capture.Sends -ne 1 -or $_.capture.TokenRequests -ne 0 -or $_.capture.FailureType}).Count){throw 'Offline GPA capture failed.'}
$artifact=[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-compiled-gpa-request-helper';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString();sma_assembly_version=[Management.Automation.InvocationInfo].Assembly.GetName().Version.ToString()
 stage='Source-derived safe suffixes passed to CommonHttpHelperRequest.SendGPARequests; Newtonsoft request DTO; synthetic InvocationInfo; fake host refuses credentials; cached module version and HTTP client replaced/restored; memory-only sender. MPACmdlet default inserted from factory source. No command ProcessRecord, response DTO decode, discovery or network.'
 cases=$cases.ToArray()
}
$artifact | ConvertTo-Json -Depth 30 | Set-Content -LiteralPath $OutputPath -Encoding utf8
"GPA request observations: $($cases.Count)"
