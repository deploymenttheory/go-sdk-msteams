package labkit

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/go-resty/resty/v2"
)

func TestUnitConfigurationRecoveryRejectsForeignTargetsAndSnapshots(t *testing.T) {
	t.Parallel()
	h := ConfigurationRestore{}
	for _, entry := range []lab.Entry{
		{ResourceID: "Tag:foreign", Snapshot: json.RawMessage(`{"value":true}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":null}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":"false"}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":false,"unknown":true}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":false} {}`)},
	} {
		if !errors.Is(h.Restore(t.Context(), entry), lab.ErrUnresolved) {
			t.Fatal("configuration recovery accepted an unsupported target or snapshot")
		}
	}
}

func TestUnitConfigurationRecoveryVerifiesTheOriginalAfterWriting(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		accepted bool
	}{{"persisted", true}, {"accepted-without-persistence", false}} {
		t.Run(tc.name, func(t *testing.T) {
			current, reads, writes := false, 0, 0
			readErr := errors.New("restoration read-back differs")
			h := ConfigurationRestore{SingletonConfiguration{Field: "setting", Get: func(context.Context) (map[string]any, *resty.Response, error) {
				reads++
				if reads > 1 && !current {
					return nil, nil, readErr
				}
				return map[string]any{"setting": &current}, nil, nil
			}, Set: func(_ context.Context, value bool) (*resty.Response, error) {
				writes++
				if !value {
					t.Fatal("recovery did not send the journaled original")
				}
				current = tc.accepted
				return nil, nil
			}}}
			err := h.Restore(t.Context(), lab.Entry{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":true}`)})
			if reads != 2 || writes != 1 || (tc.accepted && err != nil) || (!tc.accepted && !errors.Is(err, readErr)) {
				t.Fatal("restoration trusted write acceptance without independent verification")
			}
		})
	}
}
