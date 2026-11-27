# Teams app setup policies

`Client.TeamsAdminAPI.TeamsAppSetupPolicies` exposes native CRUD, explicit Set
composition and scoped assignments without a PowerShell runtime dependency.

| Method | Baseline command / parameter set |
|---|---|
| GetByIdentity, List | Get-CsTeamsAppSetupPolicy / Identity |
| ListFiltered | Get / Filter; local Identity/Tag wildcard matching |
| CreateByIdentity | New / Identity; edits on fresh lists then POST |
| PatchByIdentity | Set / Identity; one sparse native PATCH |
| UpdateByIdentity | Set / Identity; explicit read/edit/PATCH |
| DeleteByIdentity | Remove / Identity |
| AssignUser, AssignGroup, AssignTenantDefault | Grant / GrantToUser, GrantToGroup, GrantToTenant |
| Assign with missing/mixed scopes | Grant / Identity; local input error |

ScalarSettings uses policy.Set/Null for Description and nullable AllowSideLoading/
AllowUserPinning. False and null are bound values; a zero Field is omitted. No
description maximum or app-ID GUID/catalog-membership validation is invented.
Set needs a setting; Force alone rejects locally. New permits empty settings and
does not add the source's presentation-only Global comparison read. CRUD Force
is the compiled True/False string body value; tenant assignment uses a Boolean.

The six source list fields keep their own element types and comparison rules:

| List | Go element / edit | Ordinal comparison key |
|---|---|---|
| AppPresetList | App / AppEdit | Id |
| AppPresetMeetingList | App / AppEdit | Id |
| PinnedCallingBarApps | App / AppEdit | Id |
| PinnedAppBarApps | PinnedApp / PinEdit | Id and Order |
| PinnedMessageBarApps | PinnedApp / PinEdit | Id and Order |
| AdditionalCustomizationApps | CustomizationApp / CustomizationEdit | Id and AdditionalCustomizationId |

NewApp, NewPinnedApp and NewCustomizationApp are pure local constructors accepting
empty/future string values. Calling pins do not acquire an invented Order field.
Order is signed int64, including zero, negatives and Int64 maximum; no range is
guessed. Response decoding also accepts the captured numeric-string Order form
with checked decimal parsing. Null and overflowing Order setters fail. Source
missing Id/customization setters may remain unset; explicit null setters fail,
including an earlier invalid duplicate property.

Replace/Change/Clear builders are typed for each element family. Nil request edits
mean unbound; Clear represents a bound null/empty source modifier. Empty operators
clear. Remove deletes the first match per entry; Add preserves order/duplicates.
Identity and future received fields are not equality keys. Builders and Apply own
their values/pointers/raw data. Null comparator operands fail locally. Settings
uses full replacement arrays for native sparse writes: bound nil slices become
[], while explicit null array fields reject the source DTO shape.

Readonly list decoding ignores null properties and appends repeated arrays in
wire order. RawJSON retains original duplicates, presence and future fields;
RawFields offers last-occurrence inspection. Snapshots marshal their received
representation. Requests are separate and project only known DTO fields. Int64
values remain exact in model, request and lab comparisons without float conversion.

UpdateResult retains Before, ReadResponse, PatchInvoked and HTTPAccepted on partial
failure. Update uses the read model's body Identity with the caller's route;
native Patch defaults to the caller identity with an explicit override available.
204 has no fabricated Policy. Acceptance does not prove persisted/effective app
layout. No implicit read-back, wait, rollback, mutation replay or ETag/atomicity
guarantee is added. Assign requires one explicit scope; null and empty PolicyName
remain distinct and group Rank is Int32 without an inferred backend range.

Direct compiled requests use the selected Teams audience/origin and fresh
correlation, with no forest discovery or Graph lookup. List consumes a bare array
with exact Global first; filtering is local and whitespace literal. Ordinary
model/edit/result formatting is redacted; raw data remains caller-owned sensitive
data. Application/delegated synthetic modes prove routing, not tenant roles,
cloud/feature support, full shell/Newtonsoft coercion or effective layout.
Qualification remains open under #105. See [unit tests](service_test.go)
and the [lab plan](../../../docs/acceptance/teams-app-setup-policy-lab-acceptance.md).

### Executed owned policy CRUD

`TestAccAppSetupPolicyCRUD` passed with certificate and client-secret authentication.
Each workflow creates an owned unassigned policy, verifies non-default read-back,
updates Description sparsely, checks list presence and independently verifies
selected HTTP404 and list absence after deletion. Two pinned public MicrosoftTeams
8.1.0 reads match the complete known scalar/collection getter projection.
Description changes while the explicitly disabled `AllowUserPinning` setting
remains unchanged.

Actual HTTP201/204, created/updated, owned-list and deleted404 responses supply
`TestUnitOwnedCRUDUsesAcceptanceResponses`. It checks exact requests, native Set
preparation/acceptance, typed settings, complete raw snapshots and preservation
of every omitted app collection. CI requires all CRUD assertions, both pinned
comparisons and one cleaned journal entry. Global, policy assignments, installed
apps, collection-edit variants and effective user behavior remain separate tests.
