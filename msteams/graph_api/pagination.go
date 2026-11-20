package graphapi

import (
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// NewPager traverses a Graph value/@odata.nextLink collection. The complete next
// URL remains opaque and original custom headers are retained on every page. Graph
// scope validation prevents a continuation from switching authority or credentials.
func NewPager[T any](scope *client.Scope, request client.Request, options client.PagingOptions) (*client.Pager[T], error) {
	if scope == nil || scope.Backend() != client.Graph || request.Method != http.MethodGet {
		return nil, client.ErrPagination
	}
	request.Replay = client.ReplaySafe
	options.Style = client.LinkPaging
	options.Parameter = ""
	return client.NewPager(scope, request, func(body []byte) (client.PageContent[T], error) {
		return client.DecodeCollection[T](body, "value", "@odata.nextLink")
	}, options)
}
