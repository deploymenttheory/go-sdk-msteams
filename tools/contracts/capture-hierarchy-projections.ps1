param(
    [Parameter(Mandatory=$true)][string]$ModulePath,
    [Parameter(Mandatory=$true)][string]$OutputPath
)
# Offline model constructors only: no module import, connection, cmdlet or HTTP.
$ErrorActionPreference = 'Stop'
$assemblyPath = Join-Path $ModulePath 'netcoreapp3.1/Microsoft.Teams.PowerShell.TeamsCmdlets.dll'
$hash = '3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24'
if ((Get-FileHash -LiteralPath $assemblyPath -Algorithm SHA256).Hash.ToLowerInvariant() -cne $hash) { throw 'Wrong baseline assembly.' }
if ((Import-PowerShellDataFile -LiteralPath (Join-Path $ModulePath 'MicrosoftTeams.psd1')).ModuleVersion -ne '8.1.0') { throw 'Wrong baseline version.' }
$assembly = [Reflection.Assembly]::LoadFrom($assemblyPath)
$resultType = $assembly.GetType('Microsoft.Teams.PowerShell.TeamsCmdlets.GetTeamTargetingHierarchyStatus+Result',$true)
$specs = @(
    @{name='retail-ordered-error-arguments'; version='1'; input='{"id":"fixture-operation","status":"Failed","lastUpdatedAt":"wire-time","errorDetails":{"lastKnownStatus":"Ingesting","errorMessage":"[{\"errorCode\":\"First\",\"description\":\"bad {first}/{second}\",\"args\":[{\"second\":\"B\",\"first\":\"A\"}]},{\"errorCode\":\"First\",\"description\":\"extra {x}\",\"args\":[{\"x\":\"X\"}]}]"}}'},
    @{name='retail-empty-error-text'; version='1'; input='{"id":"fixture-operation","status":"Ingesting","errorDetails":{"lastKnownStatus":"Starting","errorMessage":""}}'},
    @{name='hierarchy-grouped-ordered-errors'; version='2'; input='{"id":"fixture-operation","status":"Failed","error":{"message":"native top-level message","details":[{"code":"A","message":"bad {a}/{b}","arguments":{"{a}":"{b}","{b}":"B"}},{"code":"B","message":"other","arguments":{}},{"code":"A","message":"last {x}","arguments":{"{x}":"X"}}]}}'},
    @{name='hierarchy-no-error-details'; version='2'; input='{"id":"fixture-operation","status":"Succeeded","error":{"message":"native top-level message","details":[]}}'},
    @{name='hierarchy-numeric-status'; version='2'; input='{"id":"fixture-operation","status":"1"}'},
    @{name='hierarchy-status-case'; version='2'; input='{"id":"fixture-operation","status":"running"}'},
    @{name='hierarchy-combined-enum-status'; version='2'; input='{"id":"fixture-operation","status":"Running, Succeeded"}'},
    @{name='hierarchy-status-whitespace'; version='2'; input='{"id":"fixture-operation","status":"  Running  "}'}
)
$cases = @()
foreach ($spec in $specs) {
    $modelName = if ($spec.version -eq '1') { 'RetailServiceOperation' } else { 'HierarchyServiceOperation' }
    $type = $assembly.GetType('Microsoft.Teams.PowerShell.TeamsCmdlets.Model.'+$modelName,$true)
    $native = [Newtonsoft.Json.JsonConvert]::DeserializeObject([string]$spec.input,$type)
    $projection = [Activator]::CreateInstance($resultType,[object[]]@($native))
    $cases += @{
        name=$spec.name; version=$spec.version; input=$spec.input
        projection=@{
            id=$projection.Id; status=$projection.Status
            last_known_status=$projection.LastKnownStatus; error_message=$projection.ErrorMessage
            last_updated_at=$projection.LastUpdatedAt; last_modified_by=$projection.LastModifiedBy; filename=$projection.FileName
        }
    }
}
$report = @{
    schema_version=1; baseline_version='8.1.0'; assembly_sha256=$hash
    source='GetTeamTargetingHierarchyStatus.Result(RetailServiceOperation/HierarchyServiceOperation)'
    observation='offline shipped CLR constructor; no exported cmdlet, tenant or HTTP'
    source_newline=[Environment]::NewLine; live='not-run'; cases=@($cases)
}
$json = $report | ConvertTo-Json -Depth 20
[IO.File]::WriteAllText($OutputPath,$json+"`n",[Text.UTF8Encoding]::new($false))
[Console]::Out.WriteLine('Pinned hierarchy projection capture: '+$cases.Count+' synthetic cases; no tenant access.')
