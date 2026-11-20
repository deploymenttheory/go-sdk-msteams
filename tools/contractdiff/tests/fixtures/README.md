# Historical fixtures

`8.0.0.json.gz` and `8.1.0.json.gz` contain normalized discovery metadata from isolated,
unauthenticated PowerShell imports. Their package SHA256/SHA512, original snapshot digest,
fixture digest and runtime environment are recorded in `provenance.json`. No Microsoft
binaries or tenant captures are distributed. Release-note prose is replaced by its SHA256
digest; all runtime signatures and extracted source metadata remain intact.

Primary package sources:

- [MicrosoftTeams 8.0.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.0.0)
- [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0)

Reproduce each original snapshot with `tools/baseline/baseline.py --version VERSION
--expected-sha512 HASH --output NEW_PATH`, using the corresponding recorded package hash.
The original acquisition used Gallery version metadata plus the verified archive; 8.1.0
retains the previously approved pin. Hash replacement of release notes and deterministic
GZip compression create the checked fixtures. Tests verify their uncompressed digests.

The Statistics claim records the 8.1.0 release note's documented removal of that output
field. The test's before/after boolean projection represents that claim only. A default
session signature does not reveal every output field, and no tenant call established
that behaviour here. Full resource conformance remains CALL/LAB work.
