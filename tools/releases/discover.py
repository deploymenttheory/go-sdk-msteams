"""Discover upstream candidates without changing the accepted Teams baseline."""
import argparse
import base64
import gzip
import hashlib
import html
import io
import json
import pathlib
import re
import sys
import tempfile
import urllib.parse
import xml.etree.ElementTree as ET

ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / 'tools/baseline'))
sys.path.insert(0, str(ROOT / 'tools/contractdiff'))
import baseline
from compare import compare, normalize

Error = baseline.DiscoveryError
ATOM = '{http://www.w3.org/2005/Atom}'
DATA = '{http://schemas.microsoft.com/ado/2007/08/dataservices}'
FEED = baseline.GALLERY + "/api/v2/FindPackagesById()?id='MicrosoftTeams'"
ACCEPTED = ROOT / 'tools/contractdiff/tests/fixtures/8.1.0.json.gz'
MAX_JSON = 64 * 1024 * 1024
VERSION = re.compile(r'(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([A-Za-z0-9]+(?:[.-][A-Za-z0-9]+)*))?')


def version_key(version):
    match = VERSION.fullmatch(version) if isinstance(version, str) and len(version) <= 64 else None
    if not match:
        raise Error('Unsupported Gallery version')
    major, minor, patch, label = match.groups()
    # NuGet/SemVer numeric identifiers sort numerically; non-numeric labels are case-insensitive.
    parts = tuple((0, int(p)) if p.isdigit() else (1, p.casefold()) for p in label.split('.')) if label else ()
    return (int(major), int(minor), int(patch), 0 if label else 1, parts)


def checksum(value):
    try:
        raw = base64.b64decode(value, validate=True)
    except (ValueError, TypeError) as exc:
        raise Error('Missing or malformed Gallery SHA512') from exc
    if len(raw) != 64:
        raise Error('Missing or malformed Gallery SHA512')
    return raw


def candidate_key(version, sha512):
    version_key(version)
    # Full hash of the source digest keeps the ref path bounded without truncating identity.
    return version.lower() + '-' + hashlib.sha256(checksum(sha512)).hexdigest()


def read_json(path):
    opener = gzip.open if path.suffix == '.gz' else open
    with opener(path, 'rb') as stream:
        raw = stream.read(MAX_JSON + 1)
    if len(raw) > MAX_JSON:
        raise Error('Candidate JSON exceeds size limit')
    return json.loads(raw.decode('utf-8'))


def json_bytes(value):
    return (json.dumps(value, sort_keys=True, indent=2, ensure_ascii=False) + '\n').encode('utf-8')


def parse_feed(raw):
    try:
        root = ET.fromstring(raw)
    except ET.ParseError as exc:
        raise Error('Malformed Gallery release feed') from exc
    if root.tag != ATOM + 'feed':
        raise Error('Expected an Atom release feed')
    releases = []
    for entry in root.findall(ATOM + 'entry'):
        def field(name):
            node = entry.find('.//' + DATA + name)
            return node.text if node is not None else None
        if field('Id') != 'MicrosoftTeams':
            raise Error('Release feed contains a different package')
        version = field('Version')
        version_key(version)
        channel = 'preview' if '-' in version else 'stable'
        if field('IsPrerelease') != ('true' if channel == 'preview' else 'false'):
            raise Error('Release channel metadata disagrees with version')
        if field('PackageHashAlgorithm') != 'SHA512':
            raise Error('Unsupported Gallery hash algorithm')
        sha512 = field('PackageHash')
        checksum(sha512)
        if field('Listed') == 'false':
            continue
        releases.append(dict(version=version, channel=channel, sha512=sha512))
    links = [x.get('href') for x in root.findall(ATOM + 'link') if x.get('rel') == 'next']
    if len(links) > 1:
        raise Error('Ambiguous Gallery pagination')
    return releases, links[0] if links else None


def gallery_releases(fetch=baseline.fetch):
    releases, seen = {}, set()
    url = FEED
    with tempfile.TemporaryDirectory(prefix='msteams-feed-') as directory:
        page = pathlib.Path(directory) / 'page.xml'
        while url:
            parsed = urllib.parse.urlsplit(url)
            if (parsed.scheme != 'https' or parsed.netloc != 'www.powershellgallery.com'
                    or parsed.path not in ('/api/v2/FindPackagesById()', '/api/v2/FindPackagesById')
                    or parsed.fragment or url in seen or len(seen) >= 50):
                raise Error('Unsafe, cyclic or excessive Gallery pagination')
            seen.add(url)
            fetch(url, page, 8 * 1024 * 1024)
            entries, next_url = parse_feed(page.read_bytes())
            for entry in entries:
                key = entry['version'].casefold()
                if key in releases and releases[key] != entry:
                    raise Error('Conflicting release identity across Gallery pages')
                releases[key] = entry
            url = urllib.parse.urljoin(url, next_url) if next_url else None
    if not releases:
        raise Error('Empty Gallery release inventory')
    return list(releases.values())


def select_candidates(releases, accepted):
    """Select the latest eligible release in each channel and baseline republishing."""
    version = accepted['module_version']
    current = [r for r in releases if r['version'] == version]
    if len(current) != 1:
        raise Error('Accepted baseline is missing or ambiguous in Gallery')
    selected = []
    if current[0]['sha512'] != accepted['package_sha512_base64']:
        selected.append(current[0])
    for channel in ('stable', 'preview'):
        options = [r for r in releases if r['channel'] == channel and version_key(r['version']) > version_key(version)]
        if options:
            selected.append(max(options, key=lambda r: version_key(r['version'])))
    return selected


def acquire(release, fetch=baseline.fetch, runtime=baseline.run_runtime):
    version, expected = release['version'], release['sha512']
    with tempfile.TemporaryDirectory(prefix='msteams-candidate-') as directory:
        work = pathlib.Path(directory)
        metadata_path, archive = work / 'gallery.xml', work / 'module.zip'
        fetch(f"{baseline.GALLERY}/api/v2/Packages(Id='MicrosoftTeams',Version='{version}')", metadata_path, 1024 * 1024)
        metadata = baseline.gallery_metadata(metadata_path.read_bytes(), version)
        if metadata['sha512'] != expected:
            raise Error('Exact metadata differs from discovery checksum; retry discovery')
        fetch(f'{baseline.GALLERY}/api/v2/package/MicrosoftTeams/{version}', archive)
        baseline.verify_archive(archive, metadata, expected)
        module = work / 'module'
        baseline.extract_archive(archive, module)
        observation = runtime(module, work / 'runtime.json', version)
        value = baseline.snapshot(module, archive, metadata, observation)
        # Keep released prose out of repository snapshots, matching the historical oracle.
        value['release_notes'] = 'sha256:' + hashlib.sha256(value['release_notes'].encode()).hexdigest()
        normalize(value)
        return value


def accepted_snapshot():
    value = read_json(ACCEPTED)
    normalize(value)
    if value['module_version'] != baseline.VERSION or value['package_sha512_base64'] != baseline.PIN:
        raise Error('Accepted snapshot does not match pinned baseline')
    return value


def impact_report(before, after):
    return compare(before, after, read_json(ROOT / 'contracts/coverage/8.1.0.json'),
                   read_json(ROOT / 'docs/implementation/backlog/publication/mapping.json'))


def markdown(report):
    # Only JSON-escaped, HTML-escaped data is rendered; no source-provided Markdown links.
    def cell(value):
        escaped = html.escape(str(value)).replace('\n', ' ').replace('\r', ' ')
        return ''.join('&#' + str(ord(c)) + ';' if c in '|`[]()*_!\\' else c for c in escaped)
    lines = ['# MicrosoftTeams candidate impact', '',
             f"Baseline: {cell(report['before']['version'])}. Candidate: {cell(report['after']['version'])} ({cell(report['after']['channel'])}).",
             '', f"Status: {report['status']}; changes: {len(report['changes'])}; warnings: {len(report['warnings'])}.",
             '', 'This is review evidence. It neither promotes a baseline nor establishes live tenant compatibility.', '',
             '| Contract path | Classification | Compatibility | Implementation owner |', '|---|---|---|---|']
    for change in report['changes'][:100]:
        lines.append('| ' + ' | '.join(cell(v) for v in ('/'.join(change['path']), change['classification'],
                     change['compatibility'], change['impact']['implementation_owner'] or 'assignment required')) + ' |')
    lines += ['', 'The JSON impact report contains every change, old/new value, warning and owner/test mapping.', '']
    return '\n'.join(lines)


def bundle(release, snapshot, before):
    if snapshot['module_version'] != release['version'] or snapshot['package_sha512_base64'] != release['sha512']:
        raise Error('Candidate snapshot identity differs from discovery')
    report = impact_report(before, snapshot)
    key = candidate_key(release['version'], release['sha512'])
    descriptor = dict(schema_version=1, key=key, **release,
                      package_sha256=snapshot['package_sha256'], baseline_version=before['module_version'],
                      baseline_sha512=before['package_sha512_base64'])
    # mtime=0 avoids time-dependent commits. Compress through GzipFile for a portable header.
    stream = io.BytesIO()
    with gzip.GzipFile(fileobj=stream, mode='wb', mtime=0, filename='') as output:
        output.write(json_bytes(snapshot))
    return key, {'candidate.json': json_bytes(descriptor), 'snapshot.json.gz': stream.getvalue(),
                 'impact.json': json_bytes(report), 'impact.md': markdown(report).encode('utf-8')}


def discover(output, releases=None, acquisition=acquire):
    if output.exists():
        raise Error('Use a new output directory for each discovery run')
    before = accepted_snapshot()
    selected = select_candidates(gallery_releases() if releases is None else releases, before)
    # Build everything before publication so extraction failure cannot leave a partial run.
    bundles = [bundle(release, acquisition(release), before) for release in selected]
    if not bundles:
        return 0
    output.mkdir(parents=True)
    for key, files in bundles:
        target = output / key
        target.mkdir()
        for name, data in files.items():
            (target / name).write_bytes(data)
    return len(bundles)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=pathlib.Path, required=True)
    args = parser.parse_args()
    try:
        count = discover(args.output)
        print(json.dumps({'candidates': count, 'baseline': baseline.VERSION}))
    except (Error, OSError, ValueError, KeyError, TypeError, AttributeError) as exc:
        # Avoid echoing any untrusted package/module diagnostics from a failed observation.
        print(json.dumps({'status': 'incomplete', 'error_type': type(exc).__name__,
                          'reason': str(exc) if isinstance(exc, Error) else 'Release discovery failed before publication'}))
        return 1
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
