# SIP devices and team template permission qualification

Owners: #146 (POL-73), #147 (POL-74). Baseline: MicrosoftTeams 8.1.0.

Both public surfaces are compiled Core/PolicyRP. SIP has Get/Set and nullable
BulkSignIn; team template permission has Get/New/Set/Remove, nullable Description
and a get-only, constructor-initialized HiddenTemplates list. The Id setter rejects
explicit null; list modifiers use ordinal Id equality, first-match removal and
append without deduplication. Bound null/empty modifiers clear. The nested provider
resolver serializes Id/Identity only for list objects. No scalar enum/length rules
or extra public Grant/New/Delete operations are invented.

Evidence: checked runtime signatures, Core classes/cmdlets, PSListModifierHelper,
BaseTpm/provider and BoundParametersContractResolver. Oracles verify the pinned
module/Core/provider hashes already documented in the neighboring qualifications.
Current Microsoft references: [SIP](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/get-csteamssipdevicesconfiguration?view=teams-ps),
[template permission](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamstemplatepermissionpolicy?view=teams-ps).

Both certificate/client-secret modes passed Global reads/list and native/pinned
getter comparison on 2026-11-22. SIP toggled its observed Disabled original to
Enabled, compared before/after/restoration, restored the private journal snapshot
and independently verified Disabled. It does not qualify device sign-in behavior.

Team template tests ran empty and populated owned CRUD. Populated input IDs were
selected from a complete successful tenant team-template inventory read, not from
invented strings or meeting templates. A description-only PATCH preserves the list;
Add/Remove replaces its first ID with a second real ID; clear sends an empty list.
Four pinned comparisons verify getters, with list/delete and independent GET404/
list absence. Empty-list CRUD runs separately with two comparisons. No existing
policy, template, Global permission or assignment was changed.

The first populated attempt exposed a decoder bug: its received HiddenTemplates
array contained an entry but the typed projection ignored it. A unit test using
the actual sanitized response failed, the field match was corrected, and populated
CRUD then passed in both modes. The failed attempt's intent was also independently
cleaned; all seven intents/restorations from these local attempts are clean.
Twenty-five permanent response fixtures cover reads, both CRUD variants, list edits,
clear, deletion and restoration. Successful and failed cases do not overwrite each
other. No separate response store remains. Unit replay verifies actual request/
response behavior and complete snapshots; source-derived fault vectors remain
separate from actual successful response bodies.

CI registers template reads, basic owned CRUD, a separate populated suite requiring
edit/clear plus four comparisons, and serialized SIP reads/update/restoration. The
normal repository secrets suffice. The only configured skip remains shared-calling
telephone inventory; neither new service scenario is skipped. Remaining live scope
includes null/empty scalar persistence, other setting combinations, effective user/
device behavior, delegated modes and other clouds. Selected operations do not
establish full release qualification.
