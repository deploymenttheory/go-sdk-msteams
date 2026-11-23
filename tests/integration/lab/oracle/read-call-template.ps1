# Opt-in, read-only oracle for one owned call queue template.
$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue'
$InformationPreference='SilentlyContinue'
$WarningPreference='SilentlyContinue'
$VerbosePreference='SilentlyContinue'
$connected=$false
$certificate=$null
$stage='module'
$report=@{success=$false}
try{
    if($env:MSTEAMS_LAB_READ_ONLY -ne '1' -or $env:MSTEAMS_LAB_POWERSHELL_ORACLE -ne '1'){throw 'Explicit read oracle opt-in required.'}
    $kind=$env:MSTEAMS_LAB_CALL_TEMPLATE_KIND
    $fixture=$env:MSTEAMS_LAB_CALL_TEMPLATE_FIXTURE
    $id=$env:MSTEAMS_LAB_CALL_TEMPLATE_ID
    if($kind -notin @('compliance','history','triage','tags')){throw 'Template kind required'}
    if($fixture -cnotmatch '^msteams-sdk-[a-z2-7]{26,64}-[a-z][a-z0-9-]*$'){throw 'Owned fixture required'}
    if($id -notmatch '^[0-9a-fA-F]{8}(-[0-9a-fA-F]{4}){3}-[0-9a-fA-F]{12}$'){throw 'Template ID required'}
    $module=$env:MSTEAMS_LAB_MODULE_PATH
    if((Import-PowerShellDataFile -LiteralPath (Join-Path $module 'MicrosoftTeams.psd1')).ModuleVersion -ne '8.1.0'){throw 'Wrong module baseline'}
    $assembly=Join-Path $module 'bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll'
    if((Get-FileHash -LiteralPath $assembly -Algorithm SHA256).Hash.ToLowerInvariant() -ne '8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'){throw 'Wrong module assembly'}
    $identity=Get-Content -LiteralPath $env:MSTEAMS_LAB_IDENTITY_FILE -Raw | ConvertFrom-Json
    Import-Module (Join-Path $module 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
    $stage='authentication'
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
    $stage='read'
    switch($kind){
        'compliance'{$v=Get-CsComplianceRecordingForCallQueueTemplate -Id $id -ErrorAction Stop}
        'history'{$v=Get-CsSharedCallHistoryTemplate -Id $id -ErrorAction Stop; $alias=Get-CsSharedCallQueueHistoryTemplate -Id $id -ErrorAction Stop -WarningAction SilentlyContinue}
        'triage'{$v=Get-CsSharedVoicemailTriageSettingsTemplate -Id $id -ErrorAction Stop}
        'tags'{$v=Get-CsTagsTemplate -Id $id -ErrorAction Stop}
    }
    if($null -eq $v -or $v.Name -cne $fixture -or [string]$v.Id -ne $id){throw 'Expected the owned template'}
    $stage='projection'
    $p=[ordered]@{name=[string]$v.Name;description=[string]$v.Description}
    switch($kind){
        'compliance'{
            $p.bot_id=[string]$v.BotApplicationInstanceObjectId
            $p.required_during_call=[bool]$v.RequiredDuringCall
            $p.required_before_call=[bool]$v.RequiredBeforeCall
            $p.concurrent_invitation_count=[int]$v.ConcurrentInvitationCount
        }
        'history'{
            $p.incoming_missed_calls=[int]$v.IncomingMissedCalls
            $p.answered_and_outbound_calls=[int]$v.AnsweredAndOutboundCalls
            $p.incoming_redirected_calls=[int]$v.IncomingRedirectedCalls
            $p.alias_matches=(($alias | ConvertTo-Json -Depth 4 -Compress) -ceq ($v | ConvertTo-Json -Depth 4 -Compress))
        }
        'triage'{
            $p.urgency=[bool]$v.EnableUrgencyDetection
            $p.category=[bool]$v.EnableCategoryDetection
            $p.call_to_action=[bool]$v.EnableCallToActionDetection
            $p.summary=[bool]$v.EnableVoiceToTextSummary
        }
        'tags'{
            $p.tags=@(foreach($tag in @($v.Tags)){[ordered]@{name=[string]$tag.TagName;target=[string]$tag.TagDetails.Id}})
        }
    }
    $report=@{success=$true;projection=$p}
}catch{
    $report=@{success=$false;failure_stage=$stage}
}finally{
    if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null}
    if($null -ne $certificate){$certificate.Dispose()}
}
[Console]::Out.Write(($report|ConvertTo-Json -Depth 8 -Compress))
