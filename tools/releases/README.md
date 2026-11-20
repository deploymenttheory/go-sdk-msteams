# Upstream candidate discovery

DELTA-02 polls the official MicrosoftTeams Gallery feed daily (06:23 UTC) and on
manual dispatch. It selects the latest version newer than 8.1.0 in each of the stable
and preview channels, and separately flags changed package bytes at the accepted
8.1.0 version. It follows bounded Gallery pagination and compares semantic versions,
including numeric prerelease identifiers. Historical releases below the baseline are
not proposed. The observed feed must still contain the accepted baseline.

```sh
python tools/releases/discover.py --output /tmp/new-msteams-candidates
python -m unittest discover -s tools/releases/tests -v
```

The output directory must be new. No-change discovery makes no directory and does
not import a module. Extraction verifies the exact release's SHA512 against both the
release feed and exact-version metadata before using the existing size-limited,
path-checked baseline extractor. The Gallery digest proves consistency with the feed;
it is not a separate publisher-signature attestation. Imports use a fresh PowerShell
process with a filtered environment and no tenant credentials. Module prerelease
labels are included in the observed version, so a preview cannot pass as stable.
A failure produces a nonzero exit and never modifies accepted contracts. No partial
candidate set is published when any acquisition fails; rerun after fixing the failure.

Each candidate contains `candidate.json`, `snapshot.json.gz`, `impact.json` and
`impact.md`. JSON holds every semantic delta and owner/test mapping; Markdown summarizes
the first 100 changes. Release-note prose is hashed, consistently with the checked
[8.1.0 historical oracle](../contractdiff/tests/fixtures/README.md). That complete,
compressed snapshot is the comparison baseline, validated against the accepted package
pin on every run; the compact seed inventory alone is insufficient for semantic diffs.
Environment differences remain explicit warnings. No live behavior is inferred.

## Draft publication and retry behavior

The workflow imports upstream code in a job with read-only repository access. A separate
runner with a fresh checkout receives only the candidate data artifact. It validates
paths, sizes, schemas, identity and baseline pins, then recomputes all reports before
using publication credentials. Report mismatches or extra files fail before mutation.
Publication writes only four data files under `contracts/candidates/<key>/` through the
GitHub Git Data API, with non-force reference updates.

The key contains the normalized version and a full SHA256 of its decoded source SHA512
digest. A republished version with different bytes gets a distinct key. Stable and preview
versions have distinct keys and PRs. One branch and draft PR exist per key; an identical
repeat makes no GitHub writes. Changed observations update that draft. A failed PR creation
can be retried after the branch commit succeeds, without duplicating branches or PRs.
Closed PRs are left closed, ready-for-review PRs are left under human control, and branches
with human commits are rejected rather than overwritten. Concurrent divergent commits
fail visibly; the next run rechecks state. There is no automatic merge or baseline adoption.

Only the upstream repository's main branch can run discovery/publication in Actions.
PRs and forks run offline tests with read-only permissions. Discovery and publication
are serialized per ref; checkout credentials are not persisted. Publication needs only
`contents: write` and `pull-requests: write`, confined to its own job. It uses the job's
`GITHUB_TOKEN` and cannot follow API redirects. Repository Actions settings must permit
creating pull requests. No PAT, tenant secret, cloud credential or OIDC grant is needed.
PRs created or updated by `GITHUB_TOKEN` trigger PR workflows in an approval-required state;
a repository writer must approve those runs and verify the results before adopting a
candidate. This automation does not bypass CI.

For an explicitly authorized local publication with an existing token in the environment:

```sh
python tools/releases/publish.py --input /tmp/new-msteams-candidates \
  --repository deploymenttheory/go-sdk-msteams
```

`GH_TOKEN` or `GITHUB_TOKEN` is read only for this command. Do not put token values in
arguments, reports or repository files. The command refuses a fork repository. To re-run
an intentional human-owned or closed candidate, review its state manually; the scheduler
will not silently reopen it. The scheduled service currently tracks package metadata;
remote server changes without a package release require separately supplied contract
observations and live conformance work.

## Evidence and sources

The test suite mocks Gallery/GitHub boundaries for no-change, pagination, new releases,
republished checksums, previews, malformed feeds, failed verification/import, draft
creation retry, changed observations, human edits, closed/ready PRs, artifact tampering
and fork rejection. A synthetic module tests actual PowerShell preview import; CI has
PowerShell on all three operating systems. Baseline/contractdiff regression suites remain
in place. These are discovery tests, not tenant conformance or implementation coverage.

- [Official Gallery Atom feed](https://www.powershellgallery.com/api/v2/FindPackagesById()?id=%27MicrosoftTeams%27)
- [PowerShell prerelease metadata](https://learn.microsoft.com/en-us/powershell/gallery/concepts/package-manifest-affecting-ui)
- [GitHub draft pull request API](https://docs.github.com/en/rest/pulls/pulls#create-a-pull-request)
- [GitHub workflow permissions](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#permissions)
- [Events triggered by GITHUB_TOKEN](https://docs.github.com/en/actions/how-tos/writing-workflows/choosing-when-your-workflow-runs/triggering-a-workflow#triggering-a-workflow-from-a-workflow)
