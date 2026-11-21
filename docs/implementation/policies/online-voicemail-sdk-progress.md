# POL-25 online voicemail implementation progress

Issue #98 has five baseline commands and eight observed parameter sets. This
checkpoint introduces a typed policy snapshot and a .NET TimeSpan wire value;
HTTP services, root wiring, assignments and ledger mappings remain pending.
No unsupported-operation stubs or full parity/tenant qualification are claimed.

[Pinned source research](federation-branding-voicemail-evidence.md) establishes
compiled direct-provider CRUD/list and selected modern generated assignments.
Those protocols must remain distinct: compiled null-including assignment payloads
cannot be copied onto this resource's generated Grant routes. Omitted/zero Rank,
required bound PolicyName, generated null omission and the wrapper PassThru binding
gap need explicit handling. Accepted group operations have a separate read handle;
they do not prove effective assignment propagation or a policy-resource update.

The [pure model capture](../../../tools/contracts/capture-voicemail-models.ps1)
checks pinned Core/PolicyRp hashes and serializes OnlineVoicemailPolicy setters
through Newtonsoft. Its 16 observations cover zero, signed 100ns precision,
fractions, days, Int64 extrema and null. No command processing, credential access,
audio-file reading/upload or network runs. Audio fields remain reference strings;
no backend recording-length range, language restriction or default is inferred.

TimeSpan stores the full signed tick range, formats the constant wire string and
parses that constant format with overflow checks. Go time.Duration conversion
rejects sub-tick precision rather than truncating. Broader CLR/PowerShell input
binding and locale-dependent parsing remain unqualified. Snapshot tests consume
all observations and retain omitted/null/future fields without client defaults.

Remaining work: compiled CRUD request captures and services, sparse fields and
explicit update/create compositions, generated assignment routing/discovery and
operation reads, request/negative tests, examples/docs, safe lab gates and all
command/set evidence mappings. Persisted policy behavior, assignments, feature
preview, authorization/cloud availability and propagation remain unrun.
