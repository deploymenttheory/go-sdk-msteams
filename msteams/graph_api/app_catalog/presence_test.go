package appcatalog_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	ac "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/app_catalog"
)

func TestUnitAppCatalogSourceReadStatusPresenceAndControls(t *testing.T) {
	body := []byte(`{"VALUE":[null,{}, {"id":null},{"id":""},{"id":9007199254740993},null],"@ODATA.NEXTLINK":"\t\n"}`)
	for _, status := range []int{200, 201, 202, 206, 299} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			run(t, contracttest.Case{Name: fmt.Sprintf("source-read-%d", status), Mappings: mapping("Get-TeamsApp"), Steps: []contracttest.Exchange{get(route, status, body)}, Want: contracttest.Expected{Status: status}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				result, err := sdk.GraphAPI.AppCatalog.List(ctx, ac.ListOptions{})
				if err != nil || len(result.Items) != 6 || result.Items[0] != nil || result.Items[1] == nil || result.Items[1].ID != nil || result.Items[2] == nil || result.Items[2].ID != nil || result.Items[3] == nil || result.Items[3].ID == nil || *result.Items[3].ID != "" || result.Items[4] == nil || result.Items[4].ID == nil || *result.Items[4].ID != "9007199254740993" || result.Items[5] != nil {
					t.Fatal("source nullable output/receipt changed")
				}
				return contracttest.Observation{Response: result.Response, Err: err}
			}})
		})
	}
	for index, value := range []string{"\t", "\n", "\r\n", "a\nb"} {
		t.Run(fmt.Sprintf("control-%q", value), func(t *testing.T) {
			target := route
			if value == "a\nb" {
				target = filtered("displayName eq 'a\nb'")
			}
			run(t, contracttest.Case{Name: fmt.Sprintf("source-query-control-%d", index), Mappings: mapping("Get-TeamsApp"), Steps: []contracttest.Exchange{get(target, 200, []byte(`{"value":[]}`))}, Want: contracttest.Expected{Status: 200}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				result, err := sdk.GraphAPI.AppCatalog.List(ctx, ac.ListOptions{DisplayName: value})
				return contracttest.Observation{Response: result.Response, Err: err}
			}})
		})
	}
}

func TestUnitAppCatalogSourcePageAliasesAndPartialFailure(t *testing.T) {
	next := route + "?$skiptoken=source"
	first := []byte(`{"VALUE":[null,{}],"@ODATA.NEXTLINK":"` + next + `"}`)
	run(t, contracttest.Case{Name: "source-page-aliases", Mappings: mapping("Get-TeamsApp"), Steps: []contracttest.Exchange{get(route, 201, first), get(next, 206, []byte(`{"value":[{"id":false}]}`))}, Want: contracttest.Expected{Status: 206}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		result, err := sdk.GraphAPI.AppCatalog.List(ctx, ac.ListOptions{})
		if err != nil || result.Pages != 2 || len(result.Items) != 3 || result.Items[0] != nil || result.Items[1] == nil || result.Items[1].ID != nil || result.Items[2] == nil || result.Items[2].ID == nil || *result.Items[2].ID != "false" {
			t.Fatal("nullable page handoff changed")
		}
		return contracttest.Observation{Response: result.Response, Err: err}
	}})
	run(t, contracttest.Case{Name: "source-page-decode-partial-native-extension", Mappings: mapping("Get-TeamsApp"), Steps: []contracttest.Exchange{get(route, 202, first), get(next, 200, []byte(`{"value":null}`))}, Want: contracttest.Expected{Status: 200, Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		result, err := sdk.GraphAPI.AppCatalog.List(ctx, ac.ListOptions{})
		if !errors.Is(err, client.ErrDecode) || len(result.Items) != 2 || result.Items[0] != nil || result.Items[1] == nil || result.Response == nil {
			t.Fatal("partial native output or failed-page receipt lost")
		}
		return contracttest.Observation{Response: result.Response, Err: err}
	}})
}

func TestUnitAppCatalogAcceptedNullableCreateAndEmptyUploads(t *testing.T) {
	for _, item := range []struct {
		name   string
		status int
		body   string
		null   bool
	}{
		{"empty", 202, "", true}, {"whitespace", 200, " \t\n", true}, {"null", 206, "null", true}, {"missing-id", 201, "{}", false}, {"empty-id", 299, `{"id":""}`, false},
	} {
		t.Run(item.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: "source-create-" + item.name, Mappings: mapping("New-TeamsApp"), Steps: []contracttest.Exchange{upload(http.MethodPost, route, item.status, []byte(item.body))}, Want: contracttest.Expected{Status: item.status}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				app, response, err := sdk.GraphAPI.AppCatalog.Create(ctx, ac.CreateRequest{DistributionMethod: ac.OrganizationDistribution, Package: pkg})
				if err != nil || (app == nil) != item.null || response == nil {
					t.Fatal("accepted nullable creation output lost its receipt")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
		})
	}
	for _, method := range []string{http.MethodPost, http.MethodPut} {
		t.Run("empty-package-"+method, func(t *testing.T) {
			target := route
			command := "New-TeamsApp"
			if method == http.MethodPut {
				target += "/" + appID
				command = "Set-TeamsApp"
			}
			exchange := upload(method, target, 202, nil)
			exchange.Request.Binary = []byte{}
			run(t, contracttest.Case{Name: "source-empty-package-" + strings.ToLower(method), Mappings: mapping(command), Steps: []contracttest.Exchange{exchange}, Want: contracttest.Expected{Status: 202}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				if method == http.MethodPost {
					app, response, err := sdk.GraphAPI.AppCatalog.Create(ctx, ac.CreateRequest{DistributionMethod: ac.OrganizationDistribution, Package: []byte{}})
					if app != nil {
						t.Fatal("empty accepted body invented model")
					}
					return contracttest.Observation{Response: response, Err: err}
				}
				response, err := sdk.GraphAPI.AppCatalog.UpdateByID(ctx, appID, []byte{})
				return contracttest.Observation{Response: response, Err: err}
			}})
		})
	}
	for _, status := range []int{200, 202, 206, 299} {
		t.Run(fmt.Sprintf("delete-%d", status), func(t *testing.T) {
			run(t, contracttest.Case{Name: fmt.Sprintf("source-delete-%d", status), Mappings: mapping("Remove-TeamsApp"), Steps: []contracttest.Exchange{{Request: contracttest.WireRequest{Method: http.MethodDelete, URL: route + "/" + appID, Audience: auth.GraphAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: reply(status, nil)}}, Want: contracttest.Expected{Status: status}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				response, err := sdk.GraphAPI.AppCatalog.DeleteByID(ctx, appID)
				return contracttest.Observation{Response: response, Err: err}
			}})
		})
	}
}
