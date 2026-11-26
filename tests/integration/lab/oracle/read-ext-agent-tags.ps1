# Read-only oracle for an owned EXT agent and its owned IVR tag template.
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
    $agentId=$env:MSTEAMS_LAB_AGENT_ID
    $templateId=$env:MSTEAMS_LAB_TAG_TEMPLATE_ID
    $agentName=$env:MSTEAMS_LAB_AGENT_FIXTURE
    $templateName=$env:MSTEAMS_LAB_TAG_TEMPLATE_FIXTURE
    foreach($id in @($agentId,$templateId)){if($id -notmatch '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'){throw 'Owned identity required'}}
    if($agentName -notmatch '^msteams-sdk-[a-z2-7]{26,64}-agent(-renamed)?$' -or $templateName -notmatch '^msteams-sdk-[a-z2-7]{26,64}-tags$'){throw 'Owned fixtures required'}
    $agent=Get-CsAgent -Id $agentId -ErrorAction Stop
    $template=Get-CsTagsTemplate -Id $templateId -ErrorAction Stop
    if($agent.Name -ne $agentName -or $template.Name -ne $templateName){throw 'Not owned resources'}
    $report=@{success=$true;agent_name=[string]$agent.Name;template_name=[string]$template.Name;template_id=[string]$template.Id;agent_template_id=[string]$agent.AIAgentTargetTagTemplateId;tag_count=@($template.Tags).Count}
}catch{
    $report=@{success=$false}
}finally{
    [Console]::SetOut($stdout)
    if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null}
    if($null -ne $certificate){$certificate.Dispose()}
}
[Console]::Out.Write(($report|ConvertTo-Json -Depth 8 -Compress))
