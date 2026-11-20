# Opt-in, read-only oracle for one owned tenant template.
$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue'
$InformationPreference='SilentlyContinue'
$WarningPreference='SilentlyContinue'
$VerbosePreference='SilentlyContinue'
$connected=$false
$certificate=$null
$report=@{success=$false}
try{
    $fixture=$env:MSTEAMS_LAB_TEMPLATE_FIXTURE
    $reference=$env:MSTEAMS_LAB_TEMPLATE_ID
    if($fixture -notmatch '^msteams-sdk-[a-z2-7]{26,64}-template$'){throw 'Owned fixture required'}
    if($reference -notmatch '^/api/teamtemplates/v1\.0/[^/]+/Tenant/en-US$'){throw 'Owned tenant reference required'}
    $module=$env:MSTEAMS_LAB_MODULE_PATH
    if((Import-PowerShellDataFile -LiteralPath (Join-Path $module 'MicrosoftTeams.psd1')).ModuleVersion -ne '8.1.0'){throw 'Wrong module baseline'}
    $assembly=Join-Path $module 'bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll'
    if((Get-FileHash -LiteralPath $assembly -Algorithm SHA256).Hash.ToLowerInvariant() -ne '8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'){throw 'Wrong module assembly'}
    $identity=Get-Content -LiteralPath $env:MSTEAMS_LAB_IDENTITY_FILE -Raw | ConvertFrom-Json
    Import-Module (Join-Path $module 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
    $certificate=[Security.Cryptography.X509Certificates.X509Certificate2]::CreateFromPemFile($identity.certificate_file,$identity.private_key_file)
    Connect-MicrosoftTeams -Certificate $certificate -ApplicationId $identity.client_id -TenantId $identity.tenant_id -WarningAction SilentlyContinue -InformationAction SilentlyContinue -ErrorAction Stop | Out-Null
    $connected=$true
    $values=@(Get-CsTeamTemplate -OdataId $reference -ErrorAction Stop)
    if($values.Count -ne 1 -or $values[0].DisplayName -cne $fixture){throw 'Expected one owned template'}
    $report=@{success=$true;display_name=[string]$values[0].DisplayName;short_description=[string]$values[0].ShortDescription;description=[string]$values[0].Description;channel_names=@($values[0].Channel|Where-Object{$null -ne $_}|ForEach-Object{[string]$_.DisplayName})}
}catch{
    $report=@{success=$false}
}finally{
    if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null}
    if($null -ne $certificate){$certificate.Dispose()}
}
[Console]::Out.Write(($report|ConvertTo-Json -Depth 8 -Compress))
