package recordingmessages

import (
	"encoding/json"
)

// LocalizationInput mirrors the local builder, without a locale whitelist or
// invented text limits. Empty required Language fails; whitespace is preserved.
type LocalizationInput struct {
	Language                     string
	InitiatorImplicit            *string
	ParticipantImplicit          *string
	InitiatorExplicit            *string
	ParticipantExplicitRequested *string
	ParticipantExplicitProvided  *string
	AgreementDialogue            *string
}

// NewLocalization is a pure local builder; it performs no token or HTTP work.
func NewLocalization(input LocalizationInput) (*Localization, error) {
	if input.Language == "" {
		return nil, ErrInput
	}
	return cloneLocalization(&Localization{Language: &input.Language, InitiatorImplicit: input.InitiatorImplicit, ParticipantImplicit: input.ParticipantImplicit, InitiatorExplicit: input.InitiatorExplicit, ParticipantExplicitRequested: input.ParticipantExplicitRequested, ParticipantExplicitProvided: input.ParticipantExplicitProvided, AgreementDialogue: input.AgreementDialogue}), nil
}
func cloneString(v *string) *string {
	if v == nil {
		return nil
	}
	return new(*v)
}
func cloneLocalization(v *Localization) *Localization {
	if v == nil {
		return nil
	}
	x := *v
	x.Identity = cloneString(v.Identity)
	x.Language = cloneString(v.Language)
	x.InitiatorImplicit = cloneString(v.InitiatorImplicit)
	x.ParticipantImplicit = cloneString(v.ParticipantImplicit)
	x.InitiatorExplicit = cloneString(v.InitiatorExplicit)
	x.ParticipantExplicitRequested = cloneString(v.ParticipantExplicitRequested)
	x.ParticipantExplicitProvided = cloneString(v.ParticipantExplicitProvided)
	x.AgreementDialogue = cloneString(v.AgreementDialogue)
	if v.RawFields != nil {
		x.RawFields = make(map[string]json.RawMessage, len(v.RawFields))
		for k, raw := range v.RawFields {
			x.RawFields[k] = append(json.RawMessage(nil), raw...)
		}
	}
	return &x
}
func cloneLocalizations(v []*Localization) []*Localization {
	result := make([]*Localization, len(v))
	for i, item := range v {
		result[i] = cloneLocalization(item)
	}
	return result
}

// LocalizationEdit is a local replacement or ordered remove/add operation.
// Nil on a request means unbound. ClearLocalizations represents a bound null or
// empty modifier, both of which clear the source collection.
type LocalizationEdit struct {
	replacement, add, remove []*Localization
	delta, valid             bool
}

func ReplaceLocalizations(v ...*Localization) *LocalizationEdit {
	return &LocalizationEdit{replacement: cloneLocalizations(v), valid: true}
}
func ClearLocalizations() *LocalizationEdit { return ReplaceLocalizations() }
func ChangeLocalizations(add, remove []*Localization) *LocalizationEdit {
	return &LocalizationEdit{add: cloneLocalizations(add), remove: cloneLocalizations(remove), delta: true, valid: true}
}
func (e *LocalizationEdit) Validate() error {
	if e == nil || !e.valid {
		return ErrInput
	}
	return nil
}
func equalString(a, b *string) bool { return a == nil && b == nil || a != nil && b != nil && *a == *b }
func equalLocalization(a, b *Localization) (bool, error) {
	if a == nil || b == nil {
		return false, ErrInput
	}
	return equalString(a.Language, b.Language) && equalString(a.InitiatorImplicit, b.InitiatorImplicit) && equalString(a.ParticipantImplicit, b.ParticipantImplicit) && equalString(a.InitiatorExplicit, b.InitiatorExplicit) && equalString(a.ParticipantExplicitRequested, b.ParticipantExplicitRequested) && equalString(a.ParticipantExplicitProvided, b.ParticipantExplicitProvided) && equalString(a.AgreementDialogue, b.AgreementDialogue), nil
}

// Apply copies the source list, removes the first seven-field ordinal match per
// Remove entry, then appends Add without deduplication. Identity is not a key.
func (e *LocalizationEdit) Apply(current []*Localization) ([]*Localization, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	if !e.delta {
		return cloneLocalizations(e.replacement), nil
	}
	if len(e.add) == 0 && len(e.remove) == 0 {
		return []*Localization{}, nil
	}
	values := cloneLocalizations(current)
	for _, remove := range e.remove {
		if remove == nil {
			return nil, ErrInput
		}
		for i, item := range values {
			match, err := equalLocalization(remove, item)
			if err != nil {
				return nil, err
			}
			if match {
				values = append(values[:i], values[i+1:]...)
				break
			}
		}
	}
	return append(values, cloneLocalizations(e.add)...), nil
}
func (LocalizationEdit) String() string     { return "recording localization edit (redacted)" }
func (v LocalizationEdit) GoString() string { return v.String() }

// localizationDTO excludes future received fields, just as the pinned model
// serializer does, while serializing every known nullable child property.
func localizationDTO(v []*Localization) []map[string]*string {
	result := make([]map[string]*string, len(v))
	for i, x := range v {
		if x != nil {
			result[i] = map[string]*string{"Identity": x.Identity, "Language": x.Language, "InitiatorImplicit": x.InitiatorImplicit, "ParticipantImplicit": x.ParticipantImplicit, "InitiatorExplicit": x.InitiatorExplicit, "ParticipantExplicitRequested": x.ParticipantExplicitRequested, "ParticipantExplicitProvided": x.ParticipantExplicitProvided, "AgreementDialogue": x.AgreementDialogue}
		}
	}
	return result
}
