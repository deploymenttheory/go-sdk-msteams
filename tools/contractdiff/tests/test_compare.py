import copy
import base64
import gzip
import hashlib
import json
import pathlib
import sys
import tempfile
import subprocess
import unittest

sys.path.insert(0,str(pathlib.Path(__file__).resolve().parents[1]))
from compare import ROOT, DiscoveryError, compare

FIXTURES=pathlib.Path(__file__).parent/'fixtures'


def snapshot():
    param=dict(name='Identity',type='System.String',mandatory=False,position=-2147483648,
               aliases=[],pipeline_input=False,pipeline_by_property=False,validation=[])
    return dict(schema_version=1,module_version='8.1.0',package_sha256='a'*64,
                package_sha512_base64=base64.b64encode(b'x'*64).decode(),package_url='https://example.test/package',
                runtime_environment=dict(module_version='8.1.0',powershell_version='7.6.6',platform='test',authenticated=False),
                runtime_commands=[dict(name='Get-Sample',command_type='Cmdlet',output_types=['System.String'],
                    parameter_sets=[dict(name='Identity',is_default=True,parameters=[param])])],
                gallery_commands=['Get-Sample'],release_notes='',limitations=['Synthetic fixture'],
                sources=[dict(path='module.psd1',sha256='b'*64,lexical_command_candidates=['Get-Sample']),
                         dict(path='help.xml',sha256='c'*64,help_commands=['Get-Sample']),
                         dict(path='CmdletConfig.json',sha256='d'*64,configuration={'CmdletConfiguration':{}})])


class CompareTests(unittest.TestCase):
    def setUp(self):
        self.old=snapshot()
        self.new=copy.deepcopy(self.old)
        self.coverage={'records':[dict(name='Get-Sample',research_owner='CORE-01',implementation_owner='CORE-20',
            resource='sample',implementation={'status':'not-started','symbols':[]},parameter_sets=[],
            contract={'status':'metadata-only'},live={'status':'not-run'})]}

    def compare(self,**kwargs):
        return compare(self.old,self.new,self.coverage,**kwargs)

    def test_identical_case_and_order_normalization(self):
        self.new['runtime_commands'][0]['name']='GET-SAMPLE'
        self.new['gallery_commands'][0]='GET-SAMPLE'
        self.new['sources'].reverse()
        self.assertEqual(self.compare()['status'],'no-delta')

    def test_command_add_remove_and_rename_are_explicit(self):
        self.new['runtime_commands'][0]['name']='Get-Renamed'
        changes=self.compare()['changes']
        removed=next(c for c in changes if c['path']==['commands','get-sample'])
        added=next(c for c in changes if c['path']==['commands','get-renamed'])
        self.assertTrue(removed['old']['present']);self.assertFalse(removed['new']['present'])
        self.assertEqual(removed['impact']['implementation_owner'],'CORE-20')
        self.assertTrue(added['impact']['assignment_required'])
        self.assertEqual(added['impact']['research_owner'],'DELTA-01')

    def test_parameter_type_requiredness_validation_default_and_null(self):
        old=self.old['runtime_commands'][0]['parameter_sets'][0]['parameters'][0]
        new=self.new['runtime_commands'][0]['parameter_sets'][0]['parameters'][0]
        old['default']=False
        new.update(type='System.Int32',mandatory=True,default=0,validation=[{'type':'ValidateRange','properties':{'MinRange':0,'MaxRange':9}}])
        paths=[c['path'][-1] for c in self.compare()['changes']]
        self.assertCountEqual(paths,['type','mandatory','default','validation'])
        del old['default'];new['default']=None
        change=next(c for c in self.compare()['changes'] if c['path'][-1]=='default')
        self.assertFalse(change['old']['present'])
        self.assertEqual(change['new'],{'present':True,'value':None})

    def test_parameter_sets_output_types_and_alias_targets(self):
        self.new['runtime_commands'][0]['parameter_sets'][0]['is_default']=False
        self.new['runtime_commands'][0]['output_types']=['System.Int32']
        self.assertEqual(len(self.compare()['changes']),2)
        alias=copy.deepcopy(self.new['runtime_commands'][0])
        alias.update(name='Get-Alias',command_type='Alias',alias_target='Get-Sample')
        self.new['runtime_commands'].append(alias)
        self.assertTrue(any(c['path']==['commands','get-alias'] for c in self.compare()['changes']))
        del alias['alias_target']
        with self.assertRaises(DiscoveryError):self.compare()

    def test_incomplete_and_unknown_inputs_cannot_report_clean(self):
        for field in ['sources','runtime_commands','gallery_commands']:
            candidate=copy.deepcopy(self.new);candidate[field]=[]
            with self.assertRaises(DiscoveryError):compare(self.old,candidate,self.coverage)
        self.old['future_field']=self.new['future_field']={'added':True}
        result=self.compare()
        self.assertEqual(result['status'],'needs-review')
        self.assertTrue(result['warnings'])
        self.assertEqual(result['changes'],[])

    def test_documentation_opaque_sources_and_observed_auth_changes(self):
        self.new['release_notes']='updated documentation'
        result=self.compare()
        self.assertEqual(result['counts']['documentation'],1)
        self.old['sources'].append(dict(path='loader.psm1',sha256='1'*64))
        self.new['sources'].append(dict(path='loader.psm1',sha256='2'*64))
        self.assertEqual(self.compare()['counts']['unclassified'],1)
        before=[dict(command='Get-Sample',classification='observed-behaviour',evidence='before.json',facts={'authentication':{'mode':'delegated'},'routing':{'path':'/old'}})]
        after=copy.deepcopy(before);after[0].update(evidence='after.json',facts={'authentication':{'mode':'application'},'routing':{'path':'/new'}})
        result=self.compare(contracts_before=before,contracts_after=after)
        self.assertEqual(result['counts']['observed-behaviour'],2)
        self.assertTrue(all(c['evidence']=={'before':'before.json','after':'after.json'} for c in result['changes'] if c['classification']=='observed-behaviour'))

    def test_changed_package_or_missing_provenance_cannot_be_clean(self):
        self.new['package_sha256']='f'*64
        self.assertEqual(self.compare()['status'],'needs-review')
        self.new['package_sha512_base64']=''
        with self.assertRaises(DiscoveryError):self.compare()

    def test_stable_and_preview_channels_are_not_promoted(self):
        self.new['module_version']='8.2.0-preview'
        self.new['runtime_environment']['module_version']='8.2.0-preview'
        result=self.compare()
        self.assertEqual(result['before']['channel'],'stable')
        self.assertEqual(result['after']['channel'],'preview')
        self.assertEqual(result['status'],'no-delta')
        self.assertEqual(self.old['module_version'],'8.1.0')

    def test_cli_exit_statuses_and_immutable_output(self):
        with tempfile.TemporaryDirectory() as directory:
            work=pathlib.Path(directory)
            before=work/'before.json';after=work/'after.json';output=work/'report.json'
            before.write_text(json.dumps(self.old),encoding='utf-8')
            after.write_text(json.dumps(self.new),encoding='utf-8')
            command=[sys.executable,str(ROOT/'tools/contractdiff/compare.py'),'--before',str(before),'--after',str(after),'--output',str(output)]
            result=subprocess.run(command,capture_output=True,text=True,check=False)
            self.assertEqual(result.returncode,0,result.stdout+result.stderr)
            original=output.read_bytes()
            self.new['runtime_commands'][0]['output_types']=['NewType']
            after.write_text(json.dumps(self.new),encoding='utf-8')
            result=subprocess.run(command,capture_output=True,text=True,check=False)
            self.assertEqual(result.returncode,2)
            self.assertEqual(output.read_bytes(),original)
            command[-1]=str(work/'changed.json')
            self.assertEqual(subprocess.run(command,capture_output=True,check=False).returncode,1)
            after.write_text('{}',encoding='utf-8')
            command[-1]=str(work/'incomplete.json')
            self.assertEqual(subprocess.run(command,capture_output=True,check=False).returncode,2)
            self.assertFalse((work/'incomplete.json').exists())

    def test_verified_historical_800_to_810_and_release_note_claim(self):
        provenance=json.loads((FIXTURES/'provenance.json').read_text(encoding='utf-8'))
        snapshots=[]
        for version in ['8.0.0','8.1.0']:
            raw=gzip.decompress((FIXTURES/(version+'.json.gz')).read_bytes())
            self.assertEqual(hashlib.sha256(raw).hexdigest(),provenance['snapshots'][version]['fixture_uncompressed_sha256'])
            snapshots.append(json.loads(raw))
        coverage=json.loads((ROOT/'contracts/coverage/8.1.0.json').read_text(encoding='utf-8'))
        result=compare(*snapshots,coverage)
        paths={tuple(c['path']) for c in result['changes']}
        self.assertIn(('commands','new-csphonenumberusagechangeorder'),paths)
        self.assertTrue(any('invitewithreplaces' in p for p in paths))
        self.assertTrue(any('filetypeblocklist' in p for p in paths))
        claim=dict(command='Get-CsCallQueue',classification='documentation',
            evidence='tools/contractdiff/tests/fixtures/provenance.json#/documented_changes/0',facts={'output_fields':{'Statistics':True}})
        changed=copy.deepcopy(claim);changed['facts']['output_fields']['Statistics']=False
        report=compare(*snapshots,coverage,contracts_before=[claim],contracts_after=[changed])
        removal=next(c for c in report['changes'] if c['path']==['contracts','get-cscallqueue','output_fields','Statistics'])
        self.assertEqual(removal['classification'],'documentation')
        self.assertEqual(removal['impact']['resource'],'call_queues')
        self.assertEqual(report['counts']['observed-behaviour'],0)
