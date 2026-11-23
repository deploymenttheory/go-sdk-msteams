# Opt-in oracle for the cache clear denial. Clear-CsCacheOperation is not
# exported in a tenant session, so it runs in the ConfigAPI module's scope.
# It sends only a nonexistent key for an invalid region, which cannot clear any
# cache entry, and reports the error code the module raises.
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
    $key=$env:MSTEAMS_LAB_CACHE_PROBE_KEY
    $region=$env:MSTEAMS_LAB_CACHE_PROBE_REGION
    if($key -notmatch '^msteams-sdk-[a-z0-9-]+-nonexistent-key$' -or $region -notmatch '^msteams-sdk-[a-z0-9-]+-invalid-region$'){throw 'Inert probe inputs required'}
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
    $configapi=Get-Module -All | Where-Object { $_.Name -eq 'Microsoft.Teams.ConfigAPI.Cmdlets' -and $_.ModuleType -eq 'Script' } | Select-Object -First 1
    if($null -eq $configapi){throw 'ConfigAPI module unavailable'}
    $code=''
    try{ & $configapi { param($k,$r) Clear-CsCacheOperation -KeysToDelete $k -Region $r -Confirm:$false -ErrorAction Stop } $key $region | Out-Null; $code='accepted' }catch{ $code=($_.FullyQualifiedErrorId -split ',')[0] }
    $report=@{success=$true;error_code=$code}
}catch{
    $report=@{success=$false}
}finally{
    [Console]::SetOut($stdout)
    if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null}
    if($null -ne $certificate){$certificate.Dispose()}
}
[Console]::Out.Write(($report|ConvertTo-Json -Depth 8 -Compress))
