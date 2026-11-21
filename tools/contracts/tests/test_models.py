import copy
import importlib.util
import pathlib
import unittest

PATH = pathlib.Path(__file__).resolve().parents[1] / 'verify_models.py'
SPEC = importlib.util.spec_from_file_location('verify_models', PATH)
models = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(models)


class ModelEvidenceTests(unittest.TestCase):
    def test_special_profiles_keep_origin_audience_scope_and_cloud_prefix_distinct(self):
        value = models.read(models.SPECIAL_PROFILES)
        inventory = models.read(models.CONTRACT)
        self.assertEqual(models.validate_special_profiles(value, inventory), (4, 11, 5))
        mutations = [
            lambda v: v.update(status='accepted'),
            lambda v: v['profiles'].pop(),
            lambda v: v['profiles'][0].update(audience='https://retailservices.gcc.teams.microsoft.com'),
            lambda v: v['profiles'][1]['cloud_routes']['commercial'].update(origin='https://chatsvcagg.teams.microsoft.com'),
            lambda v: v['profiles'][2].update(delegated_scope='https://migrationsvc.teams.microsoft.com/.default'),
            lambda v: v['profiles'][3]['cloud_routes']['gcch'].update(prefix='/fabric/templates/api'),
            lambda v: v['profiles'][3]['audiences_by_cloud'].update(gcch='https://templates.teams.microsoft.com'),
            lambda v: v['profiles'][0].update(isolated_selector_rejects=[]),
            lambda v: v['operations'][0].update(profile='regions'),
            lambda v: v['operations'][0].update(replay_safe=True),
            lambda v: v['commands'][0]['parameter_sets'].clear(),
        ]
        for mutate in mutations:
            changed = copy.deepcopy(value)
            mutate(changed)
            with self.assertRaises(ValueError):
                models.validate_special_profiles(changed, inventory)

    def test_compositions_bind_runtime_sets_and_reject_broken_evidence(self):
        value = models.read(models.COMPOSITIONS)
        inventory = models.read(models.CONTRACT)
        self.assertEqual(models.validate_compositions(value, inventory), (9, 29))
        mutations = [
            lambda v: v.update(captured=True),
            lambda v: v['commands'].pop(),
            lambda v: v['commands'][0]['parameter_sets'].clear(),
            lambda v: v['commands'][0]['case_refs'].append('missing'),
            lambda v: v['cases'].append(copy.deepcopy(v['cases'][0])),
            lambda v: v['cases'][0].update(parameter_sets=['not-a-runtime-set']),
            lambda v: v['cases'][0]['input'].update(NotAParameter=True),
            lambda v: v['cases'][0]['steps'][0].update(operation='missing'),
            lambda v: v['cases'][0]['steps'][0].update(method='DELETE'),
            lambda v: v['cases'][0]['steps'][0].update(path='https://foreign.invalid/v1.0/groups/fixture'),
            lambda v: v['cases'][0]['steps'][0].update(body={'kind': 'json'}),
            lambda v: v['cases'][0]['steps'][0].update(query={'$search': 'loses repeated keys'}),
        ]
        for mutate in mutations:
            changed = copy.deepcopy(value)
            mutate(changed)
            with self.assertRaises(ValueError):
                models.validate_compositions(changed, inventory)

    def test_inventory_retains_complete_assignment_and_unresolved_state(self):
        value = models.read(models.CONTRACT)
        self.assertEqual(models.validate_inventory(value), (40, 61))
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
