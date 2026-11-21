# Tenant acceptance in CI

All Go unit/offline contract tests use `TestUnit`; tenant tests use `TestAcc`.
`Example` and `TestMain` retain their Go toolchain meanings. Ordinary quality CI
runs `go test -run '^(TestUnit|Example)'`. The naming guard also reconciles every
`TestAcc` with a CI suite or a concrete [pending disposition](pending.json).
Adding an unregistered acceptance test fails that guard.

`tenant-acceptance.yml` runs on pull requests from this repository, pushes to main
and manual dispatch. Fork/Dependabot code receives no lab secrets; the aggregate
acceptance result fails if its lab jobs did not execute successfully. The workflow
does not use `pull_request_target` or execute fork code with privileged credentials.
Runs serialize without automatically cancelling an in-flight mutation.

The certificate/client-secret matrix executes the explicit selections in
[suites.json](suites.json): authentication, 25 resource read tests plus preflight,
and owned unassigned application-access-policy CRUD. The runner requires the
selected top-level tests and their child cases to pass. Skips, missing test names,
failed processes, missing credentials and disabled/unreached CRUD oracle checks
fail acceptance. CRUD additionally requires all assertion milestones, two pinned
PowerShell comparisons and a new recovery journal whose entries are cleaned.

The remaining 20 acceptance tests require additional fixtures, permission/mode
qualification, reliability fixes or the separate workload-authentication workflow.
They remain explicit in `pending.json`; a green selection is not full SDK parity
and does not complete their resource tickets. Known failed resource reads remain
failures in CI until resolved or explicitly dispositioned; they are not silently
skipped. The first supported release still requires the complete qualification
and coverage gates.

## Repository secrets

The dedicated lab identity is supplied through these repository secrets:

- `MSTEAMS_LAB_TENANT_ID`
- `MSTEAMS_LAB_CLIENT_ID`
- `MSTEAMS_LAB_CLIENT_SECRET`
- `MSTEAMS_LAB_CERTIFICATE_PEM`
- `MSTEAMS_LAB_PRIVATE_KEY_PEM`
- `MSTEAMS_LAB_RECOVERY_CERTIFICATE_PEM` (public recovery certificate)

Only the prepare step receives the secret values. It writes private mode-0600
files under an isolated runner-temp directory. Subsequent tests receive paths,
explicit enablement and the approved `teams-administrator` access profile. The
exact MicrosoftTeams 8.1.0 Gallery archive is SHA512-verified using the existing
baseline tooling before extraction. PowerShell is an oracle for tests only.

## Evidence and interrupted-run recovery

Uploaded JSON contains checked-out revision, auth method, exact selected test
names, outcomes, assertion flags and cleanup counts. It excludes credential values,
tenant/resource IDs, response bodies, dynamic child-test names and raw Go logs.
The output explicitly makes no full-coverage claim.

The finalization step always attempts to seal only recovery journals into
`recovery.cms` using CMS AES-256-GCM and RSA-OAEP-SHA256. It then removes the raw
identity, logs, certificate/key and journals. Encryption failure retains private
files and fails the job. Unresolved cleanup fails even when encryption succeeds.
Only the encrypted journal bundle is uploaded alongside structural results;
plaintext journals are never test artifacts. Retention is 14 days.

The recovery private key stays outside CI in the user-selected lab directory,
alongside `ci-recovery-certificate.pem`. After downloading a failed run's encrypted
artifact, decrypt it into a private directory outside the checkout:

```sh
umask 077
openssl cms -decrypt -binary -inform DER -in recovery.cms \
  -recip /private/lab/ci-recovery-certificate.pem \
  -inkey /private/lab/ci-recovery-private-key.pem -out recovery.tar
```

Stop the original run before recovery. Inspect/extract the journals privately and
use the matching resource's `TestAcc...Recover` handler with the same tenant,
cloud and client binding. Retain unresolved entries; never infer absence from an
empty discovery list. Finalization covers process interruption while the runner
is available, not loss of the runner filesystem before an artifact can be saved.
