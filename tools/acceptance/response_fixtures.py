"""Bounded export of existing labkit-sanitized response fixtures.

Callers select fixed service/label rules. This module never reads private lab
configuration, credentials, journals or original HTTP logs.
"""
import hashlib
import json
import os
from pathlib import Path
import re


class ExportError(ValueError):
    """A fixed diagnostic that contains no response or private path."""


def write_json(path, value):
    path.parent.mkdir(parents=True, exist_ok=True)
    fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
    with os.fdopen(fd, 'w') as stream:
        json.dump(value, stream, indent=2)
        stream.write('\n')


MAX_BYTES = 1 << 20


def state(root, rules):
    """Observe only fixed allowlisted public testdata directories, never private lab data."""
    root = root.resolve()
    state = {}
    total = 0
    for service in rules:
        directory = root / 'msteams' / 'teams_admin_api' / service / 'testdata'
        if directory.is_symlink() or directory.resolve() != directory:
            raise ExportError('Response fixture source must not traverse symlinks')
        if not directory.exists():
            continue
        for path in directory.iterdir():
            if path.is_symlink() or not path.is_file() or path.stat().st_size > MAX_BYTES:
                raise ExportError('Response fixture source must be bounded regular files')
            with path.open('rb') as stream:
                body = stream.read(MAX_BYTES + 1)
            if len(body) > MAX_BYTES:
                raise ExportError('Response fixture source exceeds file bound')
            total += len(body)
            if total > 16 * MAX_BYTES:
                raise ExportError('Response fixture source exceeds aggregate bound')
            state[path.relative_to(root).as_posix()] = body
    return state


def check_json(body):
    try:
        value = json.loads(body)
    except (ValueError, UnicodeError, RecursionError):
        raise ExportError('Response fixture is not valid JSON') from None
    if not isinstance(value, (dict, list)):
        raise ExportError('Response fixture must be an object or collection')
    guid = re.compile(r'(?i)[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}')
    email = re.compile(r'[\w.+-]+@[\w.-]+\.[a-zA-Z]{2,}')
    def check(node, key=''):
        normalized = re.sub(r'[^a-z]', '', key.lower())
        if normalized in ('serializedholidayrecords', 'serializedholidayrecord') and node not in (None, '', []):
            raise ExportError('Opaque holiday export content is not a response fixture')
        if isinstance(node, dict):
            for name, child in node.items():
                check(child, name)
        elif isinstance(node, list):
            for child in node:
                check(child, key)
        elif isinstance(node, str):
            if any(part in normalized for part in ('password', 'secret', 'token', 'apikey', 'privatekey')) and node not in ('', 'fixture-secret', 'fixture-token'):
                raise ExportError('Response fixture contains an unsanitized credential field')
            if any(not match.group().lower().startswith('00000000-0000-4000-8000-') for match in guid.finditer(node)) or any(not match.group().endswith('@example.invalid') for match in email.finditer(node)) or (node.startswith('eyJ') and node.count('.') == 2):
                raise ExportError('Response fixture contains unsanitized private data')
    check(value)


def capture(root, report_directory, before, revision, rules):
    """Export only changed, existing-writer sanitized JSON or empty receipts."""
    after = state(root, rules)
    changed = {name for name in set(before) | set(after) if before.get(name) != after.get(name)}
    outputs = []
    for name in sorted(changed):
        path = Path(name)
        service = path.parts[2]
        if name not in after or path.stem not in rules[service] or path.suffix not in ('.json', '.body'):
            raise ExportError('Unexpected response fixture output')
        body = after[name]
        if path.suffix == '.body':
            if body:
                raise ExportError('Response non-JSON receipt must be empty')
        else:
            check_json(body)
        outputs.append((path, body))
    if report_directory.is_symlink():
        raise ExportError('Response fixture report directory must not be a symlink')
    directory = report_directory.resolve() / 'response-fixtures'
    if directory.is_symlink():
        raise ExportError('Response fixture output must not traverse symlinks')
    directory.mkdir(parents=True, exist_ok=True)
    records = []
    for path, body in outputs:
        destination = directory / path.parts[2] / path.name
        if destination.parent.is_symlink() or destination.is_symlink() or destination.parent.resolve() != directory / path.parts[2]:
            raise ExportError('Response fixture output must not traverse symlinks')
        destination.parent.mkdir(exist_ok=True)
        destination.write_bytes(body)
        records.append({'source': path.as_posix(), 'artifact': destination.relative_to(report_directory.resolve()).as_posix(), 'sha256': hashlib.sha256(body).hexdigest(), 'bytes': len(body)})
    write_json(report_directory / 'response-fixtures.json', {'schema_version': 1, 'revision': revision, 'sanitizer': 'labkit.WriteResponseFixture', 'files': records})
