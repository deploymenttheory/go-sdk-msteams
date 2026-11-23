# Verify exposed migration arguments fail pinned modern binding; WhatIf guards execution.
# Pinned configuration/policy Global getters; no tenant writes.
$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue';$InformationPreference='SilentlyContinue'
$WarningPreference='SilentlyContinue';$VerbosePreference='SilentlyContinue';$DebugPreference='SilentlyContinue'
$report=@{success=$false};$connected=$false;$certificate=$null;$tokens=$null
try {
 if($env:MSTEAMS_LAB_READ_ONLY -ne '1' -or $env:MSTEAMS_LAB_POWERSHELL_ORACLE -ne '1'){throw 'Explicit read oracle opt-in required.'}
 $request=[Console]::In.ReadToEnd() | ConvertFrom-Json
 $tenant=[Guid]::Parse($request.tenant_id).ToString()
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

 $results=@()
 foreach($bind in @('false','true')){
  try {
   $params=@{Identity='msteams-sdk-binding-only@invalid.example';PolicyName='UpgradeToTeams';WhatIf=$true;MsftInternalProcessingMode='ForceModern';ErrorAction='Stop'}
   if($bind -ne 'omitted'){$params.MigrateMeetingsToTeams=($bind -eq 'true')}
   Grant-CsTeamsUpgradePolicy @params | Out-Null
   $results+=@{bind=$bind;success=$true}
  }catch{$results+=@{bind=$bind;success=$false;error_id=$_.FullyQualifiedErrorId;exception_type=$_.Exception.GetType().FullName}}
 }
 $report=@{success=$true;results=$results}
} catch {$report=@{success=$false}}
finally{if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null};if($null -ne $certificate){$certificate.Dispose()}}
$tokens=$null;$token=$null;$identity=$null
[Console]::Out.Write(($report | ConvertTo-Json -Depth 12 -Compress))
