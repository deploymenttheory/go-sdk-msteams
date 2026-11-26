# Optional pinned modern media logging policy and owned-fixture assignment reads; no tenant writes.
# kind=policy selects Global or Tag:Enabled; kind=user/kind=group read only the bound SDK fixtures.
$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue';$InformationPreference='SilentlyContinue'
$WarningPreference='SilentlyContinue';$VerbosePreference='SilentlyContinue';$DebugPreference='SilentlyContinue'
$report=@{success=$false};$connected=$false;$certificate=$null;$tokens=$null
try {
 if($env:MSTEAMS_LAB_READ_ONLY -ne '1' -or $env:MSTEAMS_LAB_POWERSHELL_ORACLE -ne '1'){throw 'Explicit read oracle opt-in required.'}
 $request=[Console]::In.ReadToEnd() | ConvertFrom-Json;$tenant=[Guid]::Parse($request.tenant_id).ToString()
 $kind=[string]$request.kind
 if($kind -ceq 'policy'){
  $selected=[string]$request.identity
  if($selected -cne 'Global' -and $selected -cne 'Tag:Enabled'){throw 'Invalid built-in policy selector.'}
 }elseif($kind -ceq 'user'){
  $user=[Guid]::Parse($request.user_id).ToString()
  $fixture=Get-Content -LiteralPath $env:MSTEAMS_LAB_USER_FIXTURE_FILE -Raw | ConvertFrom-Json
  if([Guid]::Parse($fixture.tenant_id).ToString() -ne $tenant -or [Guid]::Parse($fixture.user_id).ToString() -ne $user -or $fixture.display_name -cne 'go-sdk-msteams acceptance user' -or $fixture.user_principal_name -notlike 'go-sdk-msteams-acceptance@*'){throw 'Wrong owned user binding.'}
 }elseif($kind -ceq 'group'){
  $group=[Guid]::Parse($request.group_id).ToString()
  $fixture=Get-Content -LiteralPath $env:MSTEAMS_LAB_GROUP_FIXTURE_FILE -Raw | ConvertFrom-Json
  if([Guid]::Parse($fixture.tenant_id).ToString() -ne $tenant -or [Guid]::Parse($fixture.group_id).ToString() -ne $group -or $fixture.display_name -cne 'go-sdk-msteams acceptance policy group' -or $fixture.mail_nickname -cne 'go-sdk-msteams-acceptance-policy' -or $fixture.security_enabled -ne $true -or $fixture.mail_enabled -ne $false -or $fixture.role_assignable -ne $false -or $fixture.empty_membership_verified -ne $true){throw 'Wrong owned group binding.'}
 }else{throw 'Invalid oracle kind.'}
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
 $connected=$true
 if($kind -ceq 'policy'){
  $values=@(Get-CsTeamsMediaLoggingPolicy -MsftInternalProcessingMode ForceModern -Identity $selected -ErrorAction Stop)
  if($values.Count -ne 1 -or [string]$values[0].Identity -cne $selected){throw 'Expected selected policy.'}
  $settings=@{}
  foreach($name in @('Description','AllowMediaLogging')){
   $property=$values[0].PSObject.Properties[$name];if($null -eq $property){throw 'Incomplete media logging policy projection.'}
   $settings[$name]=$property.Value
  }
  $report=@{success=$true;tenant_id=$tenant;identity=[string]$values[0].Identity;settings=$settings}
 }elseif($kind -ceq 'user'){
  $parameters=@{Identity=$user;PolicyType='TeamsMediaLoggingPolicy';ErrorAction='Stop'}
  if((Get-Command Get-CsUserPolicyAssignment).Parameters.ContainsKey('MsftInternalProcessingMode')){$parameters.MsftInternalProcessingMode='ForceModern'}
  $policies=@()
  foreach($value in @(Get-CsUserPolicyAssignment @parameters)){
   $sources=@()
   foreach($source in @($value.PolicySource)){
    if($null -eq $source){continue}
    $item=@{}
    foreach($entry in @(@{key='policy_name';name='PolicyName'},@{key='policy_type';name='PolicyType'},@{key='assignment_type';name='AssignmentType'},@{key='reference';name='Reference'})){
     $property=$source.PSObject.Properties[$entry.name]
     $item[$entry.key]=$null
     if($null -ne $property -and $null -ne $property.Value){$item[$entry.key]=[string]$property.Value}
    }
    $sources+= $item
   }
   $policies+= @{policy_type=[string]$value.PolicyType;policy_name=[string]$value.PolicyName;sources=@($sources)}
  }
  $report=@{success=$true;tenant_id=$tenant;user_id=$user;policies=@($policies)}
 }else{
  $parameters=@{GroupId=$group;ErrorAction='Stop'}
  if((Get-Command Get-CsGroupPolicyAssignment).Parameters.ContainsKey('MsftInternalProcessingMode')){$parameters.MsftInternalProcessingMode='ForceModern'}
  $assignments=@()
  foreach($value in @(Get-CsGroupPolicyAssignment @parameters)){
   if([string]$value.PolicyType -ne 'TeamsMediaLoggingPolicy'){throw 'Unexpected policy type on owned group.'}
   $assignments+= @{group_id=[string]$value.GroupId;policy_type=[string]$value.PolicyType;policy_name=[string]$value.PolicyName;priority=[int]$value.Priority}
  }
  $report=@{success=$true;tenant_id=$tenant;group_id=$group;assignments=@($assignments)}
 }
} catch {$report=@{success=$false}}
finally{if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null};if($null -ne $certificate){$certificate.Dispose()}}
$tokens=$null;$token=$null;$identity=$null
[Console]::Out.Write(($report | ConvertTo-Json -Depth 12 -Compress))
