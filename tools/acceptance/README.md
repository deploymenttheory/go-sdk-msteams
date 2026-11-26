# Tenant acceptance in CI

All Go unit/offline contract tests use `TestUnit`; tenant tests use `TestAcc`.
`Example` and `TestMain` retain their Go toolchain meanings. Ordinary quality CI
runs `go test -run '^(TestUnit|Example)'`. The naming guard also reconciles every
`TestAcc` with exactly one partition step in [suites.json](suites.json) or a
concrete [pending disposition](pending.json), and requires each registered test
to live in its partition's Go package. Adding an unregistered acceptance test
fails that guard.

`tenant-acceptance.yml` runs on pull requests from this repository, pushes to main
and manual dispatch. Fork/Dependabot code receives no lab secrets; the aggregate
acceptance result fails if its lab jobs did not execute successfully. The workflow
does not use `pull_request_target` or execute fork code with privileged credentials.

## Partitions

[suites.json](suites.json) is the source of test names, step order and assertion
requirements. It has two kinds of partition:

- **Auth** (`auth`): one job that loads every credential kind
  (`MSTEAMS_LAB_AUTH_METHOD=all`) and proves client-secret and certificate
  authentication. Only this job receives the certificate and private key.
- **Services** (`services`): one job per `msteams/teams_admin_api/<service>`
  directory, plus `platform` (preflight and service discovery) and `graph`
  (Microsoft Graph team resources, currently pending). Every service job
  authenticates with the client secret only; `ci.py` refuses certificate auth
  for a service. Certificate support is proven once, not per resource.

A service runs its steps in order: reads/list first, then mutating steps
(owned CRUD, update/restore, assignments, populated scenarios). Validation rejects
a read-only step after a mutating one. If a step fails, the service's later
mutating steps do not start and report `not-run`.

Service jobs start after the auth job passes. `ci.py plan` splits services into
two matrices:

- Read-only services run in parallel (at most 8 at a time) without a lock.
- Services with any `mutation` or `serial` step hold the
  `msteams-lab-tenant-acceptance` concurrency group for their whole read-then-write
  sequence, because concurrent policy writes can conflict even with distinct
  resource names. `serial` marks Global reads that must not observe another job's
  temporary singleton changes; those steps never enable mutations. Job-level
  `queue: max` retains pending jobs, and an in-flight mutation is never cancelled.
  The group name is unchanged from the earlier read/write matrix. GitHub
  documents the
  [workflow queue behavior](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency).

Each job uploads its own artifact (`tenant-acceptance-auth` or
`tenant-acceptance-service-<service>`), so a failed service can be re-run on its
own. **Run workflow** also accepts `services` (comma-separated partition names) to
run only those services after the auth job. The aggregate `TestAcc result` job
then expects reports from those services only.

The `ext_integration` partition also enables the existing response-fixture
sanitizer for its two owned compositions. It exports only changed JSON or empty
HTTP receipts from four fixed service/testdata label lists, with a per-file
1 MiB bound, symlink/path checks and an additional private-data guard. The
`response-fixtures.json` manifest binds each file digest to the exact CI revision;
files live under `response-fixtures/` in that service artifact. Other partitions
never enable this writer. These files contain no plaintext journals or private
lab directory contents and do not establish coverage beyond the run's assertions.

Selected tests and child cases must pass unless explicitly excluded as described
below. The aggregate requires exactly one report per auth step (`all`) and per
selected service step (`client-secret`). It rejects missing, duplicate or
unexpected reports. Unconfigured skips, missing tests, setup failures and
unsuccessful cleanup fail CI. CRUD also requires all assertion milestones, two
PowerShell comparisons per owned CRUD test and one cleaned journal entry per test.
User assignment requires assign/read-back, clear/restoration, independent
cleanup, three pinned comparisons (before, assigned and restored), and both
policy and user-restoration journal entries. Group assignment requires all five
phase assertions, six pinned comparisons and three cleaned entries (two owned
policies and group restoration).

To run a partition locally after preparing a private lab directory:

```sh
MSTEAMS_LAB_AUTH_METHOD=all python3 tools/acceptance/ci.py run --auth --report-directory /private/tmp/acceptance-results
python3 tools/acceptance/ci.py run --service teams_call_hold_policy --report-directory /private/tmp/acceptance-results
python3 tools/acceptance/ci.py run --service teams_call_hold_policy --step reads --report-directory /private/tmp/acceptance-results
```

The remaining acceptance tests are listed in `pending.json` with their missing
fixtures, permissions, capabilities or execution mode. Resource tickets stay open
until their successful operations are qualified. The first supported release
requires complete command and parameter-set qualification.

## Selective acceptance skips

[skips.json](skips.json) maps exact registered `TestAcc` names to public reasons.
These tests stay in their normal steps. The runner excludes them before starting
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
python3 tools/acceptance/ci.py run --service teams_room_video_tele_conferencing_policy \
  --report-directory /private/tmp/acceptance-results \
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
`skip_reason`, and `include_tests`. The same inputs reach every partition job and
the aggregate. Local automation can use `MSTEAMS_ACC_SKIP_TESTS`,
`MSTEAMS_ACC_SKIP_REASON` and `MSTEAMS_ACC_INCLUDE_TESTS`. Pass the same selections
to `summarize` when aggregating local reports. Routine PR/main runs use the
checked-in skip configuration. Remove its entry after a fixture is available and
the test has passed in its service partition.

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
Results cover the tests listed in each step.

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

The `team_template` partition's `crud` step runs `expanded` and `json-body` input
variants. Each variant must pass its own create/read,
update/read, list and delete/absence assertions and two pinned PowerShell reads.
The step requires both named children and two cleaned journal entries. A passing
parent or another variant cannot substitute for a missing/skipped variant.

Cleanup checks selected HTTP404 plus a complete list without the owned template.
Recovery validates its tenant-scope reference and owned display name before a
mutation. Actual create/update receipts, nested detail envelopes, owned/deleted
lists and deleted404 responses supply the permanent unit fixtures. The public
General-channel name/ID and controlled fixture selector remain consistent across
responses. Ordinary CI never rewrites those files. No team, membership, app
installation or public template is created or changed by this step.

## Singleton messaging and migration configurations

Each singleton service's `global-reads` step is `serial: true`: the service job
holds the tenant write lock and keeps mutation enablement off for that step. Its
`update-restore` step runs in the same locked job. This prevents parallel Global reads from observing its temporary changes.
Each update journals the original non-null Boolean before mutation, toggles it,
verifies all typed getter values, compares PowerShell before/after/restoration,
and independently reads back the original state. Messaging updates preserve the
nonempty file block list and every other omitted typed setting. CI requires all
three comparisons, both assertions and one cleaned restore entry per test.

Recovery selects `TestAccRecoverOwnedPolicies` with the bound private journal;
`messaging-configuration` restores only `EnableVideoMessageCaptions`, and
`migration-configuration` restores only `EnableLegacyClientInterop` on Global.
Null or missing originals refuse mutation. This step temporarily changes Global
in the dedicated lab; it does not qualify user/client propagation.

Meeting configuration Global reads and the CustomFooterText/HelpURL update and
restoration run as that service's serialized `global-reads` and `update-restore`
steps. Its dedicated read-only oracle, `read-meeting-configuration.ps1`, compares
all 21 pinned getters, including the authentication context list, before the
update, after it and after restoration. CI requires those three comparisons,
preservation of the 19 omitted getters, and one cleaned `meeting-configuration`
restore entry. The handler restores only the journaled Global string pair,
sending an explicit null for a null original, and refuses other targets.

BYOD and desks policy runs Global read/list/filter as its `reads` step and owned
CRUD as its `crud` step, using the dedicated `read-byod-and-desks-policy.ps1`
oracle. CRUD creates `DeviceDataCollection=Disabled` (Global is `Enabled`),
updates it to `Enabled`, requires every other returned field unchanged and makes
two pinned comparisons. Its recovery kind is `byod-and-desks-policy`; grants are
not exercised live.
Tenant (Meeting Migration Service) migration configuration Global reads and the
MeetingMigrationEnabled update/restoration run as the
`tenant_migration_configuration` service's serialized `global-reads` and
`update-restore` steps. Its dedicated read-only oracle,
`read-tenant-migration-configuration.ps1`, compares the pinned getter before,
after and on restoration; CI requires those three comparisons and one cleaned
`tenant-migration-configuration` restore entry. The update only runs from an
enabled original (enabled, then disabled, then re-enabled), and the handler
restores only an enabled journaled Global value.
Meeting broadcast configuration Global reads (default and
`-ExposeSDNConfigurationJsonBlob`) and SupportURL update/restoration run as its
service's serialized `global-reads` and `update-restore` steps. Its dedicated
read-only oracle, `read-meeting-broadcast-configuration.ps1`, reports the SDN
provider token only as presence. CI requires three pinned comparisons and one
cleaned `meeting-broadcast-configuration` restore entry; the handler restores
only the journaled Global SupportURL and AllowSdnProviderForBroadcastMeeting and
never journals, reads for restoration or writes SDN provider values.

Mobility CRUD runs as its service's `crud` step, with all five getters compared
against the pinned module before and after a sparse description update. Its
recovery kind is `mobility-policy`. Multi-tenant Global reads/updates run as
serialized `global-reads` and `update-restore` steps; its `multi-tenant-configuration` recovery kind
restores only the journaled CopilotFromHomeTenant string on Global. The recovery
handler accepts only the documented Enabled/Disabled baseline used by the test.
CI reports fixed parity, persistence and restoration failure reasons without
publishing raw responses or resource names.

Guest meeting configuration Global reads and the AllowMeetNow update and
restoration run as that service's serialized `global-reads` and `update-restore`
steps. Its dedicated read-only oracle, `read-guest-meeting-configuration.ps1`,
compares all eight pinned getters before the update, after it and after
restoration. CI requires those three comparisons, preservation of the seven
omitted getters, and one cleaned `guest-meeting-configuration` restore entry. The
handler restores only the journaled Global Boolean and refuses null originals or
other targets.


Work location detection policy runs Global read/list/filter as its `reads` step
and owned CRUD as its `crud` step, using the dedicated
`read-work-location-detection-policy.ps1` oracle. CRUD creates
`EnableWorkLocationDetection=true` with `UserSettingsDefault=Enabled`, updates only
`EnableWorkLocationDetection` to false, requires `UserSettingsDefault` and every
other returned field unchanged and makes two pinned comparisons. Its recovery
kind is `work-location-detection-policy`; grants are not exercised live and no
user's work-location setting is selected or changed.

AI policy runs Global read/list/filter as its `reads` step and owned CRUD as its
`crud` step, using the dedicated `read-ai-policy.ps1` oracle for all five
getters. CRUD creates a described policy with `EnrollFace=Disabled` (Global is
`Enabled`), then sets `EnrollFace=Enabled` and clears `Description` with explicit
null in one sparse update, requires every other returned field unchanged and
makes two pinned comparisons. Its recovery kind is `ai-policy`; grants and user
enrollment settings are not exercised live.


Network roaming runs its `crud` step using modern POST/PUT/DELETE.
Notification/feeds runs its own `crud` step, whose extra explicit-null
Description check and three pinned comparisons are mandatory. Its native create
is evidenced by the pinned unexported compiled New implementation and live tenant
acceptance; public baseline mappings still contain only Get/Set/Remove. Both
policy kinds have bound owned-resource recovery handlers and require GET404 plus
list absence. Their fixtures are permanent service testdata, with no capture store.

IP phone policy runs Global read/list/filter as its `reads` step and owned CRUD
as its `crud` step, using the dedicated `read-ip-phone-policy.ps1` oracle. CRUD
creates common-area sign-in with hot desking off and a 30-minute idle timeout,
sparsely updates description, home screen and Better Together, requires every
other returned field unchanged and makes two pinned comparisons. Its bound
recovery kind is `ip-phone-policy`; grants are not exercised live.

Personal-attendant and recording-rollout policies each run a `crud` step. The former
updates only recording/transcription controls and retains the observed unprefixed
identity. The latter checks exact sparse request bytes as well as two persisted
ownership values. Both require two pinned comparisons and a cleaned create intent;
recovery kinds are `personal-policy` and `recording-rollout-policy`. These tests
do not grant policies or qualify effective attendant/recording behavior.

Room-video and unassociated shared-calling CRUD run as their services' `crud` steps.
Shared-calling's populated scenario is explicitly pending: the lab has no resource
accounts or telephone numbers. Select its private fixture with
MSTEAMS_LAB_SHARED_CALLING_FIXTURE_FILE only after providing tenant_id,
resource_account_id and two actual unassigned emergency callback numbers in
emergency_numbers. The test uses those references without changing the account or
phone assignments, checks list preservation/edit/clear, and sanitizes phone values.
Its empty-list CRUD test does not qualify the populated scenario. Recovery kinds
are room-video-policy and shared-calling-policy.

Meeting broadcast policy runs Global read/list/filter as its `reads` step and
owned CRUD as its `crud` step, using the dedicated `read-meeting-broadcast-policy.ps1`
oracle. CRUD creates scheduling off with `AlwaysDisabled` recording, sparsely
updates description and attendee visibility, requires every other returned field
unchanged and makes two pinned comparisons. Its bound recovery kind is
`meeting-broadcast-policy`; grants are not exercised live.

Shifts policy CRUD runs as its service's `crud` step and compares all six public getters
with PowerShell before/after a grace-period update. Its bound recovery kind is
`shifts-policy`. Shifts app Global reads and Boolean update/restoration run as
serialized `global-reads` and `update-restore` steps, with three pinned comparisons and one cleaned
`shifts-app-configuration` restore entry. Permanent response fixtures feed both
service unit suites; additional grants and effective behavior remain separate.

Guest calling configuration Global reads and AllowPrivateCalling update/restoration
run as its service's serialized `global-reads` and `update-restore` steps. Its
dedicated read-only oracle, `read-guest-calling-configuration.ps1`, compares the
pinned getter before, after and on restoration; CI requires those three comparisons
and one cleaned `guest-calling-configuration` restore entry. The handler restores
only the journaled Global Boolean and refuses null or foreign targets.
Media logging policy has no public New/Set/Remove, so its writes grant the built-in
`Enabled` instance rather than an owned policy. Its `reads` step compares Global and
Tag:Enabled getters with the dedicated `read-media-logging-policy.ps1` oracle. The
`user-grant` step reuses the user-assignment proof with one journaled restore entry
(no owned policy). The `group-grant` step uses the `media-logging-group-grant` proof:
grant/read-back, removal/absence and cleanup with three pinned comparisons and one
cleaned entry. Recovery kinds are `media-logging-user-assignment` and
`media-logging-group-assignment`, bound to the private user/group fixtures.
GrantToTenant is not run live because it changes every unassigned user.
First-party meeting-template configuration has its own partition. Global reads
run as the serialized `global-reads` step; `update-restore` sets a non-default
Global Description, requires the template list unchanged, and restores the
original (possibly null) Description with three pinned comparisons and one
cleaned `first-party-template-config` restore entry. Template-list writes are
not run live.

Media connectivity policy runs Global read/list/filter as its `reads` step and
owned CRUD as its `crud` step, using the dedicated `read-media-connectivity-policy.ps1`
oracle. CRUD creates `DirectConnection=Disabled`, updates it to `Enabled`, requires
every other returned field unchanged and makes two pinned comparisons. Its bound
recovery kind is `media-connectivity-policy`; grants are not exercised live.

SIP devices Global reads/update-restoration share the tenant lock. Its bound
`sip-devices-configuration` handler restores only the original Enabled/Disabled
BulkSignIn string on Global; CI requires three pinned comparisons and a cleaned
restore entry. Team template permission basic CRUD runs as its `crud` step.
Its separate `populated-crud` step reads actual
existing team-template IDs, verifies sparse omission, Add/Remove and clear with
four pinned comparisons, and independently deletes its owned policy. Recovery
kind: `template-permission-policy`. Empty and populated workflows write distinct
permanent fixture names.

Guest messaging configuration Global reads and AllowMemes/GiphyRatingType
update/restoration run as its service's serialized `global-reads` and
`update-restore` steps. Its dedicated read-only oracle,
`read-guest-messaging-configuration.ps1`, compares all ten pinned getter settings
before, after and on restoration; CI requires those three comparisons and one
cleaned `guest-messaging-configuration` restore entry. The handler restores only
the journaled Global pair and refuses null, empty or foreign targets.

Update-management policy runs as its service's `crud` step. CI requires
all CRUD assertions, DateTime persistence, message-list edit/clear, five pinned
comparisons and one cleaned owned-policy entry. Compiled CRUD and modern grant
requests remain distinct. Its fifteen actual response fixtures feed unit replay.
Tenant-abuse's current lab identity receives HTTP403/PolicyRP40301 in both SDK and
pinned PowerShell. Denial parity runs in the service partition; successful read and
UInt64 update/restoration tests are configured skips with public reasons. They
remain executable with an authorized identity. The update refuses mutation without
a readable original and its bound recovery restores only LowSeatLimit on Global.
No additional role or credential was provisioned for these cases.

VDI and virtual-appointments reads/owned CRUD are registered as the `reads` and
`crud` steps of their service partitions. Each requires two pinned comparisons, exact sparse-field
assertions, non-default read-back and independent deletion, with one cleaned
journal entry. VDI verifies both disable flags survive an optimization-only update;
virtual appointments verifies false-to-true SMS persistence. Permanent sanitized
service testdata supplies unit replay. Grants remain outside the qualified live
CRUD claim. Existing policies and assignments are untouched.

Voice-applications and workload Global reads are their services' `reads` steps.
Their `crud` steps require two pinned comparisons and one cleaned
journal entry per test. Voice applications checks all 34 getters across a
Description-only update; workload preserves all six Boolean settings across a
Description-only PUT. Each request is asserted sparse and each delete requires
independent GET404/list absence. Eighteen actual response fixtures supply units;
Grant persistence is a separate unqualified operation.

## Test layout

Each service partition's acceptance and offline tests live in
`tests/integration/lab/<service>/` (package `<service without underscores>acc`),
so `go test ./tests/integration/lab/teams_call_hold_policy` runs exactly one
service. `platform` holds preflight and service discovery, `graph` the Microsoft
Graph team resources and `recovery` the journal recovery that merges every
service's `RecoveryHandlers`. Helpers shared by several services are in
`tests/integration/lab/internal/labkit`; PowerShell oracles stay in
`tests/integration/lab/oracle`. `check-names` rejects Go files placed directly in
`tests/integration/lab`.

To recover a retained journal after a failed cleanup, with the same lab identity,
mutation opt-in and fixture files described above:

```sh
MSTEAMS_LAB_RECOVERY_FILE=/private/lab/journal.jsonl \
  go test -count=1 -run '^TestAccRecoverOwnedPolicies$' ./tests/integration/lab/recovery
```

CORE-90 extends each owned team-template CRUD variant (expanded and JSON Body)
with an explicit empty Description update. Both variants must retain the updated
short description and General channel, persist the empty description, and compare
it with pinned PowerShell. The same fixtures now set all 18 boolean fields in
Member, Guest, Fun, Messaging and Discovery settings to non-default true, then
explicit false, with native and pinned comparison after each update. The
`template-crud` proof requires all three additional milestones and five comparisons
per variant (ten total), alongside independent post-delete absence and two cleaned
journal entries. Prior four- and six-comparison runs remain evidence for their
original scopes only; they do not qualify the new settings checks.

CORE targeting-policy conformance runs as the `targeting_policy` partition's
serialized `global-reads` step. It requires exactly one Global policy in the
filtered list, matching selected/listed typed fields and one pinned PowerShell
comparison. A configured app denial, empty list or disabled oracle fails this
positive read scenario. The test makes no policy, hierarchy or assignment change;
legacy-flight eligibility and mutation effects remain separate qualification.

CORE private-channel migration conformance runs the existing bound-tenant singleton
status read in `private_channel_migration/reads`. It requires matching tenant
identity and one pinned PowerShell status/details comparison. It starts no
migration or deployment and needs no existing orchestration selector. A denied
read fails this positive scenario; it does not become a successful migration.

The auth partition can register read-only backend conformance separately from its
base provider authentication step. A backend step's `when_service` names its
required service selection; its optional package must resolve inside
`tests/integration`. Private migration uses this scope to compare certificate-connected
PowerShell with both native credential methods, because pinned 8.1.0's supplied-token
mode supports only the Graph/Skype pair. Certificate files remain auth-job inputs. CI runs base authentication and the
selected backend step in separate auth jobs: a failed backend check fails the
aggregate but does not prevent unrelated service jobs from running.
For an explicit prepared auth-only run, select the exact step:

```sh
MSTEAMS_LAB_AUTH_METHOD=all python3 tools/acceptance/ci.py run --auth --step private-migration-backend --report-directory /private/tmp/acceptance-results
```

The ordinary private migration service step still requires a successful bound-tenant
native read; HTTP 401 is a failure. Backend parity is reported in the separate
selected auth step and never inferred from token acquisition.

## Integrated EXT acceptance

The `ext_integration` partition creates one SDK-owned, unassigned feedback policy
through the generic configuration API, reads and sparsely updates it through the
typed feedback service, then replaces it through the generic API and verifies the
typed result. It requires three pinned PowerShell comparisons and one independently
cleaned journal entry. No Global configuration, user/group assignment, licensing
state or service-wide cache is changed. Changes to either participating adapter
select this partition.

The [EXT inventory](../../docs/implementation/ext/coverage-report.md) records all
owned command/parameter-set mappings, current live limitations and the fifteen
unowned nested research candidates. Mappings and expected denials do not establish
full qualification or successful tenant operations.

The same partition's `agent-target-template` step creates an owned IVR tag
template and an owned agent configuration referencing it, verifies the linkage
after rename with two pinned comparisons, and requires two independently cleaned
entries. No resource account or call flow is attached to the agent.
## Integrated CALL acceptance

The `call_integration` partition composes owned schedules and auto attendants.
`TestAccScheduleAutoAttendantComposition` creates a weekly schedule, drafts an
after-hours association, creates an unassigned attendant, verifies cross-service
identities, replaces its greeting and verifies the preserved association. It
runs two pinned PowerShell comparisons and requires both journal entries cleaned
in reverse creation order. It assigns no account, telephone number, user, group
or policy. Changes to either service select this integration partition as well.

The CALL [coverage inventory](../../docs/implementation/calls/coverage-report.md)
names every command/parameter-set mapping and separates those mappings from
actual passing evidence and remaining live prerequisites.
