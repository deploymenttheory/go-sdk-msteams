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
    if ($request.id -notmatch $uuid) { throw 'Explicit identifier required.' }
    if ($request.kind -notin @('team','users','channel-users','operation','shared-teams','associated-teams','shared-users','targeting-policy','hierarchy-status')) { throw 'Unknown oracle.' }
    $ownedRead = $request.kind -in @('team','users','channel-users')
    if ($ownedRead -and $request.fixture -notmatch '^msteams-sdk-[a-z2-7]{26,64}-(team|migration|membership)$') { throw 'Owned fixture required.' }
    if ($request.kind -eq 'operation' -and [string]::IsNullOrWhiteSpace([string]$request.operation_id)) { throw 'Explicit operation required.' }
    if ($request.kind -eq 'team' -and $request.owner -notmatch $uuid) { throw 'Explicit owner required.' }
    if ($request.kind -eq 'channel-users' -and [string]::IsNullOrWhiteSpace([string]$request.channel)) { throw 'Explicit channel required.' }
    if ($request.kind -in @('shared-teams','shared-users') -and [string]::IsNullOrWhiteSpace([string]$request.channel)) { throw 'Explicit shared channel required.' }
    if ($request.kind -eq 'shared-users' -and $request.shared_team_id -notmatch $uuid) { throw 'Explicit shared-with team required.' }
    if ($request.kind -eq 'shared-teams' -and -not [string]::IsNullOrEmpty([string]$request.shared_team_id) -and $request.shared_team_id -notmatch $uuid) { throw 'Invalid shared-with team.' }
    if ($request.kind -eq 'hierarchy-status' -and $request.api_version -notin @('1','2')) { throw 'Explicit hierarchy version required.' }
    $report.stage = 'module'
    $module = $env:MSTEAMS_LAB_MODULE_PATH
    $manifest = Join-Path $module 'MicrosoftTeams.psd1'
    if ((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0') { throw 'Wrong baseline.' }
    $assembly = Join-Path $module 'netcoreapp3.1/Microsoft.Teams.PowerShell.TeamsCmdlets.dll'
    if ((Get-FileHash -LiteralPath $assembly -Algorithm SHA256).Hash.ToLowerInvariant() -ne '3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24') { throw 'Wrong baseline assembly.' }
    Import-Module $manifest -Force -WarningAction SilentlyContinue
    $report.stage = 'authentication'
    $identity = Get-Content -LiteralPath $env:MSTEAMS_LAB_IDENTITY_FILE -Raw | ConvertFrom-Json
    if ($request.kind -in @('targeting-policy','hierarchy-status') -and [string]$request.id -ine [string]$identity.tenant_id) { throw 'Tenant binding differs.' }
    $certificate = [Security.Cryptography.X509Certificates.X509Certificate2]::CreateFromPemFile($identity.certificate_file, $identity.private_key_file)
    Connect-MicrosoftTeams -Certificate $certificate -ApplicationId $identity.client_id -TenantId $identity.tenant_id -ErrorAction Stop -WarningAction SilentlyContinue -InformationAction SilentlyContinue | Out-Null
    $connected = $true
    $report.stage = 'read'
    if ($ownedRead) {
        $values = @(Get-Team -GroupId $request.id -ErrorAction Stop)
        if ($values.Count -ne 1 -or [string]$values[0].GroupId -ine $request.id -or [string]$values[0].DisplayName -cne $request.fixture) { throw 'Expected one owned fixture.' }
    }
    function Convert-OperationObservation($value) {
        foreach ($property in @('OperationId','AsyncOperationType','AsyncOperationStatus','CreatedDateTime','LastUpdateDateTimeUtc','AttemptsCount','TargetResourceId','TargetResourceLocation')) {
            if ($null -eq $value.PSObject.Properties[$property]) { throw 'Incomplete operation projection.' }
        }
        return @{
            id=[string]$value.OperationId
            type=([string]$value.AsyncOperationType).ToLowerInvariant()
            status=([string]$value.AsyncOperationStatus).ToLowerInvariant()
            created=$value.CreatedDateTime.ToUniversalTime().ToString('O')
            updated=$value.LastUpdateDateTimeUtc.ToUniversalTime().ToString('O')
            attempts=$value.AttemptsCount
            target_id=$value.TargetResourceId
            target_location=$value.TargetResourceLocation
        }
    }
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
    function Convert-SharedTeamObservation($value, [bool]$includeHost) {
        foreach ($property in @('GroupId','DisplayName','TenantId')) {
            if ($null -eq $value.PSObject.Properties[$property]) { throw 'Incomplete associated/shared team projection.' }
        }
        $result = @{id=[string]$value.GroupId; display_name=$value.DisplayName; tenant_id=$value.TenantId; is_host=$null}
        if ($includeHost) {
            if ($null -eq $value.PSObject.Properties['IsHostTeam']) { throw 'Incomplete sharing projection.' }
            $result.is_host = $value.IsHostTeam
        }
        return $result
    }
    function Convert-HierarchyStatusObservation($value) {
        foreach ($property in @('Id','Status','LastKnownStatus','ErrorMessage','LastUpdatedAt','LastModifiedBy','FileName')) {
            if ($null -eq $value.PSObject.Properties[$property]) { throw 'Incomplete hierarchy status projection.' }
        }
        return @{id=[string]$value.Id; status=$value.Status; last_known_status=$value.LastKnownStatus; error_message=$value.ErrorMessage; last_updated_at=$value.LastUpdatedAt; last_modified_by=$value.LastModifiedBy; filename=$value.FileName}
    }
    function Convert-TargetingPolicyObservation($value) {
        if ($null -eq $value.PSObject.Properties['Identity']) { throw 'Incomplete targeting policy projection.' }
        $result = @{identity=[string]$value.Identity}
        $fields = @{
            CustomTagsMode='custom_tags_mode'; Description='description'; ManageTagsPermissionMode='manage_tags_permission_mode'
            ShiftBackedTagsMode='shift_backed_tags_mode'; SuggestedPresetTags='suggested_preset_tags'
            TeamOwnersEditWhoCanManageTagsMode='team_owners_edit_who_can_manage_tags_mode'
        }
        foreach ($name in $fields.Keys) {
            $property = $value.PSObject.Properties[$name]
            $result[$fields[$name]] = if ($null -eq $property -or $null -eq $property.Value) { $null } else { [string]$property.Value }
        }
        return $result
    }
    $report.stage = 'projection'
    if ($request.kind -eq 'hierarchy-status') {
        $observed = @(Get-TeamTargetingHierarchyStatus -ApiVersion $request.api_version -RequestId $request.request_id -ErrorAction Stop)
        if ($observed.Count -ne 1) { throw 'Expected one selected hierarchy operation.' }
        $report = @{success=$true; stage='complete'; hierarchy_status=(Convert-HierarchyStatusObservation $observed[0])}
    } elseif ($request.kind -eq 'targeting-policy') {
        $observed = @(Get-CsTeamsTargetingPolicy -Identity Global -ErrorAction Stop)
        if ($observed.Count -ne 1 -or [string]$observed[0].Identity -ine 'Global') { throw 'Expected Global targeting policy.' }
        $report = @{success=$true; stage='complete'; targeting_policy=(Convert-TargetingPolicyObservation $observed[0])}
    } elseif ($request.kind -eq 'operation') {
        $operations = @(Get-Operation -GroupId $request.id -OperationId $request.operation_id -ErrorAction Stop)
        if ($operations.Count -ne 1 -or [string]$operations[0].OperationId -cne $request.operation_id) { throw 'Expected one selected operation.' }
        $report = @{success=$true; stage='complete'; operation=(Convert-OperationObservation $operations[0])}
    } elseif ($request.kind -in @('shared-teams','associated-teams')) {
        $rows = @()
        $selected = @(
            if ($request.kind -eq 'associated-teams') {
                Get-AssociatedTeam -User $request.id -ErrorAction Stop
            } elseif ([string]::IsNullOrEmpty([string]$request.shared_team_id)) {
                Get-SharedWithTeam -HostTeamId $request.id -ChannelId $request.channel -ErrorAction Stop
            } else {
                Get-SharedWithTeam -HostTeamId $request.id -ChannelId $request.channel -SharedWithTeamId $request.shared_team_id -ErrorAction Stop
            }
        )
        foreach ($value in $selected) { $rows += Convert-SharedTeamObservation $value ($request.kind -eq 'shared-teams') }
        if ($request.kind -eq 'shared-teams') { $report = @{success=$true; stage='complete'; shared_teams=@($rows)} }
        else { $report = @{success=$true; stage='complete'; associated_teams=@($rows)} }
    } elseif ($request.kind -eq 'team') {
        $team = Convert-TeamObservation $values[0]
        $tenant = @(Get-Team -DisplayName $request.fixture -ErrorAction Stop)
        $user = @(Get-Team -User $request.owner -DisplayName $request.fixture -ErrorAction Stop)
        if ($tenant.Count -ne 1 -or $user.Count -ne 1) { throw 'Expected complete single-fixture lists.' }
        $report = @{success=$true; stage='complete'; team=$team; tenant=(Convert-TeamObservation $tenant[0]); user=(Convert-TeamObservation $user[0])}
    } else {
        $users = @()
        $members = @(
            if ($request.kind -eq 'shared-users') {
                Get-SharedWithTeamUser -HostTeamId $request.id -ChannelId $request.channel -SharedWithTeamId $request.shared_team_id -ErrorAction Stop
            } elseif ($request.kind -eq 'channel-users') {
                Get-TeamChannelUser -GroupId $request.id -DisplayName $request.channel -ErrorAction Stop
            } else {
                Get-TeamUser -GroupId $request.id -ErrorAction Stop
            }
        )
        foreach ($value in $members) {
            $row = Convert-TeamUserObservation $value
            if ($request.kind -eq 'shared-users') {
                if ($null -eq $value.PSObject.Properties['TenantId']) { throw 'Incomplete shared user projection.' }
                $row.tenant_id = $value.TenantId
            }
            $users += $row
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
