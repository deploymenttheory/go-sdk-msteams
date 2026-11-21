package lab_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
)

var (
	fixtureGUIDPattern    = regexp.MustCompile(`(?i)[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}`)
	fixtureEmailPattern   = regexp.MustCompile(`(?i)[a-z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-z0-9.-]+\.[a-z]{2,}`)
	fixtureServicePattern = regexp.MustCompile(`^(?:[a-z][a-z0-9_]*/)*[a-z][a-z0-9_]*$`)
	fixtureLabelPattern   = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	errResponseFixture    = errors.New("invalid response test fixture")
)

func fixtureString(key, parent, value string, replacements map[string]string) string {
	keys := make([]string, 0, len(replacements))
	for original := range replacements {
		keys = append(keys, original)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	for _, original := range keys {
		value = strings.ReplaceAll(value, original, replacements[original])
	}
	value = fixtureGUIDPattern.ReplaceAllStringFunc(value, func(id string) string {
		if strings.HasPrefix(id, "00000000-0000-4000-8000-") {
			return id
		}
		hash := sha256.Sum256([]byte(strings.ToLower(id)))
		return fmt.Sprintf("00000000-0000-4000-8000-%x", hash[:6])
	})
	value = fixtureEmailPattern.ReplaceAllStringFunc(value, func(address string) string {
		hash := sha256.Sum256([]byte(strings.ToLower(address)))
		return fmt.Sprintf("user-%x@example.invalid", hash[:4])
	})
	if value == "" {
		return value
	}
	if strings.EqualFold(value, "Global") {
		return value
	}
	normalized := strings.ToLower(strings.NewReplacer("_", "", "-", "", ".", "").Replace(key))
	for _, sensitive := range []string{"password", "secret", "token", "apikey", "privatekey", "contentpin", "authorization"} {
		if strings.Contains(normalized, sensitive) {
			return "fixture-secret"
		}
	}
	switch normalized {
	case "audiofileid":
		// Audio uploads can use compact IDs without UUID separators.
		if len(value) == 32 && !strings.HasPrefix(value, "00000000000040008000") {
			if _, err := hex.DecodeString(value); err == nil {
				hash := sha256.Sum256([]byte(strings.ToLower(value)))
				return fmt.Sprintf("00000000000040008000%x", hash[:6])
			}
		}
	case "message":
		switch value {
		case "Type  is not currently enabled in flighting", "Type TeamsAudioConferencingCustomPromptsConfiguration is not currently enabled in flighting":
			return value
		}
		return "Fixture message"
	case "description", "displayname", "brandeddisplayname", "labelforallowedacsusers", "label", "text", "textprompt":
		if value != "SDK lab non-default initial description" && value != "SDK lab non-default updated description" {
			return "Fixture " + normalized
		}
	case "name":
		if !strings.EqualFold(parent, "XmlRoot") && !strings.EqualFold(parent, "XmlType") {
			return "Fixture name"
		}
	case "identity":
		if fixtureGUIDPattern.FindString(value) == value {
			return value
		}
		if !strings.EqualFold(value, "Global") && value != "fixture-policy" && !strings.EqualFold(value, "Tag:fixture-policy") {
			if prefix, component, scoped := strings.Cut(value, "/"); scoped && strings.EqualFold(prefix, "Global") {
				if fixtureGUIDPattern.FindString(component) == component {
					return prefix + "/" + component
				}
				hash := sha256.Sum256([]byte(component))
				return fmt.Sprintf("%s/fixture-record-%x", prefix, hash[:4])
			}
			hash := sha256.Sum256([]byte(value))
			return fmt.Sprintf("fixture-policy-%x", hash[:4])
		}
	case "blockedusers", "restrictedsenderlist":
		return "fixture-user@example.invalid"
	case "id":
		if !fixtureGUIDPattern.MatchString(value) {
			hash := sha256.Sum256([]byte(value))
			return fmt.Sprintf("fixture-record-%x", hash[:4])
		}
	case "configid":
		hash := sha256.Sum256([]byte(value))
		return fmt.Sprintf("fixture-config-%x", hash[:4])
	}
	if strings.Contains(normalized, "domain") && strings.Contains(value, ".") {
		return "fixture.example.invalid"
	}
	if strings.HasPrefix(value, "eyJ") && strings.Count(value, ".") == 2 {
		return "fixture-token"
	}
	if lower := strings.ToLower(value); strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		u, err := url.Parse(value)
		if err == nil {
			host := strings.ToLower(u.Hostname())
			if normalized == "adminserviceendpoint" {
				u.Host = "admin.example.invalid"
			} else if host != "graph.microsoft.com" && !strings.HasSuffix(host, ".microsoft.com") {
				u.Host = "fixture.example.invalid"
			}
			u.User = nil
			u.Fragment, u.RawFragment = "", ""
			query := u.Query()
			for name := range query {
				query.Set(name, "fixture-value")
			}
			u.RawQuery = query.Encode()
			value = u.String()
		}
	}
	if normalized == "adminserviceendpoint" && !strings.Contains(value, "://") {
		value = "admin.example.invalid"
	}
	return value
}

// fixtureJSON changes private values only. It preserves member and array order,
// nulls and numeric spelling; duplicate keys and trailing JSON are rejected.
func fixtureJSON(raw []byte, replacements map[string]string) ([]byte, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var read func(string, string) ([]byte, error)
	read = func(key, parent string) ([]byte, error) {
		token, err := decoder.Token()
		if err != nil {
			return nil, errResponseFixture
		}
		switch value := token.(type) {
		case json.Delim:
			switch value {
			case '{':
				result := []byte{'{'}
				seen := map[string]bool{}
				for decoder.More() {
					token, err := decoder.Token()
					if err != nil {
						return nil, errResponseFixture
					}
					name, ok := token.(string)
					if !ok || seen[name] {
						return nil, errResponseFixture
					}
					if len(seen) > 0 {
						result = append(result, ',')
					}
					seen[name] = true
					field, _ := json.Marshal(name)
					result = append(result, field...)
					result = append(result, ':')
					child, err := read(name, key)
					if err != nil {
						return nil, err
					}
					result = append(result, child...)
				}
				if end, err := decoder.Token(); err != nil || end != json.Delim('}') {
					return nil, errResponseFixture
				}
				return append(result, '}'), nil
			case '[':
				result := []byte{'['}
				for decoder.More() {
					if len(result) > 1 {
						result = append(result, ',')
					}
					child, err := read(key, parent)
					if err != nil {
						return nil, err
					}
					result = append(result, child...)
				}
				if end, err := decoder.Token(); err != nil || end != json.Delim(']') {
					return nil, errResponseFixture
				}
				return append(result, ']'), nil
			default:
				return nil, errResponseFixture
			}
		case string:
			return json.Marshal(fixtureString(key, parent, value, replacements))
		case json.Number:
			if strings.Contains(strings.ToLower(key), "contentpin") {
				return []byte("0"), nil
			}
			return []byte(value.String()), nil
		case bool:
			return json.Marshal(value)
		case nil:
			return []byte("null"), nil
		default:
			return nil, errResponseFixture
		}
	}
	value, err := read("", "")
	if err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, errResponseFixture
	}
	var formatted bytes.Buffer
	if err := json.Indent(&formatted, value, "", "  "); err != nil {
		return nil, errResponseFixture
	}
	return formatted.Bytes(), nil
}

func writeResponseFixture(t *testing.T, service, name string, response *resty.Response, replacements map[string]string) {
	t.Helper()
	if os.Getenv("MSTEAMS_UPDATE_TESTDATA") != "1" {
		return
	}
	if response == nil {
		return // Preserve the operation's original failure diagnostic in the caller.
	}
	if !fixtureServicePattern.MatchString(service) || !fixtureLabelPattern.MatchString(name) {
		t.Fatal("response and valid service/testdata name are required")
	}
	if response.StatusCode() != 200 {
		name = fmt.Sprintf("%s_%d", name, response.StatusCode())
	}
	raw := response.Body()
	extension := ".body"
	if len(raw) != 0 {
		var err error
		raw, err = fixtureJSON(raw, replacements)
		if err != nil {
			t.Fatal("response fixture JSON could not be safely written")
		}
		extension = ".json"
		raw = append(raw, '\n')
	}
	directory := filepath.Join("..", "..", "..", "msteams", "teams_admin_api", service, "testdata")
	if err := os.MkdirAll(directory, 0o750); err != nil {
		t.Fatal("service testdata directory unavailable")
	}
	if err := os.WriteFile(filepath.Join(directory, name+extension), raw, 0o600); err != nil {
		t.Fatal("response fixture could not be written")
	}
}

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
