# Opt-in, read-only oracle for Shifts connector and sync reads. It reports
# only the output type of each read, never connection data.
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
    $types=[ordered]@{}
    # The cmdlets write correlation IDs straight to the console; keep stdout to the report.
    $stdout=[Console]::Out
    [Console]::SetOut([IO.TextWriter]::Null)
    try{
        # Sync reads by ID use reserved IDs that name no tenant object.
        $probe='00000000-0000-4000-8000-000000000001'
        $team='00000000-0000-4000-8000-000000000002'
        foreach($pair in @(@('connectors',{Get-CsTeamsShiftsConnectionConnector -ErrorAction Stop}),@('connections',{Get-CsTeamsShiftsConnection -ErrorAction Stop}),@('instances',{Get-CsTeamsShiftsConnectionInstance -ErrorAction Stop}),
            @('errorReports',{Get-CsTeamsShiftsConnectionErrorReport -ErrorAction Stop}),@('operation',{Get-CsTeamsShiftsConnectionOperation -OperationId $probe -ErrorAction Stop}),
            @('syncResult',{Get-CsTeamsShiftsConnectionSyncResult -ConnectorInstanceId $probe -TeamId $team -ErrorAction Stop}),@('teamMaps',{Get-CsTeamsShiftsConnectionTeamMap -ConnectorInstanceId $probe -ErrorAction Stop}),
            @('wfmTeams',{Get-CsTeamsShiftsConnectionWfmTeam -ConnectorInstanceId $probe -ErrorAction Stop}),@('wfmUsers',{Get-CsTeamsShiftsConnectionWfmUser -ConnectorInstanceId $probe -WfmTeamId 'msteams-sdk-probe' -ErrorAction Stop}))){
            $values=@(& $pair[1] 6>$null)
            $types[$pair[0]]=if($values.Count -eq 1){$values[0].GetType().Name}else{'count-' + $values.Count}
        }
    }finally{
        [Console]::SetOut($stdout)
    }
    $report=@{success=$true}
    foreach($key in $types.Keys){$report[$key]=$types[$key]}
}catch{
    $report=@{success=$false}
}finally{
    if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null}
    if($null -ne $certificate){$certificate.Dispose()}
}
[Console]::Out.Write(($report|ConvertTo-Json -Depth 8 -Compress))
