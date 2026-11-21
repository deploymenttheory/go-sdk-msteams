import argparse
import contextlib
import importlib.util
import io
import json
import os
from pathlib import Path
import subprocess
import tarfile
import tempfile
import unittest
from unittest.mock import patch

SPEC = importlib.util.spec_from_file_location('acceptance_ci', Path(__file__).resolve().parents[1] / 'ci.py')
ci = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(ci)


class AcceptanceGuardTests(unittest.TestCase):
    def test_skip_missing_empty_and_failed_process_cannot_pass(self):
        for events, expected, code in [
            ([{'Test': 'TestAccRead', 'Action': 'skip'}], ['TestAccRead'], 0),
            ([], ['TestAccRead'], 0),
            ([], [], 0),
            ([{'Test': 'TestAccRead', 'Action': 'pass'}], ['TestAccRead'], 1),
            ([{'Test': 'TestAccRead', 'Action': 'pass'}, {'Test': 'TestAccRead/fixture', 'Action': 'skip'}], ['TestAccRead'], 0),
        ]:
            with self.subTest(events=events, expected=expected, code=code):
                self.assertEqual(ci.evaluate(events, expected, code)['outcome'], 'failed')

    def test_failed_test_summary_names_the_test_and_http_status(self):
        events = [
            {'Test': 'TestAccPrivacyRead', 'Action': 'output', 'Output': 'Global read failed (HTTP 403); private-response'},
            {'Test': 'TestAccPrivacyRead', 'Action': 'fail'},
        ]
        report = ci.evaluate(events, ['TestAccPrivacyRead'], 1)
        report.update(suite='resource-reads', auth_method='certificate')
        with tempfile.TemporaryDirectory() as directory:
            summary = Path(directory) / 'summary.md'
            output = io.StringIO()
            with patch.dict(os.environ, {'GITHUB_STEP_SUMMARY': str(summary), 'GITHUB_ACTIONS': 'true'}), contextlib.redirect_stdout(output):
                ci.print_report(report)
                ci.write_summary([report])
            self.assertIn('TestAccPrivacyRead: HTTP 403', output.getvalue())
            self.assertIn('::error title=TestAccPrivacyRead::HTTP 403', output.getvalue())
            self.assertIn('`TestAccPrivacyRead` | HTTP 403', summary.read_text())
            self.assertNotIn('private-response', output.getvalue() + summary.read_text())

    def test_prepare_names_missing_secrets_without_exposing_present_values(self):
        with patch.dict(os.environ, {'MSTEAMS_LAB_TENANT_ID': 'private-tenant-value'}, clear=True):
            with self.assertRaises(ci.AcceptanceFailure) as failure:
                ci.prepare(argparse.Namespace(metadata=None, archive=None))
        self.assertIn('MSTEAMS_LAB_CLIENT_ID', str(failure.exception))
        self.assertIn('MSTEAMS_LAB_RECOVERY_CERTIFICATE_PEM', str(failure.exception))
        self.assertNotIn('MSTEAMS_LAB_TENANT_ID', str(failure.exception))
        self.assertNotIn('private-tenant-value', str(failure.exception))

    def test_passing_parent_does_not_hide_skipped_child_in_output(self):
        report = ci.evaluate([
            {'Test': 'TestAccRead', 'Action': 'pass'},
            {'Test': 'TestAccRead/private-identifier', 'Action': 'skip'},
        ], ['TestAccRead'], 0)
        report.update(suite='resource-reads', auth_method='certificate')
        output = io.StringIO()
        with contextlib.redirect_stdout(output):
            ci.print_report(report)
        self.assertEqual(report['outcome'], 'failed')
        self.assertIn('1 child tests skip', output.getvalue())
        self.assertNotIn('private-identifier', output.getvalue())

    def test_crud_pass_requires_both_oracle_comparisons(self):
        test = 'TestAccApplicationAccessPolicyCRUD'
        events = [{'Test': test, 'Action': 'pass'}]
        events.extend({'Test': test, 'Action': 'output', 'Output': marker} for marker in ci.PROOF_MARKERS)
        self.assertEqual(ci.evaluate(events, [test], 0, True)['outcome'], 'failed')
        events.extend({'Test': test, 'Action': 'output', 'Output': ci.ORACLE_MARKER} for _ in range(2))
        self.assertEqual(ci.evaluate(events, [test], 0, True)['outcome'], 'passed')

    def test_raw_test_output_and_dynamic_child_names_are_not_published(self):
        events = [
            {'Test': 'TestAccRead', 'Action': 'output', 'Output': 'private-token-and-tenant-data'},
            {'Test': 'TestAccRead/private-fixture-name', 'Action': 'pass'},
            {'Test': 'TestAccRead', 'Action': 'pass'},
        ]
        report = json.dumps(ci.evaluate(events, ['TestAccRead'], 0))
        self.assertNotIn('private-', report)
        self.assertIn('passed', report)

    def test_mutation_cannot_pass_without_a_new_cleaned_journal(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            private = root / 'msteams-acceptance-fixture'
            private.mkdir()
            test = 'TestAccApplicationAccessPolicyCRUD'
            (root / 'suites.json').write_text(json.dumps({'crud': {'package': './fixture', 'tests': [test], 'mutation': True}}))

            def process(command, **kwargs):
                if command[0] == 'git':
                    return subprocess.CompletedProcess(command, 0, stdout='a' * 40 + '\n')
                events = [{'Test': test, 'Action': 'pass'}]
                events.extend({'Test': test, 'Action': 'output', 'Output': marker} for marker in ci.PROOF_MARKERS + [ci.ORACLE_MARKER] * 2)
                for event in events:
                    kwargs['stdout'].write(json.dumps(event) + '\n')
                return subprocess.CompletedProcess(command, 0)

            with patch.object(ci, 'HERE', root), patch.dict(os.environ, {'MSTEAMS_LAB_PRIVATE_DIRECTORY': str(private)}), patch.object(ci.subprocess, 'run', side_effect=process):
                with contextlib.redirect_stdout(io.StringIO()):
                    outcome = ci.run(argparse.Namespace(suite=['crud'], report_directory=root / 'reports'))
            self.assertEqual(outcome, 1)
            self.assertEqual(json.loads((root / 'reports/crud.json').read_text())['cleanup_journal_entries'], 0)

    def test_naming_guard_distinguishes_test_account_from_test_acc(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = root / 'example_test.go'
            source.write_text('package example\nfunc TestUnitRead() {}\nfunc TestAccRead() {}\nfunc TestMain() {}\n')
            self.assertEqual(ci.check_names(root), {'unit': 1, 'acceptance': 1})
            source.write_text('package example\nfunc TestAccountRead() {}\n')
            with self.assertRaises(ValueError):
                ci.check_names(root)

    def test_finalization_encrypts_only_journals_before_removing_private_data(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            private = root / 'msteams-acceptance-fixture'
            (private / 'journals').mkdir(parents=True)
            (private / 'identity.json').write_text('private-credential')
            (private / 'run.jsonl').write_text('private-test-output')
            (private / 'journals/fixture.json').write_text(json.dumps({'entries': [{'state': 'applied', 'fixture': 'owned-fixture'}]}))
            archived = []

            def encrypt(command, **kwargs):
                with tarfile.open(fileobj=io.BytesIO(kwargs['input'])) as archive:
                    archived.extend(archive.getnames())
                    self.assertNotIn(b'private-credential', kwargs['input'])
                    self.assertNotIn(b'private-test-output', kwargs['input'])
                Path(command[command.index('-out') + 1]).write_bytes(b'encrypted-recovery')
                return subprocess.CompletedProcess(command, 0)

            with patch.dict(os.environ, {'MSTEAMS_LAB_PRIVATE_DIRECTORY': str(private), 'RUNNER_TEMP': str(root)}), patch.object(ci.subprocess, 'run', side_effect=encrypt):
                with contextlib.redirect_stdout(io.StringIO()):
                    outcome = ci.finalize(argparse.Namespace(report_directory=root / 'reports'))
            self.assertEqual(outcome, 1)
            self.assertFalse(private.exists())
            self.assertEqual(archived, ['fixture.json'])
            self.assertEqual(json.loads((root / 'reports/cleanup.json').read_text())['unresolved_entries'], 1)

    def test_failed_encryption_retains_private_recovery_files(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            private = root / 'msteams-acceptance-fixture'
            private.mkdir()
            with patch.dict(os.environ, {'MSTEAMS_LAB_PRIVATE_DIRECTORY': str(private), 'RUNNER_TEMP': str(root)}), patch.object(ci.subprocess, 'run', return_value=subprocess.CompletedProcess([], 1)):
                with self.assertRaises(ValueError):
                    ci.finalize(argparse.Namespace(report_directory=root / 'reports'))
            self.assertTrue(private.exists())


if __name__ == '__main__':
    unittest.main()
