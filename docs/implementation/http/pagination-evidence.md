# HTTP-03 pagination evidence

Status: implemented with synthetic offline fixtures. No live multi-page Teams list
has been qualified and no resource command ledger status is promoted by this helper.

## Source contracts

[Graph paging](https://learn.microsoft.com/en-us/graph/paging) defines the complete
`@odata.nextLink` URL as the continuation. Clients retain custom headers on subsequent
requests, and continue until the link is absent. An empty page with a next link is not
completion. The SDK does not append the initial query to a returned full URL.

Inspected the pinned MicrosoftTeams 8.1.0 package's
`bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll`, SHA256
`8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6`.
Offline reflection populated only synthetic strings/integers/empty arrays and invoked
generated `ToJsonString()` serializers. Method-body IL was read without invoking any
management operation. Observed fields and query literals:

| Generated result model | Serialized array | Continuation field | Request query key |
|---|---|---|---|
| `AiAgentQueryResult` | `agents` | `continuationToken` | `continuationToken` |
| `ApplicationInstanceSearchResults` | `applicationInstances` | `skipToken` | `skipToken` |
| `ConferencingUserSearchResults` | `users` | `@nextLink` | `$skiptoken` |
| `UserApiSearchResults` | `users` | `@nextLink` | `$skiptoken` |
| `SkypeTelephoneNumberMgmtPhoneNumberPolicyAssignmentCmdletResponse` | `Assignments` | `NextLink` | Resource-specific mapping still required |

The pinned proxy help explicitly calls `Skiptoken` a pagination marker using the
response's nextlink value. This differs from Graph's full-URL protocol. Request builders
confirm exact casing: `AiAgentGet` IL offset 359 contains `continuationToken=`,
`ApplicationInstanceV2SearchApplicationInstanceAsync` offsets 155/218 contain
`pageSize=` / `skipToken=`, `OdcUserSearch` offsets 103/166 contain `pageSize=` /
`$skiptoken=`, and `UsersSearchUsers` offsets 104/282 contain the same pair.

Offset patterns are independently evidenced: `CallQueueGet` and `AutoAttendantGet1`
contain `First=` / `Skip=` at IL offsets 104/167. `OnlineApplicationInstanceV2Get`
contains `resultSize=` / `skip=` at offsets 155/218 and a typed response callback for
an application-instance array. `PhoneNumberAssignmentGet` contains `top=` / `skip=`
at offsets 961/1026 and a typed response callback for a phone-number array. These
case-sensitive keys must come from the resource contract, never a universal guess.
`GetCallQueuesResponse` and `GetAutoAttendantsResponse` serialize `CallQueues` /
`AutoAttendants` arrays respectively, alongside a `Diagnostics` array.
The helpers do not infer resource routes or permissions from these pagination facts.

## Delivered behavior

`client.Pager[T]` binds a serial, lazy traversal to one `Scope`. Backend helpers are
`graphapi.NewPager`, `teamsadminapi.NewTokenPager` and `teamsadminapi.NewOffsetPager`.
Teams token formats explicitly specify array field, token field and query key. Offset
formats specify an array field (empty for a bare array) and exact offset query key.
This preserves wire differences without naming every future resource in shared code.

`PagingOptions.ExpectedStatus` optionally selects the successful status required
by a particular collection contract before any page decoder runs. For example,
the [native team-members collection](https://learn.microsoft.com/en-us/graph/api/team-list-members?view=graph-rest-1.0)
documents HTTP 200. A collection-shaped 202/206 response cannot establish a
complete page under that contract. A mismatch returns `ErrPageStatus`, retains
the earlier collected items and last response, and makes traversal terminal.
HTTP failures keep their existing typed errors. Zero preserves the generic
any-2xx behavior; a configured expectation outside 200–299 is rejected locally
without credentials. Resource adapters select their evidenced status explicitly.

Construction snapshots headers, encoded path/query and serialized JSON body without
network or authentication. Traversal accepts only explicit replay-safe GET/POST reads;
the Graph helper classifies its GET collection read. Other operations require the
resource's explicit classification. A decoder maps each evidenced envelope, with strict
array shape checks even when `T` is byte. Missing/null collections fail as decode errors
instead of masquerading as empty successful lists. Missing/null token/link fields mean
completion. Offset traversal stops on an empty array and advances by actual item count;
short server-capped pages do not cause early completion. It checks integer overflow.

Every full URL re-enters the scope's routing checks before token acquisition. The pager
never parses a Teams query token as a destination URL. Only the continuation query key
changes; original filters, selected fields, page-size settings, headers and JSON remain.
Opaque Graph URLs stay byte-for-byte intact on the wire. Loop detection normalizes a
separate URL key, detects repeated token values and bounds the traversal to 1,000 pages
by default. Explicit limits may override this; aggregate results default to 100,000 items.
Caller context is passed into each page and each request retains the shared HTTP budgets.
The pager has no worker goroutine and stores no operation context.

`Next` returns each decoded page, including empty intermediate/final pages. `HasNext`
means a continuation exists, not that the next request is valid or will succeed. It is
validated on the next call before any credentials are acquired. `Collect` retains the
successfully collected prefix and last response on an error. Transport, decoder, loop,
page-limit and item-limit errors never return truncated success. Errors after traversal
starts are terminal, including an oversized aggregate page: later calls cannot skip
that omitted page. Separate pagers are independent; one pager is consumed serially.

## Verification map

| Tests | Contract checked |
|---|---|
| `TestUnitGraphPaginationUsesOpaqueLinksAndKeepsHeaders` | Three pages with empty intermediate data, exact URL preservation, query/header snapshots and completion |
| `TestUnitTeamsTokenPaginationPreservesQueryBodyAndTreatsNextLinkAsToken` | URL-shaped opaque token stays a query value; body/query snapshots and Teams audience isolation |
| `TestUnitOffsetPaginationContinuesAfterShortPages` | Nonzero starting offset, actual-count advancement, retained limit/filter and final empty array |
| `TestUnitPaginationLoopAndDestinationErrorsPreservePartialResults` | URL-loop canonicalization, cross-host/relative rejection before auth and collected prefix |
| `TestUnitTeamsTokenLoopsStopBeforeAnotherCredentialRequest` | Repeated token stops without another acquisition |
| `TestUnitPaginationTransportAndDecodeErrorsNeverReturnTruncatedSuccess` | Second-page failures retain prefix and last response, and remain terminal |
| `TestUnitPaginationLimitsAreTerminal` | Page/item bounds, whole-page aggregation limit and no silent resume after dropped data |
| `TestUnitPaginationCancellationReturnsCollectedPrefix` | Cancellation during the second request retains prior data |
| `TestUnitDecodeCollectionRequiresArrayEvenForBytes` | Non-array/null/missing envelope rejection, including Go's special []byte JSON behavior |
| `TestUnitPaginationConfigurationAndOffsetOverflow` | Ambiguous/negative offsets, unsafe operations, constructor purity and numeric overflow |
| `TestUnitIndependentPagersDoNotShareTraversalState` | Concurrent independent traversals under the race detector |
| `TestUnitPagerExpectedStatusRejectsUnexpectedPageBeforeDecode` | Configured HTTP 200 rejects 201/202/204/206 before decoding and cannot resume after failure |
| `TestUnitPagerExpectedStatusRetainsEarlierItemsAndLastResponse` | An unexpected later successful status retains the prior prefix and raw response |
| `TestUnitPagerStatusOptionPreservesGenericAndHTTPFailureBehavior` | Default any-2xx and explicitly selected success remain compatible; a 403 remains a typed forbidden error |
| `TestUnitPagerExpectedStatusValidationIsCredentialFree` | Invalid non-success expectations acquire no token and send no HTTP |

These checks qualify traversal mechanics and the listed envelope/key mappings. They do
not qualify resource-specific consistency guarantees, total-result limits, additional
backend authorities, permissions or live management parity.
