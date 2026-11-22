# Optional pinned read oracle. No policy/configuration or directory writes.
$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue'
$InformationPreference='SilentlyContinue'
$WarningPreference='SilentlyContinue'
$VerbosePreference='SilentlyContinue'
$DebugPreference='SilentlyContinue'
$report=@{success=$false};$connected=$false;$certificate=$null
try {
    if($env:MSTEAMS_LAB_READ_ONLY -ne '1' -or $env:MSTEAMS_LAB_POWERSHELL_ORACLE -ne '1'){throw 'Explicit read oracle opt-in required.'}
    $raw=[Console]::In.ReadToEnd();if($raw.Length -gt 8192){throw 'Invalid selectors.'}
    $request=$raw | ConvertFrom-Json
    if($request.kind -notin @('external-policy','acs','external-users','tenant')){throw 'Unknown selector.'}
    $selected=[string]$request.identity;if([string]::IsNullOrEmpty($selected)){$selected='Global'}
    $fixture=$selected -replace '^(?i)tag:',''
    if($selected -ine 'Global' -and ($request.kind -ne 'external-policy' -or $fixture -notmatch '^msteams-sdk-[a-z2-7]{26,64}-external-access$')){throw 'Explicit owned identity required.'}
    $module=$env:MSTEAMS_LAB_MODULE_PATH;$manifest=Join-Path $module 'MicrosoftTeams.psd1'
    if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
    foreach($item in @(
        @{path='bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll';hash='8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'},
        @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';hash='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
        @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';hash='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'}
    )){if((Get-FileHash -LiteralPath (Join-Path $module $item.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $item.hash){throw 'Wrong baseline assembly.'}}
    $identity=Get-Content -LiteralPath $env:MSTEAMS_LAB_IDENTITY_FILE -Raw | ConvertFrom-Json
    if([string]$request.tenant_id -ine [string]$identity.tenant_id -or [string]$request.tenant_id -notmatch '^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$'){throw 'Tenant binding differs.'}
    Import-Module $manifest -Force -WarningAction SilentlyContinue
    if($env:MSTEAMS_LAB_AUTH_METHOD -eq 'client-secret'){
        $authTenant=[Guid]::Parse($identity.tenant_id).ToString();$tokens=@()
        foreach($scope in @('https://graph.microsoft.com/.default','48ac35b8-9aa8-4d74-927d-1f4a14a0b239/.default')){
            $token=Invoke-RestMethod -Method Post -Uri "https://login.microsoftonline.com/$authTenant/oauth2/v2.0/token" -Body @{grant_type='client_credentials';client_id=$identity.client_id;client_secret=$identity.client_secret;scope=$scope}
            $tokens+= $token.access_token
        }
        Connect-MicrosoftTeams -AccessTokens $tokens -ErrorAction Stop -InformationAction SilentlyContinue -WarningAction SilentlyContinue | Out-Null
    }elseif([string]::IsNullOrEmpty($env:MSTEAMS_LAB_AUTH_METHOD) -or $env:MSTEAMS_LAB_AUTH_METHOD -eq 'certificate'){
        $certificate=[Security.Cryptography.X509Certificates.X509Certificate2]::CreateFromPemFile($identity.certificate_file,$identity.private_key_file)
        Connect-MicrosoftTeams -Certificate $certificate -ApplicationId $identity.client_id -TenantId $identity.tenant_id -WarningAction SilentlyContinue -InformationAction SilentlyContinue -ErrorAction Stop | Out-Null
    }else{throw 'Unknown auth method.'}
    $connected=$true
    $values=@(switch($request.kind){
        'external-policy'{Get-CsExternalAccessPolicy -Identity $selected -ErrorAction Stop}
        'acs'{Get-CsTeamsAcsFederationConfiguration -Identity Global -ErrorAction Stop}
        'external-users'{Get-CsTeamsExternalAccessConfiguration -Identity Global -ErrorAction Stop}
        'tenant'{Get-CsTenantFederationConfiguration -Identity Global -ErrorAction Stop}
    })
    if($values.Count -ne 1 -or (([string]$values[0].Identity -replace '^(?i)tag:','') -ine $fixture)){throw 'Expected selected observation.'}
    $fields=switch($request.kind){
        'external-policy'{@('Description','EnableFederationAccess')}
        'acs'{@('EnableAcsUsers','RequireAcsFederationForMeeting')}
        'external-users'{@('BlockExternalUserAccess')}
        'tenant'{@('AllowFederatedUsers','AllowTeamsConsumer')}
    }
    $settings=@{}
    foreach($name in $fields){$property=$values[0].PSObject.Properties[$name];if($null -eq $property){throw 'Incomplete selected projection.'};$settings[$name]=$property.Value}
    $report=@{success=$true;kind=$request.kind;identity=[string]$values[0].Identity;settings=$settings}
} catch {
    $report=@{success=$false}
} finally {
    if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null}
    if($null -ne $certificate){$certificate.Dispose()}
}
[Console]::Out.Write(($report | ConvertTo-Json -Depth 8 -Compress))
