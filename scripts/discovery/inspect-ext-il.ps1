#requires -Version 7.0
# Read-only bounded IL/metadata inspection; inspected methods are never invoked.
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$TypeName,[string]$MethodPattern='.*')
$ErrorActionPreference='Stop'
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline'}
foreach($pin in @(@{file='Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';hash='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},@{file='Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';hash='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'},@{file='Microsoft.Teams.PowerShell.TeamsCmdlets.dll';hash='3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24'})){
 if((Get-FileHash (Join-Path (Join-Path $ModulePath 'netcoreapp3.1') $pin.file) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $pin.hash){throw 'Wrong pinned assembly'}
}
Import-Module $manifest -Force -WarningAction SilentlyContinue
foreach($assembly in [AppDomain]::CurrentDomain.GetAssemblies()){try{$null=$assembly.GetTypes()}catch{}}
$type=[AppDomain]::CurrentDomain.GetAssemblies()|% {$_.GetType($TypeName,$false)}|? {$null -ne $_}|select -First 1
$ops=@{};foreach($f in [Reflection.Emit.OpCodes].GetFields([Reflection.BindingFlags]'Public,Static')){$op=$f.GetValue($null);$ops[[int]$op.Value -band 65535]=$op}
$methods=@()
foreach($m in @($type.GetMethods([Reflection.BindingFlags]'Instance,Static,Public,NonPublic'))+@($type.GetConstructors([Reflection.BindingFlags]'Instance,Static,Public,NonPublic,DeclaredOnly'))){
 if($m.DeclaringType.FullName -ne $type.FullName -or $m.Name -notmatch $MethodPattern){continue};$b=$m.GetMethodBody();if(!$b){continue};$bytes=$b.GetILAsByteArray();if($bytes.Length -gt 131072){throw 'Method body exceeds inspection bound'};$pos=0;$ins=@()
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
@{extractor_sha256=(Get-FileHash -LiteralPath $PSCommandPath -Algorithm SHA256).Hash.ToLowerInvariant();baseline_version='8.1.0';type=$type.FullName;assembly_sha256=(Get-FileHash $type.Assembly.Location -Algorithm SHA256).Hash.ToLowerInvariant();fields=$fields;methods=$methods}|ConvertTo-Json -Depth 12 -Compress
