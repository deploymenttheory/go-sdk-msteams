"""Reconcile discovery evidence without turning absent exports into exclusions."""
import argparse
import collections
import hashlib
import json
import pathlib
import re

from baseline import DiscoveryError, PIN, ROOT, publish_snapshot


def command_literals(text):
    return re.findall(r"['\"]([A-Za-z]+-[A-Za-z0-9]+)['\"]", text)


def reconcile(snapshot, module, ownership):
    if snapshot.get('module_version') != '8.1.0' or snapshot.get('package_sha512_base64') != PIN:
        raise DiscoveryError('Reconciliation requires the pinned 8.1.0 observation')
    if not snapshot.get('runtime_commands') or not snapshot.get('gallery_commands'):
        raise DiscoveryError('Incomplete observation')
    sources = snapshot['sources']
    root = next(s for s in sources if s['path'] == 'MicrosoftTeams.psd1')
    raw = (module / root['path']).read_bytes()
    if hashlib.sha256(raw).hexdigest() != root['sha256']:
        raise DiscoveryError('Root manifest differs from the observed package')
    comments, oce = {}, {}
    in_oce = False
    for number, line in enumerate(raw.decode('utf-8-sig').splitlines(), 1):
        if '#MPA OCE cmdlets' in line:
            in_oce = True
        if line.lstrip().startswith('#preview'):
            in_oce = False
        for name in command_literals(line):
            if line.lstrip().startswith('#'):
                comments.setdefault(name.casefold(), []).append(number)
            elif in_oce:
                oce.setdefault(name.casefold(), []).append(number)
    owners = {}
    for item in ownership:
        key = item['command'].casefold()
        if key in owners:
            raise DiscoveryError('Duplicate command ownership: ' + item['command'])
        owners[key] = item
    names = {}
    manifests = collections.defaultdict(list)
    help_files = collections.defaultdict(list)
    config = {}
    for source in sources:
        for name in source.get('lexical_command_candidates', []):
            names.setdefault(name.casefold(), name)
            manifests[name.casefold()].append(source['path'])
        for name in source.get('help_commands', []):
            help_files[name.casefold()].append(source['path'])
        for name, settings in source.get('configuration', {}).get('CmdletConfiguration', {}).items():
            config[name.casefold()] = dict(path=source['path'], settings=settings)
    gallery = {n.casefold() for n in snapshot['gallery_commands']}
    names.update({n.casefold(): n for n in snapshot['gallery_commands']})
    runtime = {}
    for item in snapshot['runtime_commands']:
        key = item['name'].casefold()
        if key in runtime:
            raise DiscoveryError('Duplicate runtime command: ' + item['name'])
        runtime[key] = item
        names.setdefault(key, item['name'])
    records = []
    for key, name in sorted(names.items()):
        owner = owners.get(key)
        observed = runtime.get(key)
        setting = config.get(key)
        exports = setting['settings'].get('ExportsTo', []) if setting else []
        if observed:
            availability = 'observed-default-session'
        elif exports and 'TeamsGA' not in exports:
            availability = 'conditional-config-export'
        elif key in oce:
            availability = 'conditional-oce-declaration'
        elif key not in gallery:
            availability = 'nested-only-candidate'
        else:
            availability = 'unresolved-export'
        disposition = owner['disposition'] if owner else 'research-candidate'
        if disposition == 'management-candidate' and observed:
            disposition = 'observed-management-export'
        if owner and owner['family'] == 'attendant_builders':
            disposition = 'local-builder-candidate'
        records.append(dict(
            name=name, gallery=key in gallery, runtime_name=observed['name'] if observed else None,
            command_type=observed['command_type'] if observed else None,
            parameter_set_names=[s['name'] for s in observed['parameter_sets']] if observed else [],
            availability=availability, disposition=disposition,
            family=owner['family'] if owner else 'unassigned-nested-exports',
            research_owner=owner['research'] if owner else 'BASE-02',
            implementation_owner=owner['implementation'] if owner else None,
            evidence=dict(manifests=manifests[key], help_files=help_files[key],
                          configuration=setting, oce_manifest_lines=oce.get(key, []),
                          commented_root_lines=comments.get(key, [])),
            qualification='metadata-only', implementation_status='not-started'))
    return dict(schema_version=1, module_version='8.1.0',
                package_sha256=snapshot['package_sha256'],
                environment=snapshot['runtime_environment'],
                source_hashes={s['path']: s['sha256'] for s in sources},
                counts=dict(candidates=len(records), gallery=len(gallery), runtime=len(runtime),
                            availability=dict(sorted(collections.Counter(r['availability'] for r in records).items())),
                            dispositions=dict(sorted(collections.Counter(r['disposition'] for r in records).items()))),
                records=records,
                unresolved_exports=[r['name'] for r in records if r['availability'] == 'unresolved-export'],
                nested_candidates=[r['name'] for r in records if r['availability'] == 'nested-only-candidate'],
                limitations=[
                    'Default unauthenticated observation only; no tenant or privileged session was connected.',
                    'Conditional declarations explain absence but do not prove public support or exclusion.',
                    'Nested manifest candidates are lexical hints, not evaluated root exports.',
                    'Local-builder classification is a research assignment; behavioural proof belongs to the family contract ticket.',
                    'Help metadata does not establish current public documentation or a complete wire contract.',
                    'Only the previously approved telemetry exclusion is retained.'])


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--snapshot', required=True, type=pathlib.Path)
    parser.add_argument('--module', required=True, type=pathlib.Path)
    parser.add_argument('--output', required=True, type=pathlib.Path)
    args = parser.parse_args()
    try:
        value = reconcile(json.loads(args.snapshot.read_text(encoding='utf-8')),
                          args.module, json.loads((ROOT/'docs/implementation/backlog/coverage.json').read_text(encoding='utf-8')))
        print(publish_snapshot(args.output, value), value['counts'])
    except (OSError, ValueError, KeyError, StopIteration) as exc:
        parser.exit(1, f'Reconciliation failed: {exc}\n')


if __name__ == '__main__':
    main()
