# Microsoft Teams SDK documentation

The SDK provides native Go clients for Microsoft Graph and Teams administration.
Use the resource API docs, compiled examples and actual test results to assess
supported functionality.

## Using the SDK

- [Quick start](../README.md)
- [Client configuration](guides/client-configuration.md)
- [Authentication and permissions](guides/authentication.md)
- [Graph services](../msteams/graph_api/services.go)
- [Teams administration services](../msteams/teams_admin_api/services.go)

Resource packages contain `api_docs.md`, models, JSON response fixtures and
compiled examples. SDK use requires no PowerShell installation.

## Developing and testing

- [Development setup and quality checks](guides/development.md)
- [Service development](guides/service-development.md)
- [Unit tests, JSON fixtures and tenant acceptance](guides/testing.md)
- [Tenant setup, cleanup and recovery](../tools/acceptance/README.md)
- [Current GitHub backlog](https://github.com/deploymenttheory/go-sdk-msteams/issues/6)

## Architecture

- [Client and service composition](decisions/0001-client-and-service-composition.md)
- [Portable authentication](decisions/0002-portable-authentication.md)
- [Backend routing, retries and testing](decisions/0003-backend-routing-and-testing.md)

Current delivery covers CALL, CORE, EXT, POL and LAB. Voice administration remains
deferred because licensing is unavailable. Commercial tenant results do not prove
sovereign-cloud support. There is no separate contract-research or command-ledger
release gate.
