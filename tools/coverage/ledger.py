"""Validate capability ownership and evidence states; never infer coverage from discovery."""
import argparse
import collections
import json
import pathlib
import re

ROOT = pathlib.Path(__file__).resolve().parents[2]
BASELINE = ROOT/'contracts/powershell/8.1.0'
TELEMETRY = 'invoke-csinternalpstelemetry'
# Commands approved for exclusion, each with its recorded reason. Anything
# else in the baseline must be implemented or stay open.
APPROVED_EXCLUSIONS = {
    TELEMETRY: 'Approved module telemetry plumbing exclusion: docs/implementation/backlog/issues/INDEX.md',
    'set-csnotifycache': ('Approved: no evidenced wire contract in 8.1.0 (manifest-only export whose request helper has no caller): '
                          'docs/implementation/ext/configuration-diagnostics-contract-notes.md'),
    'get-cslogininfo': ('Approved: returns the module\'s own access token and sends no service request; SDK credentials come from msteams/auth: '
                      'docs/implementation/ext/provisioning-diagnostics-contract-notes.md'),
}
COMMON = {
    'Debug': 'client logging configuration',
    'Verbose': 'client logging configuration',
    'ErrorAction': 'Go error returns; callers decide whether to continue',
    'ErrorVariable': 'Go error returns',
    'WarningAction': 'client logging configuration',
    'WarningVariable': 'caller logging sink',
    'InformationAction': 'client logging configuration',
    'InformationVariable': 'caller logging sink',
    'OutVariable': 'Go return values',
    'OutBuffer': 'caller-owned iteration/buffering',
    'PipelineVariable': 'caller-owned iteration',
    'ProgressAction': 'optional operation progress reporting',
    'WhatIf': 'caller-owned planning; no write is sent for a preview',
    'Confirm': 'caller-owned approval before invoking a mutation',
}


def read(path):
    return json.loads(path.read_text(encoding='utf-8'))


def seed(reconciliation, runtime):
    signatures = {c['name'].casefold(): (i, c) for i, c in enumerate(runtime['commands'])}
    records = []
    for candidate in reconciliation['records']:
        name = candidate['name']
        kind = 'management'
        if name.casefold() == TELEMETRY:
            kind = 'excluded'
        elif candidate['disposition'] == 'session-equivalent':
            kind = 'session'
        elif candidate['disposition'] == 'local-builder-candidate':
            kind = 'local-builder'
        elif not candidate['implementation_owner']:
            kind = 'research-candidate'
        sets = []
        if name.casefold() in signatures:
            index, command = signatures[name.casefold()]
            for group_index, group in enumerate(command['parameter_sets']):
                sets.append(dict(name=group['name'],
                    metadata_ref=f'contracts/powershell/8.1.0/runtime.json#/commands/{index}/parameter_sets/{group_index}',
                    symbols=[], contract_tests=[], live_tests=[]))
        records.append(dict(name=name, kind=kind, resource=candidate['family'],
            implementation_owner=candidate['implementation_owner'],
            research_owner='EXT-01' if kind=='research-candidate' else candidate['research_owner'],
            metadata_ref=f'contracts/powershell/8.1.0/reconciliation.json#/records/{len(records)}',
            parameter_sets=sets,
            implementation=dict(status='not-started', unsupported_stub=False, symbols=[]),
            contract=dict(status='metadata-only', evidence=[]),
            live=dict(status='not-run', reason='Tenant qualification has not run.', evidence=[]),
            exclusion='Approved module telemetry plumbing exclusion: docs/implementation/backlog/issues/INDEX.md' if kind=='excluded' else None))
    return dict(schema_version=1, baseline_version='8.1.0',
                common_parameters=COMMON,
                session_mapping='Native configuration, provider selection and cache lifecycle; AUTH-06. No PowerShell session or remote token revocation is implied.',
                records=records)


def resolve_reference(root, reference, cache=None):
    if not isinstance(reference, str) or not reference:
        raise ValueError('missing reference')
    path, _, pointer = reference.partition('#')
    if pathlib.Path(path).is_absolute():
        raise ValueError('reference must be repository-relative')
    target = (root/path).resolve()
    if not target.is_relative_to(root.resolve()) or not target.is_file():
        raise ValueError('reference is missing or outside repository: '+path)
    if not pointer:
        return target
    if cache is None:
        cache = {}
    if target not in cache:
        cache[target] = read(target)
    value = cache[target]
    if not pointer.startswith('/'):
        raise ValueError('invalid JSON pointer')
    for part in pointer[1:].split('/'):
        part = part.replace('~1','/').replace('~0','~')
        value = value[int(part)] if isinstance(value,list) else value[part]
    return value


def validate_symbol(root, ref, test=False, acceptance=False):
    if not isinstance(ref, dict) or set(ref) != {'file','name'}:
        raise ValueError('symbol requires exactly file and name')
    path = resolve_reference(root, ref['file'])
    if not isinstance(path, pathlib.Path) or path.suffix != '.go':
        raise ValueError('symbol must reference a Go source file')
    name = ref['name']
    if not isinstance(name,str) or not re.fullmatch(r'[A-Za-z_]\w*(?:\.[A-Za-z_]\w*)?',name):
        raise ValueError('invalid Go symbol name')
    source = path.read_text(encoding='utf-8')
    # Structural lookup only; implementation review and executable tests prove semantics.
    if '.' in name:
        receiver, method = name.split('.')
        # Generic receivers carry type parameters: func (t *Templates[T]) List(
        pattern = r'(?m)^func\s+\(\s*\w+\s+\*?'+re.escape(receiver)+r'(?:\[[\w\s,]+\])?\s*\)\s+'+re.escape(method)+r'\s*\('
    else:
        pattern = r'(?m)^(?:func|type)\s+'+re.escape(name)+r'\b'
    if not re.search(pattern, source):
        raise ValueError('stale Go symbol: '+name)
    prefixes = ('TestAcc',) if acceptance else ('TestUnit', 'Example')
    if test and (not path.name.endswith('_test.go') or not name.startswith(prefixes)):
        raise ValueError('test reference must identify the correct TestUnit/TestAcc category')


def validate_shape(value, schema, definitions, path='$'):
    """Validate the JSON Schema subset used by the checked-in ledger schema."""
    if '$ref' in schema:
        schema=definitions[schema['$ref'].removeprefix('#/$defs/')]
    errors=[]
    types=schema.get('type',[])
    if isinstance(types,str):
        types=[types]
    matches={'object':isinstance(value,dict),'array':isinstance(value,list),
             'string':isinstance(value,str),'integer':type(value) is int,
             'boolean':type(value) is bool,'null':value is None}
    if types and not any(matches[t] for t in types):
        return [path+': invalid type']
    if 'const' in schema and value!=schema['const']:
        errors.append(path+': invalid constant')
    if 'enum' in schema and value not in schema['enum']:
        errors.append(path+': invalid enum value')
    if isinstance(value,str) and len(value)<schema.get('minLength',0):
        errors.append(path+': empty value')
    if isinstance(value,dict):
        for field in schema.get('required',[]):
            if field not in value:
                errors.append(path+': missing '+field)
        for key,item in value.items():
            child=schema.get('properties',{}).get(key,schema.get('additionalProperties',True))
            if child is False:
                errors.append(path+': unexpected '+key)
            elif isinstance(child,dict):
                errors.extend(validate_shape(item,child,definitions,path+'.'+key))
    if isinstance(value,list):
        for i,item in enumerate(value):
            errors.extend(validate_shape(item,schema['items'],definitions,f'{path}[{i}]'))
    return errors


def validate(ledger, reconciliation, runtime, known_tickets, root=ROOT, require_complete=False):
    schema=read(ROOT/'contracts/coverage/schema.json')
    shape_errors=validate_shape(ledger,schema,schema['$defs'])
    if shape_errors:
        return dict(result='failed',errors=[dict(command=None,message=e) for e in shape_errors],release_ready=False)
    errors = []
    cache = {}
    def resolve(reference):
        return resolve_reference(root, reference, cache)
    def fail(name, message):
        errors.append(dict(command=name, message=message))
    if ledger.get('schema_version') != 1 or ledger.get('baseline_version') != '8.1.0':
        fail(None,'unsupported ledger schema or baseline')
    if ledger.get('common_parameters') != COMMON or not ledger.get('session_mapping'):
        fail(None,'common-parameter or session mappings are missing/changed without schema review')
    expected = {r['name'].casefold():r for r in reconciliation['records']}
    observed_refs = {c['name'].casefold(): {g['name']: f'contracts/powershell/8.1.0/runtime.json#/commands/{i}/parameter_sets/{j}'
                     for j,g in enumerate(c['parameter_sets'])} for i,c in enumerate(runtime['commands'])}
    observed = {r['name'].casefold():r for r in runtime['commands']}
    seen = set()
    complete = 0
    for record in ledger.get('records', []):
        if not isinstance(record,dict):
            fail(None,'ledger record must be an object')
            continue
        try:
            required = {'name','kind','resource','implementation_owner','research_owner','metadata_ref','parameter_sets','implementation','contract','live','exclusion'}
            if set(record)!=required:
                fail(record.get('name'),'record fields do not match schema')
            for field, keys in [('implementation',{'status','unsupported_stub','symbols'}),('contract',{'status','evidence'}),('live',{'status','reason','evidence'})]:
                if not isinstance(record.get(field),dict) or set(record[field])!=keys:
                    fail(record.get('name'),field+' fields do not match schema')
            name = record['name']
            key = name.casefold()
            if key in seen:
                fail(name,'duplicate command ownership')
            seen.add(key)
            if key not in expected:
                fail(name,'command is not in the reconciled baseline')
            kind = record['kind']
            if kind not in {'management','session','local-builder','research-candidate','excluded'}:
                fail(name,'invalid disposition')
            if kind=='research-candidate' and key in expected and expected[key]['implementation_owner']:
                fail(name,'owned management command cannot become an unassigned research candidate')
            if kind=='excluded' and record.get('exclusion')!=APPROVED_EXCLUSIONS.get(key):
                fail(name,'unapproved exclusion of a management candidate')
            if record.get('exclusion') and kind!='excluded':
                fail(name,'only an excluded record may carry an exclusion')
            if record.get('research_owner') not in known_tickets:
                fail(name,'missing or unknown research owner')
            if kind not in {'excluded','research-candidate'} and record.get('implementation_owner') not in known_tickets:
                fail(name,'missing or unknown implementation owner')
            if not record.get('resource'):
                fail(name,'missing resource mapping')
            evidence = resolve(record['metadata_ref'])
            if not isinstance(evidence, dict) or evidence.get('name','').casefold()!=key:
                fail(name,'metadata reference points to another command')
            groups = record['parameter_sets']
            group_names = [g['name'] for g in groups]
            if len(group_names)!=len(set(group_names)):
                fail(name,'duplicate parameter-set mapping')
            if key in observed and set(group_names)!={g['name'] for g in observed[key]['parameter_sets']}:
                fail(name,'parameter sets differ from runtime metadata')
            impl, contract, live = record['implementation'], record['contract'], record['live']
            if impl['status'] not in {'not-started','in-progress','implemented','unsupported'}:
                fail(name,'invalid implementation status')
            if not isinstance(impl.get('unsupported_stub'),bool):
                fail(name,'unsupported-stub flag must be explicit')
            if contract['status'] not in {'metadata-only','passed','failed'}:
                fail(name,'invalid contract status')
            if live['status'] not in {'not-run','passed','failed','skipped','unavailable'}:
                fail(name,'invalid live status')
            for symbol in impl['symbols']:
                validate_symbol(root,symbol)
            for group in groups:
                if set(group)!={'name','metadata_ref','symbols','contract_tests','live_tests'}:
                    fail(name,'parameter-set fields do not match schema')
                if key in observed_refs and group['metadata_ref']!=observed_refs[key].get(group['name']):
                    fail(name,'parameter-set reference is not the observed command signature')
                signature = resolve(group['metadata_ref'])
                if not isinstance(signature,dict) or signature.get('name')!=group['name']:
                    fail(name,'parameter-set evidence is missing or mismatched')
                for symbol in group['symbols']:
                    validate_symbol(root,symbol)
                for test in group['contract_tests']:
                    validate_symbol(root,test,test=True)
                for test in group['live_tests']:
                    validate_symbol(root,test,test=True,acceptance=True)
            for ref in contract['evidence']+live['evidence']:
                resolve(ref)
            if contract['status']=='passed':
                for ref in contract['evidence']:
                    result = resolve(ref)
                    if isinstance(result,pathlib.Path) and result.suffix=='.json':
                        result=read(result)
                    if not isinstance(result,dict) or result.get('outcome')!='passed' or not result.get('revision'):
                        fail(name,'contract pass requires a successful versioned test run')
            implemented = impl['status']=='implemented'
            if implemented and (impl['unsupported_stub'] or not impl['symbols'] or not groups or any(not g['symbols'] for g in groups)):
                fail(name,'unsupported/incomplete stub cannot count as implemented')
            if contract['status']=='passed' and (not implemented or not contract['evidence'] or not groups or any(not g['contract_tests'] for g in groups)):
                fail(name,'contract pass requires implementation, evidence and tests for every parameter set')
            if live['status']=='passed':
                if contract['status']!='passed' or not live['evidence'] or any(not g['live_tests'] for g in groups):
                    fail(name,'live pass requires contract pass, evidence and live tests')
                for ref in live['evidence']:
                    result = resolve(ref)
                    if isinstance(result,pathlib.Path) and result.suffix=='.json':
                        result=read(result)
                    if not isinstance(result,dict) or result.get('outcome')!='passed' or not result.get('revision'):
                        fail(name,'skipped/failed/missing run cannot count as live pass')
            elif not live.get('reason'):
                fail(name,'non-passing live state requires a reason')
            if kind=='excluded' or (implemented and contract['status']=='passed' and kind!='research-candidate'):
                complete += 1
            elif require_complete:
                fail(name,'release requires implemented and contract-tested coverage with resolved scope')
        except (KeyError,TypeError,ValueError,IndexError,AttributeError) as exc:
            fail(record.get('name'),str(exc))
    for key in sorted(set(expected)-seen):
        fail(expected[key]['name'],'orphan candidate: no ledger record')
    return dict(result='failed' if errors else 'passed', errors=errors,
                candidates=len(expected), records=len(seen), complete=complete,
                release_ready=not errors and complete==len(expected),
                implemented=sum(r['implementation']['status']=='implemented' for r in ledger['records']),
                contract_tested=sum(r['contract']['status']=='passed' for r in ledger['records']),
                excluded=sum(r['kind']=='excluded' for r in ledger['records']),
                live_counts=dict(collections.Counter(r.get('live',{}).get('status','missing') for r in ledger.get('records',[]) if isinstance(r,dict) and isinstance(r.get('live'),dict))))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--ledger',type=pathlib.Path,default=ROOT/'contracts/coverage/8.1.0.json')
    parser.add_argument('--require-complete',action='store_true')
    args = parser.parse_args()
    try:
        result = validate(read(args.ledger),read(BASELINE/'reconciliation.json'),read(BASELINE/'runtime.json'),
                          {r['id'] for r in read(ROOT/'docs/implementation/backlog/manifest.json')['issues'] if r['type']=='task'},
                          require_complete=args.require_complete)
    except (OSError,ValueError,TypeError,KeyError) as exc:
        result = dict(result='failed',errors=[dict(command=None,message=str(exc))])
    print(json.dumps(result,indent=2))
    return 0 if result['result']=='passed' else 1


if __name__=='__main__':
    raise SystemExit(main())
