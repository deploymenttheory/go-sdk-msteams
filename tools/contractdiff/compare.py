"""Compare complete discovery snapshots and explicit contract evidence without promoting a baseline."""
import argparse
import base64
import copy
import hashlib
import gzip
import json
import pathlib
import re
import sys

ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0,str(ROOT/'tools/baseline'))
from baseline import DiscoveryError, canonical_runtime, publish_snapshot

MISSING = object()
ROOT_FIELDS = {'schema_version','module_version','package_sha256','package_sha512_base64','package_url',
               'runtime_environment','runtime_commands','gallery_commands','release_notes','sources','limitations'}
COMMAND_FIELDS = {'name','command_type','output_types','parameter_sets','alias_target'}
SET_FIELDS = {'name','is_default','parameters'}
PARAM_FIELDS = {'name','type','mandatory','position','aliases','pipeline_input','pipeline_by_property','validation','default'}
SOURCE_FIELDS = {'path','sha256','lexical_command_candidates','help_commands','configuration'}


def canonical(value):
    return json.dumps(value,sort_keys=True,separators=(',',':'),ensure_ascii=False)


def keyed(items, name='name'):
    result={}
    for item in items:
        key=item[name].casefold()
        if key in result:
            raise DiscoveryError('Duplicate case-insensitive identity: '+item[name])
        result[key]=item
    return result


def unknown(value, known, path, warnings):
    fields=sorted(set(value)-known)
    if fields:
        warnings.append(dict(path=path,reason='Unknown fields require schema review',fields=fields))
    return {key:value[key] for key in fields}


def normalize(snapshot):
    if not isinstance(snapshot,dict) or snapshot.get('schema_version')!=1:
        raise DiscoveryError('Unsupported or missing snapshot schema')
    missing=ROOT_FIELDS-set(snapshot)
    if missing:
        raise DiscoveryError('Incomplete snapshot: '+', '.join(sorted(missing)))
    if not re.fullmatch(r'\d+\.\d+\.\d+(?:-[A-Za-z0-9.-]+)?',snapshot['module_version']):
        raise DiscoveryError('Invalid module version')
    if not re.fullmatch(r'[0-9a-f]{64}',snapshot['package_sha256']):
        raise DiscoveryError('Missing package digest')
    try:
        digest=base64.b64decode(snapshot['package_sha512_base64'],validate=True)
    except (ValueError,TypeError) as exc:
        raise DiscoveryError('Missing/invalid package SHA512 provenance') from exc
    if len(digest)!=64 or not isinstance(snapshot['package_url'],str) or not snapshot['package_url']:
        raise DiscoveryError('Incomplete package provenance')
    runtime=dict(snapshot['runtime_environment'],commands=copy.deepcopy(snapshot['runtime_commands']))
    commands=canonical_runtime(runtime,snapshot['module_version'])
    if not snapshot['gallery_commands'] or not snapshot['sources']:
        raise DiscoveryError('Empty Gallery or source extraction')
    warnings=[]
    model={'commands':{},'candidates':{},'configuration':{},'documentation':{},'opaque_sources':{},
           'unknown':unknown(snapshot,ROOT_FIELDS,'$',warnings)}
    environment_extra=unknown(snapshot['runtime_environment'],{'module_version','powershell_version','platform','authenticated'},'runtime_environment',warnings)
    if environment_extra:
        model['unknown']['runtime_environment']=environment_extra
    for command in commands:
        name=command['name'].casefold()
        item={'command_type':command['command_type'],'output_types':sorted(command['output_types']),
              'parameter_sets':{},'unknown':unknown(command,COMMAND_FIELDS,'commands.'+name,warnings)}
        if command['command_type']=='Alias':
            if not command.get('alias_target'):
                raise DiscoveryError('Alias target was not extracted: '+command['name'])
            item['alias_target']=command['alias_target'].casefold()
        for group_name,group in keyed(command['parameter_sets']).items():
            params={}
            for param_name,param in keyed(group['parameters']).items():
                value={k:copy.deepcopy(v) for k,v in param.items() if k not in {'name','aliases'}}
                value['aliases']=sorted(a.casefold() for a in param['aliases'])
                unknown(param,PARAM_FIELDS,f'commands.{name}.{group_name}.{param_name}',warnings)
                params[param_name]=value
            item['parameter_sets'][group_name]={'is_default':group['is_default'],'parameters':params,
                'unknown':unknown(group,SET_FIELDS,f'commands.{name}.{group_name}',warnings)}
        model['commands'][name]=item
    for name in snapshot['gallery_commands']:
        if not isinstance(name,str) or not re.fullmatch(r'[A-Za-z]+-[A-Za-z0-9]+',name):
            raise DiscoveryError('Invalid Gallery candidate')
        model['candidates'].setdefault(name.casefold(),{})['gallery']=True
    sources=keyed(snapshot['sources'],'path')
    if not any('help_commands' in s for s in sources.values()) or not any('configuration' in s for s in sources.values()):
        raise DiscoveryError('Help/configuration extraction is incomplete')
    for path,source in sources.items():
        if not re.fullmatch(r'[0-9a-f]{64}',source['sha256']):
            raise DiscoveryError('Invalid source digest')
        extra=unknown(source,SOURCE_FIELDS,'sources.'+path,warnings)
        if extra:
            model['unknown']['source:'+path]=extra
        for name in source.get('lexical_command_candidates',[]):
            model['candidates'].setdefault(name.casefold(),{}).setdefault('manifests',[]).append(path)
        if 'help_commands' in source:
            model['documentation'][path]={'commands':sorted(n.casefold() for n in source['help_commands']),
                                           'sha256':source['sha256']}
        if 'configuration' in source:
            configuration=copy.deepcopy(source['configuration'])
            cmdlets=configuration.pop('CmdletConfiguration',{})
            seen_config=set()
            for name,settings in cmdlets.items():
                if name.casefold() in seen_config:
                    raise DiscoveryError('Duplicate configuration command: '+name)
                seen_config.add(name.casefold())
                model['configuration'].setdefault(name.casefold(),{})[path]=settings
            if configuration:
                model['configuration'].setdefault('$shared',{})[path]=configuration
        if path.endswith(('.ps1','.psm1')):
            model['opaque_sources'][path]=source['sha256']
    for candidate in model['candidates'].values():
        if 'manifests' in candidate:
            candidate['manifests']=sorted(candidate['manifests'])
    model['documentation']['release_notes_sha256']=hashlib.sha256(snapshot['release_notes'].encode()).hexdigest()
    return model,warnings


def contract_model(contracts):
    """Optional sanitized facts retain documentation versus observed provenance."""
    result={}
    for entry in contracts:
        if not isinstance(entry,dict) or set(entry)!={'command','classification','evidence','facts'}:
            raise DiscoveryError('Contract evidence requires command, classification, evidence and facts')
        if not isinstance(entry['command'],str) or not re.fullmatch(r'[A-Za-z]+-[A-Za-z0-9]+',entry['command']):
            raise DiscoveryError('Invalid contract command identity')
        if entry['classification'] not in {'documentation','observed-behaviour'} or not isinstance(entry['evidence'],str) or not entry['evidence'] or not isinstance(entry['facts'],dict):
            raise DiscoveryError('Incomplete contract evidence provenance')
        name=entry['command'].casefold()
        if name in result:
            raise DiscoveryError('Duplicate contract evidence command')
        result[name]=entry
    return result


def differences(before, after, path=()):
    if isinstance(before,dict) and isinstance(after,dict):
        for key in sorted(set(before)|set(after)):
            yield from differences(before.get(key,MISSING),after.get(key,MISSING),path+(key,))
    elif before is MISSING or after is MISSING or canonical(before)!=canonical(after):
        yield path,before,after


def evidence(value):
    return {'present':False} if value is MISSING else {'present':True,'value':value}


def impact(command, ledger, mapping):
    record=ledger.get(command)
    if record is None:
        return dict(research_owner='DELTA-01',implementation_owner=None,resource=None,
                    symbols=[],contract_tests=[],live_tests=[],assignment_required=command is not None,
                    research_issue=mapping.get('DELTA-01',{}).get('html_url'))
    research=record['research_owner'];implementation=record['implementation_owner']
    return dict(research_owner=research,implementation_owner=implementation,resource=record['resource'],
                research_issue=mapping.get(research,{}).get('html_url'),
                implementation_issue=mapping.get(implementation,{}).get('html_url'),
                symbols=record['implementation']['symbols'],
                contract_tests=[t for g in record['parameter_sets'] for t in g['contract_tests']],
                live_tests=[t for g in record['parameter_sets'] for t in g['live_tests']],
                implementation_status=record['implementation']['status'],
                contract_status=record['contract']['status'],live_status=record['live']['status'],assignment_required=False)


def compare(before, after, coverage, mapping=None, contracts_before=(), contracts_after=()):
    old,old_warnings=normalize(before);new,new_warnings=normalize(after)
    ledger=keyed(coverage['records'])
    mapping=mapping or {}
    changes=[]
    def add(path,left,right,classification,command=None,refs=None):
        severity='review'
        if len(path)==2 and path[0]=='commands':
            severity='additive' if left is MISSING else 'potentially-breaking'
        elif right is MISSING or (path[-1]=='mandatory' and right is True):
            severity='potentially-breaking'
        changes.append(dict(path=list(path),classification=classification,compatibility=severity,
                            command=command,old=evidence(left),new=evidence(right),
                            evidence=refs or {'before':'normalized-snapshot:'+('/'.join(path)),'after':'normalized-snapshot:'+('/'.join(path))},
                            impact=impact(command,ledger,mapping)))
    for path,left,right in differences(old,new):
        category=path[0]
        classification='documentation' if category=='documentation' else 'metadata'
        if category in {'unknown','opaque_sources'}:
            classification='unclassified'
        command=path[1] if len(path)>1 and category in {'commands','candidates','configuration'} and path[1]!='$shared' else None
        add(path,left,right,classification,command)
    old_contracts=contract_model(contracts_before);new_contracts=contract_model(contracts_after)
    for name in sorted(set(old_contracts)|set(new_contracts)):
        left,right=old_contracts.get(name),new_contracts.get(name)
        if left and right and left['classification']!=right['classification']:
            add(('contracts',name,'classification'),left['classification'],right['classification'],'unclassified',name)
        if left and right and left['evidence']!=right['evidence']:
            add(('contracts',name,'evidence'),left['evidence'],right['evidence'],'metadata',name,
                {'before':left['evidence'],'after':right['evidence']})
        classification=(right or left)['classification']
        for path,old_value,new_value in differences(left['facts'] if left else MISSING,right['facts'] if right else MISSING):
            add(('contracts',name)+path,old_value,new_value,classification,name,
                {'before':left['evidence'] if left else None,'after':right['evidence'] if right else None})
    warnings=[dict(side='before',**w) for w in old_warnings]+[dict(side='after',**w) for w in new_warnings]
    # Platform differences are visible because absence in one runtime is not removal proof.
    if {k:v for k,v in before['runtime_environment'].items() if k!='module_version'}!={k:v for k,v in after['runtime_environment'].items() if k!='module_version'}:
        warnings.append(dict(path='runtime_environment',reason='Observation environments differ; review platform/session effects'))
    if before['package_sha256']!=after['package_sha256']:
        warnings.append(dict(path='package_sha256',reason='Package bytes changed; metadata alone cannot establish binary or remote behaviour equivalence'))
    classification_counts={kind:sum(c['classification']==kind for c in changes) for kind in ['documentation','metadata','observed-behaviour','unclassified']}
    return dict(schema_version=1,status='needs-review' if changes or warnings else 'no-delta',
                before={'version':before['module_version'],'channel':'preview' if '-' in before['module_version'] else 'stable','package_sha256':before['package_sha256'],'snapshot_canonical_sha256':hashlib.sha256(canonical(before).encode()).hexdigest()},
                after={'version':after['module_version'],'channel':'preview' if '-' in after['module_version'] else 'stable','package_sha256':after['package_sha256'],'snapshot_canonical_sha256':hashlib.sha256(canonical(after).encode()).hexdigest()},
                counts=classification_counts,changes=changes,warnings=warnings,
                limitations=['Metadata changes do not prove remote behaviour changes.',
                             'Opaque source changes require contract research; no route or authorization semantics are guessed.',
                             'Removed/added names are not automatically interpreted as renames.',
                             'Observations and documentation are explicitly supplied evidence, not claims of lab qualification.'])


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--before',type=pathlib.Path,required=True)
    parser.add_argument('--after',type=pathlib.Path,required=True)
    parser.add_argument('--contracts-before',type=pathlib.Path)
    parser.add_argument('--contracts-after',type=pathlib.Path)
    parser.add_argument('--output',type=pathlib.Path,required=True)
    args=parser.parse_args()
    def read(path):
        opener=gzip.open if path.suffix=='.gz' else open
        with opener(path,'rb') as stream:
            raw=stream.read(64*1024*1024+1)
        if len(raw)>64*1024*1024:
            raise DiscoveryError('JSON input exceeds 64 MiB')
        return json.loads(raw.decode('utf-8'))
    try:
        result=compare(read(args.before),read(args.after),read(ROOT/'contracts/coverage/8.1.0.json'),
            read(ROOT/'docs/implementation/backlog/publication/mapping.json'),
            read(args.contracts_before) if args.contracts_before else [],read(args.contracts_after) if args.contracts_after else [])
        publish_snapshot(args.output,result)
        print(json.dumps({'status':result['status'],'changes':len(result['changes']),'warnings':len(result['warnings'])}))
    except (DiscoveryError,OSError,ValueError,KeyError,TypeError,AttributeError) as exc:
        print(json.dumps({'status':'incomplete','error':str(exc)}))
        return 2
    return 0 if result['status']=='no-delta' else 1


if __name__=='__main__':
    raise SystemExit(main())
