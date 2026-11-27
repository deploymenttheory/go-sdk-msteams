package teamsapppolicyconfiguration

import (
	"regexp"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

var uriPattern = regexp.MustCompile(`^(h|H)(t|T)(t|T)(p|P)(s|S)://.+$`)

// The pinned CultureInvariant .NET expression is a prefix/content check, not
// URL parsing. Its $ anchor also admits a single terminal LF; preserve that
// exact scalar behavior without trimming other whitespace or interior lines.
func validURI(value string) bool { return uriPattern.MatchString(strings.TrimSuffix(value, "\n")) }

func validate(value Settings) error {
	for _, field := range []policy.Field[string]{value.AppCatalogURI, value.ResourceURI} {
		if field.IsSet() && !field.IsNull() {
			v, _ := field.Value()
			if !validURI(v) {
				return ErrInput
			}
		}
	}
	return nil
}
