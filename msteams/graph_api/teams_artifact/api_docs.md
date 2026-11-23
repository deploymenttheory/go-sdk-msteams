# Teams meeting artifacts

Use `sdk.GraphAPI.TeamsArtifacts.List` to find Teams meeting artifacts
(recordings and transcripts, Loop meeting notes and whiteboards) stored in
OneDrive for Business and SharePoint, as `Get-TeamsArtifacts` does. It returns
metadata only: it never downloads content and writes no `artifacts.json`.

| Resource method | PowerShell command | Wire operations |
| --- | --- | --- |
| `List(ctx, Options)` | `Get-TeamsArtifacts` | GET `/v1.0/sites?search=*`, `/v1.0/sites/{id}/drives`, `/v1.0/users[?$filter=mail eq '…']`, `/v1.0/{drive}/items/{item}/children`, each following `@odata.nextLink` |

## Scan

Without `OneDrive` or `SharePoint`, the scan visits every SharePoint site drive
(`drives/{id}`) and then every user's OneDrive (`users/{id}/drive`). `OneDrive`
selects the user whose `mail` equals the value (single quotes doubled);
`SharePoint` selects site drives only; setting both is `ErrInput`.

From each drive root the scan applies the module's rules, compared
case-insensitively:

- Folders at depth one and two are always entered. Deeper, only
  `Recordings/View Only` and an artifact folder (`Recordings`, `Meetings`,
  `Whiteboards`) whose parent is not an artifact folder are entered.
- With `Type`, an artifact folder of another type is skipped.
- A file is returned when its folder is `Recordings` (`.mp4`), `Meetings`
  (`.loop`), `Whiteboards` (`.whiteboard`) or `Recordings/View Only` (`.mp4`),
  and its `lastModifiedDateTime` is inside `Start`/`End` when given.

Each `Artifact` has the module's output fields (name, MIME type, download URL,
last modified) plus the drive path, folder path and full drive item. Download
URLs are short-lived credentials: `Artifact` formats as `teams artifact
(redacted)`.

## Failures

The module prints a failed search, drive listing or folder read and carries on;
`List` records each one in `Result.Failures` (stage, ID and the typed error)
and carries on too, after scanning items from pages read before the failure.
`List` returns an error only for invalid options, a closed client or a done
context. Searching sites needs `Sites.Read.All`; listing users needs a
directory read permission; reading drives needs `Files.Read.All`.

The module's `-Usage` switch only prints help, and its `EXPORT_TOOL.*`
environment credentials are replaced by the client's Graph session.
