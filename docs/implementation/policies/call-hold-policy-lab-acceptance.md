# Call hold policy lab acceptance

`TestLabCallHoldPolicyReadOnly` requires `MSTEAMS_LAB_READ_ONLY=1` before it
reads the selected private identity file. It reads compiled Global and a Global
filtered list, selects one exact case-insensitive Global identity (allowing
Tag:Global as a separate match), and compares the four nullable string getters.
It performs no policy, assignment, directory, role or consent writes.

Use `MSTEAMS_LAB_IDENTITY_FILE` and `MSTEAMS_LAB_AUTH_METHOD=client-secret` or
`certificate`, with mutation/consequential opt-ins unset, then run:

```sh
go test -count=1 -run '^TestLabCallHoldPolicyReadOnly$' -v ./tests/integration/lab
```

Both actual credential modes returned resource HTTP 403 on Global on
2026-11-21. Only sanitized status was logged; no response payload was retained.
Filter/list and optional pinned comparison were not reached. No policy, audio
configuration, playback, assignment, role or consent changed. No successful
tenant qualification is inferred from denied reads or offline tests.

Optional comparison additionally requires `MSTEAMS_LAB_POWERSHELL_ORACLE=1`
and an explicit pinned `MSTEAMS_LAB_MODULE_PATH`. The certificate oracle checks
the manifest/Core/PolicyRp hashes and tenant binding, invokes only compiled Get
Global, requires all four getter properties and disconnects. It neither invokes
modern Grant nor requests flight/role changes. The bounded decoder rejects
mixed/trailing output, unknown outer fields, wrong tenant/identity and incomplete
or non-string/non-null projections. Missing native wire properties use the
nullable Core getter presentation while complete raw snapshots remain available.

Successful authorized reads and optional comparison remain open. Mutation
qualification must separately cover non-default CRUD read-back, omission/null,
cleanup, modern grant completion, cloud/flight availability and effective
streaming/audio playback. These are not established by HTTP acceptance alone.
