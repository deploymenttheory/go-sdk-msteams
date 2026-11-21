package teammanagement_test

import (
	"context"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_management"
)

func TestUnitCreationValidationPrecedesCredentialsAndWrites(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"empty-name", "empty-alias", "empty-description", "empty-classification", "invalid-template", "class-explicit-visibility", "class-empty-visibility", "invalid-visibility", "negative-timeout", "negative-attempts", "negative-interval", "negative-cleanup-timeout", "application-without-owner"} {
		mode := auth.Delegated
		want := client.ErrRequest
		if name == "application-without-owner" {
			mode = auth.Application
			want = directoryuser.ErrCurrentUser
		}
		creationCase(t, "CreateTeam", contracttest.Case{Name: name, Scope: "local", Mode: mode, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value := creationRequest()
			switch name {
			case "empty-name":
				value.DisplayName = " "
			case "empty-alias":
				empty := ""
				value.MailNickname = &empty
			case "empty-description":
				empty := ""
				value.Description = &empty
			case "empty-classification":
				empty := ""
				value.Classification = &empty
			case "invalid-template":
				value.Template = "standard"
			case "class-explicit-visibility":
				visibility := group.Private
				value.Template = management.EducationClass
				value.Visibility = &visibility
			case "class-empty-visibility":
				visibility := group.Visibility("")
				value.Template = management.EducationClass
				value.Visibility = &visibility
			case "invalid-visibility":
				visibility := group.Visibility("future")
				value.Visibility = &visibility
			case "negative-timeout":
				value.Options.Timeout = -1
			case "negative-attempts":
				value.Options.MaxAttempts = -1
			case "negative-interval":
				value.Options.Interval = -1
			case "negative-cleanup-timeout":
				value.Options.CleanupTimeout = -1
			case "application-without-owner":
				value.Owner = ""
			}
			result, err := sdk.GraphAPI.TeamManagement.Create(ctx, value)
			if err != nil {
				return contracttest.Observation{Response: result.Response, Err: err}
			}
			return contracttest.Observation{Response: result.Response, Err: err}
		}, Want: contracttest.Expected{Error: want}})
	}
}

func TestUnitArchiveValidationStopsBeforeAnyReadOrWrite(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"unarchive-spo-false", "application-spo-false", "negative-timeout", "negative-interval", "negative-polls", "propagation-not-evidenced"} {
		mode := auth.Delegated
		want := client.ErrOperation
		if name == "unarchive-spo-false" {
			want = management.ErrUnarchiveOption
		}
		if name == "application-spo-false" {
			mode = auth.Application
			want = team.ErrArchiveOption
		}
		runCase(t, "Set-TeamArchivedState", contracttest.Case{Name: name, Scope: "local", Mode: mode, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			options := management.ArchiveStateOptions{}
			archived := true
			switch name {
			case "unarchive-spo-false":
				archived = false
				disabled := false
				options.Archive.ShouldSetSpoSiteReadOnlyForMembers = &disabled
			case "application-spo-false":
				disabled := false
				options.Archive.ShouldSetSpoSiteReadOnlyForMembers = &disabled
			case "negative-timeout":
				options.Wait.Timeout = -1
			case "negative-interval":
				options.Wait.Interval = -1
			case "negative-polls":
				options.Wait.MaxPolls = -1
			case "propagation-not-evidenced":
				options.Wait.PropagationWindow = time.Minute
				options.Wait.PropagationRules = []client.PropagationRule{{StatusCode: 404, Code: "notFound"}}
			}
			result, err := sdk.GraphAPI.TeamManagement.SetArchivedState(ctx, groupID, archived, options)
			if err != nil {
				return contracttest.Observation{Response: result.Response, Err: err}
			}
			return contracttest.Observation{Response: result.Response, Err: err}
		}, Want: contracttest.Expected{Error: want}})
	}
}

func TestUnitListingLocalValidationAndCurrentUserMode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"batch-negative", "batch-oversized", "attempts-negative", "timeout-negative", "delay-negative", "pages-negative", "wrong-paging", "control-search", "current-application"} {
		want := client.ErrRequest
		mode := auth.Delegated
		if name == "pages-negative" || name == "wrong-paging" {
			want = client.ErrPagination
		}
		if name == "current-application" {
			mode = auth.Application
			want = directoryuser.ErrCurrentUser
		}
		listCase(t, contracttest.Case{Name: name, Scope: "local", Mode: mode, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			options := management.ListOptions{}
			switch name {
			case "batch-negative":
				options.BatchSize = -1
			case "batch-oversized":
				options.BatchSize = 21
			case "attempts-negative":
				options.MaxThrottleAttempts = -1
			case "timeout-negative":
				options.Timeout = -1
			case "delay-negative":
				options.ThrottleDelay = -1
			case "pages-negative":
				options.Traversal.MaxPages = -1
			case "wrong-paging":
				options.Traversal.Style = client.OffsetPaging
			case "control-search":
				options.DisplayName = "private\nsearch"
			case "current-application":
				user := ""
				options.User = &user
			}
			result, err := sdk.GraphAPI.TeamManagement.List(ctx, options)
			if err != nil {
				return contracttest.Observation{Response: result.Response, Err: err}
			}
			return contracttest.Observation{Response: result.Response, Err: err}
		}, Want: contracttest.Expected{Error: want}})
	}
}

func TestUnitInvalidInputAndCancellationPerformNoIO(t *testing.T) {
	t.Parallel()
	for _, method := range []string{"get", "update"} {
		for _, condition := range []string{"invalid-id", "nil-context", "cancelled", "closed", "nil-service", "wrong-backend"} {
			want := client.ErrRequest
			switch condition {
			case "nil-context":
				want = client.ErrNilContext
			case "cancelled":
				want = context.Canceled
			case "closed", "nil-service":
				want = client.ErrClosed
			case "wrong-backend":
				want = client.ErrDestination
			}
			runCase(t, map[string]string{"get": "Get-Team", "update": "Set-Team"}[method], contracttest.Case{Name: method + "-" + condition, Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				service, id := sdk.GraphAPI.TeamManagement, groupID
				switch condition {
				case "invalid-id":
					id = "../private"
				case "nil-context":
					if method == "get" {
						//nolint:contextcheck,staticcheck // Deliberately exercise rejection of nil contexts before authentication or I/O.
						_, err := service.GetByID(nil, id)
						return contracttest.Observation{Err: err}
					}
					//nolint:contextcheck,staticcheck // Deliberately exercise rejection of nil contexts before authentication or I/O.
					_, err := service.UpdateByID(nil, id, management.UpdateRequest{})
					return contracttest.Observation{Err: err}
				case "cancelled":
					cancelled, cancel := context.WithCancel(ctx)
					cancel()
					ctx = cancelled
				case "closed":
					if err := sdk.Close(); err != nil {
						return contracttest.Observation{Err: err}
					}
				case "nil-service":
					service = nil
				case "wrong-backend":
					service = management.NewService(sdk.TeamsAdminAPI.Scope())
				}
				if method == "get" {
					_, err := service.GetByID(ctx, id)
					return contracttest.Observation{Err: err}
				}
				_, err := service.UpdateByID(ctx, id, management.UpdateRequest{})
				return contracttest.Observation{Err: err}
			}, Want: contracttest.Expected{Error: want}})
		}
	}
}
