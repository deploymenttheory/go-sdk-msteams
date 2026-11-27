package tenantidentity_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	ti "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_identity"
)

type tenantWireObservation struct {
	Name       string
	JSON       string `json:"input_json"`
	Exception  json.RawMessage
	Properties struct {
		DisplayName           *string
		AnnouncementsDisabled *bool
		ServiceNumberCount    *int32
		CompanyTags           []*string
	}
}

func tenantWireObservations(t *testing.T) []tenantWireObservation {
	t.Helper()
	var fixture struct{ Cases []tenantWireObservation }
	if json.Unmarshal(readFile(t, "testdata/pinned-tenant-wire.json"), &fixture) != nil || len(fixture.Cases) != 9 {
		t.Fatal("own pinned tenant observations unavailable")
	}
	return fixture.Cases
}
func TestUnitTenantWireUsesOwnExactFieldsAndNumericProfile(t *testing.T) {
	for _, item := range tenantWireObservations(t) {
		t.Run(item.Name, func(t *testing.T) {
			var value ti.Tenant
			err := json.Unmarshal([]byte(item.JSON), &value)
			if string(item.Exception) != "null" {
				if !errors.Is(err, client.ErrDecode) {
					t.Fatal("own generated duplicate/format/range failure accepted")
				}
				return
			}
			if err != nil {
				t.Fatal("source-accepted standard JSON rejected")
			}
			got, _ := json.Marshal([]any{value.DisplayName, value.AnnouncementsDisabled, value.ServiceNumberCount, value.CompanyTags})
			want, _ := json.Marshal([]any{item.Properties.DisplayName, item.Properties.AnnouncementsDisabled, item.Properties.ServiceNumberCount, item.Properties.CompanyTags})
			if string(got) != string(want) {
				t.Fatal("own exact field/type projection changed")
			}
		})
	}
}
func TestUnitTenantWireTypesEveryKnownFieldAndRetainsRawDatesAndMaps(t *testing.T) {
	item := tenantWireObservations(t)[0]
	var value ti.Tenant
	if json.Unmarshal([]byte(item.JSON), &value) != nil {
		t.Fatal("full wire view rejected")
	}
	if value.ObjectID == nil || *value.ObjectID != "opaque-tenant" || value.TNMAccountID == nil || *value.TNMAccountID != "opaque-account" {
		t.Fatal("opaque tenant identifiers were constrained or lost")
	}
	if value.AnnouncementsDisabled == nil || *value.AnnouncementsDisabled || value.DirSyncEnabled == nil || !*value.DirSyncEnabled || value.NameRecordingDisabled == nil || *value.NameRecordingDisabled {
		t.Fatal("explicit false/true lost")
	}
	for _, field := range []*string{value.Country, value.DefaultPoolFQDN, value.State, value.WhenChanged} {
		if field == nil {
			t.Fatal("known tenant field was left raw-only")
		}
	}
	if value.SubscriberNumberCount == nil || *value.SubscriberNumberCount != 9 || len(value.CompanyPartnerships) != 1 || value.CompanyPartnerships[0].SupportPartner == nil || !*value.CompanyPartnerships[0].SupportPartner || value.CompanyPartnerships[0].PartnerContextID == nil {
		t.Fatal("known nested partnership/count lost")
	}
	if len(value.AssignedPlans) != 1 || value.AssignedPlans[0].ServiceInstance == nil || *value.AssignedPlans[0].ServiceInstance != "plan-instance" {
		t.Fatal("known assigned-plan instance lost")
	}
	if value.SyncInLyncADInfo == nil || value.SyncInLyncADInfo.IsUserSyncDisabled == nil || *value.SyncInLyncADInfo.IsUserSyncDisabled || value.SyncInLyncADInfo.StopSyncTimestamp == nil || string(value.SyncInLyncADInfo.RawFields["custom"]) != `{"x":1}` {
		t.Fatal("known sync object/raw extras lost")
	}
	for _, items := range [][]*string{value.CompanyTags, value.Pools, value.SIPDomains} {
		if len(items) != 2 || items[0] == nil || items[1] != nil {
			t.Fatal("nullable string positions changed")
		}
	}
	if value.WhenCreated == nil || *value.WhenCreated != "2026-11-27T03:00:00.1234567+02:00" || value.LastSyncTimeStamp == nil || *value.LastSyncTimeStamp != "2026-11-27T03:00:00" {
		t.Fatal("native wire timestamps incorrectly became CLR projections")
	}
	var input map[string]json.RawMessage
	if json.Unmarshal([]byte(item.JSON), &input) != nil {
		t.Fatal("wire input unavailable")
	}
	var expected map[string]json.RawMessage
	if json.Unmarshal(input["lastProvisionTimeStamps"], &expected) != nil {
		t.Fatal("wire map unavailable")
	}
	for key, raw := range expected {
		if string(value.LastProvisionTimeStamps[key]) != string(raw) {
			t.Fatal("object-valued wire map was coerced or dropped")
		}
	}
	if string(value.LastPublishTimeStamps["number"]) != "9007199254740993" || string(value.LastPublishTimeStamps["nil"]) != "null" {
		t.Fatal("wire map lost numeric precision or explicit null")
	}
	value.LastProvisionTimeStamps["text"] = json.RawMessage(`"changed"`)
	if string(value.RawFields["lastProvisionTimeStamps"]) != string(input["lastProvisionTimeStamps"]) {
		t.Fatal("map edit mutated original wire document")
	}
}
func TestUnitTenantWireReadRetainsMixedMapsAndMalformedReceipt(t *testing.T) {
	good := []byte(tenantWireObservations(t)[0].JSON)
	for _, item := range []struct {
		name  string
		reply []byte
		want  error
	}{
		{"mixed-map", good, nil},
		{"overflow", []byte(`{"serviceNumberCount":2147483648}`), client.ErrDecode},
		{"duplicate", []byte(`{"displayName":"a","displayName":"b"}`), client.ErrDecode},
	} {
		t.Run(item.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: "tenant-wire-" + item.name, Source: "msteams/teams_admin_api/tenant_identity/testdata/pinned-tenant-wire.json", Mappings: mapping("Get-CsTenant", "Identity"), Steps: []contracttest.Exchange{discovery(t), wire("GET", "/Teams.Tenant/tenants?defaultpropertyset=Extended", "Get-CsTenant", nil, 200, item.reply)}, Want: contracttest.Expected{Status: 200, Body: item.reply, Error: item.want}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.TenantIdentity.GetTenant(ctx)
				if response == nil || string(response.Body()) != string(item.reply) {
					t.Fatal("accepted receipt lost")
				}
				if item.want == nil && (value == nil || string(value.LastPublishTimeStamps["number"]) != "9007199254740993") {
					t.Fatal("native wire read lost map values")
				}
				if err != nil && value != nil {
					t.Fatal("decode failure leaked a partial model")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
		})
	}
}
