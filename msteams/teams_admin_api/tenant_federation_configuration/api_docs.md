# Tenant federation configuration

`Client.TeamsAdminAPI.TenantFederationConfigurations` implements the modern generated
ConfigAPI path for TenantFederationSettings. GetByIdentity/List require 200;
UpdateByIdentity/UpdateGlobal send sparse PUT and require 204. ListFiltered composes
the bare list with local Identity/Tag wildcard matching. It uses shared same-origin
forest discovery and the Teams resource token. Create/Delete, paging and PATCH
are not established for this resource.

UpdateGlobal explicitly names the omitted-Identity Global default. The primitive
requires an identity escaped as one complete segment. Nil scalar pointers omit;
pointers to false/empty strings send those values. Strings remain open; no default,
enum restriction or domain limit is invented.

NewDomainPattern is a local Domain value with a required nonempty string; it does
not prove DNS syntax, ownership or reachability. NewAllowList copies strings,
preserving order, duplicates and empties; nil/omitted input emits [].
NewAllowAllKnownDomains emits ["AllowAllKnownDomains"]. AllowedDomains JSON is the
flattened request value, distinct from the PowerShell formatting wrapper.

AllowedDomains and AllowedDomainsAsAList are mutually exclusive; the latter's wire
key is AllowedDomains. ReplaceDomains/ChangeDomains emit replacement versus
Add/Remove shapes. Empty operators are omitted; both empty sends {}.
ReplaceDomainValues/ChangeDomainValues preserve mixed strings/legacy objects after
removing only serialization metadata, without asserting arbitrary object acceptance.
Nil list fields omit; invalid zero choices/changes reject before discovery.

Response choice/discriminator semantics remain unproven. Configuration retains raw
AllowedDomains, BlockedDomains, AllowedTrialTenantDomains and all received fields.
Nullable values are not translated into asserted allow-all defaults. Marshaling a
snapshot preserves omission/null/future data; separate write requests prevent an
accidental full-object update.

Updates send once without implicit read-back, rollback or propagation wait. Shared
transport supplies context, cancellation, raw metadata, redacted errors and token/
destination isolation. Modern flight/legacy eligibility, confirmation semantics
and actual permissions remain separately qualified. [unit tests](service_test.go).
