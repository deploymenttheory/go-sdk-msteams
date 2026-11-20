"""Render and validate a complete local GitHub backlog preview. No remote writes."""
import hashlib
import json
import pathlib
import re
from collections import defaultdict

from catalog import EPICS, MILESTONES, foundational_tasks
from resource_groups import SPECIAL, classify

ROOT = pathlib.Path(__file__).resolve().parents[2]
OUT = ROOT / 'docs/implementation/backlog'
NAMESPACE = 'msteams-powershell-parity'
SOURCE = 'https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0'
TASKS = {}
BASE = json.loads((ROOT / 'contracts/powershell/8.1.0/commands.json').read_text(encoding="utf-8"))
PROVENANCE = json.loads((ROOT / 'contracts/powershell/8.1.0/provenance.json').read_text(encoding="utf-8"))
RUNTIME = {c['name'].casefold(): c for c in json.loads((ROOT / 'contracts/powershell/8.1.0/runtime.json').read_text(encoding="utf-8"))['commands']}


def ref(key):
    return f'[{key}]({key}.md)'


def add(key, title, milestone, blockers, component, implementation, acceptance, verification, commands=None, **extra):
    if key in TASKS:
        raise ValueError('Duplicate task ' + key)
    TASKS[key] = dict(id=key, title=f'[{key}] {title}', milestone=milestone, blocked_by=blockers,
                      component=component, implementation=implementation, acceptance=acceptance,
                      verification=verification, commands=commands or [], **extra)


foundational_tasks(add)
groups, sessions, exclusions = classify(BASE)
add('AUTH-06', 'Map PowerShell connection and environment helpers to portable SDK sessions', 1,
    ['AUTH-01','BASE-02'], 'sdk',
    'Map every assigned helper to explicit root-client configuration, cloud/region routing or local cache lifecycle. Resolve conditional OCE/region contexts in BASE-02 before reproducing public behaviour. Document portable certificate/browser equivalents and common-parameter mapping. Disconnect clears client-owned state, not globally shared identities; do not pretend it remotely revokes Entra access tokens.',
    'Each assigned command has an explicit SDK equivalent or reviewed internal-context disposition; concurrent clients retain independent tenant/region settings.',
    'Test client isolation, configuration replacement, cache clearing and invalid cloud/region settings. Compare session behaviour to 8.1.0 where observable without tenant writes.', sessions)

# Research batches bound contract discovery to at most roughly 40 command candidates.
# A single cohesive family is never split just to meet a task-count target.
research_for = {}
family_ids = {}
for area in ['CORE', 'POL', 'VOICE', 'CALL', 'EXT']:
    families = sorted((family, cmds) for (a, family), cmds in groups.items() if a == area)
    batches, batch, count = [], [], 0
    for family, cmds in families:
        if batch and count + len(cmds) > 40:
            batches.append(batch)
            batch, count = [], 0
        batch.append((family, cmds))
        count += len(cmds)
    if batch:
        batches.append(batch)
    if len(batches) >= 20:
        raise ValueError('Research IDs overlap resource range')
    for i, batch in enumerate(batches, 1):
        key = f'{area}-{i:02d}'
        family_names = [f for f, c in batch]
        commands = [cmd for f, c in batch for cmd in c]
        for family in family_names:
            research_for[(area, family)] = key
        research_title = 'Establish contracts for ' + ', '.join(family_names)
        if len(research_title) > 220:
            research_title = f'Establish {area.lower()} contracts from {family_names[0]} through {family_names[-1]}'
        add(key, research_title, 0,
            ['BASE-02', 'ARCH-01'], 'research',
            'Own the exact command set below for families ' + ', '.join(family_names) + '. Read the pinned package scripts, command configuration, help and binary metadata plus Microsoft documentation. Record operation routes/methods, authority/audience, discovery, parameters/defaults/validation, null/omission rules, response/error shapes, auth/role restrictions, pagination, async/retry/transfer semantics and local-only builders. Use sanitized PowerShell lab observations when available. Produce versioned contracts and representative request/response fixtures under contracts/endpoints, with source path/hash or documentation revision and confidence. Resolve the canonical Go operations and API-family routing within the agreed architecture before downstream implementation starts. Do not infer a wire route from a command name or convert lack of evidence into a fabricated endpoint.',
            'Every assigned candidate has a settled contract or evidenced internal/conditional disposition; implementation prerequisites include all necessary types and special behaviours; unknown contracts keep this task open and block affected implementations.',
            'Check every command and parameter set against the normalized snapshot; review positive, validation, authorization and failure examples. Confirm fixtures are evidence-derived and sanitized. Where the lab is unavailable, cite package/documentation evidence and state that live behaviour remains unobserved.',
            commands, families=family_names)
    for i, (family, commands) in enumerate(families, 20):
        key = f'{area}-{i:02d}'
        family_ids[(area, family)] = key
        early = (area == 'CORE' and family in ['teams','team_members','channels','channel_members','operations']) or (area == 'POL' and family in ['teams_messaging_policy','policy_assignments']) or (area == 'VOICE' and family == 'online_users')
        blockers = [research_for[(area, family)], 'BASE-03', 'HTTP-02']
        verbs = {cmd['name'].split('-')[0] for cmd in commands}
        if verbs & {'Get','Find'}:
            blockers.append('HTTP-03')
        if family in ['teams','channels','operations','policy_assignments','assignment_migration','number_orders','bulk_number_orders','deployment_migration','resource_accounts','shifts_connections','shifts_sync','devices','migration']:
            blockers.append('HTTP-04')
        if family in ['teams','meeting_branding','audio_prompts','app_catalog','shifts_sync','schedules','number_inventory','bulk_number_orders']:
            blockers.append('HTTP-05')
        special = SPECIAL.get(family, 'Preserve all exposed resource fields, identity/scope rules and command-specific validation; implement grant/assignment or local builder behaviour when present in this command set.')
        if family == 'teams_messaging_configuration':
            special += ' Include 8.1.0 FileTypeBlockList and FileTypeCustomizationSetting.'
        if family == 'teams_meeting_policy':
            special += ' Include 8.1.0 KnowledgeGenerationMeetingTypes, TranscriptionWithMeetingAI and MeetingAI.'
        if family == 'teams_meeting_configuration':
            special += ' Include 8.1.0 EnableWorkIQTranscriptAccess.'
        local_only = family == 'attendant_builders'
        if local_only:
            blockers = [research_for[(area, family)], 'BASE-03','FND-02']
        add(key, 'Implement ' + family.replace('_',' ') + ' management parity', 2 if early else 3,
            blockers, 'sdk',
            'Implement the assigned commands using the settled contracts from ' + research_for[(area, family)] + '. Add the planned ' + family + ' resource package in the contract-selected graph_api or teams_admin_api family, constructor/root-client wiring, typed request/response models, context-aware methods, colocated mocks/tests, api_docs.md and compiling examples. Use shared transport controls rather than duplicating authentication/retry loops. Keep Go methods independent of PowerShell release numbers. Map parameter sets and local constructors explicitly; preserve omitted/null/zero/empty distinctions. ' + special + ' Update each command/parameter-set ledger row with Go symbols and verification evidence. Keep unrelated resource families and shared-transport changes outside this task; defects in shared contracts go to their owners.',
            'Every assigned public management command and parameter set is implemented, with typed models and evidence-backed validation; success/authorization/failure tests pass; examples compile and permissions are documented; contract-tested and live-tested status remain distinct. A stub or generic unsupported error does not satisfy parity.',
            'Run go test on the affected resource package and its mocks, compile the corresponding examples, and run the coverage-ledger check. Test exact method/path/query/body/headers against research fixtures; cover invalid identity, missing required arguments, auth denial, malformed/error responses and cancellation. For writes, verify non-default read-back, omission/null semantics and available cleanup; for local builders, prove no HTTP call occurs. Exercise evidenced paging/async/upload/batch behaviour where present. Add opt-in PowerShell/Go acceptance cases; record unavailable live prerequisites without dropping implementation.',
            commands, family=family, contract=research_for[(area, family)])

# Reusable family contract harnesses are real prerequisites: they provide the
# backend fixture runner and ledger assertions consumed by resource tests.
# Separate policy/configuration/compound harnesses keep their contracts distinct
# and keep native GitHub dependency fan-out below the documented 50-link limit.
integration_groups = defaultdict(list)
for pair, key in family_ids.items():
    area, family = pair
    category = ('policy' if family.endswith('_policy') else 'configuration' if family.endswith('_configuration') else 'compound') if area == 'POL' else 'resources'
    integration_groups[(area, category)].append((pair, key))
qualification_gates = []
for area in ['CORE','POL','VOICE','CALL','EXT']:
    categories = sorted(category for a, category in integration_groups if a == area)
    for ordinal, category in enumerate(categories):
        members = integration_groups[(area,category)]
        harness = f'{area}-{15+ordinal:02d}'
        gate = f'{area}-{90+ordinal:02d}'
        names = ', '.join(pair[1] for pair,key in members)
        add(harness, f'Build {area.lower()} {category} contract-test fixtures and service harness', 1,
            ['BASE-03','HTTP-02','HTTP-03',research_for[members[0][0]]], 'verification',
            'Provide the reusable offline fixture runner, backend route/audience assertions, response/error comparators and ledger evidence writer consumed by resource tests for: ' + names + '. Build the initial runner against a representative settled contract; resource tickets add their own exact fixtures. Preserve scope/identity, query semantics, omission/null and authorization errors; do not normalize meaningful ordering or hide unknown fields. Local builders use local-only assertions. This owns test infrastructure, not production resource CRUD.',
            'Each family can register a deterministic case and receive exact wire/evidence comparison; a deliberate route, audience, serialization or error mismatch fails with a useful diff.',
            'Execute representative positive/negative fixture cases without tenant credentials. Verify no HTTP call for local constructors, proper cancellation and no raw secrets in generated evidence.')
        for pair,key in members:
            TASKS[key]['blocked_by'] = [d for d in TASKS[key]['blocked_by'] if d not in ['BASE-03','HTTP-02','HTTP-03']]
            TASKS[key]['blocked_by'].append(harness)
        add(gate, f'Qualify integrated {area.lower()} {category} compatibility coverage', 3,
            [key for pair,key in members], 'verification',
            'Audit and integrate the completed contracts, services, builders and fixtures for: ' + names + '. Verify shared identities/nested types and cross-service operations where applicable, compile examples together, and produce a signed-off coverage/evidence report consumed by QUAL-01. Check authorization failures and recovery across resource boundaries, not merely isolated child test counts. Missing live resources retain contract-tested status.',
            'All assigned family contracts are implemented and tested; integrated identity/type interactions work; the evidence report names every command, test and live limitation.',
            'Run the combined affected package tests, example compilation and coverage checker; compare resource command lists to BASE-03 and execute available acceptance scenarios with cleanup evidence.')
        qualification_gates.append(gate)

# Genuine data/constructor prerequisites between resource families.
cross = {
    ('CORE','team_members'): [('CORE','teams')],
    ('CORE','channels'): [('CORE','teams')],
    ('CORE','channel_members'): [('CORE','channels')],
    ('CORE','shared_channels'): [('CORE','channels')],
    ('CALL','auto_attendants'): [('CALL','attendant_builders'),('CALL','schedules'),('CALL','audio_prompts'),('CALL','resource_accounts')],
    ('CALL','call_queues'): [('CALL','resource_accounts'),('CALL','audio_prompts')],
    ('VOICE','emergency_network'): [('VOICE','emergency_addresses')],
    ('EXT','shifts_sync'): [('EXT','shifts_connections')],
}
for family, dependencies in cross.items():
    TASKS[family_ids[family]]['blocked_by'].extend(family_ids[d] for d in dependencies)
TASKS['LAB-03']['blocked_by'].extend(family_ids[k] for k in [('CORE','team_members'),('CORE','channel_members'),('POL','teams_messaging_policy'),('POL','policy_assignments'),('VOICE','online_users')])
TASKS['QUAL-01']['blocked_by'].extend(qualification_gates)
TASKS['QUAL-01']['blocked_by'].append('AUTH-06')


def ancestors(key):
    result = set()
    for dep in TASKS[key]['blocked_by']:
        result.add(dep)
        result.update(ancestors(dep))
    return result


def validate_graph():
    visiting, visited = set(), set()
    def visit(key):
        if key in visiting:
            raise ValueError('Cycle at ' + key)
        if key in visited:
            return
        visiting.add(key)
        for dep in TASKS[key]['blocked_by']:
            if dep not in TASKS:
                raise ValueError('Missing dependency ' + dep)
            if TASKS[dep]['milestone'] > TASKS[key]['milestone']:
                raise ValueError('Milestone inversion ' + key + ' -> ' + dep)
            visit(dep)
        visiting.remove(key)
        visited.add(key)
    for key in TASKS:
        visit(key)


validate_graph()
# Remove transitive edges while retaining their IDs as supporting context.
for task in TASKS.values():
    original = set(task['blocked_by'])
    redundant = set().union(*(ancestors(dep) for dep in original)) if original else set()
    task['supporting'] = sorted(original & redundant)
    task['blocked_by'] = sorted(original - redundant)
validate_graph()

RULES = ('Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. '
         'Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. '
         'Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. '
         'No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.')
ALIGN = ('Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. '
         'Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. '
         'No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.')


def command_section(commands):
    lines = []
    for cmd in commands:
        observed = cmd['runtime_observed']
        evidence = f"{cmd['parameter_sets']} observed parameter sets" if observed else 'not observed in the unauthenticated macOS import; reconcile before wire implementation'
        lines.append(f"- `{cmd['name']}` — {evidence}.")
    return '\n'.join(lines)


def task_body(task):
    key = task['id']
    area = key.split('-')[0]
    scope = '\n\nAssigned baseline commands:\n\n' + command_section(task['commands']) if task['commands'] else ''
    dependency_text = '\n'.join(f'- {ref(dep)} — provides {TASKS[dep]["title"].split("] ",1)[1].lower()}.' for dep in task['blocked_by']) or 'None; ready to start.'
    if task['supporting']:
        dependency_text += '\n\nSupporting contracts covered transitively by the direct blockers: ' + ', '.join(ref(d) for d in task['supporting']) + '. These are context, not additional native blockers.'
    title = task['title'].split('] ',1)[1]
    kind = 'feat' if task['component'] == 'sdk' else {'research':'docs','documentation':'docs','verification':'test','integration':'test','tooling':'feat','ci':'ci'}[task['component']]
    return f'''<!-- weave-backlog:{NAMESPACE}:v1:{key} -->
Backlog ID: `{key}` · Epic: {ref('EPIC-' + area)} · Product baseline: {ref('INDEX')}

## Outcome

{title}. This provides the {task['component']} contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

{task['implementation']}{scope}

Source: [MicrosoftTeams 8.1.0]({SOURCE}); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `{PROVENANCE['package_sha256']}`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] {task['acceptance']}
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

{task['verification']}

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

{dependency_text}

## Applicable product rules

{RULES}

## Planning

Milestone: {MILESTONES[task['milestone']][0]}.
Priority: {'P1 — required release capability' if task.get('family') else 'P0 — foundation, contract integrity or delivery gate'}.
Component / scope owner: {task['component']}; {area} workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `{kind}({area.lower()}): {title[0].lower() + title[1:]}`

## Architecture decision alignment

{ALIGN}
'''


def render():
    (OUT / 'issues').mkdir(parents=True, exist_ok=True)
    issues = []
    areas = defaultdict(list)
    for key, task in sorted(TASKS.items()):
        area = key.split('-')[0]
        areas[area].append(task)
        body = task_body(task)
        path = OUT / 'issues' / (key + '.md')
        path.write_text(body, encoding="utf-8", newline="\n")
        issues.append({'id':key, 'title':task['title'], 'type':'task', 'parent':'EPIC-' + area,
                       'milestone':MILESTONES[task['milestone']][0],
                       'labels':['type:task','priority:P1' if task.get('family') else 'priority:P0','area:' + area.lower(),'component:' + task['component']],
                       'blocked_by':task['blocked_by'], 'body_file':str(path.relative_to(ROOT)), 'body_sha256':hashlib.sha256(body.encode()).hexdigest()})
    for area, (title, milestone, acceptance) in EPICS.items():
        children = areas[area]
        body = f'''<!-- weave-backlog:{NAMESPACE}:v1:EPIC-{area} -->
Backlog ID: `EPIC-{area}` · Product baseline: {ref('INDEX')}

## Outcome

{title}. {acceptance} The starting SDK is template-only; this epic carries no implementation-complete claim.

## Agreed behavior

{RULES}

## Implementation tickets

''' + '\n'.join(f'- [ ] {ref(t["id"])} — {t["component"]}; {"P1" if t.get("family") else "P0"}; {t["title"].split("] ",1)[1]}.' for t in children) + f'''

## Epic acceptance

- [ ] {acceptance}
- [ ] Integrated evidence covers interactions across children, including authorization, error recovery and applicable cleanup; contract/live evidence is distinguished.
- [ ] Coverage and documentation match the reconciled baseline; the maintainer evaluates these conditions independently of child closure.

## Planning

Milestone: {MILESTONES[milestone][0]}; this is the integrated epic exit. Children may deliver in earlier gates and do not wait for parent closure.
Priority: {'P1 — required release capability' if area in ['CORE','POL','VOICE','CALL','EXT'] else 'P0 — foundation or delivery gate'}.
Component / scope owner: {area} workstream; native child membership must match this checklist.
Source rationale: agreed implementation plan at starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`, pinned 8.1.0 metadata and INDEX. Direct blockers belong to child contracts, not milestone order.

## Architecture decision alignment

{ALIGN}
'''
        path = OUT / 'issues' / ('EPIC-' + area + '.md')
        path.write_text(body, encoding="utf-8", newline="\n")
        issues.append({'id':'EPIC-'+area,'title':f'[Epic {area}] {title}', 'type':'epic', 'parent':None,
                       'milestone':MILESTONES[milestone][0], 'labels':['type:epic','priority:P1' if area in ['CORE','POL','VOICE','CALL','EXT'] else 'priority:P0','area:'+area.lower()],
                       'blocked_by':[], 'children':[t['id'] for t in children], 'body_file':str(path.relative_to(ROOT)), 'body_sha256':hashlib.sha256(body.encode()).hexdigest()})
    coverage = []
    for (area, family), commands in sorted(groups.items()):
        for cmd in commands:
            coverage.append(dict(command=cmd['name'], disposition='management-candidate', family=family, research=research_for[(area,family)], implementation=family_ids[(area,family)], runtime_observed=cmd['runtime_observed']))
    coverage += [dict(command=c['name'],disposition='session-equivalent',family='sessions',research='BASE-02',implementation='AUTH-06',runtime_observed=c['runtime_observed']) for c in sessions]
    coverage += [dict(command=c['name'],disposition='excluded-module-telemetry',family='telemetry',research='BASE-02',implementation=None,runtime_observed=c['runtime_observed']) for c in exclusions]
    coverage.sort(key=lambda x:x['command'].casefold())
    (OUT/'coverage.json').write_text(json.dumps(coverage,indent=2)+'\n', encoding="utf-8", newline="\n")
    (OUT/'coverage.md').write_text('# Command coverage and ownership\n\nCandidate ownership is not proof of runtime availability or implemented parity. BASE-02 resolves conditional/public status; no candidate is silently discarded.\n\n| Command | Disposition | Contract owner | Implementation owner | Runtime observed |\n| --- | --- | --- | --- | --- |\n' + '\n'.join(f'| `{c["command"]}` | {c["disposition"]} | [{c["research"]}](issues/{c["research"]}.md) | '+(f'[{c["implementation"]}](issues/{c["implementation"]}.md)' if c['implementation'] else 'Excluded: module telemetry only')+f' | {c["runtime_observed"]} |' for c in coverage)+'\n', encoding="utf-8", newline="\n")
    labels = [dict(name='type:'+n,color='5319e7',description=d) for n,d in [('backlog','Compatibility baseline index'),('epic','Integrated behaviour and child work'),('task','Bounded implementation or research outcome')]]
    labels += [dict(name='priority:'+n,color=c,description=d) for n,c,d in [('P0','b60205','Foundation, contract integrity or release gate'),('P1','fbca04','Required release capability')]]
    labels += [dict(name='area:'+a.lower(),color='0052cc',description=t) for a,(t,m,x) in EPICS.items()]
    labels += [dict(name='component:'+c,color='0e8a16',description='Owns '+c+' work') for c in sorted({t['component'] for t in TASKS.values()})]
    requirements = [
        ('Pinned baseline and complete management coverage',['EPIC-BASE','QUAL-01']),
        ('Jamf-style native Go client and repository structure',['EPIC-ARCH','EPIC-FND']),
        ('Portable authentication and CI federation',['EPIC-AUTH']),
        ('REST behaviours, resilience and observability',['EPIC-HTTP']),
        ('Teams, channels and membership',['EPIC-CORE']),
        ('Policies, configuration and assignment',['EPIC-POL']),
        ('Voice, numbering, routing and emergency administration',['EPIC-VOICE']),
        ('Queues, attendants and builders',['EPIC-CALL']),
        ('Applications, Shifts and remaining services',['EPIC-EXT']),
        ('Lab qualification and honest unavailable-feature evidence',['EPIC-LAB','QUAL-02']),
        ('8.2.0/new release and server change detection',['EPIC-DELTA']),
        ('Complete first supported release and consumer documentation',['EPIC-QUAL']),
    ]
    plan = (ROOT/'docs/implementation/plan.md').read_text(encoding="utf-8")
    # Inline shared decisions: published tickets remain usable before repository docs land.
    baseline_body = f'<!-- weave-backlog:{NAMESPACE}:v1:INDEX -->\nBacklog ID: `INDEX` · Source: agreed implementation plan; repository start `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`\n\n## Purpose and scope\n\nNative Go SDK with complete MicrosoftTeams 8.1.0 management parity. This backlog implements the agreed plan; SDK work is not yet delivered.\n\n'
    baseline_body += '## Product baseline\n\n' + plan.split('## Product rules\n\n')[1].split('## Architecture decisions')[0]
    baseline_body += '\n## Glossary\n\n- Candidate: command found in a source, pending reconciled public/export classification.\n- Contract-tested: implemented behaviour verified offline against evidence-derived contracts.\n- Live-tested: an implemented scenario observed against a tenant.\n- Baseline: immutable accepted upstream version; a candidate release PR does not change it.\n- Direct blocker: an unavailable contract required to execute a task, independent of parent membership.\n\n'
    baseline_body += '## Agreed architecture decisions\n\n' + plan.split('## Architecture decisions\n\n')[1].split('## Delivery sequence')[0]
    baseline_body += '\n## Delivery map\n\n| Milestone | Exit | Epics |\n| --- | --- | --- |\n' + '\n'.join(f'| {title} | {desc} | '+', '.join(ref('EPIC-'+a) for a,(_,m,_) in EPICS.items() if m==i)+' |' for i,(title,desc) in enumerate(MILESTONES))
    baseline_body += '\n\n## Requirement coverage\n\n| Requirement | Owners |\n| --- | --- |\n' + '\n'.join('| '+r+' | '+', '.join(ref(x) for x in ids)+' |' for r,ids in requirements)
    baseline_body += f'\n\nAll {len(BASE)} candidate commands are enumerated in their research/implementation task bodies. {len(exclusions)} telemetry command is explicitly excluded; {len(sessions)} helpers map to AUTH-06. The initial {PROVENANCE["runtime_count"]} exports are unauthenticated macOS evidence, not the final cross-platform surface. BASE-02 reconciles the other candidates without dropping them. Package SHA256: `{PROVENANCE["package_sha256"]}`.\n'
    baseline_body += '\n## Execution conventions\n\nType, priority, area and component labels are orthogonal. P0 denotes foundation/integrity/release gates; P1 denotes required release capabilities. Native task parents and direct blockers are required; supporting links do not add blocker edges. Milestones have no invented due dates. Research settles unknown wire contracts before dependent implementation. All tasks are initially open; metadata discovery is not SDK delivery. Existing dependency-update PRs are unrelated.\n\nSource links: '+SOURCE+'; https://github.com/deploymenttheory/go-sdk-jamfpro-v2/tree/25714da557956da9e317f725ca3d22e8818ec77e. Baseline snapshots and the local plan must accompany implementation handoff.\n'
    path=OUT/'issues/INDEX.md';path.write_text(baseline_body, encoding="utf-8", newline="\n")
    issues.append(dict(id='INDEX',title='Microsoft Teams SDK: implementation backlog and product baseline',type='backlog',parent=None,milestone=MILESTONES[5][0],labels=['type:backlog','priority:P0'],blocked_by=[],body_file=str(path.relative_to(ROOT)),body_sha256=hashlib.sha256(baseline_body.encode()).hexdigest()))
    manifest = dict(schema_version=1,repository='deploymenttheory/go-sdk-msteams',namespace=NAMESPACE,
                    source_revision='60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb',plan_sha256=hashlib.sha256(plan.encode()).hexdigest(),
                    publication_status='draft-not-approved',existing_feature_issues=[],existing_milestones=[],
                    metadata_policy='Create listed labels and milestones; leave existing standard labels and dependency PRs unchanged.',
                    relationship_capability='Native relationship writes are not yet exercised; publication must verify actual API support and permissions.',
                    labels=labels,milestones=[dict(title=t,description=d) for t,d in MILESTONES],issues=sorted(issues,key=lambda x:x['id']))
    (OUT/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n', encoding="utf-8", newline="\n")
    ready = sorted(k for k,t in TASKS.items() if not t['blocked_by'])
    preview = f'''# Microsoft Teams SDK backlog publication preview

Target: `deploymenttheory/go-sdk-msteams`. Namespace: `{NAMESPACE}`. Status: **draft, not published**.

This review set contains **{len(issues)} issues: one baseline, {len(EPICS)} epics and {len(TASKS)} tasks**, covering {len(groups)} resource families and all {len(BASE)} command candidates. It adds **{len(labels)} labels** and **{len(MILESTONES)} milestones**, with **{sum(len(t['blocked_by']) for t in TASKS.values())} direct blocker edges** and **{len(TASKS)} native parent relationships**. Existing standard labels and dependency PRs are unchanged; no feature issues or milestones existed at inspection. No obsolete work, updates or identity conflicts are proposed.

Read the [agreed plan](../plan.md), [baseline issue](issues/INDEX.md), [command coverage](coverage.md), and [exact publication manifest](manifest.json). Full final draft bodies are linked below; they are not title-only placeholders.

Ready to start: {', '.join('['+k+'](issues/'+k+'.md)' for k in ready)}. Subsequent readiness follows direct blockers, not numbering. Research tickets deliberately block implementations with unsettled wire contracts.

## Publication contract

Approval covers these exact bodies, metadata, parent sets and blockers. Each body begins with a stable identity marker. Publication resolves local `ID.md` issue links to actual GitHub URLs, without changing scope. All other body content is preserved; no source revisions or successful tests are invented. Body SHA256 values in the manifest protect this preview against accidental changes.

Create missing metadata, baseline, epics and tasks in dependency order; then attach native sub-issues/blockers, resolve links and read everything back. Native API permissions are not yet exercised. If native links are unavailable, publication is incomplete even when textual links exist. Retain the issue-ID mapping and journal for recovery; do not duplicate ambiguous writes.

## Labels

All labels below are new orthogonal dimensions; existing `enhancement`, `documentation`, `dependencies` and `github_actions` retain their meanings. No synthetic blocked/status labels or GitHub Project changes are proposed.

| Label | Color | Meaning |
| --- | --- | --- |
''' + '\n'.join(f'| `{l["name"]}` | `#{l["color"]}` | {l["description"]} |' for l in labels)
    preview += '\n\n## Delivery gates\n\n' + '\n'.join(f'- **{t}:** {d}' for t,d in MILESTONES)
    preview += '\n\nNo due dates are proposed. Cross-milestone epics target their integrated exit; their M0 research and earlier implementation children remain independently actionable. DELTA work may start once its prerequisites exist.\n\n## Complete issue bodies and direct blockers\n\n'
    for area in EPICS:
        preview += f'### {area}\n\n[Epic {area}](issues/EPIC-{area}.md)\n\n| Task | Gate | Direct blockers |\n| --- | --- | --- |\n'
        for t in areas[area]:
            preview += f'| [{t["title"]}](issues/{t["id"]}.md) | M{t["milestone"]} | '+(', '.join(t['blocked_by']) or 'None; ready')+' |\n'
        preview += '\n'
    preview += '## Local validation\n\nRun `python3 tools/backlog/build.py` to deterministically render the preview and `python3 tools/backlog/validate.py` to audit bodies, hashes, coverage, parent membership, references, cycles and milestone ordering. These checks verify planning artifacts, not SDK behaviour.\n'
    (OUT/'README.md').write_text(preview, encoding="utf-8", newline="\n")
    print(f'Rendered {len(issues)} issues, {len(TASKS)} tasks, {len(groups)} resource families; ready: {ready}')


if __name__ == '__main__':
    render()
