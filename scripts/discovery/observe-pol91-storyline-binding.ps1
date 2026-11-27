[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference = 'Stop'
$hashes = @{
    'Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll' = '68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'
    'Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll' = 'a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'
}
foreach ($name in $hashes.Keys) {
    if ((Get-FileHash (Join-Path $ModulePath "netcoreapp3.1/$name")).Hash.ToLowerInvariant() -ne $hashes[$name]) { throw 'Pinned source hash differs' }
}
Import-Module (Join-Path $ModulePath 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
$public = Get-Command Set-CsTeamsMessagingConfiguration
$parameter = $public.Parameters['Storyline']
if (@($parameter.Attributes | Where-Object { $_ -isnot [System.Management.Automation.ParameterAttribute] }).Count -ne 0) { throw 'Additional binder attributes need review' }
$enum = $parameter.ParameterType.GetGenericArguments()[0]
if (-not $enum.IsEnum -or $enum.GetEnumUnderlyingType() -ne [int]) { throw 'Source enum shape changed' }
$references = @((Get-ChildItem (Join-Path $PSHOME 'ref/*.dll')).FullName) + @([System.Management.Automation.PSCmdlet].Assembly.Location)
# Close the inert type at runtime with the actual source enum. No Core compile
# reference/version-warning suppression or management ProcessRecord is involved.
$types = Add-Type -PassThru -ReferencedAssemblies $references -TypeDefinition @'
using System.Management.Automation;
namespace Pol91StorylineBinding {
[Cmdlet("Test", "Pol91StorylineBinding")]
public sealed class Probe<T> : PSCmdlet where T : struct {
    [Parameter] public T? Storyline { get; set; }
    protected override void ProcessRecord() {
        WriteObject(new { Bound=MyInvocation.BoundParameters.ContainsKey("Storyline"), Value=Storyline });
    }
}
}
'@
$generic = $types | Where-Object FullName -eq 'Pol91StorylineBinding.Probe`1'
$closed = $generic.MakeGenericType($enum)
if ($closed.GetProperty('Storyline').PropertyType -ne $parameter.ParameterType) { throw 'Inert property type differs' }
$probe = [System.Management.Automation.CmdletInfo]::new('Test-Pol91StorylineBinding', $closed)
$model = $public.ImplementingType.BaseType.GetGenericArguments()[0]
$resolver = [AppDomain]::CurrentDomain.GetAssemblies() | ForEach-Object {
    $_.GetType('Microsoft.Teams.Policy.Administration.Cmdlets.Providers.Helpers.BoundParametersContractResolver', $false)
} | Where-Object { $null -ne $_ } | Select-Object -First 1
$cases = @(
    @{name='null';value=$null;accept=$true;wire='{"Storyline":null}'},
    @{name='disabled-name';value='Disabled';accept=$true;wire='{"Storyline":0}'},
    @{name='enabled-name';value='Enabled';accept=$true;wire='{"Storyline":1}'},
    @{name='untyped-unknown-number';value=37;accept=$false},
    @{name='untyped-unknown-numeric-string';value='-7';accept=$false},
    @{name='typed-unknown-enum';value=[Enum]::ToObject($enum,37);accept=$true;wire='{"Storyline":37}'}
)
$observations = foreach ($case in $cases) {
    $observation = $null
    $failure = $null
    try { $observation = & $probe -Storyline $case.value } catch { $failure = $_.FullyQualifiedErrorId }
    if (($null -eq $failure) -ne $case.accept) { throw 'Compiled binder outcome changed' }
    $wire = $null
    if ($null -eq $failure) {
        if (-not $observation.Bound) { throw 'Selected key lost' }
        $dto = [Activator]::CreateInstance($model)
        $model.GetProperty('Storyline').SetValue($dto, $observation.Value)
        $bound = [Collections.Generic.Dictionary[string,object]]::new()
        $bound.Add('Storyline', $observation.Value)
        $settings = [Newtonsoft.Json.JsonSerializerSettings]::new()
        $settings.ContractResolver = [Activator]::CreateInstance($resolver, [object[]]@($bound))
        $wire = [Newtonsoft.Json.JsonConvert]::SerializeObject($dto, $settings)
        if ($wire -ne $case.wire) { throw 'Actual DTO/resolver projection changed' }
    }
    [pscustomobject]@{name=$case.name;accepted=($null -eq $failure);serialized=$wire;error=$failure}
}
[pscustomobject]@{
    hashes=$hashes
    reproducer_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant()
    source_parameter=$parameter.ParameterType.FullName
    enum_members=@($enum.GetEnumNames() | ForEach-Object { [pscustomobject]@{name=$_;value=[Convert]::ToInt32([Enum]::Parse($enum,$_))} })
    observations=@($observations)
    management_lifecycle_invoked=$false
    authentication_or_http=$false
} | ConvertTo-Json -Depth 6
