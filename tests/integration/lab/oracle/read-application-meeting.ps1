# Optional read-only 8.1.0 oracle for Global application meeting configuration.
$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue'
$InformationPreference='SilentlyContinue'
$WarningPreference='SilentlyContinue'
$VerbosePreference='SilentlyContinue'
$DebugPreference='SilentlyContinue'
$report=@{success=$false}
$connected=$false
$certificate=$null
try {
    if($env:MSTEAMS_LAB_READ_ONLY -ne '1' -or $env:MSTEAMS_LAB_POWERSHELL_ORACLE -ne '1'){throw 'Explicit read oracle opt-in required.'}
    $module=$env:MSTEAMS_LAB_MODULE_PATH
    $manifest=Join-Path $module 'MicrosoftTeams.psd1'
    if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
    $assembly=Join-Path $module 'bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll'
    if((Get-FileHash -LiteralPath $assembly -Algorithm SHA256).Hash.ToLowerInvariant() -ne '8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'){throw 'Wrong baseline assembly.'}
    Import-Module $manifest -Force -WarningAction SilentlyContinue
    $identity=Get-Content -LiteralPath $env:MSTEAMS_LAB_IDENTITY_FILE -Raw | ConvertFrom-Json
    $certificate=[Security.Cryptography.X509Certificates.X509Certificate2]::CreateFromPemFile($identity.certificate_file,$identity.private_key_file)
    Connect-MicrosoftTeams -Certificate $certificate -ApplicationId $identity.client_id -TenantId $identity.tenant_id -WarningAction SilentlyContinue -InformationAction SilentlyContinue -ErrorAction Stop | Out-Null
    $connected=$true
    $values=@(Get-CsApplicationMeetingConfiguration -Identity Global -ErrorAction Stop)
    if($values.Count -ne 1 -or [string]$values[0].Identity -ine 'Global'){throw 'Expected Global configuration.'}
    $property=$values[0].PSObject.Properties['AllowRemoveParticipantAppIds']
    if($null -eq $property){throw 'Incomplete configuration projection.'}
    $ids=$null
    if($null -ne $property.Value){$ids=@($property.Value)}
    $report=@{success=$true;identity=[string]$values[0].Identity;app_ids=$ids}
} catch {
    $report=@{success=$false}
} finally {
    if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null}
    if($null -ne $certificate){$certificate.Dispose()}
}
[Console]::Out.Write(($report | ConvertTo-Json -Depth 8 -Compress))
