package applicationaccesspolicy

import (
	"errors"
	"testing"
)

func TestUnitWildcardSemanticsMatchPowerShellStringCases(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, pattern, value string
		want, invalid        bool
	}{
		{name: "case insensitive", pattern: "fIX*", value: "Fixture", want: true},
		{name: "single unit", pattern: "a?b", value: "aXb", want: true},
		{name: "emoji one unit rejected", pattern: "a?b", value: "a😀b"},
		{name: "emoji two units", pattern: "a??b", value: "a😀b", want: true},
		{name: "literal star", pattern: "a`*b", value: "a*b", want: true},
		{name: "literal trailing backtick", pattern: "abc`", value: "abc`", want: true},
		{name: "set", pattern: "[ab]", value: "A", want: true},
		{name: "range", pattern: "[a-z]", value: "Q", want: true},
		{name: "exclamation not negation", pattern: "[!a]", value: "!", want: true},
		{name: "exclamation not negation miss", pattern: "[!a]", value: "b"},
		{name: "literal close bracket", pattern: "[]]", value: "]", want: true},
		{name: "literal open bracket", pattern: "[[]", value: "[", want: true},
		{name: "unclosed", pattern: "[", value: "[", invalid: true},
		{name: "empty class", pattern: "[]", value: "]", invalid: true},
		{name: "reversed range", pattern: "[z-a]", value: "a", invalid: true},
		{name: "newline", pattern: "*", value: "a\nb", want: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			tokens, err := compilePattern(test.pattern)
			if test.invalid {
				if !errors.Is(err, ErrInput) {
					t.Fatal("invalid wildcard accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			actual, err := tokens.Match(t.Context(), test.value)
			if err != nil {
				t.Fatal(err)
			}
			if actual != test.want {
				t.Fatalf("match=%v, want %v", actual, test.want)
			}
		})
	}
}
