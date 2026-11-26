# Microsoft Teams SDK documentation

The SDK targets MicrosoftTeams PowerShell 8.1.0 management behaviour with a native
Go runtime. Authentication, transport and resource services are implemented in
stages; complete parity and release qualification remain outstanding.

## Using the SDK

- [Quick start and current project status](../README.md)
- [Client configuration and environment variables](guides/client-configuration.md)
- [Authentication methods, permissions and cloud limitations](guides/authentication.md)
- [Graph services](../msteams/graph_api/services.go), including [teams](../msteams/graph_api/team/api_docs.md), [channels](../msteams/graph_api/channel/api_docs.md) and [membership](../msteams/graph_api/team_member/api_docs.md)
- [Teams administration services](../msteams/teams_admin_api/services.go), including [meeting policies](../msteams/teams_admin_api/teams_meeting_policy/api_docs.md), [call queues](../msteams/teams_admin_api/call_queues/api_docs.md) and [applications](../msteams/teams_admin_api/unified_apps/api_docs.md)

Resource packages contain `api_docs.md` and compiled `example_test.go` examples.
Their references describe supported CRUD/actions, permissions, command mappings
and live-test limitations. Request and response models belong to each resource.

## Developing and verifying

- [Go toolchain, dependencies and quality gates](guides/development.md)
- [Service implementation and public compatibility conventions](guides/service-development.md)
- [Offline contract harness](../internal/contracttest/README.md)
- [Tenant acceptance, fixtures, cleanup and selective skips](../tools/acceptance/README.md)
- [Current capability and evidence ledger](../contracts/coverage/README.md)
- [GitHub product baseline and backlog](https://github.com/deploymenttheory/go-sdk-msteams/issues/6)

Offline tests use `TestUnit`; credential-backed tenant tests use `TestAcc`.
Run offline tests with `go test -run '^(TestUnit|Example)' ./...`. Live acceptance
requires the explicit lab setup described above. A skip or expected denial does
not qualify a successful resource operation.

## Architecture and compatibility

- [Client and service composition](decisions/0001-client-and-service-composition.md)
- [Portable authentication](decisions/0002-portable-authentication.md)
- [Backend contracts, replay safety and compatibility](decisions/0003-backend-contracts-and-compatibility.md)
- [Agreed implementation baseline](implementation/plan.md) and [subsequent amendments](implementation/plan-amendments.md)
- [Pinned upstream evidence](../contracts/powershell/8.1.0/README.md)
- [Upstream contract comparison](../tools/contractdiff/README.md)
- [Candidate release discovery and review](../tools/releases/README.md)

The original plan and publication preview preserve historical approval evidence.
Use GitHub and the coverage ledger for delivery status. SDK semantic versions are
independent of the PowerShell baseline. Commercial lab results do not establish
sovereign-cloud qualification.

Current backlog work covers CALL, CORE, EXT, POL and LAB. VOICE administration is
deferred because the lab lacks licensing for live testing; full-parity release
requirements still include it.

## Contributing and licence

See [CONTRIBUTING.md](../CONTRIBUTING.md) and [LICENSE](../LICENSE).
