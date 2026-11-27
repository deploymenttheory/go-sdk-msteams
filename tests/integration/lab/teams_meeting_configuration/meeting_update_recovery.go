package teamsmeetingconfigurationacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	meeting "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// meetingSnapshot is the journaled original pair. Both keys are required; a
// JSON null original is restored as an explicit null.
type meetingSnapshot struct {
	CustomFooterText *string
	HelpURL          *string
}

func decodeMeetingSnapshot(raw []byte) (meetingSnapshot, bool) {
	var fields map[string]json.RawMessage
	d := json.NewDecoder(bytes.NewReader(raw))
	if d.Decode(&fields) != nil || d.Decode(new(any)) != io.EOF || len(fields) != 2 {
		return meetingSnapshot{}, false
	}
	var snapshot meetingSnapshot
	for name, target := range map[string]**string{"CustomFooterText": &snapshot.CustomFooterText, "HelpURL": &snapshot.HelpURL} {
		value, ok := fields[name]
		if !ok {
			return meetingSnapshot{}, false
		}
		trimmed := bytes.TrimSpace(value)
		if bytes.Equal(trimmed, []byte("null")) {
			continue
		}
		if len(trimmed) == 0 || trimmed[0] != '"' || json.Unmarshal(trimmed, target) != nil {
			return meetingSnapshot{}, false
		}
	}
	return snapshot, true
}

func sameText(left, right *string) bool {
	return (left == nil && right == nil) || (left != nil && right != nil && *left == *right)
}

func textField(value *string) policy.Field[string] {
	if value == nil {
		return policy.Null[string]()
	}
	return policy.Set(*value)
}

// meetingConfigurationCleanup restores only the journaled Global
// CustomFooterText and HelpURL pair through the same public SDK service.
type meetingConfigurationCleanup struct{ service *meeting.Service }

func (meetingConfigurationCleanup) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}

func (meetingConfigurationCleanup) Delete(context.Context, lab.Entry) error {
	return lab.ErrUnresolved
}

func (h meetingConfigurationCleanup) Restore(ctx context.Context, entry lab.Entry) error {
	snapshot, ok := decodeMeetingSnapshot(entry.Snapshot)
	if entry.ResourceID != "Global" || !ok {
		return lab.ErrUnresolved
	}
	restored := func(v *meeting.Configuration) bool {
		return v != nil && sameText(v.CustomFooterText, snapshot.CustomFooterText) && sameText(v.HelpURL, snapshot.HelpURL)
	}
	// A previous restoration may have succeeded before its receipt was saved.
	before, _, err := h.service.GetByIdentity(ctx, "Global")
	if err != nil {
		return err
	}
	if restored(before) {
		return nil
	}
	request := meeting.UpdateRequest{ScalarSettings: meeting.ScalarSettings{
		CustomFooterText: textField(snapshot.CustomFooterText),
		HelpURL:          textField(snapshot.HelpURL),
	}}
	if _, _, err := h.service.UpdateByIdentity(ctx, "Global", request); err != nil {
		return err
	}
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		v, _, e := h.service.GetByIdentity(ctx, "Global")
		return e == nil && restored(v), e
	})
}
