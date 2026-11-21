// Package policyfilter implements the shared local configuration-policy wildcard grammar.
package policyfilter

import (
	"context"
	"errors"
	"unicode"
	"unicode/utf16"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// ErrPattern identifies an invalid local wildcard expression.
var ErrPattern = errors.New("invalid policy wildcard pattern")

// Pattern is an immutable, compiled UTF-16 wildcard matcher. Its zero value
// matches only the empty string. Match does not apply locale-specific collation.
type Pattern struct{ tokens []patternToken }

type patternToken struct {
	kind    uint16
	literal uint16
	ranges  [][2]uint16
}
type patternUnit struct {
	value   uint16
	escaped bool
}

// Compile parses the pinned helper's UTF-16 wildcard grammar without I/O.
func Compile(pattern string) (Pattern, error) {
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
				return Pattern{}, ErrPattern
			}
			token := patternToken{kind: '['}
			for item := 0; item < len(items); item++ {
				start, end := items[item].value, items[item].value
				if item+2 < len(items) && items[item+1].value == '-' && !items[item+1].escaped {
					end = items[item+2].value
					item += 2
				}
				if start > end {
					return Pattern{}, ErrPattern
				}
				token.ranges = append(token.ranges, [2]uint16{start, end})
			}
			tokens = append(tokens, token)
		default:
			tokens = append(tokens, patternToken{literal: value})
		}
	}
	return Pattern{tokens: tokens}, nil
}

func equalUnit(left, right uint16) bool {
	return left == right || unicode.ToUpper(rune(left)) == unicode.ToUpper(rune(right))
}

// Match applies a compiled pattern with caller cancellation.
func (p Pattern) Match(ctx context.Context, value string) (bool, error) {
	if ctx == nil {
		return false, client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	units := utf16.Encode([]rune(value))
	previous := make([]bool, len(units)+1)
	previous[0] = true
	for _, token := range p.tokens {
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
