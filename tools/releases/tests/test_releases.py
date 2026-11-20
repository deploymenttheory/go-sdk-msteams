import base64
import copy
import json
import pathlib
import shutil
import sys
import tempfile
import unittest
import urllib.parse
import urllib.error
from unittest.mock import patch

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parents[1]))
import discover
import publish

HASH = base64.b64encode(b'n' * 64).decode()


def release(version='8.2.0', digest=HASH):
    return {'version': version, 'channel': 'preview' if '-' in version else 'stable', 'sha512': digest}


def feed(entries, next_url=None):
    from xml.sax.saxutils import escape, quoteattr
    rows = []
    for item in entries:
        fields = dict(Id='MicrosoftTeams', Version=item['version'], PackageHash=item['sha512'],
                      IsPrerelease='true' if item['channel'] == 'preview' else 'false', PackageHashAlgorithm='SHA512')
        rows.append('<entry><properties>' + ''.join(f'<d:{k}>{escape(v)}</d:{k}>' for k, v in fields.items()) + '</properties></entry>')
    link = '<link rel="next" href=' + quoteattr(next_url) + '/>' if next_url else ''
    return ('<feed xmlns="http://www.w3.org/2005/Atom" xmlns:d="http://schemas.microsoft.com/ado/2007/08/dataservices">'
            + ''.join(rows) + link + '</feed>').encode()


def candidate_snapshot(version='8.2.0'):
    value = discover.accepted_snapshot()
    value['module_version'] = version
    value['runtime_environment']['module_version'] = version
    value['package_sha512_base64'] = HASH
    value['package_sha256'] = 'f' * 64
    value['package_url'] = discover.baseline.GALLERY + '/api/v2/package/MicrosoftTeams/' + version
    return value


class GalleryTests(unittest.TestCase):
    def test_pagination_and_stable_preview_selection(self):
        baseline = release('8.1.0', discover.baseline.PIN)
        next_url = discover.FEED + '&$skip=100'
        pages = {discover.FEED: feed([baseline, release('8.2.0-preview.2')], next_url),
                 next_url: feed([release('8.2.0'), release('8.3.0-preview.2'), release('8.3.0-preview.10')])}
        def fetch(url, path, limit):
            path.write_bytes(pages[url])
        records = discover.gallery_releases(fetch)
        selected = discover.select_candidates(records, discover.accepted_snapshot())
        self.assertEqual([r['version'] for r in selected], ['8.2.0', '8.3.0-preview.10'])

    def test_checksum_republication_has_distinct_identity(self):
        selected = discover.select_candidates([release('8.1.0'), release()], discover.accepted_snapshot())
        self.assertEqual([r['version'] for r in selected], ['8.1.0', '8.2.0'])
        self.assertNotEqual(discover.candidate_key('8.1.0', HASH), discover.candidate_key('8.1.0', discover.baseline.PIN))

    def test_invalid_pages_and_missing_baseline_fail(self):
        for body in (b'not xml', b'<feed/>', feed([release()], 'https://other.test/feed'), feed([release()], discover.FEED)):
            with self.subTest(body=body[:20]):
                def fetch(url, path, limit):
                    path.write_bytes(body)
                with self.assertRaises(discover.Error):
                    discover.gallery_releases(fetch)
        with self.assertRaises(discover.Error):
            discover.select_candidates([release()], discover.accepted_snapshot())
        with self.assertRaises(discover.Error):
            discover.parse_feed(feed([release()]).replace(b'<d:IsPrerelease>false', b'<d:IsPrerelease>true'))
        with self.assertRaises(discover.Error):
            discover.parse_feed(feed([release(digest='invalid')]))

    def test_no_change_does_not_acquire_or_write(self):
        with tempfile.TemporaryDirectory() as tmp:
            out = pathlib.Path(tmp) / 'absent'
            def unexpected(_):
                self.fail('no-change discovery acquired a package')
            self.assertEqual(discover.discover(out, [release('8.1.0', discover.baseline.PIN)], unexpected), 0)
            self.assertFalse(out.exists())
            out.mkdir()
            with self.assertRaises(discover.Error):
                discover.discover(out, [release('8.1.0', discover.baseline.PIN)], unexpected)

    def test_acquisition_failure_preserves_outputs_and_baseline(self):
        before = discover.ACCEPTED.read_bytes()
        with tempfile.TemporaryDirectory() as tmp:
            out = pathlib.Path(tmp) / 'candidates'
            def failed(_):
                raise discover.Error('failed import')
            with self.assertRaises(discover.Error):
                discover.discover(out, [release('8.1.0', discover.baseline.PIN), release()], failed)
            self.assertFalse(out.exists())
        self.assertEqual(discover.ACCEPTED.read_bytes(), before)

    def test_hash_mismatch_never_imports(self):
        def fetch(url, path, limit=0):
            if '/Packages(' in url:
                body = feed([release()]).replace(b'</properties>', b'<d:Tags>PSCommand_Get-Sample</d:Tags></properties>')
            else:
                body = b'not the expected package'
            path.write_bytes(body)
        def runtime(*_):
            self.fail('unverified module was imported')
        with self.assertRaises(discover.Error):
            discover.acquire(release(), fetch, runtime)

    def test_versions_reject_path_or_shell_input(self):
        for value in ('../8.2.0', '8.2.0$(id)', '8.2.0-preview/branch', '8.2.0-preview..one', '08.2.0', '8.2.0\n'):
            with self.subTest(value=value), self.assertRaises(discover.Error):
                discover.candidate_key(value, HASH)


class FakeGitHub:
    repository = publish.REPOSITORY

    def __init__(self):
        self.refs = {'main': 'initial'}
        self.commits = {'initial': {'message': 'base', 'tree': {'sha': 'tree0'}}}
        self.trees = {'tree0': {'README.md': {'path': 'README.md', 'type': 'blob', 'mode': '100644', 'sha': 'readme'}}}
        self.blobs, self.pulls, self.writes = {}, [], []
        self.fail_create = False

    def request(self, method, path, data=None, missing=False):
        if method != 'GET':
            self.writes.append((method, path, copy.deepcopy(data)))
        if method == 'GET' and path.startswith('pulls?'):
            head = urllib.parse.parse_qs(urllib.parse.urlsplit(path).query)['head'][0].split(':', 1)[1]
            return [copy.deepcopy(p) for p in self.pulls if p['head']['ref'] == head]
        if method == 'GET' and path.startswith('git/ref/heads/'):
            ref = path.removeprefix('git/ref/heads/')
            return {'object': {'sha': self.refs[ref]}} if ref in self.refs else None
        if method == 'GET' and path.startswith('git/commits/'):
            return copy.deepcopy(self.commits[path.removeprefix('git/commits/')])
        if method == 'GET' and path.startswith('git/trees/'):
            tree = path.removeprefix('git/trees/').split('?')[0]
            return {'truncated': False, 'tree': list(copy.deepcopy(self.trees[tree]).values())}
        if method == 'POST' and path == 'git/blobs':
            raw = base64.b64decode(data['content']); sha = publish.blob_sha(raw)
            self.blobs[sha] = raw
            return {'sha': sha}
        if method == 'POST' and path == 'git/trees':
            sha = 'tree' + str(len(self.trees))
            tree = copy.deepcopy(self.trees[data['base_tree']])
            tree.update({x['path']: copy.deepcopy(x) for x in data['tree']})
            self.trees[sha] = tree
            return {'sha': sha}
        if method == 'POST' and path == 'git/commits':
            sha = 'commit' + str(len(self.commits))
            self.commits[sha] = dict(message=data['message'], tree={'sha': data['tree']}, parents=data['parents'])
            return {'sha': sha}
        if method == 'POST' and path == 'git/refs':
            ref = data['ref'].removeprefix('refs/heads/')
            if ref in self.refs:
                raise discover.Error('conflicting ref')
            self.refs[ref] = data['sha']
            return {}
        if method == 'PATCH' and path.startswith('git/refs/heads/'):
            ref = path.removeprefix('git/refs/heads/')
            if data['force'] or self.refs[ref] not in self.commits[data['sha']]['parents']:
                raise discover.Error('non-fast-forward update')
            self.refs[ref] = data['sha']
            return {}
        if method == 'POST' and path == 'pulls':
            if self.fail_create:
                self.fail_create = False
                raise discover.Error('transient PR failure')
            self.pulls.append(dict(number=len(self.pulls)+1, state='open', title=data['title'], body=data['body'], draft=data['draft'],
                                   head={'ref': data['head'], 'repo': {'full_name': self.repository}}))
            return copy.deepcopy(self.pulls[-1])
        if method == 'PATCH' and path.startswith('pulls/'):
            self.pulls[int(path.split('/')[1])-1].update(data)
            return {}
        raise AssertionError((method, path))


class PublicationTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.snapshot = candidate_snapshot()
        cls.key, cls.files = discover.bundle(release(), cls.snapshot, discover.accepted_snapshot())

    def test_create_retry_and_no_change_have_one_draft_and_no_extra_writes(self):
        api = FakeGitHub()
        api.fail_create = True
        with self.assertRaises(discover.Error):
            publish.publish_candidate(api, self.key, self.files)
        self.assertEqual(publish.publish_candidate(api, self.key, self.files), 'created')
        writes = len(api.writes)
        self.assertEqual(publish.publish_candidate(api, self.key, self.files), 'unchanged')
        self.assertEqual(len(api.writes), writes)
        self.assertEqual(len(api.pulls), 1)
        self.assertTrue(api.pulls[0]['draft'])
        self.assertEqual(api.refs['main'], 'initial')
        self.assertTrue(all(e['path'].startswith('contracts/candidates/') for method, path, body in api.writes if path == 'git/trees' for e in body['tree']))

    def test_changed_evidence_updates_same_draft(self):
        api = FakeGitHub()
        publish.publish_candidate(api, self.key, self.files)
        changed = copy.deepcopy(self.snapshot)
        changed['runtime_environment']['powershell_version'] = '7.99.0'
        key, files = discover.bundle(release(), changed, discover.accepted_snapshot())
        self.assertEqual(key, self.key)
        self.assertEqual(publish.publish_candidate(api, key, files), 'updated')
        self.assertEqual(len(api.pulls), 1)
        self.assertTrue(api.pulls[0]['draft'])
        self.assertEqual(publish.publish_candidate(api, key, files), 'unchanged')

    def test_candidate_already_recorded_on_main_is_preserved(self):
        api = FakeGitHub()
        publish.publish_candidate(api, self.key, self.files)
        api.refs['main'] = api.refs.pop(publish.PREFIX + self.key)
        api.pulls.clear()
        writes = len(api.writes)
        self.assertEqual(publish.publish_candidate(api, self.key, self.files), 'recorded-preserved')
        self.assertEqual(len(api.writes), writes)

    def test_closed_or_ready_pr_and_human_commits_are_preserved(self):
        for status in ('closed', 'ready', 'human'):
            api = FakeGitHub()
            publish.publish_candidate(api, self.key, self.files)
            if status == 'closed': api.pulls[0]['state'] = 'closed'
            if status == 'ready': api.pulls[0]['draft'] = False
            if status == 'human': api.commits[api.refs[publish.PREFIX+self.key]]['message'] = 'human edit'
            writes = len(api.writes)
            if status == 'human':
                with self.assertRaises(discover.Error): publish.publish_candidate(api, self.key, self.files)
            else:
                self.assertTrue(publish.publish_candidate(api, self.key, self.files).endswith('preserved'))
            self.assertEqual(len(api.writes), writes)

    def test_preview_and_republished_checksum_use_separate_prs(self):
        api = FakeGitHub()
        publish.publish_candidate(api, self.key, self.files)
        preview = candidate_snapshot('8.2.1-preview')
        key, files = discover.bundle(release('8.2.1-preview'), preview, discover.accepted_snapshot())
        publish.publish_candidate(api, key, files)
        replacement = copy.deepcopy(self.snapshot)
        replacement['package_sha512_base64'] = base64.b64encode(b'z'*64).decode()
        key, files = discover.bundle(release(digest=replacement['package_sha512_base64']), replacement, discover.accepted_snapshot())
        publish.publish_candidate(api, key, files)
        self.assertEqual(len(api.pulls), 3)
        self.assertEqual(len({p['head']['ref'] for p in api.pulls}), 3)

    def test_artifact_validation_recomputes_reports_and_rejects_extra_files(self):
        with tempfile.TemporaryDirectory() as tmp:
            directory = pathlib.Path(tmp)
            entry = directory/self.key; entry.mkdir()
            for name, data in self.files.items(): (entry/name).write_bytes(data)
            self.assertEqual(publish.load_bundles(directory), [(self.key, self.files)])
            (entry/'impact.md').write_text('tampered report', encoding='utf-8')
            with self.assertRaises(discover.Error): publish.load_bundles(directory)
            (entry/'impact.md').write_bytes(self.files['impact.md'])
            (entry/'execute.sh').write_text('untrusted code', encoding='utf-8')
            with self.assertRaises(discover.Error): publish.load_bundles(directory)

    def test_fork_publication_rejected(self):
        with self.assertRaises(discover.Error): publish.GitHub('someone/fork', 'synthetic-token')
        with self.assertRaises(discover.Error): publish.GitHub(publish.REPOSITORY, '')

    def test_divergent_concurrent_commit_is_not_force_pushed(self):
        api = FakeGitHub()
        publish.publish_candidate(api, self.key, self.files)
        original = api.request
        branch = publish.PREFIX + self.key
        def concurrent(method, path, data=None, missing=False):
            if method == 'PATCH' and path.startswith('git/refs/'):
                api.refs[branch] = 'human-concurrent-commit'
            return original(method, path, data, missing)
        api.request = concurrent
        changed = copy.deepcopy(self.snapshot)
        changed['runtime_environment']['powershell_version'] = 'changed'
        key, files = discover.bundle(release(), changed, discover.accepted_snapshot())
        with self.assertRaises(discover.Error):
            publish.publish_candidate(api, key, files)
        self.assertEqual(api.refs[branch], 'human-concurrent-commit')

    def test_http_failures_are_visible_without_response_or_token_material(self):
        api = publish.GitHub(publish.REPOSITORY, 'synthetic-secret')
        with patch('publish.urllib.request.build_opener') as opener:
            opener.return_value.open.side_effect = urllib.error.HTTPError('https://api.github.com', 403, 'sensitive body', {}, None)
            with self.assertRaisesRegex(discover.Error, 'status 403') as error:
                api.request('POST', 'pulls', {'body': 'sensitive body'})
            self.assertNotIn('sensitive', str(error.exception))
            self.assertNotIn('synthetic-secret', str(error.exception))
            opener.return_value.open.side_effect = urllib.error.HTTPError('https://api.github.com', 404, 'missing', {}, None)
            self.assertIsNone(api.request('GET', 'git/ref/heads/missing', missing=True))
        self.assertIsNone(publish.NoRedirect().redirect_request(None, None, 302, '', {}, 'https://other.test'))

    def test_markdown_treats_source_paths_as_text(self):
        report = json.loads(self.files['impact.json'])
        report['changes'] = [dict(path=['[image](https://other.test)', '<script>', '`code`'],
                                  classification='metadata', compatibility='review', impact={'implementation_owner': None})]
        rendered = discover.markdown(report)
        self.assertNotIn('[image](', rendered)
        self.assertNotIn('<script>', rendered)
        self.assertNotIn('`code`', rendered)


@unittest.skipUnless(shutil.which('pwsh'), 'PowerShell oracle is unavailable')
class PreviewOracleTests(unittest.TestCase):
    def test_import_checks_prerelease_label_and_stable_mismatch(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            (root/'MicrosoftTeams.psm1').write_text('function Get-Sample { [CmdletBinding()]param([string]$Identity) $Identity }\nExport-ModuleMember -Function Get-Sample', encoding='utf-8')
            (root/'MicrosoftTeams.psd1').write_text("@{RootModule='MicrosoftTeams.psm1';ModuleVersion='8.2.0';GUID='db604b7d-6b3e-4c72-a051-bc2991113815';FunctionsToExport=@('Get-Sample');PrivateData=@{PSData=@{Prerelease='preview'}}}", encoding='utf-8')
            observed = discover.baseline.run_runtime(root, root/'runtime.json', '8.2.0-preview')
            self.assertEqual(observed['module_version'], '8.2.0-preview')
            self.assertFalse(observed['authenticated'])
            with self.assertRaises(discover.Error):
                discover.baseline.run_runtime(root, root/'wrong.json', '8.2.0')


if __name__ == '__main__':
    unittest.main()
