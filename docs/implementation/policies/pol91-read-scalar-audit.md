# POL91 compiled response scalar corrections

The earlier configuration qualifications did not establish complete response
conversion. Their canonical typed replies and binder probes remain valid bounded
evidence, but the actual provider calls plain
`JsonConvert.DeserializeObject<TPolicy>` at `HttpClientHelper.SendAsync` IL452.
A Boolean/string-only Go decoder rejected additional successful source inputs.
The nine affected families' qualification claims were withdrawn without changing
any live record.

The [portable probe](../../../scripts/discovery/observe-pol91-qualified-read-scalars.ps1)
hash-checks Core and PolicyRp before importing the pinned module. Its
[895 observations](evidence/pol91-qualified-read-scalars.json) record the actual
public compiled Get type, own DTO and individual property projections for all nine
families. The probe constructs only data models and invokes the serializer; it
never invokes a management cmdlet lifecycle, authentication or HTTP.

The opt-in internal `policyrp.ReadString` and `ReadBool` readers correct the
Client, Privacy and Education read models. Their explicit wire structs leave
public field types unchanged and keep raw snapshots separate from typed values.
They do not change request-side PowerShell binding or unrelated resource decoders.

Strings preserve nullable values and exact numeric token spelling, including
large integers, exponents and trailing zeros; Boolean tokens become lowercase
text. Nullable Boolean properties accept numeric zero/nonzero after source
floating conversion, Boolean words with source whitespace/NUL trimming, and
empty strings as null. Whitespace-only nonempty strings and other non-Boolean
words still fail. Floating underflow becomes false; overflow becomes true.
Known raw fields retain their received JSON rather than the projected value.

Public-model tests replay all 436 own Client/Privacy/Education source vectors,
including rejection cases. Separate real-root-client reads verify exact typed
values, route/response identity distinction, unknown raw fields and HTTP receipts.
Full affected execution is pending at this implementation checkpoint; these
changes do not restore qualification or establish tenant behavior.

The remaining six families need their own opt-in corrections. The packet also
records source behavior that cannot be implemented by a blanket JSON rewrite:
nullable UInt64 reads round numeric fractions, accept Boolean and numeric-string
values with checked overflow, and map empty strings to null. Storyline reads
include unchecked Int32 enum conversion for some integers and reject fractional
JSON values. Those rules remain separate from literal PowerShell binder behavior.
