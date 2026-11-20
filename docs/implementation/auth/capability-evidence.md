# AUTH-05 authentication capability and authorization evidence

The SDK now reports dated command restrictions and distinguishes operation access
failures. Authentication and permissions are separate: obtaining two tokens cannot
establish that any management command is authorized or implemented.

## Evidence layers

`msteams/auth/capabilities_8_1_0.json` retains all 745 reconciliation candidates,
their dispositions and exact parameter-set names. Its version and package SHA are
checked against the immutable 8.1.0 evidence by Go tests. The catalog includes
conditional/research candidates and the approved telemetry exclusion without
promoting their scope or implementation state.

The pinned Connect help requires both resource tokens to have the same mode and
describes managed identity as commercial-only. Source:
`en-US/Microsoft.TeamsCmdlets.PowerShell.Connect.dll-Help.xml`, lines 80 and 490–494;
SHA256 `1fc87439f1bfdbf53e836458d46d2b4144e0fb65975c5bcaa73ec8040f1ee8ff`.
The SDK binds both audience requests to one explicit mode. Tokens remain opaque;
the caller declares their mode, and the services validate their authenticity.
The SDK cannot prove a supplied token's type from the caller's declaration.

The exception overlay is from Microsoft's
[application authentication documentation](https://learn.microsoft.com/en-us/microsoftteams/teams-powershell-application-authentication),
reviewed 2026-11-20. Its 16 rules expand over the pinned names, including wildcard
suffixes and three GCCH/DoD-only application-instance exceptions. These are
**current documentation observations**, not extracted 8.1.0 behavior or live results.
`tools/auth/generate_matrix.py --check` verifies deterministic regeneration.
Changing the overlay requires reviewing the source and its review date; it does
not revise the baseline inventory.

Construct an `auth.CapabilityCatalog` with `auth.NewCapabilityCatalog` and handle
its error. The catalog's `Assess` method returns `documented-unsupported` for a matching application restriction and
`unqualified` otherwise. The latter means unknown, including delegated mode.
It never returns “authorized” or “supported.” Unknown names/modes/cloud labels
return errors. Documentation labels for sovereign clouds do not enable SDK routing;
the implemented providers and transports remain commercial-only. Returned slices
are copies and a catalog can be shared between readers.

## Permissions and roles

Resource implementations must record endpoint-specific delegated scopes, application
permissions, admin consent, directory roles and any tenant feature prerequisites.
The catalog does not infer these from a command name. Unresolved per-operation
requirements remain research work owned by each resource's existing ledger ticket.

| Evidence scope | Prerequisite | Qualification boundary |
|---|---|---|
| Teams module app onboarding | Graph `Organization.Read.All`, admin consent, suitable Teams directory role | Does not grant every Graph operation; avoid configuring a Teams resource application API permission |
| Graph user collection | Application `User.Read.All` or a documented higher permission | Organization permission alone is insufficient |
| Meeting administration | Teams Administrator or Teams Communications Administrator, according to the operation | Role guidance is not a tested endpoint permission mapping |
| Messaging administration | Teams Administrator | Still subject to operation and tenant policy |
| Resource account administration | Applicable Teams administrator role plus User Administrator | Auth-mode/cloud exceptions still apply |
| Delegated sign-in | Consent for the selected resource scopes plus the user's applicable access | Successful login does not bypass consent, role, licensing or conditional access |

Sources: [Graph user permissions](https://learn.microsoft.com/en-us/graph/api/user-list?view=graph-rest-1.0),
[Teams roles](https://learn.microsoft.com/en-us/microsoftteams/using-admin-roles),
[Connect delegated-token setup](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/connect-microsoftteams?view=teams-ps).
Non-Cs and administrative-unit permission requirements must be selected for actual
operations; the module's broad onboarding list is not an SDK blanket grant request.

## Failure contract

| Result | Go contract | Meaning |
|---|---|---|
| HTTP 401 | `errors.Is(err, auth.ErrUnauthenticated)` | Resource rejected authentication |
| HTTP 403 | `errors.Is(err, auth.ErrForbidden)` | Access denied; cause may be role, consent, policy or licensing |
| Evidenced operation/mode rejection | `auth.ErrUnsupportedOperationMode` | Resource implementation may wrap this only with a known operation contract |
| Evidenced missing tenant capability | `auth.ErrTenantCapabilityUnavailable` | Resource implementation may wrap this only after explicit capability evidence |

`errors.As(err, *client.APIError)` still exposes HTTP status and sanitized structured
diagnostics for explicit inspection. API error text remains redacted. Neither a
generic 403/404 nor an error-message substring proves a missing role, unsupported
mode or missing tenant feature. The last two sentinels are available to resource
implementations; this foundation does not invent service-specific mappings.
See [Graph status semantics](https://learn.microsoft.com/en-us/graph/errors).

## Verification map

| Contract | Test/evidence |
|---|---|
| Complete pinned union, dispositions, parameter sets, package identity | `TestCapabilityMatrixCoversPinnedInventory` |
| Every current restriction, wildcard suffix and sovereign-only exception | `TestEveryDocumentedRestrictionAndCloud` checks every candidate across four clouds and both modes |
| Same-type dual tokens, successful acquisition never authorizes a command, copy isolation | `TestCapabilityReportingDoesNotInferAuthorization` |
| 401/403 distinction, missing-role fixture, misleading response text, 404 ambiguity | `TestOperationAccessFailuresRemainDistinct` for both backend profiles |
| Live secret/certificate, both audiences and authorized Graph read | [Initial application qualification](live-application-validation.md) |
| Live permitted organization read, forbidden user read and invalid-token rejection | `TestApplicationAccessLab`, explicitly enabled against the dedicated app |

Missing-role and unavailable-feature cases are contract fixtures until resource
qualification can prove the actual prerequisite. No directory roles or additional
consent are granted by these tests. AUTH-05 does not complete command implementations
or change the release-readiness ledger.

## Additional live qualification

On 2026-11-20, `TestApplicationAccessLab` passed using the dedicated app's existing
`Organization.Read.All` consent and no directory role. Both secret and PEM certificate
flows acquired Graph and Teams audience tokens; Graph organization reads returned 200
and minimal user-list reads returned 403, correctly matching `ErrForbidden`. A deliberately
invalid supplied Graph token returned 401, correctly matching `ErrUnauthenticated`.
No token, identifier or response body was logged. No role/consent/tenant configuration
was changed. This proves an operation authorization boundary, not a missing Teams role
or the app-only exceptions' live behavior. The tested implementation revision and
sanitized local report are recorded with the private lab setup.

Repeat only for the deliberately limited lab app:

```sh
MSTEAMS_LIVE_ACCESS=1 MSTEAMS_LAB_IDENTITY_FILE=/private/path/lab-identity.json \
  go test -count=1 -v -timeout 5m ./tests/integration/auth -run '^TestApplicationAccessLab$'
```

Ordinary CI skips this test before reading files or acquiring credentials.
