"""Retain download and archive safety for real-tenant comparison setup."""
import base64
import hashlib
import io
import pathlib
import sys
import tempfile
import unittest
import zipfile
from unittest import mock

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parents[1]))
import powershell


class PowerShellPackageTests(unittest.TestCase):
    def setUp(self):
        self.work = tempfile.TemporaryDirectory()
        self.addCleanup(self.work.cleanup)
        self.root = pathlib.Path(self.work.name)
        self.archive = self.root / 'module.zip'
        with zipfile.ZipFile(self.archive, 'w') as archive:
            archive.writestr('MicrosoftTeams.psd1', "@{ ModuleVersion = '8.1.0' }")
        self.digest = base64.b64encode(hashlib.sha512(self.archive.read_bytes()).digest()).decode()
        self.xml = f'''<entry xmlns:d="http://schemas.microsoft.com/ado/2007/08/dataservices">
        <d:Id>MicrosoftTeams</d:Id><d:Version>8.1.0</d:Version>
        <d:PackageHashAlgorithm>SHA512</d:PackageHashAlgorithm><d:PackageHash>{self.digest}</d:PackageHash>
        </entry>'''.encode()

    def test_verified_package_extracts_without_command_inventory(self):
        metadata = powershell.gallery_metadata(self.xml, '8.1.0')
        powershell.verify_archive(self.archive, metadata, self.digest)
        target = self.root / 'module'
        powershell.extract_archive(self.archive, target)
        self.assertTrue((target / 'MicrosoftTeams.psd1').is_file())
        self.assertEqual(set(metadata), {'version', 'sha512'})

    def test_checksum_requires_both_metadata_and_pin(self):
        metadata = powershell.gallery_metadata(self.xml, '8.1.0')
        with self.assertRaises(powershell.PackageError):
            powershell.verify_archive(self.archive, metadata, 'wrong pin')
        self.archive.write_bytes(self.archive.read_bytes()[:-3])
        with self.assertRaises(powershell.PackageError):
            powershell.verify_archive(self.archive, metadata, self.digest)

    def test_metadata_rejects_wrong_identity_version_and_algorithm(self):
        for old, new in [(b'MicrosoftTeams', b'Other'), (b'8.1.0', b'8.2.0'), (b'SHA512', b'SHA256')]:
            with self.subTest(old=old), self.assertRaises(powershell.PackageError):
                powershell.gallery_metadata(self.xml.replace(old, new), '8.1.0')
        with self.assertRaises(powershell.PackageError):
            powershell.gallery_metadata(b'not xml', '8.1.0')

    def test_archive_rejects_traversal_collisions_symlinks_and_missing_manifest(self):
        for names in [['../outside'], ['C:/outside'], ['a', 'A'], ['other.txt']]:
            with self.subTest(names=names):
                with zipfile.ZipFile(self.archive, 'w') as archive:
                    for name in names:
                        archive.writestr(name, 'fixture')
                with self.assertRaises(powershell.PackageError):
                    powershell.extract_archive(self.archive, self.root / 'extract')
        with zipfile.ZipFile(self.archive, 'w') as archive:
            entry = zipfile.ZipInfo('link')
            entry.create_system = 3
            entry.external_attr = 0o120777 << 16
            archive.writestr(entry, 'outside')
        with self.assertRaises(powershell.PackageError):
            powershell.extract_archive(self.archive, self.root / 'extract')
        self.assertFalse((self.root.parent / 'outside').exists())

    def test_download_rejects_truncation_and_size_overrun(self):
        for length, limit in [('10', 100), ('3', 2)]:
            response = io.BytesIO(b'abc')
            response.headers = {'Content-Length': length}
            with mock.patch.object(powershell.urllib.request, 'urlopen', return_value=response):
                with self.assertRaises(powershell.PackageError):
                    powershell.fetch('https://example.invalid', self.root / 'download', limit)


if __name__ == '__main__':
    unittest.main()
