# Pinned meeting-template configuration/permission reads; no writes.
$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue';$InformationPreference='SilentlyContinue'
$WarningPreference='SilentlyContinue';$VerbosePreference='SilentlyContinue';$DebugPreference='SilentlyContinue'
$report=@{success=$false};$connected=$false;$certificate=$null;$tokens=$null
try {
 if($env:MSTEAMS_LAB_READ_ONLY -ne '1' -or $env:MSTEAMS_LAB_POWERSHELL_ORACLE -ne '1'){throw 'Explicit read oracle opt-in required.'}
 $request=[Console]::In.ReadToEnd() | ConvertFrom-Json;$tenant=[Guid]::Parse($request.tenant_id).ToString()
 $selected=[string]$request.identity
 if($request.kind -notin @('configuration','permission','firstparty')){throw 'Unknown resource.'};if($selected -ine 'Global' -and ($request.kind -ne 'permission' -or $selected -notmatch '(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-meeting-template-policy$')){throw 'Explicit Global or owned permission policy required.'}
 $manifest=Join-Path $env:MSTEAMS_LAB_MODULE_PATH 'MicrosoftTeams.psd1'
 if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
 $assemblies=@(
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
 $values=@(if($request.kind -eq 'configuration'){Get-CsTeamsMeetingTemplateConfiguration -Identity Global -ErrorAction Stop}elseif($request.kind -eq 'firstparty'){Get-CsTeamsFirstPartyMeetingTemplateConfiguration -Identity Global -ErrorAction Stop}else{Get-CsTeamsMeetingTemplatePermissionPolicy -Identity $selected -ErrorAction Stop})
 if($values.Count -ne 1 -or (([string]$values[0].Identity -replace '^(?i)Tag:','') -ine ($selected -replace '^(?i)Tag:',''))){throw 'Expected Global resource.'}
 $settings=@{Description=$values[0].Description}
 if($request.kind -ne 'permission'){
  $templates=@(foreach($template in $values[0].TeamsMeetingTemplates){
   $options=@(foreach($option in $template.TeamsMeetingOptions){@{Identity=$option.Identity;Name=$option.Name;Value=$option.Value;IsLocked=$option.IsLocked;IsHidden=$option.IsHidden}})
   @{Identity=$template.Identity;Name=$template.Name;Description=$template.Description;Category=$template.Category;TeamsMeetingOptions=$options}
  })
  $settings['TeamsMeetingTemplates']=$templates
 }else{
  $settings['DefaultMeetingTemplateId']=$values[0].DefaultMeetingTemplateId
  $settings['HiddenMeetingTemplates']=@(foreach($template in $values[0].HiddenMeetingTemplates){@{Identity=$template.Identity;Id=$template.Id}})
 }
 $report=@{success=$true;tenant_id=$tenant;identity=[string]$values[0].Identity;settings=$settings}
} catch {$report=@{success=$false}}
finally{if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null};if($null -ne $certificate){$certificate.Dispose()}}
$tokens=$null;$token=$null;$identity=$null
[Console]::Out.Write(($report | ConvertTo-Json -Depth 24 -Compress))
