# Opt-in, read-only oracle for mainline attendant reads. Mode "tenant" reports
# the tenant information defaults and counts; mode "owned" reads the three owned
# lifecycle resources by identity and reports their selected fields.
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
    $mode=$env:MSTEAMS_LAB_MAINLINE_MODE
    if($mode -ne 'tenant' -and $mode -ne 'owned'){throw 'Oracle mode required'}
    $guid='^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
    if($mode -eq 'owned'){
        foreach($value in @($env:MSTEAMS_LAB_MAINLINE_SPAM_ID,$env:MSTEAMS_LAB_MAINLINE_BOOKING_ID,$env:MSTEAMS_LAB_MAINLINE_QA_ID)){
            if($value -notmatch $guid){throw 'Owned identities required'}
        }
    }
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
    if($mode -eq 'tenant'){
        $information=Get-CsMainlineAttendantTenantInformation -ErrorAction Stop
        $languages=@(Get-CsMainlineAttendantSupportedLanguages -ErrorAction Stop)
        $voices=@(Get-CsMainlineAttendantSupportedVoices -ErrorAction Stop)
        # The module projects the defaults as DefaultLanguageId, DefaultVoiceId and DefaultTimeZoneId.
        $report=@{success=$true;default_language=[string]$information.DefaultLanguageId;default_voice=[string]$information.DefaultVoiceId;
            default_time_zone=[string]$information.DefaultTimeZoneId;language_count=$languages.Count;voice_count=$voices.Count}
    }else{
        $spam=Get-CsMainlineAttendantSpamDetectionTemplate -Identity $env:MSTEAMS_LAB_MAINLINE_SPAM_ID -ErrorAction Stop
        $booking=Get-CsMainlineAttendantAppointmentBookingFlow -Identity $env:MSTEAMS_LAB_MAINLINE_BOOKING_ID -ErrorAction Stop
        $qa=Get-CsMainlineAttendantQuestionAnswerFlow -Identity $env:MSTEAMS_LAB_MAINLINE_QA_ID -ErrorAction Stop
        if(([string]$spam.Name) -notmatch '^msteams-sdk-[a-z2-7]{26,64}-spam$' -or ([string]$booking.Name) -notmatch '^msteams-sdk-[a-z2-7]{26,64}-booking$' -or
            ([string]$qa.Name) -notmatch '^msteams-sdk-[a-z2-7]{26,64}-qa$'){throw 'Owned fixtures required'}
        $report=@{success=$true;spam_name=[string]$spam.Name;spam_description=[string]$spam.Description;spam_enabled=[bool]$spam.EnableSpamDetection;
            spam_action=[string]$spam.Action;booking_name=[string]$booking.Name;booking_description=[string]$booking.Description;
            booking_platform=[string]$booking.AppointmentBookingPlatform;qa_name=[string]$qa.Name;qa_description=[string]$qa.Description}
    }
}catch{
    $report=@{success=$false}
}finally{
    if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null}
    if($null -ne $certificate){$certificate.Dispose()}
}
[Console]::Out.Write(($report|ConvertTo-Json -Depth 8 -Compress))
