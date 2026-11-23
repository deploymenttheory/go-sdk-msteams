package teamscustombannertextacc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_custom_banner_text"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

type bannerOracleReport struct {
	Success      bool   `json:"success"`
	Tenant       string `json:"tenant_id"`
	Count        *int   `json:"count"`
	Observations []struct {
		Identity string          `json:"identity"`
		Settings json.RawMessage `json:"settings"`
	} `json:"observations"`
}

func decodeBannerOracle(raw []byte, tenant string, identities []string, count int) (map[string]json.RawMessage, error) {
	if len(raw) > labkit.MaxOracleBytes || !labkit.DirectoryID.MatchString(tenant) || len(identities) > 5 || count < len(identities) {
		return nil, labkit.ErrOracle
	}
	var v bannerOracleReport
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || !strings.EqualFold(v.Tenant, tenant) || v.Count == nil || *v.Count != count || v.Observations == nil || len(v.Observations) != len(identities) {
		return nil, labkit.ErrOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return nil, labkit.ErrOracle
	}
	allowed := map[string]bool{}
	for _, identity := range identities {
		if identity == "" || allowed[identity] {
			return nil, labkit.ErrOracle
		}
		allowed[identity] = true
	}
	result := map[string]json.RawMessage{}
	for _, o := range v.Observations {
		if !allowed[o.Identity] || result[o.Identity] != nil {
			return nil, labkit.ErrOracle
		}
		var fields map[string]json.RawMessage
		if json.Unmarshal(o.Settings, &fields) != nil || len(fields) != 3 {
			return nil, labkit.ErrOracle
		}
		for _, key := range []string{"Id", "Text", "Description"} {
			var nullable *string
			raw, ok := fields[key]
			if !ok || json.Unmarshal(raw, &nullable) != nil {
				return nil, labkit.ErrOracle
			}
		}
		// The PowerShell projection includes null for an absent CLR Text. Do not
		// feed that presentation through the SDK's explicit-null Text setter.
		projection, e := json.Marshal(map[string]json.RawMessage{"Id": fields["Id"], "Description": fields["Description"]})
		var model resource.Banner
		if e != nil || json.Unmarshal(projection, &model) != nil {
			return nil, labkit.ErrOracle
		}
		result[o.Identity] = o.Settings
	}
	return result, nil
}

func bannerProjection(v *resource.Banner) []byte {
	raw, _ := json.Marshal(map[string]*string{"Id": v.ID, "Text": v.Text, "Description": v.Description})
	return raw
}

func sameBannerCollections(left, right []*resource.Banner) bool {
	if len(left) != len(right) {
		return false
	}
	byIdentity := map[string][]byte{}
	for _, v := range right {
		if v == nil || v.Identity == nil || *v.Identity == "" || byIdentity[*v.Identity] != nil {
			return false
		}
		byIdentity[*v.Identity] = bannerProjection(v)
	}
	for _, v := range left {
		if v == nil || v.Identity == nil || !bytes.Equal(byIdentity[*v.Identity], bannerProjection(v)) {
			return false
		}
		delete(byIdentity, *v.Identity)
	}
	return len(byIdentity) == 0
}

func bannerOracle(t *testing.T, ctx context.Context, tenant string, identities []string, count int) map[string]json.RawMessage {
	t.Helper()
	input, e := json.Marshal(map[string]any{"tenant_id": tenant, "identities": identities})
	if e != nil {
		t.Fatal("invalid banner oracle selector")
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-banner-text.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"), "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned banner reads failed; no output retained")
	}
	v, e := decodeBannerOracle(output.Data, tenant, identities, count)
	if e != nil {
		t.Fatal("pinned banner projection unavailable")
	}
	return v
}

// TestAccBannerTextReadOnly lists banner records and reads at most five of them.
// It never creates banners, updates text, deletes records or grants access.
func TestAccBannerTextReadOnly(t *testing.T) {
	cfg, e := lab.Load()
	if e != nil {
		t.Fatal("invalid lab enablement")
	}
	if e := cfg.Require(lab.ReadOnly); errors.Is(e, lab.ErrDisabled) {
		t.Skip("banner reads require explicit read-only opt-in")
	} else if e != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	values, response, e := sdk.TeamsAdminAPI.TeamsCustomBannerTexts.List(ctx)
	labkit.WriteResponseFixture(t, "teams_custom_banner_text", "list", response, nil)
	if e != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("banner list failed (HTTP %d); no response payload retained", status)
	}
	filtered, _, e := sdk.TeamsAdminAPI.TeamsCustomBannerTexts.ListFiltered(ctx, "*")
	if e != nil || !sameBannerCollections(values, filtered) {
		t.Fatal("banner filtered list differs")
	}
	identities := []string{}
	projections := map[string][]byte{}
	for _, v := range values {
		if len(identities) == 5 {
			break
		}
		if v == nil || v.Identity == nil || *v.Identity == "" || projections[*v.Identity] != nil {
			t.Fatal("banner identity unavailable or duplicated")
		}
		read, fixtureResponse, e := sdk.TeamsAdminAPI.TeamsCustomBannerTexts.GetByIdentity(ctx, *v.Identity)
		labkit.WriteResponseFixture(t, "teams_custom_banner_text", "get_by_id", fixtureResponse, nil)
		if e != nil || read == nil || read.Identity == nil || *read.Identity != *v.Identity {
			t.Fatal("selected banner read failed; no output retained")
		}
		before, after := bannerProjection(v), bannerProjection(read)
		if !bytes.Equal(before, after) {
			t.Fatal("banner list/read projections differ")
		}
		if v.ID != nil && *v.Identity == "Global/"+*v.ID {
			relative, _, e := sdk.TeamsAdminAPI.TeamsCustomBannerTexts.GetByID(ctx, *v.ID)
			if e != nil || relative == nil || relative.Identity == nil || *relative.Identity != *v.Identity || !bytes.Equal(after, bannerProjection(relative)) {
				t.Fatal("banner relative-ID read differs")
			}
		}
		identities = append(identities, *v.Identity)
		projections[*v.Identity] = after
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := bannerOracle(t, ctx, tenant, identities, len(values))
		for identity, want := range projections {
			var left, right any
			if json.Unmarshal(want, &left) != nil || json.Unmarshal(actual[identity], &right) != nil || !reflect.DeepEqual(left, right) {
				t.Fatal("Go and pinned banner projections differ")
			}
		}
	}
	t.Log("bounded banner reads verified; mutations, cleanup and displayed behavior remain separately qualified")
}

func TestUnitBannerCollectionComparisonAndOracleBoundary(t *testing.T) {
	t.Parallel()
	a := &resource.Banner{Identity: new("Global/a")}
	b := &resource.Banner{Identity: new("Global/b")}
	if !sameBannerCollections([]*resource.Banner{a, b}, []*resource.Banner{b, a}) {
		t.Fatal("independent response order changed comparison")
	}
	for _, other := range [][]*resource.Banner{{a, a}, {a, nil}, {a, {Identity: b.Identity, Text: new("changed")}}} {
		if sameBannerCollections([]*resource.Banner{a, b}, other) {
			t.Fatal("duplicate, missing or changed record accepted")
		}
	}
	const tenant = "00000000-0000-4000-8000-000000000003"
	const identity = "Global/fixture"
	const observation = `{"identity":"Global/fixture","settings":{"Id":null,"Text":null,"Description":null}}`
	valid := `{"success":true,"tenant_id":"` + tenant + `","count":1,"observations":[` + observation + `]}`
	if _, e := decodeBannerOracle([]byte(valid), tenant, []string{identity}, 1); e != nil {
		t.Fatal("complete nullable projection rejected")
	}
	empty := `{"success":true,"tenant_id":"` + tenant + `","count":0,"observations":[]}`
	if _, e := decodeBannerOracle([]byte(empty), tenant, nil, 0); e != nil {
		t.Fatal("empty list projection rejected")
	}
	for _, raw := range []string{`{"success":false}`, valid + "{}", strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, identity, "Other", 1), strings.Replace(valid, `"Description":null`, `"Description":false`, 1), strings.Replace(valid, `,"Text":null`, "", 1), strings.Replace(valid, observation, observation+","+observation, 1), strings.Replace(valid, `"Id":null`, `"Id":"invalid"`, 1), strings.Replace(valid, `"count":1`, `"count":0`, 1), strings.Replace(valid, `"count":1,`, "", 1)} {
		if _, e := decodeBannerOracle([]byte(raw), tenant, []string{identity}, 1); e == nil {
			t.Fatal("invalid banner oracle output accepted")
		}
	}
}
