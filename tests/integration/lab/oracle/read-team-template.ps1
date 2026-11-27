# Opt-in, read-only oracle for one owned tenant template.
$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue'
$InformationPreference='SilentlyContinue'
$WarningPreference='SilentlyContinue'
$VerbosePreference='SilentlyContinue'
$connected=$false
$certificate=$null
$report=@{success=$false}
try{
    if($env:MSTEAMS_LAB_READ_ONLY -ne '1' -or $env:MSTEAMS_LAB_POWERSHELL_ORACLE -ne '1'){throw 'Explicit read oracle opt-in required.'}
    $fixture=$env:MSTEAMS_LAB_TEMPLATE_FIXTURE
    $reference=$env:MSTEAMS_LAB_TEMPLATE_ID
    if($fixture -notmatch '^msteams-sdk-[a-z2-7]{26,64}-template$'){throw 'Owned fixture required'}
    if($reference -notmatch '^/api/teamtemplates/v1\.0/[^/]+/Tenant/en-US$'){throw 'Owned tenant reference required'}
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
    $parameters=@{OdataId=$reference;ErrorAction='Stop'}
    if((Get-Command Get-CsTeamTemplate).Parameters.ContainsKey('MsftInternalProcessingMode')){$parameters.MsftInternalProcessingMode='ForceModern'}
    $values=@(Get-CsTeamTemplate @parameters)
    if($values.Count -ne 1 -or $values[0].DisplayName -cne $fixture){throw 'Expected one owned template'}
    # Pinned ITeam*Settings property names are singular; wire JSON is plural.
    $settings=@{}
    $fields=@{
      member_add_remove_apps=@('MemberSetting','AllowAddRemoveApp')
      member_private_channels=@('MemberSetting','AllowCreatePrivateChannel')
      member_create_update_channels=@('MemberSetting','AllowCreateUpdateChannel')
      member_connectors=@('MemberSetting','AllowCreateUpdateRemoveConnector')
      member_tabs=@('MemberSetting','AllowCreateUpdateRemoveTab')
      member_delete_channels=@('MemberSetting','AllowDeleteChannel')
      member_upload_custom_app=@('MemberSetting','UploadCustomApp')
      guest_create_update_channels=@('GuestSetting','AllowCreateUpdateChannel')
      guest_delete_channels=@('GuestSetting','AllowDeleteChannel')
      fun_custom_memes=@('FunSetting','AllowCustomMeme')
      fun_giphy=@('FunSetting','AllowGiphy')
      fun_stickers_memes=@('FunSetting','AllowStickersAndMeme')
      messaging_channel_mentions=@('MessagingSetting','AllowChannelMention')
      messaging_owner_delete=@('MessagingSetting','AllowOwnerDeleteMessage')
      messaging_team_mentions=@('MessagingSetting','AllowTeamMention')
      messaging_user_delete=@('MessagingSetting','AllowUserDeleteMessage')
      messaging_user_edit=@('MessagingSetting','AllowUserEditMessage')
      discovery_search=@('DiscoverySetting','ShowInTeamsSearchAndSuggestion')
    }
    foreach($key in $fields.Keys){
      $object=$values[0].PSObject.Properties[$fields[$key][0]]
      if($null -eq $object){throw 'Pinned settings object property missing'}
      $enabled=$false
      if($null -ne $object.Value){
        $property=$object.Value.PSObject.Properties[$fields[$key][1]]
        if($null -eq $property -or $property.Value -isnot [bool]){throw 'Pinned boolean setting missing'}
        $enabled=[bool]$property.Value
      }
      $settings[$key]=$enabled
    }
    $report=@{success=$true;settings=$settings;display_name=[string]$values[0].DisplayName;short_description=[string]$values[0].ShortDescription;description=[string]$values[0].Description;channel_names=@($values[0].Channel|Where-Object{$null -ne $_}|ForEach-Object{[string]$_.DisplayName})}
}catch{
    $report=@{success=$false}
}finally{
    if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null}
    if($null -ne $certificate){$certificate.Dispose()}
}
[Console]::Out.Write(($report|ConvertTo-Json -Depth 8 -Compress))
