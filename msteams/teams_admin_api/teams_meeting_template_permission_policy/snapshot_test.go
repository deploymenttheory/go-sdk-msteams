package teamsmeetingtemplatepermissionpolicy_test

import (
	"bytes"
	"encoding/json"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_template_permission_policy"
)

func TestUnitAcceptanceSnapshotOwnsResponseBytes(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "get_global.json")
	var value resource.Policy
	if json.Unmarshal(raw, &value) != nil {
		t.Fatal("invalid acceptance response")
	}
	expected := append([]byte(nil), bytes.TrimSpace(raw)...)
	clear(raw)
	if !bytes.Equal(expected, value.RawJSON) || value.RawFields["ConfigMetadata"] == nil {
		t.Fatal("response snapshot or unknown metadata lost")
	}
	encoded, err := json.Marshal(value)
	var a, b any
	if err != nil || json.Unmarshal(expected, &a) != nil || json.Unmarshal(encoded, &b) != nil {
		t.Fatal("snapshot serialization failed")
	}
	normalizedA, _ := json.Marshal(a)
	normalizedB, _ := json.Marshal(b)
	if !bytes.Equal(normalizedA, normalizedB) {
		t.Fatal("received JSON values changed on marshal")
	}
}
