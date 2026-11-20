package teamsadminapi

import "github.com/deploymenttheory/go-sdk-msteams/msteams/client"

// TokenPageFormat maps an evidenced Teams envelope to its exact query parameter.
// For example, the pinned UserApiSearchResults uses users/@nextLink and $skiptoken;
// @nextLink is an opaque token here, not Graph's absolute @odata.nextLink URL.
type TokenPageFormat struct{ ItemsField, TokenField, QueryParameter string }

// NewTokenPager preserves the initial query/body/headers and changes only the
// configured query token. The caller must classify the read operation ReplaySafe.
func NewTokenPager[T any](scope *client.Scope, request client.Request, format TokenPageFormat, options client.PagingOptions) (*client.Pager[T], error) {
	if scope == nil || scope.Backend() != client.TeamsAdmin || format.ItemsField == "" || format.TokenField == "" {
		return nil, client.ErrPagination
	}
	options.Style = client.TokenPaging
	options.Parameter = format.QueryParameter
	return client.NewPager(scope, request, func(body []byte) (client.PageContent[T], error) {
		return client.DecodeCollection[T](body, format.ItemsField, format.TokenField)
	}, options)
}

// NewOffsetPager traverses an evidenced offset list. itemsField is empty for bare
// arrays; offsetParameter is case-sensitive (for example Skip or skip). Initial
// page-size/filter parameters are retained. Short pages continue; an empty page ends
// the traversal. The caller must classify the read operation ReplaySafe.
func NewOffsetPager[T any](scope *client.Scope, request client.Request, itemsField, offsetParameter string, options client.PagingOptions) (*client.Pager[T], error) {
	if scope == nil || scope.Backend() != client.TeamsAdmin {
		return nil, client.ErrPagination
	}
	options.Style = client.OffsetPaging
	options.Parameter = offsetParameter
	return client.NewPager(scope, request, func(body []byte) (client.PageContent[T], error) {
		return client.DecodeCollection[T](body, itemsField, "")
	}, options)
}
