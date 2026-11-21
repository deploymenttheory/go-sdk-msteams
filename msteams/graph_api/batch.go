package graphapi

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	graphbatch "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/batch"
	"github.com/go-resty/resty/v2"
)

var (
	ErrBatchRequest  = graphbatch.ErrRequest
	ErrBatchResponse = graphbatch.ErrResponse
	ErrBatchItems    = graphbatch.ErrItems
)

// BatchRequest is one Graph JSON batch operation. The alias retains the existing
// API while resource packages can use graphbatch without importing this collection.
type BatchRequest = graphbatch.Request

// BatchResponse retains an individual Graph batch outcome, including its opaque
// body and headers. It is the same type exposed by the shared batch package.
type BatchResponse = graphbatch.Response

// ExecuteBatch sends 1–20 operations once through the shared Graph batch engine.
// It retains the existing API, wire behavior, partial results and error identities.
// Resource services own operation permissions and any safe item-specific retry.
func ExecuteBatch(ctx context.Context, scope *client.Scope, version string, requests []BatchRequest) ([]BatchResponse, *resty.Response, error) {
	return graphbatch.Execute(ctx, scope, version, requests)
}
