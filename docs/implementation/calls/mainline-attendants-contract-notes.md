# Mainline attendants contracts

Owner: CALL-26, settling the CALL-03 research inline for this family.
Baseline: MicrosoftTeams 8.1.0 (package SHA256
`78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`).

## Evidence

The 16 commands are AutoRest cmdlets with custom wrappers in
`custom/Merged_custom_PsExt.ps1`. Routes, verbs, query names and the request
and response DTOs come from the generated operation bodies and model `ToJson`
methods in `bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll`, read by IL
inspection. Request bodies were captured from the pinned module with an
`-HttpPipelinePrepend` step, and every route was confirmed against the lab with
the SDK.

| Command | Wire operation | Lab status |
| --- | --- | --- |
| `Get-CsMainlineAttendantTenantInformation`, `…SupportedLanguages`, `…SupportedVoices` | GET `/Teams.VoiceApps/information` | 200 |
| `Get-CsMainlineAttendantFlow` | GET `/Teams.VoiceApps/mainline-attendant-flow[?type=&relatedConfigurationId=&First=&Skip=&NameFilter=&SortBy=&Descending=True]` | 200 |
| `Get-CsMainlineAttendantFlow -Identity` | GET `…/mainline-attendant-flow/{id}` | 200; 404 code 100004 |
| `Get/New/Set/Remove-CsMainlineAttendantAppointmentBookingFlow` | GET/POST `…/appointment-booking`, GET/PUT/DELETE `…/appointment-booking/{id}` | 200, 201, 200, 200; repeated delete 404 |
| `Get/New/Set/Remove-CsMainlineAttendantQuestionAnswerFlow` | GET/POST `…/question-answer`, GET/PUT/DELETE `…/question-answer/{id}` | 200, 201, 200, 200 |
| `Get/New/Set/Remove-CsMainlineAttendantSpamDetectionTemplate` | GET/POST `…/spam-detection-template`, GET/PUT/DELETE `…/spam-detection-template/{id}` | 200, 201, 200, 200; repeated delete 404 |

- List responses wrap items in `MainlineAttendantFlowResponses` or
  `SpamDetectionTemplates`; a single flow read is `MainlineAttendantFlowResponse`;
  create and update return `MainlineAttendantAppointmentBookingFlow`,
  `MainlineAttendantQuestionAnswerFlow` or `SpamDetectionTemplate`. Delete
  returns `{"Diagnostics":[]}`. The SDK requires the named envelope.
- The routes also accept `ExcludeContent` and `TypeFilter` queries, which no
  cmdlet parameter binds; the SDK does not expose them.
- The three tenant information commands call the internal
  `Get-CsAutoAttendantTenantInformation`. The module projects
  `MainlineAttendantTenantInformation` with `DefaultLanguageId`,
  `DefaultVoiceId` and `DefaultTimeZoneId`; the wire names are
  `DefaultLanguage`, `DefaultVoice` and `DefaultTimeZone`.
- `CallerAuthenticationMethod` (`Sms`=0, `Email`, `VerificationLink`,
  `Voiceprint`, `UserDetails`) and `ApiAuthenticationType` (`Basic`=0, `ApiKey`,
  `BearerTokenStatic`, `BearerTokenDynamic`) travel as integers. The spam
  `Action` (`DisconnectCall`, `TransferCallToOperator`, `TransferCallToTarget`)
  travels as a string.
- `Set-*` sends the instance fields but not `Identity` (it is in the path); flow
  updates send `Type`, and the module sends `RelatedConfigurationIds` as read
  back (`[]` for a new flow). Related configurations are `{Id, Name}` objects.
- `New-CsMainlineAttendantSpamDetectionTemplate` defaults
  `EnableSpamDetection` to true and `Action` to `DisconnectCall` and always
  sends both; the wrapper validates the description as 1–500 characters (Set
  checks the 500 limit). `CallTarget` expands to the `CallableEntity` DTO
  (`Id`, `Type`, `EnableTranscription`,
  `EnableSharedVoicemailSystemPromptSuppression`, `CallPriority`,
  `SharedVoicemailHistoryTemplateId`); the switches are sent only when true and
  `CallPriority` only for `ApplicationEndpoint` or `ConfigurationEndpoint`.
- `New-CsMainlineAttendantQuestionAnswerFlow -KnowledgeBase` takes a path to a
  JSON list of `{address_type, document_address, description}`. Its
  `ReadKnowledgeBaseContent` reads `Local_File` documents, base64-encodes them
  and serializes `{"file_descriptions":[{file_name, file_content,
  file_extension, file_description}], "url_descriptions":[{url_link,
  url_description}]}` into the `KnowledgeBase` string. The module also enforces
  a maximum document size locally; the SDK leaves that check to the service.
- `ApiDefinitions` is read from a file and re-serialized by the module; the SDK
  sends caller JSON text unchanged after a local validity check.

## Lab observations

- The lab tenant has the feature: every read returns 200 and owned creates
  succeed. No licence is assigned and nothing is billed.
- A question and answer flow's `KnowledgeBase` read-back is an indexing status
  list (`document_name`, `description`, `status`, `error`), not the uploaded
  content. New documents start `InProgress`; an unreachable URL becomes
  `Failed` within seconds. Updates while indexing return 400
  (`UpdateQuestionAnswerFlowOperation_DocumentIndexingInProgress`); an update
  without `KnowledgeBase` returns 400 (`…KnowledgeBaseNullOrEmpty`). Once
  indexing settles, updates accept either new content or the status text read
  back (the module's Set does the latter).
- An omitted `ApiAuthenticationType` reads back as 0 on question and answer
  flows.
- Custom appointment booking needs an `ApiDefinitions` document with the
  operations the service requires (it reported `AuthenticateCaller` missing and
  rejected simple shapes as invalid). Microsoft Bookings with a booking link is
  accepted without definitions and reads back `ApiDefinitions` as `{}`; the
  link is not dereferenced on create.
- Error envelopes are `{"code","message","action"[,"errorDetails"]}`; messages
  embed the tenant ID, so fixtures replace them.

## Acceptance

`TestAccMainlineAttendantReads` reads the tenant information, every list and a
missing flow and template, with pinned PowerShell tenant information parity.
`TestAccMainlineAttendantLifecycle` journals and creates one owned
`msteams-sdk-*` spam detection template, Microsoft Bookings appointment booking
flow and question and answer flow (one unreachable example.com URL), waits for
indexing to settle, updates and reads each back, lists them with filters,
compares all three with the pinned PowerShell reads, then deletes them and
observes 404 absence. Recovery resolves a lost create by its unique name and
only deletes resources whose name is the journal fixture.

## Not proven live

- The Custom booking platform with real API definitions, and spam templates
  with call targets or scopes: contract-tested only, as they need real
  endpoints or targets.
- Related configurations: only an auto attendant can relate a flow, and the
  lab has no auto attendant to attach.
