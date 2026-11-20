package capabilities

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

func TestCapabilityMatrixCoversPinnedInventory(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../contracts/powershell/8.1.0/reconciliation.json")
	if err != nil {
		t.Fatal(err)
	}
	var baseline struct {
		Version string `json:"module_version"`
		SHA     string `json:"package_sha256"`
		Records []struct {
			Name        string   `json:"name"`
			Disposition string   `json:"disposition"`
			Sets        []string `json:"parameter_set_names"`
		} `json:"records"`
	}
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatal(err)
	}
	catalog, err := NewCatalog()
	if err != nil {
		t.Fatal(err)
	}
	if catalog.matrix.BaselineVersion != baseline.Version || catalog.matrix.PackageSHA256 != baseline.SHA ||
		len(catalog.records) != len(baseline.Records) || len(catalog.matrix.Records) != len(catalog.records) {
		t.Fatal("matrix must exactly cover the pinned candidate union without duplicates")
	}
	for _, record := range baseline.Records {
		report, err := catalog.Assess(record.Name, auth.Delegated, "commercial")
		if err != nil || report.Disposition != record.Disposition || !reflect.DeepEqual(report.ParameterSets, record.Sets) {
			t.Fatalf("pinned metadata drift for %s: %v", record.Name, err)
		}
		if report.Support != ModeUnqualified {
			t.Fatalf("discovery claimed authorization: %s", record.Name)
		}
	}
}

func TestEveryDocumentedRestrictionAndCloud(t *testing.T) {
	t.Parallel()
	catalog, err := NewCatalog()
	if err != nil {
		t.Fatal(err)
	}
	// An independent list protects the complete dated source overlay, including
	// wildcard suffixes and sovereign-only exceptions, from accidental removal.
	patterns := []string{"New-Team", "Set-CsOnlineApplicationInstance", "*PolicyPackage*",
		"*-CsTeamsShiftsConnection*", "*-CsBatchTeamsDeployment*", "Get-MultiGeoRegion",
		"Get-AllM365TeamsApps", "Get-M365TeamsApp", "Update-M365TeamsApp",
		"Get-M365UnifiedCustomPendingApps", "Get-M365UnifiedTenantSettings",
		"Update-M365UnifiedTenantSettings", "Update-M365UnifiedCustomPendingApp"}
	sovereign := []string{"Get-CsOnlineApplicationInstance", "New-CsOnlineApplicationInstance", "Sync-CsOnlineApplicationInstance"}
	if len(catalog.rules) != len(patterns)+len(sovereign) {
		t.Fatal("restriction inventory drift")
	}
	matched := map[string]bool{}
	for _, record := range catalog.matrix.Records {
		for _, cloud := range []string{"commercial", "gcc", "gcch", "dod"} {
			expected := false
			for _, pattern := range patterns {
				match, err := path.Match(pattern, record.Name)
				if err != nil {
					t.Fatal(err)
				}
				if match {
					expected = true
					matched[pattern] = true
				}
			}
			if (cloud == "gcch" || cloud == "dod") && slices.Contains(sovereign, record.Name) {
				expected = true
				matched[record.Name] = true
			}
			for _, mode := range []auth.Mode{auth.Delegated, auth.Application} {
				report, err := catalog.Assess(strings.ToLower(record.Name), mode, cloud)
				if err != nil {
					t.Fatal(err)
				}
				if (report.Support == ModeDocumentedUnsupported) != (expected && mode == auth.Application) {
					t.Fatalf("wrong restriction: %s/%s/%s", record.Name, mode, cloud)
				}
				if report.Support == ModeDocumentedUnsupported && len(report.RestrictionIDs) == 0 {
					t.Fatal("restriction has no evidence mapping")
				}
			}
		}
	}
	for _, pattern := range append(patterns, sovereign...) {
		if !matched[pattern] {
			t.Fatalf("untested source restriction %s", pattern)
		}
	}
}

func TestCapabilityReportingDoesNotInferAuthorization(t *testing.T) {
	t.Parallel()
	catalog, err := NewCatalog()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
		binding := auth.Binding{TenantID: "fixture", Cloud: "commercial", IdentityID: "fixture", Mode: mode}
		provider, err := auth.NewSuppliedTokenProvider(binding, map[string]auth.Token{
			"graph": {AccessToken: "fixture-graph", ExpiresAt: time.Now().Add(time.Hour)},
			"teams": {AccessToken: "fixture-teams", ExpiresAt: time.Now().Add(time.Hour)},
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(provider.Close)
		for _, audience := range []string{"graph", "teams"} {
			if _, err := provider.AcquireToken(context.Background(), binding.Request(audience)); err != nil {
				t.Fatal(err)
			}
			request := binding.Request(audience)
			if mode == auth.Application {
				request.Mode = auth.Delegated
			} else {
				request.Mode = auth.Application
			}
			if _, err := provider.AcquireToken(context.Background(), request); !errors.Is(err, auth.ErrBindingMismatch) {
				t.Fatal("same-type dual-token binding was bypassed")
			}
		}
		report, err := catalog.Assess("Get-Team", mode, "commercial")
		if err != nil || report.Support != ModeUnqualified {
			t.Fatal("tokens inferred operation authorization")
		}
	}
	first, err := catalog.Assess("New-Team", auth.Application, "commercial")
	if err != nil {
		t.Fatal(err)
	}
	first.RestrictionIDs[0] = "modified"
	first.ParameterSets[0] = "modified"
	second, err := catalog.Assess("New-Team", auth.Application, "commercial")
	if err != nil || second.RestrictionIDs[0] == "modified" || second.ParameterSets[0] == "modified" {
		t.Fatal("caller changed catalog evidence")
	}
	if _, err := catalog.Assess("Invented-Command", auth.Application, "commercial"); !errors.Is(err, ErrUnknownCommand) {
		t.Fatal("unknown command gained a capability report")
	}
	if _, err := catalog.Assess("Get-Team", "mixed", "commercial"); !errors.Is(err, ErrInvalidMode) {
		t.Fatal("mixed mode accepted")
	}
	if _, err := catalog.Assess("Get-Team", auth.Application, "unknown"); !errors.Is(err, ErrUnknownProfile) {
		t.Fatal("unknown cloud accepted")
	}
}
