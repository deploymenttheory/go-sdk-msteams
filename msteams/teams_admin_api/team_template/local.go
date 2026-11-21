// Package teamtemplate manages Teams template resources and their local values.
package teamtemplate

import "errors"

// ErrInput identifies invalid local template arguments before any request.
var ErrInput = errors.New("team template input is invalid")

// HiddenTemplate is a local ID value for a template permission policy. Creating
// it does not hide a template or persist a policy. Policy serialization belongs
// to the permission-policy service, not this local constructor.
type HiddenTemplate struct{ ID string }

// NewHiddenTemplate mirrors the local New-CsTeamsHiddenTemplate builder's
// mandatory nonempty ID. It needs no authentication or network access.
func NewHiddenTemplate(id string) (HiddenTemplate, error) {
	if id == "" {
		return HiddenTemplate{}, ErrInput
	}
	return HiddenTemplate{ID: id}, nil
}

// HiddenMeetingTemplate is the corresponding local ID value for a meeting
// template permission policy. It does not perform a management operation.
type HiddenMeetingTemplate struct{ ID string }

// NewHiddenMeetingTemplate mirrors New-CsTeamsHiddenMeetingTemplate. Validation
// is limited to the evidenced nonempty ID; it does not invent a server lookup.
func NewHiddenMeetingTemplate(id string) (HiddenMeetingTemplate, error) {
	if id == "" {
		return HiddenMeetingTemplate{}, ErrInput
	}
	return HiddenMeetingTemplate{ID: id}, nil
}
