package contracttest

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func multipartCase() Case {
	return Case{Name: "csv-upload", Family: "targeting", Scope: "primitive", TenantID: "11111111-1111-4111-8111-111111111111", Source: "internal/contracttest/multipart_test.go", Steps: []Exchange{{Request: WireRequest{Method: "POST", URL: "https://retailservices.teams.microsoft.com/api/rest/v1/nodes/upload", Audience: "https://retailservices.teams.microsoft.com", Multipart: []MultipartField{{Name: "csvFile", Filename: "fixture.csv", ContentType: "application/json", Body: []byte("a,b\r\n1,2\r\n")}}}, Reply: Reply{Status: 202}}}, Invoke: func(ctx context.Context, sdk *msteams.Client) Observation {
		scope, err := client.NewScope(client.Targeting, sdk.TeamsAdminAPI.Scope().Session())
		if err != nil {
			return Observation{Err: err}
		}
		response, err := scope.UploadMultipart(ctx, client.Request{Method: "POST", Path: []string{"api", "rest", "v1", "nodes", "upload"}}, []client.MultipartPart{{Name: "csvFile", Filename: "fixture.csv", ContentType: "application/json", Body: io.NopCloser(strings.NewReader("a,b\r\n1,2\r\n"))}}, 1024, nil)
		return Observation{Response: response, Err: err}
	}, Want: Expected{Status: 202}}
}

func TestUnitMultipartComparisonUsesPartsWithoutRetainingPayload(t *testing.T) {
	t.Parallel()
	requirePass(t, multipartCase())
	for _, name := range []string{"name", "filename", "content-type", "bytes", "extra-part"} {
		t.Run(name, func(t *testing.T) {
			c := multipartCase()
			field := &c.Steps[0].Request.Multipart[0]
			switch name {
			case "name":
				field.Name = "private-canary"
			case "filename":
				field.Filename = "private-canary.csv"
			case "content-type":
				field.ContentType = "text/csv"
			case "bytes":
				field.Body = []byte("private-canary")
			case "extra-part":
				c.Steps[0].Request.Multipart = append(c.Steps[0].Request.Multipart, MultipartField{Name: "extra"})
			}
			result, err := Run(t.Context(), c)
			if err != nil || result.Passed() || len(result.Differences()) == 0 {
				t.Fatal("multipart mismatch was accepted")
			}
			for _, difference := range result.Differences() {
				if strings.Contains(difference.Path, "private-canary") || strings.Contains(difference.Kind, "private-canary") {
					t.Fatal("multipart payload leaked into structural differences")
				}
			}
		})
	}
}

func TestUnitMultipartFixtureValidationRejectsContradictoryBodiesAndBindings(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"json-plus-multipart", "binary-plus-multipart", "empty-parts", "invalid-tenant"} {
		c := multipartCase()
		switch name {
		case "json-plus-multipart":
			c.Steps[0].Request.JSON = []byte(`{}`)
		case "binary-plus-multipart":
			c.Steps[0].Request.Binary = []byte{}
		case "empty-parts":
			c.Steps[0].Request.Multipart = []MultipartField{}
		case "invalid-tenant":
			c.TenantID = "../private-canary"
		}
		if _, err := Run(t.Context(), c); !errors.Is(err, ErrFixture) {
			t.Fatal("contradictory multipart fixture or binding accepted")
		}
	}
}

func TestUnitMultipartComparisonRejectsPathBearingFilenameAndMalformedFraming(t *testing.T) {
	t.Parallel()
	want := []MultipartField{{Name: "csvFile", Filename: "fixture.csv", ContentType: "application/json", Body: []byte("csv")}}
	raw := []byte("--fixture\r\nContent-Disposition: form-data; name=\"csvFile\"; filename=\"/private/fixture.csv\"\r\nContent-Type: application/json\r\n\r\ncsv\r\n--fixture--\r\n")
	if len(compareMultipart(want, "multipart/form-data; boundary=fixture", raw, "body")) == 0 {
		t.Fatal("directory-bearing filename was normalized away")
	}
	if len(compareMultipart(want, "application/json", raw, "body")) == 0 || len(compareMultipart(want, "multipart/form-data; boundary=fixture", []byte("broken"), "body")) == 0 {
		t.Fatal("invalid multipart framing was accepted")
	}
}
