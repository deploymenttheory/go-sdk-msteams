# Tenant acceptance scenarios

These notes retain resource setup, permission limitations, read-back and cleanup
procedures. They are scenario documentation, not a separate command-coverage or
contract qualification framework. Historical observations apply to their stated
revision and identity; consult the current tests and GitHub backlog for status.

Use the [acceptance guide](../../tools/acceptance/README.md) for credentials,
scenario selection, CI, fixture sanitization and recovery. The
[lifecycle guide](lifecycle-and-recovery.md) describes durable journals and safe
cleanup. API-specific notes are linked from each service's API docs.

Unit regression inputs belong in service `testdata/` directories. Separate
binary/CLR research, command ledgers and generated research reports are retired.
