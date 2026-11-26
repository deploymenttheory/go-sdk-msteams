[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference='Stop'
Import-Module (Join-Path $ModulePath 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
$type=[AppDomain]::CurrentDomain.GetAssemblies() | ForEach-Object { $_.GetType('Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.XdsConfiguration',$false) } | Where-Object { $_ } | Select-Object -First 1
$assemblyHash=(Get-FileHash $type.Assembly.Location).Hash.ToLowerInvariant()
if ($assemblyHash -ne '8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6') { throw 'Pinned generated assembly differs' }
$parse=$type.GetMethod('FromJsonString',[Reflection.BindingFlags]'Public,Static')
$before=$type.GetProperty('BeforeJson')
$inputEvidence=Join-Path $PSScriptRoot '../../docs/implementation/policies/evidence/pol91-broadcast-modern-read.json'
$inputs=@((Get-Content $inputEvidence -Raw | ConvertFrom-Json).observations.input_json | Select-Object -Unique)
$inputs+=@('', ' ', '{"Future":{"x":1,"x":2}}', '{"Future":{"X":1,"x":2}}')
$inputs+=@('{"Future":{"K":1,"\u212a":2}}', '{"Future":{"S":1,"\u017f":2}}',
    '{"Future":{"I":1,"\u0131":2}}', '{"Future":{"i":1,"\u0130":2}}',
    '{"Future":{"\u03a3":1,"\u03c2":2}}', '{"Future":{"\u00e9":1,"\u00c9":2}}',
    '{"SdnApiTo\u212aen":"foreign-key"}')
$rows=foreach($raw in $inputs) {
    $failure=$null; $model=$null; $original=$null; $identity=$null; $conversionFailure=$null
    try {
        $model=$parse.Invoke($null,[object[]]@([string]$raw))
        if ($null -ne $model) { $original=$before.GetValue($model); $identity=$type.GetProperty('Identity').GetValue($model) }
    } catch { $failure=$_.Exception.GetBaseException().GetType().FullName }
    if ($null -ne $original) {
        try { $null=ConvertFrom-Json -InputObject $original }
        catch { $conversionFailure=$_.Exception.GetBaseException().GetType().FullName }
    }
    [pscustomobject]@{input_json=$raw;model_present=($null -ne $model);before_json=$original;typed_identity=$identity;error=$failure;conversion_error=$conversionFailure}
}
[pscustomobject]@{
    scope='Actual generated XdsConfiguration.FromJsonString and BeforeJson/Identity data getters, followed by the pure ConvertFrom-Json conversion only. No response callback, cmdlet lifecycle, flight, legacy, authentication or HTTP.'
    assembly_sha256=$assemblyHash
    powershell_version=$PSVersionTable.PSVersion.ToString()
    reproducer_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant()
    input_evidence_sha256=(Get-FileHash $inputEvidence).Hash.ToLowerInvariant()
    observations=@($rows)
    management_lifecycle_invoked=$false
    authentication_or_http=$false
} | ConvertTo-Json -Depth 8
