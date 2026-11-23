package mainlineattendants_test

import (
	"context"
	"errors"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	ma "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/mainline_attendants"
)

// This example compiles a real SDK workflow. It performs no network operation
// during go test; the application supplies its own auth configuration and context.
func ExampleService_CreateSpamDetectionTemplate() {
	create := func(ctx context.Context, cfg *auth.Config) (id string, err error) {
		sdk, err := msteams.NewClient(cfg)
		if err != nil {
			return "", err
		}
		defer func() { err = errors.Join(err, sdk.Close()) }()
		// EnableSpamDetection and Action default to true and DisconnectCall.
		template, _, err := sdk.TeamsAdminAPI.MainlineAttendants.CreateSpamDetectionTemplate(ctx, ma.CreateSpamDetectionTemplateRequest{
			Name:        "Front desk spam filter",
			Description: "Disconnects likely spam calls",
		})
		if err != nil || template.Identity == nil {
			return "", err
		}
		return *template.Identity, nil
	}
	_ = create
}

// A question and answer flow indexes its knowledge base asynchronously; the
// service rejects updates until indexing finishes.
func ExampleService_CreateQuestionAnswerFlow() {
	create := func(ctx context.Context, sdk *msteams.Client, faq []byte) (*ma.Flow, error) {
		s := sdk.TeamsAdminAPI.MainlineAttendants
		flow, _, err := s.CreateQuestionAnswerFlow(ctx, ma.CreateQuestionAnswerFlowRequest{
			Name:        "Opening hours",
			Description: "Answers questions about opening hours",
			KnowledgeBase: ma.KnowledgeBase{
				Files: []ma.KnowledgeBaseFile{{Name: "faq.txt", Content: faq, Description: "Frequently asked questions"}},
				URLs:  []ma.KnowledgeBaseURL{{URL: "https://www.contoso.com/hours", Description: "Hours page"}},
			},
		})
		if err != nil || flow.Identity == nil {
			return nil, err
		}
		for {
			indexing, err := flow.Indexing()
			if err != nil || !indexing {
				return flow, err
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(5 * time.Second):
			}
			if flow, _, err = s.GetQuestionAnswerFlowByID(ctx, *flow.Identity); err != nil {
				return nil, err
			}
		}
	}
	_ = create
}

// GetTenantInformation returns the supported languages and voices that
// Get-CsMainlineAttendantSupportedLanguages and -SupportedVoices project.
func ExampleService_GetTenantInformation() {
	languages := func(ctx context.Context, sdk *msteams.Client) ([]string, error) {
		information, _, err := sdk.TeamsAdminAPI.MainlineAttendants.GetTenantInformation(ctx)
		if err != nil {
			return nil, err
		}
		ids := make([]string, 0, len(information.SupportedLanguages))
		for _, language := range information.SupportedLanguages {
			if language.ID != nil {
				ids = append(ids, *language.ID)
			}
		}
		return ids, nil
	}
	_ = languages
}
