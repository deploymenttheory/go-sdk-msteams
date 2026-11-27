[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$ModulePath,
    [ValidateSet('TenantAbuse','Messaging')][string]$Family = 'TenantAbuse'
)
$ErrorActionPreference = 'Stop'
$expected = @{
    'Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll' = '68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'
    'Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll' = 'a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'
}
foreach ($name in $expected.Keys) {
    if ((Get-FileHash (Join-Path $ModulePath "netcoreapp3.1/$name")).Hash.ToLowerInvariant() -ne $expected[$name]) { throw 'Pinned assembly mismatch' }
}
Import-Module (Join-Path $ModulePath 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
$field = if ($Family -eq 'TenantAbuse') { 'BlockedTenants' } else { 'FileTypeBlockList' }
$public = Get-Command ("Set-CsTeams" + $Family + "Configuration")
$parameter = $public.Parameters[$field]
if ($parameter.ParameterType -ne [System.Management.Automation.PSListModifier[string]]) { throw 'Public list parameter type changed' }
if (@($parameter.Attributes | Where-Object { $_ -isnot [System.Management.Automation.ParameterAttribute] }).Count -ne 0) { throw 'Additional public binder attributes need review' }
$references = @((Get-ChildItem (Join-Path $PSHOME 'ref/*.dll')).FullName) + @([System.Management.Automation.PSCmdlet].Assembly.Location)
$types = Add-Type -ReferencedAssemblies $references -PassThru -TypeDefinition @'
using System.Management.Automation;
namespace Pol91TenantAbuseList {
[Cmdlet("Test", "Pol91TenantAbuseList")]
public sealed class Probe : PSCmdlet {
    [Parameter] public PSListModifier<string> BlockedTenants { get; set; }
    [Parameter] public PSListModifier<string> FileTypeBlockList { get; set; }
    protected override void ProcessRecord() {
        bool blocked = MyInvocation.BoundParameters.ContainsKey("BlockedTenants");
        WriteObject(new { Bound = blocked || MyInvocation.BoundParameters.ContainsKey("FileTypeBlockList"), Value = blocked ? BlockedTenants : FileTypeBlockList });
    }
}
}
'@
$probe = [System.Management.Automation.CmdletInfo]::new('Test-Pol91TenantAbuseList', ($types | Where-Object FullName -eq 'Pol91TenantAbuseList.Probe'))
# The inert compiled binder records only; the real management lifecycle is never invoked.
$bound = if ($Family -eq 'TenantAbuse') {
    & $probe -BlockedTenants @($null, 'entry')
} else {
    & $probe -FileTypeBlockList @($null, 'entry')
}
if (-not $bound.Bound -or $null -eq $bound.Value) { throw 'Replacement binding lost' }
$model = $public.ImplementingType.BaseType.GetGenericArguments()[0]
$dto = [Activator]::CreateInstance($model)
$list = $model.GetProperty($field).GetValue($dto)
$bound.Value.ApplyTo([Collections.IList]$list)
$resolverType = [AppDomain]::CurrentDomain.GetAssemblies() | ForEach-Object {
    $_.GetType('Microsoft.Teams.Policy.Administration.Cmdlets.Providers.Helpers.BoundParametersContractResolver', $false)
} | Where-Object { $null -ne $_ } | Select-Object -First 1
$keys = [Collections.Generic.Dictionary[string,object]]::new()
$keys.Add($field, $bound.Value)
$settings = [Newtonsoft.Json.JsonSerializerSettings]::new()
$settings.ContractResolver = [Activator]::CreateInstance($resolverType, [object[]]@($keys))
$serialized = [Newtonsoft.Json.JsonConvert]::SerializeObject($dto, $settings)
if ($serialized -ne ('{"' + $field + '":[null,"entry"]}')) { throw 'Pinned nullable replacement result changed' }
$decoded = [Newtonsoft.Json.JsonConvert]::DeserializeObject($serialized, $model)
$decodedList = $model.GetProperty($field).GetValue($decoded)
if ($decodedList.Count -ne 2 -or $null -ne $decodedList[0]) { throw 'Pinned nullable read projection changed' }
$delta = [System.Management.Automation.PSListModifier[string]]::new(@{Remove=@($null,'entry'); Add=@($null,'ENTRY')})
$current = [Collections.Generic.List[string]]::new()
$initial = [System.Management.Automation.PSListModifier[string]]::new([object[]]@('entry',$null,'entry',$null))
$initial.ApplyTo([Collections.IList]$current)
$delta.ApplyTo([Collections.IList]$current)
if ([Newtonsoft.Json.JsonConvert]::SerializeObject($current) -ne '["entry",null,null,"ENTRY"]') { throw 'Pinned delta result changed' }
[pscustomobject]@{
    source_hashes = $expected
    reproducer_sha256 = (Get-FileHash $PSCommandPath).Hash.ToLowerInvariant()
    source_command = $public.Name
    source_field = $field
    source_parameter = $parameter.ParameterType.FullName
    source_model = $model.FullName
    binding_form = 'inert compiled binder, literal array with null member'
    serialized_replacement = $serialized
    deserialized_read_member_is_null = ($null -eq $decodedList[0])
    delta_after_remove_first_and_add = @($current)
    management_lifecycle_invoked = $false
    authentication_or_http = $false
} | ConvertTo-Json -Depth 6
