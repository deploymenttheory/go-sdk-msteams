package externalaccesspolicy_test

import (
	"encoding/json"
	"testing"

	external "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/external_access_policy"
)

func TestUnitPolicyDecodesNamedCommunicationModeWithoutLosingWireValue(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, wire string
		want       external.CommunicationWithExternalOrgs
	}{
		{name: "named-response", wire: `"OrganizationDefault"`, want: external.OrganizationDefault},
		{name: "allow-all", wire: `"AllowAllExternalDomains"`, want: external.AllowAllExternalDomains},
		{name: "allow-specific", wire: `"AllowSpecificExternalDomains"`, want: external.AllowSpecificExternalDomains},
		{name: "block-specific", wire: `"BlockSpecificExternalDomains"`, want: external.BlockSpecificExternalDomains},
		{name: "case-insensitive", wire: `"blockallexternaldomains"`, want: external.BlockAllExternalDomains},
		{name: "future-numeric", wire: `7`, want: 7},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var value external.Policy
			raw := []byte(`{"Identity":"Global","CommunicationWithExternalOrgs":` + tc.wire + `}`)
			if err := json.Unmarshal(raw, &value); err != nil {
				t.Fatal("valid communication mode rejected")
			}
			if value.CommunicationWithExternalOrgs == nil || *value.CommunicationWithExternalOrgs != tc.want {
				t.Fatal("typed communication mode differs")
			}
			if string(value.RawFields["CommunicationWithExternalOrgs"]) != tc.wire {
				t.Fatal("original wire representation lost")
			}
		})
	}
	for _, raw := range []string{`true`, `{}`, `2147483648`, `1.5`, `"NotAKnownMode"`} {
		t.Run("invalid-"+raw, func(t *testing.T) {
			var value external.Policy
			if json.Unmarshal([]byte(`{"CommunicationWithExternalOrgs":`+raw+`}`), &value) == nil {
				t.Fatal("invalid communication mode accepted")
			}
		})
	}
}
