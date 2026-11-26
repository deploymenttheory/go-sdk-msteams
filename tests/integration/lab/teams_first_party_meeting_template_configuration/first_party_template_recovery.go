package teamsfirstpartymeetingtemplateconfigurationacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_first_party_meeting_template_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func firstPartyTemplateProjection(value *resource.Configuration) map[string]any {
	templates := []map[string]any{}
	for _, item := range value.TeamsMeetingTemplates {
		if item == nil {
			templates = append(templates, nil)
			continue
		}
		options := []map[string]any{}
		for _, option := range item.TeamsMeetingOptions {
			if option == nil {
				options = append(options, nil)
				continue
			}
			options = append(options, map[string]any{"Identity": option.Identity, "Name": option.Name, "Value": option.Value, "IsLocked": option.IsLocked, "IsHidden": option.IsHidden})
		}
		templates = append(templates, map[string]any{"Identity": item.Identity, "Name": item.Name, "Description": item.Description, "Category": item.Category, "TeamsMeetingOptions": options})
	}
	return map[string]any{"Description": value.Description, "TeamsMeetingTemplates": templates}
}

// The journal persists only the original Global Description, which may be null.
type firstPartyTemplateDescriptionRestore struct{ service *resource.Service }

func (firstPartyTemplateDescriptionRestore) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}

func (firstPartyTemplateDescriptionRestore) Delete(context.Context, lab.Entry) error {
	return lab.ErrUnresolved
}

func decodeDescriptionSnapshot(entry lab.Entry) (*string, error) {
	var original map[string]json.RawMessage
	d := json.NewDecoder(bytes.NewReader(entry.Snapshot))
	if entry.ResourceID != "Global" || d.Decode(&original) != nil || d.Decode(new(any)) != io.EOF || len(original) != 1 {
		return nil, lab.ErrUnresolved
	}
	raw, ok := original["description"]
	if !ok {
		return nil, lab.ErrUnresolved
	}
	var value *string
	if json.Unmarshal(raw, &value) != nil {
		return nil, lab.ErrUnresolved
	}
	return value, nil
}

func sameDescription(a, b *string) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}

func (h firstPartyTemplateDescriptionRestore) Restore(ctx context.Context, entry lab.Entry) error {
	original, err := decodeDescriptionSnapshot(entry)
	if err != nil {
		return err
	}
	// A previous restoration may have succeeded before its receipt was saved.
	current, _, err := h.service.GetByIdentity(ctx, "Global")
	if err != nil {
		return err
	}
	if current != nil && sameDescription(current.Description, original) {
		return nil
	}
	description := policy.Null[string]()
	if original != nil {
		description = policy.Set(*original)
	}
	if _, _, err := h.service.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{Description: description}}); err != nil {
		return err
	}
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		v, _, err := h.service.GetByIdentity(ctx, "Global")
		return err == nil && v != nil && sameDescription(v.Description, original), err
	})
}
