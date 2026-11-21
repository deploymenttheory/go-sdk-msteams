package channelmember_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	channelmember "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_member"
)

func TestUnitReadPreservesOpaqueAndDirectoryIdentities(t *testing.T) {
	t.Parallel()
	const raw = `{"@odata.type":"#microsoft.graph.aadUserConversationMember","id":"opaque+/=:#%雪","userId":"22222222-2222-4222-8222-222222222222","tenantId":"other-tenant","roles":["owner","futureQualifier"],"displayName":null,"email":"","visibleHistoryStartDateTime":"2026-01-02T03:04:05Z","future":{"enabled":false}}`
	var value channelmember.Member
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		t.Fatal(err)
	}
	if value.ID != "opaque+/=:#%雪" || value.UserID == nil || *value.UserID != "22222222-2222-4222-8222-222222222222" || value.TenantID == nil || *value.TenantID != "other-tenant" {
		t.Fatal("membership, user or tenant identity changed")
	}
	if value.Type == nil || *value.Type != "#microsoft.graph.aadUserConversationMember" || len(value.Roles) != 2 || value.Roles[1] != "futureQualifier" || value.DisplayName != nil || value.Email == nil || *value.Email != "" || value.VisibleHistoryStartDateTime == nil {
		t.Fatal("discriminator, role, null/empty or date information lost")
	}
	if string(value.RawFields["future"]) != `{"enabled":false}` || string(value.RawFields["displayName"]) != "null" {
		t.Fatal("future or presence information lost")
	}
}

func TestUnitReadRetainsUnknownSubtypeAndNullableAADFields(t *testing.T) {
	t.Parallel()
	for _, roles := range []string{"null", "[]", `["guest"]`} {
		var value channelmember.Member
		raw := `{"id":"opaque","@odata.type":"#microsoft.graph.futureConversationMember","roles":` + roles + `,"userId":null,"email":null,"futureSubtypeField":7}`
		if json.Unmarshal([]byte(raw), &value) != nil || value.UserID != nil || value.Email != nil || value.Type == nil || *value.Type != "#microsoft.graph.futureConversationMember" || string(value.RawFields["roles"]) != roles || string(value.RawFields["futureSubtypeField"]) != "7" {
			t.Fatal("future subtype or nullable AAD fields were rejected or defaulted")
		}
		if roles == "[]" && value.Roles == nil {
			t.Fatal("explicit empty read roles became null")
		}
	}
}

func TestUnitReadRequiresMembershipIdentityAndRedactsDecodeErrors(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`null`, `[]`, `{}`, `{"userId":"directory-only"}`, `{"id":""}`, `{"id":7}`, `{"id":"opaque","roles":"private-value"}`} {
		var value channelmember.Member
		err := json.Unmarshal([]byte(raw), &value)
		if !errors.Is(err, client.ErrDecode) || strings.Contains(err.Error(), "private-value") || value.ID != "" {
			t.Fatal("malformed resource accepted, substituted a directory identity, or exposed data")
		}
	}
}

func TestUnitRolePatchDistinguishesOmissionNullDemotionAndPromotion(t *testing.T) {
	t.Parallel()
	var null []channelmember.Role
	empty := []channelmember.Role{}
	owner := []channelmember.Role{channelmember.Owner}
	for _, tc := range []struct {
		name  string
		roles *[]channelmember.Role
		want  string
	}{
		{"omitted", nil, `{}`},
		{"null-is-not-demotion", &null, `{"roles":null}`},
		{"demote", &empty, `{"roles":[]}`},
		{"promote", &owner, `{"roles":["owner"]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(channelmember.UpdateRequest{Roles: tc.roles})
			if err != nil || string(raw) != tc.want {
				t.Fatal("role patch changed omission/null/empty/owner semantics")
			}
			if strings.Contains(string(raw), "userId") || strings.Contains(string(raw), "visibleHistory") {
				t.Fatal("read-only or chat-only field entered team patch")
			}
		})
	}
}
