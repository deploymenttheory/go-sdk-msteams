# Batch deployments

Candidate native resource for the batch-deployment portion of CORE-22; the ticket
also owns migration reads that remain in progress. `BuildArchive` is a pure
context-aware constructor for exactly Teams.csv and Users.csv, using fast Deflate
and original input bytes. It uses explicit combined input/ZIP budgets, returns
independently copied bytes or Base64, and leaves readers caller-owned. It performs
no CSV parsing, filesystem operations or shared archive update. Reader operations
must return promptly or honor caller cancellation.

`sdk.TeamsAdminAPI.BatchDeployments.Create` sends that archive as DeploymentCsvs
and the supplied UsersToNotify string. It uses selected ConfigAPI routing,
same-origin service discovery, forest header and Teams audience. The generated
fixture's development trafficmanager authority is replaced by the pinned session
callback before transmission and is never a native credential destination.
Create expects 202, retains instanceId/executionId and unknown fields, and returns
response/partial data if accepted output lacks a usable instance ID.

Submission can create/manage teams and send notification email. It is sent once,
without automatic read-back, retry, polling or compensation. HTTP acceptance is
not completed deployment. The opaque instance ID addresses GetByID; InputObject
convenience maps to typed Identity/GetByIdentity using the same escaped selector,
never arbitrary URI forwarding. Status remains original text plus HTTP metadata.
Named 400/404/500 responses remain errors in Go, with raw text in the response.
No structured per-team model or inference that Completed means success is made.
No List/Update/Delete/cancel contract is established for this orchestration.

See [tests](../../../tests/integration/lab)
and [unit tests](service_test.go).
Documentation-only service limits, exact grants/cloud deployment, live behavior
and consequential acceptance remain separately qualified. This candidate has
made no tenant request or notification email and claims no complete issue parity.

## Published input layout

The [deployment guide](https://learn.microsoft.com/en-us/microsoft-365/frontline/deploy-teams-at-scale),
reviewed on 2026-11-26, specifies ordered columns for Teams.csv and Users.csv.
The [unit tests](service_test.go)
records those columns and current acceptance prerequisites. Owners need a Teams
license. This documents caller inputs; BuildArchive still preserves original bytes
and does not claim local validation of the service's CSV grammar or a typed
per-team status schema. Accepted deployment and recovery need an explicit owned
licensed-owner/recipient fixture; no orchestration is submitted for discovery.
