# Inspect static command metadata only; never construct or invoke an OCE cmdlet.
[CmdletBinding()]
param(
 [Parameter(Mandatory)][string]$ModulePath,
 [Parameter(Mandatory)][string]$OutputPath,
 [Parameter(Mandatory)][ValidateSet('net472','netcoreapp3.1')][string]$Target
)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$relative="$Target/Microsoft.Teams.Policy.Administration.Cmdlets.OCE.dll"
$hashes=@{
 net472='ea52b14ae69bfeb7ce63497176a81252dd7c0fa17ad93d6b8ae49ef358091f03'
 'netcoreapp3.1'='de6da2b13b4b675d7956a0c57af53fccc6cf457a3de2e8464273f95103e3dfc5'
}
$file=Join-Path $ModulePath $relative
if((Get-FileHash -LiteralPath $file -Algorithm SHA256).Hash.ToLowerInvariant() -ne $hashes[$Target]){throw 'Wrong pinned OCE assembly.'}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
# Import the ordinary unauthenticated dependency profile, even if the caller has
# an OCE environment marker. This script must run in its own fresh pwsh process.
$env:MSTeamsContextInternal=$null
Import-Module $manifest -Force
$assembly=[Reflection.Assembly]::LoadFrom($file)
if([IO.Path]::GetFullPath($assembly.Location) -ne [IO.Path]::GetFullPath($file)){throw 'Assembly identity was already loaded from another path; use a fresh process.'}
$namespace='Microsoft.Teams.Policy.Administration.Cmdlets.OCE.'
$names=@(
 'Policies.LROS.Operations.LROSOperations+GetBatchOperationDefinitionCmdlet',
 'Policies.LROS.Operations.LROSOperations+GetBatchOperationStatusCmdlet',
 'Policies.LROS.Operations.LROSOperations+ReprocessBatchOperationCmdlet',
 'Policies.GPA.Operations.GPAOperations+GetGroupPolicyAssignmentsCmdlet',
 'Policies.GetGPAGroupMembers+GPAGroupMembers',
 'Policies.GetGPAUserMembership+GPAUserMembership',
 'Policies.GPARefreshGroupUsers+RefreshGroupUsers',
 'Policies.GPAReprocessOperations+UpdateGPAOperationsCmdlet'
)
$common=@('Verbose','Debug','ErrorAction','WarningAction','InformationAction','ProgressAction','ErrorVariable','WarningVariable','InformationVariable','OutVariable','OutBuffer','PipelineVariable')
$commands=@(foreach($name in $names){
 $type=$assembly.GetType($namespace+$name,$true)
 $attribute=@($type.GetCustomAttributes([Management.Automation.CmdletAttribute],$false))
 if($attribute.Count -ne 1){throw 'Expected exactly one cmdlet attribute.'}
 $commandName=$attribute[0].VerbName+'-'+$attribute[0].NounName
 $info=[Management.Automation.CmdletInfo]::new($commandName,$type)
 $sets=@(foreach($set in ($info.ParameterSets|Sort-Object Name)){
  $parameters=@(foreach($p in ($set.Parameters|Sort-Object Name)){
   if($p.Name -in $common){continue}
   $enumType=[Nullable]::GetUnderlyingType($p.ParameterType)
   if($null -eq $enumType){$enumType=$p.ParameterType}
   $enumValues=@();if($enumType.IsEnum){$enumValues=@([Enum]::GetNames($enumType))}
   [ordered]@{name=$p.Name;type=$p.ParameterType.FullName;mandatory=$p.IsMandatory;position=$p.Position;aliases=@($p.Aliases|Sort-Object);value_from_pipeline=$p.ValueFromPipeline;value_from_pipeline_by_property_name=$p.ValueFromPipelineByPropertyName;enum_values=$enumValues}
  })
  [ordered]@{name=$set.Name;is_default=$set.IsDefault;parameters=$parameters}
 })
 [ordered]@{name=$commandName;declaring_type=$type.FullName;parameter_sets=$sets}
})
if($commands.Count -ne 8){throw 'Incomplete command metadata.'}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-policy-assignment-oce-static-metadata'
 package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assembly=@{path=$relative;sha256=$hashes[$Target]}
 host_powershell_version=$PSVersionTable.PSVersion.ToString()
 stage='Static CmdletInfo/ParameterSetInfo reflection on the pinned OCE assembly after ordinary unauthenticated dependency import. No OCE profile activation, command construction, ProcessRecord, credentials, sender, discovery or network. Common PowerShell parameters are omitted from this supplemental signature inventory.'
 commands=@($commands|Sort-Object name)
}|ConvertTo-Json -Depth 25|Set-Content -LiteralPath $OutputPath -Encoding utf8
"policy assignment OCE metadata observations ($Target): $($commands.Count)"
