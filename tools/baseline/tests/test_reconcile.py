import copy
import hashlib
import pathlib
import sys
import tempfile
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parents[1]))
from baseline import DiscoveryError, PIN
from reconcile import reconcile


class ReconciliationTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.module = pathlib.Path(self.tmp.name)
        raw = b"'Get-Public'\n#MPA OCE cmdlets\n,'Get-Conditional'\n#preview ,'Get-Preview'\n"
        (self.module/'MicrosoftTeams.psd1').write_bytes(raw)
        self.snapshot = dict(module_version='8.1.0', package_sha512_base64=PIN,
            package_sha256='test', runtime_environment={'authenticated': False},
            gallery_commands=['Get-Public', 'Get-Conditional', 'Get-Gated', 'Get-Unknown'],
            runtime_commands=[dict(name='Get-PUBLIC', command_type='Cmdlet', parameter_sets=[])],
            sources=[dict(path='MicrosoftTeams.psd1', sha256=hashlib.sha256(raw).hexdigest(),
                          lexical_command_candidates=['Get-Public','Get-Conditional','Get-Gated','Get-Unknown']),
                     dict(path='nested.psd1', sha256='test', lexical_command_candidates=['Get-Preview']),
                     dict(path='CmdletConfig.json', sha256='test', configuration={'CmdletConfiguration': {
                         'Get-Gated': {'ExportsTo': ['SelfHost']}}})])
        self.owners = [dict(command='Get-Public', disposition='management-candidate', family='sample',
                            research='BASE-02', implementation='EXT-01')]

    def test_union_retains_conditional_unknown_preview_and_case_spelling(self):
        result = reconcile(self.snapshot, self.module, self.owners)
        records = {r['name']: r for r in result['records']}
        self.assertEqual(len(records), 5)
        self.assertEqual(records['Get-Public']['runtime_name'], 'Get-PUBLIC')
        self.assertEqual(records['Get-Conditional']['availability'], 'conditional-oce-declaration')
        self.assertEqual(records['Get-Gated']['availability'], 'conditional-config-export')
        self.assertEqual(result['unresolved_exports'], ['Get-Unknown'])
        self.assertEqual(result['nested_candidates'], ['Get-Preview'])
        self.assertEqual(records['Get-Preview']['evidence']['commented_root_lines'], [4])
        self.assertEqual(records['Get-Preview']['research_owner'], 'BASE-02')
        self.assertTrue(all(r['qualification']=='metadata-only' for r in records.values()))

    def test_extra_runtime_export_is_retained_for_research(self):
        self.snapshot['runtime_commands'].append(dict(name='Get-NewExport', command_type='Alias', parameter_sets=[]))
        result = reconcile(self.snapshot, self.module, self.owners)
        extra = next(r for r in result['records'] if r['name']=='Get-NewExport')
        self.assertEqual(extra['research_owner'], 'BASE-02')
        self.assertEqual(extra['disposition'], 'research-candidate')
        self.assertEqual(extra['command_type'], 'Alias')

    def test_duplicate_owner_and_runtime_fail(self):
        with self.assertRaises(DiscoveryError):
            reconcile(self.snapshot, self.module, self.owners * 2)
        self.snapshot['runtime_commands'] *= 2
        with self.assertRaises(DiscoveryError):
            reconcile(self.snapshot, self.module, self.owners)

    def test_mismatched_or_incomplete_evidence_fails(self):
        for key, value in [('module_version','8.2.0'), ('package_sha512_base64','wrong'), ('runtime_commands',[])]:
            candidate = copy.deepcopy(self.snapshot)
            candidate[key] = value
            with self.assertRaises(DiscoveryError):
                reconcile(candidate, self.module, self.owners)
        (self.module/'MicrosoftTeams.psd1').write_text('changed', encoding='utf-8')
        with self.assertRaises(DiscoveryError):
            reconcile(self.snapshot, self.module, self.owners)
