# Teams notification and feeds policies

Use `sdk.TeamsAdminAPI.TeamsNotificationAndFeedsPolicies`. The pinned public module
exports compiled Get, Set and Remove commands. The native SDK additionally offers
the POST create operation evidenced by the package's compiled New implementation
and verified against the lab. This does not add an exported cmdlet to the baseline.

| Method | Behavior |
|---|---|
| `GetByIdentity`, `List`, `ListFiltered` | Direct compiled GET; bare list, exact Global first, local Identity/Tag filter |
| `CreateByIdentity` | Native POST; no implicit read or invented baseline command |
| `PatchByIdentity` | One sparse PATCH; requires at least one bound setting |
| `UpdateByIdentity` | Set preparation GET followed by sparse PATCH; retains both phases |
| `DeleteByIdentity` | Compiled DELETE with Identity and optional Force body |

Description, SuggestedFeedsEnabledType and TrendingFeedsEnabledType are nullable
source strings, with policy.Field[string] write bindings for omitted/null/empty
values. No enum or length limits are inferred from documentation examples.
Bound Force uses the compiled string form. Whole request execution is native Go;
PowerShell is only an opt-in test oracle. Mutations never replay automatically.

The compiled protocol uses MPACmdlet and command metadata without modern service
discovery. Scoped configuration identities split at the first slash; the remainder
is one escaped segment. Filter whitespace stays literal. Read models preserve
unknown fields and owned RawJSON; marshal returns the received snapshot.
UpdateResult preserves preparation and acceptance independently of persistence.

Both application authentication modes passed owned create/read/update/list/delete
and explicit-null Description persistence. The first update changes only its
non-default description and preserves SuggestedFeedsEnabledType. A subsequent
JSON-null PATCH clears Description, is independently read back and matches the
pinned getter. Three PowerShell comparisons cover created, updated and cleared
states. Deletion requires independent GET404 and list absence. SDK-owned unassigned
policies and pre-create recovery journals isolate the lab lifecycle. The tested
role is Teams Administrator; no policy grants or Global changes are performed.

CI requires all CRUD assertions, the explicit-null check, three comparisons and
one cleaned journal entry. Twenty permanent sanitized response files across this
and network roaming's service feed their unit tests, including real empty 201/204
bodies and deleted404 responses. Effective feeds, delegated flows, locale-specific
filter behavior and other clouds remain separate qualification scenarios.

See [qualification](../../../docs/implementation/policies/network-notification-lab-acceptance.md).
