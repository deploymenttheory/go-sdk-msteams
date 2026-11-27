# Bounded header closure. No management lifecycle/auth/HTTP invocation.
# Executes only fake inert cmdlets in empty runspaces and pinned pure PSInvocationInfo data adapter.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference='Stop'
function Hash-Text([string]$text){[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData([Text.Encoding]::UTF8.GetBytes($text))).ToLowerInvariant()}
$private=Join-Path $ModulePath 'bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll'
if((Get-FileHash $private).Hash.ToLowerInvariant() -ne '8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'){throw 'Wrong private assembly'}
$scriptPins=ConvertFrom-Json -AsHashtable @'
{"custom/Merged_custom_PsExt.ps1": "03dbc3a85e12c98dd1807d604782c6a90d0656569541cd9540738825a7fcb255", "internal/Merged_internal.ps1": "6ff95158174bab031dc1fc24112a2bc6ae81cf1e2f59db4ecb58afbefc8c2bd9", "exports/ProxyCmdletDefinitionsWithHelp.ps1": "5a8a5a50081b1c6986b93d5b15723e966a2d22b64459a9c300cb2a1bf69bb3b5"}
'@
foreach($key in $scriptPins.Keys){if((Get-FileHash (Join-Path $ModulePath $key)).Hash.ToLowerInvariant() -ne $scriptPins[$key]){throw "Wrong pinned script: $key"}}
Import-Module (Join-Path $ModulePath 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
foreach($assembly in [AppDomain]::CurrentDomain.GetAssemblies()){try{$null=$assembly.GetTypes()}catch{}}
function Read-Type([string]$TypeName,[string]$MethodPattern=".*"){
$type=[AppDomain]::CurrentDomain.GetAssemblies()|% {$_.GetType($TypeName,$false)}|? {$null -ne $_}|select -First 1
if(!$type){throw "Missing selected type: $TypeName"}
$ops=@{};foreach($f in [Reflection.Emit.OpCodes].GetFields([Reflection.BindingFlags]'Public,Static')){$op=$f.GetValue($null);$ops[[int]$op.Value -band 65535]=$op}
$methods=@()
foreach($m in @($type.GetMethods([Reflection.BindingFlags]'Instance,Static,Public,NonPublic,DeclaredOnly'))+@($type.GetConstructors([Reflection.BindingFlags]'Instance,Static,Public,NonPublic,DeclaredOnly'))){
 if($m.Name -notmatch $MethodPattern){continue};$b=$m.GetMethodBody();if(!$b){continue};$bytes=$b.GetILAsByteArray();$pos=0;$ins=@()
 while($pos -lt $bytes.Length){$start=$pos;$code=[int]$bytes[$pos++];if($code -eq 254){$code=65024+[int]$bytes[$pos++]};$op=$ops[$code];$k=[string]$op.OperandType;$v=$null;$n=0
 switch($k){
 'InlineNone'{} 'ShortInlineI'{$n=1;$v=[int][sbyte]::Parse($bytes[$pos].ToString('X2'),[Globalization.NumberStyles]::HexNumber)}
 'ShortInlineVar'{$n=1;$v=[int]$bytes[$pos]} 'InlineVar'{$n=2;$v=[BitConverter]::ToUInt16($bytes,$pos)}
 'ShortInlineBrTarget'{$n=1;$v=$pos+1+[int][sbyte]::Parse($bytes[$pos].ToString('X2'),[Globalization.NumberStyles]::HexNumber)}
 'InlineBrTarget'{$n=4;$v=$pos+4+[BitConverter]::ToInt32($bytes,$pos)}
 'InlineI'{$n=4;$v=[BitConverter]::ToInt32($bytes,$pos)} 'InlineI8'{$n=8;$v=[BitConverter]::ToInt64($bytes,$pos)}
 'ShortInlineR'{$n=4;$v=[BitConverter]::ToSingle($bytes,$pos)} 'InlineR'{$n=8;$v=[BitConverter]::ToDouble($bytes,$pos)}
 'InlineSwitch'{$count=[BitConverter]::ToInt32($bytes,$pos);$n=4+4*$count;$v=@();for($j=0;$j -lt $count;$j++){$v+=($pos+$n+[BitConverter]::ToInt32($bytes,$pos+4+4*$j))}}
 default{$n=4;$tok=[BitConverter]::ToInt32($bytes,$pos);try{if($k -eq 'InlineString'){$v=$m.Module.ResolveString($tok)}else{$mem=$m.Module.ResolveMember($tok,$m.DeclaringType.GetGenericArguments(),$(if($m -is [Reflection.MethodInfo]){$m.GetGenericArguments()}else{@()}));$v=[string]$mem.DeclaringType+'.'+[string]$mem}}catch{$v="token:$tok"}}
 };$pos+=$n;$ins+=@{offset=$start;op=$op.Name;operand=$v}}
 $eh=@();foreach($e in $b.ExceptionHandlingClauses){$eh+=@{flags=[string]$e.Flags;tryOffset=$e.TryOffset;tryLength=$e.TryLength;handlerOffset=$e.HandlerOffset;handlerLength=$e.HandlerLength;catch=$(if($e.Flags -eq 'Clause'){[string]$e.CatchType}else{$null})}}
 $methods+=@{method=[string]$m;il_sha256=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($bytes)).ToLowerInvariant();locals=@($b.LocalVariables|%{[string]$_});exceptions=$eh;instructions=$ins}
}
$fields=@();foreach($f in $type.GetFields([Reflection.BindingFlags]'Public,NonPublic,Static,Instance,DeclaredOnly')){$fields+=@{name=$f.Name;type=[string]$f.FieldType;literal=$f.IsLiteral;constant=$(if($f.IsLiteral){$f.GetRawConstantValue()}else{$null})}}
@{type=$type.FullName;assembly_qualified_type=$type.AssemblyQualifiedName;assembly_identity=$type.Assembly.FullName;assembly_sha256=(Get-FileHash $type.Assembly.Location -Algorithm SHA256).Hash.ToLowerInvariant();fields=$fields;methods=$methods}

}
function Scan-HeaderLiterals {
$types=@();foreach($a in [AppDomain]::CurrentDomain.GetAssemblies()){if($a.GetName().Name -eq 'Microsoft.Teams.ConfigAPI.Cmdlets.private'){$types=$a.GetTypes()}}
$targets=@('set_EventListener','set_InvocationInformation')
$opcodes=@{};foreach($f in [Reflection.Emit.OpCodes].GetFields([Reflection.BindingFlags]'Public,Static')){$op=$f.GetValue($null);$opcodes[[int]$op.Value -band 65535]=$op}
$rows=@();$bodiesScanned=0;$stringsResolved=0;$stringFailures=0;foreach($t in $types){foreach($m in @($t.GetMethods([Reflection.BindingFlags]'Instance,Static,Public,NonPublic,DeclaredOnly'))+@($t.GetConstructors([Reflection.BindingFlags]'Instance,Static,Public,NonPublic,DeclaredOnly'))){$b=$m.GetMethodBody();if(!$b){continue};$bodiesScanned++;$bytes=$b.GetILAsByteArray();$pos=0;while($pos -lt $bytes.Length){$start=$pos;$code=[int]$bytes[$pos++];if($code -eq 254){$code=65024+[int]$bytes[$pos++]};$op=$opcodes[$code];$kind=[string]$op.OperandType;$size=4;switch($kind){'InlineNone'{$size=0}'ShortInlineBrTarget'{$size=1}'ShortInlineI'{$size=1}'ShortInlineVar'{$size=1}'InlineVar'{$size=2}'InlineI8'{$size=8}'InlineR'{$size=8}'InlineSwitch'{$size=4+4*[BitConverter]::ToInt32($bytes,$pos)}};if($kind -eq 'InlineString'){try{$literal=$m.Module.ResolveString([BitConverter]::ToInt32($bytes,$pos));$stringsResolved++;if($literal -in @('X-MS-CmdletName','User-Agent')){$rows+=@{type=$t.FullName;method=[string]$m;offset=$start;literal=$literal;il_sha256=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($bytes)).ToLowerInvariant()}}}catch{$stringFailures++;throw}};$pos+=$size}}}
@{scope='All method/constructor bodies declared in every type of the pinned private assembly; every InlineString resolved, matching exact command-name/User-Agent literal operands';types_scanned=$types.Count;method_bodies_scanned=$bodiesScanned;string_operands_resolved=$stringsResolved;unresolved_string_operands=$stringFailures;writers=@($rows)}

}
$inertCSharp=@'
using System.Management.Automation;
namespace Ext90InertHeader {
 public sealed class Snapshot { public string MyCommandName;public string InvocationName;public InvocationInfo Info; }
 [Cmdlet("Get","CsTeamsSettingsCustomApp_Get")]
 public sealed class ReadName:PSCmdlet {
  protected override void ProcessRecord(){WriteObject(new Snapshot{MyCommandName=MyInvocation.MyCommand.Name,InvocationName=MyInvocation.InvocationName,Info=MyInvocation});}
 }
 [Cmdlet("Set","CsTeamsSettingsCustomApp_SetExpanded")]
 public sealed class WriteName:PSCmdlet {
  protected override void ProcessRecord(){WriteObject(new Snapshot{MyCommandName=MyInvocation.MyCommand.Name,InvocationName=MyInvocation.InvocationName,Info=MyInvocation});}
 }
}

'@
function Observe-InertNames {
$refs=@(Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll'|ForEach-Object{$_.FullName})+@([Management.Automation.PSCmdlet].Assembly.Location)
Add-Type -TypeDefinition $inertCSharp -ReferencedAssemblies $refs
$adapter=[AppDomain]::CurrentDomain.GetAssemblies()|ForEach-Object{$_.GetType('Microsoft.Teams.ConfigApi.Cmdlets.PSInvocationInfo',$false)}|Where-Object{$null -ne $_}|Select-Object -First 1
if((Get-FileHash $adapter.Assembly.Location).Hash.ToLowerInvariant() -ne '8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6'){throw 'Wrong source adapter assembly'}
$interface=$adapter.GetInterface('Microsoft.Teams.ConfigApi.Cmdlets.IInvocationInfo')
$initial=[Management.Automation.Runspaces.InitialSessionState]::Create()
$initial.Commands.Add([Management.Automation.Runspaces.SessionStateCmdletEntry]::new('Get-CsTeamsSettingsCustomApp_Get',[Ext90InertHeader.ReadName],$null))
$initial.Commands.Add([Management.Automation.Runspaces.SessionStateAliasEntry]::new('Get-LocalAlias','Get-CsTeamsSettingsCustomApp_Get'))
$proxy=@'
[CmdletBinding(DefaultParameterSetName='Get',PositionalBinding=$false)]param()
begin {
 $wrappedCmd=$ExecutionContext.InvokeCommand.GetCommand('Get-CsTeamsSettingsCustomApp_Get',[System.Management.Automation.CommandTypes]::Cmdlet)
 $scriptCmd={& $wrappedCmd @PSBoundParameters}
 $step=$scriptCmd.GetSteppablePipeline($MyInvocation.CommandOrigin)
 $step.Begin($PSCmdlet)
}
process {$step.Process($_)}
end {$step.End()}
'@
$initial.Commands.Add([Management.Automation.Runspaces.SessionStateFunctionEntry]::new('Get-CsTeamsSettingsCustomApp',$proxy))
$initial.Commands.Add([Management.Automation.Runspaces.SessionStateFunctionEntry]::new('Set-CsTeamsSettingsCustomApp','[CmdletBinding()]param() process { Get-CsTeamsSettingsCustomApp }'))
$initial.Commands.Add([Management.Automation.Runspaces.SessionStateFunctionEntry]::new('Invoke-OuterCustomGet','[CmdletBinding()]param() process { Get-CsTeamsSettingsCustomApp }'))
$runspace=[Management.Automation.Runspaces.RunspaceFactory]::CreateRunspace($initial);$runspace.Open()
$rows=@()
try{
 foreach($command in @('Get-CsTeamsSettingsCustomApp_Get','Get-CsTeamsSettingsCustomApp','Set-CsTeamsSettingsCustomApp','Invoke-OuterCustomGet','Get-LocalAlias')){
  $pipeline=[Management.Automation.PowerShell]::Create();$pipeline.Runspace=$runspace
  try{
   $null=$pipeline.AddCommand($command);$values=$pipeline.Invoke()
   if($pipeline.Streams.Error.Count -ne 0 -or $values.Count -ne 1){throw 'Inert name observation failed'}
   $snapshot=$values[0].psobject.BaseObject
   $wrapped=[Activator]::CreateInstance($adapter,[object[]]@($snapshot.Info))
   $pinnedName=$interface.GetMethod('GetCommandName').Invoke($wrapped,@())
   $rows+=@{entry_command=$command;inert_type=$snapshot.GetType().FullName;mycommand_name=$snapshot.MyCommandName;invocation_name=$snapshot.InvocationName;pinned_adapter_command_name=$pinnedName}
  }finally{$pipeline.Dispose()}
 }
}finally{$runspace.Dispose()}
$initial=[Management.Automation.Runspaces.InitialSessionState]::Create()
$initial.Commands.Add([Management.Automation.Runspaces.SessionStateCmdletEntry]::new('Set-CsTeamsSettingsCustomApp_SetExpanded',[Ext90InertHeader.WriteName],$null))
$initial.Commands.Add([Management.Automation.Runspaces.SessionStateAliasEntry]::new('Set-LocalAlias','Set-CsTeamsSettingsCustomApp_SetExpanded'))
$setProxy=$proxy.Replace('Get-CsTeamsSettingsCustomApp_Get','Set-CsTeamsSettingsCustomApp_SetExpanded').Replace("DefaultParameterSetName='Get'","DefaultParameterSetName='SetExpanded'")
$initial.Commands.Add([Management.Automation.Runspaces.SessionStateFunctionEntry]::new('Set-CsTeamsSettingsCustomApp',$setProxy))
$initial.Commands.Add([Management.Automation.Runspaces.SessionStateFunctionEntry]::new('Invoke-OuterCustomSet','[CmdletBinding()]param() process { Set-CsTeamsSettingsCustomApp }'))
$runspace=[Management.Automation.Runspaces.RunspaceFactory]::CreateRunspace($initial);$runspace.Open()
try{
 foreach($command in @('Set-CsTeamsSettingsCustomApp_SetExpanded','Set-CsTeamsSettingsCustomApp','Invoke-OuterCustomSet','Set-LocalAlias')){
  $pipeline=[Management.Automation.PowerShell]::Create();$pipeline.Runspace=$runspace
  try{
   $null=$pipeline.AddCommand($command);$values=$pipeline.Invoke()
   if($pipeline.Streams.Error.Count -ne 0 -or $values.Count -ne 1){throw 'Inert set name observation failed'}
   $snapshot=$values[0].psobject.BaseObject
   $wrapped=[Activator]::CreateInstance($adapter,[object[]]@($snapshot.Info))
   $pinnedName=$interface.GetMethod('GetCommandName').Invoke($wrapped,@())
   $rows+=@{entry_command=$command;inert_type=$snapshot.GetType().FullName;mycommand_name=$snapshot.MyCommandName;invocation_name=$snapshot.InvocationName;pinned_adapter_command_name=$pinnedName}
  }finally{$pipeline.Dispose()}
 }
}finally{$runspace.Dispose()}
@{scope='Only fake local PSCmdlet lifecycle in empty runspace; generated management types absent. Real pinned data-only PSInvocationInfo constructor/GetCommandName exercised on inert InvocationInfo.';ps_assembly_sha256=(Get-FileHash ([Management.Automation.PSCmdlet].Assembly.Location)).Hash.ToLowerInvariant();pinned_assembly_sha256=(Get-FileHash $adapter.Assembly.Location).Hash.ToLowerInvariant();script_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant();fake_sha256=(Hash-Text $inertCSharp);observations=$rows}

}

$selection=ConvertFrom-Json @'
[["Microsoft.Teams.ConfigApi.Cmdlets.SendAsyncCallback", "AddSessionInfoToHttpRequest|AddUseragentAndCmdletNameHeaders|AddRequestHeader|AddOrUpdateOCERelatedHeaders|CustomRequestHandlerByService"], ["Microsoft.Teams.ConfigApi.Cmdlets.PSInvocationInfo", "GetCommandName|\\.ctor"], ["Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Cmdlets.GetCsTeamsSettingsCustomApp_Get", "InvocationInformation"], ["Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Cmdlets.SetCsTeamsSettingsCustomApp_SetExpanded", "InvocationInformation"], ["Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Cmdlets.GetCsTeamsSettingsCustomApp_Get+<ProcessRecordAsync>d__46", "MoveNext"], ["Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Cmdlets.SetCsTeamsSettingsCustomApp_SetExpanded+<ProcessRecordAsync>d__86", "MoveNext"], ["Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Module", "^CreatePipeline$|^BeforeCreatePipeline$|^AfterCreatePipeline$|^SendAsync$|^CustomInit$|^Signal$"], ["Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Module+<>c__DisplayClass25_0", "AfterCreatePipeline"], ["Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Module+<SendAsync>d__5", "MoveNext"], ["Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Module+<Signal>d__32", "MoveNext"], ["Microsoft.Teams.ConfigAPI.Cmdlets.Telemetry.CustomCmdletUtils", "^ProcessArgs$|^GetHttpPipelineArguments$|^RemoveHttpPipelineArgs$|^\\.ctor$"], ["Microsoft.Teams.ConfigAPI.Cmdlets.Telemetry.ExecutionIdContainer", "^SendAsyncStep$|^AddAndSignalExecutionId$|^\\.ctor$"], ["Microsoft.Teams.ConfigAPI.Cmdlets.Telemetry.ExecutionIdContainer+<AddAndSignalExecutionId>d__10", "MoveNext"], ["Microsoft.Teams.ConfigAPI.Cmdlets.Telemetry.AutoRestTelemetryContainer", "^SetExecutionId$"], ["Microsoft.Teams.ConfigAPI.Cmdlets.Telemetry.AutoRestTelemetryContainer+<_EventHandler>d__4", "MoveNext"], ["Microsoft.Teams.ConfigAPI.Cmdlets.Telemetry.CmdletExecutionData", "^SetExecutionId$|^SetFlighted$"], ["Microsoft.Teams.ConfigAPI.Cmdlets.AutoRestEventListeners+<EventForwarder>d__4", "MoveNext"], ["Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Runtime.Extensions", "^Clone$"], ["Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Runtime.CustomHttpClientFactory", "^SendAsync$|^GetClient$|^CreateAndUpdateHttpClient$"], ["System.Management.Automation.SteppablePipeline", "^Begin$"]]
'@
$selected=@();foreach($entry in $selection){$selected+=Read-Type $entry[0] $entry[1]}
foreach($record in $selected){if($record.methods.Count -eq 0){throw "Empty method selection: $($record.type)"};foreach($m in $record.methods){foreach($i in $m.instructions){if([string]$i.operand -like 'token:*'){throw 'Unresolved selected metadata token'}}}}
$definitions=@();foreach($relative in $scriptPins.Keys){$path=Join-Path $ModulePath $relative;$tokens=$null;$parseErrors=$null;$ast=[Management.Automation.Language.Parser]::ParseFile($path,[ref]$tokens,[ref]$parseErrors);if($parseErrors.Count){throw 'Pinned source parse failed'};foreach($fn in $ast.FindAll({param($node)$node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -in @('Get-CsTeamsSettingsCustomApp','Set-CsTeamsSettingsCustomApp')},$true)){$definitions+=@{file=$relative;file_sha256=$scriptPins[$relative];name=$fn.Name;start_line=$fn.Extent.StartLineNumber;definition_sha256=(Hash-Text $fn.Extent.Text);body=$fn.Extent.Text}}}
$metadata=@();foreach($name in @('Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Cmdlets.GetCsTeamsSettingsCustomApp_Get','Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Cmdlets.SetCsTeamsSettingsCustomApp_SetExpanded')){$type=[AppDomain]::CurrentDomain.GetAssemblies()|ForEach-Object{$_.GetType($name,$false)}|Where-Object{$null -ne $_}|Select-Object -First 1;$metadata+=@{type=$name;assembly_qualified_type=$type.AssemblyQualifiedName;attributes=@($type.GetCustomAttributesData()|ForEach-Object{[string]$_});invocation_property_attributes=@($type.GetProperty('InvocationInformation').GetCustomAttributesData()|ForEach-Object{[string]$_})}}
$literalScan=Scan-HeaderLiterals
if($literalScan.unresolved_string_operands -ne 0){throw 'Incomplete literal scan'}
$names=Observe-InertNames
@{schema_version=1;baseline='public8.1.0';scope='Bounded complete relevant source bodies plus whole-private-assembly exact-literal writer scan; only inert cmdlet lifecycle, no management/auth/HTTP';reproducer_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant();selected_il=$selected;source_definitions=$definitions;cmdlet_metadata=$metadata;literal_writer_scan=$literalScan;inert_probe_csharp=$inertCSharp;inert_probe_powershell=${function:Observe-InertNames}.ToString();literal_scanner_source=${function:Scan-HeaderLiterals}.ToString();literal_scanner_source_sha256=(Hash-Text ${function:Scan-HeaderLiterals}.ToString());selected_metadata_operands_unresolved=0;inert_probe_proxy_source=@'
[CmdletBinding(DefaultParameterSetName='Get',PositionalBinding=$false)]param()
begin {
 $wrappedCmd=$ExecutionContext.InvokeCommand.GetCommand('Get-CsTeamsSettingsCustomApp_Get',[System.Management.Automation.CommandTypes]::Cmdlet)
 $scriptCmd={& $wrappedCmd @PSBoundParameters}
 $step=$scriptCmd.GetSteppablePipeline($MyInvocation.CommandOrigin)
 $step.Begin($PSCmdlet)
}
process {$step.Process($_)}
end {$step.End()}
'@
;inert_name_observation=$names;conclusions=@{read_header='Get-CsTeamsSettingsCustomApp_Get';write_header='Set-CsTeamsSettingsCustomApp_SetExpanded';scope='Default built-in hooks; caller-supplied pipeline hooks/global listener mutations excluded';execution_id_step='Signals SetExecutionId then forwards same request/callback; no header access';adapter='Uses MyCommand.Name, not InvocationName; alias/proxy changes InvocationName only';capture_kind='Source-derived header values, backed by inert names; no actual management request sent'}}|ConvertTo-Json -Depth 18 -Compress
