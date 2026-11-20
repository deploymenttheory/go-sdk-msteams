package lab_test

import (
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
)

func TestEnablementRequiresExplicitIndependentOptIns(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		env     map[string]string
		invalid bool
		allowed []lab.Scenario
	}{
		{name: "ordinary-tests-skip"},
		{name: "read-only", env: map[string]string{"MSTEAMS_LAB_READ_ONLY": "1"}, allowed: []lab.Scenario{lab.ReadOnly}},
		{name: "mutation-needs-read", env: map[string]string{"MSTEAMS_LAB_MUTATIONS": "1"}, invalid: true},
		{name: "mutation-needs-journal", env: map[string]string{"MSTEAMS_LAB_READ_ONLY": "1", "MSTEAMS_LAB_MUTATIONS": "1"}, invalid: true},
		{name: "consequential-needs-mutation", env: map[string]string{"MSTEAMS_LAB_READ_ONLY": "1", "MSTEAMS_LAB_CONSEQUENTIAL": "1"}, invalid: true},
		{name: "mutation", env: map[string]string{"MSTEAMS_LAB_READ_ONLY": "1", "MSTEAMS_LAB_MUTATIONS": "1", "MSTEAMS_LAB_JOURNAL_DIRECTORY": "private"}, allowed: []lab.Scenario{lab.ReadOnly, lab.Mutation}},
		{name: "consequential", env: map[string]string{"MSTEAMS_LAB_READ_ONLY": "1", "MSTEAMS_LAB_MUTATIONS": "1", "MSTEAMS_LAB_CONSEQUENTIAL": "1", "MSTEAMS_LAB_JOURNAL_DIRECTORY": "private"}, allowed: []lab.Scenario{lab.ReadOnly, lab.Mutation, lab.Consequential}},
		{name: "invalid-flag", env: map[string]string{"MSTEAMS_LAB_READ_ONLY": "true"}, invalid: true},
		{name: "unbounded-timeout", env: map[string]string{"MSTEAMS_LAB_TIMEOUT_SECONDS": "301"}, invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := lab.Parse(func(key string) string { return tc.env[key] })
			if tc.invalid {
				if !errors.Is(err, lab.ErrConfig) {
					t.Fatal("invalid configuration accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, scenario := range []lab.Scenario{lab.ReadOnly, lab.Mutation, lab.Consequential} {
				want := false
				for _, allowed := range tc.allowed {
					if allowed == scenario {
						want = true
					}
				}
				if (cfg.Require(scenario) == nil) != want {
					t.Fatal("scenario opt-in changed")
				}
			}
		})
	}
}
