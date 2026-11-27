#requires -Version 7.0
# Read-only concrete predicate/base IL; no cmdlet instance or backend operation.
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference='Stop'
$decoder=Join-Path $PSScriptRoot 'inspect-ext-il.ps1'
$base=& $decoder -ModulePath $ModulePath -TypeName 'Microsoft.Teams.Policy.Administration.Cmdlets.Providers.BaseTpmSetCmdlet`1' -MethodPattern '^ProcessRecord$' | ConvertFrom-Json
$records=@()
foreach($name in @('SetConversationRoleCmdlet','SetConversationRolesSettingCmdlet','SetTenantConfigurationCmdlet','SetTeamsAppPolicyConfigurationCmdlet')){
 $full='Microsoft.Teams.Policy.Administration.Cmdlets.Core.'+$name
 $type=[AppDomain]::CurrentDomain.GetAssemblies()|ForEach-Object{$_.GetType($full,$false)}|Where-Object{$null-ne$_}|Select-Object -First 1
 $predicate=& $decoder -ModulePath $ModulePath -TypeName $full -MethodPattern '^AreParametersSpecified$'|ConvertFrom-Json
 $keys=@($predicate.methods.instructions|Where-Object{$_.op -eq 'ldstr'}|ForEach-Object{$_.operand})
 $records+=@{type=$full;base=$type.BaseType.FullName;predicate=$predicate;bound_keys=$keys;selection='BoundParameters.ContainsKey; omission is false, a bound key counts independently of its value. Value binding/model validation is separate.'}
}
@{schema_version=1;baseline_version='8.1.0';extractor_sha256=(Get-FileHash -LiteralPath $PSCommandPath -Algorithm SHA256).Hash.ToLowerInvariant();observation='Exact concrete overrides and common base ProcessRecord before token acquisition. No settings emits ProcessRecordFailed category7 via WriteError and returns; native API uses ErrInput. Bound null/false/zero/empty count as selected keys, without claiming each value valid for every parameter type.';base_process_record=$base;records=$records}|ConvertTo-Json -Depth 20 -Compress
