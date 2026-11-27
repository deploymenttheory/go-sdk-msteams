# AppCatalog authorization and current API boundary

Checked against current Microsoft Graph documentation on 2026-11-27. These
published API facts supplement the pinned 8.1.0 SDK behaviour; they do not prove
successful tenant execution or authorize new credentials, consent or role changes.

| Operation | Documented identity/permission guidance | Boundary |
| --- | --- | --- |
| List catalog | Application AppCatalog.Read.All, or higher AppCatalog.ReadWrite.All. Organizational delegated access also supports AppCatalog.Submit/read/readwrite permissions. | [List documentation](https://learn.microsoft.com/en-us/graph/api/appcatalogs-list-teamsapps?view=graph-rest-1.0). Results follow app-management visibility policies and can lag publication; they are not an unconditional admin inventory. |
| Publish app | Delegated organizational access; AppCatalog.Submit is limited to review submission, while AppCatalog.ReadWrite.All supports catalog publication. Application access is unsupported. | [Publish documentation](https://learn.microsoft.com/en-us/graph/api/teamsapp-publish?view=graph-rest-1.0). Native Create sends the pinned collection POST without a review option. |
| Delete published app | Delegated organizational catalog management, with AppCatalog.ReadWrite.All and Teams Service administrator or higher. Application access is unsupported; Submit is limited to definitions in review. | [Delete documentation](https://learn.microsoft.com/en-us/graph/api/teamsapp-delete?view=graph-rest-1.0). Use the catalog-returned ID, distinct from manifest/external ID. |
| Current published update | Delegated organizational access with documented catalog permissions/admin prerequisites; application access is unsupported. | [Current update documentation](https://learn.microsoft.com/en-us/graph/api/teamsapp-update?view=graph-rest-1.0) specifies POST to the appDefinitions collection. This is not the pinned Set operation below. |

The pinned Set-TeamsApp implementation uses PUT to
`/v1.0/appCatalogs/teamsApps/{id}`. The SDK preserves that source-parity primitive.
Current update documentation instead specifies POST to
`/v1.0/appCatalogs/teamsApps/{id}/appDefinitions`. Neither route equivalence nor
current permission/auth support for the legacy PUT follows from that document.
No silent protocol replacement, guessed fallback, delegated fallback or unsupported
mode rejection is introduced by this source checkpoint. Its server/authorization
compatibility remains unqualified pending an approved owned fixture and evidence.

Historical client-secret 403 observations remain denials. Listing has documented
application support, but the current lab has no successful catalog-read proof.
Published write operations need appropriate delegated prerequisites; obtaining an
application write permission alone must not be presented as the remedy. No new
role, consent, license, app publication or installation is performed here.
