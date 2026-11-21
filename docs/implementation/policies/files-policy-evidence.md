# Teams files policy evidence (POL-47 / #120)

Actual 8.1.0 public Get/New/Set/Remove/Grant exports load Core cmdlets. The
provider capture verifies every implementing type/assembly. Get has Filter and
Identity sets; Grant has GrantToUser, GrantToTenant, GrantToGroup and an empty
default Identity set; New/Set/Remove have Identity. These are nine exact
command/set pairs.

## Executed source captures

The provider capture pins Core/PolicyRp/HostContract and invokes actual direct
HttpRequestHelper using synthetic InvocationInfo, a credential-refusing host and
memory sender. It records 24 provider requests and 20 actual model
constructor/setter/Newtonsoft observations. Each provider request makes one
memory send and zero token requests. Static client/version/host state is restored.
Original lifecycle, binding, auth/scope checks, discovery and network do not run.

All four settings are strings: NativeFileEntryPoints, SPChannelFilesTab,
DefaultFileUploadAppId and FileSharingInChatswithExternalUsers. Setters preserve
null/empty/future values without inferred enum, GUID, length or cross-field rules.
Captured comparisons include a long native entry mode, case-insensitive last
values, duplicate null, and wrong object/array failures. Native typed JSON need
not emulate arbitrary shell coercion.

The inherited compiled binding capture records 37 cases (32 successes, four
validation failures and one expected parameterless Grant error), plus five actual
protected Get filter selections. Owned subclasses inherit original properties,
attributes and constructors, replace Begin/Process/End with synthetic output,
and invoke actual New/Set predicates. Only no-scope Grant invokes actual private
ExecuteGrant with synthetic token/origin; its default branch emits a source error
and false result without sending HTTP. Other grants bind only. All binder/filter
observations have zero sends/token requests; memory client/host state is restored.

## Observed contract and implementation remaining

Compiled CRUD uses bare collection reads, exact Global-first local Identity/Tag
filters, first-slash scoped identity splitting, collection POST, sparse PATCH and
body-bearing DELETE. Set retains read-model Identity. Bound Force is a string;
MPACmdlet/cmdlet/correlation headers apply without a modern forest target. New
permits empty settings through its base processing; Set requires a bound setting.

User/tenant grants PATCH with additionalParameters:null; tenant uses capitalized
PolicyName and Boolean forceSwitchPresent. Group POST always includes required
signed Int32 Rank as priority, including zero/negative values. Nil/empty policy
names remain distinct. Global:false still selects the tenant set. Parameterless
Grant has no valid scope and fails without HTTP. Group 202 is decoded as the
compiled model, without an inferred polling/effective-assignment guarantee.

The initial native service exposes TeamsFilesPolicies with compiled
CRUD/list/explicit Set composition, all four optional string settings and owned
raw snapshots, plus user/tenant/group grants and a missing-scope input guard.
Go field names normalize DefaultFileUploadAppID and
FileSharingInChatsWithExternalUsers while preserving the exact source JSON tags.
Initial race comparisons cover all 24 provider requests in both synthetic auth
modes, all 20 model observations, 37 compiled bindings, five actual filters and
the parameterless Grant error. All nine command/set pairs have native mappings.

Remaining work includes partial Set and mutation failure/context/no-replay tests,
API docs/examples, gated lab/source reads, full checks and reviewed sanitized
artifacts. Issue #120 stays open.
No tenant read/write has run for this family; successful CRUD, persistence,
assignment or effective file/upload/external-sharing behavior is unqualified.
Microsoft binaries and decompiled source remain private.
