# Acceptance recovery audit — 2026-11-21

Audited revision: `e913765ca2af234ff2d0a6a314276c2cdf3791a7` (merged PR #272).
This audit records a delivery-workflow failure. Merged resource implementations
have not passed the required tenant acceptance workflow. Passing ordinary CI is
not evidence of tenant CRUD or PowerShell conformance. Keep the resource tickets
open until their acceptance criteria are demonstrated; do not close them by
moving unmet criteria to a later qualification ticket.

## Executed tenant checks

On 2026-11-21, ran 26 explicitly selected tests against the existing dedicated lab
identity: the limited-access preflight and 25 resource read-only tests. The command
used `go test -json -count=1 -timeout=8m` with explicit read-only enablement, the
private identity file and a 60-second per-test timeout. Mutation and PowerShell
oracle flags were disabled. Raw logs and the exact test selection are private.

- Preflight passed its narrow organization-200/users-403 access expectations.
- All 25 selected resource tests failed. Twenty-four tests reported HTTP 403;
  federation's four subtests also failed with HTTP 403. The application-meeting
  test reported a failed Global read without retaining its status, so its cause
  is not established by that result.
- There were 29 failing test events including federation's four child tests,
  25 failing top-level resource tests, one passing preflight, and no skips.
- No resource mutation or PowerShell/Go resource comparison ran. No resource is
  qualified by this run.

Affected resource tickets: #94–#99 and #102–#120. The existing application-access
CRUD (#93), assignment operations (#100), delegated policy packages (#101), and
core workflows were not part of this read-only selection.

The dedicated identity can acquire Graph and Teams tokens, but initially had only
Graph `Organization.Read.All` and no directory role. A fresh bootstrap-authenticated
Graph inventory confirmed that state. The requested Teams Administrator grant was
rejected by automatic approval review. Following explicit user approval of the
exact role, existing principal and tenant scope, the assignment succeeded and was
independently read back. Its assignment/removal receipt is stored in the private
lab directory. No additional Graph application permission was granted. This is
an authorization prerequisite, not evidence that an SDK wire contract is correct.

### Read-only rerun after approved authorization

Reran the same 25 resource tests with fresh certificate clients and the pinned
8.1.0 PowerShell oracle enabled. Twenty resource tests passed; five failed:

| Resource | Observed failure |
|---|---|
| Assignment migration | HTTP 400 on the status list |
| Audio conferencing custom prompts | HTTP 400 on the Global read |
| Channels policy | Pinned PowerShell oracle failed after the SDK reads |
| Federation | External-policy resource read failed with HTTP 200; three other federation subtests passed |
| Privacy configuration | HTTP 403 on the Global read |

The preflight separately failed its outdated users-403 expectation: both Graph
reads now return 200 after the authorized directory-role assignment. Its expected
access profile needs an explicit update rather than interpreting the new access
as a resource failure. None of these read-only passes qualifies mutation actions.
The private run records the complete selection, outcomes, oracle enablement and
the tested source revision. Failed cases require diagnosis against the pinned
module; do not reclassify them as accepted or unavailable without evidence.

## Acceptance coverage gaps

The ledger contains 184 CORE/POL command records with implementation symbols and
285 parameter sets. All 184 retain implementation `in-progress` and contract
`metadata-only`; live status is `not-run` for 120 and `unavailable` for 64. None
has a live pass. Only 132 parameter sets had a live-test mapping before this audit;
153 had none. Local builders need executable offline acceptance, rather than an
unnecessary tenant mutation.

1. Most policy suites only read settings. Missing acceptance includes owned
   create/update/delete, non-default persistence, omission/null behavior,
   user/group assignment and revocation, batch outcomes and verified cleanup.
   `tests/integration/lab/files_policy_test.go` explicitly excludes those writes;
   the same limitation occurs across the other policy families.
2. Thirteen mutation commands in POL-37–POL-40 were mapped to read-only live tests.
   These mappings cannot establish mutation coverage and must be removed until
   appropriate tests exist. A getter pass cannot qualify a setter or grant.
3. Assignment/migration/package probes can pass with an empty list and no selected
   resource read. OCE status acceptance only requires a non-nil observation.
   Treat these as discovery probes, not completed operation acceptance.
4. PowerShell oracle execution is optional in current live suites. A CRUD smoke
   test can pass without comparing the pinned 8.1.0 behavior.
5. Shared-channel unshare, targeting writes/import/delete, deployment submission
   and outcomes, active operation transitions and policy actions lack sufficient
   tenant acceptance. Unresolved API contracts remain on their resource tickets.

## Reliability gaps to fix before mutation qualification

- Application/external-access recovery treats one empty list after an uncertain
  create as verified absence. Delayed visibility could leave an orphan marked
  cleaned. An uncertain result must retain an unresolved journal entry.
- Team cleanup returns after DELETE acceptance, allowing the journal to mark
  completion before deletion is independently observed. Verify bounded absence
  inside the cleanup handler before reporting success.
- Application-access/template PowerShell helpers lack the output bounds and
  process completion limits already used by other oracle helpers.
- Several tests require immediate update/deletion visibility. Use bounded
  read-only convergence checks without replaying a mutation.

## Required delivery loop

For each affected resource ticket, complete its missing acceptance cases, obtain
the documented authorization and owned fixtures, and run those cases against the
actual tenant. Require the pinned PowerShell oracle for conformance claims.
Verify non-default read-back, negative/authorization behavior and cleanup.
An absent prerequisite, skip, 403, empty discovery list or disabled oracle must
remain an incomplete acceptance result for the operation concerned.

Record the exact tested revision, authentication mode/cloud, command/parameter-set
mapping, executed assertions, oracle version/comparison result, and independently
verified cleanup in sanitized evidence. Rerun after relevant fixes. Update the
ledger and ticket from those results, then close only when its criteria are met.
Integrated qualification gates remain additional checks; they do not absorb
unfinished acceptance from an implementation ticket. New feature implementation
and merges remain paused while this recovery work is addressed.

Ordinary CI currently cannot enforce this: `auth-lab.yml` is manual, runs only the
authentication package, and resource tests skip unless opted in. A required
acceptance check must verify executed per-operation evidence for the PR revision,
reject skipped selected cases, and distinguish smoke/discovery from CRUD and
PowerShell conformance. A green offline build alone must not authorize a merge.
