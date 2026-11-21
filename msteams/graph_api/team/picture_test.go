package team_test

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
)

type pictureSource struct {
	*strings.Reader
	closes   int
	closeErr error
}

func (s *pictureSource) Close() error { s.closes++; return s.closeErr }

func TestUnitUploadPictureUsesGroupPhotoRouteAndOriginalBytes(t *testing.T) {
	t.Parallel()
	for _, status := range []int{200, 204, 503} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			jpeg := "\xff\xd8\xff\x00fixture-binary\xff\xd9"
			sdk, _, calls := fixture(t, "commercial", auth.Application, exchange{method: "PUT", path: "/v1.0/groups/" + teamID + "/photo/$value", binary: jpeg, contentType: "image/jpeg", status: status})
			source := &pictureSource{Reader: strings.NewReader(jpeg)}
			response, err := sdk.GraphAPI.Teams.UploadPicture(t.Context(), teamID, source)
			if errors.Is(err, auth.ErrForbidden) {
				return
			}
			if response == nil || response.StatusCode() != status || (err != nil) != (status == 503) || source.closes != 1 || *calls != 1 {
				t.Fatal("picture transfer changed bytes, lost metadata, leaked source or replayed")
			}
		})
	}
}

func TestUnitUploadPictureValidationClosesSourcesWithoutIO(t *testing.T) {
	t.Parallel()
	sdk, p, _ := fixture(t, "commercial", auth.Application)
	for _, id := range []string{"", "../groups"} {
		source := &pictureSource{Reader: strings.NewReader("fixture")}
		if response, err := sdk.GraphAPI.Teams.UploadPicture(t.Context(), id, source); response != nil || !errors.Is(err, client.ErrRequest) || source.closes != 1 {
			t.Fatal("invalid picture identity was sent or source leaked")
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	source := &pictureSource{Reader: strings.NewReader("fixture"), closeErr: errors.New("private-close-detail")}
	if _, err := sdk.GraphAPI.Teams.UploadPicture(ctx, teamID, source); !errors.Is(err, context.Canceled) || !errors.Is(err, client.ErrStreamClose) || source.closes != 1 || strings.Contains(err.Error(), "private-close-detail") {
		t.Fatal("cancellation/close failure lost or exposed")
	}
	var typedNil *pictureSource
	if _, err := sdk.GraphAPI.Teams.UploadPicture(t.Context(), teamID, typedNil); !errors.Is(err, client.ErrRequest) {
		t.Fatal("typed nil source accepted")
	}
	var absent *team.Service
	source = &pictureSource{Reader: strings.NewReader("fixture")}
	if _, err := absent.UploadPicture(t.Context(), teamID, source); !errors.Is(err, client.ErrClosed) || source.closes != 1 {
		t.Fatal("nil service leaked source")
	}
	if p.calls != 0 {
		t.Fatal("picture validation acquired tokens")
	}
}

func TestUnitUploadPictureEnforcesGraphSizeLimit(t *testing.T) {
	t.Parallel()
	sdk, _, calls := fixture(t, "commercial", auth.Application, exchange{method: "PUT", path: "/v1.0/groups/" + teamID + "/photo/$value", status: 200})
	source := &pictureSource{Reader: strings.NewReader(strings.Repeat("x", int(team.MaxPictureBytes)+1))}
	if _, err := sdk.GraphAPI.Teams.UploadPicture(t.Context(), teamID, source); !errors.Is(err, client.ErrUploadTooLarge) || source.closes != 1 || *calls != 1 {
		t.Fatal("oversize picture accepted, source leaked or attempt replayed")
	}
}
