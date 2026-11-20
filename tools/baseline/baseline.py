"""Acquire and inspect a pinned Teams package without authenticating to a tenant.

Uses only Python's standard library. PowerShell is required for fresh runtime
introspection, but never for SDK consumers. No installed modules are modified.
"""
import argparse
import base64
import hashlib
import json
import os
import pathlib
import re
import stat
import subprocess
import tempfile
import urllib.request
import xml.etree.ElementTree as ET
import zipfile

VERSION = '8.1.0'
PIN = '842LJvX0gC9ZmIqN9YOrKWPqh9hVk+BPanw4okt3+bJ5EmknnKHvO1ioRllgLZG4j+h06c1s9y9bs/xIxPrMPQ=='
GALLERY = 'https://www.powershellgallery.com'
MAX_DOWNLOAD = 64 * 1024 * 1024
MAX_EXPANDED = 256 * 1024 * 1024
ROOT = pathlib.Path(__file__).resolve().parents[2]


class DiscoveryError(ValueError):
    """The source could not establish a complete, trustworthy observation."""


def fetch(url, destination, limit=MAX_DOWNLOAD):
    with urllib.request.urlopen(url, timeout=60) as response:
        expected = response.headers.get('Content-Length')
        total = 0
        with destination.open('wb') as stream:
            while chunk := response.read(65536):
                total += len(chunk)
                if total > limit:
                    raise DiscoveryError('Download exceeds size limit')
                stream.write(chunk)
        if not total or (expected is not None and total != int(expected)):
            raise DiscoveryError('Empty or truncated download')


def gallery_metadata(data, version):
    try:
        root = ET.fromstring(data)
    except ET.ParseError as exc:
        raise DiscoveryError('Malformed Gallery metadata') from exc
    ns = '{http://schemas.microsoft.com/ado/2007/08/dataservices}'
    def field(name):
        node = root.find('.//' + ns + name)
        return node.text if node is not None else None
    if field('Id') != 'MicrosoftTeams' or field('Version') != version:
        raise DiscoveryError('Gallery identity/version mismatch')
    if field('PackageHashAlgorithm') != 'SHA512' or not field('PackageHash'):
        raise DiscoveryError('Missing SHA512 metadata')
    names = sorted(set(t[10:] for t in (field('Tags') or '').split() if t.startswith('PSCommand_')), key=str.casefold)
    if not names or any(not re.fullmatch(r'[A-Za-z]+-[A-Za-z0-9]+', n) for n in names):
        raise DiscoveryError('Missing or invalid Gallery command inventory')
    return dict(version=version, sha512=field('PackageHash'), commands=names, release_notes=field('ReleaseNotes') or '')


def verify_archive(path, metadata, expected_hash):
    actual = base64.b64encode(hashlib.sha512(path.read_bytes()).digest()).decode()
    if actual != metadata['sha512'] or actual != expected_hash:
        raise DiscoveryError('Package SHA512 differs from metadata or explicit pin')


def extract_archive(path, target):
    try:
        with zipfile.ZipFile(path) as archive:
            entries = archive.infolist()
            if sum(i.file_size for i in entries) > MAX_EXPANDED:
                raise DiscoveryError('Expanded package exceeds size limit')
            seen = set()
            for entry in entries:
                name = entry.filename.replace('\\', '/')
                parts = pathlib.PurePosixPath(name)
                if parts.is_absolute() or '..' in parts.parts or ':' in name or '\x00' in name:
                    raise DiscoveryError('Unsafe archive path')
                if name.casefold() in seen or stat.S_ISLNK(entry.external_attr >> 16):
                    raise DiscoveryError('Duplicate archive path or symlink')
                seen.add(name.casefold())
                destination = target.joinpath(*parts.parts)
                if entry.is_dir():
                    destination.mkdir(parents=True, exist_ok=True)
                else:
                    destination.parent.mkdir(parents=True, exist_ok=True)
                    destination.write_bytes(archive.read(entry))
    except (zipfile.BadZipFile, EOFError) as exc:
        raise DiscoveryError('Invalid or truncated package archive') from exc
    manifest = target / 'MicrosoftTeams.psd1'
    if not manifest.is_file():
        raise DiscoveryError('Root module manifest is missing')


def run_runtime(module, output, version, pwsh='pwsh'):
    # Inherit only process/platform configuration, not CI cloud credentials.
    env = {k: v for k, v in os.environ.items() if k in {'PATH','SystemRoot','WINDIR','TEMP','TMP','TMPDIR','DOTNET_ROOT','LANG'}}
    try:
        result = subprocess.run([pwsh, '-NoLogo', '-NoProfile', '-NonInteractive', '-File',
                                 str(ROOT/'scripts/discovery/export-command-metadata.ps1'),
                                 '-ModulePath', str(module/'MicrosoftTeams.psd1'),
                                 '-OutputPath', str(output), '-ExpectedVersion', version],
                                capture_output=True, text=True, timeout=300, env=env)
    except (OSError, subprocess.TimeoutExpired) as exc:
        raise DiscoveryError('PowerShell import unavailable or timed out') from exc
    if result.returncode != 0 or not output.is_file():
        # Do not echo process output that could include user environment values.
        raise DiscoveryError(f'PowerShell import failed (exit {result.returncode})')
    return json.loads(output.read_text(encoding='utf-8-sig'))


def canonical_runtime(runtime, version):
    if runtime.get('module_version') != version or runtime.get('authenticated') is not False:
        raise DiscoveryError('Runtime version/authentication provenance mismatch')
    if not runtime.get('powershell_version') or not runtime.get('platform'):
        raise DiscoveryError('Runtime environment provenance is missing')
    commands = runtime.get('commands')
    if not isinstance(commands, list) or not commands:
        raise DiscoveryError('Empty runtime inventory')
    seen = set()
    normalized = []
    for command in commands:
        name = command.get('name', '')
        if not re.fullmatch(r'[A-Za-z]+-[A-Za-z0-9]+', name) or name.casefold() in seen:
            raise DiscoveryError('Invalid or duplicate runtime command')
        seen.add(name.casefold())
        sets = command.get('parameter_sets')
        if not isinstance(sets, list) or not sets or not command.get('command_type') or not isinstance(command.get('output_types'), list):
            raise DiscoveryError('Incomplete runtime command metadata: ' + name)
        for group in sets:
            if not group.get('name') or not isinstance(group.get('parameters'), list) or not isinstance(group.get('is_default'), bool):
                raise DiscoveryError('Incomplete parameter set: ' + name)
            for param in group['parameters']:
                if not param.get('name') or not param.get('type') or not isinstance(param.get('mandatory'), bool):
                    raise DiscoveryError('Incomplete parameter metadata: ' + name)
                if (not isinstance(param.get('position'), int)
                        or not isinstance(param.get('aliases'), list)
                        or not isinstance(param.get('validation'), list)
                        or not isinstance(param.get('pipeline_input'), bool)
                        or not isinstance(param.get('pipeline_by_property'), bool)):
                    raise DiscoveryError('Incomplete parameter attributes: ' + name)
                param['aliases'] = sorted(param['aliases'])
            group['parameters'].sort(key=lambda p: p['name'].casefold())
        command['parameter_sets'] = sorted(sets, key=lambda s: s['name'])
        command['output_types'] = sorted(command.get('output_types', []))
        normalized.append(command)
    return sorted(normalized, key=lambda c: c['name'].casefold())


def package_sources(module):
    sources = []
    for path in sorted(module.rglob('*')):
        if not path.is_file():
            continue
        name = path.relative_to(module).as_posix()
        lower = name.lower()
        if not (lower.endswith(('.psd1','.psm1','.ps1','help.xml')) or path.name in ['CmdletConfig.json','CmdletSettings.json']):
            continue
        raw = path.read_bytes()
        source = dict(path=name,sha256=hashlib.sha256(raw).hexdigest())
        if lower.endswith('.psd1'):
            text = raw.decode('utf-8-sig')
            active = '\n'.join(line for line in text.splitlines() if not line.lstrip().startswith('#'))
            source['lexical_command_candidates'] = sorted(set(re.findall(r"['\"]([A-Za-z]+-[A-Za-z0-9]+)['\"]",active)),key=str.casefold)
        elif lower.endswith('help.xml'):
            try:
                tree = ET.fromstring(raw)
            except ET.ParseError as exc:
                raise DiscoveryError('Invalid help XML: '+name) from exc
            # Preserve schema facts, not full copyrighted help prose.
            source['help_commands'] = sorted(set((n.text or '').strip() for n in tree.findall('.//{*}details/{*}name') if (n.text or '').strip()))
        elif path.name in ['CmdletConfig.json','CmdletSettings.json']:
            try:
                source['configuration'] = json.loads(raw.decode('utf-8-sig'))
            except (ValueError,UnicodeError) as exc:
                raise DiscoveryError('Invalid command configuration: '+name) from exc
        sources.append(source)
    if not any('help_commands' in s for s in sources) or not any('configuration' in s for s in sources):
        raise DiscoveryError('Package help/configuration extraction is incomplete')
    return sources


def snapshot(module, archive, metadata, runtime):
    commands = canonical_runtime(runtime, metadata['version'])
    sources = package_sources(module)
    return dict(schema_version=1,module_version=metadata['version'],
                package_sha512_base64=metadata['sha512'],package_sha256=hashlib.sha256(archive.read_bytes()).hexdigest(),
                package_url=f"{GALLERY}/api/v2/package/MicrosoftTeams/{metadata['version']}",
                runtime_environment={k:v for k,v in runtime.items() if k != 'commands'},
                gallery_commands=metadata['commands'],release_notes=metadata['release_notes'],
                runtime_commands=commands,sources=sources,
                limitations=['Unauthenticated runtime observation; platform/session exports require reconciliation.',
                             'Manifest candidates are lexical hints, not evaluated exports.',
                             'Metadata does not establish wire behaviour, defaults or complete output schemas.'])


def publish_snapshot(output, value):
    data = (json.dumps(value, sort_keys=True, indent=2, ensure_ascii=False)+'\n').encode()
    output.parent.mkdir(parents=True, exist_ok=True)
    if output.exists():
        if output.read_bytes() == data:
            return 'unchanged'
        raise DiscoveryError('Output exists with different content; choose a new candidate path')
    # Atomic hard-link creation cannot overwrite a concurrent publisher's output.
    with tempfile.NamedTemporaryFile(dir=output.parent, delete=False) as stream:
        staged = pathlib.Path(stream.name)
        stream.write(data)
    try:
        try:
            os.link(staged, output)
        except FileExistsError:
            if output.read_bytes() != data:
                raise DiscoveryError('Concurrent snapshot differs; output was not replaced')
    finally:
        staged.unlink()
    return 'created'


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--version', default=VERSION)
    parser.add_argument('--expected-sha512', default=PIN)
    parser.add_argument('--archive', type=pathlib.Path)
    parser.add_argument('--metadata', type=pathlib.Path, help='Saved Gallery XML; otherwise fetch exact version')
    parser.add_argument('--runtime', type=pathlib.Path, help='Saved raw exporter JSON; otherwise run isolated PowerShell')
    parser.add_argument('--output', type=pathlib.Path, required=True)
    args = parser.parse_args()
    if not re.fullmatch(r'\d+\.\d+\.\d+(?:-[A-Za-z0-9.-]+)?', args.version):
        parser.error('Invalid version')
    try:
        with tempfile.TemporaryDirectory(prefix='msteams-baseline-') as directory:
            work = pathlib.Path(directory)
            metadata_path = args.metadata or work/'gallery.xml'
            if not args.metadata:
                fetch(f"{GALLERY}/api/v2/Packages(Id='MicrosoftTeams',Version='{args.version}')",metadata_path,1024*1024)
            metadata = gallery_metadata(metadata_path.read_bytes(),args.version)
            archive = args.archive or work/'module.zip'
            if not args.archive:
                fetch(f'{GALLERY}/api/v2/package/MicrosoftTeams/{args.version}',archive)
            verify_archive(archive,metadata,args.expected_sha512)
            module = work/'module'
            extract_archive(archive,module)
            runtime = json.loads(args.runtime.read_text(encoding='utf-8-sig')) if args.runtime else run_runtime(module,work/'runtime.json',args.version)
            value = snapshot(module,archive,metadata,runtime)
            state = publish_snapshot(args.output,value)
            print(f'{state}: {args.output}; {len(value["runtime_commands"])} runtime commands, {len(value["gallery_commands"])} Gallery candidates')
    except (DiscoveryError,OSError,ValueError) as exc:
        parser.exit(1,f'Discovery failed: {exc}\n')


if __name__ == '__main__':
    main()
