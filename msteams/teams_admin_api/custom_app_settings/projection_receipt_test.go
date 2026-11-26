package customappsettings

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

type projectionClientFunc func(context.Context, client.Request, any) (*resty.Response, error)

func (f projectionClientFunc) Do(ctx context.Context, request client.Request, value any) (*resty.Response, error) {
	return f(ctx, request, value)
}

func TestUnitSourceSetAmbiguousActionReceipts(t *testing.T) {
	for _, item := range []struct {
		name     string
		response *resty.Response
		err      error
	}{
		{name: "lost-response", err: client.ErrTransport},
		{name: "accepted-canceled", response: &resty.Response{RawResponse: &http.Response{StatusCode: 204}}, err: context.Canceled},
	} {
		t.Run(item.name, func(t *testing.T) {
			attempts := 0
			service := &Service{client: projectionClientFunc(func(_ context.Context, req client.Request, _ any) (*resty.Response, error) {
				attempts++
				if req.Method != http.MethodPut || req.Replay != client.ReplayNever {
					t.Fatal("action must be an unreplayed PUT")
				}
				return item.response, item.err
			})}
			read := &resty.Response{RawResponse: &http.Response{StatusCode: 200}}
			projection, err := ProjectSideloadedAppsInteraction([]byte(`{}`), false)
			if err != nil {
				t.Fatal(err)
			}
			prepared := &SideloadedUpdate{owner: service, readResponse: read, projection: projection}
			result, response, err := service.UpdateSideloadedAppsInteraction(t.Context(), prepared)
			if !errors.Is(err, item.err) || attempts != 1 || result == nil || !result.UpdateInvoked || result.ReadResponse != read || result.UpdateResponse != item.response || response != item.response || result.HTTPAccepted != (item.response != nil) {
				t.Fatal("ambiguous/accepted failure lost action state")
			}
		})
	}
}
