# Implementation plan amendments

The [original plan](plan.md) and publication manifest are preserved as the exact
approved publication evidence. These subsequent user instructions apply to ongoing
implementation. GitHub issues/PRs and the coverage ledger record current delivery
state; the original plan's initial status text is historical.

## 2026-11-20: resource CRUD SDK

Expose supported create, read, update, delete and list operations with typed models,
plus explicit resource-specific actions. Compose these operations when needed to
reproduce MicrosoftTeams PowerShell behavior. The user explicitly accepts such
compositions; the public Go interface does not need to copy cmdlet names or make
every cmdlet a single method.

Every resource contract and implementation includes a CRUD/list/action matrix and
command-to-operation mappings. Supporting operations needed for compositions are in
scope even without a same-named cmdlet. Read and list are distinct. Do not invent an
unsupported operation, route or stub merely to complete a matrix. Record actual
service constraints and preserve local-only builders as pure constructors.

Use `Create`, `GetByID`, `UpdateByID`, `DeleteByID` and `List` for identified resources;
use the corresponding `ByIdentity` methods for identity-based resources and `Get` /
`Update` for singletons. Give actions descriptive names, such as `Archive`,
`Unarchive`, `UploadPicture`, `PromoteOwner` and `DemoteOwner`.

Composed operations expose partial results, created resource/operation identities
and failures. Compensation has an explicit ownership boundary; it never deletes
an existing caller-supplied resource. Test the primitive operations and their
compositions, and link all constituent symbols/tests from the capability ledger.

[Decision 0001](../decisions/0001-client-and-service-composition.md) carries this rule
into the shared architecture. All existing transport, authentication, evidence,
replay-safety and complete-release requirements continue to apply.
