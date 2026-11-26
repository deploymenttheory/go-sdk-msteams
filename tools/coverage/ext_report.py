#!/usr/bin/env python3
"""Render the EXT inventory, retaining unowned nested research candidates."""
import argparse
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
DESTINATION = ROOT / 'docs/implementation/ext/coverage-report.md'
FAMILIES = {
    'ai_agents': 'ai-agents-contract-notes.md',
    'app_catalog': 'app-catalog-contract-notes.md',
    'app_installation': 'app-installation-contract-notes.md',
    'cloud_call_data': 'cloud-call-data-contract-notes.md',
    'configuration_diagnostics': 'configuration-diagnostics-contract-notes.md',
    'devices': 'devices-contract-notes.md',
    'migration': 'migration-contract-notes.md',
    'provisioning_diagnostics': 'provisioning-diagnostics-contract-notes.md',
    'shifts_connections': 'shifts-connections-contract-notes.md',
    'shifts_sync': 'shifts-sync-contract-notes.md',
    'tenant_identity': 'tenant-identity-contract-notes.md',
    'unified_apps': 'unified-apps-contract-notes.md',
}


def reference(value):
    return f"[`{value['name']}`](../../../{value['file']})"


def render(ledger):
    rows = sorted((r for r in ledger['records'] if (r['implementation_owner'] or '').startswith('EXT-')),
                  key=lambda r: (r['resource'], r['name'].casefold()))
    nested = sorted((r for r in ledger['records'] if not r['implementation_owner']
                     and (r['research_owner'] or '').startswith('EXT-')), key=lambda r: r['name'].casefold())
    if {r['resource'] for r in rows} != FAMILIES.keys():
        raise ValueError('EXT family set changed; review scope before regenerating')
    sets = sum(len(r['parameter_sets']) for r in rows)
    lines = ['# EXT command and parameter-set evidence inventory', '',
             'Owner: EXT-90 (#191). Baseline: MicrosoftTeams 8.1.0.', '',
             f'The ledger assigns {len(rows)} commands and {sets} observed parameter sets to {len(FAMILIES)} EXT families.',
             f'An additional {len(nested)} unowned nested candidates retain EXT research ownership.',
             'Regenerate with `python3 tools/coverage/ext_report.py`; `--check` rejects stale output.', '',
             '## Evidence boundaries', '',
             'These are declared symbol/test mappings and recorded ledger states, not proof that a test ran',
             'or that every behavior is complete. Conditional commands with no observed parameter sets remain',
             'visible; a missing runtime signature is not an empty supported surface. No ledger state is promoted.', '',
             'The [integrated suite](../../../tests/integration/lab/ext_integration/contracts_test.go) constrains',
             'generic/typed configuration identity and presence, independent recovery after typed authorization',
             'failure, tag-template identity handed to an agent configuration, and independent template',
             'cleanup after agent authorization failure. These compositions use',
             'synthetic replies. The separate opt-in `TestAccGenericTypedConfigurationComposition` creates an',
             'owned unassigned feedback policy, crosses the generic/typed adapters, requires three pinned',
             'comparisons, and independently cleans its one journaled resource. The second opt-in',
             '`TestAccAgentTargetTemplateComposition` links an owned target template to an owned agent,',
             'verifies the reference after replacement with two pinned comparisons, and cleans both.',
             'Neither scenario assigns users/accounts or enables a live call path. Registration is not a live pass.', '',
             '## Live prerequisites and scope blockers', '',
             '- Fresh [Graph app-token claims](evidence/graph-permission-claims.json) show the listed broad',
             '  AppCatalog/Sites/Files roles absent from the bound CI identity; a fresh read-only catalog',
             '  request returned HTTP 403. This denial does not qualify catalog success or establish',
             '  every scoped permission alternative for Sites/Files.',
             '- Unified apps and meeting knowledge need delegated credentials and exact service scopes.',
             '- Shifts success needs service-authorized credentials, a WFM system and a mapped team.',
             '- Device success needs real device inventory and delegated gateway access.',
             '- Support/provisioning success needs support credentials; irreversible tenant moves, licensing',
             '  changes and service-wide actions are not made safe by a synthetic success fixture.',
             '- The remaining nested candidates below are real compiled classes but not normal root exports.',
             '  They remain research scope, not approved exclusions or delivered Go services.', '',
             '## Unowned nested candidates', '',
             'Read-only reflection against the verified pinned package found a compiled type for every one',
             'of the fifteen audited nested candidates and none exported by a normal macOS root-module import. The',
             '[assembly hashes, types and parameter declarations](evidence/nested-command-types.json)',
             'are source observations only; they do not settle public/support entitlement or wire behavior.',
             'The hashed manifest declarations make the conditional boundary explicit: five installation/license',
             'commands are commented `#preview` in the root manifest; four conversation-role commands',
             'are exported by preview/OCE manifests, and six remaining settings commands by OCE only.',
             'Four installation commands now have in-progress implementation ownership and appear below',
             'under app_installation; their preview parameter sets remain unknown. This does not authorize',
             'excluding the remaining candidates or assuming the ordinary lab identity can exercise them.', '',
             '| Candidate | Research owner | Current disposition |', '| --- | --- | --- |']
    for row in nested:
        lines.append(f"| `{row['name']}` | `{row['research_owner']}` | Unowned; not implemented or qualified |")
    lines.append('')
    for family, notes in FAMILIES.items():
        lines += [f'## {family}', '', f'Contract and live limitations: [{notes}]({notes}).', '']
        for row in (row for row in rows if row['resource'] == family):
            lines += [f"### {row['name']}", '', f"Owner: `{row['implementation_owner']}`. "
                      f"Implementation: `{row['implementation']['status']}`; "
                      f"contract: `{row['contract']['status']}`; live: `{row['live']['status']}`.", '']
            if row['exclusion']:
                lines += ['Recorded approved exclusion: ' + row['exclusion'], '']
            lines += ['| Parameter set | Go symbols | Contract test mappings | Live test mappings |',
                      '| --- | --- | --- | --- |']
            for group in row['parameter_sets']:
                cells = [f"`{group['name']}`"]
                for field in ('symbols', 'contract_tests', 'live_tests'):
                    cells.append('<br>'.join(reference(v) for v in group[field]) or 'None recorded')
                lines.append('| ' + ' | '.join(cells) + ' |')
            if not row['parameter_sets']:
                symbols = '<br>'.join(reference(v) for v in row['implementation']['symbols']) or 'None recorded'
                lines.append(f'| No observed set; conditional scope requires evidence | {symbols} | No set mapping | No set mapping |')
            lines += ['', 'Recorded live limitation: ' + row['live'].get('reason', 'None recorded.'), '']
    return '\n'.join(lines)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    report = render(json.loads((ROOT / 'contracts/coverage/8.1.0.json').read_text()))
    if args.check:
        if not DESTINATION.exists() or DESTINATION.read_text() != report:
            parser.exit(1, 'EXT report is stale; run tools/coverage/ext_report.py\n')
        print('EXT command/set and unresolved-candidate inventory matches the ledger')
    else:
        DESTINATION.write_text(report)


if __name__ == '__main__':
    main()
