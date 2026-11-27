"""Download and verify the PowerShell module used by real-tenant comparisons.

This helper only acquires a checksum-pinned package. It does not inspect commands,
produce metadata inventories or perform contract research.
"""
import base64
import hashlib
import pathlib
import stat
import urllib.request
import xml.etree.ElementTree as ET
import zipfile

VERSION = '8.1.0'
PIN = '842LJvX0gC9ZmIqN9YOrKWPqh9hVk+BPanw4okt3+bJ5EmknnKHvO1ioRllgLZG4j+h06c1s9y9bs/xIxPrMPQ=='
GALLERY = 'https://www.powershellgallery.com'
MAX_DOWNLOAD = 64 * 1024 * 1024
MAX_EXPANDED = 256 * 1024 * 1024


class PackageError(ValueError):
    """A download or archive is not the expected safe comparison module."""


def fetch(url, destination, limit=MAX_DOWNLOAD):
    with urllib.request.urlopen(url, timeout=60) as response:
        expected = response.headers.get('Content-Length')
        total = 0
        with destination.open('wb') as stream:
            while chunk := response.read(65536):
                total += len(chunk)
                if total > limit:
                    raise PackageError('Download exceeds size limit')
                stream.write(chunk)
        if not total or (expected is not None and total != int(expected)):
            raise PackageError('Empty or truncated download')


def gallery_metadata(data, version):
    try:
        root = ET.fromstring(data)
    except ET.ParseError as exc:
        raise PackageError('Malformed Gallery metadata') from exc
    ns = '{http://schemas.microsoft.com/ado/2007/08/dataservices}'
    def field(name):
        node = root.find('.//' + ns + name)
        return node.text if node is not None else None
    if field('Id') != 'MicrosoftTeams' or field('Version') != version:
        raise PackageError('Gallery identity/version mismatch')
    if field('PackageHashAlgorithm') != 'SHA512' or not field('PackageHash'):
        raise PackageError('Missing SHA512 metadata')
    return dict(version=version, sha512=field('PackageHash'))


def verify_archive(path, metadata, expected_hash):
    actual = base64.b64encode(hashlib.sha512(path.read_bytes()).digest()).decode()
    if actual != metadata['sha512'] or actual != expected_hash:
        raise PackageError('Package SHA512 differs from metadata or explicit pin')


def extract_archive(path, target):
    try:
        with zipfile.ZipFile(path) as archive:
            entries = archive.infolist()
            if sum(i.file_size for i in entries) > MAX_EXPANDED:
                raise PackageError('Expanded package exceeds size limit')
            seen = set()
            for entry in entries:
                name = entry.filename.replace('\\', '/')
                parts = pathlib.PurePosixPath(name)
                if parts.is_absolute() or '..' in parts.parts or ':' in name or '\x00' in name:
                    raise PackageError('Unsafe archive path')
                if name.casefold() in seen or stat.S_ISLNK(entry.external_attr >> 16):
                    raise PackageError('Duplicate archive path or symlink')
                seen.add(name.casefold())
                destination = target.joinpath(*parts.parts)
                if entry.is_dir():
                    destination.mkdir(parents=True, exist_ok=True)
                else:
                    destination.parent.mkdir(parents=True, exist_ok=True)
                    destination.write_bytes(archive.read(entry))
    except (zipfile.BadZipFile, EOFError) as exc:
        raise PackageError('Invalid or truncated package archive') from exc
    manifest = target / 'MicrosoftTeams.psd1'
    if not manifest.is_file():
        raise PackageError('Root module manifest is missing')
