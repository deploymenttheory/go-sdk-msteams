package callintegration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	aa "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/auto_attendants"
	aaacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/auto_attendants"
	"github.com/go-resty/resty/v2"
)

func confirmedCreateRejection(response *resty.Response, err error) bool {
	var api *client.APIError
	return response != nil && response.StatusCode() == 400 && errors.As(err, &api) && api.StatusCode == 400 && slices.Contains(api.Codes, "100001")
}

// A declared validation rejection plus a separate exact-name list can prove
// that this create did not persist. Transport/decoding failures and lost create
// receipts remain ambiguous and retain the original handler's recovery intent.
type rejectedAttendantCleanup struct {
	service          *aa.Service
	fallback         lab.Handler
	entryID, fixture string
	rejected         bool
}

func (h rejectedAttendantCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	if !h.rejected || entry.ID != h.entryID || entry.Fixture != h.fixture {
		return h.fallback.Resolve(ctx, entry)
	}
	const pageSize int32 = 25
	seen := map[string]bool{}
	for page := int32(0); page < 128; page++ {
		skip := page * pageSize
		values, _, err := h.service.List(ctx, &aa.ListOptions{NameFilter: &h.fixture, ExcludeContent: ptr(true), First: ptr(pageSize), Skip: &skip})
		if err != nil {
			return "", err
		}
		if len(values) > int(pageSize) {
			return "", lab.ErrUnresolved
		}
		for _, value := range values {
			if value == nil || value.Name == nil || value.ID == nil || *value.Name == h.fixture || seen[*value.ID] {
				return "", lab.ErrUnresolved
			}
			seen[*value.ID] = true
		}
		if len(values) < int(pageSize) {
			return "", nil
		}
	}
	return "", lab.ErrUnresolved
}
func (h rejectedAttendantCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	return h.fallback.Delete(ctx, entry)
}
func (h rejectedAttendantCleanup) Restore(ctx context.Context, entry lab.Entry) error {
	return h.fallback.Restore(ctx, entry)
}

func TestUnitRejectedCreateNeedsDeclaredFailureAndIndependentAbsence(t *testing.T) {
	t.Parallel()
	const name = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-attendant"
	for _, scenario := range []string{"declared-rejection", "ambiguous-receipt", "still-present", "full-page-then-empty"} {
		t.Run(scenario, func(t *testing.T) {
			listed := `{"AutoAttendants":[],"Diagnostics":[]}`
			if scenario == "still-present" {
				listed = `{"AutoAttendants":[{"Id":"` + attendantID + `","Name":"` + name + `"}],"Diagnostics":[]}`
			}
			var want error
			if scenario == "ambiguous-receipt" || scenario == "still-present" {
				want = lab.ErrUnresolved
			}
			path := "/Teams.VoiceApps/auto-attendants?ExcludeContent=True&First=25&Skip=0&NameFilter=" + name
			if scenario == "ambiguous-receipt" {
				path = "/Teams.VoiceApps/auto-attendants?ExcludeContent=True&NameFilter=" + name
			}
			if scenario == "full-page-then-empty" {
				values := []map[string]string{}
				for i := 0; i < 25; i++ {
					values = append(values, map[string]string{"Id": fmt.Sprintf("00000000-0000-4000-8000-%012d", i), "Name": fmt.Sprintf("other-owned-name-%d", i)})
				}
				raw, err := json.Marshal(map[string]any{"AutoAttendants": values})
				if err != nil {
					t.Fatal(err)
				}
				listed = string(raw)
			}
			steps := []testutil.Exchange{discovery(t),
				exchange("POST", "/Teams.VoiceApps/auto-attendants", "New-CsAutoAttendant", `{"Name":"`+name+`","LanguageId":"en-US","TimeZoneId":"UTC","DefaultCallFlow":{"Id":"`+flowID+`"}}`, 400, `{"code":"100001","errorDetails":[]}`),
				exchange("GET", path, "Get-CsAutoAttendant", "", 200, listed),
			}
			if scenario == "full-page-then-empty" {
				steps = append(steps, exchange("GET", "/Teams.VoiceApps/auto-attendants?ExcludeContent=True&First=25&Skip=25&NameFilter="+name, "Get-CsAutoAttendant", "", 200, `{"AutoAttendants":[]}`))
			}
			result, err := testutil.Run(t.Context(), testutil.Case{Name: scenario, Steps: steps, Want: testutil.Expected{Error: want}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				service := sdk.TeamsAdminAPI.AutoAttendants
				_, response, operationErr := service.Create(ctx, aa.Settings{Name: ptr(name), LanguageID: ptr("en-US"), TimeZoneID: ptr("UTC"), DefaultCallFlow: &aa.CallFlow{ID: ptr(flowID)}})
				rejected := confirmedCreateRejection(response, operationErr)
				if scenario == "ambiguous-receipt" {
					rejected = confirmedCreateRejection(nil, context.DeadlineExceeded)
				}
				handler := rejectedAttendantCleanup{service: service, fallback: aaacc.RecoveryHandlers(sdk)["auto-attendant"], entryID: "planned-entry", fixture: name, rejected: rejected}
				id, err := handler.Resolve(ctx, lab.Entry{ID: "planned-entry", Fixture: name})
				if id != "" {
					t.Fatal("rejected or ambiguous creation must not resolve an invented receipt")
				}
				return testutil.Observation{Err: err}
			}})
			if err != nil || !result.Passed() {
				t.Fatalf("rejected-create cleanup differed: %v %+v", err, result.Differences())
			}
		})
	}
}
