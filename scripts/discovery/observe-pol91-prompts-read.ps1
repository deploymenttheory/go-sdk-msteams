[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath)
$ErrorActionPreference='Stop'
$core=Join-Path $ModulePath 'netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll'
if((Get-FileHash $core).Hash.ToLowerInvariant()-ne'68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'){throw 'hash'}
$provider=Join-Path $ModulePath 'netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll'
if((Get-FileHash $provider).Hash.ToLowerInvariant()-ne'a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'){throw 'provider hash'}
Import-Module (Join-Path $ModulePath 'MicrosoftTeams.psd1') -WarningAction SilentlyContinue
$t=(Get-Command Get-CsTeamsAudioConferencingCustomPromptsConfiguration).ImplementingType.BaseType.GetGenericArguments()[0]
$id='00000000-0000-0000-0000-000000000001'
$cases=@();$i=0
foreach($token in @('9007199254740993','184467440737095516151234567890','1e+03','1E-03','1.2300','-0','-0.00','0.0','true','false')){
$cases+=@{name="name-$i";body='{"Prompts":[{"Id":"'+$id+'","Type":"FutureType","Locale":"xx","Name":'+$token+'}]}';field='name';token=$token}
$cases+=@{name="reference-$i";body='{"Packages":[{"Id":"'+$id+'","DefaultLocale":"xx","InboundStartRecordingPrompt":['+$token+']}]}';field='reference';token=$token};$i++
}
foreach($field in @('Id','Type','Locale','Name')){$cases+=@{name="null-$($field.ToLower())";body='{"Prompts":[{"'+$field+'":null}]}';field='prompt'}}
$cases+=@(
@{name='package-defaultlocale-null';body='{"Packages":[{"DefaultLocale":null}]}';field='package'},
@{name='numeric-id';body='{"Prompts":[{"Id":37}]}';field='prompt'},
@{name='name-object';body='{"Prompts":[{"Name":{}}]}';field='prompt'},
@{name='reference-object';body='{"Packages":[{"InboundStartRecordingPrompt":[{}]}]}';field='package'},
@{name='duplicate-reference';body='{"Packages":[{"InboundStartRecordingPrompt":["a",null],"InboundStartRecordingPrompt":[7,false]}]}';field='package'},
@{name='reference-null-after-array';body='{"Packages":[{"InboundStartRecordingPrompt":["a"],"InboundStartRecordingPrompt":null}]}';field='package'},
@{name='root-null-collections';body='{"Prompts":null,"Packages":null}';field='root'},
@{name='duplicate-prompts';body='{"Prompts":[null],"Prompts":[{"Name":"x"}]}';field='root'},
@{name='id-final-lf';body='{"Prompts":[{"Id":"'+$id+'\n"}]}';field='prompt'},
@{name='id-crlf';body='{"Prompts":[{"Id":"'+$id+'\r\n"}]}';field='prompt'},
@{name='id-double-lf';body='{"Prompts":[{"Id":"'+$id+'\n\n"}]}';field='prompt'}
)
foreach($f in @('InboundStartRecordingPrompt','InboundEndRecordingPrompt','OutboundStartRecordingPrompt','OutboundEndRecordingPrompt','MeetingStartRecordingPrompt','MeetingEndRecordingPrompt')) {
 $cases+=@{name=('missing-'+$f.ToLower());body='{"Packages":[{}]}'}
 $cases+=@{name=('null-'+$f.ToLower());body='{"Packages":[{"'+$f+'":null}]}'}
 $cases+=@{name=('empty-'+$f.ToLower());body='{"Packages":[{"'+$f+'":[]}]}'}
 $cases+=@{name=('alias-'+$f.ToLower());body='{"Packages":[{"'+$f+'":["a",null],"'+$f.ToLower()+'":[7,false],"'+$f+'":null}]}'}
}
$cases+=@{name='root-missing-collections';body='{}'}
$cases+=@{name='root-empty-collections';body='{"Prompts":[],"Packages":[]}'}
$cases+=@{name='root-alias-collections';body='{"Prompts":[null],"prompts":[{"Name":"x"}],"Prompts":null,"Packages":[null],"packages":[{}],"Packages":null}'}
$rows=@(foreach($c in $cases){
 $expected=$null;$errorType=$null
 try {
  $v=[Newtonsoft.Json.JsonConvert]::DeserializeObject($c.body,$t)
  $expected=[Newtonsoft.Json.JsonConvert]::SerializeObject([ordered]@{Identity=$v.Identity;Prompts=@($v.Prompts);Packages=@($v.Packages)})
 }catch{$errorType=$_.Exception.GetBaseException().GetType().FullName}
 [pscustomobject]@{name=$c.name;input=$c.body;accepted=($null-eq$errorType);expected=$expected;error=$errorType}
})
[pscustomobject]@{core_sha256=(Get-FileHash $core).Hash.ToLowerInvariant();provider_sha256=(Get-FileHash $provider).Hash.ToLowerInvariant();reproducer_sha256=(Get-FileHash $PSCommandPath).Hash.ToLowerInvariant();source_model=$t.FullName;deserializer='Providers.HttpClientHelper.SendAsync<TPolicy> -> JsonConvert.DeserializeObject<TPolicy> at IL452';observations=$rows;management_lifecycle_invoked=$false;authentication_or_http=$false}|ConvertTo-Json -Depth 6
