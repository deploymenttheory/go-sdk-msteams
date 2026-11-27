[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference = 'Stop'
$moduleRoot = $ModulePath
$expected = @{
    'Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll' = '68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'
    'Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll' = 'a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'
}
foreach ($name in $expected.Keys) {
    if ((Get-FileHash (Join-Path $moduleRoot "netcoreapp3.1/$name")).Hash.ToLowerInvariant() -ne $expected[$name]) { throw 'Wrong pinned source assembly' }
}
Import-Module (Join-Path $moduleRoot 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
$references = @((Get-ChildItem (Join-Path $PSHOME 'ref/*.dll')).FullName) + @([System.Management.Automation.PSCmdlet].Assembly.Location)
$fakePath = Join-Path $PSScriptRoot 'pol91-field-binding-fakes.cs'
$types = Add-Type -Path $fakePath -ReferencedAssemblies $references -PassThru
$probeType = $types | Where-Object FullName -eq 'Pol91FieldBinding.Probe'
$probe = [System.Management.Automation.CmdletInfo]::new('Test-Pol91FieldBinding', $probeType)
$resolver = [AppDomain]::CurrentDomain.GetAssemblies() | ForEach-Object {
    $_.GetType('Microsoft.Teams.Policy.Administration.Cmdlets.Providers.Helpers.BoundParametersContractResolver', $false)
} | Where-Object { $null -ne $_ } | Select-Object -First 1
if ($null -eq $resolver) { throw 'Pinned resolver missing' }
$cases = @()
foreach ($name in @('RestrictedSenderList', 'ContentPin', 'ResourceAccountContentAccess', 'UseUnifiedDomain')) {
    $cases += @{name="client-$name-null"; command='Set-CsTeamsClientConfiguration'; field=$name; value=$null}
}
$cases += @(
    @{name='client-string-empty'; command='Set-CsTeamsClientConfiguration'; field='RestrictedSenderList'; value=''},
    @{name='client-string-space'; command='Set-CsTeamsClientConfiguration'; field='RestrictedSenderList'; value=' '},
    @{name='client-bool-null'; command='Set-CsTeamsClientConfiguration'; field='AllowEmailIntoChannel'; value=$null},
    @{name='client-bool-false'; command='Set-CsTeamsClientConfiguration'; field='AllowEmailIntoChannel'; value=$false},
    @{name='validation-uri-null'; command='Set-CsOnlineVoicemailValidationConfiguration'; field='AudioFileValidationUri'; value=$null},
    @{name='validation-uri-empty'; command='Set-CsOnlineVoicemailValidationConfiguration'; field='AudioFileValidationUri'; value=''},
    @{name='validation-uri-https'; command='Set-CsOnlineVoicemailValidationConfiguration'; field='AudioFileValidationUri'; value='https://example.invalid/fixture'},
    @{name='validation-bool-null'; command='Set-CsOnlineVoicemailValidationConfiguration'; field='AudioFileValidationEnabled'; value=$null},
    @{name='validation-bool-false'; command='Set-CsOnlineVoicemailValidationConfiguration'; field='AudioFileValidationEnabled'; value=$false}
)
$rows = foreach ($case in $cases) {
    $public = Get-Command $case.command
    $sourceParameter = $public.Parameters[$case.field]
    if ($sourceParameter.ParameterType -ne $probeType.GetProperty($case.field).PropertyType) { throw 'Inert and public parameter types differ' }
    $extraAttributes = @($sourceParameter.Attributes | Where-Object { $_ -isnot [System.Management.Automation.ParameterAttribute] })
    if ($extraAttributes.Count -ne 0) { throw 'Source field has additional binding attributes requiring a more exact probe' }
    $arguments = @{Identity='Global'}
    $arguments[$case.field] = $case.value
    $bindingForm = 'hashtable-splat'
    switch ($case.name) {
        'client-RestrictedSenderList-null' {
            $bindingForm = 'literal-null'
            $observation = & $probe -Identity Global -RestrictedSenderList $null
        }
        'client-bool-null' {
            $bindingForm = 'literal-null'
            $observation = & $probe -Identity Global -AllowEmailIntoChannel $null
        }
        'client-bool-false' {
            $bindingForm = 'literal-false'
            $observation = & $probe -Identity Global -AllowEmailIntoChannel $false
        }
        'validation-uri-null' {
            $bindingForm = 'literal-null'
            $observation = & $probe -Identity Global -AudioFileValidationUri $null
        }
        default { $observation = & $probe @arguments }
    }
    $dtoType = $public.ImplementingType.BaseType.GetGenericArguments()[0]
    $dto = [Activator]::CreateInstance($dtoType)
    $boundValue = $observation.PropertyValues[$case.field]
    $setterError = $null
    $serialized = $null
    try {
        $dtoType.GetProperty($case.field).SetValue($dto, $boundValue)
        $bound = [Collections.Generic.Dictionary[string,object]]::new()
        $bound.Add($case.field, $observation.BoundValues[$case.field])
        $settings = [Newtonsoft.Json.JsonSerializerSettings]::new()
        $settings.ContractResolver = [Activator]::CreateInstance($resolver, [object[]]@($bound))
        $serialized = [Newtonsoft.Json.JsonConvert]::SerializeObject($dto, $settings)
    } catch {
        $failure = $_.Exception
        while ($failure.InnerException) { $failure = $failure.InnerException }
        $setterError = @{type=$failure.GetType().FullName; message=$failure.Message}
    }
    @{
        case=$case.name; command=$case.command; field=$case.field; binding_form=$bindingForm
        input_is_null=($null -eq $case.value); input=$case.value
        bound_is_null=($null -eq $observation.BoundValues[$case.field]); bound=$observation.BoundValues[$case.field]
        property_is_null=($null -eq $boundValue); property=$boundValue
        source_parameter_type=$sourceParameter.ParameterType.FullName
        source_model_type=$dtoType.FullName
        serialized=$serialized; setter_error=$setterError
    }
}
@{
    scope='Actual inert compiled-cmdlet parameter binding with source-matching types/attributes, then actual pinned DTO setter/resolver serialization. Only inert ProcessRecord runs; no management lifecycle, credentials, telemetry or HTTP.'
    hashes=$expected
    reproducer_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant()
    fake_sha256=(Get-FileHash $fakePath).Hash.ToLowerInvariant()
    observations=@($rows)
} | ConvertTo-Json -Depth 9 -Compress
