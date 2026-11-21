[CmdletBinding()]
param([Parameter(Mandatory)][string]$ModulePath,[Parameter(Mandatory)][string]$OutputPath)
$ErrorActionPreference='Stop'
if(Test-Path -LiteralPath $OutputPath){throw 'Output already exists.'}
$assemblies=@(
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll';sha256='68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49'},
 @{path='netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll';sha256='a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51'},
 @{path='netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll';sha256='477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af'}
)
foreach($a in $assemblies){if((Get-FileHash -LiteralPath (Join-Path $ModulePath $a.path) -Algorithm SHA256).Hash.ToLowerInvariant() -ne $a.sha256){throw 'Wrong pinned banner text dependency.'}}
$manifest=Join-Path $ModulePath 'MicrosoftTeams.psd1'
if((Import-PowerShellDataFile -LiteralPath $manifest).ModuleVersion -ne '8.1.0'){throw 'Wrong baseline.'}
$env:MSTeamsContextInternal=$null
Import-Module $manifest -Force
$references=@((Get-ChildItem (Join-Path $PSHOME 'ref') -Filter '*.dll').FullName)+@((Join-Path $PSHOME 'System.Management.Automation.dll'))+@($assemblies|ForEach-Object{Join-Path $ModulePath $_.path})+@((Join-Path $PSHOME 'Newtonsoft.Json.dll'))
Add-Type -Path @((Join-Path $PSScriptRoot 'oracle/PolicyProviderCapture.cs'),(Join-Path $PSScriptRoot 'oracle/BannerTextProviderCapture.cs')) -ReferencedAssemblies $references -CompilerOptions '/nowarn:1701'
$cases=[Collections.Generic.List[object]]::new()
$id='01234567-89ab-cdef-0123-456789abcdef'
foreach($c in @(
 @{name='read-identity';op='read';selector='identity';identity="Global/$id";relative=$id;input='{}';status=200;reply=('{"Identity":"Global/'+$id+'","Id":"'+$id+'","Text":"fixture","Description":null,"Future":false}');force=-1},
 @{name='read-relative';op='read';selector='relative';identity='';relative=$id;input='{}';status=200;reply='{}';force=-1},
 @{name='read-extra-slash';op='read';selector='identity';identity="scope /$id / Child";relative=$id;input='{}';status=200;reply='{}';force=-1},
 @{name='list';op='list';selector='default';identity='';relative='';input='{}';status=200;reply='[]';force=-1},
 @{name='create-identity';op='create';selector='identity';identity="Global/$id";relative=$id;input='{"Text":"fixture","Description":""}';status=201;reply='{}';force=-1},
 @{name='create-default-id';op='create';selector='default';identity='';relative='';input='{"Text":"fixture"}';status=201;reply='{}';force=-1},
 @{name='create-relative';op='create';selector='relative';identity='';relative=$id;input='{"Text":"fixture","Description":""}';status=201;reply='{}';force=-1},
 @{name='create-identity-force';op='create';selector='identity';identity="Global/$id";relative=$id;input='{"Text":""}';status=204;reply='';force=0},
 @{name='create-relative-force';op='create';selector='relative';identity='';relative=$id;input='{"Text":""}';status=204;reply='';force=1},
 @{name='create-relative-missing-text';op='create';selector='relative';identity='';relative=$id;input='{}';status=201;reply='{}';force=-1},
 @{name='create-identity-description-null';op='create';selector='identity';identity="Global/$id";relative=$id;input='{"Text":"fixture","Description":null}';status=204;reply='';force=-1},
 @{name='patch-identity';op='patch';selector='identity';identity="Global/$id";relative=$id;input='{"Text":"","Description":null}';status=204;reply='';force=-1},
 @{name='patch-relative';op='patch';selector='relative';identity='';relative=$id;input='{"Text":"","Description":null}';status=204;reply='';force=-1},
 @{name='patch-relative-null-id';op='patch';selector='relative';identity='';relative=$id;input='{"Description":""}';status=204;reply='';force=-1},
 @{name='patch-relative-description-only';op='patch';selector='relative';identity='';relative=$id;input='{"Description":""}';status=204;reply='';force=0},
 @{name='patch-identity-force';op='patch';selector='identity';identity="Global/$id";relative=$id;input='{"Description":""}';status=204;reply='';force=1},
 @{name='patch-text-null';op='patch';selector='identity';identity="Global/$id";relative=$id;input='{"Text":null}';status=204;reply='';force=-1},
 @{name='delete-identity';op='delete';selector='identity';identity="Global/$id";relative=$id;input='{}';status=204;reply='';force=-1},
 @{name='delete-relative';op='delete';selector='relative';identity='';relative=$id;input='{}';status=204;reply='';force=-1},
 @{name='delete-relative-force';op='delete';selector='relative';identity='';relative=$id;input='{}';status=204;reply='';force=0},
 @{name='read-null';op='read';selector='relative';identity='';relative=$id;input='{}';status=200;reply='null';force=-1},
 @{name='read-empty';op='read';selector='relative';identity='';relative=$id;input='{}';status=204;reply='';force=-1},
 @{name='list-envelope';op='list';selector='default';identity='';relative='';input='{}';status=200;reply='{"value":[]}';force=-1},
 @{name='read-denied';op='read';selector='relative';identity='';relative=$id;input='{}';status=403;reply='{"Message":"fixture-denied"}';force=-1}
)){
 $readID=if($c.name -eq 'patch-relative-null-id'){''}else{$id}
 $observed=[TeamsContractOracle.BannerTextOracle]::Run($c.op,$c.selector,$c.identity,$c.relative,$c.input,'read-model-identity',$readID,$c.status,$c.reply,$c.force)
 $cases.Add(@{name=$c.name;operation=$c.op;selector=$c.selector;identity=$c.identity;id=$c.relative;input=$c.input;status=$c.status;response=$c.reply;force_mode=$c.force;read_model_id=$readID;observed=$observed})
}
$models=[Collections.Generic.List[object]]::new()
foreach($inputJson in @('constructor','text-constructor-null','text-constructor-empty','{}','{"Text":null}','{"Text":""}','{"Description":null}','{"Description":""}','{"Id":null}','{"Id":""}','{"Id":"01234567-89AB-CDEF-0123-456789abcdef"}','{"Id":"01234567-89ab-cdef-0123-456789abcdef\n"}','{"Id":"{01234567-89ab-cdef-0123-456789abcdef}"}','{"Text":"first","text":"second"}','{"Text":"first","Text":null}','{"Text":[]}',(@{Text=('x'*2000);Description=('x'*2000)}|ConvertTo-Json -Compress))){
 $models.Add(@{input=$inputJson;observation=([TeamsContractOracle.BannerTextOracle]::Model($inputJson)|ConvertFrom-Json -AsHashtable)})
}
[ordered]@{
 schema_version=1;baseline_version='8.1.0';captured_at=[DateTime]::UtcNow.ToString('yyyy-MM-dd')
 qualification='offline-banner-text-provider-scope-and-models';package_sha256='78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7'
 assemblies=$assemblies;host_powershell_version=$PSVersionTable.PSVersion.ToString()
 stage='Actual command constructors under injected memory client (factory guarded), pure PrepareForScopeCheck, actual direct HttpRequestHelper with synthetic InvocationInfo, credential-refusing host and memory sender; actual model setters/Newtonsoft. No Begin/Process/End, public binder, scope authorization, token acquisition, discovery or network. Static client/version/host restored. New constructor GUID is observed by shape; explicit fixture ID selects deterministic requests.'
 cases=$cases.ToArray();models=$models.ToArray()
}|ConvertTo-Json -Depth 30|Set-Content -LiteralPath $OutputPath -Encoding utf8
"banner provider/scope cases: $($cases.Count); model observations: $($models.Count)"
