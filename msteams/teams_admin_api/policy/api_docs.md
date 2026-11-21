# Compiled policy request values

`policy.Field[T]` represents omission, explicit null and a bound value. Resource
request structs use `json:",omitzero"`; zero fields omit, Set(false)/Set("") retain
those values, and Null[T]() sends null. Value() distinguishes a value from null;
IsSet/IsNull report binding. Directly serializing an unset field returns
ErrUnsetField. Set copies T; callers retain responsibility for nested mutable data.

Strings copies a full replacement array, with nil/empty normalized to `[]`.
A nil *StringList property omits it. ReplaceStrings/ChangeStrings build owned local
edits for composed updates. Apply copies current values, removes one ordinal match
per Remove item, then appends Add values without deduplication. A nonnil empty edit
clears, matching the pinned cmdlet's IsEmpty branch. The invalid zero edit rejects.

These values do not emit modern generated ConfigAPI Add/Remove wire operators.
They produce full arrays after an explicit read, without atomicity, ETag, replay
or rollback guarantees. Controls distinguishes bound CRUD Force false/true from
omission; resource serialization emits its pinned string form. Assignment selects
exactly one explicit scope and preserves null/empty names and tenant Force Boolean.
