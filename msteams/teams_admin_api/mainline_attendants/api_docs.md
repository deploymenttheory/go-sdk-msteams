# Mainline attendants

Use `sdk.TeamsAdminAPI.MainlineAttendants` for the Teams mainline attendant:
appointment booking and question and answer flows, spam detection templates,
and the supported languages and voices. Operations use the shared
client/auth/transport controls through the ConfigAPI backend and need no
PowerShell runtime. Constructing the root client or this service performs no
network I/O.

All routes are under `/Teams.VoiceApps`; `{flows}` below is
`/Teams.VoiceApps/mainline-attendant-flow`.

| Resource method | PowerShell command | Wire operation | Declared successful statuses |
| --- | --- | --- | --- |
| `GetTenantInformation(ctx)` | `Get-CsMainlineAttendantTenantInformation`, `Get-CsMainlineAttendantSupportedLanguages`, `Get-CsMainlineAttendantSupportedVoices` | GET `/Teams.VoiceApps/information` | 200 |
| `ListFlows(ctx, options)` | `Get-CsMainlineAttendantFlow` | GET `{flows}[?type=&relatedConfigurationId=&First=&Skip=&NameFilter=&SortBy=&Descending=True]` | 200 |
| `GetFlowByID(ctx, id)` | `Get-CsMainlineAttendantFlow -Identity` | GET `{flows}/{id}` | 200 |
| `ListAppointmentBookingFlows(ctx, options)` | `Get-CsMainlineAttendantAppointmentBookingFlow` | GET `{flows}/appointment-booking[?First=…]` | 200 |
| `GetAppointmentBookingFlowByID(ctx, id)` | `Get-CsMainlineAttendantAppointmentBookingFlow -Identity` | GET `{flows}/appointment-booking/{id}` | 200 |
| `CreateAppointmentBookingFlow(ctx, request)` | `New-CsMainlineAttendantAppointmentBookingFlow` | POST `{flows}/appointment-booking` | 201, 200 |
| `UpdateAppointmentBookingFlowByID(ctx, id, request)` | `Set-CsMainlineAttendantAppointmentBookingFlow` | PUT `{flows}/appointment-booking/{id}` | 200 |
| `DeleteAppointmentBookingFlowByID(ctx, id)` | `Remove-CsMainlineAttendantAppointmentBookingFlow` | DELETE `{flows}/appointment-booking/{id}` | 200, 204 |
| `ListQuestionAnswerFlows(ctx, options)` | `Get-CsMainlineAttendantQuestionAnswerFlow` | GET `{flows}/question-answer[?First=…]` | 200 |
| `GetQuestionAnswerFlowByID(ctx, id)` | `Get-CsMainlineAttendantQuestionAnswerFlow -Identity` | GET `{flows}/question-answer/{id}` | 200 |
| `CreateQuestionAnswerFlow(ctx, request)` | `New-CsMainlineAttendantQuestionAnswerFlow` | POST `{flows}/question-answer` | 201, 200 |
| `UpdateQuestionAnswerFlowByID(ctx, id, request)` | `Set-CsMainlineAttendantQuestionAnswerFlow` | PUT `{flows}/question-answer/{id}` | 200 |
| `DeleteQuestionAnswerFlowByID(ctx, id)` | `Remove-CsMainlineAttendantQuestionAnswerFlow` | DELETE `{flows}/question-answer/{id}` | 200, 204 |
| `ListSpamDetectionTemplates(ctx, options)` | `Get-CsMainlineAttendantSpamDetectionTemplate` | GET `{flows}/spam-detection-template[?First=…]` | 200 |
| `GetSpamDetectionTemplateByID(ctx, id)` | `Get-CsMainlineAttendantSpamDetectionTemplate -Identity` | GET `{flows}/spam-detection-template/{id}` | 200 |
| `CreateSpamDetectionTemplate(ctx, request)` | `New-CsMainlineAttendantSpamDetectionTemplate` | POST `{flows}/spam-detection-template` | 201, 200 |
| `UpdateSpamDetectionTemplateByID(ctx, id, request)` | `Set-CsMainlineAttendantSpamDetectionTemplate` | PUT `{flows}/spam-detection-template/{id}` | 200 |
| `DeleteSpamDetectionTemplateByID(ctx, id)` | `Remove-CsMainlineAttendantSpamDetectionTemplate` | DELETE `{flows}/spam-detection-template/{id}` | 200, 204 |

Every method takes a context and returns the actual Resty response alongside
its value/error. Reads are replay-safe; writes are never replayed automatically.
Any other HTTP status is an `*client.APIError` (a missing flow or template is a
404 with code `100004`; a rejected value is a 400 with code `100001`); an
undeclared success status is `ErrResponse`. Invalid input fails with `ErrInput`
before discovery or credentials are used.

## Lists

`ListOptions` maps the module's `-First`, `-Skip`, `-SortBy`, `-Descending`
and `-NameFilter`; nil fields and a false `Descending` are omitted.
`FlowListOptions` adds `Get-CsMainlineAttendantFlow`'s `-Type`
(`FlowTypeAppointmentBooking` or `FlowTypeQuestionAnswer`) and
`-RelatedConfigurationId`. Every list returns a non-nil slice.

## Updates replace

The module's `Set-*` commands send the whole object read back. The update
requests omit nil fields, so pass the values read back for anything you want
to keep. Flow updates always send their `Type`. `RelatedConfigurationIDs` is
sent only when non-nil; an empty, non-nil slice is sent as `[]`, as the module
does for a flow with no related configurations.

## Appointment booking flows

`AppointmentBookingPlatform` is `BookingPlatformMicrosoftBookings` (with a
`BookingLink`) or `BookingPlatformCustom`. With Microsoft Bookings the service
ignores the caller authentication method, API authentication type and API
definitions. With Custom, `APIDefinitions` is JSON text the service validates;
it must define the operations the service requires (for example
`AuthenticateCaller`). The SDK checks only that it is valid JSON. The module
reads the definitions from a file and re-serializes them; the SDK sends the
text unchanged. `CallerAuthenticationMethod` and `APIAuthenticationType` travel
as the module's enum integers.

## Question and answer flows

The module reads a JSON file that lists local files and URLs and uploads the
files base64-encoded. `KnowledgeBase` takes the file bytes and URLs directly,
and `EncodeKnowledgeBase` produces the same service text (file name,
base64 content, extension and description; URL and description). At least one
file or URL is required.

The service indexes the knowledge base asynchronously. A flow read back reports
indexing state in `KnowledgeBase` (a JSON status list, not the uploaded
content): parse it with `KnowledgeBaseDocuments`, or call `Indexing`. The
service rejects updates with HTTP 400 while any document is still in progress,
and it requires `KnowledgeBase` on every update; pass the text read back (as the
module does) or new content from `EncodeKnowledgeBase`.

## Spam detection templates

`CreateSpamDetectionTemplateRequest` requires `Name` and a `Description` of at
most 500 characters (the update checks the same limit). A nil
`EnableSpamDetection` or `Action` takes the module defaults, `true` and
`SpamActionDisconnectCall`; both are always sent on create. `CallTarget` is
sent with the module's expansion: `EnableTranscription` and
`EnableSharedVoicemailSystemPromptSuppression` only when true, and
`CallPriority` only for `ApplicationEndpoint` and `ConfigurationEndpoint`
targets.

## Tenant information

`GetTenantInformation` decodes the `MainlineAttendantTenantInformation` part of
the voice apps tenant information. The module projects its defaults as
`DefaultLanguageId`, `DefaultVoiceId` and `DefaultTimeZoneId`; the SDK keeps the
wire names. The auto attendant part of the same response stays in the response
body.
