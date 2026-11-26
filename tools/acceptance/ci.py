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
AUDIO_FILE_MARKERS = [
    'audio import: passed',
    'audio get: passed',
    'audio list: passed',
    'audio export round trip: passed',
    'audio delete/soft-delete observation: passed',
]
AUDIO_FILE_ORACLE = 'PowerShell 8.1.0 / Go audio file parity: passed'
SCHEDULE_MARKERS = [
    'schedule range drafts: passed',
    'schedule create: passed',
    'schedule get/list: passed',
    'schedule update/read-back: passed',
    'holiday routes: passed',
    'schedule delete: passed',
]
SCHEDULE_ORACLE = 'PowerShell 8.1.0 / Go schedule parity: passed'
CALL_QUEUE_MARKERS = [
    'call queue missing get: passed',
    'call queue create: passed',
    'call queue get: passed',
    'call queue list: passed',
    'call queue update/read-back: passed',
    'call queue delete: passed',
]
CALL_QUEUE_ORACLE = 'PowerShell 8.1.0 / Go call queue parity: passed'
CALL_TEMPLATE_MARKERS = [
    'call template create/read-back: passed',
    'call template update/read-back: passed',
    'call template list: passed',
    'call template delete/absence: passed',
]
CALL_TEMPLATE_ORACLE = 'PowerShell 8.1.0 / Go call template parity: passed'
MAINLINE_ATTENDANT_MARKERS = [
    'spam detection template create/update/read-back: passed',
    'appointment booking flow create/update/read-back: passed',
    'question answer flow create/indexing/update/read-back: passed',
    'mainline attendant flow list/filter: passed',
    'spam detection template delete/absence: passed',
    'appointment booking flow delete/absence: passed',
    'question answer flow delete/absence: passed',
]
MAINLINE_ATTENDANT_ORACLE = 'PowerShell 8.1.0 / Go mainline attendant parity: passed'
AI_AGENT_MARKERS = [
    'AI agent create: passed',
    'AI agent get/list: passed',
    'AI agent update/read-back: passed',
    'AI agent delete/absence: passed',
]
AI_AGENT_ORACLE = 'PowerShell 8.1.0 / Go AI agent parity: passed'
AUTO_ATTENDANT_MARKERS = [
    'auto attendant create: passed',
    'auto attendant get/status: passed',
    'auto attendant list: passed',
    'auto attendant update/read-back: passed',
    'auto attendant refresh: passed',
    'auto attendant delete: passed',
]
AUTO_ATTENDANT_ORACLE = 'PowerShell 8.1.0 / Go auto attendant parity: passed'
PROOFS = {
    'ext-agent-template': (['EXT agent-template create/read-back: passed', 'EXT agent-template update/read-back: passed',
                            'EXT agent-template independent cleanup: passed'],
                           'PowerShell 8.1.0 / Go EXT agent-template parity: passed', 2),
    'ext-integration': (['integrated EXT generic-to-typed read: passed', 'integrated EXT typed-to-generic update: passed',
                         'integrated EXT generic replacement/read-back: passed', 'integrated EXT independent cleanup: passed'],
                        'PowerShell 8.1.0 / Go integrated EXT parity: passed', 3),
    'owned-crud': (PROOF_MARKERS, ORACLE_MARKER, 2),
    'owned-update-management-crud': (PROOF_MARKERS + ['update-management DateTime update/read-back: passed','update-management list edit/read-back: passed','update-management list clear/read-back: passed'], ORACLE_MARKER, 5),
    'owned-template-permission-populated-crud': (PROOF_MARKERS + ['policy hidden-list edit/read-back: passed','policy hidden-list clear/read-back: passed'], ORACLE_MARKER, 4),
    'owned-notification-crud': (PROOF_MARKERS + ['notification explicit-null update/read-back: passed'], ORACLE_MARKER, 3),
    'owned-shared-calling-populated-crud': (PROOF_MARKERS + ['shared calling populated list edit/clear: passed'], ORACLE_MARKER, 4),
    'owned-compliance-recording-crud': (PROOF_MARKERS + ['application rejection and absence: passed'], ORACLE_MARKER, 2),
    'user-assignment': (USER_ASSIGNMENT_MARKERS, USER_ASSIGNMENT_ORACLE, 3),
    'group-assignment': (GROUP_ASSIGNMENT_MARKERS, GROUP_ASSIGNMENT_ORACLE, 6),
    # Grant of a service built-in instance (no public New/Remove): no owned policy.
    'media-logging-group-grant': (['group policy grant/read-back: passed', 'group policy delete/absence: passed', 'group assignment cleanup: passed'],
                                  'PowerShell 8.1.0 / Go media logging group grant parity: passed', 3),
    'template-crud': (TEMPLATE_MARKERS, TEMPLATE_ORACLE, 4),
    'audio-file-lifecycle': (AUDIO_FILE_MARKERS, AUDIO_FILE_ORACLE, 1),
    'schedule-lifecycle': (SCHEDULE_MARKERS, SCHEDULE_ORACLE, 3),
    'call-queue-lifecycle': (CALL_QUEUE_MARKERS, CALL_QUEUE_ORACLE, 1),
    'call-template-lifecycle': (CALL_TEMPLATE_MARKERS, CALL_TEMPLATE_ORACLE, 1),
    'auto-recording-denial': (['auto recording create denial: verified', 'auto recording missing-template writes: passed'], None, 0),
    'mainline-attendant-lifecycle': (MAINLINE_ATTENDANT_MARKERS, MAINLINE_ATTENDANT_ORACLE, 1),
    'ai-agent-lifecycle': (AI_AGENT_MARKERS, AI_AGENT_ORACLE, 1),
    'auto-attendant-lifecycle': (AUTO_ATTENDANT_MARKERS, AUTO_ATTENDANT_ORACLE, 1),
    'populated-template-access': (['native populated-list create denial: verified', 'populated-list comparison cleanup: passed'], 'PowerShell 8.1.0 / Go populated-list authorization parity: passed', 1),
    'configuration-update': (['configuration update/read-back and omission: passed', 'configuration restoration/read-back: passed'], 'PowerShell 8.1.0 / Go configuration read parity: passed', 3),
    'generic-configuration-lifecycle': (['configuration create/read-back: passed', 'configuration set/read-back: passed', 'configuration remove/absence: passed'], 'PowerShell 8.1.0 / Go generic configuration parity: passed', 1),
}
# Publish fixed assertion text only. HTTP payloads and dynamic fixture names
# remain private even when a test fails.
FAILURE_REASONS = (
    'read/list comparison: response fields missing',
    'read/list comparison: property sets differ',
    'read/list comparison: identity values differ',
    'read/list comparison: setting or extension values differ',
    'read/list comparison: metadata values differ',
    'read/list metadata: Key.ScopeClass differs',
    'read/list metadata: Key.SchemaId differs',
    'read/list metadata: Key.AuthorityId differs',
    'read/list metadata: Key.DefaultXml.SchemaId differs',
    'read/list metadata: Key.DefaultXml.Data differs',
    'read/list metadata: Key.DefaultXml.ConfigObject differs',
    'read/list metadata: Key.DefaultXml.Signature differs',
    'read/list metadata: Key.DefaultXml.IsModified differs',
    'read/list metadata: Key.XmlRoot differs',
    'read/list metadata: ConfigMetadata differs',
    'read/list metadata: ConfigId differs',

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


SERVICE_ROOT = 'msteams/teams_admin_api'
# Partitions that are not teams_admin_api services: tenant/platform checks and
# Microsoft Graph team resources.
EXTRA_SERVICES = {'platform', 'graph', 'ext_integration'}
SERVICE_AUTH_METHOD = 'client-secret'
# Each service partition's tests live in their own Go package here.
LAB_ROOT = Path('tests/integration/lab')
# The auth partition loads every credential kind; its tests prove each one.
AUTH_PARTITION_METHOD = 'all'


def load_registry(here=None):
    return json.loads(((here or HERE) / 'suites.json').read_text())


def is_write(step):
    """Mutating and serialized steps hold the tenant-wide write lock."""
    return bool(step.get('mutation') or step.get('serial'))


def registry_steps(registry, service=None, auth=False):
    """Flatten the registry into ordered steps; each becomes one report."""
    units = []
    if not service:
        section = registry.get('auth', {})
        for step in section.get('steps', []):
            units.append(dict(step, partition='auth', package=section['package'], suite='auth/' + step['id']))
    if not auth:
        for name, entry in registry.get('services', {}).items():
            if service and name != service:
                continue
            for step in entry['steps']:
                units.append(dict(step, partition=name, service=name, package=entry['package'], suite=name + '/' + step['id']))
    return units


def registered_tests(registry):
    return [test for unit in registry_steps(registry) for test in unit['tests']]


def service_locked(entry):
    return any(is_write(step) for step in entry['steps'])


def validate_registry(registry, root=ROOT):
    """Reject registries that would run steps out of order or unlabelled."""
    auth = registry.get('auth') or {}
    if not auth.get('package') or not auth.get('steps') or any(not step.get('tests') for step in auth['steps']):
        raise AcceptanceFailure('the auth partition needs a package and steps with tests')
    if any(is_write(step) for step in auth['steps']):
        raise AcceptanceFailure('the auth partition must not mutate the tenant')
    services = registry.get('services') or {}
    if not services:
        raise AcceptanceFailure('no acceptance service partitions are registered')
    for name, entry in services.items():
        if name not in EXTRA_SERVICES and not (root / SERVICE_ROOT / name).is_dir():
            raise AcceptanceFailure('acceptance partition is not a teams_admin_api service: ' + name)
        if entry.get('package') != './' + (LAB_ROOT / name).as_posix():
            raise AcceptanceFailure('service partition package must be ./' + (LAB_ROOT / name).as_posix() + ': ' + name)
        ids = [step.get('id') for step in entry.get('steps', [])]
        if not ids or len(ids) != len(set(ids)) or not all(isinstance(i, str) and re.fullmatch(r'[a-z][a-z0-9-]*', i) for i in ids):
            raise AcceptanceFailure('service steps need unique lowercase ids: ' + name)
        writes = [is_write(step) for step in entry['steps']]
        if writes != sorted(writes):
            raise AcceptanceFailure('read-only steps must run before mutating or serialized steps: ' + name)
        for step in entry['steps']:
            if not step.get('tests'):
                raise AcceptanceFailure('service step has no tests: ' + name + '/' + step['id'])
            if step.get('mutation') and (step.get('proof') not in PROOFS or type(step.get('journal_entries_per_test')) is not int
                                         or step['journal_entries_per_test'] < 1):
                raise AcceptanceFailure('mutation step requires explicit checks and journal entry count: ' + name + '/' + step['id'])


def check_names(root=ROOT):
    errors = []
    counts = {'unit': 0, 'acceptance': 0}
    acceptance = {}
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
                    acceptance[name] = path.parent.relative_to(root)
    if errors:
        raise AcceptanceFailure('test names must start TestUnit or TestAcc: ' + ', '.join(sorted(set(errors))))
    if root == ROOT:
        registry = load_registry()
        validate_registry(registry, root)
        selected = registered_tests(registry)
        pending = json.loads((HERE / 'pending.json').read_text())
        if len(selected) != len(set(selected)) or set(selected) & set(pending) or set(selected) | set(pending) != set(acceptance):
            raise AcceptanceFailure('every TestAcc must have one CI step or an explicit pending disposition')
        if any(not entry.get('reason') or not (ROOT / entry['file']).is_file() for entry in pending.values()):
            raise AcceptanceFailure('pending acceptance requires an existing test file and a concrete reason')
        moved = sorted(test for test, entry in pending.items() if Path(entry['file']).parent != acceptance[test])
        if moved:
            raise AcceptanceFailure('pending acceptance file must define the test: ' + ', '.join(moved))
        stray = sorted(path.name for path in (root / LAB_ROOT).glob('*.go'))
        if stray:
            raise AcceptanceFailure('lab tests belong in a service package under ' + LAB_ROOT.as_posix() + ': ' + ', '.join(stray))
        misplaced = sorted(test for unit in registry_steps(registry) for test in unit['tests']
                           if Path(unit['package']) != acceptance[test])
        if misplaced:
            raise AcceptanceFailure('registered TestAcc must be defined in its partition package: ' + ', '.join(misplaced))
        selected_skips(registry)
    return counts


def selected_skips(registry, args=None):
    """Resolve exact test names before starting any tenant operation."""
    path = HERE / 'skips.json'
    skipped = json.loads(path.read_text()) if path.exists() else {}
    known = set(registered_tests(registry))
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


def auth_method():
    mode = os.environ.get('MSTEAMS_LAB_AUTH_METHOD', SERVICE_AUTH_METHOD)
    if mode not in {'certificate', 'client-secret', AUTH_PARTITION_METHOD}:
        raise ValueError('unsupported acceptance authentication method')
    return mode


def prepare(args):
    # Service partitions authenticate with the client secret only; the
    # certificate key material is delivered to the auth partition alone.
    keys = ['TENANT_ID', 'CLIENT_ID', 'CLIENT_SECRET', 'RECOVERY_CERTIFICATE_PEM', 'CALL_HOLD_AUDIO_FILE_ID', 'USER_FIXTURE_JSON', 'GROUP_FIXTURE_JSON']
    if auth_method() != SERVICE_AUTH_METHOD:
        keys[3:3] = ['CERTIFICATE_PEM', 'PRIVATE_KEY_PEM']
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
    if 'CERTIFICATE_PEM' in values:
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
    cache = getattr(args, 'cache_directory', None)
    cached = cache is not None and (cache / 'gallery.xml').is_file() and (cache / 'module.zip').is_file()
    metadata = args.metadata or (cache / 'gallery.xml' if cached else private / 'gallery.xml')
    archive = args.archive or (cache / 'module.zip' if cached else private / 'module.zip')
    if not args.metadata and not cached:
        baseline.fetch(f"{baseline.GALLERY}/api/v2/Packages(Id='MicrosoftTeams',Version='{baseline.VERSION}')", metadata, 1024 * 1024)
    if not args.archive and not cached:
        baseline.fetch(f'{baseline.GALLERY}/api/v2/package/MicrosoftTeams/{baseline.VERSION}', archive)
    # A cached download is only a transport optimisation: the pinned checksum
    # is verified on every run before the module is extracted.
    info = baseline.gallery_metadata(metadata.read_bytes(), baseline.VERSION)
    baseline.verify_archive(archive, info, baseline.PIN)
    if cache is not None and not cached:
        cache.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(metadata, cache / 'gallery.xml')
        shutil.copyfile(archive, cache / 'module.zip')
    baseline.extract_archive(archive, private / 'module')
    print(f'Repository secrets loaded: {len(values)}. MicrosoftTeams 8.1.0 checksum verified.', flush=True)


@functools.cache
def assertion_reasons(root=ROOT):
    """Retain checked-in assertion literals, never their format arguments."""
    reasons = set(FAILURE_REASONS)
    # Shared lab helpers live in ordinary .go files under tests/; their fixed
    # assertion text is as public as the text in the _test.go files.
    paths = set(root.rglob('*_test.go')) | set((root / 'tests').rglob('*.go'))
    for path in sorted(paths):
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
            matched = next((reason for reason in reasons if reason in event.get('Output', '')), None)
            if matched:
                failure_reasons[owner].add(matched)
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


def selected_services(registry):
    """Services the plan job selected, else those named by MSTEAMS_ACC_SERVICES
    (a dispatch rerun), else all."""
    services = registry.get('services', {})
    planned = os.environ.get('MSTEAMS_ACC_PLANNED_SERVICES', '').strip()
    if planned:
        names = set(json.loads(planned))
        if names - set(services):
            raise AcceptanceFailure('Planned acceptance services must name registered service partitions exactly.')
        return [name for name in services if name in names]
    names = {value for value in re.split(r'[\s,]+', os.environ.get('MSTEAMS_ACC_SERVICES', '').strip()) if value}
    if names - set(services):
        raise AcceptanceFailure('Acceptance service selections must name registered service partitions exactly.')
    return [name for name in services if not names or name in names]


# A pull request that changes any of these runs every service partition:
# they are shared by all clients, lab tests or tenant setup.
SHARED_PATTERNS = [re.compile(pattern) for pattern in [
    r'^go\.(mod|sum)$',
    r'^tools/baseline/',
    r'^internal/',
    r'^tests/integration/(auth|lab/internal)/',
    r'^msteams/[^/]+\.go$',
    r'^msteams/(auth|client|config|capabilities|internal)/',
    r'^msteams/[^/]+/(?!services\.go$)[^/]+\.go$',
    r'^msteams/[^/]+/(internal|policy)/',
]]


def changed_services(registry, files, base_registry, references=None):
    """The service partitions a pull request touches.

    A service is touched by a change to its lab test package, an SDK package
    of the same name, its suites.json entry, or a PowerShell oracle script its
    lab tests reference. A shared change (SHARED_PATTERNS) touches every
    service; unit test files in shared packages and acceptance workflow edits
    do not (the always-run auth partition exercises the partition steps).
    ci.py, ledger, docs and services.go wiring changes touch none: unit tests
    cover them, and pushes to main still run every partition."""
    services = registry.get('services', {})
    if any(pattern.search(path) for path in files if not path.endswith('_test.go') for pattern in SHARED_PATTERNS):
        return list(services)
    base_services = base_registry.get('services', {})
    references = references or {}
    oracles = {Path(path).name for path in files if path.startswith('tests/integration/lab/oracle/')}
    selected = []
    for name, entry in services.items():
        package = entry['package'].removeprefix('./').rstrip('/') + '/'
        sdk = re.compile(r'^msteams/[^/]+/' + re.escape(name) + '/')
        if (entry != base_services.get(name)
                or any(path.startswith(package) or sdk.search(path) for path in files)
                or (name == 'ext_integration' and any(
                    path.startswith(prefix + dependency + '/')
                    for path in files for prefix in ('msteams/teams_admin_api/', 'tests/integration/lab/')
                    for dependency in ('configuration_diagnostics', 'teams_feedback_policy', 'ai_agents', 'call_templates')))
                or oracles & set(references.get(name, ()))):
            selected.append(name)
    return selected


def oracle_references(registry):
    """Oracle script names each service's lab test package mentions."""
    references = {}
    scripts = [path.name for path in (ROOT / 'tests/integration/lab/oracle').glob('*.ps1')]
    for name, entry in registry.get('services', {}).items():
        sources = ''.join(path.read_text() for path in (ROOT / entry['package']).glob('*.go'))
        references[name] = [script for script in scripts if script in sources]
    return references


def pull_request_services(registry, base):
    """Services changed between base (the PR merge commit's base parent) and HEAD."""
    def git(*command):
        return subprocess.run(['git', *command], cwd=ROOT, capture_output=True, text=True, check=True).stdout
    files = [line for line in git('diff', '--name-only', base, 'HEAD').splitlines() if line]
    try:
        base_registry = json.loads(git('show', base + ':tools/acceptance/suites.json'))
    except (subprocess.CalledProcessError, json.JSONDecodeError):
        base_registry = {}
    return changed_services(registry, files, base_registry, oracle_references(registry))


def plan(args):
    """Emit the service job matrices: locked and unlocked partitions.

    A pull request (MSTEAMS_ACC_DIFF_BASE set) plans only the services it
    touches, so concurrent PRs don't overflow the tenant-wide lock queue."""
    registry = load_registry()
    validate_registry(registry)
    services = selected_services(registry)
    base = os.environ.get('MSTEAMS_ACC_DIFF_BASE', '').strip()
    if base and not os.environ.get('MSTEAMS_ACC_SERVICES', '').strip():
        changed = set(pull_request_services(registry, base))
        services = [name for name in services if name in changed]
    entries = registry['services']
    outputs = {
        'services': services,
        'services_unlocked': [name for name in services if not service_locked(entries[name])],
        'services_locked': [name for name in services if service_locked(entries[name])],
    }
    destination = os.environ.get('GITHUB_OUTPUT')
    for key, value in outputs.items():
        line = key + '=' + json.dumps(value, separators=(',', ':'))
        print(line, flush=True)
        if destination:
            with Path(destination).open('a') as stream:
                stream.write(line + '\n')
    return 0


def run(args):
    private = Path(os.environ['MSTEAMS_LAB_PRIVATE_DIRECTORY']).resolve()
    registry = load_registry()
    service = getattr(args, 'service', None)
    auth = bool(getattr(args, 'auth', False))
    if auth == bool(service):
        raise ValueError('choose the auth partition or one service partition')
    if service and service not in registry.get('services', {}):
        raise AcceptanceFailure('Unknown acceptance service partition: ' + service)
    units = registry_steps(registry, service=service, auth=auth)
    wanted = set(getattr(args, 'step', None) or [])
    if wanted - {unit['id'] for unit in units}:
        raise ValueError('unknown acceptance step')
    units = [unit for unit in units if not wanted or unit['id'] in wanted]
    skipped = selected_skips(registry, args)
    revision = subprocess.run(['git', 'rev-parse', 'HEAD'], cwd=ROOT, capture_output=True, text=True, check=True).stdout.strip()
    mode = auth_method()
    if auth and mode != AUTH_PARTITION_METHOD:
        raise AcceptanceFailure('The auth partition runs with MSTEAMS_LAB_AUTH_METHOD=all so it can prove every credential kind.')
    if service and mode != SERVICE_AUTH_METHOD:
        # Certificate support is proven once by the auth partition; services
        # do not repeat it for every resource.
        raise AcceptanceFailure('Service partitions run with client-secret authentication only.')
    reports = []
    blocked = False
    for unit in units:
        if unit.get('mutation') and (unit.get('proof') not in PROOFS or type(unit.get('journal_entries_per_test')) is not int or unit['journal_entries_per_test'] < 1):
            raise ValueError('mutation step requires explicit checks and journal entry count')
        active = [test for test in unit['tests'] if test not in skipped]
        env = {key: value for key, value in os.environ.items() if not key.startswith('MSTEAMS_')}
        env.update({
            'MSTEAMS_LAB_READ_ONLY': '1', 'MSTEAMS_LAB_MUTATIONS': '1' if unit.get('mutation') else '0',
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
        command = ['go', 'test', '-json', '-count=1', '-race', '-timeout=12m', '-run', pattern, unit['package']]
        previous_journals = set((private / 'journals').glob('*.json'))
        label = unit['suite'].replace('/', '--')
        log = private / (label + '.jsonl')
        events = []
        process_exit = 0
        # A failed earlier step leaves the service in an unknown state, so its
        # later mutations do not start; their tests report not-run.
        started = bool(active) and not (blocked and unit.get('mutation'))
        if started:
            fd = os.open(log, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
            with os.fdopen(fd, 'w') as stream:
                process_exit = subprocess.run(command, cwd=ROOT, env=env, stdout=stream, stderr=subprocess.STDOUT).returncode
            for line in log.read_text().splitlines():
                try:
                    events.append(json.loads(line))
                except ValueError:
                    pass
        children = {test: cases for test, cases in unit.get('children', {}).items() if test in active}
        report = evaluate(events, active, process_exit, proof=unit.get('proof'), expected_children=children)
        if not active:
            report['outcome'] = 'skipped'
        elif not started:
            report['blocked_by_earlier_step'] = True
        outcomes = {test['name']: test for test in report['tests']}
        report['tests'] = [outcomes[test] if test in outcomes else {'name': test, 'outcome': 'skip', 'skip_reason': skipped[test]} for test in unit['tests']]
        if unit.get('mutation'):
            new_journals = set((private / 'journals').glob('*.json')) - previous_journals
            entries = [entry for path in new_journals for entry in json.loads(path.read_text())['entries']]
            report['cleanup_journal_entries'] = len(entries)
            report['cleanup_journal_entries_cleaned'] = sum(entry['state'] == 'cleaned' for entry in entries)
            report['expected_cleanup_journal_entries'] = len(active) * unit['journal_entries_per_test']
            if len(entries) != report['expected_cleanup_journal_entries'] or report['cleanup_journal_entries_cleaned'] != len(entries):
                report['outcome'] = 'failed'
        report.update(suite=unit['suite'], partition=unit['partition'], step=unit['id'], phase='writes' if is_write(unit) else 'reads',
                      revision=revision, auth_method=mode, cloud='commercial', baseline_version='8.1.0', powershell_oracle_enabled=True)
        if unit.get('service'):
            report['service'] = unit['service']
        blocked = blocked or report['outcome'] == 'failed'
        reports.append(report)
        name = label
        write_json(args.report_directory / (name + '.json'), report)
        print_report(report)
    outcome = ('skipped' if all(r['outcome'] == 'skipped' for r in reports)
               else 'passed' if all(r['outcome'] in {'passed', 'skipped'} for r in reports) else 'failed')
    summary = {'revision': revision, 'auth_method': mode, 'partition': 'auth' if auth else service, 'outcome': outcome,
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

    if report.get('step') == 'tenant-access-outcomes':
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
        if report.get('step') != 'tenant-access-outcomes':
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
        raise AcceptanceFailure('No TestAcc reports were produced. Check the setup steps in the auth and service jobs.')
    registry = load_registry()
    skipped = selected_skips(registry, args)
    services = set(selected_services(registry))
    # The auth partition proves every credential kind once; service steps run
    # once each, with the client secret.
    units = {unit['suite']: unit for unit in registry_steps(registry) if unit['partition'] == 'auth' or unit['partition'] in services}
    expected = {(AUTH_PARTITION_METHOD if unit['partition'] == 'auth' else SERVICE_AUTH_METHOD, suite) for suite, unit in units.items()}
    actual = [(report.get('auth_method'), report.get('suite')) for report in reports]
    complete = set(actual) == expected and len(actual) == len(expected)
    for mode, name in sorted(expected - set(actual)):
        print(f'FAIL missing TestAcc report: {mode}/{name}', flush=True)
    if set(actual) - expected or len(actual) != len(set(actual)):
        print('FAIL unexpected or duplicate TestAcc reports', flush=True)
    for report in reports:
        print_report(report)
        if report.get('suite') in units and not report_matches_selection(report, units[report['suite']], skipped):
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
    commands.add_parser('plan', help='Write the auth and service job matrices to GITHUB_OUTPUT.')
    prep = commands.add_parser('prepare')
    prep.add_argument('--archive', type=Path)
    prep.add_argument('--metadata', type=Path)
    prep.add_argument('--cache-directory', type=Path, help='Reuse or populate a verified MicrosoftTeams download cache.')
    for name in ['run', 'finalize', 'summarize']:
        child = commands.add_parser(name)
        child.add_argument('--report-directory', type=Path, required=True)
        if name in {'run', 'summarize'}:
            child.add_argument('--skip-test', action='append', help='Skip an exact registered TestAcc name; repeat for multiple tests.')
            child.add_argument('--skip-reason', help='Public reason for the selective skips.')
            child.add_argument('--include-test', action='append', help='Run a test skipped by skips.json; repeat for multiple tests.')
        if name == 'run':
            target = child.add_mutually_exclusive_group(required=True)
            target.add_argument('--auth', action='store_true', help='Run the auth partition (MSTEAMS_LAB_AUTH_METHOD=all).')
            target.add_argument('--service', help='Run one service partition, in step order, with the client secret.')
            child.add_argument('--step', action='append', help='Run only this step id of the partition; repeatable.')
    args = parser.parse_args()
    try:
        if args.command == 'check-names':
            print(json.dumps(check_names()))
            return 0
        if args.command == 'plan':
            return plan(args)
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
