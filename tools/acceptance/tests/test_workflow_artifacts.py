"""Keep the shared upload step aligned with bounded sanitized fixture exports."""
from pathlib import Path
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[3]


def upload_patterns(workflow):
    lines = workflow.splitlines()
    start = next(index for index, line in enumerate(lines) if line.strip() == '- name: Upload test results')
    step_indent = len(lines[start]) - len(lines[start].lstrip())
    block = []
    for line in lines[start + 1:]:
        if line.strip() and len(line) - len(line.lstrip()) <= step_indent:
            break
        block.append(line)
    start = next(index for index, line in enumerate(block) if line.strip() == 'path: |')
    path_indent = len(block[start]) - len(block[start].lstrip())
    paths = []
    for line in block[start + 1:]:
        if line.strip() and len(line) - len(line.lstrip()) <= path_indent:
            break
        if line.strip():
            paths.append(line.strip())
    return paths


class WorkflowArtifactGuardTests(unittest.TestCase):
    def test_shared_step_uploads_exact_bounded_report_and_fixture_paths(self):
        workflow = (ROOT / '.github/workflows/tenant-acceptance.yml').read_text()
        prefix = '${{ runner.temp }}/acceptance-report/'
        expected = {
            '*.json', 'recovery.cms',
            'response-fixtures/schedules/integration_holidays.json',
            'response-fixtures/schedules/integration_holidays.body',
            'response-fixtures/configuration_diagnostics/integration_*.json',
            'response-fixtures/configuration_diagnostics/integration_create_201.body',
            'response-fixtures/teams_feedback_policy/integration_*.json',
            'response-fixtures/call_templates/ext_integration_*.json',
            'response-fixtures/ai_agents/ext_integration_*.json',
        }
        self.assertEqual(set(upload_patterns(workflow)), {prefix + path for path in expected})
        self.assertEqual(workflow.count('steps: *partition-steps'), 2)
        self.assertIn('steps: &partition-steps', workflow)

    def test_patterns_retain_nested_capture_bytes_without_private_trees(self):
        workflow = (ROOT / '.github/workflows/tenant-acceptance.yml').read_text()
        wanted = [
            'acceptance-report/summary.json', 'acceptance-report/recovery.cms',
            'acceptance-report/response-fixtures/schedules/integration_holidays.json',
            'acceptance-report/response-fixtures/schedules/integration_holidays.body',
            'acceptance-report/response-fixtures/configuration_diagnostics/integration_create_201.body',
            'acceptance-report/response-fixtures/configuration_diagnostics/integration_generic_get.json',
            'acceptance-report/response-fixtures/teams_feedback_policy/integration_typed_get.json',
            'acceptance-report/response-fixtures/call_templates/ext_integration_create_201.json',
            'acceptance-report/response-fixtures/ai_agents/ext_integration_read.json',
        ]
        private = [
            'msteams-acceptance-private/identity.json', 'msteams-acceptance-private/journals/owned.json',
            'msteams-acceptance-private/results.jsonl', 'acceptance-report/raw.log',
            'acceptance-report/response-fixtures/private/journals/owned.json',
            'acceptance-report/response-fixtures/schedules/recovery.json',
            'acceptance-report/response-fixtures/ai_agents/raw.body',
        ]
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            for name in wanted + private:
                path = root / name
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_bytes(b'')
            selected = set()
            for pattern in upload_patterns(workflow):
                relative = pattern.removeprefix('${{ runner.temp }}/')
                selected.update(path.relative_to(root).as_posix() for path in root.glob(relative))
            self.assertEqual(selected, set(wanted))
