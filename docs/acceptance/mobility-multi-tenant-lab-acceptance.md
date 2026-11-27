# Mobility and multi-tenant organization qualification

Owners: #136 (POL-63) and #137 (POL-64). Baseline: MicrosoftTeams 8.1.0.

Both public surfaces select compiled Core cmdlets and the direct PolicyRP provider.
Mobility has Get Identity/Filter, New/Set/Remove Identity and GrantToUser/GrantToGroup/
GrantToTenant plus the parameterless Identity set. Multi-tenant configuration has
Get Identity/Filter and Set Identity. Its CopilotFromHomeTenant type is String in
the pinned public runtime and compiled source, despite the current web reference's
inconsistent Boolean label. The command exports and source, rather than current
web syntax defaults, determine identity/Force bindings.

Primary local sources: TeamsMobilityPolicy and all five implementing cmdlets,
TeamsMultiTenantOrganizationConfiguration and its Get/Set cmdlets, and shared
BaseTpm* / HttpRequestHelper provider methods. Core hash:
68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49;
PolicyRP hash: a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51.
Read oracles validate these pinned binaries before connecting.

On 2026-11-22 the dedicated SDK SPN (Teams Administrator) read/listed Global for
both services and matched every typed getter with pinned PowerShell 8.1.0.
Mobility has five nullable strings. Its owned CRUD test creates a uniquely named,
unassigned policy with a non-default description and an explicit video setting,
verifies persistence, changes only the description, verifies omission, compares
twice with PowerShell, lists and deletes. Independent GET404 and collection
absence verify cleanup. Recovery intent precedes create.

Multi-tenant configuration has one nullable string. Its update test refuses a
missing/unsupported original setting, journals the original Enabled/Disabled
value, toggles it, reads back and compares the pinned getter, restores the original
value, and independently reads it again. Both application authentication methods
passed this scenario. Global read tests share the CI write lock; mutation opt-in
remains disabled for the serial read suite. CI requires three comparisons,
update/restoration assertions and one cleaned restore entry. No organizations,
licenses, privilege changes, policy grants or user/group assignments occur.

The committed service testdata consists exclusively of permanent sanitized live
response bodies. Unit tests load these for routing, headers, typed getters,
complete snapshots, filters and CRUD/update receipts. Deliberately malformed or
empty failing replies test failure/replay behavior separately; they are not
successful response fixtures. Compiling examples and each parameter-set mapping
are included. No separate capture store is created.

Still unqualified: mobility grants and effective client behavior, explicit null
or empty string persistence, multi-tenant meeting/Copilot effects, delegated
authentication, locale-specific filter behavior and other clouds. Successful
HTTP acceptance remains separate from persisted or effective behavior. Coverage
states and whole-ticket closure require their own complete audit.

[Microsoft's Copilot MTO guide](https://learn.microsoft.com/en-us/microsoftteams/copilot-mto)
supplies the documented Enabled/Disabled test values; source setters remain opaque
strings without invented SDK validation.
