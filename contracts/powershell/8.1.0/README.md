# MicrosoftTeams 8.1.0 discovery evidence

This is metadata discovery, not proof of implemented or live-tested SDK parity.

- `provenance.json`: package checksums, selected source-file hashes, release notes and observation limitations.
- `commands.json`: case-insensitive union of Gallery command tags, lexical manifest candidates and runtime observations.
- `runtime.json`: full observed parameter-set metadata, one command per line. Common PowerShell parameters are retained as evidence; the SDK mapping is a separate decision.

The package SHA512 was verified against Gallery metadata. The downloaded package stays in a temporary directory; no Microsoft binaries or credentials are committed.

Observed exports: 680. Candidate commands: 730. Differences are retained for BASE-02, not dropped. Case-only spelling differences are normalized for identity and the runtime spelling is retained.

To reproduce, obtain the exact package and Gallery metadata, verify its checksum, extract it into a temporary directory, and run:

```sh
pwsh -NoLogo -NoProfile -File scripts/discovery/export-command-metadata.ps1 \
  -ModulePath /path/to/module/MicrosoftTeams.psd1 \
  -OutputPath /path/to/discovery/runtime.json
python3 tools/backlog/capture_baseline.py /path/to/discovery --output /tmp/new-seed-candidate
```

The discovery directory must contain `module.zip`, `runtime.json` and `gallery.json` with `commands`, `sha512` and `release_notes`. The exporter imports the module without signing into a tenant. Repeat normalization from the same inputs is deterministic. Cross-platform/connected discovery remains future BASE-02 work.

For verified acquisition and richer source extraction, use [the baseline workflow](../../../tools/baseline/README.md). Both workflows require a new candidate destination and preserve the accepted seed.
