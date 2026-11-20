# Opt-in live PowerShell oracle. This script only connects and reads one owned fixture.
$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue'
$InformationPreference='SilentlyContinue'
$WarningPreference='SilentlyContinue'
$VerbosePreference='SilentlyContinue'
$report=@{success=$false}
$connected=$false
$certificate=$null
try{
    $fixture=$env:MSTEAMS_LAB_POLICY_FIXTURE
    if($fixture -notmatch '^msteams-sdk-[a-z2-7]{26,64}-app-access$'){throw 'Owned fixture required'}
    $module=$env:MSTEAMS_LAB_MODULE_PATH
    if((Import-PowerShellDataFile -LiteralPath (Join-Path $module 'MicrosoftTeams.psd1')).ModuleVersion -ne '8.1.0'){throw 'Wrong module baseline'}
    $assembly=Join-Path $module 'bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll'
    if((Get-FileHash -LiteralPath $assembly -Algorithm SHA256).Hash.ToLowerInvariant() -ne '8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'){throw 'Wrong module assembly'}
    $identity=Get-Content -LiteralPath $env:MSTEAMS_LAB_IDENTITY_FILE -Raw | ConvertFrom-Json
    Import-Module (Join-Path $module 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
    $certificate=[Security.Cryptography.X509Certificates.X509Certificate2]::CreateFromPemFile($identity.certificate_file,$identity.private_key_file)
    Connect-MicrosoftTeams -Certificate $certificate -ApplicationId $identity.client_id -TenantId $identity.tenant_id -WarningAction SilentlyContinue -InformationAction SilentlyContinue -ErrorAction Stop | Out-Null
    $connected=$true
    $policies=@(Get-CsApplicationAccessPolicy -Identity $fixture -ErrorAction Stop)
    if($policies.Count -ne 1){throw 'Expected one owned policy'}
    $report=@{success=$true;identity=[string]$policies[0].Identity;app_ids=@($policies[0].AppIds);description=[string]$policies[0].Description}
}catch{
    $report=@{success=$false}
}finally{
    if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null}
    if($null -ne $certificate){$certificate.Dispose()}
}
[Console]::Out.Write(($report|ConvertTo-Json -Depth 8 -Compress))
