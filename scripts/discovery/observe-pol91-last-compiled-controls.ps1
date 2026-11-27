[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference='Stop'
$hashes=@{
 'Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll'='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'
 'Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll'='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'
}
foreach($name in $hashes.Keys){if((Get-FileHash(Join-Path $ModulePath "netcoreapp3.1/$name")).Hash.ToLowerInvariant() -ne $hashes[$name]){throw 'Pinned source differs'}}
Import-Module (Join-Path $ModulePath 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
$ops = @{}
[Reflection.Emit.OpCodes].GetFields([Reflection.BindingFlags]'Public,Static') |
    ForEach-Object { $op = $_.GetValue($null); $ops[[int]$op.Value -band 65535] = $op }
function Read-MethodIL([Reflection.MethodBase]$method) {
    $body = $method.GetMethodBody()
    if ($null -eq $body) { continue }
    $bytes = $body.GetILAsByteArray()
    $typeArgs = $method.DeclaringType.GetGenericArguments()
    $methodArgs = [Type[]]@()
    if ($method -is [Reflection.MethodInfo]) { $methodArgs = $method.GetGenericArguments() }
    for ($pos = 0; $pos -lt $bytes.Length;) {
        $offset = $pos
        $code = [int]$bytes[$pos++]
        if ($code -eq 254) { $code = 65024 + [int]$bytes[$pos++] }
        $op = $ops[$code]
        if ($null -eq $op) { throw "Unknown opcode at $offset" }
        $kind = [string]$op.OperandType
        $size = switch ($kind) {
            InlineNone { 0 }
            { $_ -in 'ShortInlineBrTarget','ShortInlineI','ShortInlineVar' } { 1 }
            InlineVar { 2 }
            { $_ -in 'InlineI8','InlineR' } { 8 }
            InlineSwitch { 4 + 4 * [BitConverter]::ToInt32($bytes, $pos) }
            default { 4 }
        }
        $value = $null
        if ($kind -eq 'InlineString') {
            $value = $method.Module.ResolveString([BitConverter]::ToInt32($bytes, $pos))
        } elseif ($kind -in 'InlineMethod','InlineField','InlineType','InlineTok') {
            $token = [BitConverter]::ToInt32($bytes, $pos)
            $member = $method.Module.ResolveMember($token, $typeArgs, $methodArgs)
            $value = "$($member.DeclaringType).$member"
        } elseif ($size -eq 4 -and $kind -ne 'InlineSwitch') {
            $value = [BitConverter]::ToInt32($bytes, $pos)
        } elseif ($size -eq 1) { $value = $bytes[$pos] }
        [pscustomobject]@{ Method=$method.Name; Offset=$offset; Opcode=$op.Name; Operand=$value }
        $pos += $size
    }
}

$flags=[Reflection.BindingFlags]'Public,NonPublic,Instance,Static'
$rows=@()
foreach($name in @('Set-CsTeamsGuestMeetingConfiguration','Set-CsTeamsMeetingConfiguration')){
 $command=Get-Command $name
 if($command.CommandType -ne 'Cmdlet'){throw 'Expected compiled command'}
 $type=$command.ImplementingType
 $model=$type.BaseType.GetGenericArguments()[0]
 $methods=@($type.GetMethod('AreParametersSpecified',$flags),$type.GetMethod('UpdateProperties',$flags),$type.BaseType.GetMethod('ProcessRecord',$flags))
 foreach($p in $model.GetProperties()){
  if($p.Name -eq 'Identity'){continue}
  $publicProperty=$type.GetProperty($p.Name)
  if($null -ne $publicProperty -and $null -ne $publicProperty.SetMethod){$methods+=$publicProperty.SetMethod}
 }
 foreach($method in $methods){
  if($null -eq $method){throw 'Expected exact source method'}
  $body=$method.GetMethodBody()
  $rows+=[pscustomobject]@{command=$name;public_type=$type.FullName;declaring_type=$method.DeclaringType.ToString();method=$method.Name;signature=$method.ToString();il_sha256=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($body.GetILAsByteArray())).ToLowerInvariant();instructions=@(Read-MethodIL $method)}
 }
}
[pscustomobject]@{scope='Static owning-generic-context reflection only: concrete last compiled Set bound guards/update/setter methods and shared closed BaseTpmSet preparation ordering. No management lifecycle, authentication or HTTP.';hashes=$hashes;reproducer_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant();selected_methods=$rows;branch_operands='Encoded relative branch operands are retained as read, not presented as absolute targets.';management_lifecycle_invoked=$false;authentication_or_http=$false}|ConvertTo-Json -Depth 9
