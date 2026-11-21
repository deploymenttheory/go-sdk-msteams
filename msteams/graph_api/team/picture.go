package team

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// MaxPictureBytes is Graph's four-MB profile-photo limit.
const MaxPictureBytes int64 = 4 << 20

// UploadPicture streams the original JPEG bytes to the backing group's photo,
// matching Set-TeamPicture's route and its group-photo permission requirements.
// It takes ownership of source immediately and closes it exactly once, including
// validation failure. Close must unblock Read. No conversion, buffering, redirect
// or replay is performed; a failed upload can leave partial remote state.
func (s *Service) UploadPicture(ctx context.Context, id string, source io.ReadCloser) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, rejectPicture(ctx, source, err)
	}
	if !directoryID.MatchString(id) {
		return nil, rejectPicture(ctx, source, client.ErrRequest)
	}
	response, err := s.scope.Upload(ctx, request(http.MethodPut, "groups", id, "photo", "$value"), source, "image/jpeg", MaxPictureBytes, nil)
	if err == nil && response.StatusCode() != http.StatusOK && response.StatusCode() != http.StatusNoContent {
		err = ErrUnexpectedStatus
	}
	return response, err
}

func rejectPicture(ctx context.Context, source io.ReadCloser, reason error) error {
	// Let the shared stream boundary handle typed nils and close ownership. An
	// invalid method fails before any session access, token acquisition or I/O.
	var scope *client.Scope
	_, err := scope.Upload(ctx, client.Request{}, source, "image/jpeg", MaxPictureBytes, nil)
	if errors.Is(err, client.ErrStreamClose) {
		return errors.Join(reason, client.ErrStreamClose)
	}
	return reason
}
