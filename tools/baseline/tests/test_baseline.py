import base64
import copy
import hashlib
import io
import json
import pathlib
import subprocess
import sys
import tempfile
import unittest
import zipfile
from unittest import mock

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parents[1]))
import baseline


class BaselineTests(unittest.TestCase):
    def setUp(self):
        self.work = tempfile.TemporaryDirectory()
        self.addCleanup(self.work.cleanup)
        self.root = pathlib.Path(self.work.name)
        self.archive = self.root/'module.zip'
        self.runtime = dict(module_version='8.1.0', authenticated=False, powershell_version='7.6.6', platform='test', commands=[dict(
            name='Get-Team',command_type='Cmdlet',output_types=['Team'],parameter_sets=[dict(
                name='Default',is_default=True,parameters=[dict(name='GroupId',type='System.String',mandatory=False,
                    position=0,aliases=[],validation=[],pipeline_input=False,pipeline_by_property=False)])])])
        with zipfile.ZipFile(self.archive,'w') as archive:
            archive.writestr('MicrosoftTeams.psd1', "@{ ModuleVersion = '8.1.0'; CmdletsToExport = @('Get-Team') }")
            archive.writestr('en-US/MicrosoftTeams-help.xml','<helpItems><command><details><name>Get-Team</name></details></command></helpItems>')
            archive.writestr('custom/CmdletConfig.json','{"commands":{"Get-Team":{}}}')
        self.digest = base64.b64encode(hashlib.sha512(self.archive.read_bytes()).digest()).decode()
        self.xml = f'''<entry xmlns:d="http://schemas.microsoft.com/ado/2007/08/dataservices">
        <d:Id>MicrosoftTeams</d:Id><d:Version>8.1.0</d:Version><d:PackageHashAlgorithm>SHA512</d:PackageHashAlgorithm>
        <d:PackageHash>{self.digest}</d:PackageHash><d:Tags>PSCommand_Get-Team</d:Tags></entry>'''.encode()
        self.metadata = baseline.gallery_metadata(self.xml,'8.1.0')

    def extract(self):
        target=self.root/'module'
        baseline.extract_archive(self.archive,target)
        return target

    def test_verified_extraction_contains_all_source_kinds(self):
        baseline.verify_archive(self.archive,self.metadata,self.digest)
        result=baseline.snapshot(self.extract(),self.archive,self.metadata,self.runtime)
        self.assertEqual(result['gallery_commands'],['Get-Team'])
        self.assertEqual(result['runtime_commands'][0]['name'],'Get-Team')
        self.assertTrue(any('help_commands' in s for s in result['sources']))
        self.assertTrue(any('configuration' in s for s in result['sources']))
        self.assertTrue(any('lexical_command_candidates' in s for s in result['sources']))

    def test_checksum_must_match_metadata_and_pin(self):
        with self.assertRaisesRegex(baseline.DiscoveryError,'SHA512'):
            baseline.verify_archive(self.archive,self.metadata,'wrong pin')
        self.archive.write_bytes(self.archive.read_bytes()[:-10])
        with self.assertRaisesRegex(baseline.DiscoveryError,'SHA512'):
            baseline.verify_archive(self.archive,self.metadata,self.digest)

    def test_metadata_rejects_identity_hash_and_empty_inventory(self):
        for before,after in [(b'8.1.0',b'8.2.0'),(b'SHA512',b'SHA256'),(b'PSCommand_Get-Team',b'')]:
            with self.subTest(before=before),self.assertRaises(baseline.DiscoveryError):
                baseline.gallery_metadata(self.xml.replace(before,after),'8.1.0')

    def test_truncated_and_traversal_archives_fail(self):
        self.archive.write_bytes(b'PK broken')
        with self.assertRaises(baseline.DiscoveryError): self.extract()
        with zipfile.ZipFile(self.archive,'w') as archive: archive.writestr('../outside','bad')
        with self.assertRaisesRegex(baseline.DiscoveryError,'Unsafe'): self.extract()
        self.assertFalse((self.root/'outside').exists())

    def test_case_colliding_archive_paths_fail(self):
        with zipfile.ZipFile(self.archive,'w') as archive:
            archive.writestr('a.txt','one'); archive.writestr('A.txt','two')
        with self.assertRaisesRegex(baseline.DiscoveryError,'Duplicate'): self.extract()

    def test_incomplete_runtime_never_reports_empty_success(self):
        cases=[dict(self.runtime,commands=[]),dict(self.runtime,module_version='8.2.0'),dict(self.runtime,authenticated=True)]
        incomplete=copy.deepcopy(self.runtime);incomplete['commands'][0]['parameter_sets']=[];cases.append(incomplete)
        duplicate=copy.deepcopy(self.runtime);duplicate['commands']*=2;cases.append(duplicate)
        attributes=copy.deepcopy(self.runtime);del attributes['commands'][0]['parameter_sets'][0]['parameters'][0]['validation'];cases.append(attributes)
        output_types=copy.deepcopy(self.runtime);del output_types['commands'][0]['output_types'];cases.append(output_types)
        for value in cases:
            with self.subTest(value=value),self.assertRaises(baseline.DiscoveryError):
                baseline.canonical_runtime(value,'8.1.0')

    def test_incomplete_package_sources_fail(self):
        module=self.extract();(module/'custom/CmdletConfig.json').unlink()
        with self.assertRaisesRegex(baseline.DiscoveryError,'incomplete'):
            baseline.snapshot(module,self.archive,self.metadata,self.runtime)

    def test_import_failure_and_timeout_are_explicit(self):
        with mock.patch.object(baseline.subprocess,'run',return_value=subprocess.CompletedProcess([],1,'','secret')):
            with self.assertRaisesRegex(baseline.DiscoveryError,'import failed') as error:
                baseline.run_runtime(self.root,self.root/'runtime.json','8.1.0')
            self.assertNotIn('secret',str(error.exception))
        with mock.patch.object(baseline.subprocess,'run',side_effect=subprocess.TimeoutExpired('pwsh',300)):
            with self.assertRaisesRegex(baseline.DiscoveryError,'timed out'):
                baseline.run_runtime(self.root,self.root/'runtime.json','8.1.0')

    def test_truncated_download_and_size_limit(self):
        for length,limit in [('10',100),('3',2)]:
            response=io.BytesIO(b'abc');response.headers={'Content-Length':length}
            with mock.patch.object(baseline.urllib.request,'urlopen',return_value=response):
                with self.assertRaises(baseline.DiscoveryError):
                    baseline.fetch('https://example.invalid',self.root/'download',limit)

    def test_publish_is_repeatable_and_cannot_replace_snapshot(self):
        output=self.root/'snapshot.json'
        self.assertEqual(baseline.publish_snapshot(output,{'a':1}),'created')
        self.assertEqual(baseline.publish_snapshot(output,{'a':1}),'unchanged')
        before=output.read_bytes()
        with self.assertRaisesRegex(baseline.DiscoveryError,'different content'):
            baseline.publish_snapshot(output,{'a':2})
        self.assertEqual(output.read_bytes(),before)

    def test_offline_cli_is_deterministic(self):
        metadata_path=self.root/'gallery.xml';metadata_path.write_bytes(self.xml)
        runtime_path=self.root/'runtime.json';runtime_path.write_text(json.dumps(self.runtime))
        outputs=[self.root/'first.json',self.root/'second.json']
        for output in outputs:
            result=subprocess.run([sys.executable,str(pathlib.Path(baseline.__file__)),'--archive',str(self.archive),
                                   '--metadata',str(metadata_path),'--runtime',str(runtime_path),'--expected-sha512',self.digest,
                                   '--output',str(output)],capture_output=True,text=True)
            self.assertEqual(result.returncode,0,result.stderr)
        self.assertEqual(outputs[0].read_bytes(),outputs[1].read_bytes())


if __name__=='__main__': unittest.main()
