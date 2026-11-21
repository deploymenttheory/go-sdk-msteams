# POL-29 privacy configuration evidence

Issue #102 owns Get-CsPrivacyConfiguration (Filter and Identity sets) and
Set-CsPrivacyConfiguration (Identity set). This is a source/capture checkpoint,
not a native service, contract pass or successful tenant qualification. The
immutable default baseline and current coverage states are unchanged.

The accepted 8.1.0 package SHA256 is
`78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`.
The relevant packaged Core/PolicyRp/host SHA256 values are pinned in
[the capture script](../../../tools/contracts/capture-privacy-provider.ps1).
Local ILSpy inspection and the accepted runtime snapshot agree on all four nullable
Boolean settings: EnablePrivacyMode, AutoInitiateContacts,
PublishLocationDataDefault and DisplayPublishedPhotoDefault. The model has a
separate Identity property and simple nullable setters, with no extra validation
or manufactured defaults.

GetPrivacyConfigurationCmdlet derives from BaseTpmGetCmdlet of PrivacyConfiguration.
SetPrivacyConfigurationCmdlet derives from the corresponding BaseTpmSetCmdlet.
It considers a setting specified by BoundParameters membership, preserving false
and null, then updates only those bound properties. Force alone is not a setting.
This establishes the compiled direct protocol, not generated forest discovery or
Graph. The resource name is PrivacyConfiguration.

Native implementation should expose explicit Get/List/local Filter and sparse
PATCH, plus Set's explicit read/modify/PATCH composition. Scope/identity escaping,
Global-first ordering, Identity/Tag wildcard selection, body identity override,
Force binding and compiled headers belong to the existing policyrp adapter.
Retain preparation metadata if a composed write fails; do not invent atomicity,
ETags, automatic replay or effective-state guarantees. No Create/Delete/Grant
operation is established by these two commands.

[The development oracle](../../../tools/contracts/oracle/PrivacyProviderCapture.cs)
uses actual pinned HttpRequestHelper methods with synthetic InvocationInfo,
credential-refusing host and a memory-only sender. It replaces/restores the private
client, version and host. MPACmdlet's default is supplied from factory source.
No management command constructor/ProcessRecord, credentials, generated discovery
or network runs. The script captures reads/list, nullable/false sparse settings,
bound-false Force, empty success, an invalid collection envelope and a denial.
This is a provider/helper observation, not full shell binding or live qualification.

The next steps are to check the successful capture, add the named compiled-resource
whitelist entry and typed native service, wire the root client, map all three sets,
and run offline plus opt-in tenant reads. Tenant writes require separate owned
fixtures/read-back; lab unavailability must stay visible without dropping commands.
