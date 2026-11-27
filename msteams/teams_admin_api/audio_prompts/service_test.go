package audioprompts_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	audio "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/audio_prompts"
	"github.com/go-resty/resty/v2"
)

const (
	configOrigin = "https://api.interfaces.records.teams.microsoft.com"
	targetForest = "https://admin.example.invalid/"

	fixtureID = "fixture-record-3eb1bd43"
)

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	root, err := os.OpenRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	raw, err := root.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// silentWAV matches the acceptance upload: 0.25 s of 16 kHz mono PCM silence.
func silentWAV() []byte {
	const size = 8000
	samples := make([]byte, size)
	var out bytes.Buffer
	write := func(v any) { _ = binary.Write(&out, binary.LittleEndian, v) }
	out.WriteString("RIFF")
	write(uint32(36 + size))
	out.WriteString("WAVEfmt ")
	for _, v := range []any{uint32(16), uint16(1), uint16(1), uint32(16000), uint32(32000), uint16(2), uint16(16)} {
		write(v)
	}
	out.WriteString("data")
	write(uint32(size))
	out.Write(samples)
	return out.Bytes()
}

func discoveryBody(t *testing.T) []byte {
	t.Helper()
	body, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func discovery(t *testing.T) testutil.Exchange {
	t.Helper()
	return testutil.Exchange{Request: testutil.WireRequest{Method: http.MethodGet, URL: configOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience},
		Reply: testutil.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: discoveryBody(t)}}
}

func wire(method, path, command string, body json.RawMessage, status int, contentType string, reply []byte) testutil.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {targetForest}}
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: configOrigin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body},
		Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {contentType}}, Body: reply}}
}

func run(t *testing.T, c testutil.Case) {
	t.Helper()

	result, err := testutil.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Fatalf("audio file contract %s differed: %+v", c.Name, result.Differences())
	}
}

func TestUnitAudioFileLifecycleUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	imported, got, exported := readFile(t, "testdata/import_201.json"), readFile(t, "testdata/get.json"), readFile(t, "testdata/export.json")
	deleted, deletedRead := readFile(t, "testdata/delete_204.body"), readFile(t, "testdata/get_deleted.json")
	content := silentWAV()
	body, err := json.Marshal(map[string]string{"ApplicationId": "TenantGlobal", "OriginalFilename": "fixture-audio.wav", "Content": base64.StdEncoding.EncodeToString(content)})
	if err != nil {
		t.Fatal(err)
	}
	item := "/Teams.MediaStorage/audiofile/TenantGlobal/" + fixtureID
	const jsonType = "application/json; charset=utf-8"
	run(t, testutil.Case{Name: "actual-lifecycle",

		Steps: []testutil.Exchange{discovery(t),
			wire("POST", "/Teams.MediaStorage/audiofile", "Import-CsOnlineAudioFile", body, 201, jsonType, imported),
			wire("GET", item, "Get-CsOnlineAudioFile", nil, 200, jsonType, got),
			wire("GET", item+"/content", "Export-CsOnlineAudioFile", nil, 200, jsonType, exported),
			wire("DELETE", item, "Remove-CsOnlineAudioFile", nil, 204, "", deleted),
			wire("GET", item, "Get-CsOnlineAudioFile", nil, 200, jsonType, deletedRead),
		},
		Want: testutil.Expected{Status: 200, Body: deletedRead},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.AudioFiles
			created, response, err := s.Create(ctx, audio.ImportRequest{FileName: "fixture-audio.wav", Content: content})
			if err != nil || response.StatusCode() != 201 || created.ID == nil || *created.ID != fixtureID || created.OriginalFilename == nil ||
				*created.OriginalFilename != "fixture-audio.wav" || created.Content == nil || len(created.Content) != 0 || created.MarkedForDeletion() {
				t.Fatal("actual import receipt was changed")
			}
			value, _, err := s.GetByID(ctx, "TENANTGLOBAL", fixtureID, nil)
			if err != nil || value.Duration == nil || *value.Duration != "00:00:00.2500000" || value.DownloadURI == nil || len(value.RawFields) != 13 {
				t.Fatal("actual detail was changed")
			}
			bytesOut, _, err := s.GetContentByID(ctx, "", fixtureID)
			if err != nil || len(bytesOut) < 44 || string(bytesOut[:4]) != "RIFF" || string(bytesOut[8:12]) != "WAVE" {
				t.Fatal("actual export was not decoded to WAV bytes")
			}
			if response, err = s.DeleteByID(ctx, fixtureID); err != nil || response.StatusCode() != 204 {
				t.Fatal("actual empty delete reply changed")
			}
			value, response, err = s.GetByID(ctx, audio.TenantGlobal, fixtureID, nil)
			if err != nil || !value.MarkedForDeletion() {
				t.Fatal("soft-deleted file lost its deletion mark")
			}
			return testutil.Observation{Response: response, Err: err}
		}})
}

func TestUnitAudioFileListNormalizesApplicationAndPreservesNulls(t *testing.T) {
	t.Parallel()
	list := []byte(`[{"Id":"a","ApplicationId":"HuntGroup","Content":null,"DeletionTimestampOffset":null,"Future":{"x":1}},{"Id":"b","Content":"","DeletionTimestampOffset":"2026-01-01T00:00:00+00:00"}]`)
	for index, input := range []string{"huntgroup", "HuntGroup", "HUNTGROUP"} {
		run(t, testutil.Case{Name: fmt.Sprintf("list-case-%d", index),
			Steps: []testutil.Exchange{discovery(t), wire("GET", "/Teams.MediaStorage/audiofile/HuntGroup", "Get-CsOnlineAudioFile", nil, 200, "application/json", list)},
			Want:  testutil.Expected{Status: 200, Body: list},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				values, response, err := sdk.TeamsAdminAPI.AudioFiles.List(ctx, input)
				if err != nil || len(values) != 2 || values[0].Content != nil || values[1].Content == nil || values[0].MarkedForDeletion() || !values[1].MarkedForDeletion() || values[0].RawFields["Future"] == nil {
					t.Fatal("list items lost null/empty/future distinctions")
				}
				return testutil.Observation{Response: response, Err: err}
			}})
	}
	empty := []byte(`[]`)
	run(t, testutil.Case{Name: "list-default-and-unknown", Steps: []testutil.Exchange{discovery(t),
		wire("GET", "/Teams.MediaStorage/audiofile/TenantGlobal", "Get-CsOnlineAudioFile", nil, 200, "application/json", empty),
		wire("GET", "/Teams.MediaStorage/audiofile/FutureApp", "Get-CsOnlineAudioFile", nil, 200, "application/json", empty)},
		Want: testutil.Expected{Status: 200, Body: empty},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			values, _, err := sdk.TeamsAdminAPI.AudioFiles.List(ctx, "")
			if err != nil || values == nil || len(values) != 0 {
				t.Fatal("empty list was not a non-nil empty slice")
			}
			values, response, err := sdk.TeamsAdminAPI.AudioFiles.List(ctx, "FutureApp")
			return testutil.Observation{Value: values, Response: response, Err: err}
		}})
}

func TestUnitAudioFileGetSendsDurationOnlyWhenSet(t *testing.T) {
	t.Parallel()
	got := readFile(t, "testdata/get.json")
	item := "/Teams.MediaStorage/audiofile/OrgAutoAttendant/" + fixtureID
	run(t, testutil.Case{Name: "duration", Steps: []testutil.Exchange{discovery(t),
		wire("GET", item+"?durationInMins=0", "Get-CsOnlineAudioFile", nil, 200, "application/json", got),
		wire("GET", item, "Get-CsOnlineAudioFile", nil, 200, "application/json", got)},
		Want: testutil.Expected{Status: 200, Body: got},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			zero := int32(0)
			if _, _, err := sdk.TeamsAdminAPI.AudioFiles.GetByID(ctx, "orgautoattendant", fixtureID, &audio.GetOptions{DurationInMinutes: &zero}); err != nil {
				t.Fatal("explicit zero duration was not sent")
			}
			value, response, err := sdk.TeamsAdminAPI.AudioFiles.GetByID(ctx, audio.OrgAutoAttendant, fixtureID, &audio.GetOptions{})
			return testutil.Observation{Value: value != nil, Response: response, Err: err}
		}})
}

func TestUnitAudioFileServiceFailuresStayTyped(t *testing.T) {
	t.Parallel()
	item := "/Teams.MediaStorage/audiofile/TenantGlobal/" + fixtureID
	missing := []byte(`"Could not find requested media file.`)
	run(t, testutil.Case{Name: "missing", Steps: []testutil.Exchange{discovery(t), wire("GET", item, "Get-CsOnlineAudioFile", nil, 404, "text/plain; charset=utf-8", missing)},
		Want: testutil.Expected{Status: 404, Body: missing, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 404}},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.AudioFiles.GetByID(ctx, "", fixtureID, nil)
			return testutil.Observation{Value: value, Response: response, Err: err}
		}})
	rejected := []byte(`Exception encountered while validating the audio file before upload.`)
	run(t, testutil.Case{Name: "repeat-delete", Steps: []testutil.Exchange{discovery(t), wire("DELETE", item, "Remove-CsOnlineAudioFile", nil, 400, "text/plain; charset=utf-8", rejected)},
		Want: testutil.Expected{Status: 400, Body: rejected, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			response, err := sdk.TeamsAdminAPI.AudioFiles.DeleteByID(ctx, fixtureID)
			return testutil.Observation{Response: response, Err: err}
		}})
	run(t, testutil.Case{Name: "undeclared-status", Steps: []testutil.Exchange{discovery(t), wire("DELETE", item, "Remove-CsOnlineAudioFile", nil, 202, "", nil)},
		Want: testutil.Expected{Status: 202, Error: audio.ErrResponse},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			response, err := sdk.TeamsAdminAPI.AudioFiles.DeleteByID(ctx, fixtureID)
			return testutil.Observation{Response: response, Err: err}
		}})
}

func TestUnitAudioFileMalformedResponsesFailDecode(t *testing.T) {
	t.Parallel()
	item := "/Teams.MediaStorage/audiofile/TenantGlobal/" + fixtureID
	for name, step := range map[string]struct {
		path, command string
		body          string
		invoke        func(context.Context, *audio.Service) (any, *resty.Response, error)
	}{
		"list-null": {"/Teams.MediaStorage/audiofile/TenantGlobal", "Get-CsOnlineAudioFile", `null`, func(ctx context.Context, s *audio.Service) (any, *resty.Response, error) {
			v, r, e := s.List(ctx, "")
			return v, r, e
		}},
		"list-null-item": {"/Teams.MediaStorage/audiofile/TenantGlobal", "Get-CsOnlineAudioFile", `[null]`, func(ctx context.Context, s *audio.Service) (any, *resty.Response, error) {
			v, r, e := s.List(ctx, "")
			return v, r, e
		}},
		"get-array": {item, "Get-CsOnlineAudioFile", `[]`, func(ctx context.Context, s *audio.Service) (any, *resty.Response, error) {
			v, r, e := s.GetByID(ctx, "", fixtureID, nil)
			return v, r, e
		}},
		"get-bad-type": {item, "Get-CsOnlineAudioFile", `{"Id":7}`, func(ctx context.Context, s *audio.Service) (any, *resty.Response, error) {
			v, r, e := s.GetByID(ctx, "", fixtureID, nil)
			return v, r, e
		}},
		"get-bad-base64": {item, "Get-CsOnlineAudioFile", `{"Content":"!!"}`, func(ctx context.Context, s *audio.Service) (any, *resty.Response, error) {
			v, r, e := s.GetByID(ctx, "", fixtureID, nil)
			return v, r, e
		}},
		"export-missing": {item + "/content", "Export-CsOnlineAudioFile", `{}`, func(ctx context.Context, s *audio.Service) (any, *resty.Response, error) {
			v, r, e := s.GetContentByID(ctx, "", fixtureID)
			return v, r, e
		}},
		"export-null": {item + "/content", "Export-CsOnlineAudioFile", `null`, func(ctx context.Context, s *audio.Service) (any, *resty.Response, error) {
			v, r, e := s.GetContentByID(ctx, "", fixtureID)
			return v, r, e
		}},
	} {
		run(t, testutil.Case{Name: name, Steps: []testutil.Exchange{discovery(t), wire("GET", step.path, step.command, nil, 200, "application/json", []byte(step.body))},
			Want: testutil.Expected{Status: 200, Error: client.ErrDecode},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				value, response, err := step.invoke(ctx, sdk.TeamsAdminAPI.AudioFiles)
				if value != nil && !isNilValue(value) {
					t.Fatal("malformed response returned a value")
				}
				return testutil.Observation{Response: response, Err: err}
			}})
	}
}

func isNilValue(v any) bool {
	switch x := v.(type) {
	case []*audio.AudioFile:
		return x == nil
	case *audio.AudioFile:
		return x == nil
	case []byte:
		return x == nil
	}
	return false
}

func TestUnitAudioFileExportAcceptsURLSafeUnpaddedBase64(t *testing.T) {
	t.Parallel()
	payload := []byte{0xfb, 0xff, 0xfe, 0x01}
	body := []byte(`{"Content":"` + base64.RawURLEncoding.EncodeToString(payload) + `"}`)
	run(t, testutil.Case{Name: "url-safe", Steps: []testutil.Exchange{discovery(t), wire("GET", "/Teams.MediaStorage/audiofile/HuntGroup/x/content", "Export-CsOnlineAudioFile", nil, 200, "application/json", body)},
		Want: testutil.Expected{Status: 200, Body: body},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.AudioFiles.GetContentByID(ctx, "HuntGroup", "x")
			if err != nil || !bytes.Equal(value, payload) {
				t.Fatal("URL-safe content was not decoded")
			}
			return testutil.Observation{Response: response, Err: err}
		}})
}

func TestUnitAudioFileInvalidInputFailsBeforeIO(t *testing.T) {
	t.Parallel()
	for name, call := range map[string]func(context.Context, *audio.Service) error{
		"get-empty-id": func(ctx context.Context, s *audio.Service) error { _, _, e := s.GetByID(ctx, "", "", nil); return e },
		"get-dot":      func(ctx context.Context, s *audio.Service) error { _, _, e := s.GetByID(ctx, "", "..", nil); return e },
		"get-slash":    func(ctx context.Context, s *audio.Service) error { _, _, e := s.GetByID(ctx, "", "a/b", nil); return e },
		"get-control": func(ctx context.Context, s *audio.Service) error {
			_, _, e := s.GetByID(ctx, "", "a\nb", nil)
			return e
		},
		"list-bad-app":     func(ctx context.Context, s *audio.Service) error { _, _, e := s.List(ctx, "a\\b"); return e },
		"list-blank-app":   func(ctx context.Context, s *audio.Service) error { _, _, e := s.List(ctx, "  "); return e },
		"export-empty":     func(ctx context.Context, s *audio.Service) error { _, _, e := s.GetContentByID(ctx, "", ""); return e },
		"delete-traversal": func(ctx context.Context, s *audio.Service) error { _, e := s.DeleteByID(ctx, "."); return e },
		"import-no-name": func(ctx context.Context, s *audio.Service) error {
			_, _, e := s.Create(ctx, audio.ImportRequest{Content: []byte{1}})
			return e
		},
		"import-empty": func(ctx context.Context, s *audio.Service) error {
			_, _, e := s.Create(ctx, audio.ImportRequest{FileName: "a.wav"})
			return e
		},
		"import-control-name": func(ctx context.Context, s *audio.Service) error {
			_, _, e := s.Create(ctx, audio.ImportRequest{FileName: "a\r.wav", Content: []byte{1}})
			return e
		},
	} {
		run(t, testutil.Case{Name: name, Want: testutil.Expected{Error: audio.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				return testutil.Observation{Err: call(ctx, sdk.TeamsAdminAPI.AudioFiles)}
			}})
	}
}

func TestUnitAudioFileNilContextAndServiceAreRejected(t *testing.T) {
	t.Parallel()
	var s *audio.Service
	//nolint:staticcheck // A nil context is the guarded input.
	if _, _, err := s.List(nil, ""); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, err := s.DeleteByID(t.Context(), "x"); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
	if got := (audio.AudioFile{Content: []byte("secret")}).String(); got != "audio file (redacted)" {
		t.Fatal("audio file formatting is not redacted")
	}
}
