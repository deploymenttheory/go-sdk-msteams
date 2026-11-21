# Group-first team creation and migration

`TeamManagement.Create` implements the New-Team/CreateTeam composition.
`MigrateGroup` implements the separate MigrateGroup parameter set. Native
`Teams.Create` remains the asynchronous POST /teams API, and native
`Teams.CreateFromGroup` remains a single PUT with no hidden propagation wait.

Create resolves an explicit owner ID/UPN or uses /me when Owner is empty and the
client is configured for delegated authentication. Application clients require an
explicit owner; there is no authentication fallback. UUID selection performs no
directory read and does not prove existence or authority. Both initial owner and
member references bind that user through the selected Graph cloud.

The request creates a Unified group with mail enabled, security disabled and
Private visibility by default. Explicit description, classification and empty
strings are retained. Visibility accepts the pinned Public/Private/HiddenMembership
values, ignoring input case. Pinned education presets select the captured flags,
site-app roles and education extension; they do not prove that the tenant supports
education provisioning. EDU_Class forces HiddenMembership and rejects explicitly
supplied nonblank visibility. These group presets are distinct from native Teams
template IDs. Original captures are unchanged.

A supplied mail nickname is used directly, without normalization, suffixing or
preflight. Omission generates `msteams_` plus six cryptographically random hex
characters. Each candidate checks a Unified-group alias query through all pages
before accepting absence. Up to ten candidates and sixteen pages per lookup are
allowed. Denied, malformed or incomplete reads cannot permit a group write. The
query is not an atomic reservation; a later POST rejection is returned without
blindly creating again.

After group creation, both workflows read group availability, PUT the group/team
settings with all five containers, then read paired group/team state. MigrateGroup
accepts only its group ID and SDK wait controls, without creation-only settings,
owner or template inputs. It never infers ownership of the supplied group.

`CreationOptions` defaults to a five-minute foreground timeout, four attempts
per propagation stage (initial plus three retries), and ten seconds between
received Graph 404s. Retry-After can increase the delay. No other response or
transport failure triggers this retry. Caller cancellation and client close
apply to the foreground workflow. The pinned five-attempt/20-second behavior is
retained as source characterization rather than imposed on every native call.
The [current create-from-group contract](https://learn.microsoft.com/en-us/graph/api/team-put-teams?view=graph-rest-1.0)
describes replication-related 404s and a successful 201 team response. It also
states that SharePoint provisioning is a background process whose completion
cannot be determined. Paired team readback does not qualify site provisioning.

`CreationResult` retains the group ID, selected alias/owner, creation proof,
teamification snapshot and ordered stages. `Value` requires successful paired
readback. `CreatedGroupOwned` records origin, even after cleanup; it does not mean
the group still exists. A supplied group is never automatically deleted. A
successful teamification is retained when later enrichment fails.

With RetainCreatedGroup, an owned group's failed workflow leaves it untouched.
Otherwise, compensation can delete only the group created successfully by this
call. Before any teamification attempt, that new group can be cleaned directly.
After an attempt, a positive team read retains it. A received rejection plus a
subsequent team 404 permits compensation; this is an SDK lifecycle choice, not
an atomic absence guarantee. Transport loss, 5xx, unexpected success or malformed
creation responses remain ambiguous. A lone 404 after such an outcome never
permits deletion. Unknown group IDs remain unresolved and retain the alias and
creation response for private reconciliation; no identity is fabricated.

Cleanup runs under a separate bounded context (thirty seconds by default), which
may continue after caller cancellation. A closed client cannot be bypassed.
PrimaryError/PrimaryResponse and Cleanup.Error/response remain separate, and the
returned error joins primary and cleanup failures. Local validation errors may
return before lifecycle diagnostics are populated. Failed or ambiguous deletes
are not replayed or reported as successful. Keep recovery metadata privately
with the tenant/cloud/identity binding; default formatting is redacted.

The SDK does not create directory users or grant permissions. Access must cover
each selected directory/group/team operation, including cleanup when requested.
Current [group-create permissions](https://learn.microsoft.com/en-us/graph/api/group-post-groups?view=graph-rest-1.0)
also cover reading user objects used in initial bindings. Native application and
delegated routing tests are separate from PowerShell cmdlet eligibility; the
pinned module's documented app-auth exclusions are not relaxed by these tests.

Primary API documentation reviewed 2026-11-21. Source-derived SDK tests emit
three revision-bound creation/migration cases. Additional tests check actual
education captures, both modes/seven routing profiles, owner resolution,
generated aliases, deadlines, lost responses and reconciliation/cleanup. All
report `coverage_claim: false` and `live: not-run`; successful journalled tenant
qualification remains required.
