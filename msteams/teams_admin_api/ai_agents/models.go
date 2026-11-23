package aiagents

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// AgentTypeMicrosoftCopilotStudio is the only agent type the module documents.
// The service compares agent types case-sensitively; other values pass through
// unchanged.
const AgentTypeMicrosoftCopilotStudio = "MicrosoftCopilotStudio"

// ErrInput identifies a request rejected locally before discovery or credentials.
var ErrInput = errors.New("AI agent input is invalid")

// ErrResponse identifies an undeclared successful HTTP status.
var ErrResponse = errors.New("unexpected AI agent response")

// decodeFields unmarshals raw into value (a pointer to a plain alias type) and
// keeps every service field in fields, so unknown properties survive.
func decodeFields(raw []byte, value any, fields *map[string]json.RawMessage) error {
	if json.Unmarshal(raw, value) != nil || json.Unmarshal(raw, fields) != nil || *fields == nil {
		return client.ErrDecode
	}
	return nil
}

// Agent is an AI agent configuration in the voice apps service
// (Get/New/Set-CsAgent). The module renames ConfigurationId to Id and prefixes
// the agent fields with AI; the wire names are kept here. Absent and null
// fields stay nil; RawFields keeps the complete service object.
type Agent struct {
	ConfigurationID          *string                    `json:"ConfigurationId"`
	Name                     *string                    `json:"Name"`
	AgentID                  *string                    `json:"AgentId"`
	AgentType                *string                    `json:"AgentType"`
	AgentTargetTagTemplateID *string                    `json:"AgentTargetTagTemplateId"`
	RawFields                map[string]json.RawMessage `json:"-"`
}

func (v *Agent) UnmarshalJSON(raw []byte) error {
	type plain Agent
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = Agent(value)
	return nil
}

// AgentList is one page of AI agent configurations. AgentCount is the
// service's total, which can exceed len(Agents) when First limits the page.
type AgentList struct {
	Agents     []*Agent                   `json:"AIAgents"`
	AgentCount *int32                     `json:"AgentCount"`
	RawFields  map[string]json.RawMessage `json:"-"`
}

func (v *AgentList) UnmarshalJSON(raw []byte) error {
	type plain AgentList
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	if value.Agents == nil {
		return client.ErrDecode
	}
	for _, agent := range value.Agents {
		if agent == nil {
			return client.ErrDecode
		}
	}
	*v = AgentList(value)
	return nil
}

// ListOptions are the voice apps list query parameters. The module's public
// Get-CsAgent sends none of them; they are exposed from its generated internal
// cmdlet. Nil fields are omitted.
type ListOptions struct {
	First          *int32
	Skip           *int32
	ExcludeContent *bool
	NameFilter     *string
	TypeFilter     *string
	SortBy         *string
	Descending     *bool
}

// AgentRequest is the body of New-CsAgent and Set-CsAgent. Name, AgentID and
// AgentType are required. A non-nil AgentTargetTagTemplateID must name an
// existing tag template; the service rejects unknown IDs with HTTP 400.
type AgentRequest struct {
	Name                     string
	AgentID                  string
	AgentType                string
	AgentTargetTagTemplateID *string
}

// QueryRequest filters Get-CsAiAgents. ProviderID is required and must be a
// GUID; the service rejects other values with HTTP 400. Nil fields are omitted.
type QueryRequest struct {
	ProviderID          string
	AgentID             *string
	AgentIDs            *string
	Channel             *string
	ContinuationToken   *string
	DisplayNameContains *string
	DisplayNamePrefix   *string
	IsTeamsIvrEnabled   *bool
	MaxResults          *int32
	ShowCount           *bool
}

// QueryResult is one page of platform AI agents (Get-CsAiAgents).
// ContinuationToken is nil on the last page; AgentCount is set when ShowCount
// was requested.
type QueryResult struct {
	Agents            []*PlatformAgent           `json:"agents"`
	ContinuationToken *string                    `json:"continuationToken"`
	AgentCount        *int32                     `json:"agentCount"`
	RawFields         map[string]json.RawMessage `json:"-"`
}

func (v *QueryResult) UnmarshalJSON(raw []byte) error {
	type plain QueryResult
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	if value.Agents == nil {
		return client.ErrDecode
	}
	for _, agent := range value.Agents {
		if agent == nil {
			return client.ErrDecode
		}
	}
	*v = QueryResult(value)
	return nil
}

// PlatformAgent is an AI agent registered with the Teams platform service, as
// returned by Get-CsAiAgents.
type PlatformAgent struct {
	MsaAppID            *string                    `json:"msaAppId"`
	MsaAppTenantID      *string                    `json:"msaAppTenantId"`
	BotHandle           *string                    `json:"botHandle"`
	IsTeamsIvrEnabled   *bool                      `json:"isTeamsIvrEnabled"`
	ProviderID          *string                    `json:"providerId"`
	PublishState        *string                    `json:"publishState"`
	CID                 *int64                     `json:"cid"`
	CallingAPIVersion   *string                    `json:"callingApiVersion"`
	MessagingAPIVersion *string                    `json:"messagingApiVersion"`
	Channels            []string                   `json:"channels"`
	IsTeamsVoiceEnabled *bool                      `json:"isTeamsVoiceEnabled"`
	Description         *PlatformAgentDescription  `json:"description"`
	RawFields           map[string]json.RawMessage `json:"-"`
}

func (v *PlatformAgent) UnmarshalJSON(raw []byte) error {
	type plain PlatformAgent
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = PlatformAgent(value)
	return nil
}

// PlatformAgentDescription is the catalog description of a platform AI agent.
// Less common catalog fields (tiles, terms, disclosures, locales) stay in
// RawFields.
type PlatformAgentDescription struct {
	DisplayName      *string                    `json:"displayName"`
	SkypeDisplayName *string                    `json:"skypeDisplayName"`
	Description      *string                    `json:"description"`
	AgentID          *string                    `json:"agentId"`
	AgentType        *string                    `json:"agentType"`
	DeveloperName    *string                    `json:"developerName"`
	PublisherEmail   *string                    `json:"publisherEmail"`
	PublisherCountry *string                    `json:"publisherCountry"`
	IsTrusted        *bool                      `json:"isTrusted"`
	Capabilities     []string                   `json:"capabilities"`
	Categories       []string                   `json:"categories"`
	Keywords         []string                   `json:"keywords"`
	StarRating       *float64                   `json:"starRating"`
	ETag             *string                    `json:"eTag"`
	RawFields        map[string]json.RawMessage `json:"-"`
}

func (v *PlatformAgentDescription) UnmarshalJSON(raw []byte) error {
	type plain PlatformAgentDescription
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = PlatformAgentDescription(value)
	return nil
}

// KnowledgeRequest identifies a meeting or call's AI-generated knowledge.
// ThreadID is required; CallID narrows the request to one call in the thread.
type KnowledgeRequest struct {
	ThreadID string
	CallID   *string
}
