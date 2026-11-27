// Package licensereport composes documented Graph activity, user and license reads.
package licensereport

import (
	"encoding/json"
	"errors"
	"math/big"
	"strconv"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

var (
	ErrInput       = errors.New("invalid license report request")
	ErrPartial     = errors.New("license report contains incomplete outcomes")
	ErrIdentity    = errors.New("license report identity resolution is missing or ambiguous")
	ErrCount       = errors.New("license report count is unavailable or not an integer")
	ErrSourceRange = errors.New("license report value is outside pinned Int32 projection")
	ErrLicenseMode = errors.New("Graph license details documents delegated authentication only")
)

// Period selects the four documented aggregation windows.
type Period int

const (
	Period7Days   Period = 7
	Period30Days  Period = 30
	Period90Days  Period = 90
	Period180Days Period = 180
)

func validPeriod(v Period) bool {
	return v == Period7Days || v == Period30Days || v == Period90Days || v == Period180Days
}

// SourcePlanID is the pinned source's service-plan presence predicate. Matching
// it is not proof of current entitlement, billing state or effective authority.
const SourcePlanID = "6dc145d6-95dd-4191-b9c3-185575ee6f6b"

// DirectoryUser is this composition's minimal directory read snapshot. Graph
// defines its ID as an opaque String, not a GUID; ID and UPN remain distinct.
// The existing UUID-only resolver is deliberately not used for these reads.
type DirectoryUser struct {
	ID                string                     `json:"id"`
	DisplayName       *string                    `json:"displayName"`
	UserPrincipalName *string                    `json:"userPrincipalName"`
	RawFields         map[string]json.RawMessage `json:"-"`
	RawJSON           json.RawMessage            `json:"-"`
}

func (v *DirectoryUser) UnmarshalJSON(raw []byte) error {
	type snapshot DirectoryUser
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil || next.ID == "" {
		return client.ErrDecode
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*v = DirectoryUser(next)
	return nil
}
func (v DirectoryUser) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type snapshot DirectoryUser
	return json.Marshal(snapshot(v))
}
func (DirectoryUser) String() string     { return "Graph report directory user (redacted)" }
func (v DirectoryUser) GoString() string { return v.String() }

// Activity retains documented report identity/count fields and the whole response.
// Graph CSDL declares Edm.Stream and the JSON documentation specifies no count
// width; json.Number preserves precision without imposing the source's Int32 cap.
type Activity struct {
	UserID                  *string                    `json:"userId"`
	UserPrincipalName       *string                    `json:"userPrincipalName"`
	PrivateChatMessageCount *json.Number               `json:"privateChatMessageCount"`
	TeamChatMessageCount    *json.Number               `json:"teamChatMessageCount"`
	RawFields               map[string]json.RawMessage `json:"-"`
	RawJSON                 json.RawMessage            `json:"-"`
}

func (v *Activity) UnmarshalJSON(raw []byte) error {
	type snapshot Activity
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*v = Activity(next)
	return nil
}
func (v Activity) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type snapshot Activity
	return json.Marshal(snapshot(v))
}
func (Activity) String() string     { return "Graph activity report row (redacted)" }
func (v Activity) GoString() string { return v.String() }

// ServicePlan describes a returned plan. ProvisioningStatus is retained but is
// deliberately not used by the fixed source-plan presence predicate.
type ServicePlan struct {
	ID                 *string                    `json:"servicePlanId"`
	Name               *string                    `json:"servicePlanName"`
	ProvisioningStatus *string                    `json:"provisioningStatus"`
	AppliesTo          *string                    `json:"appliesTo"`
	RawFields          map[string]json.RawMessage `json:"-"`
	RawJSON            json.RawMessage            `json:"-"`
}

func (v *ServicePlan) UnmarshalJSON(raw []byte) error {
	type snapshot ServicePlan
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*v = ServicePlan(next)
	return nil
}
func (v ServicePlan) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type snapshot ServicePlan
	return json.Marshal(snapshot(v))
}
func (ServicePlan) String() string     { return "Graph service plan (redacted)" }
func (v ServicePlan) GoString() string { return v.String() }

// LicenseDetail is a read snapshot, never a license assignment request.
type LicenseDetail struct {
	ID            string                     `json:"id"`
	SKUID         *string                    `json:"skuId"`
	SKUPartNumber *string                    `json:"skuPartNumber"`
	ServicePlans  []*ServicePlan             `json:"servicePlans"`
	RawFields     map[string]json.RawMessage `json:"-"`
	RawJSON       json.RawMessage            `json:"-"`
}

func (v *LicenseDetail) UnmarshalJSON(raw []byte) error {
	type snapshot LicenseDetail
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil || next.ID == "" || next.ServicePlans == nil {
		return client.ErrDecode
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*v = LicenseDetail(next)
	return nil
}
func (v LicenseDetail) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type snapshot LicenseDetail
	return json.Marshal(snapshot(v))
}
func (LicenseDetail) String() string     { return "Graph license detail (redacted)" }
func (v LicenseDetail) GoString() string { return v.String() }

// ListOptions bounds each collection traversal. Zero uses the shared pager's
// finite defaults. Continuations must retain the selected cloud/resource identity.
type ListOptions struct{ MaxPages, MaxItems int }

// LicenseOutcome retains one resolved directory user's partial read. Unknown or
// denied license data never becomes a false source-plan verdict.
type LicenseOutcome struct {
	DirectoryUserID       string
	Items                 []LicenseDetail
	Response              *resty.Response
	Pages                 int
	Complete              bool
	PlanDataComplete      bool
	HasMatchingSourcePlan *bool
	Err                   error
}

func (LicenseOutcome) String() string     { return "Graph user license outcome (redacted)" }
func (v LicenseOutcome) GoString() string { return v.String() }

// Row keeps report identity, resolved directory identity and UPN distinct.
// MessageCount and HasMatchingSourcePlan are nil when their evidence is missing.
type Row struct {
	Activity              Activity
	DirectoryUser         *DirectoryUser
	DirectoryUserID       string
	MessageCount          *big.Int
	HasMatchingSourcePlan *bool
	Err                   error
}

func (Row) String() string     { return "native license report row (redacted)" }
func (v Row) GoString() string { return v.String() }

// ReportResult preserves completed stages and partial rows instead of emitting
// mixed error strings as successful report objects. Complete is a traversal and
// data-availability result, never subscription or authorization qualification.
type ReportResult struct {
	Activity                                                    client.ListResult[Activity]
	Users                                                       client.ListResult[DirectoryUser]
	Licenses                                                    []LicenseOutcome
	Rows                                                        []Row
	ActivityComplete, UsersComplete, LicensesComplete, Complete bool
	DocumentedLicenseMode                                       bool
}

func (ReportResult) String() string     { return "Graph license report result (redacted)" }
func (v ReportResult) GoString() string { return v.String() }

func integer(v *json.Number) (*big.Int, error) {
	if v == nil {
		return nil, ErrCount
	}
	// Preserve arbitrary wire numbers, but bound derived arithmetic independently
	// of schema width so a hostile short exponent cannot allocate unbounded memory.
	text := v.String()
	if len(text) == 0 || len(text) > 4096 || (!(text[0] >= '0' && text[0] <= '9') && text[0] != '-') || !json.Valid([]byte(text)) {
		return nil, ErrCount
	}
	if index := strings.IndexAny(text, "eE"); index >= 0 {
		exponent, e := strconv.ParseInt(text[index+1:], 10, 32)
		if e != nil || exponent > 4096 || exponent < -4096 {
			return nil, ErrCount
		}
	}
	rational, ok := new(big.Rat).SetString(text)
	if !ok || !rational.IsInt() {
		return nil, ErrCount
	}
	return new(big.Int).Set(rational.Num()), nil
}
func total(activity Activity) (*big.Int, error) {
	a, e := integer(activity.PrivateChatMessageCount)
	if e != nil {
		return nil, e
	}
	b, e := integer(activity.TeamChatMessageCount)
	if e != nil {
		return nil, e
	}
	return new(big.Int).Add(a, b), nil
}
