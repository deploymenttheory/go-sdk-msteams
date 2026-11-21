package teamtemplate_test

import (
	"bytes"
	"encoding/json"
	"os"
	"sync"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
)

type serviceEvidence struct {
	mu      sync.Mutex
	results []contracttest.Result
	want    map[contracttest.Mapping]bool
	seen    map[contracttest.Mapping]bool
}

func collectServiceEvidence(t *testing.T, commands ...string) *serviceEvidence {
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
	report := &serviceEvidence{results: []contracttest.Result{}, want: map[contracttest.Mapping]bool{}, seen: map[contracttest.Mapping]bool{}}
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
		revision := os.Getenv("MSTEAMS_CONTRACT_REVISION")
		if revision == "" || t.Failed() {
			return
		}
		var encoded bytes.Buffer
		if err := contracttest.WriteEvidence(&encoded, revision, report.results); err != nil {
			t.Error("service contract evidence could not be encoded")
			return
		}
		root, err := os.OpenRoot(t.ArtifactDir())
		if err != nil {
			t.Error("service contract artifact directory unavailable")
			return
		}
		defer func() { _ = root.Close() }()
		file, err := root.Create("team-template-service.json")
		if err != nil {
			t.Error("service contract artifact file unavailable")
			return
		}
		_, writeErr := file.Write(encoded.Bytes())
		closeErr := file.Close()
		if writeErr != nil || closeErr != nil {
			t.Error("service contract artifact could not be written")
		}
	})
	return report
}

func (e *serviceEvidence) add(c contracttest.Case, result contracttest.Result) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.results = append(e.results, result)
	for _, mapping := range c.Mappings {
		e.seen[mapping] = true
	}
}
