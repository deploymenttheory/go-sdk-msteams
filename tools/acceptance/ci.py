"""Run TestAcc tests and print their results."""
import argparse
import importlib.util
import io
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import tarfile
import tempfile

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
PROOF_MARKERS = [
    'policy create/read-back: passed',
    'policy update/read-back and omission: passed',
    'policy list: passed',
    'policy delete/absence: passed',
]
ORACLE_MARKER = 'PowerShell 8.1.0 / Go policy read parity: passed'
USER_ASSIGNMENT_MARKERS = [
    'user policy assignment/read-back: passed',
    'user policy clear/restoration: passed',
    'user assignment cleanup: passed',
]
USER_ASSIGNMENT_ORACLE = 'PowerShell 8.1.0 / Go user assignment parity: passed'
PROOFS = {
    'owned-crud': (PROOF_MARKERS, ORACLE_MARKER, 2),
    'user-assignment': (USER_ASSIGNMENT_MARKERS, USER_ASSIGNMENT_ORACLE, 3),
}


class AcceptanceFailure(ValueError):
    """A diagnostic containing public test/configuration names only."""


def write_json(path, value):
    path.parent.mkdir(parents=True, exist_ok=True)
    fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
    with os.fdopen(fd, 'w') as stream:
        json.dump(value, stream, indent=2)
        stream.write('\n')


def write_private(path, value):
    fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    with os.fdopen(fd, 'w') as stream:
        stream.write(value)


def check_names(root=ROOT):
    errors = []
    counts = {'unit': 0, 'acceptance': 0}
    acceptance = set()
    for path in root.rglob('*_test.go'):
        if '.git' in path.parts:
            continue
        for name in re.findall(r'^func (Test\w+)\(', path.read_text(), re.M):
            if name == 'TestMain':
                continue
            if not re.fullmatch(r'Test(?:Unit|Acc)(?:[A-Z0-9_]\w*)?', name):
                errors.append(str(path.relative_to(root)))
            else:
                counts['acceptance' if name.startswith('TestAcc') else 'unit'] += 1
                if name.startswith('TestAcc'):
                    acceptance.add(name)
    if errors:
        raise ValueError('test names must start TestUnit or TestAcc: ' + ', '.join(sorted(set(errors))))
    if root == ROOT:
        suites = json.loads((HERE / 'suites.json').read_text())
        selected = [name for suite in suites.values() for name in suite['tests']]
        pending = json.loads((HERE / 'pending.json').read_text())
        if len(selected) != len(set(selected)) or set(selected) & set(pending) or set(selected) | set(pending) != acceptance:
            raise ValueError('every TestAcc must have one CI suite or an explicit pending disposition')
        if any(not entry.get('reason') or not (ROOT / entry['file']).is_file() for entry in pending.values()):
            raise ValueError('pending acceptance requires an existing test file and a concrete reason')
    return counts


def user_fixture(values):
    try:
        fixture = json.loads(values['USER_FIXTURE_JSON'])
        valid = (set(fixture) == {'tenant_id', 'user_id', 'user_principal_name', 'display_name'}
                 and fixture['tenant_id'].lower() == values['TENANT_ID'].lower()
                 and re.fullmatch(r'[0-9a-fA-F]{8}(?:-[0-9a-fA-F]{4}){3}-[0-9a-fA-F]{12}', fixture['user_id'])
                 and re.fullmatch(r'go-sdk-msteams-acceptance@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}', fixture['user_principal_name'])
                 and fixture['display_name'] == 'go-sdk-msteams acceptance user')
    except (ValueError, TypeError, AttributeError, KeyError):
        valid = False
    if not valid:
        raise AcceptanceFailure('Invalid repository Actions secret: MSTEAMS_LAB_USER_FIXTURE_JSON')
    return fixture


def prepare(args):
    keys = ['TENANT_ID', 'CLIENT_ID', 'CLIENT_SECRET', 'CERTIFICATE_PEM', 'PRIVATE_KEY_PEM', 'RECOVERY_CERTIFICATE_PEM', 'CALL_HOLD_AUDIO_FILE_ID', 'USER_FIXTURE_JSON']
    values = {key: os.environ.get('MSTEAMS_LAB_' + key, '') for key in keys}
    missing = ['MSTEAMS_LAB_' + key for key, value in values.items() if not value.strip()]
    if missing:
        raise AcceptanceFailure('Missing repository Actions secrets: ' + ', '.join(missing))
    if not re.fullmatch(r'(?:[0-9a-fA-F]{32}|[0-9a-fA-F]{8}(?:-[0-9a-fA-F]{4}){3}-[0-9a-fA-F]{12})', values['CALL_HOLD_AUDIO_FILE_ID']):
        raise AcceptanceFailure('Invalid repository Actions secret: MSTEAMS_LAB_CALL_HOLD_AUDIO_FILE_ID')
    fixture = user_fixture(values)
    base = Path(os.environ['RUNNER_TEMP']).resolve()
    private = Path(tempfile.mkdtemp(prefix='msteams-acceptance-', dir=base))
    certificate = private / 'certificate.pem'
    key = private / 'private-key.pem'
    write_private(certificate, values['CERTIFICATE_PEM'])
    write_private(key, values['PRIVATE_KEY_PEM'])
    write_private(private / 'recovery-certificate.pem', values['RECOVERY_CERTIFICATE_PEM'])
    write_json(private / 'user-fixture.json', fixture)
    write_json(private / 'identity.json', {
        'tenant_id': values['TENANT_ID'], 'client_id': values['CLIENT_ID'],
        'client_secret': values['CLIENT_SECRET'],
        'call_hold_audio_file_id': values['CALL_HOLD_AUDIO_FILE_ID'],
        'certificate_file': str(certificate), 'private_key_file': str(key),
    })
    # Publish paths early so finalization also runs after a download/import failure.
    with Path(os.environ['GITHUB_ENV']).open('a') as stream:
        stream.write('MSTEAMS_LAB_PRIVATE_DIRECTORY=' + str(private) + '\n')
    spec = importlib.util.spec_from_file_location('pinned_baseline', ROOT / 'tools/baseline/baseline.py')
    baseline = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(baseline)
    metadata = args.metadata or private / 'gallery.xml'
    archive = args.archive or private / 'module.zip'
    if not args.metadata:
        baseline.fetch(f"{baseline.GALLERY}/api/v2/Packages(Id='MicrosoftTeams',Version='{baseline.VERSION}')", metadata, 1024 * 1024)
    if not args.archive:
        baseline.fetch(f'{baseline.GALLERY}/api/v2/package/MicrosoftTeams/{baseline.VERSION}', archive)
    info = baseline.gallery_metadata(metadata.read_bytes(), baseline.VERSION)
    baseline.verify_archive(archive, info, baseline.PIN)
    baseline.extract_archive(archive, private / 'module')
    print(f'Repository secrets loaded: {len(values)}. MicrosoftTeams 8.1.0 checksum verified.', flush=True)


def evaluate(events, expected, process_exit, require_crud=False, proof=None):
    proof = proof or ('owned-crud' if require_crud else None)
    proof_markers, oracle_marker, oracle_count = PROOFS[proof] if proof else ([], None, 0)
    outcomes = {}
    markers = {name: [] for name in expected}
    child_counts = {'pass': 0, 'fail': 0, 'skip': 0}
    http_statuses = {name: [] for name in expected}
    service_codes = {name: [] for name in expected}
    for event in events:
        test = event.get('Test', '')
        action = event.get('Action')
        if test in expected and action in {'pass', 'fail', 'skip'}:
            outcomes[test] = action
        if '/' in test and test.split('/', 1)[0] in expected and action in child_counts:
            child_counts[action] += 1
        owner = test.split('/', 1)[0]
        if owner in http_statuses and action == 'output':
            http_statuses[owner].extend(int(value) for value in re.findall(r'HTTP (\d{3})\b', event.get('Output', '')))
            service_codes[owner].extend(int(value) for value in re.findall(r'PolicyRP (\d+)\b', event.get('Output', '')))
        if test in markers and action == 'output':
            for marker in proof_markers + ([oracle_marker] if oracle_marker else []):
                if marker in event.get('Output', ''):
                    markers[test].append(marker)
    tests = [{'name': name, 'outcome': outcomes.get(name, 'not-run')} for name in expected]
    passed = bool(expected) and len(set(expected)) == len(expected) and process_exit == 0
    passed = passed and all(item['outcome'] == 'pass' for item in tests) and not child_counts['skip'] and not child_counts['fail']
    if proof:
        passed = passed and all(all(marker in markers[name] for marker in proof_markers) and markers[name].count(oracle_marker) == oracle_count for name in expected)
    return {
        'outcome': 'passed' if passed else 'failed', 'tests': tests,
        'child_outcomes': child_counts, 'process_exit': process_exit,
        'http_statuses': {name: sorted(set(values)) for name, values in http_statuses.items() if values},
        'service_error_codes': {name: sorted(set(values)) for name, values in service_codes.items() if values},
        'mutation_assertions_observed': {(name + ': ' if len(expected) > 1 else '') + marker: marker in markers[name] for name in expected for marker in proof_markers},
        'mutation_powershell_comparisons': sum(values.count(oracle_marker) for values in markers.values()) if oracle_marker else 0,
        'mutation_powershell_comparisons_by_test': {name: values.count(oracle_marker) for name, values in markers.items()} if proof else {},
        'expected_powershell_comparisons': oracle_count,
    }


def run(args):
    private = Path(os.environ['MSTEAMS_LAB_PRIVATE_DIRECTORY']).resolve()
    suites = json.loads((HERE / 'suites.json').read_text())
    kind = getattr(args, 'kind', None)
    if kind not in {None, 'reads', 'writes'} or kind and args.suite:
        raise ValueError('choose a suite selection or a read/write group')
    selected = [name for name, suite in suites.items()
                if (not args.suite or name in args.suite)
                and (not kind or bool(suite.get('mutation')) == (kind == 'writes'))]
    if not selected or args.suite and set(args.suite) - set(suites):
        raise ValueError('unknown acceptance suite')
    revision = subprocess.run(['git', 'rev-parse', 'HEAD'], cwd=ROOT, capture_output=True, text=True, check=True).stdout.strip()
    mode = os.environ.get('MSTEAMS_LAB_AUTH_METHOD', 'certificate')
    if mode not in {'certificate', 'client-secret'}:
        raise ValueError('unsupported acceptance authentication method')
    reports = []
    for name in selected:
        suite = suites[name]
        if suite.get('mutation') and (suite.get('proof') not in PROOFS or type(suite.get('journal_entries_per_test')) is not int or suite['journal_entries_per_test'] < 1):
            raise ValueError('mutation suite requires explicit checks and journal entry count')
        env = {key: value for key, value in os.environ.items() if not key.startswith('MSTEAMS_')}
        env.update({
            'MSTEAMS_LAB_READ_ONLY': '1', 'MSTEAMS_LAB_MUTATIONS': '1' if suite.get('mutation') else '0',
            'MSTEAMS_LAB_CONSEQUENTIAL': '0', 'MSTEAMS_LAB_TIMEOUT_SECONDS': '300',
            'MSTEAMS_LAB_IDENTITY_FILE': str(private / 'identity.json'),
            'MSTEAMS_LAB_USER_FIXTURE_FILE': str(private / 'user-fixture.json'),
            'MSTEAMS_LAB_MODULE_PATH': str(private / 'module'),
            'MSTEAMS_LAB_JOURNAL_DIRECTORY': str(private / 'journals'),
            'MSTEAMS_LAB_POWERSHELL_ORACLE': '1', 'MSTEAMS_LAB_AUTH_METHOD': mode,
            'MSTEAMS_LAB_ACCESS_PROFILE': 'teams-administrator', 'MSTEAMS_LIVE_ACCESS': '1',
        })
        pattern = '^(' + '|'.join(re.escape(test) for test in suite['tests']) + ')$'
        command = ['go', 'test', '-json', '-count=1', '-race', '-timeout=12m', '-run', pattern, suite['package']]
        previous_journals = set((private / 'journals').glob('*.json'))
        log = private / (name + '.jsonl')
        fd = os.open(log, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
        with os.fdopen(fd, 'w') as stream:
            result = subprocess.run(command, cwd=ROOT, env=env, stdout=stream, stderr=subprocess.STDOUT)
        events = []
        for line in log.read_text().splitlines():
            try:
                events.append(json.loads(line))
            except ValueError:
                pass
        report = evaluate(events, suite['tests'], result.returncode, proof=suite.get('proof'))
        if suite.get('mutation'):
            new_journals = set((private / 'journals').glob('*.json')) - previous_journals
            entries = [entry for path in new_journals for entry in json.loads(path.read_text())['entries']]
            report['cleanup_journal_entries'] = len(entries)
            report['cleanup_journal_entries_cleaned'] = sum(entry['state'] == 'cleaned' for entry in entries)
            report['expected_cleanup_journal_entries'] = len(suite['tests']) * suite['journal_entries_per_test']
            if len(entries) != report['expected_cleanup_journal_entries'] or report['cleanup_journal_entries_cleaned'] != len(entries):
                report['outcome'] = 'failed'
        report.update(suite=name, revision=revision, auth_method=mode, cloud='commercial', baseline_version='8.1.0',
                      powershell_oracle_enabled=True)
        reports.append(report)
        write_json(args.report_directory / (name + '.json'), report)
        print_report(report)
    summary = {'revision': revision, 'auth_method': mode, 'outcome': 'passed' if all(r['outcome'] == 'passed' for r in reports) else 'failed',
               'suites': [{'name': r['suite'], 'outcome': r['outcome']} for r in reports], 'coverage_claim': False}
    write_json(args.report_directory / 'summary.json', summary)
    write_summary(reports)
    return 0 if summary['outcome'] == 'passed' else 1


def print_report(report):
    passed = sum(test['outcome'] == 'pass' for test in report['tests'])
    print(f"{report['suite']} ({report['auth_method']}): {passed}/{len(report['tests'])} passed", flush=True)
    for test in report['tests']:
        if test['outcome'] == 'pass':
            continue
        statuses = report.get('http_statuses', {}).get(test['name'], [])
        detail = 'HTTP ' + ', '.join(map(str, statuses)) if statuses else test['outcome']
        message = test['name'] + ': ' + detail
        print('FAIL ' + message, flush=True)
        if os.environ.get('GITHUB_ACTIONS') == 'true':
            print('::error title=' + test['name'] + '::' + detail, flush=True)

    if report['suite'] == 'tenant-access-outcomes':
        for test in report['tests']:
            if test['outcome'] == 'pass':
                statuses = ', '.join(map(str, report.get('http_statuses', {}).get(test['name'], [])))
                codes = ', '.join(map(str, report.get('service_error_codes', {}).get(test['name'], [])))
                print(f"PASS {test['name']}: verified HTTP {statuses}; PolicyRP {codes}", flush=True)
    if report.get('process_exit', 0):
        print(f"FAIL {report['suite']}: go test exited {report['process_exit']}", flush=True)
    children = report.get('child_outcomes', {})
    for state in ['fail', 'skip']:
        if children.get(state, 0):
            print(f"FAIL {report['suite']}: {children[state]} child tests {state}", flush=True)
    assertions = report.get('mutation_assertions_observed', report.get('crud_assertions_observed', {}))
    for name, observed in assertions.items():
        if not observed:
            print('FAIL missing mutation check: ' + name.removesuffix(': passed'), flush=True)
    expected_comparisons = report.get('expected_powershell_comparisons', 2)
    for name, count in report.get('mutation_powershell_comparisons_by_test', report.get('crud_powershell_comparisons_by_test', {})).items():
        if count != expected_comparisons:
            print(f'FAIL {name} PowerShell comparisons: {count}/{expected_comparisons}', flush=True)
    entries = report.get('cleanup_journal_entries')
    cleaned = report.get('cleanup_journal_entries_cleaned', 0)
    expected_entries = report.get('expected_cleanup_journal_entries', entries)
    if entries is not None and (entries == 0 or entries != cleaned or entries != expected_entries):
        print(f'FAIL cleanup journals: {cleaned}/{entries} entries completed', flush=True)
        if entries != expected_entries:
            print(f'FAIL cleanup journal count: {entries}/{expected_entries} required entries', flush=True)


def write_summary(reports):
    lines = ['## TestAcc results', '', '| Auth | Suite | Result | Passed | Failed | Skipped | Not run |', '|---|---|---|---:|---:|---:|---:|']
    failed = []
    for report in reports:
        counts = {state: sum(test['outcome'] == state for test in report['tests']) for state in ['pass', 'fail', 'skip', 'not-run']}
        lines.append(f"| {report['auth_method']} | {report['suite']} | {report['outcome']} | {counts['pass']} | {counts['fail']} | {counts['skip']} | {counts['not-run']} |")
        for test in report['tests']:
            if test['outcome'] != 'pass':
                statuses = report.get('http_statuses', {}).get(test['name'], [])
                detail = 'HTTP ' + ', '.join(map(str, statuses)) if statuses else test['outcome']
                failed.append(f"| {report['auth_method']} | `{test['name']}` | {detail} |")
    if failed:
        lines += ['', '| Auth | Failed test | Result |', '|---|---|---|'] + failed
    access = []
    for report in reports:
        if report['suite'] != 'tenant-access-outcomes':
            continue
        for test in report['tests']:
            statuses = ', '.join(map(str, report.get('http_statuses', {}).get(test['name'], [])))
            codes = ', '.join(map(str, report.get('service_error_codes', {}).get(test['name'], [])))
            access.append(f"| {report['auth_method']} | `{test['name']}` | {test['outcome']} | {statuses} | {codes} |")
    if access:
        lines += ['', '### Tenant access outcomes', '', '| Auth | Denial check | Result | HTTP | PolicyRP |', '|---|---|---|---|---|'] + access
    if os.environ.get('GITHUB_STEP_SUMMARY'):
        with Path(os.environ['GITHUB_STEP_SUMMARY']).open('a') as stream:
            stream.write('\n'.join(lines) + '\n')


def summarize(args):
    reports = []
    for path in sorted(args.report_directory.rglob('*.json')):
        report = json.loads(path.read_text())
        if 'suite' in report and 'tests' in report:
            reports.append(report)
    if not reports:
        raise AcceptanceFailure('No TestAcc reports were produced. Check the setup steps in the authentication jobs.')
    suites = json.loads((HERE / 'suites.json').read_text())
    expected = {(mode, name) for mode in ['certificate', 'client-secret'] for name in suites}
    actual = [(report.get('auth_method'), report.get('suite')) for report in reports]
    complete = set(actual) == expected and len(actual) == len(expected)
    for mode, name in sorted(expected - set(actual)):
        print(f'FAIL missing TestAcc suite: {mode}/{name}', flush=True)
    if set(actual) - expected or len(actual) != len(set(actual)):
        print('FAIL unexpected or duplicate TestAcc suite reports', flush=True)
    for report in reports:
        print_report(report)
    write_summary(reports)
    return 0 if complete and all(report['outcome'] == 'passed' for report in reports) else 1


def finalize(args):
    private = Path(os.environ['MSTEAMS_LAB_PRIVATE_DIRECTORY']).resolve()
    base = Path(os.environ['RUNNER_TEMP']).resolve()
    if private.parent != base or not private.name.startswith('msteams-acceptance-'):
        raise ValueError('refusing cleanup outside the prepared private runner directory')
    payload = io.BytesIO()
    entries = []
    with tarfile.open(fileobj=payload, mode='w') as archive:
        for path in sorted((private / 'journals').glob('*.json')):
            if path.is_symlink() or not path.is_file():
                raise ValueError('invalid recovery journal')
            content = path.read_bytes()
            entries.extend(json.loads(content)['entries'])
            info = tarfile.TarInfo(path.name)
            info.size, info.mode = len(content), 0o600
            archive.addfile(info, io.BytesIO(content))
    args.report_directory.mkdir(parents=True, exist_ok=True)
    encrypted = args.report_directory / 'recovery.cms'
    result = subprocess.run(['openssl', 'cms', '-encrypt', '-aes-256-gcm', '-binary', '-outform', 'DER',
                             '-recip', str(private / 'recovery-certificate.pem'), '-keyopt', 'rsa_padding_mode:oaep',
                             '-keyopt', 'rsa_oaep_md:sha256', '-out', str(encrypted)], input=payload.getvalue(), capture_output=True)
    if result.returncode or not encrypted.is_file() or encrypted.stat().st_size == 0:
        raise ValueError('journal encryption failed; private recovery data retained')
    cleaned = sum(entry['state'] == 'cleaned' for entry in entries)
    write_json(args.report_directory / 'cleanup.json', {
        'journal_entries': len(entries), 'cleaned_entries': cleaned, 'unresolved_entries': len(entries) - cleaned,
        'recovery_encryption': 'CMS AES-256-GCM / RSA-OAEP-SHA256', 'recovery_private_key_in_ci': False,
    })
    shutil.rmtree(private)
    print(f'Cleanup: {cleaned}/{len(entries)} entries completed. Recovery journals encrypted; credential files removed.', flush=True)
    return 0 if cleaned == len(entries) else 1


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest='command', required=True)
    commands.add_parser('check-names')
    prep = commands.add_parser('prepare')
    prep.add_argument('--archive', type=Path)
    prep.add_argument('--metadata', type=Path)
    for name in ['run', 'finalize', 'summarize']:
        child = commands.add_parser(name)
        child.add_argument('--report-directory', type=Path, required=True)
        if name == 'run':
            child.add_argument('--suite', action='append')
            child.add_argument('--kind', choices=['reads', 'writes'])
    args = parser.parse_args()
    try:
        if args.command == 'check-names':
            print(json.dumps(check_names()))
            return 0
        if args.command == 'prepare':
            prepare(args)
            return 0
        if args.command == 'run':
            return run(args)
        return summarize(args) if args.command == 'summarize' else finalize(args)
    except AcceptanceFailure as error:
        print('ERROR: ' + str(error), flush=True)
        return 1
    except (OSError, ValueError, KeyError, subprocess.SubprocessError) as error:
        # Never echo exceptions that may include private paths, selectors or responses.
        print('ERROR: ' + args.command + ': ' + type(error).__name__, flush=True)
        return 1


if __name__ == '__main__':
    raise SystemExit(main())
