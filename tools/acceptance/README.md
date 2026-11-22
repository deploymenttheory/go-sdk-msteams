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
Write jobs serialize without automatically cancelling an in-flight mutation;
read-only jobs run in parallel.

The certificate/client-secret matrix executes the explicit selections in
[suites.json](suites.json): authentication, 22 resource read tests, preflight,
service discovery, three tenant access-outcome checks, and owned unassigned
application-access, external-access, calling, call-hold, call-park, channels, Cortana, events and enhanced-encryption policy CRUD. A separate owned-user suite assigns and clears an application-access
policy on the dedicated SDK user.
The matrix separates `reads` and `writes` for each authentication method.
Read-only jobs can run concurrently. Every write job shares the
`msteams-lab-tenant-acceptance` concurrency group because concurrent policy writes
can conflict even with distinct resource names. Job-level `queue: max` retains
pending write jobs. The group name is unchanged, so older workflow-level locks
also exclude new write jobs during migration. GitHub documents the
[workflow queue behavior](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency). Selected tests
and child cases must pass. The aggregate result requires every selected suite
in both authentication modes and rejects missing or duplicate reports. Skips,
missing tests, setup failures and unsuccessful cleanup fail CI. CRUD also requires
all assertion milestones, two PowerShell comparisons per owned CRUD test and
one cleaned journal entry per test. User assignment requires assign/read-back,
clear/restoration, independent cleanup, three pinned comparisons (before,
assigned and restored), and both policy and user-restoration journal entries.

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
- `MSTEAMS_LAB_USER_FIXTURE_JSON` (bound dedicated user identifiers; no password)

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

The dedicated `go-sdk-msteams acceptance user` is an ordinary, unlicensed account.
Its local credentials and receipt remain in the private lab directory. CI receives
only `tenant_id`, `user_id`, `user_principal_name` and `display_name` in the fixture
secret. Preparation rejects a different tenant, unrelated account label, invalid
ID or extra fields. No bootstrap credentials or user password are uploaded.
Local assignment tests select the private receipt with
`MSTEAMS_LAB_USER_FIXTURE_FILE`. The test requires an empty initial application
assignment, journals restoration before grant, verifies the direct source,
clears it and independently deletes its owned policy. Certificate and client-secret
lab runs both passed without assigning a license or administrator role.

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

For application-access, external-access, calling, call-hold, call-park, channels, Cortana, events and enhanced-encryption policy journals, run
`TestAccRecoverOwnedPolicies` with `MSTEAMS_LAB_RECOVERY_FILE` pointing to the
private decrypted journal, the same lab identity and normal mutation opt-in.
It verifies ownership and requires typed GET404 plus list absence after deletion.

User-assignment journals additionally need `MSTEAMS_LAB_USER_FIXTURE_FILE` when
running `TestAccRecoverOwnedPolicies`. The bound handler clears only the configured
SDK user and verifies its original empty direct-assignment state. Missing fixture
configuration leaves the restoration entry unresolved; it never selects another
user automatically.
