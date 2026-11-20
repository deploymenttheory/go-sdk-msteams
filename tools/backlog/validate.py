"""Audit local bodies, native relationship limits, coverage and source hashes."""
import hashlib
import json
import pathlib
import re
from collections import Counter

ROOT = pathlib.Path(__file__).resolve().parents[2]
OUT = ROOT / 'docs/implementation/backlog'
HEADINGS = ['Outcome','Implementation','Acceptance criteria','Verification','Dependencies','Applicable product rules','Planning','Architecture decision alignment']


def require(condition, message):
    if not condition:
        raise ValueError(message)


def validate(manifest, coverage):
    items = manifest['issues']
    issues = {i['id']: i for i in items}
    require(len(issues) == len(items), 'Duplicate identity')
    labels = {l['name'] for l in manifest['labels']}
    milestones = {m['title']: n for n,m in enumerate(manifest['milestones'])}
    children, blocking, bodies = Counter(), Counter(), {}
    for issue in items:
        key = issue['id']
        require(len(issue['title']) <= 256, 'Title too long: ' + key)
        require(set(issue['labels']) <= labels, 'Unknown label: ' + key)
        require(issue['milestone'] in milestones, 'Unknown milestone: ' + key)
        body = (ROOT / issue['body_file']).read_text(encoding="utf-8")
        bodies[key] = body
        require(len(body) <= 65536, 'Body too long: ' + key)
        require(hashlib.sha256(body.encode()).hexdigest() == issue['body_sha256'], 'Body hash mismatch: ' + key)
        require(body.startswith(f'<!-- weave-backlog:{manifest["namespace"]}:v1:{key} -->\n'), 'Marker mismatch: ' + key)
        require(not re.search(r'\b(?:TODO|TBD)\b|<placeholder>', body), 'Placeholder: ' + key)
        for target in re.findall(r'\]\(([A-Z]+(?:-[A-Z0-9]+)?).md\)', body):
            require(target in issues, f'Broken reference: {key} -> {target}')
        if issue['type'] == 'task':
            require(re.findall(r'^## (.+)$',body,re.M) == HEADINGS, 'Wrong sections: ' + key)
            require('Suggested PR title:' in body, 'Missing PR title: ' + key)
            parent = issue['parent']
            require(parent in issues and issues[parent]['type'] == 'epic', 'Missing parent: ' + key)
            require(key in issues[parent]['children'], 'Parent mismatch: ' + key)
            children[parent] += 1
            dependency_section = body.split('## Dependencies\n',1)[1].split('## Applicable product rules',1)[0].split('Supporting contracts',1)[0]
            direct_refs = re.findall(r'^- \[([^]]+)\]\(',dependency_section,re.M)
            require(set(direct_refs)==set(issue['blocked_by']), 'Body/manifest blockers differ: ' + key)
        if issue['type'] == 'epic':
            checklist = re.findall(r'^- \[ \] \[([^]]+)\]\(',body,re.M)
            require(set(checklist)==set(issue['children']), 'Body/manifest children differ: ' + key)
        require(len(set(issue['blocked_by'])) == len(issue['blocked_by']), 'Duplicate blocker: ' + key)
        require(len(issue['blocked_by']) <= 50, 'Blocked-by limit: ' + key)
        for dep in issue['blocked_by']:
            require(dep in issues and issues[dep]['type'] == 'task', 'Invalid blocker: ' + key)
            require(milestones[issues[dep]['milestone']] <= milestones[issue['milestone']], 'Milestone inversion: ' + key)
            blocking[dep] += 1
    for key,count in children.items():
        require(count <= 100, 'Sub-issue limit: ' + key)
        require(set(issues[key]['children']) == {i['id'] for i in items if i['parent']==key}, 'Incorrect children: ' + key)
    for key,count in blocking.items():
        require(count <= 50, 'Blocking limit: ' + key)
    visiting, visited = set(), set()
    def visit(key):
        require(key not in visiting, 'Dependency cycle: ' + key)
        if key in visited:
            return
        visiting.add(key)
        for dep in issues[key]['blocked_by']:
            visit(dep)
        visiting.remove(key)
        visited.add(key)
    for key in issues:
        visit(key)
    baseline = json.loads((ROOT / 'contracts/powershell/8.1.0/commands.json').read_text(encoding="utf-8"))
    require(Counter(c['command'].casefold() for c in coverage)==Counter(c['name'].casefold() for c in baseline), 'Missing or duplicate command coverage')
    for row in coverage:
        require(row['research'] in issues, 'Missing research owner: ' + row['command'])
        owner = row['implementation']
        if owner is None:
            require(row['disposition']=='excluded-module-telemetry' and row['command'].casefold()=='invoke-csinternalpstelemetry', 'Unapproved exclusion')
        else:
            require(owner in issues, 'Missing implementation owner')
            require('`'+row['command']+'`' in bodies[owner], 'Command missing from issue body: ' + row['command'])
    require(hashlib.sha256((ROOT/'docs/implementation/plan.md').read_bytes()).hexdigest() == manifest['plan_sha256'], 'Source plan changed')
    for label in manifest['labels']:
        require(bool(re.fullmatch('[0-9a-f]{6}',label['color'])), 'Invalid label color')
        require(len(label['description'])<=100, 'Label description too long')
    expected_files = {pathlib.Path(i['body_file']).name for i in items}
    require(expected_files=={p.name for p in (OUT/'issues').glob('*.md')}, 'Orphan or missing issue draft file')
    return dict(issues=len(items),tasks=sum(i['type']=='task' for i in items),epics=sum(i['type']=='epic' for i in items),
                commands=len(coverage),labels=len(labels),milestones=len(milestones),parent_relationships=sum(children.values()),
                blocker_relationships=sum(blocking.values()),ready=[i['id'] for i in items if i['type']=='task' and not i['blocked_by']],
                largest_body=max(len(b) for b in bodies.values()),largest_title=max(len(i['title']) for i in items))


if __name__ == '__main__':
    print(json.dumps({'result':'passed',**validate(json.loads((OUT/'manifest.json').read_text(encoding="utf-8")),json.loads((OUT/'coverage.json').read_text(encoding="utf-8")))},indent=2))
