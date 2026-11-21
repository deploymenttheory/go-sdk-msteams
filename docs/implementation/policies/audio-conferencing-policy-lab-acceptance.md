# Audio conferencing policy lab acceptance

`TestLabAudioConferencingPolicyReadOnly` requires `MSTEAMS_LAB_READ_ONLY=1`
before reading the explicitly selected private identity. It reads Global and a
Global-filtered collection, selects one exact case-insensitive Global identity
without confusing Tag:Global, and compares the two known fields. It never
creates, updates, deletes, assigns, changes roles or grants consent.

Select `MSTEAMS_LAB_IDENTITY_FILE` and `MSTEAMS_LAB_AUTH_METHOD=client-secret`
or `certificate`, with mutation/consequential opt-ins unset, then run:

```sh
go test -count=1 -run '^TestLabAudioConferencingPolicyReadOnly$' -v ./tests/integration/lab
```

Both actual credential modes returned resource HTTP 403 on Global on
2026-11-21. Only sanitized status was logged; no payload or tenant state was
retained. Filter/list and optional pinned comparison were not reached. No
privilege changes were applied. Live Get remains unavailable; writes/read-back,
group completion, server list operators, cleanup and effective invites remain
separately unqualified.

Optional comparison additionally requires `MSTEAMS_LAB_POWERSHELL_ORACLE=1`
and an explicit pinned `MSTEAMS_LAB_MODULE_PATH`. The certificate script checks
the manifest, private assembly, command mapping, wrapper and export hashes and
tenant binding. It requests the source ForceModern processing mode explicitly,
so the comparison selects the mapped REST path. It reads only Global and
requires both projection properties before returning a bounded JSON report.

The decoder rejects unknown/trailing outer output, wrong tenant/identity,
missing fields, non-Boolean toll-free values and non-array/non-null phone data.
Array contents and nulls remain available for source-compatible future values;
JSON comparisons use exact numbers and retain ordering. No coercion or effective
feature behavior is inferred from offline or denied requests.
