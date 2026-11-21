# Read-only comparison of known tenant capability/authorization outcomes.
$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue';$InformationPreference='SilentlyContinue'
$WarningPreference='SilentlyContinue';$VerbosePreference='SilentlyContinue';$DebugPreference='SilentlyContinue'
$report=@{success=$false};$connected=$false;$certificate=$null;$tokens=$null
try {
 if($env:MSTEAMS_LAB_READ_ONLY -ne '1' -or $env:MSTEAMS_LAB_POWERSHELL_ORACLE -ne '1'){throw 'Explicit read oracle opt-in required.'}
 $request=[Console]::In.ReadToEnd() | ConvertFrom-Json
 $tenant=[Guid]::Parse($request.tenant_id).ToString()
 $commands=@{migration='Get-DirectToGroupAssignmentsMigrationStatus';prompts='Get-CsTeamsAudioConferencingCustomPromptsConfiguration';privacy='Get-CsPrivacyConfiguration'}
 if(-not $commands.ContainsKey([string]$request.resource)){throw 'Unknown read selector.'}
 $manifest=Join-Path $env:MSTEAMS_LAB_MODULE_PATH 'MicrosoftTeams.psd1'
 if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
 foreach($entry in @(
  @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';hash='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
  @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';hash='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'}
 )){if((Get-FileHash -LiteralPath (Join-Path $env:MSTEAMS_LAB_MODULE_PATH $entry.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $entry.hash){throw 'Wrong pinned assembly.'}}
 Import-Module $manifest -Force -WarningAction SilentlyContinue
 $identity=Get-Content -LiteralPath $env:MSTEAMS_LAB_IDENTITY_FILE -Raw | ConvertFrom-Json
 if([Guid]::Parse($identity.tenant_id).ToString() -ne $tenant){throw 'Tenant binding differs.'}
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
 $connected=$true;$outcome=$null;$code=$null;$errorCode=$null
 try {
  if($request.resource -eq 'migration'){
   $values=@(Get-DirectToGroupAssignmentsMigrationStatus -ErrorAction Stop)
   if($values.Count -eq 1 -and $values[0] -is [string] -and $values[0] -ceq 'No migration status to show, please verify the migration event id.'){$outcome='migration-status-message'}
  }elseif($request.resource -eq 'prompts'){
   Get-CsTeamsAudioConferencingCustomPromptsConfiguration -Identity Global -ErrorAction Stop | Out-Null
  }else{
   Get-CsPrivacyConfiguration -Identity Global -ErrorAction Stop | Out-Null
  }
 } catch {
  $exception=$_.Exception
  for($depth=0;$null -ne $exception -and $depth -lt 8;$depth++){
   try {
    $body=$exception.Message | ConvertFrom-Json -ErrorAction Stop
    if($request.resource -eq 'prompts' -and $body.code -ceq 'ClientError' -and $body.errorCode -eq 40003 -and $body.message -ceq 'Type TeamsAudioConferencingCustomPromptsConfiguration is not currently enabled in flighting'){
     $outcome='feature-unavailable';$code='ClientError';$errorCode=40003;break
    }
    if($request.resource -eq 'privacy' -and $body.code -ceq 'Forbidden' -and $body.errorCode -eq 40301){
     $outcome='forbidden';$code='Forbidden';$errorCode=40301;break
    }
   }catch{}
   $exception=$exception.InnerException
  }
 }
 if($null -ne $outcome){$report=@{success=$true;tenant_id=$tenant;command=$commands[[string]$request.resource];outcome=$outcome;code=$code;error_code=$errorCode}}
} catch {$report=@{success=$false}}
finally {
 if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null}
 if($null -ne $certificate){$certificate.Dispose()}
 $tokens=$null;$token=$null;$identity=$null
}
[Console]::Out.Write(($report | ConvertTo-Json -Depth 5 -Compress))
