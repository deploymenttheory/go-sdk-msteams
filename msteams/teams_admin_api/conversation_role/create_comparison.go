package conversationrole

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/go-resty/resty/v2"
)

// CreateResult retains the optional Global comparison separately from creation.
// ComparisonErr is diagnostic: a successful action remains successful even if
// comparison failed. CreateInvoked means the action was attempted; HTTPAccepted
// records only a 2xx receipt, never successful parsing or effective persistence.
type CreateResult struct {
	Identity                                       string
	Comparison                                     *Role
	ComparisonResponse                             *resty.Response
	ComparisonErr                                  error
	After                                          *Role
	CreateResponse                                 *resty.Response
	ComparisonInvoked, CreateInvoked, HTTPAccepted bool
}

func (CreateResult) String() string     { return "conversation create comparison result (redacted)" }
func (v CreateResult) GoString() string { return v.String() }

// CreateWithGlobalComparison composes the source's optional presentation read
// with the frozen creation request. The Global read uses New cmdlet metadata;
// ordinary GetByIdentity uses different metadata. Comparison data is never cloned
// into creation. Non-Global comparison errors are retained while creation
// continues; caller cancellation stops the action. Direct Create performs no read.
func (s *Service) CreateWithGlobalComparison(ctx context.Context, input CreateRequest) (*CreateResult, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	// Validate the opaque frozen request before acquiring comparison data. Callers
	// cannot replace its body through Projection's owned view.
	if input.identity == "" || input.projection.Command != "New-CsConversationRole" || len(input.projection.Body) == 0 {
		return nil, nil, ErrInput
	}
	var body map[string]json.RawMessage
	if json.Unmarshal(input.projection.Body, &body) != nil || body == nil {
		return nil, nil, ErrInput
	}
	result := &CreateResult{Identity: input.identity}
	if !strings.EqualFold(input.identity, "Global") {
		result.ComparisonInvoked = true
		result.Comparison, result.ComparisonResponse, result.ComparisonErr = s.resource.Get(ctx, "Global", input.projection.Command)
	}
	if e := ctx.Err(); e != nil {
		return result, result.ComparisonResponse, e
	}
	result.CreateInvoked = true
	var e error
	result.After, result.CreateResponse, e = s.Create(ctx, input)
	result.HTTPAccepted = result.CreateResponse != nil && result.CreateResponse.StatusCode() >= 200 && result.CreateResponse.StatusCode() < 300
	return result, result.CreateResponse, e
}
