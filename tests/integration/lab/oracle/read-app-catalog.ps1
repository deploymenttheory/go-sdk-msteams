# Opt-in, read-only oracle for the EXT-21 app catalog commands. It reports only
# outcome classes and counts: Get-CsTeamsSettingsCustomApp's sideloading flag,
# Get-TeamsApp's outcome and, for client-secret runs, Get-TeamsArtifacts
# -SharePoint artifact/error counts. No app, artifact or URL is written.
$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue'
$InformationPreference='SilentlyContinue'
$WarningPreference='SilentlyContinue'
$VerbosePreference='SilentlyContinue'
$connected=$false
$certificate=$null
$scratch=$null
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
    $secretAuth=$env:MSTEAMS_LAB_AUTH_METHOD -eq 'client-secret'
    if($secretAuth){
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
    $sideloaded=[bool](Get-CsTeamsSettingsCustomApp -ErrorAction Stop).isSideloadedAppsInteractionEnabled
    $app='unknown'
    try{
        $null=Get-TeamsApp -DistributionMethod organization -ErrorAction Stop
        $app='listed'
    }catch{
        if(([string]$_.Exception.Message) -match 'Code: Forbidden'){$app='forbidden'}
    }
    $artifacts=@{outcome='skipped'}
    if($secretAuth){
        # The export tool authenticates itself from these process variables and
        # writes artifacts.json to the current directory.
        $scratch=Join-Path ([IO.Path]::GetTempPath()) ('msteams-sdk-artifacts-'+[Guid]::NewGuid().ToString('N'))
        $null=New-Item -ItemType Directory -Path $scratch
        [Environment]::SetEnvironmentVariable('EXPORT_TOOL.CLIENT_ID',$identity.client_id)
        [Environment]::SetEnvironmentVariable('EXPORT_TOOL.TENANT_ID',$tenant)
        [Environment]::SetEnvironmentVariable('EXPORT_TOOL.CLIENT_SECRET',$identity.client_secret)
        $stdout=[Console]::Out
        $console=[IO.StringWriter]::new()
        [Console]::SetOut($console)
        # The tool resolves artifacts.json against the process directory.
        $processDirectory=[Environment]::CurrentDirectory
        [Environment]::CurrentDirectory=$scratch
        Push-Location -LiteralPath $scratch
        try{
            Get-TeamsArtifacts -SharePoint -ErrorAction Stop 6>$null | Out-Null
        }finally{
            Pop-Location
            [Environment]::CurrentDirectory=$processDirectory
            [Console]::SetOut($stdout)
            [Environment]::SetEnvironmentVariable('EXPORT_TOOL.CLIENT_SECRET',$null)
        }
        $text=$console.ToString()
        $errors=-1
        if($text -match 'There were (\d+) errors'){$errors=[int]$Matches[1]}elseif($text -match 'No errors encountered'){$errors=0}
        $exported=@(Get-Content -LiteralPath (Join-Path $scratch 'artifacts.json') -Raw | ConvertFrom-Json)
        $artifacts=@{outcome='scanned';count=$exported.Count;errors=$errors;sites_forbidden=($text -match 'Access to search tenant SharePoint sites is forbidden')}
    }
    $report=@{success=$true;sideloaded=$sideloaded;app=$app;artifacts=$artifacts}
}catch{
    $report=@{success=$false}
}finally{
    [Environment]::SetEnvironmentVariable('EXPORT_TOOL.CLIENT_SECRET',$null)
    if($null -ne $scratch -and (Test-Path -LiteralPath $scratch)){Remove-Item -LiteralPath $scratch -Recurse -Force -ErrorAction SilentlyContinue}
    if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null}
    if($null -ne $certificate){$certificate.Dispose()}
}
[Console]::Out.Write(($report|ConvertTo-Json -Depth 8 -Compress))
