# Bounded pure-data observation: only model constructors, JsonConvert deserialization,
# auto-property getters and IL-audited pure license predicates. No report cmdlet invocation.
$ErrorActionPreference='Stop'
Import-Module '/private/tmp/msteams-core90-pinned/module/MicrosoftTeams.psd1' -Force -WarningAction SilentlyContinue
foreach($a in [AppDomain]::CurrentDomain.GetAssemblies()){try{$null=$a.GetTypes()}catch{}}
function Find-Type([string]$name){[AppDomain]::CurrentDomain.GetAssemblies()|ForEach-Object{$_.GetType($name,$false)}|Where-Object{$null -ne $_}|Select-Object -First 1}
function Sha([string]$s){[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData([Text.Encoding]::UTF8.GetBytes($s))).ToLowerInvariant()}
$license=Find-Type 'Microsoft.Teams.PowerShell.TeamsCmdlets.Model.LicenseDetails'
$plan=Find-Type 'Microsoft.Teams.PowerShell.TeamsCmdlets.Model.ServicePlanInfo'
$lambda=Find-Type 'Microsoft.Teams.PowerShell.TeamsCmdlets.GetLicenseReportForChangeNotificationSubscription+<>c'
$instance=[Activator]::CreateInstance($lambda,$true)
$flags=[Reflection.BindingFlags]'Instance,Public,NonPublic'
$licensePredicate=$lambda.GetMethod('<GenerateDlpUserLicenseReports>b__14_1',$flags)
$planPredicate=$lambda.GetMethod('<GenerateDlpUserLicenseReports>b__14_2',$flags)
function Observe([string]$label,[string]$json,[Type]$type,[Reflection.MethodInfo]$predicate,[string]$field){
 $data=[Newtonsoft.Json.JsonConvert]::DeserializeObject($json,$type)
 $p=$type.GetProperty($field)
 $value=$null;if($null -ne $data){$value=$p.GetValue($data)}
 $outcome=$null;$errorType=$null;$param=$null
 try{$outcome=$predicate.Invoke($instance,[object[]]@(,$data))}catch{$err=$_.Exception;while($err.InnerException){$err=$err.InnerException};$errorType=$err.GetType().FullName;if($err -is [ArgumentException]){$param=$err.ParamName}}
 @{case=$label;json=$json;json_sha256=(Sha $json);object_is_null=($null -eq $data);property=$field;property_is_null=($null -eq $value);property_value=$value;predicate_result=$outcome;exception_type=$errorType;exception_parameter=$param}
}
$rows=@()
foreach($case in @(
 @('license-null','null'),@('plans-missing','{}'),@('plans-null','{"servicePlans":null}'),@('plans-empty','{"servicePlans":[]}'),@('plan-null-item','{"servicePlans":[null]}'),@('plan-id-missing','{"servicePlans":[{}]}'),@('plan-id-null','{"servicePlans":[{"servicePlanId":null}]}'),@('plan-id-empty','{"servicePlans":[{"servicePlanId":""}]}'),@('plan-id-empty-guid','{"servicePlans":[{"servicePlanId":"00000000-0000-0000-0000-000000000000"}]}'),@('plan-id-other','{"servicePlans":[{"servicePlanId":"other"}]}'),@('plan-id-match','{"servicePlans":[{"servicePlanId":"6dc145d6-95dd-4191-b9c3-185575ee6f6b"}]}'),@('plan-id-uppercase','{"servicePlans":[{"servicePlanId":"6DC145D6-95DD-4191-B9C3-185575EE6F6B"}]}'),@('plan-match-before-null','{"servicePlans":[{"servicePlanId":"6dc145d6-95dd-4191-b9c3-185575ee6f6b"},null]}'),@('plan-null-before-match','{"servicePlans":[null,{"servicePlanId":"6dc145d6-95dd-4191-b9c3-185575ee6f6b"}]}')
)){$rows+=Observe $case[0] $case[1] $license $licensePredicate 'ServicePlans'}
foreach($case in @(@('plan-object-missing','{}'),@('plan-object-null-id','{"servicePlanId":null}'),@('plan-object-null','null'))){$rows+=Observe $case[0] $case[1] $plan $planPredicate 'ServicePlanId'}
$outer=@()
$delegateType=[Func``2].MakeGenericType($license,[bool])
$predicateDelegate=$licensePredicate.CreateDelegate($delegateType,$instance)
$any=([Linq.Enumerable].GetMethods()|Where-Object{$_.Name -eq 'Any' -and $_.IsGenericMethodDefinition -and $_.GetParameters().Count -eq 2}|Select-Object -First 1).MakeGenericMethod($license)
foreach($case in @(@('licenses-null','null'),@('licenses-empty','[]'),@('licenses-missing-plans','[{}]'),@('licenses-null-item','[null]'),@('licenses-empty-plans','[{"servicePlans":[]}]'),@('licenses-match','[{"servicePlans":[{"servicePlanId":"6dc145d6-95dd-4191-b9c3-185575ee6f6b"}]}]'),@('licenses-match-before-null','[{"servicePlans":[{"servicePlanId":"6dc145d6-95dd-4191-b9c3-185575ee6f6b"}]},null]'),@('licenses-null-before-match','[null,{"servicePlans":[{"servicePlanId":"6dc145d6-95dd-4191-b9c3-185575ee6f6b"}]}]'))){
 $data=[Newtonsoft.Json.JsonConvert]::DeserializeObject($case[1],$license.MakeArrayType());$result=$null;$errType=$null;$param=$null
 try{$argsArray=[object[]]::new(2);$argsArray[0]=$data;$argsArray[1]=$predicateDelegate;$result=$any.Invoke($null,$argsArray)}catch{$err=$_.Exception;while($err.InnerException){$err=$err.InnerException};$errType=$err.GetType().FullName;if($err -is [ArgumentException]){$param=$err.ParamName}}
 $outer+=@{case=$case[0];json=$case[1];json_sha256=(Sha $case[1]);collection_is_null=($null -eq $data);any_result=$result;exception_type=$errType;exception_parameter=$param;projector_null_bypass=($null -eq $data)}
}
$defaults=@();foreach($t in @($license,$plan)){$obj=[Activator]::CreateInstance($t);$defaults+=@{type=$t.FullName;properties=@($t.GetProperties()|ForEach-Object{@{name=$_.Name;type=$_.PropertyType.FullName;default=$_.GetValue($obj);attributes=@($_.GetCustomAttributesData()|ForEach-Object{[string]$_})}})}}
@{assembly_sha256=(Get-FileHash $license.Assembly.Location -Algorithm SHA256).Hash.ToLowerInvariant();reproducer_sha256=(Get-FileHash $PSCommandPath -Algorithm SHA256).Hash.ToLowerInvariant();predicate_methods=@($licensePredicate.ToString(),$planPredicate.ToString());defaults=$defaults;observations=$rows;outer_any_observations=$outer}|ConvertTo-Json -Depth 12
