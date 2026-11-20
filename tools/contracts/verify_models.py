"""Validate CORE-01 inventory and reproduce pinned model/request observations offline."""
import argparse
import copy
import json
import os
import pathlib
import subprocess
import sys
import tempfile

ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / 'tools/baseline'))
import baseline  # noqa: E402

CONTRACT = ROOT / 'contracts/endpoints/8.1.0/core.json'
FIXTURES = {
    'core-team-serialization.json': 'capture-team-serialization.ps1',
    'core-model-serialization.json': 'capture-core-models.ps1',
    'policy-model-serialization.json': 'capture-policy-models.ps1',
    'generated-request-dispatch.json': 'capture-generated-requests.ps1',
}


def read(path):
    return json.loads(path.read_text(encoding='utf-8-sig'))


def require(condition, message):
    if not condition:
        raise ValueError(message)


def reference(path):
    require(isinstance(path, str), 'Evidence reference must be a path')
    target = (ROOT / path).resolve()
    require(target.is_relative_to(ROOT) and target.is_file(), 'Evidence reference is missing or outside repository')


def validate_inventory(value):
    require(value['baseline_version'] == '8.1.0', 'Wrong baseline')
    require(value['status'] == 'in-progress', 'Inventory cannot claim accepted contract status')
    require(value['package_sha256'] == read(ROOT / 'contracts/endpoints/8.1.0/fixtures/core-team-serialization.json')['package_sha256'], 'Package identity differs')
    reconciliation = read(ROOT / 'contracts/powershell/8.1.0/reconciliation.json')['records']
    assigned = {c['name']: c for c in reconciliation if c['research_owner'] == 'CORE-01'}
    commands = value['commands']
    require(len(commands) == len(assigned) and {c['name'] for c in commands} == set(assigned), 'CORE-01 command assignment differs')
    operations = {o['id']: o for o in value['operations']}
    require(len(operations) == len(value['operations']), 'Duplicate operation identity')
    for operation in operations.values():
        require(operation['kind'] in {'create', 'read', 'update', 'delete', 'list', 'action'}, 'Unknown operation kind')
        require(operation['method'] in {'GET', 'POST', 'PATCH', 'PUT', 'DELETE'}, 'Unknown method')
        require(operation['path'].startswith('/') and not operation['path'].startswith('//') and '://' not in operation['path'], 'Operation path must be relative to selected backend')
        reference(operation['evidence'])
    runtime = read(ROOT / 'contracts/powershell/8.1.0/runtime.json')['commands']
    for command in commands:
        name = command['name']
        require(command['family'] == assigned[name]['family'] and command['implementation_owner'] == assigned[name]['implementation_owner'], 'Command ownership differs')
        index, metadata = next((i, c) for i, c in enumerate(runtime) if c['name'] == name)
        expected_sets = [dict(name=s['name'], metadata_ref=f'contracts/powershell/8.1.0/runtime.json#/commands/{index}/parameter_sets/{j}') for j, s in enumerate(metadata['parameter_sets'])]
        require(command['parameter_sets'] == expected_sets, 'Parameter-set evidence differs')
        require(all(o in operations for o in command['operation_refs']), 'Missing operation reference')
        require(all(q in value['open_questions'] for q in command['open_questions']), 'Missing question reference')
        if command['disposition'] == 'settled-local-builder':
            require(not command['operation_refs'], 'Local builder must not have HTTP operations')
            require(name in {'New-CsTeamsHiddenTemplate', 'New-CsTeamsHiddenMeetingTemplate'}, 'Unexpected local-builder disposition')
        else:
            require(command['disposition'] == 'static-evidence-in-progress' and command['open_questions'] and command['operation_refs'], 'Remote inventory must retain incomplete-evidence state')
        reference(command['evidence'])
    return len(commands), len(operations)


def differences(expected, actual, path=''):
    """Return structural paths only; never print values from a capture."""
    if type(expected) is not type(actual):
        return [path or '/']
    if isinstance(expected, dict):
        result = [path + '/' + key for key in expected.keys() ^ actual.keys()]
        for key in expected.keys() & actual.keys():
            result.extend(differences(expected[key], actual[key], path + '/' + key))
        return sorted(result)
    if isinstance(expected, list):
        if len(expected) != len(actual):
            return [path]
        return [difference for i, (left, right) in enumerate(zip(expected, actual)) for difference in differences(left, right, f'{path}/{i}')]
    return [] if expected == actual else [path or '/']


def compare(expected, actual):
    left, right = copy.deepcopy(expected), copy.deepcopy(actual)
    # Capture date is provenance, not model behavior. All other fields/order stay exact.
    left.pop('captured_at', None)
    right.pop('captured_at', None)
    return differences(left, right)


def capture(module, directory, pwsh):
    # The development oracle must not inherit tenant or CI cloud credentials.
    allowed = {'PATH', 'SystemRoot', 'WINDIR', 'TEMP', 'TMP', 'TMPDIR', 'DOTNET_ROOT', 'LANG'}
    environment = {k: v for k, v in os.environ.items() if k in allowed}
    counts = dict(models=0, requests=0)
    for fixture, script in FIXTURES.items():
        output = directory / fixture
        result = subprocess.run([pwsh, '-NoLogo', '-NoProfile', '-NonInteractive', '-File',
                                 str(ROOT / 'tools/contracts' / script), '-ModulePath', str(module),
                                 '-OutputPath', str(output)], capture_output=True, text=True,
                                timeout=120, env=environment)
        require(result.returncode == 0 and output.is_file(), f'Model oracle failed: {script}; exit {result.returncode}')
        expected = read(ROOT / 'contracts/endpoints/8.1.0/fixtures' / fixture)
        delta = compare(expected, read(output))
        require(not delta, f'Model observation differs: {fixture}; paths: {", ".join(delta[:12])}')
        category = 'requests' if expected['qualification'] == 'offline-generated-request-and-dispatch' else 'models'
        counts[category] += len(expected['cases'])
    return counts


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    mode = parser.add_mutually_exclusive_group()
    mode.add_argument('--acquire', action='store_true', help='Download, verify SHA512 and extract the pinned package before recapture')
    mode.add_argument('--module', type=pathlib.Path, help='Recapture from an already verified extracted package')
    parser.add_argument('--pwsh', default='pwsh')
    args = parser.parse_args()
    try:
        commands, operations = validate_inventory(read(CONTRACT))
        counts = dict(models=0, requests=0)
        if args.acquire or args.module:
            with tempfile.TemporaryDirectory(prefix='msteams-contracts-') as directory:
                work = pathlib.Path(directory)
                module = args.module.resolve() if args.module else work / 'module'
                if args.acquire:
                    archive = work / 'module.zip'
                    baseline.fetch(f'{baseline.GALLERY}/api/v2/package/MicrosoftTeams/{baseline.VERSION}', archive)
                    baseline.verify_archive(archive, {'sha512': baseline.PIN}, baseline.PIN)
                    baseline.extract_archive(archive, module)
                counts = capture(module, work, args.pwsh)
        print(json.dumps(dict(result='passed', commands=commands, operations=operations,
                              recaptured_model_cases=counts['models'], recaptured_request_cases=counts['requests'], contract_status='in-progress')))
    except (ValueError, OSError, subprocess.TimeoutExpired) as exc:
        parser.exit(1, f'Contract evidence verification failed: {exc}\n')


if __name__ == '__main__':
    main()
