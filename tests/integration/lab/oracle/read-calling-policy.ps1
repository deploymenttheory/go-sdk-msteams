# Optional pinned compiled calling Global read; no tenant writes.
$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue';$InformationPreference='SilentlyContinue'
$WarningPreference='SilentlyContinue';$VerbosePreference='SilentlyContinue';$DebugPreference='SilentlyContinue'
$report=@{success=$false};$connected=$false;$certificate=$null;$stage='enablement'
try {
 if($env:MSTEAMS_LAB_READ_ONLY -ne '1' -or $env:MSTEAMS_LAB_POWERSHELL_ORACLE -ne '1'){throw 'Explicit read oracle opt-in required.'}
 $stage='selectors';$request=[Console]::In.ReadToEnd() | ConvertFrom-Json;$tenant=[Guid]::Parse($request.tenant_id).ToString()
 $selected=[string]$request.identity;if([string]::IsNullOrEmpty($selected)){$selected='Global'}
 $fixture=$selected -replace '^(?i)tag:',''
 if($selected -ine 'Global' -and $fixture -notmatch '^msteams-sdk-[a-z2-7]{26,64}-calling-policy$'){throw 'Explicit owned identity required.'}
 $stage='baseline';$manifest=Join-Path $env:MSTEAMS_LAB_MODULE_PATH 'MicrosoftTeams.psd1'
 if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
 $assemblies=@(
  @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
  @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'}
 )
 foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $env:MSTEAMS_LAB_MODULE_PATH $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned assembly.'}}
 $stage='module';Import-Module $manifest -Force -WarningAction SilentlyContinue
 $stage='identity';$identity=Get-Content -LiteralPath $env:MSTEAMS_LAB_IDENTITY_FILE -Raw | ConvertFrom-Json
 if([Guid]::Parse($identity.tenant_id).ToString() -ne $tenant){throw 'Tenant binding differs.'}
 $stage='authentication';if($env:MSTEAMS_LAB_AUTH_METHOD -eq 'client-secret'){
   $authTenant=[Guid]::Parse($identity.tenant_id).ToString();$tokens=@()
   foreach($scope in @('https://graph.microsoft.com/.default','48ac35b8-9aa8-4d74-927d-1f4a14a0b239/.default')){
     $token=Invoke-RestMethod -Method Post -Uri "https://login.microsoftonline.com/$authTenant/oauth2/v2.0/token" -Body @{grant_type='client_credentials';client_id=$identity.client_id;client_secret=$identity.client_secret;scope=$scope}
     $tokens+= $token.access_token
   }
   Connect-MicrosoftTeams -AccessTokens $tokens -ErrorAction Stop -InformationAction SilentlyContinue -WarningAction SilentlyContinue | Out-Null
 }elseif([string]::IsNullOrEmpty($env:MSTEAMS_LAB_AUTH_METHOD) -or $env:MSTEAMS_LAB_AUTH_METHOD -eq 'certificate'){
   $certificate=[Security.Cryptography.X509Certificates.X509Certificate2]::CreateFromPemFile($identity.certificate_file,$identity.private_key_file)
   Connect-MicrosoftTeams -Certificate $certificate -ApplicationId $identity.client_id -TenantId $identity.tenant_id -ErrorAction Stop -InformationAction SilentlyContinue -WarningAction SilentlyContinue | Out-Null
 }else{throw 'Unknown auth method.'}
 $connected=$true;$stage='read';$values=@(Get-CsTeamsCallingPolicy -Identity $selected -ErrorAction Stop)
 $stage='selected-result';if($values.Count -ne 1 -or (([string]$values[0].Identity -replace '^(?i)tag:','') -ine $fixture)){throw 'Expected Global policy.'}
 $settings=@{}
 foreach($name in @('Description','AllowPrivateCalling','AllowWebPSTNCalling','AllowSIPDevicesCalling','AllowVoicemail','AllowCallGroups','AllowDelegation','AllowCallForwardingToUser','AllowCallForwardingToPhone','PreventTollBypass','BusyOnBusyEnabledType','MusicOnHoldEnabledType','AllowCloudRecordingForCalls','ExplicitRecordingConsent','PreventComplianceRecording','EnableRecordingAndTranscriptionCustomMessage','RecordingAndTranscriptionCustomMessageIdentifier','AllowMeetingKnowledgeGeneration','AllowTranscriptionForCalling','RecordingAndTranscriptionAudioNotification','PopoutForIncomingPstnCalls','PopoutAppPathForIncomingPstnCalls','LiveCaptionsEnabledTypeForCalling','AutoAnswerEnabledType','SpamFilteringEnabledType','CallRecordingExpirationDays','AllowCallRedirect','InboundPstnCallRoutingTreatment','InboundFederatedCallRoutingTreatment','EnableWebPstnMediaBypass','EnableSpendLimits','CallingSpendUserLimit','Copilot','ShowTeamsCallsInCallLog','RealTimeText','AIInterpreter','VoiceSimulationInInterpreter','ReportCall','VoicePhishingDetection')){
  $stage='property:'+$name;$property=$values[0].PSObject.Properties[$name];if($null -eq $property){throw 'Incomplete calling projection.'};$settings[$name]=$property.Value
 }
 $report=@{success=$true;tenant_id=$tenant;identity=[string]$values[0].Identity;settings=$settings}
} catch {$report=@{success=$false;failure_stage=$stage}}
finally{if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null};if($null -ne $certificate){$certificate.Dispose()}}
[Console]::Out.Write(($report | ConvertTo-Json -Depth 12 -Compress))
