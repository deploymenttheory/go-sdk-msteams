# App catalog

Use `sdk.GraphAPI.AppCatalog` to list, publish, update and remove apps through
Graph `/v1.0/appCatalogs/teamsApps`, following the pinned compiled TeamsApp
commands. Constructors perform no authentication or network I/O.

| Method | Public source command | HTTP operation | Source status predicate |
| --- | --- | --- | --- |
| `List` / `NewPager` | `Get-TeamsApp` | GET collection and trusted continuation | Any 2xx, then decode collection |
| `Create` | `New-TeamsApp` | POST collection, application/zip bytes | Any 2xx, then decode nullable app |
| `UpdateByID` | `Set-TeamsApp` | PUT app ID path, application/zip bytes | Any 2xx, no model decode |
| `DeleteByID` | `Remove-TeamsApp` | DELETE app ID path | Any 2xx, no model decode |

These are HTTP acceptance rules, not proof of publishing completion, persistent
state or deletion. A malformed accepted response retains its receipt and reports
`client.ErrDecode`. An empty/null Create response returns a nil model with its
HTTP response and no fabricated identity. Never replay an accepted/uncertain
write blindly, including after a decoding failure. Uploads and deletion are never
replayed automatically. Reads use the native shared replay-safe policy.

## Nullable response shape

This unreleased API corrects `TeamsApp.ID` from `string` to `*string`, and
`NewPager`/`List` entries from `TeamsApp` to `*TeamsApp`. A null row remains nil;
a nonnull model can have nil, empty or whitespace ID. Consumers must check both
the row and ID before reading or using an identity. Required safe mutation-ID
validation remains separate and unchanged.

All four source string fields preserve nullability and source primitive coercion.
Numbers become their exact JSON lexemes, including large opaque identifiers,
fractions and exponents; no float64 conversion occurs. Booleans become lowercase
text. Case aliases are accepted; scalar duplicates use the last wire value.
`RawFields` preserves the original document as a native extension, separately
from these projected fields; it never selects which case alias wins.

The resource-local page decoder follows case-insensitive `value` and
`@odata.nextLink` matching. Repeated collection aliases append arrays in wire
order; null resets the collection and a later array starts a new one. Missing/null
collections fail rather than reporting empty success. Whitespace-only continuation
text ends traversal; all other continuation values pass the existing trusted
operation/origin, cycle and paging bounds. Native partial items and the last
response remain available on failure; the source buffers every page before
emitting any output. An explicit native `Paging.ExpectedStatus` can narrow the
otherwise any-2xx read rule.

## Filters and packages

The four filters retain the source order: id, externalId, displayName,
distributionMethod. Whitespace-only values are omitted, including whitespace-only
control strings. Mixed controls remain data encoded by the query builder. The
SDK doubles literal OData quotes; source leaves them undoubled. Neither behavior
relaxes mutation path or continuation trust checks.

`CreateRequest.DistributionMethod` is required and must be organization in any
letter case. Source metadata is mandatory, its CLR default is null, and no
organization default is inferred. Omission, null and empty fail source binding;
other nonempty values bind but the processing body rejects them. The SDK rejects
unsupported values locally before authentication.

The source's required Path maps to caller-owned file I/O and `Package` bytes;
compiled examples retain that boundary. A nil package means omitted input and
fails locally. An explicit nonnil empty slice represents a loaded zero-length
file and is sent unchanged as application/zip, as the source permits. This does
not establish that Graph accepts an empty package. Required mutation IDs still
reject blank/control/path-navigation inputs. The dedicated lab's historical 403 results
are denials, not successful lifecycle qualification.

See [unit tests](service_test.go).

## Authorization and current protocol

Application catalog reads have documented AppCatalog.Read.All support. Current
publish/delete/update documentation instead requires delegated prerequisites
and marks application access unsupported. The pinned Set PUT is a separate
legacy source operation from the currently documented appDefinitions POST;
that newer permission table does not prove the legacy PUT's current auth support.
See the [sourced permission and protocol table](../../../docs/guides/app-catalog-permissions.md)
for operation-specific guidance. The SDK acquires no new consent/role and does
not switch identity or protocol automatically.
