package labkit

import (
	"strings"
	"testing"
)

func TestUnitAccessOutcomeRejectsUnverifiedOrDifferentResults(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	good := `{"success":true,"tenant_id":"` + tenant + `","command":"Get-CsPrivacyConfiguration","outcome":"forbidden","code":"Forbidden","error_code":40301}`
	for _, test := range []struct {
		name, raw string
		valid     bool
	}{
		{"verified", good, true},
		{"failed-setup", strings.Replace(good, `"success":true`, `"success":false`, 1), false},
		{"wrong-tenant", strings.Replace(good, tenant, "00000000-0000-4000-8000-000000000004", 1), false},
		{"wrong-code", strings.Replace(good, "40301", "40003", 1), false},
		{"wrong-outcome", strings.Replace(good, "forbidden", "success", 1), false},
		{"trailing-data", good + ` {}`, false},
		{"unknown-property", strings.Replace(good, `"success":true`, `"extra":true,"success":true`, 1), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := DecodeAccessOutcome([]byte(test.raw), tenant, "Get-CsPrivacyConfiguration", "forbidden")
			if (err == nil) != test.valid {
				t.Fatal("unverified access outcome accepted or verified outcome rejected")
			}
		})
	}
}
