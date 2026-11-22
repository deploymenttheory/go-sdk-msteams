# Call park policy lab acceptance

`TestAccCallParkPolicyReadOnly` requires `MSTEAMS_LAB_READ_ONLY=1` before it
reads an explicitly selected private identity. It reads Global and a Global
filtered list, selects one exact case-insensitive Global identity without
confusing Tag:Global, and compares all five known properties with exact numbers.
It performs no policy, assignment, role, consent or directory writes.

Use `MSTEAMS_LAB_IDENTITY_FILE` and `MSTEAMS_LAB_AUTH_METHOD=client-secret` or
`certificate`, with mutation/consequential opt-ins unset, then run:

```sh
go test -count=1 -run '^TestAccCallParkPolicyReadOnly$' -v ./tests/integration/lab/teams_call_park_policy
```

Both actual credential modes returned resource HTTP 403 on Global on
2026-11-21. Only sanitized status was logged; no response payload was retained.
Filter/list and optional pinned comparison were not reached. No policy,
assignment, call parking setting, role or consent changed.

Optional comparison requires `MSTEAMS_LAB_POWERSHELL_ORACLE=1` and an explicit
pinned `MSTEAMS_LAB_MODULE_PATH`. The certificate oracle checks the manifest,
private assembly, command-map, wrapper and export hashes and tenant binding.
It selects ForceModern explicitly, reads only Global, requires all five source
properties and disconnects. Its bounded decoder rejects unknown/trailing outer
output, wrong tenant/identity, missing fields, invalid scalar types, Int64
overflow and fractional values. Canonical comparison uses exact JSON numbers,
retaining zero/null/omission distinctions and all original raw snapshots.

Successful authorized reads and optional comparison remain open. Mutation
qualification must separately establish pickup/timeout validation, non-default
read-back, omission behavior, cleanup, group completion, cloud/flight availability
and effective call parking. Offline checks and HTTP acceptance do not establish
those outcomes.
