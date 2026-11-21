# POL-26 voicemail validation configuration SDK candidate

Native methods cover issue #99's two baseline commands and three observed parameter
sets. The qualification issue remains open; implementation is in-progress and contract
metadata-only. Successful tenant behavior and full parity remain unqualified.

## Source and native behavior

Pinned Core GetOnlineVoicemailValidationConfigurationCmdlet inherits
BaseTpmGetCmdlet<OnlineVoicemailValidationConfiguration>; Set inherits the matching
BaseTpmSetCmdlet. Provider methods establish direct Get/List and Set's read/modify/
PATCH composition. The Get class was additionally decompiled directly from the
pinned Core assembly. No generated forest discovery, Graph lookup, Create/Delete
or assignment endpoint is inferred.

Source is MicrosoftTeams 8.1.0, package SHA256
78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7;
Core DLL 68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49;
PolicyRp DLL a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51;
host DLL 477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af.
Named classes and BaseTpmGet/Set/HttpRequestHelper were inspected with ILSpy;
no source is vendored. Production Go loads neither PowerShell/.NET nor fixtures.

TeamsAdminAPI.OnlineVoicemailValidationConfigurations exposes GetByIdentity/List/
ListFiltered, native PatchByIdentity and explicit UpdateByIdentity. Both settings
preserve omitted/null/false/value distinctions. Non-null AudioFileValidationUri
uses the source ASCII-case HTTPS regex, including final-LF behavior. Null is valid;
URI parsing, reachability, size/length limits and audio upload are not invented.
The URI is never dereferenced. Snapshot decoding validates every setter occurrence,
including case variants/duplicate keys, while retaining received presence/future data.

Settings validate before I/O. Sparse writes contain only bound fields and string
Force. Update reads under the Set command metadata, binds the read model's body
identity and retains preparation/acceptance on failure. No mutation replay,
automatic read-back/wait/rollback or concurrency/ETag guarantee is implicit.
Shared destination/audience/context guards, escaped scoped identities and fresh
compiled-policy correlations remain in place.

## Executed evidence

The pure setter capture checks Core/PolicyRp hashes and executes actual model
setters/Newtonsoft serialization with a C# true-null branch, avoiding PowerShell
string-property coercion. Eleven observations (six accepted/five rejected) feed
snapshot/request validation tests. The script runs no command, credential access,
request helper or URI read. Framework-reference warning 1701 is explicitly unified
as in existing pinned adapters; captures are local, not ordinary Go CI execution.

The direct request capture checks Core/PolicyRp/host hashes and invokes actual
HttpRequestHelper methods with synthetic InvocationInfo. Fake host refuses tokens;
private client/version/host are replaced/restored and the only sender is in memory.
MPACmdlet default is inserted from factory source. Ten observations establish
Get/List/PATCH routes/body/metadata, explicit null/URI/false Force, empty replies,
invalid URI decoding (including lowercase names), wrong list envelopes and denial.
No cmdlet constructor/ProcessRecord, authentication, generated discovery or network
runs. Go CI consumes these goldens without PowerShell or credentials.

Root-client tests use both synthetic modes to prove request routing, not tenant
mode eligibility. Compositions cover omission, partial read/write/decode failures,
normalized error/status redaction, non-replay, validation before tokens, cancellation,
wrong backend and local wildcard/Global ordering. The shared header spy includes
this resource. A field inventory matches both exposed settings to pinned metadata;
examples compile. Structural reports do not promote full command/live coverage.

## Tenant qualification

TestLabVoicemailValidationReadOnly ran with dedicated-SPN secret and certificate;
both Global requests returned resource HTTP 403. Filtered list and optional pinned
PowerShell scalar comparison were not reached. No mode restriction or feature/RBAC
cause is inferred. The lab plan separates these reads from tenant-setting writes
with snapshots/restore and effective-behavior checks. Those writes, permission/cloud
availability and complete binding/coercion remain unrun qualification work. No tenant
settings, roles or consent were changed for this candidate.
