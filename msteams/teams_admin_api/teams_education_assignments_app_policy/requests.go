package teamseducationassignmentsapppolicy

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// Settings preserves omitted/null/empty strings. Only the URL has a local
// source model constraint; enabled-type strings remain opaque.
type Settings struct {
	ParentDigestEnabledType policy.Field[string] `json:"ParentDigestEnabledType,omitzero"`
	MakeCodeEnabledType     policy.Field[string] `json:"MakeCodeEnabledType,omitzero"`
	TurnItInEnabledType     policy.Field[string] `json:"TurnItInEnabledType,omitzero"`
	TurnItInAPIURL          policy.Field[string] `json:"TurnItInApiUrl,omitzero"`
	TurnItInAPIKey          policy.Field[string] `json:"TurnItInApiKey,omitzero"`
}

type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}
type UpdateRequest struct {
	Settings
	policy.Controls
}

// UpdateResult retains the preceding read and submission status after failure.
// HTTPAccepted does not establish persistence or effective app behavior.
type UpdateResult struct {
	Before, After              *Policy
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func fields(input Settings) (map[string]json.RawMessage, error) {
	if value, bound := input.TurnItInAPIURL.Value(); bound && !validAPIURL(value) {
		return nil, ErrInput
	}
	return policyrp.Fields(input)
}
func (Settings) String() string         { return "education assignments settings (redacted)" }
func (p Settings) GoString() string     { return p.String() }
func (UpdateResult) String() string     { return "education assignments update result (redacted)" }
func (p UpdateResult) GoString() string { return p.String() }
