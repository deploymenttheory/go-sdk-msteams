# Pinned meeting broadcast configuration Global getter; no tenant writes.
# SdnApiToken is never emitted: only whether the exposed value is non-empty.
$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue';$InformationPreference='SilentlyContinue'
$WarningPreference='SilentlyContinue';$VerbosePreference='SilentlyContinue';$DebugPreference='SilentlyContinue'
$report=@{success=$false};$connected=$false;$certificate=$null;$tokens=$null
try {
 if($env:MSTEAMS_LAB_READ_ONLY -ne '1' -or $env:MSTEAMS_LAB_POWERSHELL_ORACLE -ne '1'){throw 'Explicit read oracle opt-in required.'}
 $request=[Console]::In.ReadToEnd() | ConvertFrom-Json;$tenant=[Guid]::Parse($request.tenant_id).ToString()
 $selected=[string]$request.identity
 if($selected -ine 'Global' -or $request.kind -ne 'meeting-broadcast'){throw 'Explicit configuration read required.'}
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
 $values=@(Get-CsTeamsMeetingBroadcastConfiguration -Identity Global -ErrorAction Stop)
 if($values.Count -ne 1 -or [string]$values[0].Identity -ine 'Global'){throw 'Expected Global configuration.'}
 $settings=@{}
 foreach($name in @('SupportURL','AllowSdnProviderForBroadcastMeeting')){$property=$values[0].PSObject.Properties[$name];if($null -eq $property){throw 'Missing configuration getter.'};$settings[$name]=$property.Value}
 $exposed=@(Get-CsTeamsMeetingBroadcastConfiguration -Identity Global -ExposeSDNConfigurationJsonBlob -ErrorAction Stop)
 if($exposed.Count -ne 1 -or [string]$exposed[0].Identity -ine 'Global'){throw 'Expected exposed Global configuration.'}
 $sdn=@{}
 foreach($name in @('SupportURL','AllowSdnProviderForBroadcastMeeting','SdnName','SdnLicenseId','SdnAzureSubscriptionId','SdnApiTemplateUrl','SdnRuntimeconfiguration','SdnAttendeeFallbackCount')){$property=$exposed[0].PSObject.Properties[$name];if($null -eq $property){throw 'Missing exposed getter.'};$sdn[$name]=$property.Value}
 $secret=$exposed[0].PSObject.Properties['SdnApiToken']
 if($null -eq $secret){throw 'Missing exposed getter.'}
 $sdn['SdnApiTokenPresent']=-not [string]::IsNullOrEmpty([string]$secret.Value)
 $secret=$null;$exposed=$null
 $report=@{success=$true;tenant_id=$tenant;identity=[string]$values[0].Identity;settings=$settings;sdn=$sdn}
} catch {$report=@{success=$false}}
finally{if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null};if($null -ne $certificate){$certificate.Dispose()}}
$tokens=$null;$token=$null;$identity=$null;$secret=$null;$exposed=$null
[Console]::Out.Write(($report | ConvertTo-Json -Depth 12 -Compress))
