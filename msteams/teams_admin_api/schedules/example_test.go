package schedules_test

import (
	"context"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	sched "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/schedules"
)

// This example compiles a real SDK workflow. It performs no network operation
// during go test; the application supplies its own auth configuration and context.
// The service validates the business hours before the schedule is created.
func ExampleService_Create() {
	create := func(ctx context.Context, cfg *auth.Config) (id string, err error) {
		sdk, err := msteams.NewClient(cfg)
		if err != nil {
			return "", err
		}
		defer func() { err = errors.Join(err, sdk.Close()) }()
		schedules := sdk.TeamsAdminAPI.Schedules
		hours, _, err := schedules.NewTimeRange(ctx, "9:00", "17:00")
		if err != nil {
			return "", err
		}
		weekdays := []sched.TimeRange{*hours}
		schedule, _, err := schedules.Create(ctx, sched.CreateRequest{
			Name: "Business hours",
			WeeklyRecurrentSchedule: &sched.WeeklyRecurrentSchedule{
				MondayHours: weekdays, TuesdayHours: weekdays, WednesdayHours: weekdays,
				ThursdayHours: weekdays, FridayHours: weekdays,
			},
		})
		if err != nil || schedule.ID == nil {
			return "", err
		}
		return *schedule.ID, nil
	}
	_ = create
}

// ListHolidays and the export/import pair need an existing auto attendant.
func ExampleService_GetHolidaysContent() {
	copyHolidays := func(ctx context.Context, sdk *msteams.Client, from, to string) error {
		schedules := sdk.TeamsAdminAPI.Schedules
		content, _, err := schedules.GetHolidaysContent(ctx, from)
		if err != nil {
			return err
		}
		result, _, err := schedules.ImportHolidays(ctx, to, content)
		if err != nil {
			return err
		}
		for _, status := range result.Statuses {
			if status.Succeeded == nil || !*status.Succeeded {
				return errors.New("holiday import failed")
			}
		}
		return nil
	}
	_ = copyHolidays
}
