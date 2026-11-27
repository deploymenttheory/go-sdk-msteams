# Privacy configuration

`Client.TeamsAdminAPI.PrivacyConfigurations` manages the compiled
`PrivacyConfiguration` resource with native Go methods.

| Method | Baseline command / parameter set | Behavior |
|---|---|---|
| GetByIdentity | Get-CsPrivacyConfiguration / Identity | Read one explicit identity |
| List | Get-CsPrivacyConfiguration / omitted Identity | Read the bare collection; exact Global first |
| ListFiltered | Get-CsPrivacyConfiguration / Filter | List then apply local Identity/Tag wildcard matching |
| PatchByIdentity | Set-CsPrivacyConfiguration / Identity | Send one sparse PATCH |
| UpdateByIdentity | Set-CsPrivacyConfiguration / Identity | Read current, apply selected settings, PATCH once |

The four nullable Boolean settings are EnablePrivacyMode, AutoInitiateContacts,
PublishLocationDataDefault and DisplayPublishedPhotoDefault. Use `policy.Set(false)`
to bind false, `policy.Null[bool]()` to bind null, and the zero Field to omit a
setting. At least one setting must be bound; Force alone rejects locally before
credentials or HTTP. Force, when supplied, uses the compiled `"True"`/`"False"`
body convention. Inputs use Go Boolean types rather than arbitrary shell coercion.
ListFiltered rejects a bound empty pattern; use List for an omitted filter. A
whitespace pattern stays a literal filter, following the source's IsNullOrEmpty rule.

Snapshots retain received absence, null, false and future fields. Decode requires
typed Boolean values; raw snapshots are intended as received data, while Settings
express mutations. Ordinary formatting is redacted. Native PATCH defaults body
Identity to the requested route; Update uses the preceding read model's identity,
including null when absent, without changing the caller's route.

UpdateResult retains Before, ReadResponse, PatchInvoked and HTTPAccepted if
preparation, submission or decoding fails. HTTP acceptance does not prove persisted
or effective privacy behavior. No implicit read-back, polling, rollback, mutation
replay or atomicity/ETag guarantee is added. A 204 write has no fabricated result.

Requests use the selected Teams origin/audience and direct compiled-policy headers:
MPACmdlet, command name and fresh UUID correlation. Scope/name segments follow the
pinned escaping rule. No generated forest bootstrap or Graph identity lookup runs.
The baseline establishes Get and Set; it provides no Create, Delete or assignment
command for this family.

Application/delegated synthetic fixtures prove routing with the selected mode,
not tenant RBAC, cloud availability or production parity. Teams administrative
authorization is required independently of token acquisition; exact least-privilege
role qualification remains open. See [unit tests](service_test.go)
and the [lab plan](../../../docs/acceptance/privacy-configuration-lab-acceptance.md).
