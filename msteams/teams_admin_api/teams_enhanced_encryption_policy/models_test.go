package teamsenhancedencryptionpolicy_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_enhanced_encryption_policy"
)

func TestEnhancedEncryptionKnownProjectionsAndOwnedSnapshot(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"Identity":"Global","Description":"","CallingEndtoEndEncryptionEnabledType":"FutureCalling","MeetingEndToEndEncryption":"FutureMeeting","Future":{"n":9007199254740993}}`)
	original := append([]byte(nil), raw...)
	var value resource.Policy
	if json.Unmarshal(raw, &value) != nil || value.Description == nil || *value.Description != "" || value.CallingEndToEndEncryptionEnabledType == nil || *value.CallingEndToEndEncryptionEnabledType != "FutureCalling" || value.MeetingEndToEndEncryption == nil || *value.MeetingEndToEndEncryption != "FutureMeeting" {
		t.Fatal("known projection/presence changed")
	}
	clear(raw)
	*value.CallingEndToEndEncryptionEnabledType = "changed"
	value.RawFields["Future"][0] = 'x'
	got, e := json.Marshal(value)
	if e != nil || !bytes.Equal(got, original) {
		t.Fatal("original snapshot borrowed storage or changed with projections")
	}
	for _, format := range []string{"%v", "%+v", "%#v"} {
		if strings.Contains(fmt.Sprintf(format, value), "FutureMeeting") {
			t.Fatal("ordinary formatting exposed policy content")
		}
	}
}

func TestEnhancedEncryptionUnprojectableValuesStayRaw(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"CallingEndtoEndEncryptionEnabledType":{},"MeetingEndToEndEncryption":false,"Description":[]}`, `{"CallingEndtoEndEncryptionEnabledType":0,"MeetingEndToEndEncryption":[],"Description":{}}`} {
		var value resource.Policy
		if json.Unmarshal([]byte(raw), &value) != nil || value.CallingEndToEndEncryptionEnabledType != nil || value.MeetingEndToEndEncryption != nil || value.Description != nil || len(value.RawFields["CallingEndtoEndEncryptionEnabledType"]) == 0 {
			t.Fatal("future shape lost or projected as a zero value")
		}
	}
	var value resource.Policy
	if json.Unmarshal([]byte(`{"MeetingEndToEndEncryption":null,"CallingEndtoEndEncryptionEnabledType":""}`), &value) != nil || value.MeetingEndToEndEncryption != nil || value.Description != nil || value.CallingEndToEndEncryptionEnabledType == nil || *value.CallingEndToEndEncryptionEnabledType != "" {
		t.Fatal("omitted/null/empty collapsed")
	}
}
