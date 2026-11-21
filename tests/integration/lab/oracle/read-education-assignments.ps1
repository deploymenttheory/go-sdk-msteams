# Optional pinned compiled education assignments policy Global read; no tenant writes.
$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue';$InformationPreference='SilentlyContinue'
$WarningPreference='SilentlyContinue';$VerbosePreference='SilentlyContinue';$DebugPreference='SilentlyContinue'
$report=@{success=$false};$connected=$false;$certificate=$null
try {
 if($env:MSTEAMS_LAB_READ_ONLY -ne '1' -or $env:MSTEAMS_LAB_POWERSHELL_ORACLE -ne '1'){throw 'Explicit read oracle opt-in required.'}
 $request=[Console]::In.ReadToEnd() | ConvertFrom-Json;$tenant=[Guid]::Parse($request.tenant_id).ToString()
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
 $certificate=[Security.Cryptography.X509Certificates.X509Certificate2]::CreateFromPemFile($identity.certificate_file,$identity.private_key_file)
 Connect-MicrosoftTeams -Certificate $certificate -ApplicationId $identity.client_id -TenantId $identity.tenant_id -ErrorAction Stop -InformationAction SilentlyContinue -WarningAction SilentlyContinue | Out-Null
 $connected=$true;$values=@(Get-CsTeamsEducationAssignmentsAppPolicy -Identity Global -ErrorAction Stop)
 if($values.Count -ne 1 -or [string]$values[0].Identity -ine 'Global'){throw 'Expected Global policy.'}
 $settings=@{}
 foreach($name in @('ParentDigestEnabledType','MakeCodeEnabledType','TurnItInEnabledType','TurnItInApiUrl','TurnItInApiKey')){
  $property=$values[0].PSObject.Properties[$name];if($null -eq $property){throw 'Incomplete education assignments policy projection.'};$settings[$name]=$property.Value
 }
 $report=@{success=$true;tenant_id=$tenant;identity=[string]$values[0].Identity;settings=$settings}
} catch {$report=@{success=$false}}
finally{if($connected){Disconnect-MicrosoftTeams -Confirm:$false -ErrorAction SilentlyContinue | Out-Null};if($null -ne $certificate){$certificate.Dispose()}}
[Console]::Out.Write(($report | ConvertTo-Json -Depth 12 -Compress))
