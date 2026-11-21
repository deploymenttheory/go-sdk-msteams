import copy
import json
import pathlib
import sys
import tempfile
import unittest

sys.path.insert(0,str(pathlib.Path(__file__).resolve().parents[1]))
from ledger import read, seed, validate


class LedgerTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = pathlib.Path(self.tmp.name)
        self.reconciliation = {'records':[dict(name='Get-Sample',disposition='observed-management-export',
            family='sample',implementation_owner='EXT-20',research_owner='EXT-01')]}
        self.runtime = {'commands':[dict(name='Get-Sample',parameter_sets=[dict(name='Identity',parameters=[])])]}
        self.ledger = seed(self.reconciliation,self.runtime)
        self.tickets = {'EXT-20','EXT-01','BASE-02'}
        self.write('contracts/powershell/8.1.0/reconciliation.json',self.reconciliation)
        self.write('contracts/powershell/8.1.0/runtime.json',self.runtime)
        (self.root/'sample.go').write_text('package sample\nfunc Get() {}\n',encoding='utf-8')
        (self.root/'sample_test.go').write_text('package sample\nfunc TestUnitGet() {}\nfunc TestAccGet() {}\n',encoding='utf-8')
        self.record = self.ledger['records'][0]

    def write(self,path,value):
        p=self.root/path
        p.parent.mkdir(parents=True,exist_ok=True)
        p.write_text(json.dumps(value),encoding='utf-8')

    def check(self,complete=False):
        return validate(self.ledger,self.reconciliation,self.runtime,self.tickets,self.root,complete)

    def implemented(self):
        symbol=dict(file='sample.go',name='Get')
        self.record['implementation'].update(status='implemented',symbols=[symbol])
        self.record['parameter_sets'][0]['symbols']=[symbol]
        self.record['parameter_sets'][0]['contract_tests']=[dict(file='sample_test.go',name='TestUnitGet')]
        self.write('contract-result.json',dict(outcome='passed',revision='fixture-revision'))
        self.record['contract']=dict(status='passed',evidence=['contract-result.json'])

    def test_metadata_is_valid_but_not_release_ready(self):
        result=self.check()
        self.assertEqual(result['result'],'passed')
        self.assertFalse(result['release_ready'])
        self.assertEqual(result['complete'],0)
        self.assertEqual(self.check(complete=True)['result'],'failed')

    def test_unit_and_acceptance_references_cannot_be_interchanged(self):
        self.implemented()
        group = self.record['parameter_sets'][0]
        group['live_tests'] = [dict(file='sample_test.go', name='TestUnitGet')]
        self.assertEqual(self.check()['result'], 'failed')
        group['live_tests'] = [dict(file='sample_test.go', name='TestAccGet')]
        self.assertEqual(self.check()['result'], 'passed')
        group['contract_tests'] = [dict(file='sample_test.go', name='TestAccGet')]
        self.assertEqual(self.check()['result'], 'failed')

    def test_missing_and_duplicate_ownership(self):
        self.record['implementation_owner']=None
        self.assertEqual(self.check()['result'],'failed')
        self.record['implementation_owner']='EXT-20'
        self.ledger['records'].append(copy.deepcopy(self.record))
        self.assertTrue(any('duplicate' in e['message'] for e in self.check()['errors']))

    def test_new_candidate_or_missing_record_fails(self):
        candidate=copy.deepcopy(self.reconciliation['records'][0]);candidate['name']='Get-New'
        self.reconciliation['records'].append(candidate)
        self.assertTrue(any('orphan' in e['message'] for e in self.check()['errors']))

    def test_stale_symbol_or_missing_evidence_fails(self):
        self.implemented()
        self.assertTrue(self.check()['release_ready'])
        (self.root/'sample.go').write_text('package sample\n',encoding='utf-8')
        self.assertEqual(self.check()['result'],'failed')
        self.record['metadata_ref']='missing.json#/records/0'
        self.assertEqual(self.check()['result'],'failed')

    def test_management_exclusion_or_research_reclassification_fails(self):
        self.record.update(kind='excluded',exclusion='not in my tenant')
        self.assertEqual(self.check()['result'],'failed')
        self.record.update(kind='research-candidate',exclusion=None,implementation_owner=None)
        self.assertEqual(self.check()['result'],'failed')

    def test_unsupported_stub_never_counts_as_implemented(self):
        self.implemented()
        self.record['implementation']['unsupported_stub']=True
        self.assertEqual(self.check()['result'],'failed')
        self.record['implementation']['status']='unsupported'
        self.assertEqual(self.check()['result'],'failed')

    def test_skipped_live_run_cannot_be_passed(self):
        self.implemented()
        self.record['parameter_sets'][0]['live_tests']=[dict(file='sample_test.go',name='TestAccGet')]
        self.write('live-result.json',dict(outcome='skipped',revision='fixture-revision'))
        self.record['live']=dict(status='passed',reason='',evidence=['live-result.json'])
        self.assertEqual(self.check()['result'],'failed')
        self.write('live-result.json',dict(outcome='passed',revision='fixture-revision'))
        self.assertEqual(self.check()['result'],'passed')
        self.record['live']=dict(status='unavailable',reason='Tenant lacks test capability.',evidence=[])
        self.assertEqual(self.check()['result'],'passed')
        self.assertEqual(self.check()['live_counts'],{'unavailable':1})

    def test_contract_run_must_pass_and_cover_all_parameter_sets(self):
        self.implemented()
        self.write('contract-result.json',dict(outcome='failed',revision='fixture-revision'))
        self.assertEqual(self.check()['result'],'failed')
        self.write('contract-result.json',dict(outcome='passed',revision='fixture-revision'))
        self.record['parameter_sets'][0]['contract_tests']=[]
        self.assertEqual(self.check()['result'],'failed')

    def test_mismatched_signature_and_escaping_reference_fail(self):
        self.record['parameter_sets'][0]['name']='Wrong'
        self.assertEqual(self.check()['result'],'failed')
        self.record['metadata_ref']='../outside.json'
        self.assertEqual(self.check()['result'],'failed')

    def test_unknown_fields_and_invalid_records_fail(self):
        self.record['live_tested']=True
        self.assertEqual(self.check()['result'],'failed')
        self.ledger['records']=[None]
        self.assertEqual(self.check()['result'],'failed')
