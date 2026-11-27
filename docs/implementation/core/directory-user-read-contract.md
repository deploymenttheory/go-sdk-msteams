# Directory-user read projection

This bounded CORE-90 correction concerns the supporting directory read snapshot,
not directory-user lifecycle or mutation qualification. The actual pinned type is
`Microsoft.Teams.PowerShell.TeamsCmdlets.Model.AadUser`, rather than a CLR type
named GraphUserIdentityDTO. Its five properties represented natively (Id,
DisplayName, UserPrincipalName, UserType and Mail) are all System.String with
plain backing-field setters. The observer records actual getter/setter IL bytes,
assembly SHA256 and its own producer hash before running actual JsonConvert
operations. It invokes no management lifecycle, authentication or HTTP.

[The source packet](evidence/directory-user-read-source.json) records 70 actual
model observations: null/empty/string, Boolean, integer, exact large integer,
fraction, exponent, overflow exponent, negative zero and rejected array/object
for each of the five properties. Eleven additional property controls prove ASCII aliases, last assignment/null
reset, exact duplicates, immediate failure of an earlier invalid known field,
and ignored long-s foreign keys even when their values are arrays or objects.
The reader explicitly walks properties in wire order; Go Unicode SimpleFold
struct matching is not used. Unknown properties remain in RawFields.

Six versioned consumer IL bodies cover both GetUserId overloads, GetUpnForUsers,
both generic HttpUtilities.Get overloads and generic GetAll. Every instruction
records offsets and resolved operands using the owning generic context, plus
complete IL bytes. This makes the lookup/current/mail/parser call-chain claims
reproducible from the hash-verified assembly without invoking those consumers.

The portable reproduction is:

```powershell
pwsh -NoProfile -File scripts/discovery/observe-core-directory-user.ps1 -ModulePath <verified-8.1.0-module>
```

The assembly is Microsoft.Teams.PowerShell.TeamsCmdlets.dll SHA256
`3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24`.
Fresh GraphUtilities.GetUserId IL confirms the direct response is parsed with
JsonConvert.DeserializeObject<AadUser> at offset68 and returns get_Id at77.
The current-user helper returns get_Id at26. Only received404 activates its
historical myorganization/mail branch; First<AadUser> at149 and get_Id at154
return a string without Guid.Parse. Canonical native fallback, complete traversal
and ambiguity detection remain explicit native behavior rather than claims of
identical PowerShell selection semantics.

The native reader now opts only these five fields into that evidenced string
profile and retains original wire tokens in independent RawFields. It keeps the
existing public ID string representation: missing/null become empty in a read
snapshot. Resolve requires a nonempty usable ID after direct/current read and
for every mail row. It never turns an unusable200 into404 or falls back to mail.
Duplicate UUID spellings remain case-insensitive; opaque IDs compare exactly,
so opaque-A and opaque-a cannot establish false uniqueness. This native safety
rule does not claim the pinned first-result fallback proves uniqueness.
An accepted202 remains pending with an inspectable snapshot and no usable ID.

Input UUID passthrough remains an unverified shortcut. LookupPrincipalNames still
validates explicit requested UUIDs and rejects unrequested/duplicate returned
identities atomically for the current batch. Team/group/member/path/tenant IDs
and owner/reference writes retain all existing independent validators. Reading
an opaque ID therefore does not authorize, validate existence or prove a later
write will be accepted. Prepared team-user composition tests retain the original
resolution receipt when a subsequent local reference guard rejects the ID;
no reference mutation or automatic recovery is sent.

New success/malformed/partial tests invoke the real SDK through the contract
harness directly. Their protocol responses are authored from the actual model
contract, not captured tenant evidence. Existing requestBoundary denial tests
remain bounded request-construction evidence. After headroom recovered, the
[corrected execution](../../../contracts/coverage/evidence/core90-directory-receipt-execution.json)
at ec4b0dd passed all 675 test/subtest events across ten affected packages, with
zero failures/skips. Full configured lint on those packages returned zero issues;
formatter comparison and post-run scoped diff were clean, and both owned scratch
directories were removed. Current-head full CI remains required. No qualification
or live state changes accompany this source/implementation checkpoint.
