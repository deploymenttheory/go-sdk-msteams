# Teams messaging policy qualification

Baseline: MicrosoftTeams 8.1.0. Implementation: issue #134.

The pinned `Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll` model and
New/Set cmdlets define 39 settings. The public five-command surface has
eight parameter sets: Get Identity/Filter; Grant Identity/GrantToTenant/GrantToGroup;
New, Set and Remove Identity. CRUD uses `BaseTpm*` and PolicyRP `HttpRequestHelper`.
Public Grant dispatches through `CmdletConfig.json`,
`Merged_custom_PsExt.ps1::Grant-CsTeamsPolicy` and generated user/tenant/group helpers.
The SDK preserves this dispatch distinction. It does not execute PowerShell.

Pinned assembly SHA256:

- Core: `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49`
- PolicyRP provider: `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51`
- ConfigAPI: `8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6`

On 2026-11-22, the dedicated lab SPN with Teams Administrator completed Global
read/list and owned policy create/read/update/list/delete under certificate and
client-secret authentication. All 39 getters were compared to the pinned
PowerShell output after create and update. Description-only updates preserved
`AllowGiphy: false`. Cleanup independently verified GET404 and list absence.
The permanent sanitized response bodies are in the service's `testdata` directory
and are used directly by its unit tests. There is no additional response store.

These tests create only their uniquely named, unassigned policies. They do not
change Global, existing policies, assignments, directory roles or licenses.
The existing SDK SPN permissions suffice; no permission changes were made.

CI selects both acceptance tests and both authentication methods. The CRUD proof
requires create/read-back, update/read-back plus omission, list and deletion
milestones, two pinned comparisons and a resolved cleanup entry for every test.
This is selected-operation proof; issue #134 remains open for full criteria.

Still unqualified: live user/group/tenant grant and clear, assignment propagation,
effective meeting/chat behavior, the remaining bound/null/Force scenarios,
delegated flows, national clouds and flight/licensing differences. Unit tests
validate every source setting name/type and inject failures for native request
contracts; these are separate from live acceptance results.
