package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_management"
)

const lifecycleFixtureKind = "team-lifecycle"

var lifecycleDirectoryID = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
var lifecycleFixtureName = regexp.MustCompile(`^msteams-sdk-[a-z2-7]{26,64}-(team|migration)$`)

type lifecycleOwner struct {
	Tenant string `json:"tenant_id"`
	UserID string `json:"owner_user_id"`
}

func loadLifecycleOwner(file, binding string) (lifecycleOwner, error) {
	var value lifecycleOwner
	raw, err := labkit.PrivateFile(file)
	if err != nil {
		return value, lab.ErrConfig
	}
	defer clear(raw)
	tenant, _, ok := strings.Cut(binding, "|")
	if json.Unmarshal(raw, &value) != nil || !ok || !strings.EqualFold(value.Tenant, tenant) || !lifecycleDirectoryID.MatchString(value.UserID) {
		return lifecycleOwner{}, lab.ErrConfig
	}
	return value, nil
}

// Recovery only deletes a positively identified run-owned team and its matching
// Unified backing group. Failed/incomplete creation can require manual recovery.
type lifecycleCleanup struct{ sdk *msteams.Client }

func (h lifecycleCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	if !lifecycleFixtureName.MatchString(entry.Fixture) {
		return "", lab.ErrUnresolved
	}
	listed, err := h.sdk.GraphAPI.Groups.ListAll(ctx, group.ListOptions{
		Filter: "displayName eq '" + entry.Fixture + "'", Select: []string{"id", "displayName", "mailNickname"}}, client.PagingOptions{MaxPages: 16})
	if err != nil {
		return "", err
	}
	id := ""
	for _, value := range listed.Items {
		if value.DisplayName == nil || *value.DisplayName != entry.Fixture || value.MailNickname == nil || *value.MailNickname != entry.Fixture || !lifecycleDirectoryID.MatchString(value.ID) {
			return "", lab.ErrUnresolved
		}
		if id != "" && !strings.EqualFold(id, value.ID) {
			return "", lab.ErrUnresolved
		}
		id = value.ID
	}
	if id == "" {
		return "", lab.ErrUnresolved
	}
	return id, nil
}

func (h lifecycleCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !lifecycleFixtureName.MatchString(entry.Fixture) || !lifecycleDirectoryID.MatchString(entry.ResourceID) {
		return lab.ErrUnresolved
	}
	paired, err := h.sdk.GraphAPI.TeamManagement.GetByID(ctx, entry.ResourceID)
	if err != nil {
		return err
	}
	value := paired.Value
	if value == nil || value.GroupSnapshot == nil || value.TeamSnapshot == nil || value.DisplayName == nil || *value.DisplayName != entry.Fixture || value.MailNickname == nil || *value.MailNickname != entry.Fixture ||
		value.TeamSnapshot.DisplayName == nil || *value.TeamSnapshot.DisplayName != entry.Fixture {
		return lab.ErrUnresolved
	}
	unified := false
	for _, kind := range value.GroupSnapshot.GroupTypes {
		unified = unified || kind == "Unified"
	}
	if !unified {
		return lab.ErrUnresolved
	}
	_, err = h.sdk.GraphAPI.Teams.DeleteByID(ctx, entry.ResourceID)
	return err
}

func (lifecycleCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }

func lifecycleLabConfig(t *testing.T) lab.Config {
	t.Helper()
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.Consequential); errors.Is(err, lab.ErrDisabled) {
		t.Skip("team lifecycle lab requires explicit read-only, mutation and consequential opt-ins")
	} else if err != nil {
		t.Fatal("invalid team lifecycle enablement")
	}
	return cfg
}

func lifecycleJournal(t *testing.T, cfg lab.Config, sdk *msteams.Client, binding, suffix string) (*lab.Journal, string, string) {
	t.Helper()
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private lifecycle journal unavailable")
	}
	t.Logf("private recovery journal=%s", journal.FileName())
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{lifecycleFixtureKind: lifecycleCleanup{sdk: sdk}}); err != nil {
			t.Error("team cleanup remains unresolved; retain private journal for recovery")
		}
		if err := journal.Close(); err != nil {
			t.Error("private lifecycle journal close failed")
		}
	})
	fixture, err := journal.FixtureName(suffix)
	if err != nil {
		t.Fatal("lifecycle fixture allocation failed")
	}
	intent, err := journal.PlanCreate(lifecycleFixtureKind, fixture)
	if err != nil {
		t.Fatal("lifecycle creation intent could not be persisted")
	}
	return journal, fixture, intent
}

// TestAccTeamLifecycle creates a fresh fixture for each creation path. Existing
// resource IDs are never accepted as mutation targets. No tenant roles are set.
func TestAccTeamLifecycle(t *testing.T) {
	cfg := lifecycleLabConfig(t)
	_ = labkit.OracleEnabled(t) // Validate the optional oracle before fixture writes.
	sdk, binding, _ := labkit.SDKClient(t)
	owner, err := loadLifecycleOwner(os.Getenv("MSTEAMS_LAB_TEAM_OWNER_FILE"), binding)
	if err != nil {
		t.Fatal("explicit private same-tenant owner fixture is required")
	}
	for _, mode := range []string{"team", "migration"} {
		t.Run(mode, func(t *testing.T) {
			journal, fixture, intent := lifecycleJournal(t, cfg, sdk, binding, mode)
			ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
			defer cancel()
			var result management.CreationResult
			if mode == "team" {
				result, err = sdk.GraphAPI.TeamManagement.Create(ctx, management.CreateRequest{
					DisplayName: fixture, MailNickname: &fixture, Owner: owner.UserID,
					RetainCreatedGroup: true, Options: management.CreationOptions{Timeout: cfg.Timeout}})
				if result.CreatedGroupOwned && result.ID != "" {
					if err := journal.Applied(intent, result.ID); err != nil {
						t.Fatal("created group identity could not be journalled")
					}
				}
			} else {
				request, buildErr := group.NewMicrosoft365Group(fixture, fixture)
				if buildErr != nil {
					t.Fatal("migration fixture input invalid")
				}
				request.OwnerIDs, request.MemberIDs = []string{owner.UserID}, []string{owner.UserID}
				created, _, createErr := sdk.GraphAPI.Groups.Create(ctx, request)
				if created != nil && lifecycleDirectoryID.MatchString(created.ID) {
					if err := journal.Applied(intent, created.ID); err != nil {
						t.Fatal("migration group identity could not be journalled")
					}
				}
				if createErr != nil || created == nil {
					t.Fatal("migration group creation failed; retain journal")
				}
				result, err = sdk.GraphAPI.TeamManagement.MigrateGroup(ctx, created.ID, management.CreationOptions{Timeout: cfg.Timeout})
			}
			if err != nil || result.Value == nil || !lifecycleDirectoryID.MatchString(result.ID) {
				t.Fatal("team creation/migration was not confirmed; retain journal")
			}
			cancel()
			checkLifecycleTeam(t, cfg, sdk, result.ID, fixture, owner.UserID)
			deleteCtx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
			defer stop()
			// The same ownership-checked journal handler performs the explicit
			// delete and persists its result before the post-delete observation.
			if _, err := journal.Recover(deleteCtx, map[string]lab.Handler{lifecycleFixtureKind: lifecycleCleanup{sdk: sdk}}); err != nil {
				t.Fatal("explicit owned-team deletion remains unresolved")
			}
			if err := lifecycleObserveAbsent(deleteCtx, sdk, result.ID); err != nil {
				t.Fatal("team/group readback did not confirm deletion")
			}
			t.Log("team lifecycle observations passed; education and delegated qualification remain separate")
		})
	}
}

func checkLifecycleTeam(t *testing.T, cfg lab.Config, sdk *msteams.Client, id, fixture, ownerID string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	paired, err := sdk.GraphAPI.TeamManagement.GetByID(ctx, id)
	if err != nil || paired.Value == nil || paired.Value.DisplayName == nil || *paired.Value.DisplayName != fixture {
		t.Fatal("paired read did not confirm creation")
	}
	comparePowerShellTeam(t, ctx, paired.Value, fixture, ownerID)
	for _, user := range []*string{nil, &ownerID} {
		listed, err := sdk.GraphAPI.TeamManagement.List(ctx, management.ListOptions{User: user, DisplayName: fixture, Traversal: client.PagingOptions{MaxPages: 16}, Timeout: cfg.Timeout})
		if err != nil || !listed.Complete {
			t.Fatal("bounded tenant/user enriched list did not complete")
		}
		found := false
		for _, value := range listed.Items {
			found = found || strings.EqualFold(value.ID, id)
		}
		if !found {
			t.Fatal("tenant/user enriched list did not confirm fixture")
		}
	}
	description, disabled := "SDK lifecycle updated description", false
	updated, err := sdk.GraphAPI.TeamManagement.UpdateByID(ctx, id, management.UpdateRequest{
		Properties: group.UpdateRequest{Description: &description}, Settings: team.Settings{
			FunSettings: &team.FunSettings{AllowGiphy: &disabled}}})
	if err != nil || updated.Value == nil || updated.Value.Description == nil || *updated.Value.Description != description || updated.Value.TeamSnapshot == nil ||
		updated.Value.TeamSnapshot.FunSettings == nil || updated.Value.TeamSnapshot.FunSettings.AllowGiphy == nil || *updated.Value.TeamSnapshot.FunSettings.AllowGiphy {
		t.Fatal("paired readback did not confirm group/settings update")
	}
	comparePowerShellTeam(t, ctx, updated.Value, fixture, ownerID)
	cancel()
	for _, archived := range []bool{true, false} {
		actionCtx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
		result, err := sdk.GraphAPI.TeamManagement.SetArchivedState(actionCtx, id, archived, management.ArchiveStateOptions{Wait: client.WaitOptions{Timeout: cfg.Timeout}})
		if err != nil || !result.Confirmed {
			stop()
			t.Fatal("archive transition was not confirmed; do not resubmit")
		}
		comparePowerShellTeam(t, actionCtx, result.After, fixture, ownerID)
		stop()
	}
	photoCtx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	var photo bytes.Buffer
	pixels := image.NewRGBA(image.Rect(0, 0, 48, 48))
	draw.Draw(pixels, pixels.Bounds(), image.NewUniform(color.RGBA{R: 40, G: 110, B: 190, A: 255}), image.Point{}, draw.Src)
	if err := jpeg.Encode(&photo, pixels, nil); err != nil {
		t.Fatal("synthetic valid JPEG unavailable")
	}
	if _, err := sdk.GraphAPI.Teams.UploadPicture(photoCtx, id, io.NopCloser(bytes.NewReader(photo.Bytes()))); err != nil {
		t.Fatal("owned-team picture upload failed")
	}
	// Graph may resize/re-encode a stored photo. Confirm a nonempty valid JPEG,
	// rather than inventing a server byte-identity guarantee.
	var downloaded bytes.Buffer
	response, count, err := sdk.GraphAPI.Scope().Download(photoCtx, client.Request{Method: http.MethodGet,
		Path: []string{"v1.0", "groups", id, "photo", "$value"}}, &downloaded, team.MaxPictureBytes)
	if err != nil || response == nil || response.StatusCode() != http.StatusOK || count == 0 {
		t.Fatal("uploaded group photo readback failed")
	}
	if _, err := jpeg.DecodeConfig(bytes.NewReader(downloaded.Bytes())); err != nil {
		t.Fatal("stored group photo was not a valid JPEG")
	}
}

func lifecycleObserveAbsent(ctx context.Context, sdk *msteams.Client, id string) error {
	for {
		_, _, groupErr := sdk.GraphAPI.Groups.GetByID(ctx, id, group.ReadOptions{})
		_, _, teamErr := sdk.GraphAPI.Teams.GetByID(ctx, id)
		missing := func(err error) bool {
			var apiError *client.APIError
			return errors.As(err, &apiError) && apiError.Backend == client.Graph && apiError.StatusCode == http.StatusNotFound
		}
		if missing(groupErr) && missing(teamErr) {
			return nil
		}
		for _, err := range []error{groupErr, teamErr} {
			if err != nil && !missing(err) {
				return err
			}
		}
		timer := time.NewTimer(2 * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func TestAccRecoverTeamLifecycle(t *testing.T) {
	cfg := lifecycleLabConfig(t)
	sdk, binding, _ := labkit.SDKClient(t)
	file := os.Getenv("MSTEAMS_LAB_RECOVERY_JOURNAL")
	if file == "" {
		t.Skip("explicit private recovery journal is required")
	}
	journal, err := lab.OpenJournal(file, binding)
	if err != nil {
		t.Fatal("private lifecycle recovery journal unavailable for this binding")
	}
	defer func() { _ = journal.Close() }()
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	if _, err := journal.Recover(ctx, map[string]lab.Handler{lifecycleFixtureKind: lifecycleCleanup{sdk: sdk}}); err != nil {
		t.Fatal("lifecycle recovery remains unresolved")
	}
}
