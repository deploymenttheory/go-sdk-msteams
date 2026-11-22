# Room video and shared calling qualification

Owners: #142 (POL-69) and #143 (POL-70). Baseline: MicrosoftTeams 8.1.0.

Room video's public Functions map to modern ConfigAPI Get/New/Set/Remove helpers
and Grant-CsTeamsPolicy. Its seven settings are six strings plus a non-nullable
public Boolean. Three grant sets select user, group with nullable Int32 rank, or
tenant; PassThru controls output, and Force is not a public parameter. The native
adapter follows discovery and whole-name escaping, with POST201/PUT204/bodyless
DELETE204 and separately retained submission/read-back results.

Shared calling is compiled Core/PolicyRP: nullable Description/ResourceAccount,
a constructor-initialized emergency string list, local PSListModifier processing,
Get Identity/Filter, New/Set/Remove Identity, and four compiled grant sets. The
list setter copies via ToList and rejects null; New/Set clear null/empty bound
modifiers or apply them to the current list. The static phone regex is not
executed by the inspected setter, so no invented validation is added.

Primary local sources are the public proxies, CmdletConfig, merged modern helpers,
generated ConfigAPI methods, shared-calling model/cmdlets and BaseTpm/provider
methods. Oracles verify the pinned module hashes already documented in the
network/notification qualification and check version/tenant binding.

On 2026-11-22 certificate and client-secret tests passed Global read/list and owned
CRUD for both resources. Room video creates a non-default description with
Enabled=false, changes only the description, preserves the Boolean and compares
all seven getters twice with the pinned module. Shared calling creates an
unassociated policy with a non-default description and explicit empty emergency
list, changes only description, retains the list and compares all three getters.
Both list, delete and require independent GET404 and complete collection absence.
Every create intent is journaled beforehand; bound recovery handlers are registered.

A populated shared-calling attempt using emergency dial codes returned HTTP400 /
PolicyRP40013. Those codes are not valid callback-number fixtures. Authorized
Get-CsOnlineApplicationInstance and Get-CsPhoneNumberAssignment reads succeeded
and returned zero accounts and zero numbers; policy inventory contained Global
only. Private inventory data remains in the user-selected lab directory. No
numbers, accounts, licenses or assignments were provisioned or changed.

The full populated-list test requires an explicit bound resource-account/number
fixture, and remains registered as pending. It retains create/read, sparse
preservation, Add/Remove, clear, list and delete. It will not count as a passing
empty-list case. Callback numbers are sanitized before writing permanent fixtures.
The lab fixture question is pending; no credentials are missing.

Eighteen permanent sanitized response files from the available successful cases
supply unit decoding, snapshot, request, CRUD and actual-error checks. Decoder
edge vectors and controlled failures test list/null/binding/replay behavior without
inventing successful server responses. Unit grant checks cover exact bodies and
rank behavior; grant persistence/propagation is not qualified.

[Microsoft's shared-calling guide](https://learn.microsoft.com/en-us/microsoftteams/shared-calling-setup)
requires actual callback phone inventory, matching type/country and resource-account
configuration. Remaining live scope includes populated shared calling, room call
settings beyond the tested Boolean, scoped grants, effective calls, delegated modes,
locale-specific filters and other clouds. Selected-operation CI does not close
these whole-ticket or release requirements.
