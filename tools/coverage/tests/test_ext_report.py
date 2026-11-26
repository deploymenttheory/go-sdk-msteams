"""The generated inventory must be independent of a Windows locale default."""
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

SPEC = importlib.util.spec_from_file_location('ext_report', Path(__file__).parents[1] / 'ext_report.py')
REPORT = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(REPORT)


class ExtReportEncodingTests(unittest.TestCase):
    def test_generation_and_check_use_utf8_under_cp1252_default(self):
        # Exercise both command modes with a valid, empty family inventory;
        # the rendered explanatory text includes the actual Shifts arrow.
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            ledger = root / 'contracts/coverage/8.1.0.json'
            ledger.parent.mkdir(parents=True)
            ledger.write_text(json.dumps({'records': []}), encoding='utf-8')
            destination = root / 'inventory.md'
            read_text, write_text = Path.read_text, Path.write_text

            def windows_read(path, *args, **kwargs):
                kwargs.setdefault('encoding', 'cp1252')
                return read_text(path, *args, **kwargs)

            def windows_write(path, data, *args, **kwargs):
                kwargs.setdefault('encoding', 'cp1252')
                kwargs.setdefault('newline', '\r\n')
                return write_text(path, data, *args, **kwargs)

            with patch.object(REPORT, 'ROOT', root), patch.object(REPORT, 'DESTINATION', destination), \
                    patch.dict(REPORT.FAMILIES, {}, clear=True), patch.object(Path, 'read_text', windows_read), \
                    patch.object(Path, 'write_text', windows_write):
                with patch('sys.argv', ['ext_report.py']):
                    REPORT.main()
                raw = destination.read_bytes()
                self.assertIn('→'.encode('utf-8'), raw)
                self.assertNotIn(b'\r\n', raw)
                with patch('sys.argv', ['ext_report.py', '--check']):
                    REPORT.main()
