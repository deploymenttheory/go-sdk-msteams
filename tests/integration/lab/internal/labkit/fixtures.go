package labkit

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
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
	fixtureGUIDPattern  = regexp.MustCompile(`(?i)[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}`)
	fixtureEmailPattern = regexp.MustCompile(`(?i)[a-z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-z0-9.-]+\.[a-z]{2,}`)
	// E.164 numbers, bare or in tel: URIs. +1555… fixture numbers are kept.
	fixturePhonePattern   = regexp.MustCompile(`\+[1-9][0-9]{6,14}\b`)
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
	value = fixturePhonePattern.ReplaceAllStringFunc(value, func(number string) string {
		if strings.HasPrefix(number, "+1555") {
			return number
		}
		hash := sha256.Sum256([]byte(number))
		return fmt.Sprintf("+1555%07d", (uint32(hash[0])<<16|uint32(hash[1])<<8|uint32(hash[2]))%10000000)
	})
	if value == "" {
		return value
	}
	if address := fixtureAddress(key, value); address != "" {
		return address
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
	if (normalized == "name" || normalized == "displayname") && value == "fixture-template" {
		return value
	}
	if (normalized == "id" || normalized == "name" || normalized == "displayname") && value == "General" {
		return value
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
	case "city", "street", "postalcode", "state", "stateorprovince", "telephonenumber", "phonenumber", "mobile", "faxnumber", "mailnickname", "userprincipalname", "givenname", "surname", "companyname", "department", "jobtitle", "officelocation", "streetaddress":
		// Directory tenant and user records carry address, contact and naming data.
		return "Fixture " + normalized
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
	case "groupid":
		// Media storage group IDs embed the tenant ID without separators.
		if fixtureGUIDPattern.FindString(value) != value && !strings.HasPrefix(value, "fixture-group-") {
			hash := sha256.Sum256([]byte(strings.ToLower(value)))
			return fmt.Sprintf("fixture-group-%x", hash[:4])
		}
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

// fixtureAddress maps a whole-value IP address to a documentation address.
// IPv4 is only treated as an address under an IP/address key, since version
// strings such as 1.0.0.0 parse as IPv4 too.
func fixtureAddress(key, value string) string {
	ip := net.ParseIP(value)
	if ip == nil {
		return ""
	}
	hash := sha256.Sum256([]byte(value))
	if ip.To4() == nil {
		return fmt.Sprintf("2001:db8::%x", hash[:2])
	}
	normalized := strings.ToLower(key)
	if !strings.Contains(normalized, "ip") && !strings.Contains(normalized, "address") {
		return ""
	}
	return fmt.Sprintf("192.0.2.%d", hash[0]%254+1)
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

func WriteResponseFixture(t *testing.T, service, name string, response *resty.Response, replacements map[string]string) {
	t.Helper()
	WriteResponseFixtureWhen(t, os.Getenv("MSTEAMS_UPDATE_TESTDATA") == "1", service, name, response, replacements)
}

// WriteResponseFixtureWhen applies the existing sanitizer to an explicitly
// selected capture. It does not alter process-wide fixture writer opt-in.
func WriteResponseFixtureWhen(t *testing.T, enabled bool, service, name string, response *resty.Response, replacements map[string]string) {
	t.Helper()
	if !enabled {
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
	// Resolve from the module root: lab tests run in per-service packages.
	directory := filepath.Join(Root, "msteams", "teams_admin_api", service, "testdata")
	if err := os.MkdirAll(directory, 0o750); err != nil {
		t.Fatal("service testdata directory unavailable")
	}
	if err := os.WriteFile(filepath.Join(directory, name+extension), raw, 0o600); err != nil {
		t.Fatal("response fixture could not be written")
	}
}
