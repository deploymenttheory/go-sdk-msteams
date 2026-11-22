# Audio prompts contracts

Owner: CALL-21, settling the CALL-01..03 research inline for this family.
Baseline: MicrosoftTeams 8.1.0 (package SHA256
`78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`).

## Evidence

The four audio file commands are AutoRest cmdlets (`custom/CmdletConfig.json`)
with custom wrappers in `custom/Merged_custom_PsExt.ps1`. Routes, verbs, the
optional query and the request DTO come from the generated operation bodies in
`bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll`, read by IL inspection,
and were confirmed against the lab with the SDK's own headers.

| Command | Wire operation | Lab status |
| --- | --- | --- |
| `Get-CsOnlineAudioFile` (no Identity) | GET `/Teams.MediaStorage/audiofile/{ApplicationId}` | 200, JSON array |
| `Get-CsOnlineAudioFile -Identity` | GET `/Teams.MediaStorage/audiofile/{ApplicationId}/{Identity}?durationInMins=` | 200; 404 text/plain when absent |
| `Import-CsOnlineAudioFile` | POST `/Teams.MediaStorage/audiofile` | 201; 400 text/plain for invalid audio |
| `Export-CsOnlineAudioFile` | GET `/Teams.MediaStorage/audiofile/{ApplicationId}/{Identity}/content` | 200, `{"Content":"<base64>"}` |
| `Remove-CsOnlineAudioFile` | DELETE `/Teams.MediaStorage/audiofile/TenantGlobal/{Identity}` | 204; 400 when repeated |

- The wrappers normalize `ApplicationId`: empty or `-like TenantGlobal`,
  `OrgAutoAttendant` and `HuntGroup` (case-insensitive) become the canonical
  spelling; any other value is passed through. Remove has no ApplicationId
  parameter and always sends `TenantGlobal`.
- `DurationInMin` (Int32) is sent as the `durationInMins` query only when bound.
- Import sends the `TenantAudioFileDto` with nulls omitted. The wrapper binds
  only `ApplicationId`, `OriginalFilename` (from `-FileName`) and `Content`
  (base64 of `-Content` bytes); the SDK sends exactly those three.
- Export's wrapper base64-decodes `Content` to `byte[]`; the generated model
  accepts URL-safe characters, so the SDK accepts standard or URL-safe,
  padded or unpadded base64.
- The response DTO has 13 members (`ApplicationId`, `Content`, `ContextId`,
  `ConvertedFilename`, `DeletionTimestampOffset`, `DownloadUri`,
  `DownloadUriExpiryTimestampOffset`, `Duration`, `GroupId`, `Id`,
  `LastAccessedTimestampOffset`, `OriginalFilename`,
  `UploadedTimestampOffset`). The module projects them to an `AudioFile` with
  `MarkedForDeletion`, which the SDK exposes as a method over
  `DeletionTimestampOffset`.

## Lab observations

- A file is visible only under its own application ID (other IDs: 404).
- Delete is soft: the file stays readable and listed with
  `DeletionTimestampOffset` set; a second delete returns 400.
- The service transcodes stored audio: an 8 kHz mono WAV is exported as a
  16 kHz file with a different header. Export parity is therefore checked
  against the pinned PowerShell export (SHA-256 of the bytes), not the upload.
- Error bodies are plain text, not a JSON envelope; the SDK reports them as
  `*client.APIError` with the actual status.
- `DownloadUri` is a short-lived signed blob URL, and `GroupId` embeds a
  tenant-derived value; fixtures redact both.

## Custom prompt builders

`New-CsCustomPrompt` and `New-CsCustomPromptPackage` are compiled Policy Core
cmdlets (`NewCustomPromptCmdlet`, `NewCustomPromptPackageCmdlet`) deriving from
`BaseTpmNewElementCmdlet`. Their IL only constructs `CustomPrompt` /
`CustomPromptPackage` and copies bound properties: no service call. The
element setters validate `Id` as a GUID and reject null `Type`, `Locale` and
`DefaultLocale`. These elements are only consumed by
`Set-CsTeamsAudioConferencingCustomPromptsConfiguration`, so they map to the
existing `audioconferencingprompts.NewPrompt` and `NewPromptPackage`
constructors rather than a duplicate in the audio file package.

## Acceptance

`TestAccAudioFileLifecycle` uploads one journaled `msteams-sdk-*-audio.wav`
file to TenantGlobal, reads it (including the duration query and the
cross-application 404), lists it, exports it with PowerShell export parity, and
deletes it with the soft-delete observation. It never touches pre-existing lab
audio files. Recovery resolves a lost upload by its unique file name and treats
a set deletion timestamp or a 404 as absence.
