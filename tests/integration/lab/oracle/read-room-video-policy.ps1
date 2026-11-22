# Optional pinned modern Cortana policy selected read; no tenant writes.
$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue';$InformationPreference='SilentlyContinue'
$WarningPreference='SilentlyContinue';$VerbosePreference='SilentlyContinue';$DebugPreference='SilentlyContinue'
$report=@{success=$false};$connected=$false;$certificate=$null;$tokens=$null
try {
 if($env:MSTEAMS_LAB_READ_ONLY -ne '1' -or $env:MSTEAMS_LAB_POWERSHELL_ORACLE -ne '1'){throw 'Explicit read oracle opt-in required.'}
 $request=[Console]::In.ReadToEnd() | ConvertFrom-Json;$tenant=[Guid]::Parse($request.tenant_id).ToString()
 $selected=[string]$request.identity
 if($selected -ine 'Global' -and $selected -notmatch '(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-room-video-policy$'){throw 'Invalid owned policy selector.'}
 $manifest=Join-Path $env:MSTEAMS_LAB_MODULE_PATH 'MicrosoftTeams.psd1'
 if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
 $assemblies=@(
  @{path='bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll';sha256='8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'},
  @{path='custom/Merged_custom_PsExt.ps1';sha256='03dbc3a85e12c98dd1807d604782c6a90d0656569541cd9540738825a7fcb255'},
  @{path='custom/CmdletConfig.json';sha256='754d149c1a5904158e9f01eb55ba44d15c17b7e66f7556385ea894f4ff84d906'},
  @{path='exports/ProxyCmdletDefinitionsWithHelp.ps1';sha256='5a8a5a50081b1c6986b93d5b15723e966a2d22b64459a9c300cb2a1bf69bb3b5'}
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
 $connected=$true;$values=@(Get-CsTeamsRoomVideoTeleConferencingPolicy -MsftInternalProcessingMode ForceModern -Identity $selected -ErrorAction Stop)
 if($values.Count -ne 1 -or ([string]$values[0].Identity -replace '^(?i)Tag:','') -ine ($selected -replace '^(?i)Tag:','')){throw 'Expected selected policy.'}
 $settings=@{}
 foreach($name in @('Description','Enabled','AreaCode','ReceiveExternalCalls','ReceiveInternalCalls','PlaceExternalCalls','PlaceInternalCalls')){
  $property=$values[0].PSObject.Properties[$name];if($null -eq $property){throw 'Incomplete Cortana policy projection.'}
  $settings[$name]=$property.Value
 }
 $report=@{success=$true;tenant_id=$tenant;identity=[string]$values[0].Identity;settings=$settings}
} catch {$report=@{success=$false}}
finally{if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null};if($null -ne $certificate){$certificate.Dispose()}}
$tokens=$null;$token=$null;$identity=$null
[Console]::Out.Write(($report | ConvertTo-Json -Depth 12 -Compress))
