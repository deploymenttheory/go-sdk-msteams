package labkit

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestUnitResponseFixturePreservesJSONTypesAndRedactsPrivateValues(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"Identity":"Global","Description":null,"XmlRoot":{"name":"TeamsFilesPolicy"},"Values":[false,null,9007199254740993,1.00],"TurnItInApiKey":"private-key-value","ContentPin":"123456","Contact":"operator@tenant.example","Uri":"https://private.example/file?sig=private-value"}`)
	got, err := fixtureJSON(raw, nil)
	if err != nil || !bytes.Contains(got, []byte("9007199254740993")) || !bytes.Contains(got, []byte("1.00")) {
		t.Fatal("response numeric values changed")
	}
	for _, private := range []string{"private-key-value", "123456", "operator@tenant.example", "private.example", "private-value"} {
		if bytes.Contains(got, []byte(private)) {
			t.Fatal("private fixture value retained")
		}
	}
	if !bytes.Contains(got, []byte(`"Identity": "Global"`)) || !bytes.Contains(got, []byte(`"Description": null`)) {
		t.Fatal("public identity or null changed")
	}
	if !bytes.Contains(got, []byte(`"name": "TeamsFilesPolicy"`)) {
		t.Fatal("technical schema name changed")
	}
	for _, invalid := range []string{`{"a":1,"a":2}`, `{} {}`, `{"a":`} {
		if _, err := fixtureJSON([]byte(invalid), nil); err == nil {
			t.Fatal("invalid/ambiguous JSON fixture accepted")
		}
	}
}

func TestUnitResponseFixturePreservesCaseVariantMemberOrder(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"allowEmailIntoChannel":false,"AllowEmailIntoChannel":true}`)
	got, err := fixtureJSON(raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	var original, sanitized struct{ AllowEmailIntoChannel bool }
	if json.Unmarshal(raw, &original) != nil || json.Unmarshal(got, &sanitized) != nil || original != sanitized {
		t.Fatal("fixture changed case-insensitive last-value decoding")
	}
}

func TestUnitResponseFixturePreservesScopedIdentityRelationship(t *testing.T) {
	t.Parallel()
	for _, id := range []string{"11111111-1111-4111-8111-111111111111", "private-record"} {
		t.Run(id, func(t *testing.T) {
			raw := []byte(`{"Identity":"Global/` + id + `","Id":"` + id + `"}`)
			got, err := fixtureJSON(raw, nil)
			var value struct{ Identity, Id string }
			if err != nil || json.Unmarshal(got, &value) != nil || value.Identity != "Global/"+value.Id {
				t.Fatal("fixture broke scoped identity/ID relationship")
			}
		})
	}
}

func TestUnitResponseFixtureRedactsConfiguredLabelsAndURLFragments(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"BrandedDisplayName":"private-value","LabelForAllowedAcsUsers":"private-value","Branding":{"Name":"private-value","Label":"private-value"},"Uri":"https://private.example/file#access_token=private-value","XmlRoot":{"name":"TeamsFilesPolicy"}}`)
	got, err := fixtureJSON(raw, nil)
	if err != nil || bytes.Contains(got, []byte("private-value")) {
		t.Fatal("private label or URL fragment retained")
	}
	if !bytes.Contains(got, []byte("TeamsFilesPolicy")) {
		t.Fatal("technical schema name changed")
	}
}

func TestUnitResponseFixtureRedactsCompactAudioID(t *testing.T) {
	t.Parallel()
	const privateID = "123456781234412381231234567890ab"
	raw, err := fixtureJSON([]byte(`{"AudioFileId":"`+privateID+`"}`), nil)
	if err != nil || bytes.Contains(raw, []byte(privateID)) {
		t.Fatal("private audio ID retained")
	}
	var value struct {
		ID string `json:"AudioFileId"`
	}
	if json.Unmarshal(raw, &value) != nil || len(value.ID) != 32 {
		t.Fatal("compact ID representation changed")
	}
	again, err := fixtureJSON(raw, nil)
	if err != nil || !bytes.Equal(again, raw) {
		t.Fatal("sanitized fixture is not stable")
	}
}

func TestUnitResponseFixtureRedactsMediaGroupID(t *testing.T) {
	t.Parallel()
	const private = "tenantglobal123456781234412381231234567890ab"
	raw, err := fixtureJSON([]byte(`{"GroupId":"`+private+`"}`), nil)
	if err != nil || bytes.Contains(raw, []byte("123456781234412381231234567890ab")) {
		t.Fatal("tenant-derived media group ID retained")
	}
	again, err := fixtureJSON(raw, nil)
	if err != nil || !bytes.Equal(again, raw) {
		t.Fatal("sanitized fixture is not stable")
	}
}

func TestUnitResponseFixtureKeepsOwnedTemplateSelectorsConsistent(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"name":"owned-template","displayName":"owned-template","channel":{"id":"General","displayName":"General"},"other":{"name":"private-name"}}`)
	value, err := fixtureJSON(raw, map[string]string{"owned-template": "fixture-template"})
	if err != nil || !bytes.Contains(value, []byte(`"name": "fixture-template"`)) || !bytes.Contains(value, []byte(`"displayName": "fixture-template"`)) || !bytes.Contains(value, []byte(`"id": "General"`)) || !bytes.Contains(value, []byte(`"displayName": "General"`)) || bytes.Contains(value, []byte("private-name")) {
		t.Fatal("template selector consistency or private-name redaction changed")
	}
}

func TestUnitResponseFixtureRedactsDirectoryAddressAndContact(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"objectId":"11111111-1111-4111-8111-111111111111","displayName":"private-name","city":"private-city","street":"private-street","postalCode":"private-postcode","state":"private-state","telephoneNumber":"+15555550123","mailNickname":"private-alias","userPrincipalName":"private-upn","verifiedDomains":[{"name":"private.example","status":"Verified"}],"sipDomains":["private.example"],"countryLetterCode":"GB"}`)
	got, err := fixtureJSON(raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, private := range []string{"11111111-1111-4111-8111-111111111111", "private-name", "private-city", "private-street", "private-postcode", "private-state", "+15555550123", "private-alias", "private-upn", "private.example"} {
		if bytes.Contains(got, []byte(private)) {
			t.Fatalf("private directory value retained: %s", private)
		}
	}
	if !bytes.Contains(got, []byte(`"status": "Verified"`)) || !bytes.Contains(got, []byte(`"countryLetterCode": "GB"`)) {
		t.Fatal("public directory state changed")
	}
	again, err := fixtureJSON(got, nil)
	if err != nil || !bytes.Equal(again, got) {
		t.Fatal("sanitized fixture is not stable")
	}
}
