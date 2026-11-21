# Initial live application authentication qualification

Executed 2026-11-20 at 14:01:56 UTC against the user-authorized lab tenant with
SDK main revision `53c4a1d` (after AUTH-03 and HTTP-02). The user supplied a bootstrap
credential source and authorized a separate service principal for SDK development.
All identity identifiers and credentials remain outside this repository, in the
user-selected local testing directory. No tokens or tenant records are attached here.

A dedicated single-tenant application/service principal was created with a 30-day
RSA certificate and a 29-day client secret. Its only application permission for these
checks is Microsoft Graph `Organization.Read.All`; no directory role was assigned.
No Teams resource API permission was configured, following Microsoft's
[application authentication setup](https://learn.microsoft.com/en-us/microsoftteams/teams-powershell-application-authentication).
Transient directory propagation errors during provisioning were resolved before
validation; the existing recorded objects were reused rather than duplicated.

| SDK flow | Graph token | Teams token | Graph organization read | Supplied-token reuse |
|---|---|---|---|---|
| Client secret | Passed | Passed | HTTP 200 | HTTP 200 |
| PEM certificate/private key | Passed | Passed | HTTP 200 | HTTP 200 |

The validator constructed the actual public SDK client, acquired each exact resource
through its session, and called `GET /v1.0/organization?$select=id` through the shared
Graph transport. It checked the response status and a nonempty organization result
without printing returned identifiers. A second client used explicit audience-keyed
supplied tokens for the same read. Each client was closed after validation. The
secret-free report and repeatable validator are stored with the private lab setup.

This qualifies those application-token and transport paths only. It does not qualify
Teams management authorization, resource command parity, tenant policy mutations,
delegated sign-in, managed identity or workload federation. AUTH-04 subsequently added
machine identity providers with offline contracts; live GitHub federation setup awaits
explicit approval. Resource services and their command ledger entries retain their
existing status. AUTH-05's [tests](../../../msteams/auth) adds explicit
permitted/forbidden/unauthenticated checks without extending the lab app's grants.
