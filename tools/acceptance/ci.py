"""Run TestAcc tests and print their results."""
import argparse
import functools
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
GROUP_ASSIGNMENT_MARKERS = [
    'group policy create/read-back: passed',
    'group policy update/read-back: passed',
    'group policy delete/absence: passed',
    'group policy grant/read-back: passed',
    'group assignment cleanup: passed',
]
GROUP_ASSIGNMENT_ORACLE = 'PowerShell 8.1.0 / Go group assignment parity: passed'
TEMPLATE_MARKERS = [
    'template create/read-back and General channel: passed',
    'template update/read-back: passed',
    'template list: passed',
    'template delete/absence: passed',
]
TEMPLATE_ORACLE = 'PowerShell 8.1.0 / Go template read parity: passed'
PROOFS = {
    'owned-crud': (PROOF_MARKERS, ORACLE_MARKER, 2),
    'owned-notification-crud': (PROOF_MARKERS + ['notification explicit-null update/read-back: passed'], ORACLE_MARKER, 3),
    'owned-shared-calling-populated-crud': (PROOF_MARKERS + ['shared calling populated list edit/clear: passed'], ORACLE_MARKER, 4),
    'user-assignment': (USER_ASSIGNMENT_MARKERS, USER_ASSIGNMENT_ORACLE, 3),
    'group-assignment': (GROUP_ASSIGNMENT_MARKERS, GROUP_ASSIGNMENT_ORACLE, 6),
    'template-crud': (TEMPLATE_MARKERS, TEMPLATE_ORACLE, 4),
    'populated-template-access': (['native populated-list create denial: verified', 'populated-list comparison cleanup: passed'], 'PowerShell 8.1.0 / Go populated-list authorization parity: passed', 1),
    'configuration-update': (['configuration update/read-back and omission: passed', 'configuration restoration/read-back: passed'], 'PowerShell 8.1.0 / Go configuration read parity: passed', 3),
}
# Publish fixed assertion text only. HTTP payloads and dynamic fixture names
# remain private even when a test fails.
FAILURE_REASONS = (
    'read/list comparison: response fields missing',
    'read/list comparison: property sets differ',
    'read/list comparison: identity values differ',
    'read/list comparison: setting or extension values differ',
    'read/list comparison: metadata values differ',
    'pinned federation read oracle failed; no output retained',
    'pinned federation read could not be verified',
    'pinned federation identity differs',
    'PowerShell and Go selected federation settings differ',
    'selected federation read failed',
    'Global federation read/list observations differ',
    'canonical tenant binding required',
    'pinned voicemail read failed; no output retained',
    'pinned voicemail projection unavailable',
    'voicemail Global read/list observations differ',
    'Go and pinned voicemail scalar projections differ',
    'pinned files policy read failed; no output retained',
    'pinned files policy projection unavailable',
    'SDK and PowerShell files policy projections differ',
    'oracle failure stage: module',
    'oracle failure stage: authentication',
    'oracle failure stage: read',
    'oracle failure stage: projection',
    'oracle failure stage: unknown',
    'compiled scalar Global read failed',
    'compiled scalar Global identity differs',
    'compiled scalar Global read/list differ',
    'Go and PowerShell compiled scalar getters differ',
    'Go and PowerShell configuration update getters differ',
    'configuration original Boolean is unavailable; update refused',
    'configuration update or omitted-field preservation did not converge',
    'configuration original Boolean restoration failed',
    'configuration original getter state was not independently restored',
    'pinned mobility policy projection unavailable',
    'SDK and PowerShell mobility policy projections differ',
    'mobility non-default create read-back did not converge',
    'mobility update or omitted-field preservation did not converge',
    'mobility deletion and independent absence were not verified',
    'supported original multi-tenant setting unavailable; update refused',
    'Go and PowerShell multi-tenant configuration getters differ',
    'multi-tenant configuration update failed',
    'multi-tenant configuration update did not persist',
    'multi-tenant original setting restoration failed',
    'multi-tenant original state not independently restored',
    'pinned roaming policy projection unavailable',
    'pinned notification policy projection unavailable',
    'SDK and PowerShell NetworkRoaming policy projections differ',
    'SDK and PowerShell notification policy projections differ',
    'NetworkRoaming non-default create read-back did not converge',
    'NetworkRoaming update or omitted-field preservation did not converge',
    'notification non-default create read-back did not converge',
    'notification update or omitted-field preservation did not converge',
    'notification explicit-null description update failed',
    'notification explicit-null read-back did not converge',
    'pinned personal attendant policy projection unavailable',
    'pinned recording rollout policy projection unavailable',
    'SDK and PowerShell personal attendant policy projections differ',
    'SDK and PowerShell recording rollout policy projections differ',
    'personal attendant non-default create read-back did not converge',
    'personal attendant update or omitted-field preservation did not converge',
    'recording rollout non-default create read-back did not converge',
    'recording rollout update or omitted-field preservation did not converge',
    'recording rollout sparse update included an unbound field',
)
FEDERATION_CASES = {'external-policy', 'acs', 'external-users', 'tenant'}


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
        selected_skips(suites)
    return counts


def selected_skips(suites, args=None):
    """Resolve exact test names before starting any tenant operation."""
    path = HERE / 'skips.json'
    skipped = json.loads(path.read_text()) if path.exists() else {}
    known = {test for suite in suites.values() for test in suite.get('tests', [])}
    if not isinstance(skipped, dict):
        raise AcceptanceFailure('Acceptance skips must map exact TestAcc names to public reasons.')

    def names(option, variable):
        values = list(getattr(args, option, None) or [])
        values += re.split(r'[\s,]+', os.environ.get(variable, '').strip())
        result = {value for value in values if value}
        if any(not re.fullmatch(r'TestAcc[A-Za-z0-9_]+', value) for value in result) or result - known:
            raise AcceptanceFailure('Acceptance skip/include selections must name existing registered TestAcc tests exactly.')
        return result

    def reason_valid(reason):
        return isinstance(reason, str) and bool(reason.strip()) and len(reason) <= 500 and not any(ord(char) < 32 for char in reason)

    if set(skipped) - known or any(not reason_valid(reason) for reason in skipped.values()):
        raise AcceptanceFailure('Each configured acceptance skip needs a registered TestAcc name and a public single-line reason.')
    excluded = names('skip_test', 'MSTEAMS_ACC_SKIP_TESTS')
    included = names('include_test', 'MSTEAMS_ACC_INCLUDE_TESTS')
    if excluded & included or included - set(skipped):
        raise AcceptanceFailure('Include only configured skipped tests; do not both skip and include a test.')
    reason = getattr(args, 'skip_reason', None) or os.environ.get('MSTEAMS_ACC_SKIP_REASON', '')
    if excluded and not reason_valid(reason):
        raise AcceptanceFailure('Selective acceptance skips require --skip-reason or MSTEAMS_ACC_SKIP_REASON with a public single-line reason.')
    for test in included:
        del skipped[test]
    skipped.update({test: reason.strip() for test in excluded})
    return skipped


def report_matches_selection(report, suite, skipped):
    tests = report.get('tests', [])
    expected = suite.get('tests', [])
    if [test.get('name') for test in tests] != expected:
        return False
    for test in tests:
        reason = skipped.get(test['name'])
        if reason is not None:
            if test.get('outcome') != 'skip' or test.get('skip_reason') != reason:
                return False
        elif test.get('outcome') != 'pass' or 'skip_reason' in test:
            return False
    outcome = 'skipped' if all(test in skipped for test in expected) else 'passed'
    return bool(expected) and report.get('outcome') == outcome


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


def group_fixture(values):
    try:
        fixture = json.loads(values['GROUP_FIXTURE_JSON'])
        valid = (set(fixture) == {'tenant_id', 'group_id', 'display_name', 'mail_nickname',
                                 'security_enabled', 'mail_enabled', 'role_assignable', 'empty_membership_verified'}
                 and fixture['tenant_id'].lower() == values['TENANT_ID'].lower()
                 and re.fullmatch(r'[0-9a-fA-F]{8}(?:-[0-9a-fA-F]{4}){3}-[0-9a-fA-F]{12}', fixture['group_id'])
                 and fixture['display_name'] == 'go-sdk-msteams acceptance policy group'
                 and fixture['mail_nickname'] == 'go-sdk-msteams-acceptance-policy'
                 and fixture['security_enabled'] is True and fixture['mail_enabled'] is False
                 and fixture['role_assignable'] is False and fixture['empty_membership_verified'] is True)
    except (ValueError, TypeError, AttributeError, KeyError):
        valid = False
    if not valid:
        raise AcceptanceFailure('Invalid repository Actions secret: MSTEAMS_LAB_GROUP_FIXTURE_JSON')
    return fixture


def shared_calling_fixture(tenant_id, raw):
    try:
        fixture = json.loads(raw)
        numbers = fixture['emergency_numbers']
        valid = (set(fixture) == {'tenant_id', 'resource_account_id', 'emergency_numbers'}
                 and fixture['tenant_id'].lower() == tenant_id.lower()
                 and re.fullmatch(r'[0-9a-fA-F]{8}(?:-[0-9a-fA-F]{4}){3}-[0-9a-fA-F]{12}', fixture['resource_account_id'])
                 and isinstance(numbers, list) and len(numbers) >= 2
                 and all(isinstance(number, str) and re.fullmatch(r'\+[1-9][0-9]{1,14}', number) for number in numbers)
                 and len(set(numbers)) == len(numbers))
    except (ValueError, TypeError, AttributeError, KeyError):
        valid = False
    if not valid:
        raise AcceptanceFailure('Invalid optional repository Actions secret: MSTEAMS_LAB_SHARED_CALLING_FIXTURE_JSON')
    return fixture


def prepare(args):
    keys = ['TENANT_ID', 'CLIENT_ID', 'CLIENT_SECRET', 'CERTIFICATE_PEM', 'PRIVATE_KEY_PEM', 'RECOVERY_CERTIFICATE_PEM', 'CALL_HOLD_AUDIO_FILE_ID', 'USER_FIXTURE_JSON', 'GROUP_FIXTURE_JSON']
    values = {key: os.environ.get('MSTEAMS_LAB_' + key, '') for key in keys}
    missing = ['MSTEAMS_LAB_' + key for key, value in values.items() if not value.strip()]
    if missing:
        raise AcceptanceFailure('Missing repository Actions secrets: ' + ', '.join(missing))
    if not re.fullmatch(r'(?:[0-9a-fA-F]{32}|[0-9a-fA-F]{8}(?:-[0-9a-fA-F]{4}){3}-[0-9a-fA-F]{12})', values['CALL_HOLD_AUDIO_FILE_ID']):
        raise AcceptanceFailure('Invalid repository Actions secret: MSTEAMS_LAB_CALL_HOLD_AUDIO_FILE_ID')
    fixture = user_fixture(values)
    group = group_fixture(values)
    phone_json = os.environ.get('MSTEAMS_LAB_SHARED_CALLING_FIXTURE_JSON', '')
    phone = shared_calling_fixture(values['TENANT_ID'], phone_json) if phone_json else None
    base = Path(os.environ['RUNNER_TEMP']).resolve()
    private = Path(tempfile.mkdtemp(prefix='msteams-acceptance-', dir=base))
    certificate = private / 'certificate.pem'
    key = private / 'private-key.pem'
    write_private(certificate, values['CERTIFICATE_PEM'])
    write_private(key, values['PRIVATE_KEY_PEM'])
    write_private(private / 'recovery-certificate.pem', values['RECOVERY_CERTIFICATE_PEM'])
    write_json(private / 'user-fixture.json', fixture)
    write_json(private / 'group-fixture.json', group)
    if phone is not None:
        write_json(private / 'shared-calling-fixture.json', phone)
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


@functools.cache
def assertion_reasons(root=ROOT):
    """Retain checked-in assertion literals, never their format arguments."""
    reasons = set(FAILURE_REASONS)
    for path in root.rglob('*_test.go'):
        if '.git' in path.parts:
            continue
        for literal in re.findall(r'\bt\.(?:Fatalf?|Errorf?)\(\s*("(?:[^"\\]|\\.)*")', path.read_text()):
            try:
                prefix = json.loads(literal).split('%', 1)[0].strip()
            except ValueError:
                continue
            if 8 <= len(prefix) <= 240 and not any(char in prefix for char in '\r\n'):
                reasons.add(prefix)
    return tuple(sorted(reasons, key=lambda value: (-len(value), value)))


def evaluate(events, expected, process_exit, require_crud=False, proof=None, expected_children=None):
    proof = proof or ('owned-crud' if require_crud else None)
    proof_markers, oracle_marker, oracle_count = PROOFS[proof] if proof else ([], None, 0)
    outcomes = {}
    markers = {name: [] for name in expected}
    child_counts = {'pass': 0, 'fail': 0, 'skip': 0}
    http_statuses = {name: [] for name in expected}
    service_codes = {name: [] for name in expected}
    failure_reasons = {name: set() for name in expected}
    reasons = assertion_reasons()
    child_results = {name + '/' + child: None for name, children in (expected_children or {}).items() for child in children}
    child_markers = {name: [] for name in child_results}
    for event in events:
        test = event.get('Test', '')
        action = event.get('Action')
        if test in child_results and action in {'pass', 'fail', 'skip'}:
            child_results[test] = action
        if test in expected and action in {'pass', 'fail', 'skip'}:
            outcomes[test] = action
        if '/' in test and test.split('/', 1)[0] in expected and action in child_counts:
            child_counts[action] += 1
        owner = test.split('/', 1)[0]
        if owner in http_statuses and action == 'output':
            http_statuses[owner].extend(int(value) for value in re.findall(r'HTTP (\d{3})\b', event.get('Output', '')))
            service_codes[owner].extend(int(value) for value in re.findall(r'PolicyRP (\d+)\b', event.get('Output', '')))
            case = test.removeprefix('TestAccFederationReadOnly/')
            prefix = case + ': ' if owner == 'TestAccFederationReadOnly' and case in FEDERATION_CASES else ''
            matched = next((reason for reason in reasons if reason in event.get('Output', '')), None)
            if matched:
                failure_reasons[owner].add(prefix + matched)
        if owner in markers and action == 'output':
            for marker in proof_markers + ([oracle_marker] if oracle_marker else []):
                if marker in event.get('Output', ''):
                    markers[owner].append(marker)
                    if test in child_markers:
                        child_markers[test].append(marker)
    tests = [{'name': name, 'outcome': outcomes.get(name, 'not-run')} for name in expected]
    passed = bool(expected) and len(set(expected)) == len(expected) and process_exit == 0
    passed = passed and all(item['outcome'] == 'pass' for item in tests) and not child_counts['skip'] and not child_counts['fail']
    passed = passed and all(value == 'pass' for value in child_results.values())
    if proof:
        passed = passed and all(all(marker in markers[name] for marker in proof_markers) and markers[name].count(oracle_marker) == oracle_count for name in expected)
    child_checks = {}
    if proof and expected_children:
        for name, seen in child_markers.items():
            count = len(expected_children[name.split('/', 1)[0]])
            child_checks[name] = (oracle_count % count == 0 and all(marker in seen for marker in proof_markers)
                                  and seen.count(oracle_marker) == oracle_count // count)
        passed = passed and all(child_checks.values())
    return {
        'outcome': 'passed' if passed else 'failed', 'tests': tests,
        'child_outcomes': child_counts, 'process_exit': process_exit,
        'required_child_outcomes': child_results,
        'required_child_checks': child_checks,
        'http_statuses': {name: sorted(set(values)) for name, values in http_statuses.items() if values},
        'service_error_codes': {name: sorted(set(values)) for name, values in service_codes.items() if values},
        'failure_reasons': {name: sorted(values) for name, values in failure_reasons.items() if values},
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
                and (not kind or bool(suite.get('mutation') or suite.get('serial')) == (kind == 'writes'))]
    if not selected or args.suite and set(args.suite) - set(suites):
        raise ValueError('unknown acceptance suite')
    skipped = selected_skips(suites, args)
    revision = subprocess.run(['git', 'rev-parse', 'HEAD'], cwd=ROOT, capture_output=True, text=True, check=True).stdout.strip()
    mode = os.environ.get('MSTEAMS_LAB_AUTH_METHOD', 'certificate')
    if mode not in {'certificate', 'client-secret'}:
        raise ValueError('unsupported acceptance authentication method')
    reports = []
    for name in selected:
        suite = suites[name]
        if suite.get('mutation') and (suite.get('proof') not in PROOFS or type(suite.get('journal_entries_per_test')) is not int or suite['journal_entries_per_test'] < 1):
            raise ValueError('mutation suite requires explicit checks and journal entry count')
        active = [test for test in suite['tests'] if test not in skipped]
        env = {key: value for key, value in os.environ.items() if not key.startswith('MSTEAMS_')}
        env.update({
            'MSTEAMS_LAB_READ_ONLY': '1', 'MSTEAMS_LAB_MUTATIONS': '1' if suite.get('mutation') else '0',
            'MSTEAMS_LAB_CONSEQUENTIAL': '0', 'MSTEAMS_LAB_TIMEOUT_SECONDS': '300',
            'MSTEAMS_LAB_IDENTITY_FILE': str(private / 'identity.json'),
            'MSTEAMS_LAB_USER_FIXTURE_FILE': str(private / 'user-fixture.json'),
            'MSTEAMS_LAB_GROUP_FIXTURE_FILE': str(private / 'group-fixture.json'),
            'MSTEAMS_LAB_MODULE_PATH': str(private / 'module'),
            'MSTEAMS_LAB_JOURNAL_DIRECTORY': str(private / 'journals'),
            'MSTEAMS_LAB_POWERSHELL_ORACLE': '1', 'MSTEAMS_LAB_AUTH_METHOD': mode,
            'MSTEAMS_LAB_ACCESS_PROFILE': 'teams-administrator', 'MSTEAMS_LIVE_ACCESS': '1',
        })
        if (private / 'shared-calling-fixture.json').is_file():
            env['MSTEAMS_LAB_SHARED_CALLING_FIXTURE_FILE'] = str(private / 'shared-calling-fixture.json')
        pattern = '^(' + '|'.join(re.escape(test) for test in active) + ')$'
        command = ['go', 'test', '-json', '-count=1', '-race', '-timeout=12m', '-run', pattern, suite['package']]
        previous_journals = set((private / 'journals').glob('*.json'))
        log = private / (name + '.jsonl')
        events = []
        process_exit = 0
        if active:
            fd = os.open(log, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
            with os.fdopen(fd, 'w') as stream:
                process_exit = subprocess.run(command, cwd=ROOT, env=env, stdout=stream, stderr=subprocess.STDOUT).returncode
            for line in log.read_text().splitlines():
                try:
                    events.append(json.loads(line))
                except ValueError:
                    pass
        children = {test: cases for test, cases in suite.get('children', {}).items() if test in active}
        report = evaluate(events, active, process_exit, proof=suite.get('proof'), expected_children=children)
        if not active:
            report['outcome'] = 'skipped'
        outcomes = {test['name']: test for test in report['tests']}
        report['tests'] = [outcomes[test] if test in outcomes else {'name': test, 'outcome': 'skip', 'skip_reason': skipped[test]} for test in suite['tests']]
        if suite.get('mutation'):
            new_journals = set((private / 'journals').glob('*.json')) - previous_journals
            entries = [entry for path in new_journals for entry in json.loads(path.read_text())['entries']]
            report['cleanup_journal_entries'] = len(entries)
            report['cleanup_journal_entries_cleaned'] = sum(entry['state'] == 'cleaned' for entry in entries)
            report['expected_cleanup_journal_entries'] = len(active) * suite['journal_entries_per_test']
            if len(entries) != report['expected_cleanup_journal_entries'] or report['cleanup_journal_entries_cleaned'] != len(entries):
                report['outcome'] = 'failed'
        report.update(suite=name, revision=revision, auth_method=mode, cloud='commercial', baseline_version='8.1.0',
                      powershell_oracle_enabled=True)
        reports.append(report)
        write_json(args.report_directory / (name + '.json'), report)
        print_report(report)
    outcome = ('skipped' if all(r['outcome'] == 'skipped' for r in reports)
               else 'passed' if all(r['outcome'] in {'passed', 'skipped'} for r in reports) else 'failed')
    summary = {'revision': revision, 'auth_method': mode, 'outcome': outcome,
               'passed_tests': sum(test['outcome'] == 'pass' for report in reports for test in report['tests']),
               'skipped_tests': sum(test['outcome'] == 'skip' for report in reports for test in report['tests']),
               'suites': [{'name': r['suite'], 'outcome': r['outcome']} for r in reports], 'coverage_claim': False}
    write_json(args.report_directory / 'summary.json', summary)
    write_summary(reports)
    return 0 if summary['outcome'] in {'passed', 'skipped'} else 1


def failure_detail(report, test):
    statuses = report.get('http_statuses', {}).get(test['name'], [])
    details = (['HTTP ' + ', '.join(map(str, statuses))] if statuses else [])
    details += report.get('failure_reasons', {}).get(test['name'], [])
    return '; '.join(details) if details else test['outcome']


def print_report(report):
    passed = sum(test['outcome'] == 'pass' for test in report['tests'])
    print(f"{report['suite']} ({report['auth_method']}): {passed}/{len(report['tests'])} passed", flush=True)
    for test in report['tests']:
        if test['outcome'] == 'pass':
            continue
        if test['outcome'] == 'skip' and test.get('skip_reason'):
            print(f"SKIP {test['name']}: {test['skip_reason']}", flush=True)
            continue
        detail = failure_detail(report, test)
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
    for name, outcome in report.get('required_child_outcomes', {}).items():
        if outcome != 'pass':
            print(f'FAIL {name}: {outcome or "not-run"}', flush=True)
    for name, passed in report.get('required_child_checks', {}).items():
        if not passed:
            print(f'FAIL {name}: missing CRUD assertion or pinned comparison', flush=True)
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
    if entries is not None and (entries != cleaned or entries != expected_entries):
        print(f'FAIL cleanup journals: {cleaned}/{entries} entries completed', flush=True)
        if entries != expected_entries:
            print(f'FAIL cleanup journal count: {entries}/{expected_entries} required entries', flush=True)


def write_summary(reports):
    lines = ['## TestAcc results', '', '| Auth | Suite | Result | Passed | Failed | Skipped | Not run |', '|---|---|---|---:|---:|---:|---:|']
    failed = []
    skipped = []
    for report in reports:
        counts = {state: sum(test['outcome'] == state for test in report['tests']) for state in ['pass', 'fail', 'skip', 'not-run']}
        lines.append(f"| {report['auth_method']} | {report['suite']} | {report['outcome']} | {counts['pass']} | {counts['fail']} | {counts['skip']} | {counts['not-run']} |")
        for test in report['tests']:
            if test['outcome'] == 'skip' and test.get('skip_reason'):
                reason = test['skip_reason'].replace('|', '\\|').replace('`', '\\`')
                skipped.append(f"| {report['auth_method']} | `{test['name']}` | {reason} |")
            elif test['outcome'] != 'pass':
                detail = failure_detail(report, test)
                failed.append(f"| {report['auth_method']} | `{test['name']}` | {detail} |")
    if failed:
        lines += ['', '| Auth | Failed test | Result |', '|---|---|---|'] + failed
    if skipped:
        lines += ['', '### Configured skips', '', 'Skipped tests did not run and provide no acceptance evidence.', '', '| Auth | Skipped test | Reason |', '|---|---|---|'] + skipped
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
    skipped = selected_skips(suites, args)
    expected = {(mode, name) for mode in ['certificate', 'client-secret'] for name in suites}
    actual = [(report.get('auth_method'), report.get('suite')) for report in reports]
    complete = set(actual) == expected and len(actual) == len(expected)
    for mode, name in sorted(expected - set(actual)):
        print(f'FAIL missing TestAcc suite: {mode}/{name}', flush=True)
    if set(actual) - expected or len(actual) != len(set(actual)):
        print('FAIL unexpected or duplicate TestAcc suite reports', flush=True)
    for report in reports:
        print_report(report)
        if report.get('suite') in suites and not report_matches_selection(report, suites[report['suite']], skipped):
            print(f"FAIL TestAcc selection or outcomes differ: {report['auth_method']}/{report['suite']}", flush=True)
            complete = False
    write_summary(reports)
    return 0 if complete and all(report['outcome'] in {'passed', 'skipped'} for report in reports) else 1


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
        if name in {'run', 'summarize'}:
            child.add_argument('--skip-test', action='append', help='Skip an exact registered TestAcc name; repeat for multiple tests.')
            child.add_argument('--skip-reason', help='Public reason for the selective skips.')
            child.add_argument('--include-test', action='append', help='Run a test skipped by skips.json; repeat for multiple tests.')
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
