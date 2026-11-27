package teamsfeedbackpolicy_test

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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_feedback_policy"
)

func TestUnitFeedbackPolicyRequestdCompiledBindings(t *testing.T) {
	t.Parallel()

	fixture := struct {
		Cases []struct {
			Command, Name string
			Arguments     map[string]json.RawMessage
			Observed      struct {
				Failure *string

				Output *struct {
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
				Failure *string

				Output *struct {
					Bound     map[string]json.RawMessage
					Specified *bool
					NoScope   bool `json:"no_scope_result"`
				}
			}
		}{

			{
				Command:   "Get-CsTeamsFeedbackPolicy",
				Name:      "list",
				Arguments: map[string]json.RawMessage{},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{},
					},
				},
			},

			{
				Command: "Get-CsTeamsFeedbackPolicy",
				Name:    "filter",
				Arguments: map[string]json.RawMessage{
					"Filter": json.RawMessage("\"Fixture*\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"Filter": json.RawMessage("\"Fixture*\""),
						},
					},
				},
			},

			{
				Command: "Get-CsTeamsFeedbackPolicy",
				Name:    "empty-filter",
				Arguments: map[string]json.RawMessage{
					"Filter": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Failure: new("System.Management.Automation.ValidationMetadataException"),
				},
			},

			{
				Command: "Get-CsTeamsFeedbackPolicy",
				Name:    "empty-identity",
				Arguments: map[string]json.RawMessage{
					"Identity": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Failure: new("System.Management.Automation.ValidationMetadataException"),
				},
			},

			{
				Command: "New-CsTeamsFeedbackPolicy",
				Name:    "identity-only",
				Arguments: map[string]json.RawMessage{
					"Identity": json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"Identity": json.RawMessage("\"Fixture\""),
						},
						Specified: new(false),
					},
				},
			},

			{
				Command: "New-CsTeamsFeedbackPolicy",
				Name:    "empty-identity",
				Arguments: map[string]json.RawMessage{
					"Identity": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Failure: new("System.Management.Automation.ParameterBindingValidationException"),
				},
			},

			{
				Command: "Set-CsTeamsFeedbackPolicy",
				Name:    "identity-only",
				Arguments: map[string]json.RawMessage{
					"Identity": json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"Identity": json.RawMessage("\"Fixture\""),
						},
						Specified: new(false),
					},
				},
			},

			{
				Command: "Set-CsTeamsFeedbackPolicy",
				Name:    "force-only",
				Arguments: map[string]json.RawMessage{
					"Force":    json.RawMessage("false"),
					"Identity": json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"Force":    json.RawMessage("{\"IsPresent\":false}"),
							"Identity": json.RawMessage("\"Fixture\""),
						},
						Specified: new(false),
					},
				},
			},

			{
				Command: "Set-CsTeamsFeedbackPolicy",
				Name:    "empty-identity",
				Arguments: map[string]json.RawMessage{
					"Identity":          json.RawMessage("\"\""),
					"UserInitiatedMode": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Failure: new("System.Management.Automation.ValidationMetadataException"),
				},
			},

			{
				Command: "Remove-CsTeamsFeedbackPolicy",
				Name:    "force-false",
				Arguments: map[string]json.RawMessage{
					"Force":    json.RawMessage("false"),
					"Identity": json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"Force":    json.RawMessage("{\"IsPresent\":false}"),
							"Identity": json.RawMessage("\"Fixture\""),
						},
					},
				},
			},

			{
				Command: "New-CsTeamsFeedbackPolicy",
				Name:    "UserInitiatedMode-null",
				Arguments: map[string]json.RawMessage{
					"Identity":          json.RawMessage("\"Fixture\""),
					"UserInitiatedMode": json.RawMessage("null"),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"Identity":          json.RawMessage("\"Fixture\""),
							"UserInitiatedMode": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsFeedbackPolicy",
				Name:    "UserInitiatedMode-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":          json.RawMessage("\"Fixture\""),
					"UserInitiatedMode": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"Identity":          json.RawMessage("\"Fixture\""),
							"UserInitiatedMode": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsFeedbackPolicy",
				Name:    "ReceiveSurveysMode-null",
				Arguments: map[string]json.RawMessage{
					"Identity":           json.RawMessage("\"Fixture\""),
					"ReceiveSurveysMode": json.RawMessage("null"),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"Identity":           json.RawMessage("\"Fixture\""),
							"ReceiveSurveysMode": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsFeedbackPolicy",
				Name:    "ReceiveSurveysMode-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":           json.RawMessage("\"Fixture\""),
					"ReceiveSurveysMode": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"Identity":           json.RawMessage("\"Fixture\""),
							"ReceiveSurveysMode": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsFeedbackPolicy",
				Name:    "AllowScreenshotCollection-null",
				Arguments: map[string]json.RawMessage{
					"AllowScreenshotCollection": json.RawMessage("null"),
					"Identity":                  json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"AllowScreenshotCollection": json.RawMessage("null"),
							"Identity":                  json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsFeedbackPolicy",
				Name:    "AllowScreenshotCollection-false",
				Arguments: map[string]json.RawMessage{
					"AllowScreenshotCollection": json.RawMessage("false"),
					"Identity":                  json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"AllowScreenshotCollection": json.RawMessage("false"),
							"Identity":                  json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsFeedbackPolicy",
				Name:    "AllowEmailCollection-null",
				Arguments: map[string]json.RawMessage{
					"AllowEmailCollection": json.RawMessage("null"),
					"Identity":             json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"AllowEmailCollection": json.RawMessage("null"),
							"Identity":             json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsFeedbackPolicy",
				Name:    "AllowEmailCollection-false",
				Arguments: map[string]json.RawMessage{
					"AllowEmailCollection": json.RawMessage("false"),
					"Identity":             json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"AllowEmailCollection": json.RawMessage("false"),
							"Identity":             json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsFeedbackPolicy",
				Name:    "AllowLogCollection-null",
				Arguments: map[string]json.RawMessage{
					"AllowLogCollection": json.RawMessage("null"),
					"Identity":           json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"AllowLogCollection": json.RawMessage("null"),
							"Identity":           json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsFeedbackPolicy",
				Name:    "AllowLogCollection-false",
				Arguments: map[string]json.RawMessage{
					"AllowLogCollection": json.RawMessage("false"),
					"Identity":           json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"AllowLogCollection": json.RawMessage("false"),
							"Identity":           json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsFeedbackPolicy",
				Name:    "EnableFeatureSuggestions-null",
				Arguments: map[string]json.RawMessage{
					"EnableFeatureSuggestions": json.RawMessage("null"),
					"Identity":                 json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"EnableFeatureSuggestions": json.RawMessage("null"),
							"Identity":                 json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsFeedbackPolicy",
				Name:    "EnableFeatureSuggestions-false",
				Arguments: map[string]json.RawMessage{
					"EnableFeatureSuggestions": json.RawMessage("false"),
					"Identity":                 json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"EnableFeatureSuggestions": json.RawMessage("false"),
							"Identity":                 json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "New-CsTeamsFeedbackPolicy",
				Name:    "all-settings",
				Arguments: map[string]json.RawMessage{
					"AllowEmailCollection":      json.RawMessage("true"),
					"AllowLogCollection":        json.RawMessage("true"),
					"AllowScreenshotCollection": json.RawMessage("true"),
					"EnableFeatureSuggestions":  json.RawMessage("true"),
					"Identity":                  json.RawMessage("\"Fixture\""),
					"ReceiveSurveysMode":        json.RawMessage("\"Future-ReceiveSurveysMode\""),
					"UserInitiatedMode":         json.RawMessage("\"Future-UserInitiatedMode\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"AllowEmailCollection":      json.RawMessage("true"),
							"AllowLogCollection":        json.RawMessage("true"),
							"AllowScreenshotCollection": json.RawMessage("true"),
							"EnableFeatureSuggestions":  json.RawMessage("true"),
							"Identity":                  json.RawMessage("\"Fixture\""),
							"ReceiveSurveysMode":        json.RawMessage("\"Future-ReceiveSurveysMode\""),
							"UserInitiatedMode":         json.RawMessage("\"Future-UserInitiatedMode\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsFeedbackPolicy",
				Name:    "UserInitiatedMode-null",
				Arguments: map[string]json.RawMessage{
					"Identity":          json.RawMessage("\"Fixture\""),
					"UserInitiatedMode": json.RawMessage("null"),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"Identity":          json.RawMessage("\"Fixture\""),
							"UserInitiatedMode": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsFeedbackPolicy",
				Name:    "UserInitiatedMode-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":          json.RawMessage("\"Fixture\""),
					"UserInitiatedMode": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"Identity":          json.RawMessage("\"Fixture\""),
							"UserInitiatedMode": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsFeedbackPolicy",
				Name:    "ReceiveSurveysMode-null",
				Arguments: map[string]json.RawMessage{
					"Identity":           json.RawMessage("\"Fixture\""),
					"ReceiveSurveysMode": json.RawMessage("null"),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"Identity":           json.RawMessage("\"Fixture\""),
							"ReceiveSurveysMode": json.RawMessage("null"),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsFeedbackPolicy",
				Name:    "ReceiveSurveysMode-empty",
				Arguments: map[string]json.RawMessage{
					"Identity":           json.RawMessage("\"Fixture\""),
					"ReceiveSurveysMode": json.RawMessage("\"\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"Identity":           json.RawMessage("\"Fixture\""),
							"ReceiveSurveysMode": json.RawMessage("\"\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsFeedbackPolicy",
				Name:    "AllowScreenshotCollection-null",
				Arguments: map[string]json.RawMessage{
					"AllowScreenshotCollection": json.RawMessage("null"),
					"Identity":                  json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"AllowScreenshotCollection": json.RawMessage("null"),
							"Identity":                  json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsFeedbackPolicy",
				Name:    "AllowScreenshotCollection-false",
				Arguments: map[string]json.RawMessage{
					"AllowScreenshotCollection": json.RawMessage("false"),
					"Identity":                  json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"AllowScreenshotCollection": json.RawMessage("false"),
							"Identity":                  json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsFeedbackPolicy",
				Name:    "AllowEmailCollection-null",
				Arguments: map[string]json.RawMessage{
					"AllowEmailCollection": json.RawMessage("null"),
					"Identity":             json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"AllowEmailCollection": json.RawMessage("null"),
							"Identity":             json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsFeedbackPolicy",
				Name:    "AllowEmailCollection-false",
				Arguments: map[string]json.RawMessage{
					"AllowEmailCollection": json.RawMessage("false"),
					"Identity":             json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"AllowEmailCollection": json.RawMessage("false"),
							"Identity":             json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsFeedbackPolicy",
				Name:    "AllowLogCollection-null",
				Arguments: map[string]json.RawMessage{
					"AllowLogCollection": json.RawMessage("null"),
					"Identity":           json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"AllowLogCollection": json.RawMessage("null"),
							"Identity":           json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsFeedbackPolicy",
				Name:    "AllowLogCollection-false",
				Arguments: map[string]json.RawMessage{
					"AllowLogCollection": json.RawMessage("false"),
					"Identity":           json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"AllowLogCollection": json.RawMessage("false"),
							"Identity":           json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsFeedbackPolicy",
				Name:    "EnableFeatureSuggestions-null",
				Arguments: map[string]json.RawMessage{
					"EnableFeatureSuggestions": json.RawMessage("null"),
					"Identity":                 json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"EnableFeatureSuggestions": json.RawMessage("null"),
							"Identity":                 json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsFeedbackPolicy",
				Name:    "EnableFeatureSuggestions-false",
				Arguments: map[string]json.RawMessage{
					"EnableFeatureSuggestions": json.RawMessage("false"),
					"Identity":                 json.RawMessage("\"Fixture\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"EnableFeatureSuggestions": json.RawMessage("false"),
							"Identity":                 json.RawMessage("\"Fixture\""),
						},
						Specified: new(true),
					},
				},
			},

			{
				Command: "Set-CsTeamsFeedbackPolicy",
				Name:    "all-settings",
				Arguments: map[string]json.RawMessage{
					"AllowEmailCollection":      json.RawMessage("true"),
					"AllowLogCollection":        json.RawMessage("true"),
					"AllowScreenshotCollection": json.RawMessage("true"),
					"EnableFeatureSuggestions":  json.RawMessage("true"),
					"Identity":                  json.RawMessage("\"Fixture\""),
					"ReceiveSurveysMode":        json.RawMessage("\"Future-ReceiveSurveysMode\""),
					"UserInitiatedMode":         json.RawMessage("\"Future-UserInitiatedMode\""),
				},
				Observed: struct {
					Failure *string

					Output *struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}
				}{
					Output: &struct {
						Bound     map[string]json.RawMessage
						Specified *bool
						NoScope   bool `json:"no_scope_result"`
					}{
						Bound: map[string]json.RawMessage{
							"AllowEmailCollection":      json.RawMessage("true"),
							"AllowLogCollection":        json.RawMessage("true"),
							"AllowScreenshotCollection": json.RawMessage("true"),
							"EnableFeatureSuggestions":  json.RawMessage("true"),
							"Identity":                  json.RawMessage("\"Fixture\""),
							"ReceiveSurveysMode":        json.RawMessage("\"Future-ReceiveSurveysMode\""),
							"UserInitiatedMode":         json.RawMessage("\"Future-UserInitiatedMode\""),
						},
						Specified: new(true),
					},
				},
			},
		},
	}

	for _, f := range fixture.Cases {

		bound := f.Arguments
		if f.Observed.Output != nil {
			bound = f.Observed.Output.Bound

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
		if json.Unmarshal(bound["Identity"], &identity) != nil {
			identity = ""
		}
		if json.Unmarshal(bound["Filter"], &pattern) != nil {
			pattern = ""
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
			case "New-CsTeamsFeedbackPolicy", setCommand, "Remove-CsTeamsFeedbackPolicy":
				method, path, reply, status = "POST", collection, "", 400
				fields := map[string]json.RawMessage{}
				if f.Command != "Remove-CsTeamsFeedbackPolicy" {
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
				if f.Command == "Remove-CsTeamsFeedbackPolicy" {
					method, path = "DELETE", collection+"/configuration/Fixture"
				}
			default:
				t.Fatal("unknown source command")
			}
			c.Steps = []testutil.Exchange{compiledWire(method, path, f.Command, body, status, reply)}
		}
		if len(c.Steps) != 0 {
			c.Want = testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
		}
		c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.TeamsFeedbackPolicies
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
			case "New-CsTeamsFeedbackPolicy":
				v, r, e := s.CreateByIdentity(ctx, identity, resource.CreateRequest{Settings: input, Controls: controls})
				return testutil.Observation{Value: v, Response: r, Err: e}
			case setCommand:
				v, r, e := s.PatchByIdentity(ctx, identity, resource.PatchRequest{Settings: input, Controls: controls})
				return testutil.Observation{Value: v, Response: r, Err: e}
			case "Remove-CsTeamsFeedbackPolicy":
				v, r, e := s.DeleteByIdentity(ctx, identity, controls)
				return testutil.Observation{Value: v, Response: r, Err: e}
			}
			return testutil.Observation{Err: errors.New("unknown source command")}
		}
		run(t, c)
	}
}
