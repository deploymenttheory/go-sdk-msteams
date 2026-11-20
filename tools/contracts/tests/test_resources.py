import copy
import pathlib
import sys
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parents[1]))
import verify_resources as resources  # noqa: E402


class ResourceEvidenceTests(unittest.TestCase):
    def test_crud_actions_and_parameter_sets_bind_to_actual_request_captures(self):
        value = resources.evidence.read(resources.CONTRACT)
        self.assertEqual(resources.validate(value), (5, 8, 8))
        mutations = [
            lambda v: v.update(status='accepted'),
            lambda v: v['open_questions'].clear(),
            lambda v: v['commands'].pop(),
            lambda v: v['commands'][0]['parameter_sets'].clear(),
            lambda v: v['commands'][0]['parameter_sets'][0].update(operation_refs=['missing']),
            lambda v: v['service_surface'].update(get_by_identity='list'),
            lambda v: v['operations'][3].update(method='PATCH'),
            lambda v: v['operations'][3].update(replay_safe=True),
            lambda v: v['operations'][3].update(path='/Skype.Policy/configurations/Other/configuration/{identity}'),
            lambda v: v['operations'][3]['response'].update(statuses=[200]),
            lambda v: v['operations'][3]['request_cases'][0].update(case='missing'),
            lambda v: v['profile'].update(audience='wrong-resource'),
            lambda v: v['profile']['discovery'].update(cross_origin_send=True),
        ]
        for mutate in mutations:
            changed = copy.deepcopy(value)
            mutate(changed)
            with self.assertRaises(ValueError):
                resources.validate(changed)


if __name__ == '__main__':
    unittest.main()
