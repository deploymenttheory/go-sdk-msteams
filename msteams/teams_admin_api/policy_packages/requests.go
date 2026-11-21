package policypackages

// PolicyPair is a typed package entry. Nil fields follow the generated DTO's
// omitted-null semantics; pointers to empty strings remain explicit values.
type PolicyPair struct {
	PolicyType *string `json:"PolicyType,omitempty"`
	PolicyName *string `json:"PolicyName,omitempty"`
}

// PolicyRanking always serializes Rank, including its zero default. The pinned
// generated DTO uses non-nullable Int32 rather than an optional ranking field.
type PolicyRanking struct {
	PolicyType *string `json:"PolicyType,omitempty"`
	Rank       int32   `json:"Rank"`
}

type CustomPackageRequest struct {
	Name        *string      `json:"PackageName,omitempty"`
	Description *string      `json:"Description,omitempty"`
	Policies    []PolicyPair `json:"PolicyList,omitzero"`
}

type GroupAssignmentRequest struct {
	GroupID     string          `json:"GroupId,omitempty"`
	PackageName *string         `json:"PackageName,omitempty"`
	Rankings    []PolicyRanking `json:"PolicyRankings,omitzero"`
}

// UserAssignmentRequest is used by separate synchronous and asynchronous routes.
// Identity order/duplicates remain unchanged. Nil UserIDs is unbound; an explicit
// empty slice is distinct. PackageName maps to the service's PackageType key.
type UserAssignmentRequest struct {
	PackageName *string  `json:"PackageType,omitempty"`
	UserIDs     []string `json:"UserIds,omitzero"`
}

// Identity maps InputObject.PackageName rather than a caller-supplied URI.
type Identity struct{ PackageName string }

func (PolicyPair) String() string                 { return "policy package pair (redacted)" }
func (v PolicyPair) GoString() string             { return v.String() }
func (PolicyRanking) String() string              { return "policy package ranking (redacted)" }
func (v PolicyRanking) GoString() string          { return v.String() }
func (CustomPackageRequest) String() string       { return "custom policy package request (redacted)" }
func (v CustomPackageRequest) GoString() string   { return v.String() }
func (GroupAssignmentRequest) String() string     { return "group policy package request (redacted)" }
func (v GroupAssignmentRequest) GoString() string { return v.String() }
func (UserAssignmentRequest) String() string      { return "user policy package request (redacted)" }
func (v UserAssignmentRequest) GoString() string  { return v.String() }
func (Identity) String() string                   { return "policy package identity (redacted)" }
func (v Identity) GoString() string               { return v.String() }
