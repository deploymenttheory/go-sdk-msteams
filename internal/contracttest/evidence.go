package contracttest

import (
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"slices"
)

var (
	ErrEvidence = errors.New("contracttest: evidence encoding failed")
	revisionID  = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)
)

type caseEvidence struct {
	Name          string       `json:"name"`
	Family        string       `json:"family"`
	Scope         string       `json:"scope"`
	Cloud         string       `json:"cloud"`
	Mode          string       `json:"mode"`
	Source        string       `json:"source"`
	Mappings      []Mapping    `json:"mappings"`
	Outcome       string       `json:"outcome"`
	HTTPAttempts  int          `json:"http_attempts"`
	TokenRequests int          `json:"token_requests"`
	Differences   []Difference `json:"differences"`
}

// WriteEvidence emits a deterministic candidate run artifact suitable for review
// alongside ledger symbol/parameter-set mappings. It never updates the ledger or
// promotes command coverage. Primitive/local checks cannot establish service
// parity; every artifact explicitly leaves live qualification as not-run.
// Names, mappings and Source must be curated synthetic metadata, not tenant data.
func WriteEvidence(writer io.Writer, revision string, results []Result) error {
	if writer == nil || !revisionID.MatchString(revision) || len(results) == 0 {
		return ErrEvidence
	}
	ordered := slices.Clone(results)
	slices.SortFunc(ordered, func(a, b Result) int {
		if a.family != b.family {
			return compareString(a.family, b.family)
		}
		return compareString(a.name, b.name)
	})
	entries := make([]caseEvidence, 0, len(ordered))
	outcome := "passed"
	previous := ""
	for _, result := range ordered {
		key := result.family + "/" + result.name
		if result.name == "" || key == previous {
			return ErrEvidence
		}
		previous = key
		state := "passed"
		if !result.Passed() {
			state, outcome = "failed", "failed"
		}
		mappings := append([]Mapping{}, result.mappings...)
		slices.SortFunc(mappings, func(a, b Mapping) int {
			if a.Command != b.Command {
				return compareString(a.Command, b.Command)
			}
			return compareString(a.ParameterSet, b.ParameterSet)
		})
		entries = append(entries, caseEvidence{Name: result.name, Family: result.family,
			Scope: result.scope, Cloud: result.cloud, Mode: result.mode, Source: result.source, Mappings: mappings, Outcome: state,
			HTTPAttempts: result.httpCalls, TokenRequests: result.tokenCalls,
			Differences: append([]Difference{}, result.differences...)})
	}
	artifact := struct {
		SchemaVersion   int            `json:"schema_version"`
		BaselineVersion string         `json:"baseline_version"`
		Revision        string         `json:"revision"`
		Verification    string         `json:"verification"`
		Outcome         string         `json:"outcome"`
		CoverageClaim   bool           `json:"coverage_claim"`
		Live            string         `json:"live"`
		Cases           []caseEvidence `json:"cases"`
	}{SchemaVersion: 1, BaselineVersion: "8.1.0", Revision: revision,
		Verification: "offline-contract", Outcome: outcome, CoverageClaim: false,
		Live: "not-run", Cases: entries}
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(artifact); err != nil {
		return ErrEvidence
	}
	return nil
}

func compareString(a, b string) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
