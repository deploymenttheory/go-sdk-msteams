# Tenant-abuse configuration

`TeamsAdminAPI.TeamsTenantAbuseConfigurations` implements the three public pinned
8.1.0 Get/Set parameter sets through compiled PolicyRP. Construction has no I/O.
GetByIdentity/List/ListFiltered retain full owned snapshots and nullable typed
settings, stable Global-first ordering and local Identity/Tag wildcard filtering.
PatchByIdentity sends a single sparse PATCH; UpdateByIdentity composes Set's
preparation GET, local list edits and PATCH while retaining partial receipts.
There is no public create, delete or grant surface.

Seven thresholds are nullable UInt64; TrialTenantValidation and SkipValidation
are nullable opaque strings. `policy.Field` retains omitted/null/zero/empty values.
The BlockedTenants get-only list materializes empty and appends repeated arrays;
null JSON does not replace the collection. Optional ListEdit applies first-match
ordinal removal then Add without deduplication, or replacement/clear. Requests send
full arrays. Static source regex declarations are not executed by the setters,
so no GUID validation is invented. Force uses compiled string encoding, and
Force alone is not a setting. Context and write no-replay rules are shared.

The lab Teams Administrator application receives HTTP403 / PolicyRP40301 from both
SDK and pinned PowerShell with certificate and client-secret authentication. The
mandatory denial test verifies the exact error; it does not qualify successful
reads or writes. These remain configured exact-name skips with reasons. No roles
or tenant-abuse settings changed. Units use this actual denial body for error
propagation; standalone source vectors prove all UInt64 bits, fields and bindings.
No successful response fixture is fabricated.

TestAccTenantAbuseConfigurationUpdateRestore remains executable for an authorized
identity. It refuses mutation without a readable non-null original, journals the
UInt64 before changing LowSeatLimit, verifies all getters/omitted fields, compares
PowerShell before/after/restoration and independently verifies the original. Bound
recovery kind `tenant-abuse-configuration` restores only LowSeatLimit on Global.
Unsigned snapshots reject negative/overflow/null/unknown fields.

Pinned sources: TeamsTenantAbuseConfiguration, BlockedTenants, Get/Set classes,
BaseTpm and PolicyRP provider. Current application-auth documentation describes
RBAC-bound app access; this lab denial is not classified as universal app-mode
unsupported. See the repository qualification notes and [Microsoft authentication
reference](https://learn.microsoft.com/en-us/microsoftteams/teams-powershell-application-authentication).

The canonical BlockedTenants response field is now `[]*string`: null entries
remain distinct from empty strings. This corrects its earlier lossy `[]string`
projection and changes that public field type. `policy.NullableStrings`,
`ReplaceNullableStrings`, `ChangeNullableStrings` and `ApplyNullable` preserve
null members and copy pointed-to values. Existing Strings/ReplaceStrings/
ChangeStrings and Apply([]string) signatures and nonnull behavior are retained.
A nil edit omits the setting; an empty replacement/delta clears it. Prepared
edits remove the first ordinal/null match, then append without deduplication.
