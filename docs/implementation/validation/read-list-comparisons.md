# Comparing management reads and lists

The failed PR #296 read run reported differences in the server XML backing
Key.SchemaId, Key.DefaultXml.SchemaId, Key.AuthorityId and Key.XmlRoot while
management getters continued to pass. Those fields include internal XML schema
and XName objects. The earlier read test incorrectly treated the entire backing
Key as immutable management state across two separate HTTP requests.

Pinned Providers BasePolicy declares only Identity. Compiled HttpRequestHelper
Get/GetAll deserialize JSON into the specific TPolicy/List<TPolicy>; these types
project resource settings and do not declare Key. The SDK deliberately retains
more wire data in RawFields/RawJSON. A complete snapshot must be preserved from
its own response; separate requests do not promise identical server XML backing
metadata. Modern raw responses retain that backing data as well.

Cross-endpoint acceptance compares Identity and every other top-level property,
including unknown extensions, ConfigId and ConfigMetadata. Only Key is excluded
from this temporal equality assertion. All typed getters and pinned PowerShell
comparisons remain required. No retry or selective skip replaces the assertion.
Unit response/model/CRUD tests continue to compare each complete actual response,
including its Key, against permanent sanitized acceptance fixtures. Regression
vectors prove metadata variation does not hide changed/missing resource values,
unknown extensions or numeric changes, and does not mutate original snapshots.

Evidence: PR #296 run 44998762751, both app modes; fixed path diagnostics on
TestAccShiftsPolicyReadOnly in the client-secret job. Earlier PR #295 failures
showed metadata differences in room-video/roaming; local 150 repeated pairs and
nine controlled mutation/read pairs did not reproduce the variation. The source
establishes the public model boundary; the server-internal reason for differing
XML serialization is not established and is not attributed to an SDK wire change.
