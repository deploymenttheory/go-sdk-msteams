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

The certificate/client-secret matrix executes the selections in
[suites.json](suites.json): authentication, resource reads, tenant access-outcome
checks, owned unassigned policy CRUD, dedicated-user/group assignments, team
templates and journaled singleton update/restoration. The registry is the
current source of test names and assertion requirements.
The matrix separates `reads` and `writes` for each authentication method.
Read-only jobs can run concurrently. Every write job shares the
`msteams-lab-tenant-acceptance` concurrency group because concurrent policy writes
can conflict even with distinct resource names. Job-level `queue: max` retains
pending write jobs. The group name is unchanged, so older workflow-level locks
also exclude new write jobs during migration. GitHub documents the
[workflow queue behavior](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency). Selected tests
and child cases must pass unless explicitly excluded as described below. The
aggregate result requires every suite report in both authentication modes and
rejects missing or duplicate reports. Unconfigured skips, missing tests, setup
failures and unsuccessful cleanup fail CI. CRUD also requires
all assertion milestones, two PowerShell comparisons per owned CRUD test and
one cleaned journal entry per test. User assignment requires assign/read-back,
clear/restoration, independent cleanup, three pinned comparisons (before,
assigned and restored), and both policy and user-restoration journal entries. Group assignment requires
all five phase assertions, six pinned comparisons and three cleaned entries
(two owned policies and group restoration).

The remaining acceptance tests are listed in `pending.json` with their missing
fixtures, permissions, capabilities or execution mode. Resource tickets stay open
until their successful operations are qualified. The first supported release
requires complete command and parameter-set qualification.

## Selective acceptance skips

[skips.json](skips.json) maps exact registered `TestAcc` names to public reasons.
These tests stay in their normal suites. The runner excludes them before starting
Go, prints `SKIP <test>: <reason>`, and shows their names and reasons in the CI
summary. Skipped tests count separately from passes and provide no acceptance
evidence. The aggregate checks those exclusions against the same configuration.
A test that calls `t.Skip` without a configured exclusion still fails CI.

The first configured skip is `TestAccSharedCallingRoutingPolicyPopulatedCRUD`:
the lab has no resource account or phone inventory. Shared-calling reads and
unassociated CRUD continue to run. Its populated scenario retains mandatory list
edit/clear checks, four PowerShell comparisons and verified cleanup when enabled.

For a temporary local exclusion, after normal lab preparation:

```sh
python3 tools/acceptance/ci.py run --kind writes --report-directory /private/tmp/acceptance-results \
  --skip-test TestAccRoomVideoTeleConferencingPolicyCRUD --skip-reason 'Lab maintenance.'
```

Repeat `--skip-test` to select multiple exact names. Wildcards, unknown names,
missing reasons and conflicting selections are errors. To run a configured skip,
use `--include-test TestAccSharedCallingRoutingPolicyPopulatedCRUD`; this requires
its real fixture and normal mutation opt-in. It does not turn missing fixtures
into a pass. For prepared local runs, store its private JSON as
`$MSTEAMS_LAB_PRIVATE_DIRECTORY/shared-calling-fixture.json` with mode 0600.
The file has `tenant_id`, `resource_account_id` and `emergency_numbers` (at least
two distinct real E.164 callback numbers). Never use example numbers as lab inventory.

For CI, **Run workflow** provides `skip_tests` (comma-separated names),
`skip_reason`, and `include_tests`. The same inputs reach each auth job and the
aggregate. Local automation can use `MSTEAMS_ACC_SKIP_TESTS`,
`MSTEAMS_ACC_SKIP_REASON` and `MSTEAMS_ACC_INCLUDE_TESTS`. Pass the same selections
to `summarize` when aggregating local reports. Routine PR/main runs use the
checked-in skip configuration. Remove its entry after a fixture is available and
the test has passed in both auth modes.

The optional repository secret `MSTEAMS_LAB_SHARED_CALLING_FIXTURE_JSON` supplies
that phone fixture when available. Preparation checks its tenant, account GUID,
number format and distinctness, writes it privately and removes it at finalization.
No secret is required for a deliberately skipped phone scenario. Number ownership,
availability and calling behavior are verified by the live test and service.

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
- `MSTEAMS_LAB_GROUP_FIXTURE_JSON` (bound dedicated empty security-group identifiers/flags)

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

The dedicated `go-sdk-msteams acceptance policy group` is an empty security group,
with mail disabled and no directory-role assignment capability. The private
receipt is selected locally with `MSTEAMS_LAB_GROUP_FIXTURE_FILE`. CI accepts only
the same tenant, GUID, exact SDK label/nickname and expected boolean flags in its
fixture secret. Group configuration tests verify rank 1 and a policy-name update;
they do not qualify priority reordering across groups or member propagation. Both
authentication modes passed with six pinned read comparisons. Recovery restores
the original empty assignment state before deleting either owned policy; failed
group restoration prevents policy deletion. No tenant-default assignment changes.

## Test results and interrupted-run recovery

The job summary lists failed test names, their HTTP status codes and allowlisted
fixed federation/voicemail assertion reasons. Raw output and dynamic names remain private. Missing
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

Group recovery also requires `MSTEAMS_LAB_GROUP_FIXTURE_FILE` for group journals.
The handler is bound to that group and an empty original snapshot; policy cleanup
refuses deletion while its application-access assignment remains.

## Owned team templates

The template suite runs `expanded` and `json-body` input variants in both
certificate and client-secret modes. Each variant must pass its own create/read,
update/read, list and delete/absence assertions and two pinned PowerShell reads.
The suite requires both named children and two cleaned journal entries. A passing
parent or another variant cannot substitute for a missing/skipped variant.

Cleanup checks selected HTTP404 plus a complete list without the owned template.
Recovery validates its tenant-scope reference and owned display name before a
mutation. Actual create/update receipts, nested detail envelopes, owned/deleted
lists and deleted404 responses supply the permanent unit fixtures. The public
General-channel name/ID and controlled fixture selector remain consistent across
responses. Ordinary CI never rewrites those files. No team, membership, app
installation or public template is created or changed by this suite.

## Singleton messaging and migration configurations

The `singleton-configuration-reads` suite shares the tenant write lock through
`serial: true` and keeps mutation enablement off. The update suite uses the same
lock. This prevents parallel Global reads from observing its temporary changes.
Each update journals the original non-null Boolean before mutation, toggles it,
verifies all typed getter values, compares PowerShell before/after/restoration,
and independently reads back the original state. Messaging updates preserve the
nonempty file block list and every other omitted typed setting. CI requires all
three comparisons, both assertions and one cleaned restore entry per test.

Recovery selects `TestAccRecoverOwnedPolicies` with the bound private journal;
`messaging-configuration` restores only `EnableVideoMessageCaptions`, and
`migration-configuration` restores only `EnableLegacyClientInterop` on Global.
Null or missing originals refuse mutation. This suite temporarily changes Global
in the dedicated lab; it does not qualify user/client propagation.

Mobility CRUD now runs in `owned-policy-crud`, with all five getters compared
against the pinned module before and after a sparse description update. Its
recovery kind is `mobility-policy`. Multi-tenant Global reads/updates run in the
serialized singleton suites; its `multi-tenant-configuration` recovery kind
restores only the journaled CopilotFromHomeTenant string on Global. The recovery
handler accepts only the documented Enabled/Disabled baseline used by the test.
CI reports fixed parity, persistence and restoration failure reasons without
publishing raw responses or resource names.

Network roaming runs in the owned-policy CRUD suite using modern POST/PUT/DELETE.
Notification/feeds runs in `owned-notification-policy`, whose extra explicit-null
Description check and three pinned comparisons are mandatory. Its native create
is evidenced by the pinned unexported compiled New implementation and live tenant
acceptance; public baseline mappings still contain only Get/Set/Remove. Both
policy kinds have bound owned-resource recovery handlers and require GET404 plus
list absence. Their fixtures are permanent service testdata, with no capture store.

Personal-attendant and recording-rollout policies join owned CRUD. The former
updates only recording/transcription controls and retains the observed unprefixed
identity. The latter checks exact sparse request bytes as well as two persisted
ownership values. Both require two pinned comparisons and a cleaned create intent;
recovery kinds are `personal-policy` and `recording-rollout-policy`. These tests
do not grant policies or qualify effective attendant/recording behavior.

Room-video and unassociated shared-calling CRUD run with normal owned policy tests.
Shared-calling's populated scenario is explicitly pending: the lab has no resource
accounts or telephone numbers. Select its private fixture with
MSTEAMS_LAB_SHARED_CALLING_FIXTURE_FILE only after providing tenant_id,
resource_account_id and two actual unassigned emergency callback numbers in
emergency_numbers. The test uses those references without changing the account or
phone assignments, checks list preservation/edit/clear, and sanitizes phone values.
Its empty-list CRUD test does not qualify the populated scenario. Recovery kinds
are room-video-policy and shared-calling-policy.

Shifts policy CRUD runs in `owned-policy-crud` and compares all six public getters
with PowerShell before/after a grace-period update. Its bound recovery kind is
`shifts-policy`. Shifts app Global reads and Boolean update/restoration run in the
serialized singleton suites, with three pinned comparisons and one cleaned
`shifts-app-configuration` restore entry. Permanent response fixtures feed both
service unit suites; additional grants and effective behavior remain separate.

SIP devices Global reads/update-restoration share the tenant lock. Its bound
`sip-devices-configuration` handler restores only the original Enabled/Disabled
BulkSignIn string on Global; CI requires three pinned comparisons and a cleaned
restore entry. Team template permission basic CRUD runs in `owned-policy-crud`.
Its separate `owned-template-permission-populated-policy` suite reads actual
existing team-template IDs, verifies sparse omission, Add/Remove and clear with
four pinned comparisons, and independently deletes its owned policy. Recovery
kind: `template-permission-policy`. Empty and populated workflows write distinct
permanent fixture names. Both new service scenarios run in each auth mode.

Update-management policy runs in `owned-update-management-policy`. CI requires
all CRUD assertions, DateTime persistence, message-list edit/clear, five pinned
comparisons and one cleaned owned-policy entry. Compiled CRUD and modern grant
requests remain distinct. Its fifteen actual response fixtures feed unit replay.
Tenant-abuse's current lab identity receives HTTP403/PolicyRP40301 in both SDK and
pinned PowerShell. Denial parity runs in each auth mode; successful read and
UInt64 update/restoration tests are configured skips with public reasons. They
remain executable with an authorized identity. The update refuses mutation without
a readable original and its bound recovery restores only LowSeatLimit on Global.
No additional role or credential was provisioned for these cases.
