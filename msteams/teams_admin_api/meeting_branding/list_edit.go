package meetingbranding

import (
	"errors"
	"slices"
)

var ErrCannotRemoveElements = errors.New("branding policy object-list comparator cannot remove elements")

// AssetEdit is an owned local list edit. Nil *AssetEdit omits the property.
// Empty replacement/delta clears it, matching the pinned Core extension helper.
type AssetEdit[T Asset] struct {
	replace, add, remove *AssetList[T]
	delta, valid         bool
}

func ReplaceAssets[T Asset](values ...*T) (*AssetEdit[T], error) {
	list, err := Assets(values...)
	if err != nil {
		return nil, err
	}
	return &AssetEdit[T]{replace: list, valid: true}, nil
}
func ChangeAssets[T Asset](add, remove []*T) (*AssetEdit[T], error) {
	a, err := Assets(add...)
	if err != nil {
		return nil, err
	}
	r, err := Assets(remove...)
	if err != nil {
		return nil, err
	}
	return &AssetEdit[T]{add: a, remove: r, delta: true, valid: true}, nil
}
func (v *AssetEdit[T]) Validate() error {
	if v == nil || !v.valid {
		return ErrInput
	}
	return nil
}

// Apply never mutates current. Policy asset removals fail when the source's
// throwing comparator would run, even for a missing match. On an empty list
// no comparison runs. Reaction removals compare five fields, removing one match.
func (v *AssetEdit[T]) Apply(current []*T) (*AssetList[T], error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	if !v.delta {
		values, err := v.replace.values()
		if err != nil {
			return nil, err
		}
		return Assets(values...)
	}
	add, err := v.add.values()
	if err != nil {
		return nil, err
	}
	remove, err := v.remove.values()
	if err != nil {
		return nil, err
	}
	if len(add) == 0 && len(remove) == 0 {
		return Assets[T]()
	}
	values := slices.Clone(current)
	for _, item := range remove {
		if len(values) > 0 {
			if _, reaction := any(item).(*Reaction); !reaction {
				return nil, ErrCannotRemoveElements
			}
		}
		if item == nil {
			return nil, ErrInput // Missing-match warning calls ToString on null.
		}
		for i, candidate := range values {
			a, aOK := any(item).(*Reaction)
			b, bOK := any(candidate).(*Reaction)
			if !aOK || !bOK || a == nil || b == nil {
				return nil, ErrInput
			}
			if same(a.ID, b.ID) && same(a.ReactionImageURL, b.ReactionImageURL) && same(a.ReactionSpriteImageURL, b.ReactionSpriteImageURL) && same(a.ReactionType, b.ReactionType) && same(a.BrandedDisplayName, b.BrandedDisplayName) {
				values = slices.Delete(values, i, i+1)
				break
			}
		}
	}
	values = append(values, add...)
	return Assets(values...)
}

func same[T comparable](a, b *T) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
func (AssetEdit[T]) String() string     { return "branding asset edit (redacted)" }
func (v AssetEdit[T]) GoString() string { return v.String() }
