package teamtemplate_test

import (
	"encoding/json"
	"os"
	"sync"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
)

type serviceCoverage struct {
	mu   sync.Mutex
	want map[contracttest.Mapping]bool
	seen map[contracttest.Mapping]bool
}

func collectServiceCoverage(t *testing.T, commands ...string) *serviceCoverage {
	t.Helper()
	raw, err := os.ReadFile("../../../contracts/powershell/8.1.0/runtime.json")
	if err != nil {
		t.Fatal(err)
	}
	var runtime struct {
		Commands []struct {
			Name string `json:"name"`
			Sets []struct {
				Name string `json:"name"`
			} `json:"parameter_sets"`
		} `json:"commands"`
	}
	if json.Unmarshal(raw, &runtime) != nil {
		t.Fatal("pinned runtime is unavailable")
	}
	report := &serviceCoverage{want: map[contracttest.Mapping]bool{}, seen: map[contracttest.Mapping]bool{}}
	for _, name := range commands {
		found := false
		for _, command := range runtime.Commands {
			if command.Name == name {
				found = true
				for _, set := range command.Sets {
					report.want[contracttest.Mapping{Command: name, ParameterSet: set.Name}] = true
				}
			}
		}
		if !found {
			t.Fatal("evidence command disappeared from pinned runtime")
		}
	}
	t.Cleanup(func() {
		// Cleanup runs after parallel subtests; this lock also makes ownership
		// explicit to reviewers and the race detector.
		report.mu.Lock()
		defer report.mu.Unlock()
		for mapping := range report.want {
			if !report.seen[mapping] {
				t.Error("one or more observed parameter sets lack passing SDK evidence")
			}
		}
		for mapping := range report.seen {
			if !report.want[mapping] {
				t.Error("SDK evidence contains an unobserved parameter-set mapping")
			}
		}

	})
	return report
}

func (e *serviceCoverage) add(c contracttest.Case) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, mapping := range c.Mappings {
		e.seen[mapping] = true
	}
}
