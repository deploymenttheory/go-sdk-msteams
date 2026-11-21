package targetinghierarchy

import (
	"encoding/json"
	"errors"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func checkProjection(t *testing.T, want []byte, value any) {
	t.Helper()
	got, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var a, b any
	if json.Unmarshal(want, &a) != nil || json.Unmarshal(got, &b) != nil || !reflect.DeepEqual(a, b) {
		t.Fatal("local status projection differs")
	}
}
func TestUnitStatusProjectionFromLocalOperationValues(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, input, want string }{
		{"hierarchy-status-projection-NotStarted", "{\"id\":\"fixture-operation\",\"status\":\"NotStarted\",\"lastActionDateTime\":\"2026-01-01T00:00:00Z\",\"fileName\":\"fixture.csv\"}", "{\"fileName\":\"fixture.csv\",\"id\":\"fixture-operation\",\"lastUpdatedAt\":\"2026-01-01T00:00:00Z\",\"status\":\"Starting\"}"},
		{"hierarchy-status-projection-Running", "{\"id\":\"fixture-operation\",\"status\":\"Running\",\"lastActionDateTime\":\"2026-01-01T00:00:00Z\",\"fileName\":\"fixture.csv\"}", "{\"fileName\":\"fixture.csv\",\"id\":\"fixture-operation\",\"lastUpdatedAt\":\"2026-01-01T00:00:00Z\",\"status\":\"Ingesting\"}"},
		{"hierarchy-status-projection-Succeeded", "{\"id\":\"fixture-operation\",\"status\":\"Succeeded\",\"lastActionDateTime\":\"2026-01-01T00:00:00Z\",\"fileName\":\"fixture.csv\"}", "{\"fileName\":\"fixture.csv\",\"id\":\"fixture-operation\",\"lastUpdatedAt\":\"2026-01-01T00:00:00Z\",\"status\":\"Successful\"}"},
		{"hierarchy-status-projection-Failed", "{\"id\":\"fixture-operation\",\"status\":\"Failed\",\"lastActionDateTime\":\"2026-01-01T00:00:00Z\",\"fileName\":\"fixture.csv\"}", "{\"fileName\":\"fixture.csv\",\"id\":\"fixture-operation\",\"lastUpdatedAt\":\"2026-01-01T00:00:00Z\",\"status\":\"Failed\"}"},
		{"hierarchy-status-projection-UnknownFutureValue", "{\"id\":\"fixture-operation\",\"status\":\"UnknownFutureValue\",\"lastActionDateTime\":\"2026-01-01T00:00:00Z\",\"fileName\":\"fixture.csv\"}", "{\"fileName\":\"fixture.csv\",\"id\":\"fixture-operation\",\"lastUpdatedAt\":\"2026-01-01T00:00:00Z\",\"status\":\"Unknown\"}"},
		{"hierarchy-status-projection-FutureStatus", "{\"id\":\"fixture-operation\",\"status\":\"FutureStatus\",\"lastActionDateTime\":\"2026-01-01T00:00:00Z\",\"fileName\":\"fixture.csv\"}", "{\"fileName\":\"fixture.csv\",\"id\":\"fixture-operation\",\"lastUpdatedAt\":\"2026-01-01T00:00:00Z\",\"status\":\"FutureStatus\"}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var value HierarchyOperation
			if json.Unmarshal([]byte(tc.input), &value) != nil {
				t.Fatal("invalid local operation vector")
			}
			projected, err := projectHierarchy(&value)
			if err != nil {
				t.Fatal(err)
			}
			checkProjection(t, []byte(tc.want), map[string]any{"id": projected.ID, "status": projected.Status, "lastUpdatedAt": projected.LastUpdatedAt, "fileName": projected.Filename})
		})
	}
}
func TestUnitStatusErrorProjectionFromLocalOperationValues(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, version, input, want string }{
		{"retail-ordered-error-arguments", "1", "{\"id\":\"fixture-operation\",\"status\":\"Failed\",\"lastUpdatedAt\":\"wire-time\",\"errorDetails\":{\"lastKnownStatus\":\"Ingesting\",\"errorMessage\":\"[{\\\"errorCode\\\":\\\"First\\\",\\\"description\\\":\\\"bad {first}/{second}\\\",\\\"args\\\":[{\\\"second\\\":\\\"B\\\",\\\"first\\\":\\\"A\\\"}]},{\\\"errorCode\\\":\\\"First\\\",\\\"description\\\":\\\"extra {x}\\\",\\\"args\\\":[{\\\"x\\\":\\\"X\\\"}]}]\"}}", "{\"status\":\"Failed\",\"filename\":null,\"id\":\"fixture-operation\",\"last_known_status\":\"Ingesting\",\"error_message\":\"2 error(s) were found.\\r\\nError: First\\nDescriptions:\\nbad {first}/B\\nbad A/{second}\\nError: First\\nDescriptions:\\nextra X\\n\",\"last_updated_at\":\"wire-time\",\"last_modified_by\":null}"},
		{"retail-empty-error-text", "1", "{\"id\":\"fixture-operation\",\"status\":\"Ingesting\",\"errorDetails\":{\"lastKnownStatus\":\"Starting\",\"errorMessage\":\"\"}}", "{\"status\":\"Ingesting\",\"filename\":null,\"id\":\"fixture-operation\",\"last_known_status\":\"Starting\",\"error_message\":null,\"last_updated_at\":null,\"last_modified_by\":null}"},
		{"hierarchy-grouped-ordered-errors", "2", "{\"id\":\"fixture-operation\",\"status\":\"Failed\",\"error\":{\"message\":\"native top-level message\",\"details\":[{\"code\":\"A\",\"message\":\"bad {a}/{b}\",\"arguments\":{\"{a}\":\"{b}\",\"{b}\":\"B\"}},{\"code\":\"B\",\"message\":\"other\",\"arguments\":{}},{\"code\":\"A\",\"message\":\"last {x}\",\"arguments\":{\"{x}\":\"X\"}}]}}", "{\"status\":\"Failed\",\"filename\":null,\"id\":\"fixture-operation\",\"last_known_status\":null,\"error_message\":\"3 error(s) were found.\\r\\nError: A\\nDescriptions:\\nbad B/B\\nlast X\\nError: B\\nDescriptions:\\nother\\n\",\"last_updated_at\":null,\"last_modified_by\":null}"},
		{"hierarchy-no-error-details", "2", "{\"id\":\"fixture-operation\",\"status\":\"Succeeded\",\"error\":{\"message\":\"native top-level message\",\"details\":[]}}", "{\"status\":\"Successful\",\"filename\":null,\"id\":\"fixture-operation\",\"last_known_status\":null,\"error_message\":null,\"last_updated_at\":null,\"last_modified_by\":null}"},
		{"hierarchy-numeric-status", "2", "{\"id\":\"fixture-operation\",\"status\":\"1\"}", "{\"status\":\"Ingesting\",\"filename\":null,\"id\":\"fixture-operation\",\"last_known_status\":null,\"error_message\":null,\"last_updated_at\":null,\"last_modified_by\":null}"},
		{"hierarchy-status-case", "2", "{\"id\":\"fixture-operation\",\"status\":\"running\"}", "{\"status\":\"running\",\"filename\":null,\"id\":\"fixture-operation\",\"last_known_status\":null,\"error_message\":null,\"last_updated_at\":null,\"last_modified_by\":null}"},
		{"hierarchy-combined-enum-status", "2", "{\"id\":\"fixture-operation\",\"status\":\"Running, Succeeded\"}", "{\"status\":\"Failed\",\"filename\":null,\"id\":\"fixture-operation\",\"last_known_status\":null,\"error_message\":null,\"last_updated_at\":null,\"last_modified_by\":null}"},
		{"hierarchy-status-whitespace", "2", "{\"id\":\"fixture-operation\",\"status\":\"  Running  \"}", "{\"status\":\"Ingesting\",\"filename\":null,\"id\":\"fixture-operation\",\"last_known_status\":null,\"error_message\":null,\"last_updated_at\":null,\"last_modified_by\":null}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var status *Status
			var err error
			if tc.version == "1" {
				var value RetailOperation
				if json.Unmarshal([]byte(tc.input), &value) != nil {
					t.Fatal("invalid retail vector")
				}
				status, err = projectRetail(&value)
			} else {
				var value HierarchyOperation
				if json.Unmarshal([]byte(tc.input), &value) != nil {
					t.Fatal("invalid hierarchy vector")
				}
				status, err = projectHierarchy(&value)
			}
			if err != nil {
				t.Fatal(err)
			}
			var want map[string]any
			if json.Unmarshal([]byte(tc.want), &want) != nil {
				t.Fatal("invalid expected projection")
			}
			if message, ok := want["error_message"].(string); ok {
				head, body, ok := strings.Cut(message, "\r\n")
				if !ok {
					t.Fatal("missing error header")
				}
				body = strings.ReplaceAll(body, "\r\n", "\n")
				if runtime.GOOS == "windows" {
					body = strings.ReplaceAll(body, "\n", "\r\n")
				}
				want["error_message"] = head + "\r\n" + body
			}
			encoded, _ := json.Marshal(want)
			checkProjection(t, encoded, map[string]any{"id": status.ID, "status": status.Status, "last_known_status": status.LastKnownStatus, "error_message": status.ErrorMessage, "last_updated_at": status.LastUpdatedAt, "last_modified_by": status.LastModifiedBy, "filename": status.Filename})
		})
	}
}
func TestUnitMalformedNestedErrorsRetainLocalOperation(t *testing.T) {
	t.Parallel()
	var retail RetailOperation
	if json.Unmarshal([]byte(`{"id":"operation","status":"Failed","errorDetails":{"lastKnownStatus":"Ingesting","errorMessage":"not-json"}}`), &retail) != nil {
		t.Fatal("invalid local retail vector")
	}
	v, err := projectRetail(&retail)
	if !errors.Is(err, client.ErrDecode) || v == nil || v.Retail != &retail {
		t.Fatal("malformed error lost readable operation")
	}
	var hierarchy HierarchyOperation
	if json.Unmarshal([]byte(`{"id":"operation","future":5}`), &hierarchy) != nil {
		t.Fatal("invalid local hierarchy vector")
	}
	v, err = projectHierarchy(&hierarchy)
	if !errors.Is(err, client.ErrDecode) || v == nil || v.Hierarchy != &hierarchy {
		t.Fatal("missing status lost readable operation")
	}
}
