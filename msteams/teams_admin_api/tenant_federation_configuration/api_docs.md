# Tenant federation domain values

This initial package exposes local domain constructors. Resource Get/List/Update
and the remaining federation services are tracked in issue #96 and are not yet
part of this checkpoint.

`NewDomainPattern(domain)` returns a local Domain value and rejects a missing
mandatory string. It performs no DNS or ownership validation. `NewAllowList`
copies supplied string domains, retaining order, duplicate and empty values;
omitted input is an explicit empty list. Existing DomainPattern values can supply
their Domain field. `NewAllowAllKnownDomains` creates the request sentinel.

The AllowedDomains JSON value is the flattened native array consumed by modern
tenant Set, including `["AllowAllKnownDomains"]` for the sentinel and `[]` for
an empty allow-list. It is distinct from the PowerShell presentation object.
The invalid zero choice returns ErrInput. Marshaling a choice performs no HTTP
or token acquisition, and nullable response choice semantics are not inferred.

See [implementation progress and provenance](../../../docs/implementation/policies/federation-sdk-progress.md).
