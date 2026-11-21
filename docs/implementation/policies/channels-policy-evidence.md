# Teams channels policy evidence (POL-38 / #111)

The MicrosoftTeams 8.1.0 baseline exports compiled Core Get/New/Set/Remove
commands. Grant is a Function backed by private GrantCsTeamsChannelsPolicy,
a TrpsPSCmdlet. The pinned CmdletConfig maps it to Grant-CsTeamsPolicy with
PolicyType TeamsChannelsPolicy. Using the Core grant class would select a
different protocol and parameter contract.

Get has Filter/Identity sets; New/Set/Remove have Identity sets. Exported Grant
has GrantToGroup, GrantToTenant and Identity, for eight exact command/set pairs.
Native symbols and tests cover the selected behavior below. Discovery and offline results do not establish complete command/tenant qualification.

## Capture seams

`capture-channels-policy-provider.ps1` verifies the 8.1.0 manifest and Core,
PolicyRp and HostContract hashes. Its owned ChannelsPolicyProviderCapture adapter
uses the actual direct HttpRequestHelper with synthetic InvocationInfo, fixture
token, credential-refusing host and memory sender. Client, version and host are
restored. No command constructor/lifecycle, public binder, credentials, discovery
or network runs. Fifteen provider observations each dispatch once with zero token
requests. Sixty-nine actual model/Newtonsoft observations cover all 20 settings.
Two invalid object-valued primitive projections fail. These are compiled CRUD
observations; they do not select or execute the exported Grant transport.

`capture-channels-policy-grant-wrappers.ps1` verifies the wrapper/configuration
hashes and Grant mapping, extracts the unchanged mapped AST function and executes
it under fake telemetry/internal-command seams. Eleven observations capture
scope, policy-name and Rank behavior; missing policy/scope and both PassThru forms
fail. No exported TrpsPSCmdlet, flights, runtime attributes or tenant execution is
claimed. The generic function's own binding is separate from the public export.

`capture-channels-policy-grant-requests.ps1` verifies the private generated
assembly hash and invokes actual generated DTO/helper request and response paths
through an injected memory sender. Seven observations cover user/tenant grants,
empty policy names, group priority presence/zero and 202/204 responses. Host,
headers, session discovery and production pipeline remain separate seams.

The three promoted fixtures live in `contracts/endpoints/8.1.0/fixtures/`.
PowerShell and these owned capture tools are development-only, never a runtime
SDK dependency. No Microsoft binaries or decompiled source are committed.

## Selected behavior

- Compiled CRUD has MPACmdlet and command/correlation headers, no forest target
  header. POST uses request Identity; PATCH uses model Identity and only bound
  settings; Delete has an Identity/optional Force JSON body. A bound Force switch
  serializes as a string, including false. Empty/null reads succeed. Empty list
  sets the source result flag false without HTTP failure; envelopes fail.
- All 20 settings are nullable: 14 strings and six Booleans. In particular,
  AllowPrivateChannelCreation is Boolean and AllowCreatePrivateChannel is a
  separate string. Both remain independently represented. All modern creation,
  outside-team membership/sharing and default-type strings are retained.
- Setters accept null/empty/future strings, Boolean null/false/true and 2000-byte
  Description. Discarded generated length constants do not enforce a limit.
  No local enum, length or cross-field constraint is inferred. Server/cloud/
  flight availability and effective channel restrictions remain unqualified.
- Model projections accept case variants and later duplicate members. The native
  service must retain future/raw fields separately and use known typed values
  for canonical comparisons. Native JSON requires actual string/Boolean types;
  arbitrary Newtonsoft primitive coercions are not a public binder guarantee.
- Exported grants use modern session forest discovery and generated headers.
  User/tenant 200 differs from group 202/204. The mapped string binder turns a null
  policy name into empty. Rank omitted or zero is omitted by the mapped wrapper;
  the native generated DTO can send explicit priority zero. Tenant helper Force
  is Boolean; the public Grant does not expose Force. No implicit poll, replay,
  restore, elevation, effective-clear or completion guarantee follows acceptance.

## Remaining work

Native compiled CRUD/list, explicit Set Get/PATCH composition, all 20 typed presence-aware settings and modern grants are present. Tests compare source provider/generated requests and models, accepted final settings in sparse requests, protocol/header isolation, partial Set results, all mutation errors, context/backend/input guards and actual POST versus discovery acceptance. Accepted model settings compose captured model and helper behavior; they are not a source binder/tenant execution claim. Actual lab evidence is recorded in `channels-policy-lab-acceptance.md`. Full command and tenant qualification and issue #111 remain open.
