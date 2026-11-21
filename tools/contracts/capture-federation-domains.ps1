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
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@($assemblyFile,(Join-Path $PSHOME 'System.Management.Automation.dll'))
# The pinned assembly targets SMA 3; this offline host supplies SMA 7. Suppress
# only that known compile-time reference-unification warning.
Add-Type -Path (Join-Path $PSScriptRoot 'oracle/FederationDomainCapture.cs') -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$cases=[Collections.Generic.List[object]]::new()
$pattern=[Microsoft.Teams.ConfigApi.Cmdlets.NewEdgeDomainPatternHelper]::ConvertToDomainPattern('*.example.invalid')
$cases.Add(@{name='domain-pattern';wire=@{Domain=[string]$pattern.PSObject.Properties['Domain'].Value}})
foreach($c in @(
    @{name='allow-list';all=$false;domains=@('b.example.invalid','a.example.invalid','b.example.invalid','')},
    @{name='empty-allow-list';all=$false;domains=@()},
    @{name='allow-all';all=$true;domains=@()}
)) {
    $wire=[TeamsContractOracle.FederationDomains]::Normalize($c.all,[string[]]$c.domains) | ConvertFrom-Json -AsHashtable
    $cases.Add(@{name=$c.name;wire=$wire.AllowedDomains})
}
$artifact=[ordered]@{
    schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
    qualification='offline-federation-domain-normalization';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
    assembly=@{path=$assemblyPath;sha256=$expected}
    stage='ConvertToDomainPattern, synthetic PSObject choice inputs, SetCsTenantFederationConfiguration.ValidateAndUpdatePropertyBagParams, XdsConfiguration serialization. No cmdlet processing or network.'
    cases=$cases.ToArray()
}
$artifact | ConvertTo-Json -Depth 16 | Set-Content -LiteralPath $OutputPath -Encoding utf8
"federation domain observations: $($cases.Count)"
