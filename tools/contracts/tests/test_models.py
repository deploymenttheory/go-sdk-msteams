import copy
import importlib.util
import pathlib
import unittest

PATH = pathlib.Path(__file__).resolve().parents[1] / 'verify_models.py'
SPEC = importlib.util.spec_from_file_location('verify_models', PATH)
models = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(models)


class ModelEvidenceTests(unittest.TestCase):
    def test_inventory_retains_complete_assignment_and_unresolved_state(self):
        value = models.read(models.CONTRACT)
        self.assertEqual(models.validate_inventory(value), (40, 60))
        mutations = [
            lambda v: v['commands'].pop(),
            lambda v: v['commands'][0]['parameter_sets'].clear(),
            lambda v: v['commands'][0]['operation_refs'].append('missing'),
            lambda v: v.update(status='accepted'),
            lambda v: v['commands'][0].update(open_questions=[]),
            lambda v: v['operations'][0].update(path='https://foreign.example/'),
        ]
        for mutate in mutations:
            changed = copy.deepcopy(value)
            mutate(changed)
            with self.assertRaises(ValueError):
                models.validate_inventory(changed)

    def test_comparison_preserves_meaningful_json_distinctions(self):
        expected = {'captured_at': 'one', 'wire': {'enabled': False, 'roles': [], 'name': ''}}
        same = copy.deepcopy(expected)
        same['captured_at'] = 'two'
        self.assertEqual(models.compare(expected, same), [])
        for change in [None, 0, 'false', True]:
            actual = copy.deepcopy(expected)
            actual['wire']['enabled'] = change
            self.assertEqual(models.compare(expected, actual), ['/wire/enabled'])
        actual = copy.deepcopy(expected)
        del actual['wire']['roles']
        self.assertEqual(models.compare(expected, actual), ['/wire/roles'])
        self.assertEqual(models.differences(['a', 'b'], ['b', 'a']), ['/0', '/1'])

    def test_capture_failure_reports_structure_without_values(self):
        self.assertEqual(models.differences({'value': 'fixture-private'}, {'value': 'other-private'}), ['/value'])


if __name__ == '__main__':
    unittest.main()
