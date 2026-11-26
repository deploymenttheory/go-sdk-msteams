# POL91 GuestMeeting and Meeting read correction

The last two compiled configuration families remain unqualified while their full
command decisions are reviewed. The [portable own-model probe](../../../scripts/discovery/observe-pol91-last-compiled-read.ps1)
checks the pinned Core and PolicyRp hashes before import and records all four
public compiled declarations, exact sets/effective attributes, normalized names
and concrete DTO properties. Its [725 observations](evidence/pol91-last-compiled-read.json)
cover GuestMeeting (160), Meeting configuration (433) and its nested published
Entra context (132). Only data-model reflection/deserialization ran; no management
lifecycle, authentication or HTTP ran in the probe.

Explicit nullable read adapters preserve source string/Boolean conversion and
all six checked UInt32 properties. The UInt32 reader reuses the separately proven
UInt64 conversion within a checked UInt32 bound; integer tokens retain precision,
floats use source ties-to-even conversion, and strings/Booleans/null follow the
actual per-property observations. All six own DTO setters additionally reject zero (the source positiveInteger
constraint) while accepting the full nonzero UInt32 range. A resource-local
reader enforces this after conversion. No 65535 ceiling, cross-field port-range,
URL or enum constraint is invented.

Meeting's getter-only context list starts empty, appends repeated arrays in wire
order, ignores null and primitive assignments, and rejects object assignments.
Each context Id setter is validated as it occurs: omission retains the null
constructor default, explicit null fails, and only c1 through c99 with the
source's optional single final LF succeeds. A later alias cannot hide an invalid
assignment. Existing typed write validation stays distinct from this read rule.
Raw snapshots are retained independently of the typed complete projections.

Public model tests replay every own observation through the actual read types
and compare a recursive typed projection without invoking raw-view MarshalJSON.
Root-client reads assert the typed values, exact numeric strings/rounding, list
ordering/null members, untouched future fields and HTTP receipt. Invalid context
Id preparation tests retain the 200 response and stop before PATCH.

The paired Meeting write correction has its own helper/serializer source probe
and explicit known-field write DTO; raw read JSON is an inspection view. These
changes have not yet passed the combined affected check at this checkpoint and
restore no qualification or live state. The first private probe failed during
PowerShell tuple construction before observations; the corrected producer exited
zero and produced the versioned complete artifact.

The initial combined check at `4e7a06d` exposed the missing resource-specific
positiveInteger restriction in the new read adapter: six zero-valued variants
for each of six fields were incorrectly accepted. That real SDK gap is retained
in the failed execution artifact; expected source observations were unchanged.
The corrected resource reader rejects zero after UInt32 conversion.
