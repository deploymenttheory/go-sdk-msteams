# Microsoft Teams Go SDK

A native Go SDK project targeting the management behaviour of MicrosoftTeams PowerShell **8.1.0**, with Linux, macOS and Windows support and portable CI/CD authentication.

**Status:** baseline discovery and the Go module foundation are implemented; local client construction and session ownership are implemented. Authentication providers and resource operations are still under development.

- [Agreed implementation plan](docs/implementation/plan.md)
- [Published GitHub backlog](https://github.com/deploymenttheory/go-sdk-msteams/issues/6)
- [Approved backlog preview](docs/implementation/backlog/README.md)
- [Publication verification](docs/implementation/backlog/publication/README.md)
- [Command coverage and ownership](docs/implementation/backlog/coverage.md)
- [Pinned PowerShell discovery evidence](contracts/powershell/8.1.0/README.md)
- [Reproduce the pinned baseline](tools/baseline/README.md)
- [Compare upstream contract changes](tools/contractdiff/README.md)
- [Discover releases and review candidate draft PRs](tools/releases/README.md)

The architecture follows the [Jamf Pro Go SDK](https://github.com/deploymenttheory/go-sdk-jamfpro-v2): a root client, shared transport, resource-specific services, typed models, examples and acceptance tests. PowerShell is a discovery and comparison tool, not a planned runtime dependency.

The first supported release requires complete implementation and contract-test coverage of the reconciled management baseline. Available tenant capabilities will also receive live conformance tests; unavailable live scenarios remain explicitly identified.

## Development

Requires the pinned Go version in [go.mod](go.mod) (currently 1.27.4). The module is
`github.com/deploymenttheory/go-sdk-msteams`; import the SDK entry point as
`github.com/deploymenttheory/go-sdk-msteams/msteams`.

```sh
go build ./...
go vet ./...
go test -race -count=1 -shuffle=on ./...
go tool govulncheck -test ./...
go run ./examples/baseline
go run ./examples/client-construction
```

Both examples need no credentials. The client example constructs and closes a local session without signing in. The baseline example prints the target contract version;
it does not demonstrate an implemented Teams operation. See the
[foundation guide](docs/guides/development.md) and [client configuration guide](docs/guides/client-configuration.md) for pinned dependencies, CI gates and
package ownership. CI runs on Linux, macOS and Windows without tenant secrets.

Planned layout, following the agreed Jamf composition:

```text
go.mod, go.sum           module metadata only; no Go source at repository root
msteams/                 SDK entry point, client composition and package tests
  client.go, doc.go       public client and package documentation
  client/                request boundary and shared transport
  auth/                  portable audience-aware providers
  config/, constants/    validated configuration and cloud definitions
  shared/                narrowly scoped wire/model primitives
  graph_api/<resource>/  typed Graph services
  teams_admin_api/<resource>/  typed administration services and local builders
  mocks/, acceptance/    offline doubles and opt-in tenant conformance
examples/                compiled usage examples
contracts/powershell/     immutable upstream evidence and coverage
internal/dependencycheck/ cross-platform dependency compilation gate
```

## Review and validate the backlog

```sh
python3 tools/backlog/build.py
python3 tools/backlog/validate.py
```

These commands operate locally and do not publish GitHub issues. The publication preview includes full issue bodies, metadata, command ownership and native relationship requirements. Publication follows review of that exact set.

## Licence

See [LICENSE](LICENSE). This community project is not affiliated with Microsoft.
