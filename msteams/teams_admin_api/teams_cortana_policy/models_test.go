package teamscortanapolicy_test

import (
	"bytes"
	"encoding/json"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_cortana_policy"
)

func TestUnitCortanaKnownProjectionsAndOwnedRawSnapshot(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"Identity":"Global","Description":"","CortanaVoiceInvocationMode":"FutureMode","AllowCortanaVoiceInvocation":false,"AllowCortanaAmbientListening":true,"AllowCortanaInContextSuggestions":false,"Future":{"n":9007199254740993}}`)
	original := append([]byte(nil), raw...)
	var value resource.Policy
	if json.Unmarshal(raw, &value) != nil || value.Description == nil || *value.Description != "" || value.CortanaVoiceInvocationMode == nil || *value.CortanaVoiceInvocationMode != "FutureMode" || value.AllowCortanaVoiceInvocation == nil || *value.AllowCortanaVoiceInvocation || value.AllowCortanaAmbientListening == nil || !*value.AllowCortanaAmbientListening || value.AllowCortanaInContextSuggestions == nil || *value.AllowCortanaInContextSuggestions {
		t.Fatal("known presence/projection changed")
	}
	clear(raw)
	*value.CortanaVoiceInvocationMode = "changed"
	value.RawFields["Future"][0] = 'x'
	got, e := json.Marshal(value)
	if e != nil || !bytes.Equal(got, original) {
		t.Fatal("snapshot borrowed storage or changed with projections")
	}
}
func TestUnitCortanaUnprojectableValuesStayRaw(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"AllowCortanaVoiceInvocation":"future","CortanaVoiceInvocationMode":{},"Description":[]}`, `{"AllowCortanaVoiceInvocation":0,"CortanaVoiceInvocationMode":false,"Description":{}}`} {
		var value resource.Policy
		if json.Unmarshal([]byte(raw), &value) != nil || value.AllowCortanaVoiceInvocation != nil || value.CortanaVoiceInvocationMode != nil || value.Description != nil || len(value.RawFields["AllowCortanaVoiceInvocation"]) == 0 {
			t.Fatal("future shape lost or projected as a zero value")
		}
	}
	var value resource.Policy
	if json.Unmarshal([]byte(`{"AllowCortanaAmbientListening":false,"AllowCortanaInContextSuggestions":null,"CortanaVoiceInvocationMode":""}`), &value) != nil || value.AllowCortanaAmbientListening == nil || *value.AllowCortanaAmbientListening || value.AllowCortanaInContextSuggestions != nil || value.AllowCortanaVoiceInvocation != nil || value.CortanaVoiceInvocationMode == nil || *value.CortanaVoiceInvocationMode != "" {
		t.Fatal("omitted/null/false/empty collapsed")
	}
}
