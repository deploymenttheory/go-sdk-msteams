[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference = 'Stop'
$hashes = @{
    'Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll' = '68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'
    'Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll' = 'a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'
}
foreach ($name in $hashes.Keys) {
    if ((Get-FileHash (Join-Path $ModulePath "netcoreapp3.1/$name")).Hash.ToLowerInvariant() -ne $hashes[$name]) { throw 'Pinned source differs' }
}
Import-Module (Join-Path $ModulePath 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
$public = Get-Command Set-CsTeamsAudioConferencingCustomPromptsConfiguration
$model = $public.ImplementingType.BaseType.GetGenericArguments()[0]
$get = Get-Command Get-CsTeamsAudioConferencingCustomPromptsConfiguration
$normalizer = [AppDomain]::CurrentDomain.GetAssemblies() | ForEach-Object {
    $_.GetType('Microsoft.Teams.Policy.Administration.Cmdlets.Providers.Helpers.ConfigurationNameNormalization`1', $false)
} | Where-Object { $null -ne $_ } | Select-Object -First 1
$normalized = $normalizer.MakeGenericType($model).GetMethod('GetNormalizedNameForPolicyRP',[Reflection.BindingFlags]'Public,NonPublic,Static').Invoke($null,@())
$helper = $model.Assembly.GetType('Microsoft.Teams.Policy.Administration.Cmdlets.Core.PSListModifierHelper')
$apply = $helper.GetMethods([Reflection.BindingFlags]'Public,NonPublic,Static') | Where-Object { $_.Name -eq 'ApplyTo' -and $_.IsGenericMethodDefinition -and $_.GetParameters().Count -eq 3 } | Select-Object -First 1
$closure = $public.ImplementingType.GetNestedType('<>c', [Reflection.BindingFlags]'NonPublic')
$singleton = $closure.GetField('<>9', [Reflection.BindingFlags]'Public,NonPublic,Static').GetValue($null)
$resolver = [AppDomain]::CurrentDomain.GetAssemblies() | ForEach-Object {
    $_.GetType('Microsoft.Teams.Policy.Administration.Cmdlets.Providers.Helpers.BoundParametersContractResolver', $false)
} | Where-Object { $null -ne $_ } | Select-Object -First 1
$requestHelperType = [AppDomain]::CurrentDomain.GetAssemblies() | ForEach-Object {
    $_.GetType('Microsoft.Teams.Policy.Administration.Cmdlets.Providers.HttpRequestHelper`1', $false)
} | Where-Object { $null -ne $_ } | Select-Object -First 1
$nested = $requestHelperType.MakeGenericType($model).GetMethod('GetNestedBoundParameters',[Reflection.BindingFlags]'Public,NonPublic,Static')
$observations = foreach ($field in @('Prompts','Packages')) {
    $modifierType = $public.Parameters[$field].ParameterType
    if (@($public.Parameters[$field].Attributes | Where-Object { $_ -isnot [System.Management.Automation.ParameterAttribute] }).Count -ne 0) { throw 'Additional field attributes need review' }
    $itemType = $modifierType.GetGenericArguments()[0]
    $method = $closure.GetMethod($(if ($field -eq 'Prompts') { '<UpdateProperties>b__2_0' } else { '<UpdateProperties>b__2_1' }), [Reflection.BindingFlags]'Public,NonPublic,Instance')
    $funcType = [Func[object,object,bool]].GetGenericTypeDefinition().MakeGenericType([Type[]]@($itemType,$itemType,[bool]))
    $predicate = $method.CreateDelegate($funcType,$singleton)
    $closed = $apply.MakeGenericMethod($itemType)
    $ctor = $modifierType.GetConstructor([Type[]]@([hashtable]))
    foreach ($case in @('empty-clears-populated','null-modifier-empty','null-modifier-populated','null-remove-empty','null-remove-populated','null-add','null-replace','replace-entry','compare-remove-empty','compare-remove-populated')) {
        $dto = [Activator]::CreateInstance($model)
        $model.GetProperty('Identity').SetValue($dto,'Global')
        $list = $model.GetProperty($field).GetValue($dto)
        $arguments = if ($field -eq 'Prompts') { [object[]]@('00000000-0000-0000-0000-000000000001','FutureType','xx-future') } else { [object[]]@('00000000-0000-0000-0000-000000000002','xx-future') }
        $entry = [Activator]::CreateInstance($itemType,$arguments)
        if ($case -in @('empty-clears-populated','null-modifier-populated','null-remove-populated','compare-remove-populated')) { $list.Add($entry) }
        $modifier = switch ($case) {
            { $_ -in @('null-modifier-empty','null-modifier-populated') } { $null }
            { $_ -in @('null-remove-empty','null-remove-populated') } { $ctor.Invoke([object[]]@(@{Remove=@($null)})) }
            'null-add' { $ctor.Invoke([object[]]@(@{Add=@($null)})) }
            'null-replace' { $ctor.Invoke([object[]]@(@{Replace=@($null)})) }
            'replace-entry' { $ctor.Invoke([object[]]@(@{Replace=@($entry)})) }
            { $_ -in @('compare-remove-empty','compare-remove-populated') } { $ctor.Invoke([object[]]@(@{Remove=@($entry)})) }
            default { [Activator]::CreateInstance($modifierType) }
        }
        $failure = $null
        $body = $null
        $stage = 'helper'
        try {
            $null = $closed.Invoke($null,[object[]]@($modifier,$list,$predicate))
            $stage = 'serializer'
            $bound = [Collections.Generic.Dictionary[string,object]]::new()
            $bound.Add('Identity','Global'); $bound.Add($field,$modifier)
            $bound = $nested.Invoke($null,[object[]]@($bound,$model,$false))
            $settings = [Newtonsoft.Json.JsonSerializerSettings]::new()
            $settings.ContractResolver = [Activator]::CreateInstance($resolver,[object[]]@($bound))
            $body = [Newtonsoft.Json.JsonConvert]::SerializeObject($dto,$settings)
            $stage = 'complete'
        } catch { $failure = $_.Exception.GetBaseException().GetType().FullName }
        [pscustomobject]@{field=$field;case=$case;remaining=$list.Count;stage=$stage;error=$failure;serialized=$body}
    }
}
# Source-authored DTO projection for successful SDK read tests. It is not a
# captured service response, and no management lifecycle/authentication/HTTP runs.
$read = [Activator]::CreateInstance($model)
$model.GetProperty('Identity').SetValue($read,'Global')
$promptType = $public.Parameters['Prompts'].ParameterType.GetGenericArguments()[0]
$prompt = [Activator]::CreateInstance($promptType,[object[]]@('00000000-0000-0000-0000-000000000001','FutureType','xx-future'))
$promptType.GetProperty('Name').SetValue($prompt,'example')
$promptType.GetProperty('TextPrompt').SetValue($prompt,'authored text')
$model.GetProperty('Prompts').GetValue($read).Add($prompt)
$packageType = $public.Parameters['Packages'].ParameterType.GetGenericArguments()[0]
$package = [Activator]::CreateInstance($packageType,[object[]]@('00000000-0000-0000-0000-000000000002','xx-future'))
$packageType.GetProperty('Name').SetValue($package,'example package')
foreach ($name in @('InboundStartRecordingPrompt','InboundEndRecordingPrompt','OutboundStartRecordingPrompt','OutboundEndRecordingPrompt','MeetingStartRecordingPrompt','MeetingEndRecordingPrompt')) {
    $values = $packageType.GetProperty($name).GetValue($package)
    $strings = [System.Management.Automation.PSListModifier[string]]::new([object[]]@('00000000-0000-0000-0000-000000000001',$null))
    $strings.ApplyTo([Collections.IList]$values)
}
$model.GetProperty('Packages').GetValue($read).Add($package)
$selected = [Collections.Generic.Dictionary[string,object]]::new()
$selected.Add('Identity','Global'); $selected.Add('Prompts',[Activator]::CreateInstance($public.Parameters['Prompts'].ParameterType)); $selected.Add('Packages',[Activator]::CreateInstance($public.Parameters['Packages'].ParameterType))
$settings = [Newtonsoft.Json.JsonSerializerSettings]::new()
$settings.ContractResolver = [Activator]::CreateInstance($resolver,[object[]]@($selected))
if ($normalized -ne 'TeamsAudioConferencingCustomPromptsConfiguration') { throw 'Own normalized route name changed' }
foreach ($o in $observations) {
    $fails = $o.case -in @('null-remove-empty','null-remove-populated') -or ($o.field -eq 'Packages' -and $o.case -eq 'compare-remove-populated')
    if ($fails) {
        if ($o.stage -ne 'helper' -or $null -eq $o.error -or $null -ne $o.serialized) { throw 'Expected helper failure changed' }
    } elseif ($o.stage -ne 'complete' -or $null -ne $o.error) { throw 'Expected accepted local helper/serializer outcome changed' }
    if ($o.case -in @('empty-clears-populated','null-modifier-populated','compare-remove-empty') -and $o.remaining -ne 0) { throw 'Clear/empty removal outcome changed' }
    if ($o.case -in @('null-add','null-replace') -and ($o.remaining -ne 1 -or $o.serialized -notlike '*[[]null[]]*')) { throw 'Nullable entry was rejected or lost' }
}
$readJson = [Newtonsoft.Json.JsonConvert]::SerializeObject([ordered]@{Identity=$read.Identity;Prompts=@($prompt);Packages=@($package)})
[pscustomobject]@{
    hashes=$hashes
    reproducer_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant()
    public_get_kind=$get.CommandType.ToString()
    public_get_type=$get.ImplementingType.FullName
    public_get_base=$get.ImplementingType.BaseType.ToString()
    public_set_type=$public.ImplementingType.FullName
    normalized_resource=$normalized
    observations=@($observations)
    authored_read_projection=$readJson
    authored_fixture_sha256=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData([Text.Encoding]::UTF8.GetBytes($readJson+"`n"))).ToLowerInvariant()
    management_lifecycle_invoked=$false
    authentication_or_http=$false
} | ConvertTo-Json -Depth 6
