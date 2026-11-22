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


def registry(*steps, service='svc', package='./fixture'):
    """A registry with one auth step and one service partition."""
    return {'auth': {'package': './auth', 'steps': [{'id': 'authentication', 'tests': ['TestAccAuth']}]},
            'services': {service: {'package': package, 'steps': list(steps)}}}


def service_run(root, **kwargs):
    return argparse.Namespace(auth=False, service='svc', step=None, report_directory=root / 'reports', **kwargs)


class AcceptanceGuardTests(unittest.TestCase):
    def test_skip_selection_requires_exact_registered_names_and_public_reasons(self):
        suites = registry({'id': 'crud', 'tests': ['TestAccRead', 'TestAccPhone']})
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'skips.json').write_text(json.dumps({'TestAccPhone': 'No lab phone inventory.'}))
            with patch.object(ci, 'HERE', root), patch.dict(os.environ, {}, clear=True):
                self.assertEqual(ci.selected_skips(suites), {'TestAccPhone': 'No lab phone inventory.'})
                self.assertEqual(ci.selected_skips(suites, argparse.Namespace(include_test=['TestAccPhone'])), {})
                self.assertEqual(ci.selected_skips(suites, argparse.Namespace(skip_test=['TestAccRead'], skip_reason='Maintenance.')),
                                 {'TestAccPhone': 'No lab phone inventory.', 'TestAccRead': 'Maintenance.'})
                for args in [
                    argparse.Namespace(skip_test=['TestAccRead']),
                    argparse.Namespace(skip_test=['TestAccUnknown'], skip_reason='Maintenance.'),
                    argparse.Namespace(skip_test=['TestAcc.*'], skip_reason='Maintenance.'),
                    argparse.Namespace(include_test=['TestAccRead']),
                    argparse.Namespace(include_test=['TestAccPhone'], skip_test=['TestAccPhone'], skip_reason='Maintenance.'),
                    argparse.Namespace(skip_test=['TestAccRead'], skip_reason='reason\n::error::injection'),
                ]:
                    with self.subTest(args=args), self.assertRaises(ci.AcceptanceFailure):
                        ci.selected_skips(suites, args)
                with patch.dict(os.environ, {'MSTEAMS_ACC_SKIP_TESTS': 'TestAccRead', 'MSTEAMS_ACC_SKIP_REASON': 'Maintenance.', 'MSTEAMS_ACC_INCLUDE_TESTS': 'TestAccPhone'}):
                    self.assertEqual(ci.selected_skips(suites), {'TestAccRead': 'Maintenance.'})
                (root / 'skips.json').write_text(json.dumps({'TestAccUnknown': 'Stale selection.'}))
                with self.assertRaises(ci.AcceptanceFailure):
                    ci.selected_skips(suites)

    def test_selective_crud_skip_runs_other_tests_and_counts_only_their_journals(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            private = root / 'private'
            journals = private / 'journals'
            journals.mkdir(parents=True)
            active, skipped = 'TestAccRead', 'TestAccPhone'
            (root / 'suites.json').write_text(json.dumps(registry({'id': 'crud', 'tests': [active, skipped], 'mutation': True, 'proof': 'owned-crud', 'journal_entries_per_test': 1})))
            (root / 'skips.json').write_text(json.dumps({skipped: 'No lab phone inventory.'}))
            invoked = []

            def process(command, **kwargs):
                if command[0] == 'git':
                    return subprocess.CompletedProcess(command, 0, stdout='a' * 40 + '\n')
                invoked.append(command)
                for marker in ci.PROOF_MARKERS + [ci.ORACLE_MARKER] * 2:
                    kwargs['stdout'].write(json.dumps({'Test': active, 'Action': 'output', 'Output': marker}) + '\n')
                kwargs['stdout'].write(json.dumps({'Test': active, 'Action': 'pass'}) + '\n')
                (journals / 'owned.json').write_text(json.dumps({'entries': [{'state': 'cleaned'}]}))
                return subprocess.CompletedProcess(command, 0)

            with patch.object(ci, 'HERE', root), patch.dict(os.environ, {'MSTEAMS_LAB_PRIVATE_DIRECTORY': str(private)}, clear=True), patch.object(ci.subprocess, 'run', side_effect=process), contextlib.redirect_stdout(io.StringIO()):
                self.assertEqual(ci.run(service_run(root)), 0)
            self.assertEqual(len(invoked), 1)
            pattern = invoked[0][invoked[0].index('-run') + 1]
            self.assertEqual(pattern, '^(TestAccRead)$')
            report = json.loads((root / 'reports/svc--crud.json').read_text())
            self.assertEqual(report['expected_cleanup_journal_entries'], 1)
            self.assertEqual((report['suite'], report['service'], report['step'], report['phase']), ('svc/crud', 'svc', 'crud', 'writes'))
            self.assertEqual(report['cleanup_journal_entries_cleaned'], 1)
            self.assertEqual(report['tests'][1], {'name': skipped, 'outcome': 'skip', 'skip_reason': 'No lab phone inventory.'})
            self.assertTrue(ci.report_matches_selection(report, {'tests': [active, skipped]}, {skipped: 'No lab phone inventory.'}))

    def test_entire_skipped_suite_never_starts_go_or_counts_as_passed(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            private = root / 'private'
            private.mkdir()
            (root / 'suites.json').write_text(json.dumps(registry({'id': 'phone', 'tests': ['TestAccPhone'], 'mutation': True, 'proof': 'owned-crud', 'journal_entries_per_test': 1})))
            (root / 'skips.json').write_text(json.dumps({'TestAccPhone': 'No lab phone inventory.'}))
            with patch.object(ci, 'HERE', root), patch.dict(os.environ, {'MSTEAMS_LAB_PRIVATE_DIRECTORY': str(private)}, clear=True), patch.object(ci.subprocess, 'run', return_value=subprocess.CompletedProcess([], 0, stdout='a' * 40 + '\n')) as process, contextlib.redirect_stdout(io.StringIO()) as output:
                self.assertEqual(ci.run(service_run(root)), 0)
            self.assertEqual(process.call_count, 1)
            self.assertEqual(process.call_args.args[0][0], 'git')
            report = json.loads((root / 'reports/svc--phone.json').read_text())
            self.assertEqual(report['outcome'], 'skipped')
            self.assertEqual(report['expected_cleanup_journal_entries'], 0)
            self.assertIn('SKIP TestAccPhone: No lab phone inventory.', output.getvalue())
            self.assertNotIn('FAIL', output.getvalue())
            self.assertFalse(list(private.glob('*.jsonl')))

    def test_aggregate_accepts_only_the_requested_skips_and_displays_reasons(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            reports = root / 'reports'
            reports.mkdir()
            (root / 'suites.json').write_text(json.dumps(registry({'id': 'crud', 'tests': ['TestAccRead', 'TestAccPhone']})))
            (root / 'skips.json').write_text(json.dumps({'TestAccPhone': 'No lab phone inventory.'}))
            (reports / 'auth.json').write_text(json.dumps({'suite': 'auth/authentication', 'auth_method': 'all', 'outcome': 'passed',
                                                           'tests': [{'name': 'TestAccAuth', 'outcome': 'pass'}]}))
            (reports / 'svc.json').write_text(json.dumps({'suite': 'svc/crud', 'auth_method': 'client-secret', 'outcome': 'passed', 'tests': [
                {'name': 'TestAccRead', 'outcome': 'pass'},
                {'name': 'TestAccPhone', 'outcome': 'skip', 'skip_reason': 'No lab phone inventory.'}]}))
            summary = root / 'summary.md'
            args = argparse.Namespace(report_directory=reports)
            with patch.object(ci, 'HERE', root), patch.dict(os.environ, {'GITHUB_STEP_SUMMARY': str(summary)}, clear=True), contextlib.redirect_stdout(io.StringIO()):
                self.assertEqual(ci.summarize(args), 0)
                rendered = summary.read_text()
                self.assertIn('### Configured skips', rendered)
                self.assertIn('No lab phone inventory.', rendered)
                self.assertIn('provide no acceptance evidence', rendered)
                self.assertNotIn('Failed test', rendered)
                # A test's own t.Skip is still a failure, even if the parent report says passed.
                path = reports / 'svc.json'
                original = json.loads(path.read_text())
                for change in ['runtime-skip', 'wrong-reason', 'missing', 'duplicate', 'include-configured']:
                    report = json.loads(json.dumps(original))
                    overrides = args
                    if change == 'runtime-skip':
                        report['tests'][0]['outcome'] = 'skip'
                    elif change == 'wrong-reason':
                        report['tests'][1]['skip_reason'] = 'Different reason.'
                    elif change == 'missing':
                        report['tests'].pop()
                    elif change == 'duplicate':
                        report['tests'].append(report['tests'][1])
                    else:
                        overrides = argparse.Namespace(report_directory=reports, include_test=['TestAccPhone'])
                    path.write_text(json.dumps(report))
                    with self.subTest(change=change):
                        self.assertEqual(ci.summarize(overrides), 1)

    def test_populated_shared_calling_requires_list_edits_and_four_comparisons(self):
        test = 'TestAccSharedCallingRoutingPolicyPopulatedCRUD'
        proof = 'owned-shared-calling-populated-crud'
        markers, oracle, count = ci.PROOFS[proof]
        events = [{'Test': test, 'Action': 'output', 'Output': marker} for marker in markers + [oracle] * count]
        events.append({'Test': test, 'Action': 'pass'})
        self.assertEqual(ci.evaluate(events, [test], 0, proof=proof)['outcome'], 'passed')
        for absent in ['shared calling populated list edit/clear: passed', oracle]:
            missing = events.copy()
            missing.remove(next(event for event in missing if event.get('Output') == absent))
            with self.subTest(absent=absent):
                self.assertEqual(ci.evaluate(missing, [test], 0, proof=proof)['outcome'], 'failed')

    def test_optional_phone_fixture_rejects_foreign_tenants_and_dial_codes(self):
        tenant = '11111111-1111-1111-1111-111111111111'
        fixture = {'tenant_id': tenant, 'resource_account_id': '22222222-2222-2222-2222-222222222222',
                   'emergency_numbers': ['+12025550101', '+12025550102']}
        # Parser vectors only; these reserved numbers are never submitted to a tenant.
        self.assertEqual(ci.shared_calling_fixture(tenant, json.dumps(fixture)), fixture)
        for update in [
            {'tenant_id': '33333333-3333-3333-3333-333333333333'},
            {'resource_account_id': 'not-an-account-id'},
            {'emergency_numbers': ['112', '999']},
            {'emergency_numbers': ['+12025550101']},
            {'emergency_numbers': ['+12025550101', '+12025550101']},
            {'emergency_numbers': ['+12025550101', None]},
            {'unexpected': 'private-value'},
        ]:
            with self.subTest(update=update), self.assertRaises(ci.AcceptanceFailure) as failure:
                ci.shared_calling_fixture(tenant, json.dumps(fixture | update))
            self.assertNotIn('private-value', str(failure.exception))

    def test_read_mismatch_diagnostics_publish_only_fixed_reasons(self):
        test = 'TestAccNotificationAndFeedsPolicyReadOnly'
        events = [
            {'Test': test, 'Action': 'output', 'Output': 'read/list comparison: metadata values differ private-tenant-value'},
            {'Test': test, 'Action': 'fail'},
        ]
        report = ci.evaluate(events, [test], 1)
        self.assertEqual(report['failure_reasons'][test], ['read/list comparison: metadata values differ'])
        self.assertNotIn('private-tenant-value', json.dumps(report))

    def test_metadata_paths_publish_fixed_schema_names_without_values(self):
        test = 'TestAccRoomVideoTeleConferencingPolicyReadOnly'
        paths = ['Key.ScopeClass', 'Key.SchemaId', 'Key.AuthorityId',
                 'Key.DefaultXml.SchemaId', 'Key.DefaultXml.Data', 'Key.DefaultXml.ConfigObject',
                 'Key.DefaultXml.Signature', 'Key.DefaultXml.IsModified', 'Key.XmlRoot',
                 'ConfigMetadata', 'ConfigId']
        for path in paths:
            reason = 'read/list metadata: ' + path + ' differs'
            report = ci.evaluate([
                {'Test': test, 'Action': 'output', 'Output': reason + ' private-value'},
                {'Test': test, 'Action': 'output', 'Output': 'read/list metadata: private-property differs'},
                {'Test': test, 'Action': 'fail'},
            ], [test], 1)
            self.assertEqual(report['failure_reasons'][test], [reason])
            self.assertNotIn('private-', json.dumps(report))

    def test_populated_team_templates_require_edit_clear_and_four_comparisons(self):
        test = 'TestAccTemplatePermissionPopulatedCRUD'
        proof = 'owned-template-permission-populated-crud'
        markers, oracle, count = ci.PROOFS[proof]
        events = [{'Test': test, 'Action': 'output', 'Output': marker} for marker in markers + [oracle] * count]
        events.append({'Test': test, 'Action': 'pass'})
        self.assertEqual(ci.evaluate(events, [test], 0, proof=proof)['outcome'], 'passed')
        for absent in ['policy hidden-list edit/read-back: passed', 'policy hidden-list clear/read-back: passed', oracle]:
            missing = events.copy()
            missing.remove(next(event for event in missing if event.get('Output') == absent))
            with self.subTest(absent=absent):
                self.assertEqual(ci.evaluate(missing, [test], 0, proof=proof)['outcome'], 'failed')

    def test_update_management_requires_datetime_list_phases_and_five_comparisons(self):
        test = 'TestAccUpdateManagementPolicyCRUD'
        proof = 'owned-update-management-crud'
        markers, oracle, count = ci.PROOFS[proof]
        events = [{'Test': test, 'Action': 'output', 'Output': marker} for marker in markers + [oracle] * count]
        events.append({'Test': test, 'Action': 'pass'})
        self.assertEqual(ci.evaluate(events, [test], 0, proof=proof)['outcome'], 'passed')
        for absent in ['update-management DateTime update/read-back: passed', 'update-management list edit/read-back: passed', 'update-management list clear/read-back: passed', oracle]:
            missing = events.copy()
            missing.remove(next(event for event in missing if event.get('Output') == absent))
            with self.subTest(absent=absent):
                self.assertEqual(ci.evaluate(missing, [test], 0, proof=proof)['outcome'], 'failed')

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
        report.update(suite='svc/reads', auth_method='client-secret')
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

    def test_fixed_assertions_keep_the_reason_without_raw_output(self):
        for name, reason in [
            ('TestAccOnlineVoicemailReadOnly', 'pinned voicemail projection unavailable'),
            ('TestAccMobilityPolicyReadOnly', 'Go and PowerShell compiled scalar getters differ'),
            ('TestAccMultiTenantOrganizationConfigurationUpdateRestore', 'multi-tenant original state not independently restored'),
            ('TestAccRecordingRollOutPolicyCRUD', 'recording rollout sparse update included an unbound field'),
        ]:
            with self.subTest(name=name):
                report = ci.evaluate([
                    {'Test': name, 'Action': 'output', 'Output': reason + ': private-response'},
                    {'Test': name, 'Action': 'fail'},
                ], [name], 1)
                report.update(suite='svc/reads', auth_method='client-secret')
                output = io.StringIO()
                with contextlib.redirect_stdout(output):
                    ci.print_report(report)
                self.assertIn(name + ': ' + reason, output.getvalue())
                self.assertNotIn('private-response', output.getvalue() + json.dumps(report))

    def test_failure_reason_keeps_fixed_assertion_only(self):
        name = 'TestAccACSFederationConfigurationReadOnly'
        reason = 'pinned federation read oracle failed; no output retained'
        events = [
            {'Test': name, 'Action': 'output', 'Output': reason + ': private-response'},
            {'Test': name + '/private-user', 'Action': 'output', 'Output': 'Global federation read/list observations differ private-token'},
            {'Test': name, 'Action': 'fail'},
        ]
        report = ci.evaluate(events, [name], 1)
        report.update(suite='svc/reads', auth_method='client-secret')
        output = io.StringIO()
        with tempfile.TemporaryDirectory() as directory:
            summary = Path(directory) / 'summary.md'
            with patch.dict(os.environ, {'GITHUB_STEP_SUMMARY': str(summary)}), contextlib.redirect_stdout(output):
                ci.print_report(report)
                ci.write_summary([report])
            published = json.dumps(report) + output.getvalue() + summary.read_text()
        self.assertIn(reason, report['failure_reasons'][name])
        self.assertIn(name + ': Global federation read/list observations differ; ' + reason, published)
        self.assertIn('Global federation read/list observations differ', published)
        self.assertNotIn('private-', published)

    def test_passing_parent_does_not_hide_skipped_child_in_output(self):
        report = ci.evaluate([
            {'Test': 'TestAccRead', 'Action': 'pass'},
            {'Test': 'TestAccRead/private-identifier', 'Action': 'skip'},
        ], ['TestAccRead'], 0)
        report.update(suite='svc/reads', auth_method='client-secret')
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
        report.update(suite='assignment_migration/tenant-access-outcomes', step='tenant-access-outcomes', auth_method='client-secret')
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

    def test_files_oracle_failure_retains_fixed_reason_and_stage(self):
        name = 'TestAccFilesPolicyCRUD'
        reason = 'pinned files policy projection unavailable'
        events = [
            {'Test': name, 'Action': 'output', 'Output': 'oracle failure stage: read private-response'},
            {'Test': name, 'Action': 'output', 'Output': reason + ' private-token'},
            {'Test': name, 'Action': 'fail'},
        ]
        report = ci.evaluate(events, [name], 1)
        self.assertEqual(report['failure_reasons'][name], ['oracle failure stage: read', reason])
        self.assertNotIn('private-', json.dumps(report))

    def test_assertion_catalog_excludes_format_arguments_and_dynamic_output(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'resource_test.go').write_text('''package resource
func TestAccRead(t *testing.T) {
 t.Fatalf("policy parity differs for %s: %v", privateID, privateResponse)
 t.Fatal("policy parity differs")
 t.Log("private-log-message")
}
''')
            reasons = ci.assertion_reasons(root)
            self.assertIn('policy parity differs for', reasons)
            self.assertNotIn('private-log-message', reasons)
            with patch.object(ci, 'assertion_reasons', return_value=reasons):
                report = ci.evaluate([
                    {'Test': 'TestAccRead', 'Action': 'output', 'Output': 'resource_test.go:3: policy parity differs for private-id: private-response'},
                    {'Test': 'TestAccRead', 'Action': 'fail'},
                ], ['TestAccRead'], 1)
            self.assertEqual(report['failure_reasons']['TestAccRead'], ['policy parity differs for'])
            self.assertNotIn('private-', json.dumps(report))

    def test_mutation_cannot_pass_without_a_new_cleaned_journal(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            private = root / 'msteams-acceptance-fixture'
            private.mkdir()
            test = 'TestAccApplicationAccessPolicyCRUD'
            (root / 'suites.json').write_text(json.dumps(registry({'id': 'crud', 'tests': [test], 'mutation': True, 'proof': 'owned-crud', 'journal_entries_per_test': 1})))

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
                    outcome = ci.run(service_run(root))
            self.assertEqual(outcome, 1)
            self.assertEqual(json.loads((root / 'reports/svc--crud.json').read_text())['cleanup_journal_entries'], 0)

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
            (root / 'suites.json').write_text(json.dumps(registry({'id': 'assign', 'tests': [name],
                'mutation': True, 'proof': 'user-assignment', 'journal_entries_per_test': 2})))

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
                    result = ci.run(service_run(root))
            report = json.loads((root / 'reports/svc--assign.json').read_text())
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

    def test_template_variants_each_require_crud_and_two_comparisons(self):
        name = 'TestAccTeamTemplateCRUD'
        children = {name: ['expanded', 'json-body']}
        events = [{'Test': name, 'Action': 'pass'}]
        for child in children[name]:
            events.append({'Test': name + '/' + child, 'Action': 'pass'})
            events += [{'Test': name + '/' + child, 'Action': 'output', 'Output': marker}
                       for marker in ci.TEMPLATE_MARKERS + [ci.TEMPLATE_ORACLE] * 2]
        self.assertEqual(ci.evaluate(events, [name], 0, proof='template-crud', expected_children=children)['outcome'], 'passed')
        for removed in [
            lambda event: event['Test'].endswith('/json-body') and event['Action'] == 'pass',
            lambda event: event['Test'].endswith('/json-body') and event.get('Output') == ci.TEMPLATE_MARKERS[2],
            lambda event: event['Test'].endswith('/json-body') and event.get('Output') == ci.TEMPLATE_ORACLE,
        ]:
            incomplete = [event for event in events if not removed(event)]
            self.assertEqual(ci.evaluate(incomplete, [name], 0, proof='template-crud', expected_children=children)['outcome'], 'failed')
        skipped = [dict(event, Action='skip') if event['Test'].endswith('/json-body') and event['Action'] == 'pass' else event for event in events]
        self.assertEqual(ci.evaluate(skipped, [name], 0, proof='template-crud', expected_children=children)['outcome'], 'failed')

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

    def test_aggregate_requires_one_auth_report_and_client_secret_service_reports(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'suites.json').write_text(json.dumps(registry({'id': 'reads', 'tests': ['TestAccFixture']},
                                                                  {'id': 'crud', 'tests': ['TestAccWrite'], 'mutation': True,
                                                                   'proof': 'owned-crud', 'journal_entries_per_test': 1})))
            reports = root / 'reports'
            reports.mkdir()

            def write(name, suite, mode, test):
                (reports / (name + '.json')).write_text(json.dumps({
                    'suite': suite, 'auth_method': mode, 'outcome': 'passed', 'tests': [{'name': test, 'outcome': 'pass'}]}))

            write('auth--authentication', 'auth/authentication', 'all', 'TestAccAuth')
            write('svc--reads', 'svc/reads', 'client-secret', 'TestAccFixture')
            write('svc--crud', 'svc/crud', 'client-secret', 'TestAccWrite')
            with patch.object(ci, 'HERE', root), patch.dict(os.environ, {}, clear=True), contextlib.redirect_stdout(io.StringIO()):
                self.assertEqual(ci.summarize(argparse.Namespace(report_directory=reports)), 0)
                # A service report from certificate auth is unexpected, not extra evidence.
                write('svc--reads-certificate', 'svc/reads', 'certificate', 'TestAccFixture')
                self.assertEqual(ci.summarize(argparse.Namespace(report_directory=reports)), 1)
                (reports / 'svc--reads-certificate.json').unlink()
                (reports / 'auth--authentication.json').unlink()
                self.assertEqual(ci.summarize(argparse.Namespace(report_directory=reports)), 1)
                write('auth--authentication', 'auth/authentication', 'all', 'TestAccAuth')
                (reports / 'svc--crud.json').unlink()
                self.assertEqual(ci.summarize(argparse.Namespace(report_directory=reports)), 1)
                # A dispatch that reruns selected services expects only their reports.
                with patch.dict(os.environ, {'MSTEAMS_ACC_SERVICES': 'svc'}):
                    self.assertEqual(ci.summarize(argparse.Namespace(report_directory=reports)), 1)
                write('svc--crud', 'svc/crud', 'client-secret', 'TestAccWrite')
                (reports / 'duplicate.json').write_text((reports / 'svc--reads.json').read_text())
                self.assertEqual(ci.summarize(argparse.Namespace(report_directory=reports)), 1)

    def test_service_steps_run_in_order_and_only_mutating_steps_enable_mutations(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            private = root / 'msteams-acceptance-fixture'
            (private / 'journals').mkdir(parents=True)
            (root / 'suites.json').write_text(json.dumps(registry(
                {'id': 'reads', 'tests': ['TestAccRead']},
                {'id': 'crud', 'tests': ['TestAccWrite'], 'mutation': True, 'proof': 'owned-crud', 'journal_entries_per_test': 1})))
            invoked = []

            def process(command, **kwargs):
                if command[0] == 'git':
                    return subprocess.CompletedProcess(command, 0, stdout='a' * 40 + '\n')
                name = command[command.index('-run') + 1][2:-2]
                invoked.append((name, kwargs['env']['MSTEAMS_LAB_MUTATIONS'], kwargs['env']['MSTEAMS_LAB_AUTH_METHOD']))
                kwargs['stdout'].write(json.dumps({'Test': name, 'Action': 'pass'}) + '\n')
                return subprocess.CompletedProcess(command, 0)

            with patch.object(ci, 'HERE', root), patch.dict(os.environ, {'MSTEAMS_LAB_PRIVATE_DIRECTORY': str(private)}, clear=True), patch.object(ci.subprocess, 'run', side_effect=process):
                with contextlib.redirect_stdout(io.StringIO()):
                    ci.run(service_run(root))
                    self.assertEqual(invoked, [('TestAccRead', '0', 'client-secret'), ('TestAccWrite', '1', 'client-secret')])
                    invoked.clear()
                    self.assertEqual(ci.run(argparse.Namespace(auth=False, service='svc', step=['reads'], report_directory=root / 'only-reads')), 0)
            self.assertEqual(invoked, [('TestAccRead', '0', 'client-secret')])
            self.assertFalse((root / 'only-reads/svc--crud.json').exists())

    def test_failed_step_blocks_later_mutations_of_the_service(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            private = root / 'msteams-acceptance-fixture'
            (private / 'journals').mkdir(parents=True)
            (root / 'suites.json').write_text(json.dumps(registry(
                {'id': 'reads', 'tests': ['TestAccRead']},
                {'id': 'crud', 'tests': ['TestAccWrite'], 'mutation': True, 'proof': 'owned-crud', 'journal_entries_per_test': 1})))
            invoked = []

            def process(command, **kwargs):
                if command[0] == 'git':
                    return subprocess.CompletedProcess(command, 0, stdout='a' * 40 + '\n')
                invoked.append(command)
                kwargs['stdout'].write(json.dumps({'Test': 'TestAccRead', 'Action': 'fail'}) + '\n')
                return subprocess.CompletedProcess(command, 1)

            with patch.object(ci, 'HERE', root), patch.dict(os.environ, {'MSTEAMS_LAB_PRIVATE_DIRECTORY': str(private)}, clear=True), patch.object(ci.subprocess, 'run', side_effect=process):
                with contextlib.redirect_stdout(io.StringIO()):
                    self.assertEqual(ci.run(service_run(root)), 1)
            self.assertEqual(len(invoked), 1)
            report = json.loads((root / 'reports/svc--crud.json').read_text())
            self.assertEqual(report['outcome'], 'failed')
            self.assertTrue(report['blocked_by_earlier_step'])
            self.assertEqual(report['tests'], [{'name': 'TestAccWrite', 'outcome': 'not-run'}])

    def test_service_partitions_refuse_certificate_authentication(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'suites.json').write_text(json.dumps(registry({'id': 'reads', 'tests': ['TestAccRead']})))
            environment = {'MSTEAMS_LAB_PRIVATE_DIRECTORY': str(root), 'MSTEAMS_LAB_AUTH_METHOD': 'certificate'}
            with patch.object(ci, 'HERE', root), patch.dict(os.environ, environment, clear=True), patch.object(ci.subprocess, 'run', return_value=subprocess.CompletedProcess([], 0, stdout='a' * 40 + '\n')) as process:
                with self.assertRaises(ci.AcceptanceFailure):
                    ci.run(service_run(root))
            self.assertEqual([call.args[0][0] for call in process.call_args_list], ['git'])

    def test_auth_partition_loads_every_credential_kind(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'suites.json').write_text(json.dumps(registry({'id': 'reads', 'tests': ['TestAccRead']})))
            invoked = []

            def process(command, **kwargs):
                if command[0] == 'git':
                    return subprocess.CompletedProcess(command, 0, stdout='a' * 40 + '\n')
                invoked.append((command[-1], kwargs['env']['MSTEAMS_LAB_AUTH_METHOD']))
                kwargs['stdout'].write(json.dumps({'Test': 'TestAccAuth', 'Action': 'pass'}) + '\n')
                return subprocess.CompletedProcess(command, 0)

            for mode in ['certificate', 'client-secret']:
                environment = {'MSTEAMS_LAB_PRIVATE_DIRECTORY': str(root), 'MSTEAMS_LAB_AUTH_METHOD': mode}
                with self.subTest(mode=mode), patch.object(ci, 'HERE', root), patch.dict(os.environ, environment, clear=True), patch.object(ci.subprocess, 'run', side_effect=process):
                    with self.assertRaises(ci.AcceptanceFailure):
                        ci.run(argparse.Namespace(auth=True, service=None, step=None, report_directory=root / 'reports'))
            environment = {'MSTEAMS_LAB_PRIVATE_DIRECTORY': str(root), 'MSTEAMS_LAB_AUTH_METHOD': 'all'}
            with patch.object(ci, 'HERE', root), patch.dict(os.environ, environment, clear=True), patch.object(ci.subprocess, 'run', side_effect=process):
                with contextlib.redirect_stdout(io.StringIO()):
                    self.assertEqual(ci.run(argparse.Namespace(auth=True, service=None, step=None, report_directory=root / 'reports')), 0)
            self.assertEqual(invoked, [('./auth', 'all')])
            report = json.loads((root / 'reports/auth--authentication.json').read_text())
            self.assertEqual((report['suite'], report['partition'], report['auth_method']), ('auth/authentication', 'auth', 'all'))

    def test_naming_guard_distinguishes_test_account_from_test_acc(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = root / 'example_test.go'
            source.write_text('package example\nfunc TestUnitRead() {}\nfunc TestAccRead() {}\nfunc TestMain() {}\n')
            self.assertEqual(ci.check_names(root), {'unit': 1, 'acceptance': 1})
            source.write_text('package example\nfunc TestAccountRead() {}\n')
            with self.assertRaises(ValueError):
                ci.check_names(root)

    def test_plan_locks_services_with_serialized_reads_or_mutations(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            output = root / 'github-output'
            fixture = {'auth': registry()['auth'], 'services': {
                'platform': {'package': './tests/integration/lab/platform', 'steps': [{'id': 'reads', 'tests': ['TestAccOrdinaryRead']}]},
                'teams_messaging_configuration': {'package': './tests/integration/lab/teams_messaging_configuration',
                                                  'steps': [{'id': 'global-reads', 'serial': True, 'tests': ['TestAccSingletonRead']}]},
                'teams_call_hold_policy': {'package': './tests/integration/lab/teams_call_hold_policy', 'steps': [
                    {'id': 'reads', 'tests': ['TestAccCallHoldRead']},
                    {'id': 'crud', 'mutation': True, 'proof': 'owned-crud', 'journal_entries_per_test': 1, 'tests': ['TestAccCallHoldWrite']}]}}}
            (root / 'suites.json').write_text(json.dumps(fixture))
            with patch.object(ci, 'HERE', root), patch.dict(os.environ, {'GITHUB_OUTPUT': str(output)}, clear=True), contextlib.redirect_stdout(io.StringIO()):
                self.assertEqual(ci.plan(argparse.Namespace()), 0)
                lines = dict(line.split('=', 1) for line in output.read_text().splitlines())
                self.assertNotIn('auth', lines)
                self.assertEqual(json.loads(lines['services_unlocked']), ['platform'])
                self.assertEqual(json.loads(lines['services_locked']), ['teams_messaging_configuration', 'teams_call_hold_policy'])
                output.write_text('')
                with patch.dict(os.environ, {'MSTEAMS_ACC_SERVICES': 'teams_call_hold_policy'}):
                    ci.plan(argparse.Namespace())
                lines = dict(line.split('=', 1) for line in output.read_text().splitlines())
                self.assertEqual((json.loads(lines['services_unlocked']), json.loads(lines['services_locked'])), ([], ['teams_call_hold_policy']))
                with patch.dict(os.environ, {'MSTEAMS_ACC_SERVICES': 'teams_unknown'}), self.assertRaises(ci.AcceptanceFailure):
                    ci.plan(argparse.Namespace())

    def test_registry_rejects_unordered_unknown_or_unchecked_partitions(self):
        base = registry({'id': 'reads', 'tests': ['TestAccRead']}, service='platform', package='./tests/integration/lab/platform')
        ci.validate_registry(base)
        crud = {'id': 'crud', 'tests': ['TestAccWrite'], 'mutation': True, 'proof': 'owned-crud', 'journal_entries_per_test': 1}
        for change in [
            lambda r: r['services'].update(teams_unknown_service=r['services'].pop('platform')),
            lambda r: r['services']['platform'].update(package='./tests/integration/lab'),
            lambda r: r['services']['platform'].update(package='./tests/integration/lab/graph'),
            lambda r: r['services']['platform']['steps'].insert(0, crud),
            lambda r: r['services']['platform']['steps'].append(dict(crud, proof='unknown')),
            lambda r: r['services']['platform']['steps'].append(dict(crud, id='reads')),
            lambda r: r['auth'].update(steps=[]),
            lambda r: r['auth']['steps'][0].update(mutation=True),
        ]:
            broken = json.loads(json.dumps(base))
            change(broken)
            with self.subTest(registry=broken), self.assertRaises(ci.AcceptanceFailure):
                ci.validate_registry(broken)

    def test_serialized_reads_never_enable_mutations(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            private = root / 'msteams-acceptance-fixture'
            (private / 'journals').mkdir(parents=True)
            (root / 'suites.json').write_text(json.dumps(registry({'id': 'global-reads', 'serial': True, 'tests': ['TestAccSingletonRead']})))
            invoked = []

            def process(command, **kwargs):
                if command[0] == 'git':
                    return subprocess.CompletedProcess(command, 0, stdout='a' * 40 + '\n')
                invoked.append(kwargs['env']['MSTEAMS_LAB_MUTATIONS'])
                kwargs['stdout'].write(json.dumps({'Test': 'TestAccSingletonRead', 'Action': 'pass'}) + '\n')
                return subprocess.CompletedProcess(command, 0)

            with patch.object(ci, 'HERE', root), patch.dict(os.environ, {'MSTEAMS_LAB_PRIVATE_DIRECTORY': str(private)}, clear=True), patch.object(ci.subprocess, 'run', side_effect=process):
                with contextlib.redirect_stdout(io.StringIO()):
                    self.assertEqual(ci.run(service_run(root)), 0)
            self.assertEqual(invoked, ['0'])
            self.assertEqual(json.loads((root / 'reports/svc--global-reads.json').read_text())['phase'], 'writes')

    def test_prepare_requires_certificate_material_only_for_certificate_auth(self):
        for mode, required in [('all', True), ('certificate', True), ('client-secret', False)]:
            with self.subTest(mode=mode), patch.dict(os.environ, {'MSTEAMS_LAB_AUTH_METHOD': mode}, clear=True):
                with self.assertRaises(ci.AcceptanceFailure) as failure:
                    ci.prepare(argparse.Namespace(metadata=None, archive=None))
                self.assertEqual('MSTEAMS_LAB_CERTIFICATE_PEM' in str(failure.exception), required)
                self.assertEqual('MSTEAMS_LAB_PRIVATE_KEY_PEM' in str(failure.exception), required)

    def test_configuration_update_requires_restoration_and_three_comparisons(self):
        name = 'TestAccMessagingConfigurationUpdateRestore'
        markers, oracle, count = ci.PROOFS['configuration-update']
        events = [{'Test': name, 'Action': 'output', 'Output': marker} for marker in markers + [oracle] * count]
        events.append({'Test': name, 'Action': 'pass'})
        self.assertEqual(ci.evaluate(events, [name], 0, proof='configuration-update')['outcome'], 'passed')
        for omitted in markers:
            incomplete = [event for event in events if event.get('Output') != omitted]
            self.assertEqual(ci.evaluate(incomplete, [name], 0, proof='configuration-update')['outcome'], 'failed')
        events.remove(next(event for event in events if event.get('Output') == oracle))
        self.assertEqual(ci.evaluate(events, [name], 0, proof='configuration-update')['outcome'], 'failed')

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


    def test_populated_template_denial_needs_parity_and_cleanup(self):
        name = 'TestAccMeetingTemplatePermissionPopulatedAccess'
        markers, oracle, count = ci.PROOFS['populated-template-access']
        events = [{'Test': name, 'Action': 'output', 'Output': marker} for marker in markers + [oracle]]
        events.append({'Test': name, 'Action': 'pass'})
        self.assertEqual(ci.evaluate(events, [name], 0, proof='populated-template-access')['outcome'], 'passed')
        self.assertEqual(count, 1)
        for omitted in markers + [oracle]:
            incomplete = [event for event in events if event.get('Output') != omitted]
            self.assertEqual(ci.evaluate(incomplete, [name], 0, proof='populated-template-access')['outcome'], 'failed')

    def test_notification_crud_requires_the_null_read_and_three_comparisons(self):
        name = 'TestAccNotificationAndFeedsPolicyCRUD'
        markers, oracle, count = ci.PROOFS['owned-notification-crud']
        events = [{'Test': name, 'Action': 'output', 'Output': marker} for marker in markers + [oracle] * count]
        events.append({'Test': name, 'Action': 'pass'})
        self.assertEqual(ci.evaluate(events, [name], 0, proof='owned-notification-crud')['outcome'], 'passed')
        self.assertEqual(count, 3)
        incomplete = [event for event in events if event.get('Output') != 'notification explicit-null update/read-back: passed']
        self.assertEqual(ci.evaluate(incomplete, [name], 0, proof='owned-notification-crud')['outcome'], 'failed')
        events.remove(next(event for event in events if event.get('Output') == oracle))
        self.assertEqual(ci.evaluate(events, [name], 0, proof='owned-notification-crud')['outcome'], 'failed')


if __name__ == '__main__':
    unittest.main()
