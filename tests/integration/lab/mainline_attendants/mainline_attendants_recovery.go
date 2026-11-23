package mainlineattendantsacc

import (
	"context"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	ma "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/mainline_attendants"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	"github.com/go-resty/resty/v2"
)

// Journal kinds and fixture labels for the three owned resource types.
const (
	spamKind     = "mainline-spam-template"
	bookingKind  = "mainline-booking-flow"
	qaKind       = "mainline-qa-flow"
	spamLabel    = "spam"
	bookingLabel = "booking"
	qaLabel      = "qa"
)

// ownedFixture accepts only journal fixture names with the expected label.
func ownedFixture(name, label string) bool {
	return strings.HasPrefix(name, "msteams-sdk-") && strings.HasSuffix(name, "-"+label)
}

// ownedCleanup deletes one owned resource type. The read, list and delete
// functions abstract over spam templates and the two flow types; each returns
// the resource name so ownership is checked before anything is removed.
type ownedCleanup struct {
	label  string
	list   func(context.Context) (map[string]string, error)
	read   func(context.Context, string) (string, error)
	remove func(context.Context, string) (*resty.Response, error)
}

func (h ownedCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	if !ownedFixture(entry.Fixture, h.label) {
		return "", lab.ErrUnresolved
	}
	names, err := h.list(ctx)
	if err != nil {
		return "", err
	}
	id := ""
	for identity, name := range names {
		if name != entry.Fixture {
			continue
		}
		if id != "" || !labkit.DirectoryID.MatchString(identity) {
			return "", lab.ErrUnresolved
		}
		id = identity
	}
	if id == "" {
		// An empty list cannot prove a lost create never happened.
		return "", lab.ErrUnresolved
	}
	return id, nil
}

func (h ownedCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !ownedFixture(entry.Fixture, h.label) || !labkit.DirectoryID.MatchString(entry.ResourceID) {
		return lab.ErrUnresolved
	}
	name, err := h.read(ctx, entry.ResourceID)
	if labkit.IsHTTP(err, 404) {
		return nil
	}
	if err != nil {
		return err
	}
	if name != entry.Fixture {
		return lab.ErrUnresolved
	}
	if _, err := h.remove(ctx, entry.ResourceID); err != nil && !labkit.IsHTTP(err, 404) {
		return err
	}
	return h.absent(ctx, entry.ResourceID)
}

func (h ownedCleanup) absent(ctx context.Context, id string) error {
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		_, err := h.read(ctx, id)
		if labkit.IsHTTP(err, 404) {
			return true, nil
		}
		return false, err
	})
}

func (ownedCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }

func text(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func flowNames(values []*ma.Flow) map[string]string {
	names := map[string]string{}
	for _, value := range values {
		names[text(value.Identity)] = text(value.Name)
	}
	return names
}

func spamCleanup(service *ma.Service) ownedCleanup {
	return ownedCleanup{label: spamLabel,
		list: func(ctx context.Context) (map[string]string, error) {
			values, _, err := service.ListSpamDetectionTemplates(ctx, nil)
			names := map[string]string{}
			for _, value := range values {
				names[text(value.Identity)] = text(value.Name)
			}
			return names, err
		},
		read: func(ctx context.Context, id string) (string, error) {
			value, _, err := service.GetSpamDetectionTemplateByID(ctx, id)
			if err != nil {
				return "", err
			}
			return text(value.Name), nil
		},
		remove: service.DeleteSpamDetectionTemplateByID}
}

func bookingCleanup(service *ma.Service) ownedCleanup {
	return ownedCleanup{label: bookingLabel,
		list: func(ctx context.Context) (map[string]string, error) {
			values, _, err := service.ListAppointmentBookingFlows(ctx, nil)
			return flowNames(values), err
		},
		read: func(ctx context.Context, id string) (string, error) {
			value, _, err := service.GetAppointmentBookingFlowByID(ctx, id)
			if err != nil {
				return "", err
			}
			return text(value.Name), nil
		},
		remove: service.DeleteAppointmentBookingFlowByID}
}

func qaCleanup(service *ma.Service) ownedCleanup {
	return ownedCleanup{label: qaLabel,
		list: func(ctx context.Context) (map[string]string, error) {
			values, _, err := service.ListQuestionAnswerFlows(ctx, nil)
			return flowNames(values), err
		},
		read: func(ctx context.Context, id string) (string, error) {
			value, _, err := service.GetQuestionAnswerFlowByID(ctx, id)
			if err != nil {
				return "", err
			}
			return text(value.Name), nil
		},
		remove: service.DeleteQuestionAnswerFlowByID}
}
