package teamseventspolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_events_policy"
)

func TestUnitEventsPolicyRequestdCompiledBindings(t *testing.T) {
	t.Parallel()

	fixture := struct {
		Cases []struct {
			Command, Name string
			Arguments     map[string]json.RawMessage
			Observed      struct {
				Failure       *string
				Sends, Tokens int
				Output        *struct {
					Target    string `json:"target"`
					Bound     map[string]json.RawMessage
					Specified *bool
					NoScope   bool `json:"no_scope_result"`
				}
			}
		}
	}{
		Cases: []struct {
			Command, Name string
			Arguments     map[string]json.RawMessage
			Observed      struct {
				Failure       *string
				Sends, Tokens int
				Output        *struct {
					Target    string `json:"target"`
					Bound     map[string]json.RawMessage
					Specified *bool
					NoScope   bool `json:"no_scope_result"`
				}
			}
		}{

			{
				Command:   "Get-CsTeamsEventsPolicy",
				Name:      "list",
				Arguments: map[string]json.RawMessage{},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound:  map[string]json.RawMessage{},
					},
				},
			},

			{
				Command: "Get-CsTeamsEventsPolicy",
				Name:    "filter",
				Arguments: map[string]json.RawMessage{
					"Filter": json.RawMessage("\"Fixture*\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "filter",
						Bound: map[string]json.RawMessage{
							"Filter": json.RawMessage("\"Fixture*\""),
						},
					},
				},
			},

			{
				Command: "Get-CsTeamsEventsPolicy",
				Name:    "empty-filter",
				Arguments: map[string]json.RawMessage{
					"Filter": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Failure: new("System.Management.Automation.ValidationMetadataException"),
				},
			},

			{
				Command: "Get-CsTeamsEventsPolicy",
				Name:    "empty-identity",
				Arguments: map[string]json.RawMessage{
					"Identity": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Failure: new("System.Management.Automation.ValidationMetadataException"),
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "identity-only",
				Arguments: map[string]json.RawMessage{
					"Identity": json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity": json.RawMessage("\"Fixture\""),
						},
						Specified: new(false),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "empty-identity",
				Arguments: map[string]json.RawMessage{
					"Identity": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Failure: new("System.Management.Automation.ParameterBindingValidationException"),
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "identity-only",
				Arguments: map[string]json.RawMessage{
					"Identity": json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity": json.RawMessage("\"Fixture\""),
						},
						Specified: new(false),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "force-only",
				Arguments: map[string]json.RawMessage{
					"Force":    json.RawMessage("false"),
					"Identity": json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Force":    json.RawMessage("{\"IsPresent\":false}"),
							"Identity": json.RawMessage("\"Fixture\""),
						},
						Specified: new(false),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "empty-identity",
				Arguments: map[string]json.RawMessage{
					"Description": json.RawMessage("\"\""),
					"Identity":    json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Failure: new("System.Management.Automation.ValidationMetadataException"),
				},
			},

			{
				Command: "Remove-CsTeamsEventsPolicy",
				Name:    "force-false",
				Arguments: map[string]json.RawMessage{
					"Force":    json.RawMessage("false"),
					"Identity": json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Force":    json.RawMessage("{\"IsPresent\":false}"),
							"Identity": json.RawMessage("\"Fixture\""),
						},
					},
				},
			},

			{
				Command:   "Grant-CsTeamsEventsPolicy",
				Name:      "no-scope",
				Arguments: map[string]json.RawMessage{},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Failure: new("System.Exception"),
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target:  "none",
						Bound:   map[string]json.RawMessage{},
						NoScope: true,
					},
				},
			},

			{
				Command: "Grant-CsTeamsEventsPolicy",
				Name:    "user-omitted-name",
				Arguments: map[string]json.RawMessage{
					"Identity": json.RawMessage("\"user/fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "user",
						Bound: map[string]json.RawMessage{
							"Identity": json.RawMessage("\"user/fixture\""),
						},
						NoScope: false,
					},
				},
			},

			{
				Command: "Grant-CsTeamsEventsPolicy",
				Name:    "user-null-name",
				Arguments: map[string]json.RawMessage{
					"Identity":   json.RawMessage("\"user/fixture\""),
					"PolicyName": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "user",
						Bound: map[string]json.RawMessage{
							"Identity":   json.RawMessage("\"user/fixture\""),
							"PolicyName": json.RawMessage("null"),
						},
						NoScope: false,
					},
				},
			},

			{
				Command: "Grant-CsTeamsEventsPolicy",
				Name:    "user-empty-name",
				Arguments: map[string]json.RawMessage{
					"Identity":   json.RawMessage("\"user/fixture\""),
					"PolicyName": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "user",
						Bound: map[string]json.RawMessage{
							"Identity":   json.RawMessage("\"user/fixture\""),
							"PolicyName": json.RawMessage("\"\""),
						},
						NoScope: false,
					},
				},
			},

			{
				Command: "Grant-CsTeamsEventsPolicy",
				Name:    "tenant-false-selector",
				Arguments: map[string]json.RawMessage{
					"Force":      json.RawMessage("false"),
					"Global":     json.RawMessage("false"),
					"PolicyName": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "tenant",
						Bound: map[string]json.RawMessage{
							"Force":      json.RawMessage("{\"IsPresent\":false}"),
							"Global":     json.RawMessage("{\"IsPresent\":false}"),
							"PolicyName": json.RawMessage("null"),
						},
						NoScope: false,
					},
				},
			},

			{
				Command: "Grant-CsTeamsEventsPolicy",
				Name:    "tenant-true-force",
				Arguments: map[string]json.RawMessage{
					"Force":      json.RawMessage("true"),
					"Global":     json.RawMessage("true"),
					"PolicyName": json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "tenant",
						Bound: map[string]json.RawMessage{
							"Force":      json.RawMessage("{\"IsPresent\":true}"),
							"Global":     json.RawMessage("{\"IsPresent\":true}"),
							"PolicyName": json.RawMessage("\"Fixture\""),
						},
						NoScope: false,
					},
				},
			},

			{
				Command: "Grant-CsTeamsEventsPolicy",
				Name:    "group-rank-zero",
				Arguments: map[string]json.RawMessage{
					"Group":      json.RawMessage("\"group/fixture\""),
					"PolicyName": json.RawMessage("null"),
					"Rank":       json.RawMessage("0"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "group",
						Bound: map[string]json.RawMessage{
							"Group":      json.RawMessage("\"group/fixture\""),
							"PolicyName": json.RawMessage("null"),
							"Rank":       json.RawMessage("0"),
						},
						NoScope: false,
					},
				},
			},

			{
				Command: "Grant-CsTeamsEventsPolicy",
				Name:    "group-rank-negative",
				Arguments: map[string]json.RawMessage{
					"Group":      json.RawMessage("\"group/fixture\""),
					"PolicyName": json.RawMessage("\"\""),
					"Rank":       json.RawMessage("-1"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "group",
						Bound: map[string]json.RawMessage{
							"Group":      json.RawMessage("\"group/fixture\""),
							"PolicyName": json.RawMessage("\"\""),
							"Rank":       json.RawMessage("-1"),
						},
						NoScope: false,
					},
				},
			},

			{
				Command: "Grant-CsTeamsEventsPolicy",
				Name:    "group-rank-two",
				Arguments: map[string]json.RawMessage{
					"Group":      json.RawMessage("\"group/fixture\""),
					"PolicyName": json.RawMessage("\"Fixture\""),
					"Rank":       json.RawMessage("2"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "group",
						Bound: map[string]json.RawMessage{
							"Group":      json.RawMessage("\"group/fixture\""),
							"PolicyName": json.RawMessage("\"Fixture\""),
							"Rank":       json.RawMessage("2"),
						},
						NoScope: false,
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "AllowWebinars-null",
				Arguments: map[string]json.RawMessage{
					"AllowWebinars": json.RawMessage("null"),
					"Identity":      json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowWebinars": json.RawMessage("null"),
							"Identity":      json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "AllowWebinars-empty",
				Arguments: map[string]json.RawMessage{
					"AllowWebinars": json.RawMessage("\"\""),
					"Identity":      json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowWebinars": json.RawMessage("\"\""),
							"Identity":      json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "EventAccessType-null",
				Arguments: map[string]json.RawMessage{
					"EventAccessType": json.RawMessage("null"),
					"Identity":        json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"EventAccessType": json.RawMessage("null"),
							"Identity":        json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "EventAccessType-empty",
				Arguments: map[string]json.RawMessage{
					"EventAccessType": json.RawMessage("\"\""),
					"Identity":        json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"EventAccessType": json.RawMessage("\"\""),
							"Identity":        json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "AllowTownhalls-null",
				Arguments: map[string]json.RawMessage{
					"AllowTownhalls": json.RawMessage("null"),
					"Identity":       json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowTownhalls": json.RawMessage("null"),
							"Identity":       json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "AllowTownhalls-empty",
				Arguments: map[string]json.RawMessage{
					"AllowTownhalls": json.RawMessage("\"\""),
					"Identity":       json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowTownhalls": json.RawMessage("\"\""),
							"Identity":       json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "TownhallEventAttendeeAccess-null",
				Arguments: map[string]json.RawMessage{
					"Identity":                    json.RawMessage("\"Fixture\""),
					"TownhallEventAttendeeAccess": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":                    json.RawMessage("\"Fixture\""),
							"TownhallEventAttendeeAccess": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "TownhallEventAttendeeAccess-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":                    json.RawMessage("\"Fixture\""),
					"TownhallEventAttendeeAccess": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":                    json.RawMessage("\"Fixture\""),
							"TownhallEventAttendeeAccess": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "ExternalPresenterJoinVerification-null",
				Arguments: map[string]json.RawMessage{
					"ExternalPresenterJoinVerification": json.RawMessage("null"),
					"Identity":                          json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"ExternalPresenterJoinVerification": json.RawMessage("null"),
							"Identity":                          json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "ExternalPresenterJoinVerification-empty",
				Arguments: map[string]json.RawMessage{
					"ExternalPresenterJoinVerification": json.RawMessage("\"\""),
					"Identity":                          json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"ExternalPresenterJoinVerification": json.RawMessage("\"\""),
							"Identity":                          json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "AllowEmailEditing-null",
				Arguments: map[string]json.RawMessage{
					"AllowEmailEditing": json.RawMessage("null"),
					"Identity":          json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowEmailEditing": json.RawMessage("null"),
							"Identity":          json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "AllowEmailEditing-empty",
				Arguments: map[string]json.RawMessage{
					"AllowEmailEditing": json.RawMessage("\"\""),
					"Identity":          json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowEmailEditing": json.RawMessage("\"\""),
							"Identity":          json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "AllowedQuestionTypesInRegistrationForm-null",
				Arguments: map[string]json.RawMessage{
					"AllowedQuestionTypesInRegistrationForm": json.RawMessage("null"),
					"Identity":                               json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowedQuestionTypesInRegistrationForm": json.RawMessage("null"),
							"Identity":                               json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "AllowedQuestionTypesInRegistrationForm-empty",
				Arguments: map[string]json.RawMessage{
					"AllowedQuestionTypesInRegistrationForm": json.RawMessage("\"\""),
					"Identity":                               json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowedQuestionTypesInRegistrationForm": json.RawMessage("\"\""),
							"Identity":                               json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "AllowEventIntegrations-null",
				Arguments: map[string]json.RawMessage{
					"AllowEventIntegrations": json.RawMessage("null"),
					"Identity":               json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowEventIntegrations": json.RawMessage("null"),
							"Identity":               json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "AllowEventIntegrations-false",
				Arguments: map[string]json.RawMessage{
					"AllowEventIntegrations": json.RawMessage("false"),
					"Identity":               json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowEventIntegrations": json.RawMessage("false"),
							"Identity":               json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "AllowedWebinarTypesForRecordingPublish-null",
				Arguments: map[string]json.RawMessage{
					"AllowedWebinarTypesForRecordingPublish": json.RawMessage("null"),
					"Identity":                               json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowedWebinarTypesForRecordingPublish": json.RawMessage("null"),
							"Identity":                               json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "AllowedWebinarTypesForRecordingPublish-empty",
				Arguments: map[string]json.RawMessage{
					"AllowedWebinarTypesForRecordingPublish": json.RawMessage("\"\""),
					"Identity":                               json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowedWebinarTypesForRecordingPublish": json.RawMessage("\"\""),
							"Identity":                               json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "AllowedTownhallTypesForRecordingPublish-null",
				Arguments: map[string]json.RawMessage{
					"AllowedTownhallTypesForRecordingPublish": json.RawMessage("null"),
					"Identity": json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowedTownhallTypesForRecordingPublish": json.RawMessage("null"),
							"Identity": json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "AllowedTownhallTypesForRecordingPublish-empty",
				Arguments: map[string]json.RawMessage{
					"AllowedTownhallTypesForRecordingPublish": json.RawMessage("\"\""),
					"Identity": json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowedTownhallTypesForRecordingPublish": json.RawMessage("\"\""),
							"Identity": json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "RecordingForTownhall-null",
				Arguments: map[string]json.RawMessage{
					"Identity":             json.RawMessage("\"Fixture\""),
					"RecordingForTownhall": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":             json.RawMessage("\"Fixture\""),
							"RecordingForTownhall": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "RecordingForTownhall-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":             json.RawMessage("\"Fixture\""),
					"RecordingForTownhall": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":             json.RawMessage("\"Fixture\""),
							"RecordingForTownhall": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "RecordingForWebinar-null",
				Arguments: map[string]json.RawMessage{
					"Identity":            json.RawMessage("\"Fixture\""),
					"RecordingForWebinar": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":            json.RawMessage("\"Fixture\""),
							"RecordingForWebinar": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "RecordingForWebinar-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":            json.RawMessage("\"Fixture\""),
					"RecordingForWebinar": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":            json.RawMessage("\"Fixture\""),
							"RecordingForWebinar": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "TranscriptionForTownhall-null",
				Arguments: map[string]json.RawMessage{
					"Identity":                 json.RawMessage("\"Fixture\""),
					"TranscriptionForTownhall": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":                 json.RawMessage("\"Fixture\""),
							"TranscriptionForTownhall": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "TranscriptionForTownhall-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":                 json.RawMessage("\"Fixture\""),
					"TranscriptionForTownhall": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":                 json.RawMessage("\"Fixture\""),
							"TranscriptionForTownhall": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "TranscriptionForWebinar-null",
				Arguments: map[string]json.RawMessage{
					"Identity":                json.RawMessage("\"Fixture\""),
					"TranscriptionForWebinar": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":                json.RawMessage("\"Fixture\""),
							"TranscriptionForWebinar": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "TranscriptionForWebinar-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":                json.RawMessage("\"Fixture\""),
					"TranscriptionForWebinar": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":                json.RawMessage("\"Fixture\""),
							"TranscriptionForWebinar": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "TownhallChatExperience-null",
				Arguments: map[string]json.RawMessage{
					"Identity":               json.RawMessage("\"Fixture\""),
					"TownhallChatExperience": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":               json.RawMessage("\"Fixture\""),
							"TownhallChatExperience": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "TownhallChatExperience-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":               json.RawMessage("\"Fixture\""),
					"TownhallChatExperience": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":               json.RawMessage("\"Fixture\""),
							"TownhallChatExperience": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "BroadcastPremiumApps-null",
				Arguments: map[string]json.RawMessage{
					"BroadcastPremiumApps": json.RawMessage("null"),
					"Identity":             json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"BroadcastPremiumApps": json.RawMessage("null"),
							"Identity":             json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "BroadcastPremiumApps-empty",
				Arguments: map[string]json.RawMessage{
					"BroadcastPremiumApps": json.RawMessage("\"\""),
					"Identity":             json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"BroadcastPremiumApps": json.RawMessage("\"\""),
							"Identity":             json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "UseMicrosoftECDN-null",
				Arguments: map[string]json.RawMessage{
					"Identity":         json.RawMessage("\"Fixture\""),
					"UseMicrosoftECDN": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":         json.RawMessage("\"Fixture\""),
							"UseMicrosoftECDN": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "UseMicrosoftECDN-false",
				Arguments: map[string]json.RawMessage{
					"Identity":         json.RawMessage("\"Fixture\""),
					"UseMicrosoftECDN": json.RawMessage("false"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":         json.RawMessage("\"Fixture\""),
							"UseMicrosoftECDN": json.RawMessage("false"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "ImmersiveEvents-null",
				Arguments: map[string]json.RawMessage{
					"Identity":        json.RawMessage("\"Fixture\""),
					"ImmersiveEvents": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":        json.RawMessage("\"Fixture\""),
							"ImmersiveEvents": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "ImmersiveEvents-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":        json.RawMessage("\"Fixture\""),
					"ImmersiveEvents": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":        json.RawMessage("\"Fixture\""),
							"ImmersiveEvents": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "TownhallMaxResolution-null",
				Arguments: map[string]json.RawMessage{
					"Identity":              json.RawMessage("\"Fixture\""),
					"TownhallMaxResolution": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":              json.RawMessage("\"Fixture\""),
							"TownhallMaxResolution": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "TownhallMaxResolution-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":              json.RawMessage("\"Fixture\""),
					"TownhallMaxResolution": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":              json.RawMessage("\"Fixture\""),
							"TownhallMaxResolution": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "HighBitrateForTownhall-null",
				Arguments: map[string]json.RawMessage{
					"HighBitrateForTownhall": json.RawMessage("null"),
					"Identity":               json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"HighBitrateForTownhall": json.RawMessage("null"),
							"Identity":               json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "HighBitrateForTownhall-empty",
				Arguments: map[string]json.RawMessage{
					"HighBitrateForTownhall": json.RawMessage("\"\""),
					"Identity":               json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"HighBitrateForTownhall": json.RawMessage("\"\""),
							"Identity":               json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "BackroomChat-null",
				Arguments: map[string]json.RawMessage{
					"BackroomChat": json.RawMessage("null"),
					"Identity":     json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"BackroomChat": json.RawMessage("null"),
							"Identity":     json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "BackroomChat-empty",
				Arguments: map[string]json.RawMessage{
					"BackroomChat": json.RawMessage("\"\""),
					"Identity":     json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"BackroomChat": json.RawMessage("\"\""),
							"Identity":     json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "Registration-null",
				Arguments: map[string]json.RawMessage{
					"Identity":     json.RawMessage("\"Fixture\""),
					"Registration": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":     json.RawMessage("\"Fixture\""),
							"Registration": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "Registration-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":     json.RawMessage("\"Fixture\""),
					"Registration": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":     json.RawMessage("\"Fixture\""),
							"Registration": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "AllowEngagementReport-null",
				Arguments: map[string]json.RawMessage{
					"AllowEngagementReport": json.RawMessage("null"),
					"Identity":              json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowEngagementReport": json.RawMessage("null"),
							"Identity":              json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "AllowEngagementReport-empty",
				Arguments: map[string]json.RawMessage{
					"AllowEngagementReport": json.RawMessage("\"\""),
					"Identity":              json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowEngagementReport": json.RawMessage("\"\""),
							"Identity":              json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "InfoShownInReportMode-null",
				Arguments: map[string]json.RawMessage{
					"Identity":              json.RawMessage("\"Fixture\""),
					"InfoShownInReportMode": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":              json.RawMessage("\"Fixture\""),
							"InfoShownInReportMode": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "InfoShownInReportMode-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":              json.RawMessage("\"Fixture\""),
					"InfoShownInReportMode": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":              json.RawMessage("\"Fixture\""),
							"InfoShownInReportMode": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "Description-null",
				Arguments: map[string]json.RawMessage{
					"Description": json.RawMessage("null"),
					"Identity":    json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Description": json.RawMessage("null"),
							"Identity":    json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "Description-empty",
				Arguments: map[string]json.RawMessage{
					"Description": json.RawMessage("\"\""),
					"Identity":    json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Description": json.RawMessage("\"\""),
							"Identity":    json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsEventsPolicy",
				Name:    "all-settings",
				Arguments: map[string]json.RawMessage{
					"AllowEmailEditing":                       json.RawMessage("\"Future-AllowEmailEditing\""),
					"AllowEngagementReport":                   json.RawMessage("\"Future-AllowEngagementReport\""),
					"AllowEventIntegrations":                  json.RawMessage("true"),
					"AllowTownhalls":                          json.RawMessage("\"Future-AllowTownhalls\""),
					"AllowWebinars":                           json.RawMessage("\"Future-AllowWebinars\""),
					"AllowedQuestionTypesInRegistrationForm":  json.RawMessage("\"Future-AllowedQuestionTypesInRegistrationForm\""),
					"AllowedTownhallTypesForRecordingPublish": json.RawMessage("\"Future-AllowedTownhallTypesForRecordingPublish\""),
					"AllowedWebinarTypesForRecordingPublish":  json.RawMessage("\"Future-AllowedWebinarTypesForRecordingPublish\""),
					"BackroomChat":                            json.RawMessage("\"Future-BackroomChat\""),
					"BroadcastPremiumApps":                    json.RawMessage("\"Future-BroadcastPremiumApps\""),
					"Description":                             json.RawMessage("\"Future-Description\""),
					"EventAccessType":                         json.RawMessage("\"Future-EventAccessType\""),
					"ExternalPresenterJoinVerification":       json.RawMessage("\"Future-ExternalPresenterJoinVerification\""),
					"HighBitrateForTownhall":                  json.RawMessage("\"Future-HighBitrateForTownhall\""),
					"Identity":                                json.RawMessage("\"Fixture\""),
					"ImmersiveEvents":                         json.RawMessage("\"Future-ImmersiveEvents\""),
					"InfoShownInReportMode":                   json.RawMessage("\"Future-InfoShownInReportMode\""),
					"RecordingForTownhall":                    json.RawMessage("\"Future-RecordingForTownhall\""),
					"RecordingForWebinar":                     json.RawMessage("\"Future-RecordingForWebinar\""),
					"Registration":                            json.RawMessage("\"Future-Registration\""),
					"TownhallChatExperience":                  json.RawMessage("\"Future-TownhallChatExperience\""),
					"TownhallEventAttendeeAccess":             json.RawMessage("\"Future-TownhallEventAttendeeAccess\""),
					"TownhallMaxResolution":                   json.RawMessage("\"Future-TownhallMaxResolution\""),
					"TranscriptionForTownhall":                json.RawMessage("\"Future-TranscriptionForTownhall\""),
					"TranscriptionForWebinar":                 json.RawMessage("\"Future-TranscriptionForWebinar\""),
					"UseMicrosoftECDN":                        json.RawMessage("true"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowEmailEditing":                       json.RawMessage("\"Future-AllowEmailEditing\""),
							"AllowEngagementReport":                   json.RawMessage("\"Future-AllowEngagementReport\""),
							"AllowEventIntegrations":                  json.RawMessage("true"),
							"AllowTownhalls":                          json.RawMessage("\"Future-AllowTownhalls\""),
							"AllowWebinars":                           json.RawMessage("\"Future-AllowWebinars\""),
							"AllowedQuestionTypesInRegistrationForm":  json.RawMessage("\"Future-AllowedQuestionTypesInRegistrationForm\""),
							"AllowedTownhallTypesForRecordingPublish": json.RawMessage("\"Future-AllowedTownhallTypesForRecordingPublish\""),
							"AllowedWebinarTypesForRecordingPublish":  json.RawMessage("\"Future-AllowedWebinarTypesForRecordingPublish\""),
							"BackroomChat":                            json.RawMessage("\"Future-BackroomChat\""),
							"BroadcastPremiumApps":                    json.RawMessage("\"Future-BroadcastPremiumApps\""),
							"Description":                             json.RawMessage("\"Future-Description\""),
							"EventAccessType":                         json.RawMessage("\"Future-EventAccessType\""),
							"ExternalPresenterJoinVerification":       json.RawMessage("\"Future-ExternalPresenterJoinVerification\""),
							"HighBitrateForTownhall":                  json.RawMessage("\"Future-HighBitrateForTownhall\""),
							"Identity":                                json.RawMessage("\"Fixture\""),
							"ImmersiveEvents":                         json.RawMessage("\"Future-ImmersiveEvents\""),
							"InfoShownInReportMode":                   json.RawMessage("\"Future-InfoShownInReportMode\""),
							"RecordingForTownhall":                    json.RawMessage("\"Future-RecordingForTownhall\""),
							"RecordingForWebinar":                     json.RawMessage("\"Future-RecordingForWebinar\""),
							"Registration":                            json.RawMessage("\"Future-Registration\""),
							"TownhallChatExperience":                  json.RawMessage("\"Future-TownhallChatExperience\""),
							"TownhallEventAttendeeAccess":             json.RawMessage("\"Future-TownhallEventAttendeeAccess\""),
							"TownhallMaxResolution":                   json.RawMessage("\"Future-TownhallMaxResolution\""),
							"TranscriptionForTownhall":                json.RawMessage("\"Future-TranscriptionForTownhall\""),
							"TranscriptionForWebinar":                 json.RawMessage("\"Future-TranscriptionForWebinar\""),
							"UseMicrosoftECDN":                        json.RawMessage("true"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "AllowWebinars-null",
				Arguments: map[string]json.RawMessage{
					"AllowWebinars": json.RawMessage("null"),
					"Identity":      json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowWebinars": json.RawMessage("null"),
							"Identity":      json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "AllowWebinars-empty",
				Arguments: map[string]json.RawMessage{
					"AllowWebinars": json.RawMessage("\"\""),
					"Identity":      json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowWebinars": json.RawMessage("\"\""),
							"Identity":      json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "EventAccessType-null",
				Arguments: map[string]json.RawMessage{
					"EventAccessType": json.RawMessage("null"),
					"Identity":        json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"EventAccessType": json.RawMessage("null"),
							"Identity":        json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "EventAccessType-empty",
				Arguments: map[string]json.RawMessage{
					"EventAccessType": json.RawMessage("\"\""),
					"Identity":        json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"EventAccessType": json.RawMessage("\"\""),
							"Identity":        json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "AllowTownhalls-null",
				Arguments: map[string]json.RawMessage{
					"AllowTownhalls": json.RawMessage("null"),
					"Identity":       json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowTownhalls": json.RawMessage("null"),
							"Identity":       json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "AllowTownhalls-empty",
				Arguments: map[string]json.RawMessage{
					"AllowTownhalls": json.RawMessage("\"\""),
					"Identity":       json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowTownhalls": json.RawMessage("\"\""),
							"Identity":       json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "TownhallEventAttendeeAccess-null",
				Arguments: map[string]json.RawMessage{
					"Identity":                    json.RawMessage("\"Fixture\""),
					"TownhallEventAttendeeAccess": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":                    json.RawMessage("\"Fixture\""),
							"TownhallEventAttendeeAccess": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "TownhallEventAttendeeAccess-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":                    json.RawMessage("\"Fixture\""),
					"TownhallEventAttendeeAccess": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":                    json.RawMessage("\"Fixture\""),
							"TownhallEventAttendeeAccess": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "ExternalPresenterJoinVerification-null",
				Arguments: map[string]json.RawMessage{
					"ExternalPresenterJoinVerification": json.RawMessage("null"),
					"Identity":                          json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"ExternalPresenterJoinVerification": json.RawMessage("null"),
							"Identity":                          json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "ExternalPresenterJoinVerification-empty",
				Arguments: map[string]json.RawMessage{
					"ExternalPresenterJoinVerification": json.RawMessage("\"\""),
					"Identity":                          json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"ExternalPresenterJoinVerification": json.RawMessage("\"\""),
							"Identity":                          json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "AllowEmailEditing-null",
				Arguments: map[string]json.RawMessage{
					"AllowEmailEditing": json.RawMessage("null"),
					"Identity":          json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowEmailEditing": json.RawMessage("null"),
							"Identity":          json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "AllowEmailEditing-empty",
				Arguments: map[string]json.RawMessage{
					"AllowEmailEditing": json.RawMessage("\"\""),
					"Identity":          json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowEmailEditing": json.RawMessage("\"\""),
							"Identity":          json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "AllowedQuestionTypesInRegistrationForm-null",
				Arguments: map[string]json.RawMessage{
					"AllowedQuestionTypesInRegistrationForm": json.RawMessage("null"),
					"Identity":                               json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowedQuestionTypesInRegistrationForm": json.RawMessage("null"),
							"Identity":                               json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "AllowedQuestionTypesInRegistrationForm-empty",
				Arguments: map[string]json.RawMessage{
					"AllowedQuestionTypesInRegistrationForm": json.RawMessage("\"\""),
					"Identity":                               json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowedQuestionTypesInRegistrationForm": json.RawMessage("\"\""),
							"Identity":                               json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "AllowEventIntegrations-null",
				Arguments: map[string]json.RawMessage{
					"AllowEventIntegrations": json.RawMessage("null"),
					"Identity":               json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowEventIntegrations": json.RawMessage("null"),
							"Identity":               json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "AllowEventIntegrations-false",
				Arguments: map[string]json.RawMessage{
					"AllowEventIntegrations": json.RawMessage("false"),
					"Identity":               json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowEventIntegrations": json.RawMessage("false"),
							"Identity":               json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "AllowedWebinarTypesForRecordingPublish-null",
				Arguments: map[string]json.RawMessage{
					"AllowedWebinarTypesForRecordingPublish": json.RawMessage("null"),
					"Identity":                               json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowedWebinarTypesForRecordingPublish": json.RawMessage("null"),
							"Identity":                               json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "AllowedWebinarTypesForRecordingPublish-empty",
				Arguments: map[string]json.RawMessage{
					"AllowedWebinarTypesForRecordingPublish": json.RawMessage("\"\""),
					"Identity":                               json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowedWebinarTypesForRecordingPublish": json.RawMessage("\"\""),
							"Identity":                               json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "AllowedTownhallTypesForRecordingPublish-null",
				Arguments: map[string]json.RawMessage{
					"AllowedTownhallTypesForRecordingPublish": json.RawMessage("null"),
					"Identity": json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowedTownhallTypesForRecordingPublish": json.RawMessage("null"),
							"Identity": json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "AllowedTownhallTypesForRecordingPublish-empty",
				Arguments: map[string]json.RawMessage{
					"AllowedTownhallTypesForRecordingPublish": json.RawMessage("\"\""),
					"Identity": json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowedTownhallTypesForRecordingPublish": json.RawMessage("\"\""),
							"Identity": json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "RecordingForTownhall-null",
				Arguments: map[string]json.RawMessage{
					"Identity":             json.RawMessage("\"Fixture\""),
					"RecordingForTownhall": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":             json.RawMessage("\"Fixture\""),
							"RecordingForTownhall": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "RecordingForTownhall-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":             json.RawMessage("\"Fixture\""),
					"RecordingForTownhall": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":             json.RawMessage("\"Fixture\""),
							"RecordingForTownhall": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "RecordingForWebinar-null",
				Arguments: map[string]json.RawMessage{
					"Identity":            json.RawMessage("\"Fixture\""),
					"RecordingForWebinar": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":            json.RawMessage("\"Fixture\""),
							"RecordingForWebinar": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "RecordingForWebinar-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":            json.RawMessage("\"Fixture\""),
					"RecordingForWebinar": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":            json.RawMessage("\"Fixture\""),
							"RecordingForWebinar": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "TranscriptionForTownhall-null",
				Arguments: map[string]json.RawMessage{
					"Identity":                 json.RawMessage("\"Fixture\""),
					"TranscriptionForTownhall": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":                 json.RawMessage("\"Fixture\""),
							"TranscriptionForTownhall": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "TranscriptionForTownhall-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":                 json.RawMessage("\"Fixture\""),
					"TranscriptionForTownhall": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":                 json.RawMessage("\"Fixture\""),
							"TranscriptionForTownhall": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "TranscriptionForWebinar-null",
				Arguments: map[string]json.RawMessage{
					"Identity":                json.RawMessage("\"Fixture\""),
					"TranscriptionForWebinar": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":                json.RawMessage("\"Fixture\""),
							"TranscriptionForWebinar": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "TranscriptionForWebinar-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":                json.RawMessage("\"Fixture\""),
					"TranscriptionForWebinar": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":                json.RawMessage("\"Fixture\""),
							"TranscriptionForWebinar": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "TownhallChatExperience-null",
				Arguments: map[string]json.RawMessage{
					"Identity":               json.RawMessage("\"Fixture\""),
					"TownhallChatExperience": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":               json.RawMessage("\"Fixture\""),
							"TownhallChatExperience": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "TownhallChatExperience-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":               json.RawMessage("\"Fixture\""),
					"TownhallChatExperience": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":               json.RawMessage("\"Fixture\""),
							"TownhallChatExperience": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "BroadcastPremiumApps-null",
				Arguments: map[string]json.RawMessage{
					"BroadcastPremiumApps": json.RawMessage("null"),
					"Identity":             json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"BroadcastPremiumApps": json.RawMessage("null"),
							"Identity":             json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "BroadcastPremiumApps-empty",
				Arguments: map[string]json.RawMessage{
					"BroadcastPremiumApps": json.RawMessage("\"\""),
					"Identity":             json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"BroadcastPremiumApps": json.RawMessage("\"\""),
							"Identity":             json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "UseMicrosoftECDN-null",
				Arguments: map[string]json.RawMessage{
					"Identity":         json.RawMessage("\"Fixture\""),
					"UseMicrosoftECDN": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":         json.RawMessage("\"Fixture\""),
							"UseMicrosoftECDN": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "UseMicrosoftECDN-false",
				Arguments: map[string]json.RawMessage{
					"Identity":         json.RawMessage("\"Fixture\""),
					"UseMicrosoftECDN": json.RawMessage("false"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":         json.RawMessage("\"Fixture\""),
							"UseMicrosoftECDN": json.RawMessage("false"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "ImmersiveEvents-null",
				Arguments: map[string]json.RawMessage{
					"Identity":        json.RawMessage("\"Fixture\""),
					"ImmersiveEvents": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":        json.RawMessage("\"Fixture\""),
							"ImmersiveEvents": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "ImmersiveEvents-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":        json.RawMessage("\"Fixture\""),
					"ImmersiveEvents": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":        json.RawMessage("\"Fixture\""),
							"ImmersiveEvents": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "TownhallMaxResolution-null",
				Arguments: map[string]json.RawMessage{
					"Identity":              json.RawMessage("\"Fixture\""),
					"TownhallMaxResolution": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":              json.RawMessage("\"Fixture\""),
							"TownhallMaxResolution": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "TownhallMaxResolution-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":              json.RawMessage("\"Fixture\""),
					"TownhallMaxResolution": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":              json.RawMessage("\"Fixture\""),
							"TownhallMaxResolution": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "HighBitrateForTownhall-null",
				Arguments: map[string]json.RawMessage{
					"HighBitrateForTownhall": json.RawMessage("null"),
					"Identity":               json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"HighBitrateForTownhall": json.RawMessage("null"),
							"Identity":               json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "HighBitrateForTownhall-empty",
				Arguments: map[string]json.RawMessage{
					"HighBitrateForTownhall": json.RawMessage("\"\""),
					"Identity":               json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"HighBitrateForTownhall": json.RawMessage("\"\""),
							"Identity":               json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "BackroomChat-null",
				Arguments: map[string]json.RawMessage{
					"BackroomChat": json.RawMessage("null"),
					"Identity":     json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"BackroomChat": json.RawMessage("null"),
							"Identity":     json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "BackroomChat-empty",
				Arguments: map[string]json.RawMessage{
					"BackroomChat": json.RawMessage("\"\""),
					"Identity":     json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"BackroomChat": json.RawMessage("\"\""),
							"Identity":     json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "Registration-null",
				Arguments: map[string]json.RawMessage{
					"Identity":     json.RawMessage("\"Fixture\""),
					"Registration": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":     json.RawMessage("\"Fixture\""),
							"Registration": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "Registration-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":     json.RawMessage("\"Fixture\""),
					"Registration": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":     json.RawMessage("\"Fixture\""),
							"Registration": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "AllowEngagementReport-null",
				Arguments: map[string]json.RawMessage{
					"AllowEngagementReport": json.RawMessage("null"),
					"Identity":              json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowEngagementReport": json.RawMessage("null"),
							"Identity":              json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "AllowEngagementReport-empty",
				Arguments: map[string]json.RawMessage{
					"AllowEngagementReport": json.RawMessage("\"\""),
					"Identity":              json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowEngagementReport": json.RawMessage("\"\""),
							"Identity":              json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "InfoShownInReportMode-null",
				Arguments: map[string]json.RawMessage{
					"Identity":              json.RawMessage("\"Fixture\""),
					"InfoShownInReportMode": json.RawMessage("null"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":              json.RawMessage("\"Fixture\""),
							"InfoShownInReportMode": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "InfoShownInReportMode-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":              json.RawMessage("\"Fixture\""),
					"InfoShownInReportMode": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Identity":              json.RawMessage("\"Fixture\""),
							"InfoShownInReportMode": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "Description-null",
				Arguments: map[string]json.RawMessage{
					"Description": json.RawMessage("null"),
					"Identity":    json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Description": json.RawMessage("null"),
							"Identity":    json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "Description-empty",
				Arguments: map[string]json.RawMessage{
					"Description": json.RawMessage("\"\""),
					"Identity":    json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"Description": json.RawMessage("\"\""),
							"Identity":    json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsEventsPolicy",
				Name:    "all-settings",
				Arguments: map[string]json.RawMessage{
					"AllowEmailEditing":                       json.RawMessage("\"Future-AllowEmailEditing\""),
					"AllowEngagementReport":                   json.RawMessage("\"Future-AllowEngagementReport\""),
					"AllowEventIntegrations":                  json.RawMessage("true"),
					"AllowTownhalls":                          json.RawMessage("\"Future-AllowTownhalls\""),
					"AllowWebinars":                           json.RawMessage("\"Future-AllowWebinars\""),
					"AllowedQuestionTypesInRegistrationForm":  json.RawMessage("\"Future-AllowedQuestionTypesInRegistrationForm\""),
					"AllowedTownhallTypesForRecordingPublish": json.RawMessage("\"Future-AllowedTownhallTypesForRecordingPublish\""),
					"AllowedWebinarTypesForRecordingPublish":  json.RawMessage("\"Future-AllowedWebinarTypesForRecordingPublish\""),
					"BackroomChat":                            json.RawMessage("\"Future-BackroomChat\""),
					"BroadcastPremiumApps":                    json.RawMessage("\"Future-BroadcastPremiumApps\""),
					"Description":                             json.RawMessage("\"Future-Description\""),
					"EventAccessType":                         json.RawMessage("\"Future-EventAccessType\""),
					"ExternalPresenterJoinVerification":       json.RawMessage("\"Future-ExternalPresenterJoinVerification\""),
					"HighBitrateForTownhall":                  json.RawMessage("\"Future-HighBitrateForTownhall\""),
					"Identity":                                json.RawMessage("\"Fixture\""),
					"ImmersiveEvents":                         json.RawMessage("\"Future-ImmersiveEvents\""),
					"InfoShownInReportMode":                   json.RawMessage("\"Future-InfoShownInReportMode\""),
					"RecordingForTownhall":                    json.RawMessage("\"Future-RecordingForTownhall\""),
					"RecordingForWebinar":                     json.RawMessage("\"Future-RecordingForWebinar\""),
					"Registration":                            json.RawMessage("\"Future-Registration\""),
					"TownhallChatExperience":                  json.RawMessage("\"Future-TownhallChatExperience\""),
					"TownhallEventAttendeeAccess":             json.RawMessage("\"Future-TownhallEventAttendeeAccess\""),
					"TownhallMaxResolution":                   json.RawMessage("\"Future-TownhallMaxResolution\""),
					"TranscriptionForTownhall":                json.RawMessage("\"Future-TranscriptionForTownhall\""),
					"TranscriptionForWebinar":                 json.RawMessage("\"Future-TranscriptionForWebinar\""),
					"UseMicrosoftECDN":                        json.RawMessage("true"),
				},
				Observed: struct {
					Failure       *string
					Sends, Tokens int
					Output        *struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Target    string `json:"target"`
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Target: "none",
						Bound: map[string]json.RawMessage{
							"AllowEmailEditing":                       json.RawMessage("\"Future-AllowEmailEditing\""),
							"AllowEngagementReport":                   json.RawMessage("\"Future-AllowEngagementReport\""),
							"AllowEventIntegrations":                  json.RawMessage("true"),
							"AllowTownhalls":                          json.RawMessage("\"Future-AllowTownhalls\""),
							"AllowWebinars":                           json.RawMessage("\"Future-AllowWebinars\""),
							"AllowedQuestionTypesInRegistrationForm":  json.RawMessage("\"Future-AllowedQuestionTypesInRegistrationForm\""),
							"AllowedTownhallTypesForRecordingPublish": json.RawMessage("\"Future-AllowedTownhallTypesForRecordingPublish\""),
							"AllowedWebinarTypesForRecordingPublish":  json.RawMessage("\"Future-AllowedWebinarTypesForRecordingPublish\""),
							"BackroomChat":                            json.RawMessage("\"Future-BackroomChat\""),
							"BroadcastPremiumApps":                    json.RawMessage("\"Future-BroadcastPremiumApps\""),
							"Description":                             json.RawMessage("\"Future-Description\""),
							"EventAccessType":                         json.RawMessage("\"Future-EventAccessType\""),
							"ExternalPresenterJoinVerification":       json.RawMessage("\"Future-ExternalPresenterJoinVerification\""),
							"HighBitrateForTownhall":                  json.RawMessage("\"Future-HighBitrateForTownhall\""),
							"Identity":                                json.RawMessage("\"Fixture\""),
							"ImmersiveEvents":                         json.RawMessage("\"Future-ImmersiveEvents\""),
							"InfoShownInReportMode":                   json.RawMessage("\"Future-InfoShownInReportMode\""),
							"RecordingForTownhall":                    json.RawMessage("\"Future-RecordingForTownhall\""),
							"RecordingForWebinar":                     json.RawMessage("\"Future-RecordingForWebinar\""),
							"Registration":                            json.RawMessage("\"Future-Registration\""),
							"TownhallChatExperience":                  json.RawMessage("\"Future-TownhallChatExperience\""),
							"TownhallEventAttendeeAccess":             json.RawMessage("\"Future-TownhallEventAttendeeAccess\""),
							"TownhallMaxResolution":                   json.RawMessage("\"Future-TownhallMaxResolution\""),
							"TranscriptionForTownhall":                json.RawMessage("\"Future-TranscriptionForTownhall\""),
							"TranscriptionForWebinar":                 json.RawMessage("\"Future-TranscriptionForWebinar\""),
							"UseMicrosoftECDN":                        json.RawMessage("true"),
						},
						Specified: new(true),
					},
				},
			},
		},
	}

	for _, f := range fixture.Cases {

		bound, target := f.Arguments, "none"
		if f.Observed.Output != nil {
			bound = f.Observed.Output.Bound
			target = f.Observed.Output.Target
		} else if _, ok := bound["Filter"]; ok {
			target = "filter"
		}
		encoded, e := json.Marshal(bound)
		if e != nil {
			t.Fatal(e)
		}
		var input resource.Settings
		if json.Unmarshal(encoded, &input) != nil {
			t.Fatal("source bound values are not typed settings")
		}
		var identity, pattern string
		var name *string
		if json.Unmarshal(bound["Identity"], &identity) != nil {
			identity = ""
		}
		if json.Unmarshal(bound["Filter"], &pattern) != nil {
			pattern = ""
		}
		if value, ok := bound["PolicyName"]; ok && json.Unmarshal(value, &name) != nil {
			t.Fatal("invalid source policy name")
		}
		controls := policy.Controls{}
		if value, ok := bound["Force"]; ok {
			var v struct{ IsPresent bool }
			if json.Unmarshal(value, &v) != nil {
				t.Fatal("invalid source switch")
			}
			controls.Force = new(v.IsPresent)
		}
		c := testutil.Case{Name: strings.ToLower(f.Command + "-" + f.Name)}
		invalid := f.Observed.Failure != nil || f.Command == setCommand && (f.Observed.Output == nil || f.Observed.Output.Specified == nil || !*f.Observed.Output.Specified)
		if invalid {
			c.Want.Error = resource.ErrInput
		} else {
			method, path, reply, status := "GET", collection, "", 400
			var body json.RawMessage
			switch f.Command {
			case getCommand:
				if identity != "" {
					path = collection + "/configuration/Fixture"
					reply = `{"Identity":"Fixture"}`
				}
				c.Want.JSON = json.RawMessage(reply)
			case "New-CsTeamsEventsPolicy", setCommand, "Remove-CsTeamsEventsPolicy":
				method, path, reply, status = "POST", collection, "", 400
				fields := map[string]json.RawMessage{}
				if f.Command != "Remove-CsTeamsEventsPolicy" {
					settings, _ := json.Marshal(input)
					if json.Unmarshal(settings, &fields) != nil {
						t.Fatal("invalid settings projection")
					}
				}
				fields["Identity"], _ = json.Marshal(identity)
				if controls.Force != nil {
					value := "False"
					if *controls.Force {
						value = "True"
					}
					fields["Force"], _ = json.Marshal(value)
				}
				body, _ = json.Marshal(fields)
				if f.Command == setCommand {
					method, path = "PATCH", collection+"/configuration/Fixture"
				}
				if f.Command == "Remove-CsTeamsEventsPolicy" {
					method, path = "DELETE", collection+"/configuration/Fixture"
				}
			case grantCommand:
				reply, status = "", 400
				var fields map[string]any
				switch target {
				case "user":
					method, path = "PATCH", "/Skype.Policy/users/user%2Ffixture/policies/TeamsEventsPolicy"
					fields = map[string]any{"policyName": name, "additionalParameters": nil}
				case "tenant":
					method, path = "PATCH", "/Skype.Policy/tenants/policies/TeamsEventsPolicy"
					fields = map[string]any{"PolicyName": name, "additionalParameters": nil, "forceSwitchPresent": controls.Force != nil && *controls.Force}
				case "group":
					var rank int32
					if json.Unmarshal(bound["Rank"], &rank) != nil {
						t.Fatal("required source Rank missing")
					}
					method, path = "POST", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsEventsPolicy"
					fields = map[string]any{"policyName": name, "priority": rank}
				default:
					t.Fatal("successful no-scope source grant")
				}
				body, _ = json.Marshal(fields)
			default:
				t.Fatal("unknown source command")
			}
			c.Steps = []testutil.Exchange{compiledWire(method, path, f.Command, body, status, reply)}
		}
		if len(c.Steps) != 0 {
			c.Want = testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
		}
		c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.TeamsEventsPolicies
			switch f.Command {
			case getCommand:
				if _, ok := bound["Filter"]; ok {
					v, r, e := s.ListFiltered(ctx, pattern)
					return testutil.Observation{Value: v, Response: r, Err: e}
				}
				if _, ok := bound["Identity"]; ok {
					v, r, e := s.GetByIdentity(ctx, identity)
					return testutil.Observation{Value: v, Response: r, Err: e}
				}
				v, r, e := s.List(ctx)
				return testutil.Observation{Value: v, Response: r, Err: e}
			case "New-CsTeamsEventsPolicy":
				v, r, e := s.CreateByIdentity(ctx, identity, resource.CreateRequest{Settings: input, Controls: controls})
				return testutil.Observation{Value: v, Response: r, Err: e}
			case setCommand:
				v, r, e := s.PatchByIdentity(ctx, identity, resource.PatchRequest{Settings: input, Controls: controls})
				return testutil.Observation{Value: v, Response: r, Err: e}
			case "Remove-CsTeamsEventsPolicy":
				v, r, e := s.DeleteByIdentity(ctx, identity, controls)
				return testutil.Observation{Value: v, Response: r, Err: e}
			case grantCommand:
				switch target {
				case "user":
					v, r, e := s.AssignUser(ctx, identity, name)
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "tenant":
					v, r, e := s.AssignTenantDefault(ctx, name, controls.Force != nil && *controls.Force)
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "group":
					var rank int32
					var id string
					if json.Unmarshal(bound["Rank"], &rank) != nil || json.Unmarshal(bound["Group"], &id) != nil {
						return testutil.Observation{Err: errors.New("invalid typed source grant")}
					}
					v, r, e := s.AssignGroup(ctx, id, name, rank)
					return testutil.Observation{Value: v, Response: r, Err: e}
				default:
					v, r, e := s.Assign(ctx, policy.Assignment{})
					return testutil.Observation{Value: v, Response: r, Err: e}
				}
			}
			return testutil.Observation{Err: errors.New("unknown source command")}
		}
		run(t, c)
	}
}
