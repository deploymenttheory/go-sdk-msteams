"""Run explicit tenant acceptance selections; publish only structural outcomes."""
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


def prepare(args):
    keys = ['TENANT_ID', 'CLIENT_ID', 'CLIENT_SECRET', 'CERTIFICATE_PEM', 'PRIVATE_KEY_PEM', 'RECOVERY_CERTIFICATE_PEM']
    values = {key: os.environ.get('MSTEAMS_LAB_' + key, '') for key in keys}
    if any(not value for value in values.values()):
        raise ValueError('required repository lab secrets are missing')
    base = Path(os.environ['RUNNER_TEMP']).resolve()
    private = Path(tempfile.mkdtemp(prefix='msteams-acceptance-', dir=base))
    certificate = private / 'certificate.pem'
    key = private / 'private-key.pem'
    write_private(certificate, values['CERTIFICATE_PEM'])
    write_private(key, values['PRIVATE_KEY_PEM'])
    write_private(private / 'recovery-certificate.pem', values['RECOVERY_CERTIFICATE_PEM'])
    write_json(private / 'identity.json', {
        'tenant_id': values['TENANT_ID'], 'client_id': values['CLIENT_ID'],
        'client_secret': values['CLIENT_SECRET'],
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
    print('Private lab identity prepared; MicrosoftTeams 8.1.0 package checksum verified.')


def evaluate(events, expected, process_exit, require_crud=False):
    outcomes = {}
    markers = []
    child_counts = {'pass': 0, 'fail': 0, 'skip': 0}
    http_statuses = {name: [] for name in expected}
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
        if test == 'TestAccApplicationAccessPolicyCRUD' and action == 'output':
            for marker in PROOF_MARKERS + [ORACLE_MARKER]:
                if marker in event.get('Output', ''):
                    markers.append(marker)
    tests = [{'name': name, 'outcome': outcomes.get(name, 'not-run')} for name in expected]
    passed = bool(expected) and len(set(expected)) == len(expected) and process_exit == 0
    passed = passed and all(item['outcome'] == 'pass' for item in tests) and not child_counts['skip'] and not child_counts['fail']
    if require_crud:
        passed = passed and all(marker in markers for marker in PROOF_MARKERS) and markers.count(ORACLE_MARKER) == 2
    return {
        'outcome': 'passed' if passed else 'failed', 'tests': tests,
        'child_outcomes': child_counts, 'process_exit': process_exit,
        'http_statuses': {name: sorted(set(values)) for name, values in http_statuses.items() if values},
        'crud_assertions_observed': {marker: marker in markers for marker in PROOF_MARKERS} if require_crud else {},
        'crud_powershell_comparisons': markers.count(ORACLE_MARKER),
    }


def run(args):
    private = Path(os.environ['MSTEAMS_LAB_PRIVATE_DIRECTORY']).resolve()
    suites = json.loads((HERE / 'suites.json').read_text())
    selected = [name for name in suites if not args.suite or name in args.suite]
    if not selected or args.suite and set(args.suite) - set(suites):
        raise ValueError('unknown acceptance suite')
    revision = subprocess.run(['git', 'rev-parse', 'HEAD'], cwd=ROOT, capture_output=True, text=True, check=True).stdout.strip()
    mode = os.environ.get('MSTEAMS_LAB_AUTH_METHOD', 'certificate')
    if mode not in {'certificate', 'client-secret'}:
        raise ValueError('unsupported acceptance authentication method')
    reports = []
    for name in selected:
        suite = suites[name]
        env = {key: value for key, value in os.environ.items() if not key.startswith('MSTEAMS_')}
        env.update({
            'MSTEAMS_LAB_READ_ONLY': '1', 'MSTEAMS_LAB_MUTATIONS': '1' if suite.get('mutation') else '0',
            'MSTEAMS_LAB_CONSEQUENTIAL': '0', 'MSTEAMS_LAB_TIMEOUT_SECONDS': '300',
            'MSTEAMS_LAB_IDENTITY_FILE': str(private / 'identity.json'),
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
        report = evaluate(events, suite['tests'], result.returncode, suite.get('mutation', False))
        if suite.get('mutation'):
            new_journals = set((private / 'journals').glob('*.json')) - previous_journals
            entries = [entry for path in new_journals for entry in json.loads(path.read_text())['entries']]
            report['cleanup_journal_entries'] = len(entries)
            report['cleanup_journal_entries_cleaned'] = sum(entry['state'] == 'cleaned' for entry in entries)
            if not entries or report['cleanup_journal_entries_cleaned'] != len(entries):
                report['outcome'] = 'failed'
        report.update(suite=name, revision=revision, auth_method=mode, cloud='commercial', baseline_version='8.1.0',
                      scope='explicit selected acceptance cases; no complete SDK parity claim', coverage_claim=False,
                      powershell_oracle_enabled=True)
        reports.append(report)
        write_json(args.report_directory / (name + '.json'), report)
        print(name + ': ' + report['outcome'] + ' (' + str(len(report['tests'])) + ' selected tests)')
    summary = {'revision': revision, 'auth_method': mode, 'outcome': 'passed' if all(r['outcome'] == 'passed' for r in reports) else 'failed',
               'suites': [{'name': r['suite'], 'outcome': r['outcome']} for r in reports], 'coverage_claim': False}
    write_json(args.report_directory / 'summary.json', summary)
    return 0 if summary['outcome'] == 'passed' else 1


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
    print('Encrypted recovery journals retained; private identity, logs and raw journals removed.')
    return 0 if cleaned == len(entries) else 1


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest='command', required=True)
    commands.add_parser('check-names')
    prep = commands.add_parser('prepare')
    prep.add_argument('--archive', type=Path)
    prep.add_argument('--metadata', type=Path)
    for name in ['run', 'finalize']:
        child = commands.add_parser(name)
        child.add_argument('--report-directory', type=Path, required=True)
        if name == 'run':
            child.add_argument('--suite', action='append')
    args = parser.parse_args()
    try:
        if args.command == 'check-names':
            print(json.dumps(check_names()))
            return 0
        if args.command == 'prepare':
            prepare(args)
            return 0
        return run(args) if args.command == 'run' else finalize(args)
    except (OSError, ValueError, KeyError, subprocess.SubprocessError):
        # Never echo exceptions that may include private paths, selectors or responses.
        print('Acceptance ' + args.command + ' failed; required inputs or execution evidence are incomplete.')
        return 1


if __name__ == '__main__':
    raise SystemExit(main())
