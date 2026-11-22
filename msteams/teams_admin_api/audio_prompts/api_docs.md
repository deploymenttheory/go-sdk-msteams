# Audio files

Use `sdk.TeamsAdminAPI.AudioFiles` for tenant audio files in Teams media
storage: the greetings, menu prompts and music used by auto attendants, call
queues and call hold. Operations use the shared client/auth/transport controls
through the ConfigAPI backend and need no PowerShell runtime. Constructing the
root client or this service performs no network I/O.

| Resource method | PowerShell command | Wire operation | Declared successful statuses |
| --- | --- | --- | --- |
| `List(ctx, applicationID)` | `Get-CsOnlineAudioFile` | GET `/Teams.MediaStorage/audiofile/{applicationId}` | 200 |
| `GetByID(ctx, applicationID, id, options)` | `Get-CsOnlineAudioFile -Identity` | GET `/Teams.MediaStorage/audiofile/{applicationId}/{id}[?durationInMins=]` | 200 |
| `Create(ctx, request)` | `Import-CsOnlineAudioFile` | POST `/Teams.MediaStorage/audiofile` | 200, 201 |
| `GetContentByID(ctx, applicationID, id)` | `Export-CsOnlineAudioFile` | GET `/Teams.MediaStorage/audiofile/{applicationId}/{id}/content` | 200 |
| `DeleteByID(ctx, id)` | `Remove-CsOnlineAudioFile` | DELETE `/Teams.MediaStorage/audiofile/TenantGlobal/{id}` | 200, 204 |

Every method takes a context and returns the actual Resty response alongside
its value/error. Reads are replay-safe; `Create` and `DeleteByID` are never
replayed automatically.

## Application IDs

`TenantGlobal`, `OrgAutoAttendant` and `HuntGroup` scope a file to its consumer.
As in the module, an empty application ID means `TenantGlobal` and the three
known names are matched case-insensitively and sent in canonical spelling.
Other values are passed through for the service to judge. A file is only
visible under the application it was uploaded to; reading it under another
application returns HTTP 404. `DeleteByID` always addresses `TenantGlobal`,
because `Remove-CsOnlineAudioFile` has no application parameter.

## Content

`ImportRequest.Content` is the raw file; it is base64-encoded on the wire with
the file name sent as `OriginalFilename`. The service validates the file and
rejects unsupported content with a text/plain HTTP 400. Stored audio is
transcoded (for example an 8 kHz WAV is returned as 16 kHz), so
`GetContentByID` returns the stored file, not necessarily the uploaded bytes.

`AudioFile.Content` is nil when the service sends null or omits it, and a
non-nil empty slice when it sends `""`; list and detail reads return it empty.
Standard and URL-safe base64, padded or not, are accepted. Timestamps, the
`Duration` text and the short-lived `DownloadUri` (a signed URL: treat it as a
secret) are kept as service text. `RawFields` retains every member, including
future ones. `AudioFile` and `ImportRequest` redact themselves when formatted.

## Deletion

Delete is a soft delete. The file stays readable (and listed) with
`DeletionTimestampOffset` set, which `MarkedForDeletion` reports, until the
service purges it. Deleting it again returns HTTP 400. Check
`MarkedForDeletion` rather than expecting a 404 after delete.

## Errors

Local validation (empty, `.`/`..`, slash, backslash or control-character IDs;
an empty file name or content) returns `ErrInput` before discovery or token
acquisition. HTTP failures are `*client.APIError` with the actual status; the
service's error bodies are plain text and are left in the Resty response.
Malformed success bodies return `client.ErrDecode`, and undeclared success
statuses return `ErrResponse`.

## Custom prompts

`New-CsCustomPrompt` and `New-CsCustomPromptPackage` build local values for
the audio conferencing custom prompts configuration. They are
`audioconferencingprompts.NewPrompt` and `NewPromptPackage` in the
`audio_conferencing_prompts` package, next to the configuration they edit.
