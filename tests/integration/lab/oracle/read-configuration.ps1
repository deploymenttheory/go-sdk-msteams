# Opt-in, read-only oracle for the generic configuration route. The tenant
# session does not export Get-CsConfiguration, so it runs the module's own
# internal command in that module's scope. It reports the instance count of
# one type, the selected instance's identity and one setting, and the error
# code for a missing instance.
$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue'
$InformationPreference='SilentlyContinue'
$WarningPreference='SilentlyContinue'
$VerbosePreference='SilentlyContinue'
$connected=$false
$certificate=$null
$report=@{success=$false}
# The generated cmdlets print correlation IDs to the console on errors; keep
# stdout for the JSON report only.
$stdout=[Console]::Out
[Console]::SetOut([IO.TextWriter]::Null)
try{
    if($env:MSTEAMS_LAB_READ_ONLY -ne '1' -or $env:MSTEAMS_LAB_POWERSHELL_ORACLE -ne '1'){throw 'Explicit read oracle opt-in required.'}
    $type=$env:MSTEAMS_LAB_CONFIGURATION_TYPE
    $name=$env:MSTEAMS_LAB_CONFIGURATION_NAME
    if($type -ne 'TeamsFeedbackPolicy'){throw 'Unsupported configuration type'}
    if($name -ne 'Global' -and $name -notmatch '^msteams-sdk-[a-z0-9-]+$'){throw 'Global or owned instance required'}
    $module=$env:MSTEAMS_LAB_MODULE_PATH
    if((Import-PowerShellDataFile -LiteralPath (Join-Path $module 'MicrosoftTeams.psd1')).ModuleVersion -ne '8.1.0'){throw 'Wrong module baseline'}
    $assembly=Join-Path $module 'bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll'
    if((Get-FileHash -LiteralPath $assembly -Algorithm SHA256).Hash.ToLowerInvariant() -ne '8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'){throw 'Wrong module assembly'}
    $identity=Get-Content -LiteralPath $env:MSTEAMS_LAB_IDENTITY_FILE -Raw | ConvertFrom-Json
    Import-Module (Join-Path $module 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
    $tenant=[Guid]::Parse($identity.tenant_id).ToString()
 if($env:MSTEAMS_LAB_AUTH_METHOD -eq 'client-secret'){
  $tokens=@()
  foreach($scope in @('https://graph.microsoft.com/.default','48ac35b8-9aa8-4d74-927d-1f4a14a0b239/.default')){
   $token=Invoke-RestMethod -Method Post -Uri "https://login.microsoftonline.com/$tenant/oauth2/v2.0/token" -Body @{grant_type='client_credentials';client_id=$identity.client_id;client_secret=$identity.client_secret;scope=$scope}
   $tokens+= $token.access_token
  }
  Connect-MicrosoftTeams -AccessTokens $tokens -ErrorAction Stop -InformationAction SilentlyContinue -WarningAction SilentlyContinue | Out-Null
 }elseif([string]::IsNullOrEmpty($env:MSTEAMS_LAB_AUTH_METHOD) -or $env:MSTEAMS_LAB_AUTH_METHOD -eq 'certificate'){
  $certificate=[Security.Cryptography.X509Certificates.X509Certificate2]::CreateFromPemFile($identity.certificate_file,$identity.private_key_file)
  Connect-MicrosoftTeams -Certificate $certificate -ApplicationId $identity.client_id -TenantId $tenant -ErrorAction Stop -InformationAction SilentlyContinue -WarningAction SilentlyContinue | Out-Null
 }else{throw 'Unknown auth method.'}
    $connected=$true
    $internal=Get-Module -All Microsoft.Teams.ConfigAPI.Cmdlets.internal | Select-Object -First 1
    if($null -eq $internal){throw 'Internal module unavailable'}
    $listed=@(& $internal { param($t) Get-CsConfiguration -ConfigType $t -ErrorAction Stop } $type)
    $value=(& $internal { param($t,$n) Get-CsConfiguration -ConfigType $t -ConfigName $n -ErrorAction Stop } $type $name) | ConvertFrom-Json
    $missing=''
    try{ & $internal { param($t) Get-CsConfiguration -ConfigType $t -ConfigName 'msteams-sdk-missing' -ErrorAction Stop } $type | Out-Null }catch{ $missing=($_.FullyQualifiedErrorId -split ',')[0] }
    $report=@{success=$true;list_count=$listed.Count;identity=[string]$value.Identity;user_initiated_mode=[string]$value.UserInitiatedMode;missing_code=$missing}
}catch{
    $report=@{success=$false}
}finally{
    [Console]::SetOut($stdout)
    if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null}
    if($null -ne $certificate){$certificate.Dispose()}
}
[Console]::Out.Write(($report|ConvertTo-Json -Depth 8 -Compress))
