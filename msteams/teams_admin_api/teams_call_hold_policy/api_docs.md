# Teams call hold policies

Use `sdk.TeamsAdminAPI.TeamsCallHoldPolicies`. One scope-bound service handles
the two exported baseline transports; construction performs no I/O.

| Method | Transport and behavior |
| --- | --- |
| `GetByIdentity`, `List`, `ListFiltered` | Compiled direct GET, bare collection and local Identity/Tag filtering |
| `CreateByIdentity` | Compiled POST with explicit identity and bound settings |
| `PatchByIdentity` | One sparse compiled PATCH |
| `UpdateByIdentity` | Explicit Set-metadata Get/PATCH composition |
| `DeleteByIdentity` | Compiled DELETE with Identity/optional Force JSON body |
| `AssignToUser`, `AssignTenantDefault`, `AssignToGroup` | Modern generated grants with forest discovery |
| `Assign` | Select one grant scope, omitting zero group Rank like the mapped wrapper |

Compiled CRUD includes MPACmdlet and command/correlation headers and performs
no forest discovery. Grant uses ConfigAPI discovery, the normalized forest root
in X-MS-Target-Uri and no unconditional MPACmdlet. Both share the same Teams
scope, token session and operation replay controls. Interleaved tests verify
that discovery/cache state and protocol headers do not leak into CRUD calls.

`Settings` has Description, AudioFileID, StreamingSourceURL and
StreamingSourceAuthType as `policy.Field[string]`. Unset omits, `policy.Null`
sends null, and `policy.Set("")` sends empty. No GUID, URL, auth-type enum or
discarded Description length limit is invented. Streaming URL values are data;
the SDK does not fetch them. The source accepts future strings and a 2,000
character description at its model layer; service acceptance remains separate.

Response Policy keeps nullable strings, complete RawJSON and last exact-key
RawFields. Known duplicate/case-varied scalar setters use the later value.
Marshaling a received policy returns its original raw snapshot; writes use
explicit known-field settings. Future data is not silently included in updates.
Arbitrary shell/Newtonsoft conversions remain unqualified.

Create sends empty settings if requested and does not implicitly read/copy Global
for source ShouldProcess presentation. Patch requires at least one bound setting;
Force alone is insufficient. Compiled Force distinguishes omission from string
False/True. Patch BodyIdentity may differ from route identity, including null.
Update retains Before/ReadResponse, uses the read Identity for the sparse body,
and retains PatchInvoked/HTTPAccepted on partial failure. After is the decoded
PATCH response, not a persisted read-back. Delete sends the compiled JSON body.

Grant requires an explicit native PolicyName string argument. Empty is the
captured clear form; the source string binder turns bound null into empty.
`AssignToGroup` can send explicit priority zero; `Assign` omits Rank nil/zero.
The mapped tenant helper/DTO supports a native force argument, although this
public Grant wrapper does not expose Force. Bound PassThru is rejected by the
mapped modern source function, including false. Default modern selection in
PowerShell is command/flight-dependent; the SDK has no PowerShell/TRPS fallback.

Group Accepted means the actual POST returned 202. Discovery 202 cannot establish
acceptance; malformed/empty accepted POSTs retain acceptance with a decoding
error. Raw accepted JSON and unknown fields are retained; 204 stays distinct.
All mutations use ReplayNever. No implicit polling, restoration, ETag/atomicity
or effective playback/assignment guarantee is added.

Compiled identity escaping separates its first scope slash and preserves the
remaining opaque identity. Modern grant identifiers are whole path segments;
tests normalize equivalent Go/.NET colon spellings per segment once. No server
requirement for a particular equivalent byte spelling is claimed.

Actual Global reads on 2026-11-21 returned 403 with client-secret and certificate
credentials. Successful authorization, cloud/flight availability, mutations,
read-back/cleanup and effective playback/grants remain open. See
[unit tests](service_test.go)
and [lab acceptance](../../../docs/acceptance/call-hold-policy-lab-acceptance.md).

## Owned CRUD acceptance

`TestAccCallHoldPolicyCRUD` creates a uniquely named unassigned policy, verifies
non-default read-back, updates the description, verifies omitted settings,
checks list membership, deletes the policy and confirms GET404 plus list absence.
Created/updated reads are compared with MicrosoftTeams 8.1.0. Certificate and
client-secret lab runs passed. CI runs both authentication modes and requires
verified cleanup and two PowerShell comparisons for each workflow.

Seven actual acceptance responses live in `testdata`; `TestUnitOwnedCRUDUsesAcceptanceResponses`
loads them to exercise requests, typed fields, update behavior, native statuses
and deleted-resource errors. Call-hold requires a real uploaded audio fixture; see [lab setup](../../../tools/acceptance/README.md). The non-default audio reference survives the description-only update. User/tenant/group grants, clear operations, effective client behavior,
delegated authentication and other clouds remain unqualified.
