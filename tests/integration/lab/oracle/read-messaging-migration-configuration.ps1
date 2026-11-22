# Pinned configuration/policy Global getters; no tenant writes.
$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue';$InformationPreference='SilentlyContinue'
$WarningPreference='SilentlyContinue';$VerbosePreference='SilentlyContinue';$DebugPreference='SilentlyContinue'
$report=@{success=$false};$connected=$false;$certificate=$null;$tokens=$null
try {
 if($env:MSTEAMS_LAB_READ_ONLY -ne '1' -or $env:MSTEAMS_LAB_POWERSHELL_ORACLE -ne '1'){throw 'Explicit read oracle opt-in required.'}
 $request=[Console]::In.ReadToEnd() | ConvertFrom-Json;$tenant=[Guid]::Parse($request.tenant_id).ToString()
 $selected=[string]$request.identity
 if($selected -ine 'Global' -or $request.kind -notin @('messaging','migration','mobility','multi-tenant','roaming','notification','personal','recording-rollout')){throw 'Explicit configuration read required.'}
 $manifest=Join-Path $env:MSTEAMS_LAB_MODULE_PATH 'MicrosoftTeams.psd1'
 if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
 $assemblies=@(
 @{path='bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll';sha256='8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'},
  @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
  @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'}
 )
 foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $env:MSTEAMS_LAB_MODULE_PATH $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned assembly.'}}
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
 $connected=$true
 $values=@(if($request.kind -eq 'messaging'){Get-CsTeamsMessagingConfiguration -Identity Global -ErrorAction Stop}elseif($request.kind -eq 'migration'){Get-CsTeamsMigrationConfiguration -Identity Global -ErrorAction Stop}elseif($request.kind -eq 'mobility'){Get-CsTeamsMobilityPolicy -Identity Global -ErrorAction Stop}elseif($request.kind -eq 'roaming'){Get-CsTeamsNetworkRoamingPolicy -Identity Global -ErrorAction Stop}elseif($request.kind -eq 'notification'){Get-CsTeamsNotificationAndFeedsPolicy -Identity Global -ErrorAction Stop}elseif($request.kind -eq 'personal'){Get-CsTeamsPersonalAttendantPolicy -Identity Global -ErrorAction Stop}elseif($request.kind -eq 'recording-rollout'){Get-CsTeamsRecordingRollOutPolicy -Identity Global -ErrorAction Stop}else{Get-CsTeamsMultiTenantOrganizationConfiguration -Identity Global -ErrorAction Stop})
 if($values.Count -ne 1 -or [string]$values[0].Identity -ine 'Global'){throw 'Expected Global configuration.'}
 $names=if($request.kind -eq 'messaging'){@('EnableVideoMessageCaptions','EnableInOrganizationChatControl','CustomEmojis','Storyline','Communities','MessagingNotes','FileTypeCheck','UrlReputationCheck','ContentBasedPhishingCheck','ReportIncorrectSecurityDetections','FileTypeCustomizationSetting','FileTypeBlockList')}elseif($request.kind -eq 'migration'){@('EnableLegacyClientInterop')}elseif($request.kind -eq 'mobility'){@('Description','IPAudioMobileMode','IPVideoMobileMode','MobileDialerPreference','LinksInTeams')}elseif($request.kind -eq 'roaming'){@('Description','AllowIPVideo','MediaBitRateKb')}elseif($request.kind -eq 'notification'){@('Description','SuggestedFeedsEnabledType','TrendingFeedsEnabledType')}elseif($request.kind -eq 'personal'){@('PersonalAttendant','CallScreening','CalendarBookings','InboundInternalCalls','InboundFederatedCalls','InboundPSTNCalls','AutomaticTranscription','AutomaticRecording')}elseif($request.kind -eq 'recording-rollout'){@('MeetingRecordingOwnership')}else{@('CopilotFromHomeTenant')}
 $settings=@{}
 foreach($name in $names){$property=$values[0].PSObject.Properties[$name];if($null -eq $property){throw 'Missing configuration getter.'};$settings[$name]=$property.Value}
 $report=@{success=$true;tenant_id=$tenant;identity=[string]$values[0].Identity;settings=$settings}
} catch {$report=@{success=$false}}
finally{if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null};if($null -ne $certificate){$certificate.Dispose()}}
$tokens=$null;$token=$null;$identity=$null
[Console]::Out.Write(($report | ConvertTo-Json -Depth 12 -Compress))
