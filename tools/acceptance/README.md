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
[suites.json](suites.json): authentication, 22 resource read tests, preflight,
service discovery, three tenant access-outcome checks, and owned unassigned
application-access, external-access, calling, call-hold and call-park policy CRUD.
Authentication jobs run one at a time because concurrent writes to the tenant's
policy collection can conflict even when resource names differ. The shared
workflow group uses `queue: max`, so a new PR, main push or release update waits
instead of replacing an existing pending run. GitHub documents the
[workflow queue behavior](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency). Selected tests
and child cases must pass. Skips,
missing tests, setup failures and unsuccessful cleanup fail CI. CRUD also requires
all assertion milestones, two PowerShell comparisons for each selected mutation test and new cleaned journals.

The remaining 23 acceptance tests are listed in `pending.json` with their missing
fixtures, permissions, capabilities or execution mode. Resource tickets stay open
until their successful operations are qualified. The first supported release
requires complete command and parameter-set qualification.

## Repository secrets

The dedicated lab identity is supplied through these repository secrets:

- `MSTEAMS_LAB_TENANT_ID`
- `MSTEAMS_LAB_CLIENT_ID`
- `MSTEAMS_LAB_CLIENT_SECRET`
- `MSTEAMS_LAB_CERTIFICATE_PEM`
- `MSTEAMS_LAB_PRIVATE_KEY_PEM`
- `MSTEAMS_LAB_RECOVERY_CERTIFICATE_PEM` (public recovery certificate)
- `MSTEAMS_LAB_CALL_HOLD_AUDIO_FILE_ID` (dedicated lab audio fixture)

Only the prepare step receives the secret values. It writes private mode-0600
files under an isolated runner-temp directory. Subsequent tests receive paths,
explicit enablement and the approved `teams-administrator` access profile. The
exact MicrosoftTeams 8.1.0 Gallery archive is SHA512-verified using the existing
baseline tooling before extraction. PowerShell is an oracle for tests only.

Call-hold CRUD references the dedicated `go-sdk-msteams-ci-call-hold.wav` fixture
uploaded with `Import-CsOnlineAudioFile -ApplicationId TenantGlobal`. This silent
PCM file remains available for subsequent CI runs; the tests do not assign its
policies to users. Its receipt is stored in the private lab directory. Local
identity JSON uses `call_hold_audio_file_id`; CI writes that field from the
repository secret. The ID is never printed or committed.

If replacing the fixture, first confirm no owned policy references it. Verify
the replacement with `Get-CsOnlineAudioFile -ApplicationId TenantGlobal`, update
the private configuration and repo secret, then remove the old fixture using
`Remove-CsOnlineAudioFile -ApplicationId TenantGlobal -Identity <old-id>`.
[Microsoft's music-on-hold setup](https://learn.microsoft.com/en-us/microsoftteams/music-on-hold)
documents this prerequisite. Streaming/audio playback and policy assignment
remain separate acceptance scenarios.

## Test results and interrupted-run recovery

The job summary lists failed test names and their HTTP status codes. Missing
credentials name the repository secret that needs fixing.

Uploaded JSON contains checked-out revision, auth method, exact selected test
names, outcomes, assertion flags and cleanup counts. It excludes credential values,
tenant/resource IDs, response bodies, dynamic child-test names and raw Go logs.
Results cover the tests listed in each suite.

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

## Service response fixtures

`TestAcc` tests write sanitized HTTP response bodies directly into each service's
`testdata/` directory when `MSTEAMS_UPDATE_TESTDATA=1` is explicitly set locally.
`TestUnit` tests load those committed files to check decoding, raw field retention
and request behavior. Ordinary CI reads the fixtures and never rewrites them.
An empty HTTP body stays empty; error responses retain their actual status.
Refresh fixtures using the relevant tenant test with its normal lab enablement
and credentials. Review the diff for tenant information before committing.

## Tenant access outcomes

The current lab returns PolicyRP 40003 for assignment migration and custom audio
prompts, and 40301 for privacy configuration. The access-outcome tests verify the
specific SDK error, retain its HTTP diagnostics and compare the actual PowerShell
8.1.0 result. Authentication/setup errors and any different service response fail
these tests. Changes in tenant access therefore trigger requalification.

Successful reads and mutations for these resources remain pending on issues #95,
#106 and #102. The original successful-read TestAcc tests remain registered in
`pending.json` until a suitable tenant or identity is available. No resource
coverage state is promoted by a passing denial test.

For application-access, external-access, calling, call-hold and call-park policy journals, run
`TestAccRecoverOwnedPolicies` with `MSTEAMS_LAB_RECOVERY_FILE` pointing to the
private decrypted journal, the same lab identity and normal mutation opt-in.
It verifies ownership and requires typed GET404 plus list absence after deletion.
