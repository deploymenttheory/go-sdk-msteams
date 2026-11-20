# Pure pinned environment maps/selectors. No override-file getter, auth or request methods.
[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$ModulePath,
    [Parameter(Mandatory)][string]$OutputPath
)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists; choose a new candidate path.'}
$assemblies=@(
    @{path='netcoreapp3.1/Microsoft.TeamsCmdlets.PowerShell.Connect.dll';sha256='c1c2b2d1c4a5306fb70f373df7ed5bfca4f9db58e90be358afd8a1d48ed46aea'},
    @{path='netcoreapp3.1/Microsoft.Teams.PowerShell.TeamsCmdlets.dll';sha256='3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24'}
)
foreach($assembly in $assemblies){
    if((Get-FileHash -LiteralPath (Join-Path $ModulePath $assembly.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $assembly.sha256){throw 'Profile assembly differs from the pinned baseline.'}
}
$moduleFile=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $moduleFile).ModuleVersion -ne '8.1.0'){throw 'Module version is not 8.1.0.'}
Import-Module $moduleFile -Force
$azureType=[Microsoft.TeamsCmdlets.PowerShell.Connect.Models.AzureEnvironment]
$teamsType=[Microsoft.TeamsCmdlets.PowerShell.Connect.Models.TeamsEnvironment]
$endpointType=[Microsoft.TeamsCmdlets.PowerShell.Connect.Common.Endpoint]
# Bypass PublicEnvironments: that getter permits a local OverrideEnvironment.config.
$maps=$azureType.GetField('environments',[Reflection.BindingFlags]'NonPublic,Static').GetValue($null)
$baseUri=[Microsoft.Teams.PowerShell.TeamsCmdlets.GetTenantPrivateChannelMigrationStatus].GetMethod('GetTemplatesServiceApiBaseUrl',[Reflection.BindingFlags]'NonPublic,Static')
$originalName=$teamsType::Name
$cases=[Collections.Generic.List[object]]::new()
try{
    foreach($cloud in @(
        @{name='commercial';pinned='TeamsCloud'},@{name='gcc';pinned='TeamsGCC'},
        @{name='gcch';pinned='TeamsGCCH'},@{name='dod';pinned='TeamsDOD'},
        @{name='china';pinned='TeamsChina'},@{name='france';pinned='TeamsFrance'},@{name='germany';pinned='TeamsGermany'}
    )){
        $nameRejected=$false
        try{$baseName=$teamsType::GetAzureEnvironmentName($cloud.pinned)}catch{
            if($cloud.name -ne 'gcc' -or $_.Exception.GetBaseException() -isnot [ArgumentException]){throw}
            # GCC is detected after commercial login. It has no entry in this
            # name mapper; its selector receives the existing AzureCloud tuple.
            $nameRejected=$true
            $baseName=[Microsoft.TeamsCmdlets.PowerShell.Connect.Models.AzureEnvironment+EnvironmentName]::AzureCloud
        }
        $template=$maps[$baseName]
        $endpoints=@{}
        foreach($endpoint in @('TeamsTargetingEndpointResourceId','TeamsTargetingTokenResourceId','TeamsHierarchyEndpointResourceId','TeamsHierarchyTokenResourceId','TeamsMigrationEndPointResourceId','TeamsMigrationTokenResourceId','TeamsAndChannelsServiceResourceId')){
            $endpoints[$endpoint]=$template.GetEndpoint([Enum]::Parse($endpointType,$endpoint))
        }
        $selectors=@{}
        foreach($selector in @(
            @{name='targeting';method='setTeamsTargetingEndPoint';endpoint='TeamsTargetingEndpointResourceId'},
            @{name='regions';method='setMigrationEndPoint';endpoint='TeamsMigrationEndPointResourceId'}
        )){
            $environment=$azureType::new();$environment.Name=$baseName
            foreach($entry in $template.Endpoints.GetEnumerator()){$environment.Endpoints.Add($entry.Key,$entry.Value)}
            $rejected=$false
            try{$teamsType.GetMethod($selector.method).Invoke($null,[object[]]@($cloud.pinned,$environment))}catch{
                if($_.Exception.GetBaseException() -isnot [ArgumentException]){throw}
                $rejected=$true
            }
            $selectors[$selector.name]=@{rejected=$rejected;origin_after_attempt=$environment.GetEndpoint([Enum]::Parse($endpointType,$selector.endpoint))}
        }
        $teamsType::Name=[Enum]::Parse([Microsoft.TeamsCmdlets.PowerShell.Connect.Models.TeamsEnvironment+EnvironmentName],$cloud.pinned)
        $setup=if($nameRejected){'source-supplied AzureCloud for post-login GCC selection'}else{'GetAzureEnvironmentName output'}
        $cases.Add(@{name=$cloud.name;teams_environment_input=$cloud.pinned;azure_environment=$baseName.ToString();azure_environment_setup=$setup;name_mapper_rejected=$nameRejected;base_endpoints=$endpoints;selectors=$selectors;private_channel_base_uri=$baseUri.Invoke($null,@())})
    }
}finally{$teamsType::Name=$originalName}
$artifact=[ordered]@{
    schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
    qualification='offline-profile-map-and-selection';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
    assemblies=$assemblies;purpose='Pure static maps, cloned environment selectors and private base-URI selector; no service availability, auth, token, HTTP or override-file claim.';cases=$cases.ToArray()
}
$artifact|ConvertTo-Json -Depth 30|Set-Content -LiteralPath $OutputPath -Encoding utf8
"profile map/selection observations: $($cases.Count)"
