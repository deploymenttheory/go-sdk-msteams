# Reproducible MicrosoftTeams baseline discovery

Run from the repository root with Python 3.10+ and PowerShell 7:

```sh
python3 tools/baseline/baseline.py --output /tmp/msteams-8.1.0-candidate.json
python3 -m unittest discover -s tools/baseline/tests -v
```

The command fetches the exact Gallery metadata and package, checks both the Gallery SHA512 and the pinned 8.1.0 checksum, extracts into a disposable directory, and invokes a fresh noninteractive PowerShell process. It never installs a module globally, connects to a tenant or modifies the accepted `contracts/powershell/8.1.0` seed.

The candidate contains normalized Gallery/runtime command identities and parameter sets, supplementary manifest candidates, help-command names, configuration metadata, source hashes and limitations. Help prose and Microsoft binaries are not copied into the repository. Runtime/platform variations are genuine observations for BASE-02; the normalizer does not hide different exports. PowerShell/platform provenance is retained separately in `runtime_environment` so later semantic comparison can distinguish environment changes from command changes.

An identical repeat is a no-op. Different content at an existing output path is an error: choose a new candidate path and review the delta. Publication uses exclusive atomic file creation to avoid concurrent replacement. There is no force-overwrite switch.

For offline reproduction, retain the downloaded archive, Gallery XML and raw exporter JSON outside the repository:

```sh
python3 tools/baseline/baseline.py \
  --archive /path/to/module.zip \
  --metadata /path/to/gallery.xml \
  --runtime /path/to/runtime.json \
  --output /tmp/reproduced.json
```

The saved runtime must identify version 8.1.0, unauthenticated discovery and a nonempty, structurally complete command inventory. It is metadata evidence, not proof of wire behaviour or live tenant parity. Failed downloads/imports, empty/incomplete metadata, checksum mismatches and invalid archives fail visibly without publishing an output.

To investigate another version, supply both `--version` and an independently verified `--expected-sha512`; the accepted baseline is not promoted. CI runs offline negative cases on Linux/macOS/Windows and two independent fresh 8.1.0 extractions on Linux. The fresh extraction job may fail on Gallery availability and intentionally reports that failure rather than an empty successful inventory.

Command reconciliation is implemented by `reconcile.py`. See
[the 8.1.0 evidence report](../../contracts/powershell/8.1.0/reconciliation.md)
for conditional exports, nested candidates, unresolved names and reproduction steps.
It preserves all candidates; metadata findings never imply implementation or live coverage.

Current implementation and verification states are maintained separately in the
[capability ledger](../../contracts/coverage/README.md); discovery never marks a
capability implemented or tested.
