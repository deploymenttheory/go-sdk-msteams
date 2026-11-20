// Package capabilities reports pinned command metadata and dated operation-mode
// restrictions. It performs no authentication or tenant capability discovery.
package capabilities

import (
	_ "embed"
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

//go:embed capabilities_8_1_0.json
var capabilityData []byte

var (
	ErrUnknownCommand = errors.New("command is absent from the pinned capability inventory")
	ErrInvalidMode    = errors.New("capability assessment requires a single authentication mode")
	ErrUnknownProfile = errors.New("unknown capability documentation profile")
)

// ModeSupport reports evidence, not operation authorization or implementation.
type ModeSupport string

const (
	ModeUnqualified           ModeSupport = "unqualified"
	ModeDocumentedUnsupported ModeSupport = "documented-unsupported"
)

// OperationCapability keeps pinned discovery separate from current documentation.
// Unqualified means unknown, including when token acquisition has succeeded.
// An unsupported documentation observation is advisory, not a runtime probe.
type OperationCapability struct {
	Command               string
	BaselineVersion       string
	Disposition           string
	ParameterSets         []string
	Mode                  auth.Mode
	Cloud                 string
	Support               ModeSupport
	RestrictionIDs        []string
	DocumentationURL      string
	DocumentationReviewed string
}

type capabilityRecord struct {
	Name          string   `json:"name"`
	Disposition   string   `json:"disposition"`
	ParameterSets []string `json:"parameter_sets"`
	Restrictions  []string `json:"current_restrictions"`
}

type capabilityRule struct {
	ID      string   `json:"id"`
	Pattern string   `json:"pattern"`
	Clouds  []string `json:"clouds"`
}

type capabilityMatrix struct {
	BaselineVersion       string             `json:"baseline_version"`
	PackageSHA256         string             `json:"package_sha256"`
	DocumentationURL      string             `json:"documentation_url"`
	DocumentationReviewed string             `json:"documentation_reviewed"`
	Rules                 []capabilityRule   `json:"rules"`
	Records               []capabilityRecord `json:"records"`
}

// Catalog is an immutable, offline snapshot of the 8.1.0 candidate
// inventory and separately dated documentation restrictions. It grants no access,
// discovers no tenant features, and does not enable sovereign-cloud routing.
type Catalog struct {
	matrix  capabilityMatrix
	records map[string]capabilityRecord
	rules   map[string]capabilityRule
}

// NewCatalog loads embedded public evidence without credentials or I/O.
func NewCatalog() (*Catalog, error) {
	c := &Catalog{}
	if err := json.Unmarshal(capabilityData, &c.matrix); err != nil {
		return nil, errors.New("invalid embedded capability evidence")
	}
	c.records = make(map[string]capabilityRecord, len(c.matrix.Records))
	for _, record := range c.matrix.Records {
		c.records[strings.ToLower(record.Name)] = record
	}
	c.rules = make(map[string]capabilityRule, len(c.matrix.Rules))
	for _, rule := range c.matrix.Rules {
		c.rules[rule.ID] = rule
	}
	return c, nil
}

// Assess returns copied evidence for one command/mode/cloud. Cloud names here are
// documentation labels: commercial, gcc, gcch and dod. An absent restriction never
// means support. Application restrictions also apply to managed/workload identities
// and supplied tokens whose caller-declared mode is auth.Application.
func (c *Catalog) Assess(command string, mode auth.Mode, cloud string) (OperationCapability, error) {
	if c == nil || c.records == nil {
		return OperationCapability{}, ErrUnknownCommand
	}
	if mode != auth.Delegated && mode != auth.Application {
		return OperationCapability{}, ErrInvalidMode
	}
	if !slices.Contains([]string{"commercial", "gcc", "gcch", "dod"}, cloud) {
		return OperationCapability{}, ErrUnknownProfile
	}
	record, ok := c.records[strings.ToLower(command)]
	if !ok {
		return OperationCapability{}, ErrUnknownCommand
	}
	report := OperationCapability{
		Command: record.Name, BaselineVersion: c.matrix.BaselineVersion,
		Disposition: record.Disposition, ParameterSets: slices.Clone(record.ParameterSets),
		Mode: mode, Cloud: cloud, Support: ModeUnqualified,
		DocumentationURL:      c.matrix.DocumentationURL,
		DocumentationReviewed: c.matrix.DocumentationReviewed,
	}
	if mode == auth.Application {
		for _, id := range record.Restrictions {
			if slices.Contains(c.rules[id].Clouds, cloud) {
				report.Support = ModeDocumentedUnsupported
				report.RestrictionIDs = append(report.RestrictionIDs, id)
			}
		}
	}
	return report, nil
}
