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
        self.assertIn('MSTEAMS_LAB_CALL_HOLD_AUDIO_FILE_ID', str(failure.exception))
        self.assertIn('MSTEAMS_LAB_USER_FIXTURE_JSON', str(failure.exception))
        self.assertIn('MSTEAMS_LAB_GROUP_FIXTURE_JSON', str(failure.exception))
        self.assertNotIn('MSTEAMS_LAB_TENANT_ID', str(failure.exception))
        self.assertNotIn('private-tenant-value', str(failure.exception))

    def test_voicemail_failure_keeps_the_assertion_without_raw_output(self):
        name = 'TestAccOnlineVoicemailReadOnly'
        reason = 'pinned voicemail projection unavailable'
        report = ci.evaluate([
            {'Test': name, 'Action': 'output', 'Output': reason + ': private-response'},
            {'Test': name, 'Action': 'fail'},
        ], [name], 1)
        report.update(suite='resource-reads', auth_method='client-secret')
        output = io.StringIO()
        with contextlib.redirect_stdout(output):
            ci.print_report(report)
        self.assertIn(name + ': ' + reason, output.getvalue())
        self.assertNotIn('private-response', output.getvalue() + json.dumps(report))

    def test_failure_reason_keeps_fixed_assertion_and_public_case_only(self):
        name = 'TestAccFederationReadOnly'
        reason = 'pinned federation read oracle failed; no output retained'
        events = [
            {'Test': name + '/acs', 'Action': 'output', 'Output': reason + ': private-response'},
            {'Test': name + '/private-user', 'Action': 'output', 'Output': 'Global federation read/list observations differ private-token'},
            {'Test': name, 'Action': 'fail'},
        ]
        report = ci.evaluate(events, [name], 1)
        report.update(suite='resource-reads', auth_method='client-secret')
        output = io.StringIO()
        with tempfile.TemporaryDirectory() as directory:
            summary = Path(directory) / 'summary.md'
            with patch.dict(os.environ, {'GITHUB_STEP_SUMMARY': str(summary)}), contextlib.redirect_stdout(output):
                ci.print_report(report)
                ci.write_summary([report])
            published = json.dumps(report) + output.getvalue() + summary.read_text()
        self.assertIn('acs: ' + reason, published)
        self.assertIn('Global federation read/list observations differ', published)
        self.assertNotIn('private-', published)

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

    def test_denial_suite_reports_the_verified_access_outcome(self):
        report = ci.evaluate([
            {'Test': 'TestAccPromptsCapabilityUnavailable', 'Action': 'output', 'Output': 'confirmed capability denial: HTTP 400, PolicyRP 40003; private-response'},
            {'Test': 'TestAccPromptsCapabilityUnavailable', 'Action': 'pass'},
        ], ['TestAccPromptsCapabilityUnavailable'], 0)
        report.update(suite='tenant-access-outcomes', auth_method='certificate')
        with tempfile.TemporaryDirectory() as directory:
            summary = Path(directory) / 'summary.md'
            output = io.StringIO()
            with patch.dict(os.environ, {'GITHUB_STEP_SUMMARY': str(summary)}), contextlib.redirect_stdout(output):
                ci.print_report(report)
                ci.write_summary([report])
            self.assertIn('verified HTTP 400; PolicyRP 40003', output.getvalue())
            self.assertIn('Denial check', summary.read_text())
            self.assertIn('| pass | 400 | 40003 |', summary.read_text())
            self.assertNotIn('private-response', output.getvalue() + summary.read_text())

    def test_crud_pass_requires_both_oracle_comparisons(self):
        test = 'TestAccApplicationAccessPolicyCRUD'
        events = [{'Test': test, 'Action': 'pass'}]
        events.extend({'Test': test, 'Action': 'output', 'Output': marker} for marker in ci.PROOF_MARKERS)
        self.assertEqual(ci.evaluate(events, [test], 0, True)['outcome'], 'failed')
        events.extend({'Test': test, 'Action': 'output', 'Output': ci.ORACLE_MARKER} for _ in range(2))
        self.assertEqual(ci.evaluate(events, [test], 0, True)['outcome'], 'passed')

    def test_each_mutation_requires_its_own_checks_and_oracle_comparisons(self):
        names = ['TestAccPolicyOneCRUD', 'TestAccPolicyTwoCRUD']
        events = []
        for name in names:
            events.append({'Test': name, 'Action': 'pass'})
            events.extend({'Test': name, 'Action': 'output', 'Output': marker} for marker in ci.PROOF_MARKERS)
        events.extend({'Test': names[0], 'Action': 'output', 'Output': ci.ORACLE_MARKER} for _ in range(2))
        self.assertEqual(ci.evaluate(events, names, 0, True)['outcome'], 'failed')
        events.extend({'Test': names[1], 'Action': 'output', 'Output': ci.ORACLE_MARKER} for _ in range(2))
        self.assertEqual(ci.evaluate(events, names, 0, True)['outcome'], 'passed')
        events = [event for event in events if not (event['Test'] == names[1] and event.get('Output') == ci.PROOF_MARKERS[0])]
        self.assertEqual(ci.evaluate(events, names, 0, True)['outcome'], 'failed')

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
            (root / 'suites.json').write_text(json.dumps({'crud': {'package': './fixture', 'tests': [test], 'mutation': True, 'proof': 'owned-crud', 'journal_entries_per_test': 1}}))

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

    def test_user_assignment_requires_clear_cleanup_and_three_comparisons(self):
        name = 'TestAccApplicationAccessPolicyUserAssignment'
        events = [{'Test': name, 'Action': 'pass'}]
        events += [{'Test': name, 'Action': 'output', 'Output': marker} for marker in ci.USER_ASSIGNMENT_MARKERS]
        events += [{'Test': name, 'Action': 'output', 'Output': ci.USER_ASSIGNMENT_ORACLE} for _ in range(2)]
        self.assertEqual(ci.evaluate(events, [name], 0, proof='user-assignment')['outcome'], 'failed')
        events.append({'Test': name, 'Action': 'output', 'Output': ci.USER_ASSIGNMENT_ORACLE})
        self.assertEqual(ci.evaluate(events, [name], 0, proof='user-assignment')['outcome'], 'passed')
        events = [event for event in events if event.get('Output') != ci.USER_ASSIGNMENT_MARKERS[1]]
        self.assertEqual(ci.evaluate(events, [name], 0, proof='user-assignment')['outcome'], 'failed')

    def test_user_fixture_secret_cannot_select_a_foreign_tenant_or_label(self):
        tenant = '00000000-0000-4000-8000-000000000001'
        fixture = {'tenant_id': tenant, 'user_id': '00000000-0000-4000-8000-000000000010',
                   'user_principal_name': 'go-sdk-msteams-acceptance@fixture.example',
                   'display_name': 'go-sdk-msteams acceptance user'}
        values = {'TENANT_ID': tenant, 'USER_FIXTURE_JSON': json.dumps(fixture)}
        self.assertEqual(ci.user_fixture(values), fixture)
        for change in [{'tenant_id': '00000000-0000-4000-8000-000000000002'},
                       {'user_id': 'private-user-id'}, {'user_principal_name': 'private-user@fixture.example'},
                       {'display_name': 'private-user'}, {'password': 'private-password'}]:
            with self.subTest(field=next(iter(change))):
                with self.assertRaises(ci.AcceptanceFailure) as failure:
                    ci.user_fixture(dict(values, USER_FIXTURE_JSON=json.dumps(dict(fixture, **change))))
                self.assertEqual(str(failure.exception), 'Invalid repository Actions secret: MSTEAMS_LAB_USER_FIXTURE_JSON')
                self.assertNotIn('private-', str(failure.exception))

    def test_cleaned_policy_without_assignment_restore_entry_cannot_pass(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            private = root / 'msteams-acceptance-fixture'
            (private / 'journals').mkdir(parents=True)
            name = 'TestAccApplicationAccessPolicyUserAssignment'
            (root / 'suites.json').write_text(json.dumps({'assign': {'package': './fixture', 'tests': [name],
                'mutation': True, 'proof': 'user-assignment', 'journal_entries_per_test': 2}}))

            def process(command, **kwargs):
                if command[0] == 'git':
                    return subprocess.CompletedProcess(command, 0, stdout='a' * 40 + '\n')
                events = [{'Test': name, 'Action': 'pass'}]
                events += [{'Test': name, 'Action': 'output', 'Output': marker}
                           for marker in ci.USER_ASSIGNMENT_MARKERS + [ci.USER_ASSIGNMENT_ORACLE] * 3]
                for event in events:
                    kwargs['stdout'].write(json.dumps(event) + '\n')
                (private / 'journals/only-policy.json').write_text(json.dumps({'entries': [{'state': 'cleaned'}]}))
                return subprocess.CompletedProcess(command, 0)

            with patch.object(ci, 'HERE', root), patch.dict(os.environ, {'MSTEAMS_LAB_PRIVATE_DIRECTORY': str(private)}), patch.object(ci.subprocess, 'run', side_effect=process):
                with contextlib.redirect_stdout(io.StringIO()):
                    result = ci.run(argparse.Namespace(suite=['assign'], report_directory=root / 'reports'))
            report = json.loads((root / 'reports/assign.json').read_text())
            self.assertEqual(result, 1)
            self.assertEqual(report['cleanup_journal_entries_cleaned'], 1)
            self.assertEqual(report['expected_cleanup_journal_entries'], 2)

    def test_group_assignment_requires_every_phase_and_six_comparisons(self):
        name = 'TestAccApplicationAccessPolicyGroupAssignment'
        events = [{'Test': name, 'Action': 'pass'}]
        events += [{'Test': name, 'Action': 'output', 'Output': marker} for marker in ci.GROUP_ASSIGNMENT_MARKERS]
        events += [{'Test': name, 'Action': 'output', 'Output': ci.GROUP_ASSIGNMENT_ORACLE} for _ in range(5)]
        self.assertEqual(ci.evaluate(events, [name], 0, proof='group-assignment')['outcome'], 'failed')
        events.append({'Test': name, 'Action': 'output', 'Output': ci.GROUP_ASSIGNMENT_ORACLE})
        self.assertEqual(ci.evaluate(events, [name], 0, proof='group-assignment')['outcome'], 'passed')
        for marker in ci.GROUP_ASSIGNMENT_MARKERS:
            with self.subTest(marker=marker):
                incomplete = [event for event in events if event.get('Output') != marker]
                self.assertEqual(ci.evaluate(incomplete, [name], 0, proof='group-assignment')['outcome'], 'failed')

    def test_group_fixture_secret_is_bound_to_the_dedicated_empty_security_group(self):
        tenant = '00000000-0000-4000-8000-000000000001'
        fixture = {'tenant_id': tenant, 'group_id': '00000000-0000-4000-8000-000000000020',
                   'display_name': 'go-sdk-msteams acceptance policy group',
                   'mail_nickname': 'go-sdk-msteams-acceptance-policy', 'security_enabled': True,
                   'mail_enabled': False, 'role_assignable': False, 'empty_membership_verified': True}
        values = {'TENANT_ID': tenant, 'GROUP_FIXTURE_JSON': json.dumps(fixture)}
        self.assertEqual(ci.group_fixture(values), fixture)
        for change in [{'tenant_id': '00000000-0000-4000-8000-000000000002'},
                       {'group_id': 'private-group'}, {'display_name': 'private-group'},
                       {'mail_nickname': 'private-group'}, {'security_enabled': False},
                       {'mail_enabled': True}, {'role_assignable': True},
                       {'empty_membership_verified': False}, {'role_assignable': 0},
                       {'password': 'private-password'}]:
            with self.subTest(field=next(iter(change))):
                with self.assertRaises(ci.AcceptanceFailure) as failure:
                    ci.group_fixture(dict(values, GROUP_FIXTURE_JSON=json.dumps(dict(fixture, **change))))
                self.assertEqual(str(failure.exception), 'Invalid repository Actions secret: MSTEAMS_LAB_GROUP_FIXTURE_JSON')

    def test_aggregate_requires_every_suite_in_both_authentication_modes(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'suites.json').write_text(json.dumps({'reads': {}, 'writes': {}}))
            reports = root / 'reports'
            reports.mkdir()
            for mode in ['certificate', 'client-secret']:
                for suite in ['reads', 'writes']:
                    (reports / (mode + '-' + suite + '.json')).write_text(json.dumps({
                        'suite': suite, 'auth_method': mode, 'outcome': 'passed',
                        'tests': [{'name': 'TestAccFixture', 'outcome': 'pass'}]}))
            with patch.object(ci, 'HERE', root), contextlib.redirect_stdout(io.StringIO()):
                self.assertEqual(ci.summarize(argparse.Namespace(report_directory=reports)), 0)
                (reports / 'client-secret-writes.json').unlink()
                self.assertEqual(ci.summarize(argparse.Namespace(report_directory=reports)), 1)
                (reports / 'duplicate.json').write_text((reports / 'certificate-reads.json').read_text())
                self.assertEqual(ci.summarize(argparse.Namespace(report_directory=reports)), 1)

    def test_read_selection_cannot_enable_mutations(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            private = root / 'msteams-acceptance-fixture'
            private.mkdir()
            (root / 'suites.json').write_text(json.dumps({
                'reads': {'package': './fixture', 'tests': ['TestAccRead']},
                'writes': {'package': './fixture', 'tests': ['TestAccWrite'], 'mutation': True,
                           'proof': 'owned-crud', 'journal_entries_per_test': 1}}))
            invoked = []

            def process(command, **kwargs):
                if command[0] == 'git':
                    return subprocess.CompletedProcess(command, 0, stdout='a' * 40 + '\n')
                invoked.append((command, kwargs['env']['MSTEAMS_LAB_MUTATIONS']))
                kwargs['stdout'].write(json.dumps({'Test': 'TestAccRead', 'Action': 'pass'}) + '\n')
                return subprocess.CompletedProcess(command, 0)

            with patch.object(ci, 'HERE', root), patch.dict(os.environ, {'MSTEAMS_LAB_PRIVATE_DIRECTORY': str(private)}), patch.object(ci.subprocess, 'run', side_effect=process):
                with contextlib.redirect_stdout(io.StringIO()):
                    result = ci.run(argparse.Namespace(suite=None, kind='reads', report_directory=root / 'reports'))
            self.assertEqual(result, 0)
            self.assertEqual(len(invoked), 1)
            self.assertEqual(invoked[0][1], '0')
            self.assertIn('TestAccRead', ' '.join(invoked[0][0]))
            self.assertNotIn('TestAccWrite', ' '.join(invoked[0][0]))
            self.assertFalse((root / 'reports/writes.json').exists())

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
