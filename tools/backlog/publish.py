"""Publish the explicitly approved preview, with resumable identity-safe writes."""
import argparse
import datetime
import hashlib
import json
import pathlib
import re
import subprocess
import time
from urllib.parse import quote

from validate import validate

ROOT = pathlib.Path(__file__).resolve().parents[2]
OUT = ROOT / 'docs/implementation/backlog'
STATE = OUT / 'publication'
APPROVED_HASH = '9c04a10526600ab33ab9abd484057c665ac420e3823eff5add416776d516f78f'


def save(path, data):
    tmp = path.with_suffix(path.suffix + '.tmp')
    tmp.write_text(json.dumps(data, indent=2) + '\n', encoding="utf-8", newline="\n")
    tmp.replace(path)


def journal(event, **data):
    with (STATE/'journal.jsonl').open('a') as stream:
        stream.write(json.dumps(dict(time=datetime.datetime.now(datetime.timezone.utc).isoformat(),event=event,**data))+'\n')


class API:
    def __init__(self, repo):
        self.base = 'repos/' + repo
        self.last_write = 0

    def call(self, path, method='GET', data=None, paginate=False):
        if method != 'GET':
            time.sleep(max(0, 1.05 - (time.monotonic()-self.last_write)))
        args = ['gh','api', self.base+'/'+path, '--method',method,'-H','Accept: application/vnd.github+json']
        if paginate:
            args += ['--paginate','--slurp']
        if data is not None:
            args += ['--input','-']
        journal('attempt',method=method,path=path)
        result = subprocess.run(args,input=json.dumps(data) if data is not None else None,text=True,capture_output=True)
        if method != 'GET':
            self.last_write = time.monotonic()
        if result.returncode:
            journal('request_failed',method=method,path=path,error=result.stderr.strip())
            raise RuntimeError(result.stderr.strip() + '\n' + result.stdout[:1500])
        value = json.loads(result.stdout) if result.stdout.strip() else None
        journal('response_received',method=method,path=path)
        if paginate:
            return [item for page in value for item in page]
        return value

    def listing(self, path):
        return self.call(path+('&' if '?' in path else '?')+'per_page=100',paginate=True)


def identity_map(api, namespace):
    found = {}
    pattern = re.compile(r'^<!-- weave-backlog:'+re.escape(namespace)+r':v1:([^ ]+) -->')
    for issue in api.listing('issues?state=all'):
        if 'pull_request' in issue:
            continue
        match = pattern.match(issue.get('body') or '')
        if match:
            key=match[1]
            if key in found:
                raise ValueError('Duplicate remote identity: '+key)
            found[key]=issue
    return found


def main():
    parser=argparse.ArgumentParser()
    parser.add_argument('--stage',choices=['metadata','create','bodies','relationships','audit','all'],default='all')
    args=parser.parse_args()
    manifest_bytes=(OUT/'manifest.json').read_bytes()
    if hashlib.sha256(manifest_bytes).hexdigest()!=APPROVED_HASH:
        raise ValueError('The preview differs from the approved manifest')
    manifest=json.loads(manifest_bytes)
    validate(manifest,json.loads((OUT/'coverage.json').read_text(encoding="utf-8")))
    STATE.mkdir(exist_ok=True)
    save(STATE/'approval.json',dict(manifest_sha256=APPROVED_HASH,authorization='User: approved. raise the tickets on gh',repository=manifest['repository']))
    api=API(manifest['repository'])
    desired={i['id']:i for i in manifest['issues']}
    mapping=identity_map(api,manifest['namespace'])
    if set(mapping)-set(desired):
        raise ValueError('Unexpected managed identities exist remotely')

    def persist():
        save(STATE/'mapping.json',{k:{f:v[f] for f in ['number','id','html_url','state','updated_at']} for k,v in mapping.items()})
    persist()
    stages=['metadata','create','bodies','relationships','audit'] if args.stage=='all' else [args.stage]
    for stage in stages:
        print('STAGE '+stage,flush=True)
        if stage=='metadata':
            labels={l['name']:l for l in api.listing('labels')}
            for wanted in manifest['labels']:
                name=wanted['name']
                if name in labels:
                    if any(labels[name][k]!=wanted[k] for k in ['color','description']):
                        raise ValueError('Label conflict: '+name)
                else:
                    try:
                        api.call('labels','POST',wanted)
                    except RuntimeError:
                        labels={l['name']:l for l in api.listing('labels')}
                        if name not in labels or any(labels[name][k]!=wanted[k] for k in ['color','description']):
                            raise
                    actual=api.call('labels/'+quote(name,safe=''))
                    if any(actual[k]!=wanted[k] for k in ['name','color','description']):
                        raise ValueError('Label read-back mismatch: '+name)
                    journal('verified_label',name=name)
            milestones={m['title']:m for m in api.listing('milestones?state=all')}
            for wanted in manifest['milestones']:
                name=wanted['title']
                if name in milestones:
                    if milestones[name]['state']!='open' or milestones[name]['description']!=wanted['description']:
                        raise ValueError('Milestone conflict: '+name)
                else:
                    try:
                        created=api.call('milestones','POST',wanted)
                    except RuntimeError:
                        current={m['title']:m for m in api.listing('milestones?state=all')}
                        if name not in current:
                            raise
                        created=current[name]
                    actual=api.call('milestones/'+str(created['number']))
                    if actual['title']!=name or actual['description']!=wanted['description']:
                        raise ValueError('Milestone read-back mismatch')
                    journal('verified_milestone',title=name,number=actual['number'])
            print('Verified metadata: 25 labels, 6 milestones',flush=True)
        elif stage=='create':
            milestones={m['title']:m for m in api.listing('milestones?state=all')}
            order=['INDEX']+sorted(k for k,i in desired.items() if i['type']=='epic')
            pending={k for k,i in desired.items() if i['type']=='task'}
            while pending:
                ready=sorted(k for k in pending if set(desired[k]['blocked_by'])<=set(order))
                if not ready:
                    raise ValueError('Unresolvable creation order')
                order.extend(ready);pending-=set(ready)
            for index,key in enumerate(order,1):
                wanted=desired[key]
                if key not in mapping:
                    payload=dict(title=wanted['title'],body=(ROOT/wanted['body_file']).read_text(encoding="utf-8"),labels=wanted['labels'],milestone=milestones[wanted['milestone']]['number'])
                    try:
                        created=api.call('issues','POST',payload)
                    except RuntimeError:
                        refreshed=identity_map(api,manifest['namespace'])
                        if key not in refreshed:
                            raise
                        created=refreshed[key]
                    actual=api.call('issues/'+str(created['number']))
                    if actual['body']!=payload['body'] or actual['title']!=payload['title']:
                        raise ValueError('Created body differs: '+key)
                    mapping[key]=actual;persist()
                    journal('verified_issue_created',identity=key,number=actual['number'])
                    print(f'CREATED {index}/{len(order)} {key} #{actual["number"]}',flush=True)
            print('All issue identities exist',flush=True)
        elif stage=='bodies':
            if set(mapping)!=set(desired):
                raise ValueError('All identities must exist before resolving links')
            for index,(key,wanted) in enumerate(desired.items(),1):
                original=(ROOT/wanted['body_file']).read_text(encoding="utf-8")
                def substitute(match):
                    target=match[1]
                    if target not in mapping:
                        raise ValueError('Unknown link '+target)
                    return '('+mapping[target]['html_url']+')'
                body=re.sub(r'\(([A-Z]+(?:-[A-Z0-9]+)?)\.md\)',substitute,original)
                actual=api.call('issues/'+str(mapping[key]['number']))
                if actual['body'] not in [original,body]:
                    raise ValueError('Remote body changed outside this publication: '+key)
                if actual['body']!=body:
                    try:
                        api.call('issues/'+str(actual['number']),'PATCH',dict(body=body))
                    except RuntimeError:
                        current=api.call('issues/'+str(actual['number']))
                        if current['body']!=body:
                            raise
                    current=api.call('issues/'+str(actual['number']))
                    if current['body']!=body:
                        raise ValueError('Resolved body read-back mismatch: '+key)
                    mapping[key]=current
                journal('verified_body',identity=key)
                if index%20==0: print(f'RESOLVED {index}/{len(desired)} bodies',flush=True)
            persist()
        elif stage=='relationships':
            if set(mapping)!=set(desired):
                raise ValueError('All issues must exist before relationships')
            for key,wanted in desired.items():
                if wanted['type']!='epic':continue
                number=mapping[key]['number']
                current={i['id'] for i in api.listing(f'issues/{number}/sub_issues')}
                target={mapping[c]['id'] for c in wanted['children']}
                if current-target:raise ValueError('Unexpected child membership '+key)
                for child in wanted['children']:
                    child_id=mapping[child]['id']
                    if child_id not in current:
                        try:api.call(f'issues/{number}/sub_issues','POST',dict(sub_issue_id=child_id))
                        except RuntimeError:
                            checked={i['id'] for i in api.listing(f'issues/{number}/sub_issues')}
                            if child_id not in checked:raise
                checked={i['id'] for i in api.listing(f'issues/{number}/sub_issues')}
                if checked!=target:raise ValueError('Child read-back mismatch '+key)
                journal('verified_children',identity=key,count=len(target))
                print(f'PARENT {key}: {len(target)} children verified',flush=True)
            count=0
            for key,wanted in desired.items():
                if wanted['type']!='task':continue
                number=mapping[key]['number']
                current={i['id'] for i in api.listing(f'issues/{number}/dependencies/blocked_by')}
                target={mapping[d]['id'] for d in wanted['blocked_by']}
                if current-target:raise ValueError('Unexpected blockers '+key)
                for dep in wanted['blocked_by']:
                    if mapping[dep]['id'] not in current:
                        try:api.call(f'issues/{number}/dependencies/blocked_by','POST',dict(issue_id=mapping[dep]['id']))
                        except RuntimeError:
                            checked={i['id'] for i in api.listing(f'issues/{number}/dependencies/blocked_by')}
                            if mapping[dep]['id'] not in checked:raise
                checked={i['id'] for i in api.listing(f'issues/{number}/dependencies/blocked_by')}
                if checked!=target:raise ValueError('Blocker read-back mismatch '+key)
                journal('verified_blockers',identity=key,count=len(target))
                count+=1
                if count%10==0:print(f'BLOCKERS {count}/179 task sets verified',flush=True)
        elif stage=='audit':
            fresh=identity_map(api,manifest['namespace'])
            if set(fresh)!=set(desired):raise ValueError('Identity set mismatch')
            labels={l['name']:l for l in api.listing('labels')}
            milestones={m['title']:m for m in api.listing('milestones?state=all')}
            for wanted in manifest['labels']:
                if any(labels[wanted['name']][k]!=wanted[k] for k in ['color','description']):raise ValueError('Label audit mismatch')
            for wanted in manifest['milestones']:
                if milestones[wanted['title']]['description']!=wanted['description']:raise ValueError('Milestone audit mismatch')
            ready=[];parent_count=0;edge_count=0
            for index,(key,wanted) in enumerate(desired.items(),1):
                actual=fresh[key]
                original=(ROOT/wanted['body_file']).read_text(encoding="utf-8")
                expected=re.sub(r'\(([A-Z]+(?:-[A-Z0-9]+)?)\.md\)',lambda m:'('+fresh[m[1]]['html_url']+')',original)
                if actual['body']!=expected or actual['title']!=wanted['title']:raise ValueError('Content audit mismatch '+key)
                if set(l['name'] for l in actual['labels'])!=set(wanted['labels']):raise ValueError('Issue labels mismatch '+key)
                if actual['milestone']['title']!=wanted['milestone']:raise ValueError('Issue milestone mismatch '+key)
                if wanted['type']=='epic':
                    children=api.listing(f'issues/{actual["number"]}/sub_issues')
                    if {c['id'] for c in children}!={fresh[c]['id'] for c in wanted['children']}:raise ValueError('Audit child set mismatch '+key)
                    parent_count+=len(children)
                if wanted['type']=='task':
                    parent=api.call(f'issues/{actual["number"]}/parent')
                    if parent['id']!=fresh[wanted['parent']]['id']:raise ValueError('Audit parent mismatch '+key)
                    deps=api.listing(f'issues/{actual["number"]}/dependencies/blocked_by')
                    if {d['id'] for d in deps}!={fresh[d]['id'] for d in wanted['blocked_by']}:raise ValueError('Audit blockers mismatch '+key)
                    edge_count+=len(deps)
                    if actual['state']=='open' and all(d['state']=='closed' and d.get('state_reason')=='completed' for d in deps):ready.append(key)
                if index%25==0:print(f'AUDIT {index}/193 issues verified',flush=True)
            report=dict(status='complete',repository=manifest['repository'],approved_manifest_sha256=APPROVED_HASH,issues=len(fresh),epics=13,tasks=179,labels=25,milestones=6,parent_relationships=parent_count,blocker_relationships=edge_count,ready=ready,index_url=fresh['INDEX']['html_url'],epic_urls={k:v['html_url'] for k,v in fresh.items() if k.startswith('EPIC-')},ready_urls={k:fresh[k]['html_url'] for k in ready},verified_at=datetime.datetime.now(datetime.timezone.utc).isoformat())
            save(STATE/'audit.json',report)
            mapping=fresh;persist()
            journal('publication_complete',**report)
            print(json.dumps(report,indent=2),flush=True)


if __name__=='__main__':
    main()
