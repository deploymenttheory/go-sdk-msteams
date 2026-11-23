package teamtemplateacc

import (
	"context"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	template "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/team_template"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const templateLocale = "en-US"

type templateCleanup struct{ service *template.Service }

func ownedTemplateReference(reference string) bool {
	parts := strings.Split(reference, "/")
	return len(parts) == 7 && parts[0] == "" && parts[1] == "api" && strings.EqualFold(parts[2], "teamtemplates") &&
		parts[3] == "v1.0" && parts[4] != "" && parts[4] != "." && parts[4] != ".." && parts[5] == "Tenant" && strings.EqualFold(parts[6], templateLocale)
}

func ownedTemplateName(name string) bool {
	return strings.HasPrefix(name, "msteams-sdk-") && strings.HasSuffix(name, "-template")
}

func (h templateCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	if !ownedTemplateName(entry.Fixture) {
		return "", lab.ErrUnresolved
	}
	items, _, err := h.service.List(ctx, nil)
	if err != nil {
		return "", err
	}
	if items.IsIncomplete {
		return "", lab.ErrUnresolved
	}
	count, reference := 0, ""
	for _, item := range items.Items {
		if item.Name == nil || *item.Name != entry.Fixture {
			continue
		}
		count++
		if item.ODataID == nil || !ownedTemplateReference(*item.ODataID) || (item.Scope != nil && !strings.EqualFold(*item.Scope, "Tenant")) {
			return "", lab.ErrUnresolved
		}
		reference = *item.ODataID
	}
	if count > 1 {
		return "", lab.ErrUnresolved
	}
	if count == 0 {
		// A planned creation has no previously verified reference. An empty
		// projection/list alone cannot prove that an accepted/lost creation
		// never existed; retain it for later reconciliation.
		return "", lab.ErrUnresolved
	}
	value, _, err := h.service.GetByID(ctx, reference)
	if err != nil {
		return "", err
	}
	if value.DisplayName == nil || *value.DisplayName != entry.Fixture {
		return "", lab.ErrUnresolved
	}
	return reference, nil
}

func (h templateCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !ownedTemplateName(entry.Fixture) || !ownedTemplateReference(entry.ResourceID) {
		return lab.ErrUnresolved
	}
	value, _, err := h.service.GetByID(ctx, entry.ResourceID)
	if labkit.IsHTTP(err, 404) {
		return h.absent(ctx, entry)
	}
	if err != nil {
		return err
	}
	if value.DisplayName == nil || *value.DisplayName != entry.Fixture {
		return lab.ErrUnresolved
	}
	_, err = h.service.DeleteByID(ctx, entry.ResourceID)
	if err != nil && !labkit.IsHTTP(err, 404) {
		return err
	}
	return h.absent(ctx, entry)

}

func (h templateCleanup) absent(ctx context.Context, entry lab.Entry) error {
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		_, _, err := h.service.GetByID(ctx, entry.ResourceID)
		if err == nil {
			return false, nil
		}
		if !labkit.IsHTTP(err, 404) {
			return false, err
		}
		values, _, err := h.service.List(ctx, nil)
		if err != nil {
			return false, err
		}
		if values == nil || values.IsIncomplete {
			return false, lab.ErrUnresolved
		}
		for _, value := range values.Items {
			if value.Name != nil && *value.Name == entry.Fixture {
				return false, nil
			}
		}
		return true, nil
	})
}

func (templateCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }
