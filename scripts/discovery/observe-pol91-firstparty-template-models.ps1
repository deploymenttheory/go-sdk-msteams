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
$public = Get-Command Set-CsTeamsFirstPartyMeetingTemplateConfiguration
$model = $public.ImplementingType.BaseType.GetGenericArguments()[0]
$get = Get-Command Get-CsTeamsFirstPartyMeetingTemplateConfiguration
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
$observations = foreach ($field in @('TeamsMeetingTemplates')) {
    $modifierType = $public.Parameters[$field].ParameterType
    if (@($public.Parameters[$field].Attributes | Where-Object { $_ -isnot [System.Management.Automation.ParameterAttribute] }).Count -ne 0) { throw 'Additional field attributes need review' }
    $itemType = $modifierType.GetGenericArguments()[0]
    $method = $closure.GetMethod('<UpdateProperties>b__2_0', [Reflection.BindingFlags]'Public,NonPublic,Instance')
    $funcType = [Func[object,object,bool]].GetGenericTypeDefinition().MakeGenericType([Type[]]@($itemType,$itemType,[bool]))
    $predicate = $method.CreateDelegate($funcType,$singleton)
    $closed = $apply.MakeGenericMethod($itemType)
    $ctor = $modifierType.GetConstructor([Type[]]@([hashtable]))
    foreach ($case in @('empty-clears-populated','null-modifier-empty','null-modifier-populated','null-remove-empty','null-remove-populated','null-add','null-replace','replace-entry','compare-remove-empty','compare-remove-populated')) {
        $dto = [Activator]::CreateInstance($model)
        $model.GetProperty('Identity').SetValue($dto,'Global')
        $model.GetProperty('Description').SetValue($dto,'retained description')
        $list = $model.GetProperty($field).GetValue($dto)
        $arguments = [object[]]@('entry')
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

$extra=foreach($case in @('duplicate-first-match','ordinal-case','unmatched-name','null-name','existing-null-entry')){
 $dto=[Activator]::CreateInstance($model)
 $model.GetProperty('Identity').SetValue($dto,'Global')
 $list=$model.GetProperty('TeamsMeetingTemplates').GetValue($dto)
 $remove=[Activator]::CreateInstance($itemType,[object[]]@('a'))
 switch($case){
  'duplicate-first-match' {
   foreach($id in @('first','second')){$e=[Activator]::CreateInstance($itemType,[object[]]@('a'));$itemType.GetProperty('Identity').SetValue($e,$id);$list.Add($e)}
  }
  'ordinal-case' {$list.Add([Activator]::CreateInstance($itemType,[object[]]@('A')))}
  'unmatched-name' {$list.Add([Activator]::CreateInstance($itemType,[object[]]@('other')))}
  'null-name' {$list.Add([Activator]::CreateInstance($itemType));$remove=[Activator]::CreateInstance($itemType)}
  'existing-null-entry' {$list.Add($null)}
 }
 $modifier=$ctor.Invoke([object[]]@(@{Remove=@($remove)}))
 $failure=$null;$body=$null;$stage='helper'
 try{
  $null=$closed.Invoke($null,[object[]]@($modifier,$list,$predicate))
  $stage='serializer'
  $bound=[Collections.Generic.Dictionary[string,object]]::new();$bound.Add('Identity','Global');$bound.Add('TeamsMeetingTemplates',$modifier)
  $bound=$nested.Invoke($null,[object[]]@($bound,$model,$false))
  $settings=[Newtonsoft.Json.JsonSerializerSettings]::new();$settings.ContractResolver=[Activator]::CreateInstance($resolver,[object[]]@($bound))
  $body=[Newtonsoft.Json.JsonConvert]::SerializeObject($dto,$settings);$stage='complete'
 }catch{$failure=$_.Exception.GetBaseException().GetType().FullName}
 [pscustomobject]@{field='TeamsMeetingTemplates';case=$case;remaining=$list.Count;stage=$stage;error=$failure;serialized=$body}
}
$observations=@($observations)+@($extra)
[pscustomobject]@{
 scope='Actual own FirstParty meeting template helper, Name predicate, nested expansion and sparse serializer. No management lifecycle, authentication or HTTP.'
 hashes=$hashes
 reproducer_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant()
 public_get_type=$get.ImplementingType.FullName
 public_get_base=$get.ImplementingType.BaseType.ToString()
 public_set_type=$public.ImplementingType.FullName
 normalized_resource=$normalized
 observations=@($observations)
 management_lifecycle_invoked=$false
 authentication_or_http=$false
}|ConvertTo-Json -Depth 8
