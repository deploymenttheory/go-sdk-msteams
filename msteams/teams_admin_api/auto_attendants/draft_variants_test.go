package autoattendants_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	aa "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/auto_attendants"
	"github.com/go-resty/resty/v2"
)

// These serializer-derived replies exercise optional DTO members and the
// Prompt parameter-set differences. They are synthetic, not live acceptance.
func TestUnitDraftBuildersPreserveNonDefaultMembers(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, route, command, set, key, body string
		invoke                               func(context.Context, *aa.Service) (any, *resty.Response, error)
	}{
		{"prompt-audio", "/prompts/draft", "New-CsAutoAttendantPrompt", "AudioFileParamSet", "Prompt",
			`{"ActiveType":"AudioFile","AudioFilePrompt":{"Id":"` + probe + `"}}`,
			func(ctx context.Context, s *aa.Service) (any, *resty.Response, error) {
				return s.DraftPrompt(ctx, aa.Prompt{ActiveType: ptr(aa.PromptAudioFile), AudioFilePrompt: &aa.AudioFile{ID: ptr(probe)}})
			}},
		{"prompt-dual", "/prompts/draft", "New-CsAutoAttendantPrompt", "DualParamSet", "Prompt",
			`{"ActiveType":"AudioFile","TextToSpeechPrompt":"","AudioFilePrompt":{"Id":"` + probe + `"}}`,
			func(ctx context.Context, s *aa.Service) (any, *resty.Response, error) {
				return s.DraftPrompt(ctx, aa.Prompt{ActiveType: ptr(aa.PromptAudioFile), TextToSpeechPrompt: ptr(""), AudioFilePrompt: &aa.AudioFile{ID: ptr(probe)}})
			}},
		{"callable-entity", "/callable-entities/draft", "New-CsAutoAttendantCallableEntity", "__AllParameterSets", "CallableEntity",
			`{"Id":"` + probe + `","Type":"ApplicationEndpoint","EnableTranscription":false,"EnableSharedVoicemailSystemPromptSuppression":true,"CallPriority":0,"SharedVoicemailHistoryTemplateId":"` + attendantID + `"}`,
			func(ctx context.Context, s *aa.Service) (any, *resty.Response, error) {
				return s.DraftCallableEntity(ctx, aa.CallableEntity{ID: ptr(probe), Type: ptr(aa.EntityApplicationEndpoint), EnableTranscription: ptr(false), EnableSharedVoicemailSystemPromptSuppression: ptr(true), CallPriority: ptr[int32](0), SharedVoicemailHistoryTemplateID: ptr(attendantID)})
			}},
		{"menu-option", "/menus/menu-options/draft", "New-CsAutoAttendantMenuOption", "__AllParameterSets", "MenuOption",
			`{"Action":"AgentsAndQueues","DtmfResponse":"Tone1","VoiceResponses":["agent"],"Description":"","AgentTargetType":0,"AgentTarget":"` + probe + `","AgentTargetTagTemplateId":"` + attendantID + `"}`,
			func(ctx context.Context, s *aa.Service) (any, *resty.Response, error) {
				return s.DraftMenuOption(ctx, aa.MenuOption{Action: ptr(aa.ActionAgentsAndQueues), DtmfResponse: ptr(aa.DtmfTone1), VoiceResponses: []string{"agent"}, Description: ptr(""), AgentTargetType: ptr[int32](0), AgentTarget: ptr(probe), AgentTargetTagTemplateID: ptr(attendantID)})
			}},
		{"menu", "/menus/draft", "New-CsAutoAttendantMenu", "__AllParameterSets", "Menu",
			`{"Name":"menu","MenuOptions":[{"Action":"DisconnectCall","DtmfResponse":"Automatic"}],"DialByNameEnabled":false,"DirectorySearchMethod":"None"}`,
			func(ctx context.Context, s *aa.Service) (any, *resty.Response, error) {
				return s.DraftMenu(ctx, aa.Menu{Name: ptr("menu"), MenuOptions: []aa.MenuOption{{Action: ptr(aa.ActionDisconnectCall), DtmfResponse: ptr(aa.DtmfAutomatic)}}, DialByNameEnabled: ptr(false), DirectorySearchMethod: ptr(aa.SearchNone)})
			}},
		{"call-flow", "/call-flows/draft", "New-CsAutoAttendantCallFlow", "__AllParameterSets", "CallFlow",
			`{"Name":"flow","Menu":{"Name":"menu"},"ForceListenMenuEnabled":false,"RingResourceAccountDelegates":false,"TimeoutThreshold":0,"TimeoutDisconnectPromptType":"Custom","TimeoutDisconnectPromptCustomText":"","AiDisclaimerType":"None","AiDisclaimerCustomText":""}`,
			func(ctx context.Context, s *aa.Service) (any, *resty.Response, error) {
				return s.DraftCallFlow(ctx, aa.CallFlow{Name: ptr("flow"), Menu: &aa.Menu{Name: ptr("menu")}, ForceListenMenuEnabled: ptr(false), RingResourceAccountDelegates: ptr(false), TimeoutThreshold: ptr[int32](0), TimeoutDisconnectPromptType: ptr(aa.PromptKindCustom), TimeoutDisconnectPromptCustomText: ptr(""), AiDisclaimerType: ptr(aa.PromptKindNone), AiDisclaimerCustomText: ptr("")})
			}},
		{"disabled-association", "/call-handling-associations/draft", "New-CsAutoAttendantCallHandlingAssociation", "__AllParameterSets", "CallHandlingAssociation",
			`{"Type":"AfterHours","ScheduleId":"` + probe + `","CallFlowId":"` + defaultFlowID + `","Enabled":false}`,
			func(ctx context.Context, s *aa.Service) (any, *resty.Response, error) {
				return s.DraftCallHandlingAssociation(ctx, aa.CallHandlingAssociation{Type: ptr(aa.AssociationAfterHours), ScheduleID: ptr(probe), CallFlowID: ptr(defaultFlowID), Enabled: ptr(false)})
			}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Source: "msteams/teams_admin_api/auto_attendants/draft_variants_test.go", Mappings: []contracttest.Mapping{{Command: tc.command, ParameterSet: tc.set}}, Steps: []contracttest.Exchange{discovery(t), wire("POST", attendants+tc.route, tc.command, json.RawMessage(tc.body), 200, []byte(`{"`+tc.key+`":`+tc.body+`}`))}, Want: contracttest.Expected{JSON: json.RawMessage(tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, _, err := tc.invoke(ctx, sdk.TeamsAdminAPI.AutoAttendants)
				return contracttest.Observation{Value: value, Err: err}
			}})
		})
	}
}
