"""Validate versioned resource contracts against pinned command and oracle evidence."""
import json
import re
import sys
from urllib.parse import urlsplit

import verify_models as evidence

CONTRACT = evidence.ROOT / 'contracts/endpoints/8.1.0/application-access-policies.json'


def fixture(reference):
    evidence.reference(reference['artifact'])
    value = evidence.read(evidence.ROOT / reference['artifact'])
    matches = [case for case in value['cases'] if case['name'] == reference['case']]
    evidence.require(len(matches) == 1, 'Resource fixture case is missing or ambiguous')
    evidence.require(value['baseline_version'] == '8.1.0', 'Resource fixture baseline differs')
    return matches[0]


def validate(value):
    require = evidence.require
    require(value['schema_version'] == 1 and value['baseline_version'] == '8.1.0', 'Wrong resource schema/baseline')
    require(value['resource'] == 'application_access_policy', 'Wrong resource family')
    require(value['status'] == 'in-progress' and value['open_questions'], 'Conditional resource contract must retain incomplete state')
    require(value['package_sha256'] == evidence.read(evidence.CONTRACT)['package_sha256'], 'Resource package differs')
    evidence.reference(value['evidence'])
    evidence.reference(value['profile']['evidence'])
    require(value['profile']['name'] == 'config-api-generated'
            and value['profile']['audience'] == '48ac35b8-9aa8-4d74-927d-1f4a14a0b239', 'Resource profile/audience differs')
    require(value['profile']['discovery']['cross_origin_send'] is False, 'Discovery must not authorize cross-origin sends')
    reconciliation = evidence.read(evidence.ROOT / 'contracts/powershell/8.1.0/reconciliation.json')['records']
    assigned = {c['name'] for c in reconciliation if c['research_owner'] == 'POL-01' and c['family'] == value['resource']}
    commands = {c['name']: c for c in value['commands']}
    require(set(commands) == assigned and len(commands) == len(value['commands']), 'Resource command assignment differs')
    operations = {o['id']: o for o in value['operations']}
    require(len(operations) == len(value['operations']), 'Duplicate resource operation')
    surface = value['service_surface']
    for method, kind in [('create', 'create'), ('get_by_identity', 'read'), ('update_by_identity', 'update'), ('delete_by_identity', 'delete'), ('list', 'list')]:
        require(surface[method] in operations and operations[surface[method]]['kind'] == kind, 'CRUD/list surface is missing or misclassified')
    require(all(name in operations and operations[name]['kind'] == 'action' for name in surface['additional_operations']), 'Additional action is missing or misclassified')
    surface_refs = {surface[name] for name in ('create', 'get_by_identity', 'update_by_identity', 'delete_by_identity', 'list')}
    require(surface_refs | set(surface['additional_operations']) == set(operations), 'Resource surface does not expose every operation')
    runtime = evidence.read(evidence.ROOT / 'contracts/powershell/8.1.0/runtime.json')['commands']
    referenced = set()
    sets = 0
    for name, command in commands.items():
        index, metadata = next((i, c) for i, c in enumerate(runtime) if c['name'] == name)
        expected = [dict(name=s['name'], metadata_ref=f'contracts/powershell/8.1.0/runtime.json#/commands/{index}/parameter_sets/{j}') for j, s in enumerate(metadata['parameter_sets'])]
        actual = [dict(name=s['name'], metadata_ref=s['metadata_ref']) for s in command['parameter_sets']]
        require(actual == expected, 'Resource runtime-set mapping differs')
        require(command['composition'], 'Command lacks composition semantics')
        evidence.reference(command['evidence'])
        for parameter_set in command['parameter_sets']:
            names = parameter_set['operation_refs']
            require(names and all(name in operations for name in names), 'Missing resource operation binding')
            referenced.update(names)
        sets += len(expected)
    require(referenced == set(operations), 'Resource operation is unbound')
    for operation in operations.values():
        evidence.reference(operation['evidence'])
        require(operation['profile'] == value['profile']['name'], 'Operation uses a different profile')
        require(operation['replay_safe'] is (operation['method'] == 'GET'), 'Mutation must not become replay-safe')
        require(operation['path'].startswith('/') and not operation['path'].startswith('//'), 'Resource path must be relative')
        pattern = re.sub(r'\\\{[^}]+\\\}', '[^/?#]+', re.escape(operation['path']))
        require(operation['request_cases'], 'Operation has no pinned request case')
        statuses = set()
        for reference in operation['request_cases']:
            case = fixture(reference)
            require(case['method'] == operation['method'], 'Operation verb differs from pinned capture')
            require(re.fullmatch(pattern, urlsplit(case['generated_url']).path) is not None, 'Operation route differs from pinned capture')
            require(case['response_status'] in operation['response']['statuses'], 'Operation status differs from pinned capture')
            statuses.add(case['response_status'])
        require(statuses == set(operation['response']['statuses']), 'Success status lacks a pinned dispatch case')
    for reference in value['shared_cases'] + value['input_contract']['app_ids']['cases']:
        fixture(reference)
    return len(commands), sets, len(operations)


def main():
    try:
        commands, parameter_sets, operations = validate(evidence.read(CONTRACT))
        print(json.dumps(dict(result='passed', resource='application_access_policy', commands=commands,
                              parameter_sets=parameter_sets, operations=operations, contract_status='in-progress')))
    except (ValueError, OSError, KeyError, StopIteration) as exc:
        sys.exit(f'Resource contract verification failed: {exc}')


if __name__ == '__main__':
    main()
