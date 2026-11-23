package labkit

import "testing"

func TestUnitFederationOracleRejectsMixedAndIncompleteOutput(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"success":false}`, `{"success":true,"kind":"acs","identity":"Global","settings":{}}`, `{"success":true,"kind":"acs","identity":"Global","settings":{"EnableAcsUsers":false,"RequireAcsFederationForMeeting":null,"extra":false}}`, `{"success":true,"kind":"tenant","identity":"Global","settings":{"EnableAcsUsers":false,"RequireAcsFederationForMeeting":null}}`} {
		if _, err := decodeFederationOracle([]byte(raw), "acs"); err == nil {
			t.Fatal("invalid oracle projection accepted")
		}
	}
	if _, err := decodeFederationOracle([]byte(`{"success":true,"kind":"acs","identity":"Global","settings":{"EnableAcsUsers":false,"RequireAcsFederationForMeeting":null}}`), "acs"); err != nil {
		t.Fatal("valid selected nullable projection rejected")
	}
}
