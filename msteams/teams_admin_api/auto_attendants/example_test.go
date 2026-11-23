package autoattendants_test

import (
	"context"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	aa "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/auto_attendants"
)

// This example compiles a real SDK workflow. It performs no network operation
// during go test; the application supplies its own auth configuration and context.
func ExampleService_Create() {
	create := func(ctx context.Context, cfg *auth.Config) (id string, err error) {
		sdk, err := msteams.NewClient(cfg)
		if err != nil {
			return "", err
		}
		defer func() { err = errors.Join(err, sdk.Close()) }()
		name, language, zone := "Main line", "en-US", "UTC"
		active, text := aa.PromptTextToSpeech, "Thanks for calling."
		action, dtmf := aa.ActionDisconnectCall, aa.DtmfAutomatic
		flowName, menuName := "Default call flow", "Default menu"
		attendant, _, err := sdk.TeamsAdminAPI.AutoAttendants.Create(ctx, aa.Settings{
			Name:       &name,
			LanguageID: &language,
			TimeZoneID: &zone,
			DefaultCallFlow: &aa.CallFlow{
				Name:      &flowName,
				Greetings: []aa.Prompt{{ActiveType: &active, TextToSpeechPrompt: &text}},
				Menu:      &aa.Menu{Name: &menuName, MenuOptions: []aa.MenuOption{{Action: &action, DtmfResponse: &dtmf}}},
			},
		})
		if err != nil || attendant.ID == nil {
			return "", err
		}
		return *attendant.ID, nil
	}
	_ = create
}

// The service replaces the whole attendant, so update from a fresh read as
// Set-CsAutoAttendant -Instance does.
func ExampleAutoAttendant_Settings() {
	update := func(ctx context.Context, sdk *msteams.Client, id string) error {
		s := sdk.TeamsAdminAPI.AutoAttendants
		attendant, _, err := s.GetByID(ctx, id)
		if err != nil {
			return err
		}
		settings := attendant.Settings()
		zone := "GMT Standard Time"
		settings.TimeZoneID = &zone
		_, _, err = s.UpdateByID(ctx, id, settings)
		return err
	}
	_ = update
}
