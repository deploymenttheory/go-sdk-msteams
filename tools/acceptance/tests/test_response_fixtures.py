import argparse
import contextlib
import importlib.util
import io
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

SPEC = importlib.util.spec_from_file_location('acceptance_ci', Path(__file__).resolve().parents[1] / 'ci.py')
ci = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(ci)


def registry(*steps, service='svc', package='./fixture'):
    return {'auth': {'package': './auth', 'steps': [{'id': 'authentication', 'tests': ['TestAccAuth']}]},
            'services': {service: {'package': package, 'steps': list(steps)}}}


class CallFixtureExportGuardTests(unittest.TestCase):
    def fixture(self, root, name='integration_holidays.json'):
        path = root / 'msteams/teams_admin_api/schedules/testdata' / name
        path.parent.mkdir(parents=True, exist_ok=True)
        return path

    def test_exports_only_changed_allowlisted_sanitized_files(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory).resolve()
            existing = self.fixture(root, 'existing.json')
            existing.write_text('{}')
            before = ci.response_fixtures.state(root, ci.CALL_RESPONSE_FIXTURES)
            current = self.fixture(root)
            current.write_text('{"Identity":"Tag:fixture-policy","Enabled":false,"Items":[],"Future":null}')
            receipt = self.fixture(root, 'integration_holidays.body')
            receipt.write_bytes(b'')
            private = root / 'private/journals'
            private.mkdir(parents=True)
            (private / 'owned.json').write_text('{"secret":"private-journal"}')
            reports = root / 'reports'
            ci.response_fixtures.capture(root, reports, before, 'a' * 40, ci.CALL_RESPONSE_FIXTURES)
            manifest = json.loads((reports / 'response-fixtures.json').read_text())
            self.assertEqual(manifest['revision'], 'a' * 40)
            self.assertEqual(len(manifest['files']), 2)
            for entry in manifest['files']:
                payload = (reports / entry['artifact']).read_bytes()
                self.assertEqual(entry['sha256'], ci.response_fixtures.hashlib.sha256(payload).hexdigest())
            self.assertEqual(len(list((reports / 'response-fixtures').rglob('*.*'))), 2)
            self.assertNotIn('private-journal', ''.join(p.read_text() for p in reports.rglob('*') if p.is_file()))

    def test_rejects_unexpected_outputs_and_nonempty_raw_receipts(self):
        for name, body in [('recovery.json', b'{}'), ('integration_holidays.body', b'private-body'), ('integration_holidays.txt', b'{}')]:
            with self.subTest(name=name), tempfile.TemporaryDirectory() as directory:
                root = Path(directory).resolve()
                self.fixture(root, name).write_bytes(body)
                with self.assertRaises(ci.response_fixtures.ExportError):
                    ci.response_fixtures.capture(root, root / 'reports', {}, 'a' * 40, ci.CALL_RESPONSE_FIXTURES)
                self.assertFalse((root / 'reports').exists())

    def test_rejects_unsanitized_or_malformed_json_before_export(self):
        bodies = ['not-json', '"scalar"', '{"id":"11111111-1111-4111-8111-111111111111"}',
                  '{"contact":"private-user@private.example"}', '{"value":"eyJprivate.payload.signature"}',
                  '{"client_secret":"private-value"}', '{"ExportHolidaysResult":{"SerializedHolidayRecords":"b3BhcXVlLXByaXZhdGUtYnl0ZXM="}}']
        for body in bodies:
            with self.subTest(body=body), tempfile.TemporaryDirectory() as directory:
                root = Path(directory).resolve()
                self.fixture(root).write_text(body)
                with self.assertRaises(ci.response_fixtures.ExportError):
                    ci.response_fixtures.capture(root, root / 'reports', {}, 'a' * 40, ci.CALL_RESPONSE_FIXTURES)
                self.assertFalse((root / 'reports').exists())

    def test_rejects_file_bound_and_symlink_or_escape_paths(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory).resolve()
            path = self.fixture(root)
            path.write_bytes(b' ' * (ci.response_fixtures.MAX_BYTES + 1))
            with self.assertRaises(ci.response_fixtures.ExportError):
                ci.response_fixtures.state(root, ci.CALL_RESPONSE_FIXTURES)
            path.write_text('{}')
            # Exercise the same filesystem observations on platforms that do
            # not permit an unprivileged test process to create real symlinks.
            original = Path.is_symlink
            with patch.object(Path, 'is_symlink', lambda value: value == path or original(value)):
                with self.assertRaises(ci.response_fixtures.ExportError):
                    ci.response_fixtures.state(root, ci.CALL_RESPONSE_FIXTURES)
            reports = root / 'reports'
            with patch.object(Path, 'is_symlink', lambda value: value == reports or original(value)):
                with self.assertRaises(ci.response_fixtures.ExportError):
                    ci.response_fixtures.capture(root, reports, {}, 'a' * 40, ci.CALL_RESPONSE_FIXTURES)
            source_directory = path.parent
            original_resolve = Path.resolve
            def resolve(value, *args, **kwargs):
                return root / 'outside' if value == source_directory else original_resolve(value, *args, **kwargs)
            with patch.object(Path, 'resolve', resolve):
                with self.assertRaises(ci.response_fixtures.ExportError):
                    ci.response_fixtures.state(root, ci.CALL_RESPONSE_FIXTURES)

    def test_writer_opt_in_is_scoped_to_call_integration(self):
        for service in ['svc', 'call_integration']:
            with self.subTest(service=service), tempfile.TemporaryDirectory() as directory:
                root = Path(directory).resolve()
                private = root / 'private'
                private.mkdir()
                (root / 'suites.json').write_text(json.dumps(registry({'id':'reads','tests':['TestAccRead']}, service=service)))
                environments = []
                def process(command, **kwargs):
                    if command[0] == 'git':
                        return subprocess.CompletedProcess(command, 0, stdout='a' * 40)
                    environments.append(kwargs['env'])
                    kwargs['stdout'].write(json.dumps({'Test':'TestAccRead','Action':'pass'}) + '\n')
                    return subprocess.CompletedProcess(command, 0)
                args = argparse.Namespace(auth=False,service=service,step=None,report_directory=root/'reports')
                with patch.object(ci,'ROOT',root), patch.object(ci,'HERE',root), patch.dict(os.environ,{'MSTEAMS_LAB_PRIVATE_DIRECTORY':str(private),'MSTEAMS_LAB_CAPTURE_HOLIDAY_FIXTURE':'1','MSTEAMS_UPDATE_TESTDATA':'1'},clear=True), patch.object(ci.subprocess,'run',side_effect=process), contextlib.redirect_stdout(io.StringIO()):
                    self.assertEqual(ci.run(args), 0)
                self.assertEqual(environments[0].get('MSTEAMS_LAB_CAPTURE_HOLIDAY_FIXTURE'), '1' if service=='call_integration' else None)
                self.assertNotIn('MSTEAMS_UPDATE_TESTDATA', environments[0])
