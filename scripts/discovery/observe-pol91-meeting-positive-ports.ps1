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
$public = Get-Command Set-CsTeamsMeetingConfiguration
$model = $public.ImplementingType.BaseType.GetGenericArguments()[0]
$get = Get-Command Get-CsTeamsMeetingConfiguration
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
$fields=@('ClientAudioPort','ClientAudioPortRange','ClientVideoPort','ClientVideoPortRange','ClientAppSharingPort','ClientAppSharingPortRange')
$observations=foreach($field in $fields){
 $paramSetter=$public.ImplementingType.GetProperty($field).SetMethod
 foreach($case in @('null','zero','one','maximum')){
  $dto=[Activator]::CreateInstance($model);$model.GetProperty('Identity').SetValue($dto,'Global')
  $value=switch($case){'null'{$null};'zero'{[uint32]0};'one'{[uint32]1};'maximum'{[uint32]::MaxValue}}
  $failure=$null;$serialized=$null;$stage='dto-setter'
  try{
   $model.GetProperty($field).SetValue($dto,$value)
   $stage='serializer'
   $bound=[Collections.Generic.Dictionary[string,object]]::new();$bound.Add('Identity','Global');$bound.Add($field,$value)
   $bound=$nested.Invoke($null,[object[]]@($bound,$model,$false))
   $settings=[Newtonsoft.Json.JsonSerializerSettings]::new();$settings.ContractResolver=[Activator]::CreateInstance($resolver,[object[]]@($bound))
   $serialized=[Newtonsoft.Json.JsonConvert]::SerializeObject($dto,$settings);$stage='complete'
  }catch{$failure=$_.Exception.GetBaseException().GetType().FullName}
  [pscustomobject]@{field=$field;case=$case;parameter_setter_il=[Convert]::ToHexString($paramSetter.GetMethodBody().GetILAsByteArray()).ToLowerInvariant();stage=$stage;error=$failure;serialized=$serialized}
 }
}
[pscustomobject]@{scope='Actual own nullable positiveInteger DTO setters, selected nested-bound serializer and public parameter setter IL. No management lifecycle, authentication or HTTP.';hashes=$hashes;reproducer_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant();public_set_type=$public.ImplementingType.FullName;observations=@($observations);management_lifecycle_invoked=$false;authentication_or_http=$false}|ConvertTo-Json -Depth 7
