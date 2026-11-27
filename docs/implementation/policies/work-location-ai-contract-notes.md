# Work-location and AI policy contracts in progress

Owners: #156 (POL-83), #157 (POL-84). Baseline: MicrosoftTeams 8.1.0.

Checked runtime resolves all assigned Get/New/Set/Remove/Grant commands as
compiled Cmdlets. Each family has nine public parameter sets, including the
compiled four-set Grant shape with required signed Int32 group rank and tenant
Force. CmdletConfig has no modern mapping for these two families. The unused
parameterless Identity Grant set reports an error, as established by the shared
BaseTpmGrantCmdlet.

TeamsWorkLocationDetectionPolicy has nullable EnableWorkLocationDetection and
opaque nullable UserSettingsDefault. TeamsAIPolicy has five nullable opaque
strings: Description, EnrollFace, EnrollVoice, SpeakerAttributionForBYOD and
PassiveVoiceEnrollment. Inspected model setters do not execute enum/range rules;
description length constants alone do not prove throwing validation. New/Set
property mapping and provider type normalization still require review.

Next: explicit native CRUD/list/grant, complete typed models, actual owned lab CRUD,
pinned getter comparisons, independent cleanup, permanent sanitized responses
consumed by units, docs/examples and all parameter-set mappings. No implementation
or live success is claimed for this pair yet. Tests will use uniquely owned
unassigned policies; no existing user location or enrollment setting is selected.

Sources inspected: checked runtime.json and CmdletConfig; pinned Core
TeamsWorkLocationDetectionPolicy and TeamsAIPolicy models. Preceding PR #297
contains voice-applications/workload; PR #296 repairs the invalid cross-request
XML Key equality assumption while preserving each complete raw response.
