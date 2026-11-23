# Opt-in, read-only oracle for tenant identity reads. It reports the tenant ID,
# collection counts, the owned acceptance group search result and the
# deprecated/forbidden outcomes only; no tenant attribute values are retained.
$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue'
$InformationPreference='SilentlyContinue'
$WarningPreference='SilentlyContinue'
$VerbosePreference='SilentlyContinue'
$connected=$false
$certificate=$null
$report=@{success=$false}
try{
    if($env:MSTEAMS_LAB_READ_ONLY -ne '1' -or $env:MSTEAMS_LAB_POWERSHELL_ORACLE -ne '1'){throw 'Explicit read oracle opt-in required.'}
    $query=$env:MSTEAMS_LAB_GROUP_SEARCH
    if($query -cne 'go-sdk-msteams acceptance policy group'){throw 'Owned group search required'}
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
    $value=Get-CsTenant -ErrorAction Stop
    $groups=@(Find-CsGroup -SearchQuery $query -ExactMatchOnly $true -MaxResults 10 -ErrorAction Stop)
    $deprecated=$false
    try{Get-CsOnlineDirectoryTenant -ErrorAction Stop | Out-Null}catch{
        $deprecated=$_.Exception.Message -ceq 'This cmdlet is no longer supported, please consult public documentation or use Get-CsTenant and Get-CsOnlineDialinConferencingBridge to get the relevant attributes.'
    }
    # The failing cmdlet writes its correlation ID straight to the console;
    # keep stdout to the single JSON report.
    $forbidden=$false
    $console=[Console]::Out
    try{
        [Console]::SetOut([IO.TextWriter]::Null)
        Get-CsTenantLicensingConfiguration -ErrorAction Stop | Out-Null
    }catch{
        $forbidden=$_.FullyQualifiedErrorId -like 'Forbidden,*'
    }finally{
        [Console]::SetOut($console)
    }
    $report=@{success=$true;tenant_id=[string]$value.TenantId;verified_domain_count=@($value.VerifiedDomains).Count;sip_domain_count=@($value.SipDomain).Count;assigned_plan_count=@($value.AssignedPlan).Count;group_ids=@($groups|ForEach-Object{[string]$_.Id});directory_tenant_deprecated=$deprecated;licensing_forbidden=$forbidden}
}catch{
    $report=@{success=$false}
}finally{
    if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null}
    if($null -ne $certificate){$certificate.Dispose()}
}
[Console]::Out.Write(($report|ConvertTo-Json -Depth 8 -Compress))
