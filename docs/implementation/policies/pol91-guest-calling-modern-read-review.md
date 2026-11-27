# GuestCalling conditional modern read review

This bounded investigation covers the two unqualified GuestCalling functions
within #185. It changes no Go behavior, qualification, live record or legacy
contract. Actual public Get/Set resolve to proxy Functions; the matching private
runtime types derive from TrpsPSCmdlet. Available compiled Core types are not
therefore evidence that the public default uses compiled PolicyRp decoding.
The CmdletConfig Remoting mapping names Get/Set-CsConfigurationModern with
ConfigType=TeamsGuestCallingConfiguration; the existing
[conditional dispatch research](pol04-primary-source-audit.md#conditional-dispatch-and-the-precise-legacy-boundary)
keeps session-flight selection and the unknown remote contract separate.

The [portable own observer](../../../scripts/discovery/observe-pol91-guest-calling-modern-read.ps1)
AST-extracts the actual five helpers from hash-checked Merged_custom_PsExt.ps1,
then executes only JSON/PSSerializer conversion, own GuestCalling formatting and
stable Global-first sorting. It does not execute telemetry, the management
helper, flight selection, a generated HTTP command or a tenant request. There
is no Broadcast-only exposure selector in this family.

The [33 observations](evidence/pol91-guest-calling-modern-read.json) have 29
complete results and four isolated-stage exceptions: three null/empty-array
format-input failures and one case-collision JSON failure. Identity and
AllowPrivateCalling each receive eleven token shapes. AllowPrivateCalling:1,
"text" and {} retain number/string/object values rather than becoming Boolean
or null. Unknown nested data survives. The current Go decoder instead requires
*bool and returns ErrDecode before returning its RawFields/RawJSON for these
inputs. That is an explicit unresolved dynamic-result versus typed-view
boundary, not evidence that compiled ReadBool coercion should be applied.

These are downstream stage observations, not complete accepted HTTP roots.
The [shared generated Xds investigation](pol91-broadcast-modern-read-review.md#remaining-contract-work)
shows why that boundary matters: its parser rejects exact duplicates before
ConvertFrom-Json's apparent last-wins stage, and nonobject roots have their own
parser/callback behavior. Selected/list callbacks, statuses, headers and the
actual GuestCalling wrapper argument path still need closure before a service
change or qualification decision.

The public proxy has optional String Identity/Filter and a nonnullable Boolean
AllowPrivateCalling on Set, without extra value validators. Native omission
versus false/true is preserved. The modern generic helper defaults Set Identity
to Global only when the property-bag key is absent. Get's internal helper omits
ConfigName for whitespace Identity and filters only non-whitespace patterns;
its own mandatory parameter binding must not be inferred from those process
conditions. In particular, public explicit-empty Filter/Identity forwarding has
not yet been settled across the private Trps bridge. No filter guard is changed.

Next work is the actual public/private argument and response path plus a
coordinated dynamic/raw API design with Broadcast. The native modern service
must not be described as proving both session-selected source branches. Existing
[tenant observations](guest-calling-configuration-lab-acceptance.md) retain their
specific modern read/update/restore scope; no new tenant operation occurred.
