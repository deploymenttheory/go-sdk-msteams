package applicationaccesspolicy

import (
	"context"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

type patternToken struct {
	kind    uint16
	literal uint16
	ranges  [][2]uint16
}
type patternUnit struct {
	value   uint16
	escaped bool
}

func compilePattern(pattern string) ([]patternToken, error) {
	units := utf16.Encode([]rune(pattern))
	tokens := make([]patternToken, 0, len(units))
	for index := 0; index < len(units); index++ {
		value := units[index]
		switch value {
		case '*', '?':
			tokens = append(tokens, patternToken{kind: value})
		case '`':
			if index+1 < len(units) {
				index++
				value = units[index]
			}
			tokens = append(tokens, patternToken{literal: value})
		case '[':
			items := make([]patternUnit, 0)
			closed := false
			for index++; index < len(units); index++ {
				value = units[index]
				if value == ']' && len(items) > 0 {
					closed = true
					break
				}
				escaped := false
				if value == '`' && index+1 < len(units) {
					index++
					value = units[index]
					escaped = true
				}
				items = append(items, patternUnit{value: value, escaped: escaped})
			}
			if !closed {
				return nil, ErrInput
			}
			token := patternToken{kind: '['}
			for item := 0; item < len(items); item++ {
				start, end := items[item].value, items[item].value
				if item+2 < len(items) && items[item+1].value == '-' && !items[item+1].escaped {
					end = items[item+2].value
					item += 2
				}
				if start > end {
					return nil, ErrInput
				}
				token.ranges = append(token.ranges, [2]uint16{start, end})
			}
			tokens = append(tokens, token)
		default:
			tokens = append(tokens, patternToken{literal: value})
		}
	}
	return tokens, nil
}

func equalUnit(left, right uint16) bool {
	return left == right || unicode.ToUpper(rune(left)) == unicode.ToUpper(rune(right))
}

func matchesPattern(ctx context.Context, tokens []patternToken, value string) (bool, error) {
	units := utf16.Encode([]rune(value))
	previous := make([]bool, len(units)+1)
	previous[0] = true
	for _, token := range tokens {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		next := make([]bool, len(previous))
		if token.kind == '*' {
			next[0] = previous[0]
		}
		for position, unit := range units {
			switch token.kind {
			case '*':
				next[position+1] = previous[position+1] || next[position]
			case '?':
				next[position+1] = previous[position]
			case '[':
				for _, interval := range token.ranges {
					lower, upper := unicode.ToLower(rune(unit)), unicode.ToUpper(rune(unit))
					inside := unit >= interval[0] && unit <= interval[1]
					folded := lower >= rune(interval[0]) && lower <= rune(interval[1]) || upper >= rune(interval[0]) && upper <= rune(interval[1])
					if inside || folded {
						next[position+1] = previous[position]
						break
					}
				}
			default:
				next[position+1] = previous[position] && equalUnit(unit, token.literal)
			}
		}
		previous = next
	}
	return previous[len(units)], nil
}

// ListFiltered composes List with the pinned helper's local Identity/Tag:Identity
// wildcard filtering. It supports *, ?, bracket sets/ranges and backtick escapes;
// matching uses portable Unicode case folding and UTF-16 character positions.
// Locale-specific PowerShell collation is not inferred from the tenant locale.
func (s *Service) ListFiltered(ctx context.Context, pattern string) ([]Policy, *resty.Response, error) {
	if ctx == nil {
		return nil, nil, client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(pattern) == "" {
		return s.List(ctx)
	}
	plain, err := compilePattern(pattern)
	if err != nil {
		return nil, nil, err
	}
	tagged, err := compilePattern("Tag:" + pattern)
	if err != nil {
		return nil, nil, err
	}
	items, response, err := s.List(ctx)
	if err != nil {
		return nil, response, err
	}
	filtered := make([]Policy, 0)
	for _, item := range items {
		matches, err := matchesPattern(ctx, plain, item.Identity)
		if err != nil {
			return nil, response, err
		}
		if !matches {
			matches, err = matchesPattern(ctx, tagged, item.Identity)
			if err != nil {
				return nil, response, err
			}
		}
		if matches {
			filtered = append(filtered, item)
		}
	}
	return filtered, response, nil
}
