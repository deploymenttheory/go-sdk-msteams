# Audio conferencing custom prompts lab acceptance

`TestAccAudioConferencingPromptsReadOnly` requires `MSTEAMS_LAB_READ_ONLY=1`
before reading an explicitly selected private identity file. It reads Global,
then a Global-filtered collection, selects a unique case-insensitive exact
Global identity (without confusing Tag:Global) and compares known projections.
It performs no prompt, package, role, consent or directory writes.

Use `MSTEAMS_LAB_AUTH_METHOD=client-secret` or `certificate` and
`MSTEAMS_LAB_IDENTITY_FILE` pointing to the private lab identity. Keep mutation
and consequential opt-ins unset. Normal repository tests skip live reads.
The selected family can be run with:

```sh
go test -count=1 -run '^TestAccAudioConferencingPromptsReadOnly$' -v ./tests/integration/lab
```

On 2026-11-21 both actual credential modes reached the selected resource and
returned HTTP 403 on Global. The test failed with only a sanitized status; no
response payload was retained. Filter/list and optional PowerShell comparison
were not reached. No tenant state changed and no privilege grant was applied.

Optional comparison additionally requires `MSTEAMS_LAB_POWERSHELL_ORACLE=1` and
an explicit `MSTEAMS_LAB_MODULE_PATH` containing pinned MicrosoftTeams 8.1.0.
The certificate oracle checks manifest/Core/PolicyRp hashes and tenant binding,
reads only Global, requires all source projection properties, and disconnects.
Its bounded output is schema-checked for both lists, all seven prompt fields,
all ten package fields and non-null reference arrays before comparison. Null
elements and reference strings remain distinct; list order is not normalized.
Mixed/trailing output, wrong tenant/identity and incomplete/type-invalid data
fail without output retention.

Successful authorization and optional oracle comparison remain outstanding.
Mutation qualification must separately cover non-default read-back, omission,
empty/null semantics, package removal constraints, restoration/cleanup and
effective audio playback. Offline source/structural evidence does not establish
those outcomes or complete command parity.
