# Opt-in, read-only oracle for auto attendants. Mode "reads" projects the
# tenant information and supported language/time zone reads; mode
# "attendant" projects one owned auto attendant.
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
    $mode=$env:MSTEAMS_LAB_AUTO_ATTENDANT_MODE
    if($mode -notin @('reads','attendant')){throw 'Oracle mode required'}
    if($mode -eq 'attendant'){
        $fixture=$env:MSTEAMS_LAB_AUTO_ATTENDANT_FIXTURE
        $id=$env:MSTEAMS_LAB_AUTO_ATTENDANT_ID
        if($fixture -notmatch '^msteams-sdk-[a-z2-7]{26,64}-attendant$'){throw 'Owned fixture required'}
        if($id -notmatch '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'){throw 'Auto attendant ID required'}
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
    if($mode -eq 'reads'){
        $info=Get-CsAutoAttendantTenantInformation -ErrorAction Stop
        $languages=@(Get-CsAutoAttendantSupportedLanguage -ErrorAction Stop)
        $english=Get-CsAutoAttendantSupportedLanguage -Identity 'en-US' -ErrorAction Stop
        $zones=@(Get-CsAutoAttendantSupportedTimeZone -ErrorAction Stop)
        $utc=Get-CsAutoAttendantSupportedTimeZone -Identity 'UTC' -ErrorAction Stop
        $report=@{success=$true;default_language=[string]$info.DefaultLanguageId;default_time_zone=[string]$info.DefaultTimeZoneId;language_count=$languages.Count;time_zone_count=$zones.Count;en_us_voices=@($english.Voices).Count;utc_display_name=[string]$utc.DisplayName}
    }else{
        $attendant=Get-CsAutoAttendant -Identity $id -ErrorAction Stop
        if($attendant.Name -ne $fixture){throw 'Not the owned attendant'}
        $targets=[string[]]@($attendant.DefaultCallFlow.Menu.MenuOptions | ForEach-Object { [string]$_.CallTarget.Id })
        $flows=[string[]]@($attendant.CallFlows | ForEach-Object { [string]$_.Id })
        $report=@{success=$true;name=[string]$attendant.Name;language_id=[string]$attendant.LanguageId;time_zone_id=[string]$attendant.TimeZoneId;voice_id=[string]$attendant.VoiceId;default_call_flow_id=[string]$attendant.DefaultCallFlow.Id;greeting=[string]$attendant.DefaultCallFlow.Greetings[0].TextToSpeechPrompt;menu_option_targets=$targets;call_flow_ids=$flows}
    }
}catch{
    $report=@{success=$false}
}finally{
    if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null}
    if($null -ne $certificate){$certificate.Dispose()}
}
[Console]::Out.Write(($report|ConvertTo-Json -Depth 8 -Compress))
