# POL-26 voicemail validation configuration progress

Issue #99 has two baseline commands and three observed parameter sets. The pinned
Core GetOnlineVoicemailValidationConfigurationCmdlet inherits
BaseTpmGetCmdlet<OnlineVoicemailValidationConfiguration>; Set inherits the matching
BaseTpmSetCmdlet. They use compiled direct HttpRequestHelper calls, including Set's
read/modify/PATCH composition, rather than generated forest discovery. A Get class
decompiled directly from the pinned Core DLL confirms the backend selection.

This checkpoint adds typed snapshot/settings/update models. The two source fields
are nullable AudioFileValidationEnabled and AudioFileValidationUri. The URI setter
allows null and applies the simple ASCII-case HTTPS regex to non-null strings,
including its final-LF behavior; it establishes no URI parsing, size/length or
reachability rules. Reads retain omitted/null/false/future data and validate each
setter occurrence, including case variants/duplicates. No URI is dereferenced.

Source is MicrosoftTeams 8.1.0, package SHA256
78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7;
Core DLL 68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49;
PolicyRp DLL a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51.
The named model/Set classes and provider BaseTpmGet/Set/HttpRequestHelper methods
were inspected locally with ILSpy; no source is vendored or management call run.

The pinned setter capture now executes the model in C# with a true-null branch
to avoid PowerShell string-property coercion. Eleven observations cover six accepted
and five rejected URI values; snapshot/request tests consume all cases. The script
checks Core/PolicyRp hashes, uses Newtonsoft serialization and runs no command,
credential access, request helper or URI dereference. Host framework-reference
warning 1701 is unified explicitly, as in the existing pinned adapters.

Remaining work: pinned request captures, native Get/List/ListFiltered,
PatchByIdentity/UpdateByIdentity, whitelist/root wiring, request/error tests,
examples/API docs, opt-in lab reads and exact ledger/evidence mappings. No Create,
Delete or assignment endpoint is established for these two commands. Full command
parity, successful tenant/RBAC/cloud behavior and persisted changes remain unqualified.
