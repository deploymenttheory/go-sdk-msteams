"""Publish verified candidate data as draft PRs; never promote the baseline."""
import argparse
import base64
import hashlib
import json
import os
import pathlib
import urllib.error
import urllib.parse
import urllib.request

from discover import Error, accepted_snapshot, bundle, candidate_key, read_json, version_key

REPOSITORY = 'deploymenttheory/go-sdk-msteams'
PREFIX = 'automation/microsoftteams/'
FILES = {'candidate.json', 'snapshot.json.gz', 'impact.json', 'impact.md'}


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        # Publication credentials must never follow a redirected API request.
        return None


class GitHub:
    def __init__(self, repository, token):
        if repository != REPOSITORY or not token:
            raise Error('Publication requires the upstream repository and an explicit GitHub token')
        self.repository, self.token = repository, token

    def request(self, method, path, data=None, missing=False):
        url = 'https://api.github.com/repos/' + self.repository + '/' + path
        payload = json.dumps(data).encode() if data is not None else None
        request = urllib.request.Request(url, data=payload, method=method, headers={
            'Authorization': 'Bearer ' + self.token, 'Accept': 'application/vnd.github+json',
            'X-GitHub-Api-Version': '2022-11-28', 'User-Agent': 'msteams-candidate-discovery',
            'Content-Type': 'application/json'})
        try:
            with urllib.request.build_opener(NoRedirect).open(request, timeout=60) as response:
                raw = response.read(32 * 1024 * 1024 + 1)
                if len(raw) > 32 * 1024 * 1024:
                    raise Error('GitHub response exceeded size limit')
                return json.loads(raw) if raw else None
        except urllib.error.HTTPError as exc:
            status = exc.code
            exc.close()
            if missing and status == 404:
                return None
            raise Error(f'GitHub {method} failed with status {status}; retry after inspection') from None
        except urllib.error.URLError:
            raise Error('GitHub request failed; retry after inspection') from None


def load_bundles(directory):
    before = accepted_snapshot()
    if not directory.is_dir() or directory.is_symlink():
        raise Error('Candidate artifact directory is missing or unsafe')
    entries = sorted(directory.iterdir())
    if not entries or len(entries) > 3:
        raise Error('Expected one to three candidate bundles')
    result = []
    for entry in entries:
        if entry.is_symlink() or not entry.is_dir() or {p.name for p in entry.iterdir()} != FILES:
            raise Error('Unexpected artifact paths')
        for path in entry.iterdir():
            if path.is_symlink() or not path.is_file() or path.stat().st_size > 64 * 1024 * 1024:
                raise Error('Unsafe candidate artifact')
        descriptor = read_json(entry / 'candidate.json')
        version, sha512 = descriptor['version'], descriptor['sha512']
        key = candidate_key(version, sha512)
        channel = 'preview' if '-' in version else 'stable'
        if (entry.name != key or descriptor['channel'] != channel
                or version_key(version) < version_key(before['module_version'])
                or (version == before['module_version'] and sha512 == before['package_sha512_base64'])):
            raise Error('Invalid candidate identity or channel')
        snapshot = read_json(entry / 'snapshot.json.gz')
        generated_key, generated = bundle(dict(version=version, sha512=sha512, channel=channel), snapshot, before)
        # Recompute every report from a fresh trusted checkout; artifact strings are data only.
        if generated_key != key or any((entry / name).read_bytes() != data for name, data in generated.items()):
            raise Error('Candidate content differs from trusted recomputation')
        result.append((key, generated))
    return result


def blob_sha(data):
    return hashlib.sha1(b'blob ' + str(len(data)).encode() + b'\0' + data).hexdigest()


def pr_metadata(key, files):
    descriptor = json.loads(files['candidate.json'])
    report = json.loads(files['impact.json'])
    marker = '<!-- msteams-upstream-candidate:' + key + ' -->'
    title = f"chore(delta): review MicrosoftTeams {descriptor['version']} ({descriptor['channel']})"
    body = '\n'.join([
        marker, '', '## Summary', '',
        '```text', f"8.1.0 accepted baseline -> {descriptor['version']} {descriptor['channel']} candidate",
        '  snapshot + JSON/Markdown impact reports', '  human contract review -> separate adoption decision', '```', '',
        f"Source SHA512: `{descriptor['sha512']}`.",
        f"Package SHA256: `{descriptor['package_sha256']}`.", '',
        '## Evidence', '',
        f"- **Before:** immutable MicrosoftTeams {descriptor['baseline_version']} baseline.",
        f"  **After:** {len(report['changes'])} compared changes and {len(report['warnings'])} review warnings.",
        f"- Reports and candidate provenance: `contracts/candidates/{key}/`.",
        '- SHA512 verified against Gallery metadata before isolated PowerShell import; unauthenticated metadata only.',
        '- Stable and preview candidates are separate. This draft does not change the accepted baseline or SDK parity claims.', '',
        '## Merge Danger', '', '**Door:** Two-way. Candidate evidence only; adoption requires a separate reviewed change.', '',
        '**Blast Radius:** Discovery. No service code, tenant operations, automatic merge or baseline promotion.', ''])
    return marker, title, body


def publish_candidate(api, key, files, base='main'):
    if base != 'main':
        raise Error('Candidate base must be main')
    branch = PREFIX + key
    owner = api.repository.split('/')[0]
    marker, title, body = pr_metadata(key, files)
    query = urllib.parse.urlencode({'state': 'all', 'head': owner + ':' + branch, 'base': base, 'per_page': 100})
    pulls = api.request('GET', 'pulls?' + query)
    if len(pulls) > 1:
        raise Error('Duplicate candidate PRs require manual reconciliation')
    pull = pulls[0] if pulls else None
    if pull:
        if marker not in (pull.get('body') or ''):
            raise Error('Candidate branch belongs to an unrelated PR')
        if pull['state'] != 'open':
            return 'closed-preserved'
        if not pull['draft']:
            return 'review-preserved'
        if pull['head']['repo']['full_name'] != api.repository or pull['head']['ref'] != branch:
            raise Error('Candidate PR has an unexpected source')
    reference = api.request('GET', 'git/ref/heads/' + branch, missing=True)
    if not reference and pull:
        raise Error('Candidate PR branch is missing')
    if reference:
        parent = reference['object']['sha']
        commit = api.request('GET', 'git/commits/' + parent)
        if marker not in commit['message']:
            raise Error('Candidate branch has human or unrelated changes; refusing to overwrite')
    else:
        main = api.request('GET', 'git/ref/heads/' + base)
        parent = main['object']['sha']
        commit = api.request('GET', 'git/commits/' + parent)
    tree = api.request('GET', 'git/trees/' + commit['tree']['sha'] + '?recursive=1')
    if tree.get('truncated'):
        raise Error('Cannot validate a truncated repository tree')
    inventory = {entry['path']: entry for entry in tree['tree']}
    prefix = 'contracts/candidates/' + key + '/'
    paths = {prefix + name: data for name, data in files.items()}
    if not reference and any(p.startswith(prefix) for p in inventory):
        return 'recorded-preserved'
    if reference:
        existing = {p for p in inventory if p.startswith(prefix) and inventory[p]['type'] != 'tree'}
        if existing != set(paths):
            raise Error('Candidate branch contains unexpected bundle paths')
    changed = {p: data for p, data in paths.items()
               if inventory.get(p, {}).get('sha') != blob_sha(data) or inventory.get(p, {}).get('mode') != '100644'}
    if changed:
        entries = []
        for path, data in sorted(changed.items()):
            blob = api.request('POST', 'git/blobs', {'content': base64.b64encode(data).decode(), 'encoding': 'base64'})
            entries.append({'path': path, 'mode': '100644', 'type': 'blob', 'sha': blob['sha']})
        new_tree = api.request('POST', 'git/trees', {'base_tree': commit['tree']['sha'], 'tree': entries})
        new_commit = api.request('POST', 'git/commits', {'message': title + '\n\n' + marker,
                              'tree': new_tree['sha'], 'parents': [parent]})
        if reference:
            # Non-force update rejects divergent concurrent commits; retry never overwrites them.
            api.request('PATCH', 'git/refs/heads/' + branch, {'sha': new_commit['sha'], 'force': False})
        else:
            api.request('POST', 'git/refs', {'ref': 'refs/heads/' + branch, 'sha': new_commit['sha']})
    if not pull:
        api.request('POST', 'pulls', {'title': title, 'body': body, 'head': branch, 'base': base,
                                     'draft': True, 'maintainer_can_modify': False})
        return 'created'
    if pull['title'] != title or pull['body'] != body:
        api.request('PATCH', 'pulls/' + str(pull['number']), {'title': title, 'body': body})
        return 'updated'
    return 'updated' if changed else 'unchanged'


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--input', type=pathlib.Path, required=True)
    parser.add_argument('--repository', required=True)
    args = parser.parse_args()
    try:
        # Complete validation before any GitHub mutation.
        candidates = load_bundles(args.input)
        api = GitHub(args.repository, os.environ.get('GH_TOKEN') or os.environ.get('GITHUB_TOKEN'))
        for key, files in candidates:
            print(key + ': ' + publish_candidate(api, key, files))
    except (Error, OSError, ValueError, KeyError, TypeError, AttributeError) as exc:
        print(json.dumps({'status': 'incomplete', 'error_type': type(exc).__name__,
                          'reason': str(exc) if isinstance(exc, Error) else 'Candidate publication failed; retry after inspection'}))
        return 1
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
