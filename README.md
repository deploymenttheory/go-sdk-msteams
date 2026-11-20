# Microsoft Teams Go SDK

A native Go SDK project targeting the management behaviour of MicrosoftTeams PowerShell **8.1.0**, with Linux, macOS and Windows support and portable CI/CD authentication.

**Status:** baseline acquisition and discovery tooling is implemented; the SDK client and resource services are not implemented yet.

- [Agreed implementation plan](docs/implementation/plan.md)
- [Published GitHub backlog](https://github.com/deploymenttheory/go-sdk-msteams/issues/6)
- [Approved backlog preview](docs/implementation/backlog/README.md)
- [Publication verification](docs/implementation/backlog/publication/README.md)
- [Command coverage and ownership](docs/implementation/backlog/coverage.md)
- [Pinned PowerShell discovery evidence](contracts/powershell/8.1.0/README.md)
- [Reproduce the pinned baseline](tools/baseline/README.md)

The architecture follows the [Jamf Pro Go SDK](https://github.com/deploymenttheory/go-sdk-jamfpro-v2): a root client, shared transport, resource-specific services, typed models, examples and acceptance tests. PowerShell is a discovery and comparison tool, not a planned runtime dependency.

The first supported release requires complete implementation and contract-test coverage of the reconciled management baseline. Available tenant capabilities will also receive live conformance tests; unavailable live scenarios remain explicitly identified.

## Review and validate the backlog

```sh
python3 tools/backlog/build.py
python3 tools/backlog/validate.py
```

These commands operate locally and do not publish GitHub issues. The publication preview includes full issue bodies, metadata, command ownership and native relationship requirements. Publication follows review of that exact set.

## Licence

See [LICENSE](LICENSE). This community project is not affiliated with Microsoft.
