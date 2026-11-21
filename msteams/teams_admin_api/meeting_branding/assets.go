package meetingbranding

import (
	"encoding/json"
	"regexp"
	"slices"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

var ErrInput = policy.ErrInput

// The pinned .NET regex uses ASCII case alternatives and accepts one final LF.
// It is not a URL parser, a length limit or a permission to dereference the URI.
var httpsPattern = regexp.MustCompile(`^[hH][tT][tT][pP][sS]://[^\n]+\n?$`)

func copyPointer[T any](v *T) *T {
	if v == nil {
		return nil
	}
	value := *v
	return &value
}

type BackgroundImageOptions struct{ IsRequired, IsHidden *bool }
type AssuranceSlateOptions struct{ IsDefault *bool }
type ReactionOptions struct{ SpriteImageURL, BrandedDisplayName *string }
type ThemeOptions struct {
	ThemeFields
	MeetingReactions *AssetEdit[Reaction]
}

// NewBackgroundImage constructs a local value, with no HTTP, upload or URI read.
// Required strings follow ValidateNotNullOrEmpty; Order has no inferred range.
func NewBackgroundImage(id string, order int64, name, uri string, options BackgroundImageOptions) (*BackgroundImage, error) {
	if id == "" || name == "" || !httpsPattern.MatchString(uri) {
		return nil, ErrInput
	}
	return &BackgroundImage{ID: &id, Order: &order, Name: &name, ImageURI: &uri, IsRequired: copyPointer(options.IsRequired), IsHidden: copyPointer(options.IsHidden)}, nil
}

// NewAssuranceSlate constructs a local NDI slate; the policy owns persistence.
func NewAssuranceSlate(id, name, uri string, options AssuranceSlateOptions) (*AssuranceSlate, error) {
	if id == "" || name == "" || !httpsPattern.MatchString(uri) {
		return nil, ErrInput
	}
	return &AssuranceSlate{ID: &id, Name: &name, NDIImageURI: &uri, IsDefault: copyPointer(options.IsDefault)}, nil
}

// NewReaction is a local helper for the theme's typed reaction list. The pinned
// model allows empty ID/type strings, but requires a non-null HTTPS image value.
func NewReaction(id, uri, kind string, options ReactionOptions) (*Reaction, error) {
	if !httpsPattern.MatchString(uri) || options.SpriteImageURL != nil && !httpsPattern.MatchString(*options.SpriteImageURL) {
		return nil, ErrInput
	}
	return &Reaction{ID: &id, ReactionImageURL: &uri, ReactionType: &kind, ReactionSpriteImageURL: copyPointer(options.SpriteImageURL), BrandedDisplayName: copyPointer(options.BrandedDisplayName)}, nil
}

// NewBrandingTheme constructs a local theme. A nil reaction edit leaves the
// initialized empty list; explicit edits use the pinned five-field comparison.
func NewBrandingTheme(identity string, options ThemeOptions) (*BrandingTheme, error) {
	if identity == "" {
		return nil, ErrInput
	}
	raw, err := json.Marshal(options.ThemeFields)
	if err != nil {
		return nil, ErrInput
	}
	var fields ThemeFields
	if json.Unmarshal(raw, &fields) != nil {
		return nil, ErrInput
	}
	value := &BrandingTheme{Identity: &identity, ThemeFields: fields, MeetingReactions: []*Reaction{}}
	if options.MeetingReactions != nil {
		list, err := options.MeetingReactions.Apply(nil)
		if err != nil {
			return nil, err
		}
		value.MeetingReactions, err = list.values()
		if err != nil {
			return nil, err
		}
		for _, r := range value.MeetingReactions {
			if r != nil {
				r.RawFields = nil
			}
		}
	}
	return value, nil
}

// Asset is the closed set of nested model values used by this resource.
type Asset interface {
	BackgroundImage | AssuranceSlate | BrandingTheme | Reaction
}

// AssetList owns a full projected JSON array. Nil *AssetList omits a setting;
// Assets[T]() sends []. The zero value is invalid. No server list operator is sent.
type AssetList[T Asset] struct {
	items []json.RawMessage
	valid bool
}

// Assets copies the pinned public-property projection, discarding unknown raw
// snapshot fields. Nil elements remain JSON null, as in the provider serializer.
func Assets[T Asset](values ...*T) (*AssetList[T], error) {
	result := &AssetList[T]{items: []json.RawMessage{}, valid: true}
	for _, value := range values {
		raw, err := assetJSON(value)
		if err != nil {
			return nil, err
		}
		result.items = append(result.items, raw)
	}
	return result, nil
}

func (v AssetList[T]) MarshalJSON() ([]byte, error) {
	if !v.valid {
		return nil, ErrInput
	}
	return json.Marshal(v.items)
}
func (AssetList[T]) String() string     { return "branding asset list (redacted)" }
func (v AssetList[T]) GoString() string { return v.String() }
func (v *AssetList[T]) values() ([]*T, error) {
	if v == nil || !v.valid {
		return nil, ErrInput
	}
	result := make([]*T, 0, len(v.items))
	for _, raw := range v.items {
		value, err := decodeProjection[T](raw)
		if err != nil {
			return nil, ErrInput
		}
		result = append(result, value)
	}
	return result, nil
}

func assetJSON[T Asset](value *T) (json.RawMessage, error) {
	if value == nil {
		return json.RawMessage("null"), nil
	}
	var raw []byte
	var err error
	switch v := any(value).(type) {
	case *BackgroundImage:
		if v.ImageURI != nil && !httpsPattern.MatchString(*v.ImageURI) {
			return nil, ErrInput
		}
		type projection BackgroundImage
		copy := projection(*v)
		if copy.Order == nil {
			var zero int64
			copy.Order = &zero // CLR public getter default, only in a write projection.
		}
		raw, err = json.Marshal(copy)
	case *AssuranceSlate:
		if v.NDIImageURI != nil && !httpsPattern.MatchString(*v.NDIImageURI) {
			return nil, ErrInput
		}
		type projection AssuranceSlate
		raw, err = json.Marshal(projection(*v))
	case *Reaction:
		if v.ReactionImageURL != nil && !httpsPattern.MatchString(*v.ReactionImageURL) || v.ReactionSpriteImageURL != nil && !httpsPattern.MatchString(*v.ReactionSpriteImageURL) {
			return nil, ErrInput
		}
		type projection Reaction
		raw, err = json.Marshal(projection(*v))
	case *BrandingTheme:
		reactions, e := Assets(v.MeetingReactions...)
		if e != nil {
			return nil, e
		}
		type projection BrandingTheme
		raw, err = json.Marshal(struct {
			projection
			Reactions *AssetList[Reaction] `json:"MeetingReactions"`
		}{projection(*v), reactions})
	}
	if err != nil {
		return nil, ErrInput
	}
	return slices.Clone(raw), nil
}
