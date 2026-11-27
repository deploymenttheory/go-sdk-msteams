param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference='Stop'
$dll=Join-Path $ModulePath 'netcoreapp3.1/Microsoft.Teams.PowerShell.TeamsCmdlets.dll'
$expected='3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24'
if((Get-FileHash $dll -Algorithm SHA256).Hash.ToLowerInvariant() -ne $expected){throw 'Pinned TeamsCmdlets hash mismatch'}
Import-Module (Join-Path $ModulePath 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
$t=[Microsoft.Teams.PowerShell.TeamsCmdlets.Model.AadUser]
$fields=[ordered]@{id='Id';displayName='DisplayName';userPrincipalName='UserPrincipalName';userType='UserType';mail='Mail'}
$metadata=@($fields.Values|ForEach-Object{$p=$t.GetProperty($_);if($p.PropertyType -ne [string]){throw 'Unexpected property type'};[ordered]@{name=$p.Name;type=$p.PropertyType.FullName;setter_il=[Convert]::ToHexString($p.SetMethod.GetMethodBody().GetILAsByteArray());getter_il=[Convert]::ToHexString($p.GetMethod.GetMethodBody().GetILAsByteArray())}})
$rows=@()
foreach($field in $fields.Keys){
 foreach($token in @('null','""','"opaque-user"','" \u00e9 "','true','false','37','9007199254740993','1.5','1e2','1e400','-0','[]','{}')){
  $inputJson='{"'+$field+'":'+$token+'}'
  try{$value=[Newtonsoft.Json.JsonConvert]::DeserializeObject($inputJson,$t);$rows+= [ordered]@{field=$field;input_json=$inputJson;value=$value.($fields[$field]);exception=$null}}
  catch{$rows+=[ordered]@{field=$field;input_json=$inputJson;value=$null;exception=$_.Exception.GetBaseException().GetType().FullName}}
 }
}
$controls=@()
foreach($inputJson in @(
 '{"diſplayName":"foreign","uſerPrincipalName":"foreign"}',
 '{"diſplayName":{},"uſerPrincipalName":[]}',
 '{"displayName":"first","DISPLAYNAME":"last"}',
 '{"displayName":"first","DISPLAYNAME":null}',
 '{"id":"first","ID":null}',
 '{"ID":null,"id":"last"}',
 '{"id":"first","id":"last"}',
 '{"id":{},"ID":"later"}',
 '{"DISPLAYNAME":37,"USERPRINCIPALNAME":false,"USERTYPE":true,"MAIL":1e2}',
 '{"displayName":"known","diſplayName":{}}',
 '{"diſplayName":{},"displayName":"known"}'
)){
 try{$value=[Newtonsoft.Json.JsonConvert]::DeserializeObject($inputJson,$t);$projection=[ordered]@{};foreach($field in $fields.Keys){$projection[$field]=$value.($fields[$field])};$controls+=[ordered]@{input_json=$inputJson;projection=$projection;exception=$null}}
 catch{$controls+=[ordered]@{input_json=$inputJson;projection=$null;exception=$_.Exception.GetBaseException().GetType().FullName}}
}
# Decode data-only method bodies with the owning generic context. No method
# under inspection is invoked, including the management/HTTP consumers.
$ops=@{}
foreach($f in [Reflection.Emit.OpCodes].GetFields([Reflection.BindingFlags]'Public,Static')){$op=$f.GetValue($null);$ops[[int]$op.Value-band65535]=$op}
function Read-MethodIL($method){
 $body=$method.GetMethodBody();if(!$body){return}
 $bytes=$body.GetILAsByteArray();$position=0;$instructions=@()
 while($position-lt$bytes.Length){
  $offset=$position;$code=[int]$bytes[$position];$position++
  if($code-eq254){$code=65024+[int]$bytes[$position];$position++}
  $op=$ops[$code];$kind=[string]$op.OperandType;$size=0;$operand=$null
  switch($kind){
   InlineNone {$size=0}
   ShortInlineBrTarget {$size=1;$operand=$position+1+[sbyte]::Parse($bytes[$position].ToString('X2'),[Globalization.NumberStyles]::HexNumber)}
   ShortInlineI {$size=1;$operand=[int]$bytes[$position]}
   ShortInlineVar {$size=1;$operand=[int]$bytes[$position]}
   InlineVar {$size=2;$operand=[BitConverter]::ToUInt16($bytes,$position)}
   InlineI8 {$size=8;$operand=[BitConverter]::ToInt64($bytes,$position)}
   InlineR {$size=8;$operand=[BitConverter]::ToDouble($bytes,$position)}
   ShortInlineR {$size=4;$operand=[BitConverter]::ToSingle($bytes,$position)}
   InlineSwitch {$count=[BitConverter]::ToInt32($bytes,$position);$size=4+4*$count;$operand=[Convert]::ToHexString($bytes[$position..($position+$size-1)])}
   InlineBrTarget {$size=4;$operand=$position+4+[BitConverter]::ToInt32($bytes,$position)}
   InlineI {$size=4;$operand=[BitConverter]::ToInt32($bytes,$position)}
   InlineString {$size=4;$operand=$method.Module.ResolveString([BitConverter]::ToInt32($bytes,$position))}
   InlineSig {$size=4;$operand=[Convert]::ToHexString($method.Module.ResolveSignature([BitConverter]::ToInt32($bytes,$position)))}
   default {$size=4;$member=$method.Module.ResolveMember([BitConverter]::ToInt32($bytes,$position),$method.DeclaringType.GetGenericArguments(),$method.GetGenericArguments());$operand=[string]$member.DeclaringType+'.'+[string]$member}
  }
  $instructions+=[ordered]@{offset=$offset;opcode=$op.Name;operand=$operand};$position+=$size
 }
 [ordered]@{declaring_type=$method.DeclaringType.FullName;signature=[string]$method;il_hex=[Convert]::ToHexString($bytes);instructions=$instructions}
}
$consumers=@()
foreach($typeName in @('Microsoft.Teams.PowerShell.TeamsCmdlets.Utils.GraphUtilities','Microsoft.Teams.PowerShell.TeamsCmdlets.Utils.HttpUtilities')){
 $type=$t.Assembly.GetType($typeName,$true)
 foreach($method in $type.GetMethods([Reflection.BindingFlags]'Public,NonPublic,Static,Instance,DeclaredOnly')){
  if(($type.Name-eq'GraphUtilities'-and$method.Name-in@('GetUserId','GetUpnForUsers'))-or($type.Name-eq'HttpUtilities'-and$method.Name-in@('Get','GetAll'))){$consumers+=Read-MethodIL $method}
 }
}

[ordered]@{schema_version=1;scope='Pure actual AadUser JsonConvert data operations and consumer IL inspection; no management lifecycle, auth or HTTP';powershell_version=$PSVersionTable.PSVersion.ToString();assembly_sha256=$expected;producer_sha256=(Get-FileHash $PSCommandPath -Algorithm SHA256).Hash.ToLowerInvariant();type=$t.FullName;properties=$metadata;observations=$rows;property_controls=$controls;consumer_methods=$consumers}|ConvertTo-Json -Depth 16
