# Optional read-only 8.1.0 oracle. It never creates, changes or deletes a resource.
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
$InformationPreference = 'SilentlyContinue'
$WarningPreference = 'SilentlyContinue'
$VerbosePreference = 'SilentlyContinue'
$DebugPreference = 'SilentlyContinue'
$report = @{success=$false; stage='input'}
$connected = $false
$certificate = $null
try {
    if ($env:MSTEAMS_LAB_READ_ONLY -ne '1' -or $env:MSTEAMS_LAB_POWERSHELL_ORACLE -ne '1') { throw 'Explicit read oracle opt-in required.' }
    $raw = [Console]::In.ReadToEnd()
    if ($raw.Length -gt 8192) { throw 'Invalid selectors.' }
    $request = $raw | ConvertFrom-Json
    $uuid = '^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$'
    if ($request.id -notmatch $uuid -or $request.fixture -notmatch '^msteams-sdk-[a-z2-7]{26,64}-(team|migration|membership)$') { throw 'Owned fixture required.' }
    if ($request.kind -notin @('team','users')) { throw 'Unknown oracle.' }
    if ($request.kind -eq 'team' -and $request.owner -notmatch $uuid) { throw 'Explicit owner required.' }
    $report.stage = 'module'
    $module = $env:MSTEAMS_LAB_MODULE_PATH
    $manifest = Join-Path $module 'MicrosoftTeams.psd1'
    if ((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0') { throw 'Wrong baseline.' }
    $assembly = Join-Path $module 'netcoreapp3.1/Microsoft.Teams.PowerShell.TeamsCmdlets.dll'
    if ((Get-FileHash -LiteralPath $assembly -Algorithm SHA256).Hash.ToLowerInvariant() -ne '3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24') { throw 'Wrong baseline assembly.' }
    Import-Module $manifest -Force -WarningAction SilentlyContinue
    $report.stage = 'authentication'
    $identity = Get-Content -LiteralPath $env:MSTEAMS_LAB_IDENTITY_FILE -Raw | ConvertFrom-Json
    $certificate = [Security.Cryptography.X509Certificates.X509Certificate2]::CreateFromPemFile($identity.certificate_file, $identity.private_key_file)
    Connect-MicrosoftTeams -Certificate $certificate -ApplicationId $identity.client_id -TenantId $identity.tenant_id -ErrorAction Stop -WarningAction SilentlyContinue -InformationAction SilentlyContinue | Out-Null
    $connected = $true
    $report.stage = 'read'
    $values = @(Get-Team -GroupId $request.id -ErrorAction Stop)
    if ($values.Count -ne 1 -or [string]$values[0].GroupId -ine $request.id -or [string]$values[0].DisplayName -cne $request.fixture) { throw 'Expected one owned fixture.' }
    function Convert-TeamObservation($value) {
        foreach ($property in @('GroupId','DisplayName','Description','Visibility','MailNickName','Classification','Archived','AllowGiphy')) {
            if ($null -eq $value.PSObject.Properties[$property]) { throw 'Incomplete team projection.' }
        }
        return @{
            id=[string]$value.GroupId; display_name=$value.DisplayName
            description=$value.Description; visibility=$value.Visibility
            alias=$value.MailNickName; classification=$value.Classification
            archived=$value.Archived; allow_giphy=$value.AllowGiphy
        }
    }
    function Convert-TeamUserObservation($value) {
        foreach ($property in @('UserId','User','Name','Role')) {
            if ($null -eq $value.PSObject.Properties[$property]) { throw 'Incomplete user projection.' }
        }
        return @{id=[string]$value.UserId; user=$value.User; name=$value.Name; role=([string]$value.Role).ToLowerInvariant()}
    }
    $report.stage = 'projection'
    if ($request.kind -eq 'team') {
        $team = Convert-TeamObservation $values[0]
        $tenant = @(Get-Team -DisplayName $request.fixture -ErrorAction Stop)
        $user = @(Get-Team -User $request.owner -DisplayName $request.fixture -ErrorAction Stop)
        if ($tenant.Count -ne 1 -or $user.Count -ne 1) { throw 'Expected complete single-fixture lists.' }
        $report = @{success=$true; stage='complete'; team=$team; tenant=(Convert-TeamObservation $tenant[0]); user=(Convert-TeamObservation $user[0])}
    } else {
        $users = @()
        foreach ($value in @(Get-TeamUser -GroupId $request.id -ErrorAction Stop)) {
            $users += Convert-TeamUserObservation $value
        }
        $report = @{success=$true; stage='complete'; users=@($users)}
    }
} catch {
    $report = @{success=$false; stage=$report.stage}
} finally {
    if ($connected) { Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null }
    if ($null -ne $certificate) { $certificate.Dispose() }
}
[Console]::Out.Write(($report | ConvertTo-Json -Depth 8 -Compress))
