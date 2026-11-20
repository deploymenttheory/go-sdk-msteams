package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
)

var (
	ErrPagination       = errors.New("invalid pagination configuration or continuation")
	ErrContinuationLoop = errors.New("pagination continuation repeated")
	ErrPageLimit        = errors.New("pagination page limit exceeded")
	ErrItemLimit        = errors.New("pagination aggregate item limit exceeded")
)

// PagingStyle distinguishes opaque full URLs from opaque query tokens and numeric
// offsets. A field named nextLink does not establish which protocol a backend uses.
type PagingStyle uint8

const (
	LinkPaging PagingStyle = iota + 1
	TokenPaging
	OffsetPaging
)

// PagingOptions bounds traversal. Zero limits select 1,000 pages and 100,000
// aggregate items. Parameter is the exact query key for token/offset protocols.
// Offset traversal increments by the actual returned item count and ends on an
// empty page; it does not mistake a short server-capped page for the final page.
type PagingOptions struct {
	Style     PagingStyle
	Parameter string
	MaxPages  int
	MaxItems  int
}

// PageContent is the decoder's normalized view of one evidenced backend envelope.
// Next contains a full URL for LinkPaging or an opaque token for TokenPaging. Offset
// decoders leave it empty; offsets advance using the returned item count.
type PageContent[T any] struct {
	Items []T
	Next  string
}

// PageDecoder must decode a complete page without retaining/mutating the body.
// Decoder diagnostics are replaced with ErrDecode because they may contain payloads.
type PageDecoder[T any] func([]byte) (PageContent[T], error)

// Page exposes data and buffered response metadata for one successfully decoded
// response. Raw response metadata may contain tenant data. HasNext does not imply
// that a subsequent page is authorized, available or within traversal limits.
type Page[T any] struct {
	Items    []T
	Response *resty.Response
	HasNext  bool
}

// ListResult exposes partial items and the last HTTP response even when traversal
// fails. A non-nil error always means the result must not be treated as complete.
type ListResult[T any] struct {
	Items    []T
	Response *resty.Response
	Pages    int
}

// Pager is a serial, lazy traversal bound to one scope. It owns copies of the initial
// query, headers, path and serialized body. Consume it from one goroutine at a time;
// separate pagers are independent. It stores no caller context and starts no worker.
type Pager[T any] struct {
	scope   *Scope
	request Request
	decoder PageDecoder[T]
	options PagingOptions
	seen    map[string]struct{}
	offset  int64
	pages   int
	done    bool
	failure error
	last    *resty.Response
}

// NewPager validates a read traversal without acquiring a token. Only explicit
// replay-safe GET/POST operations are accepted. Backend packages provide decoders
// and exact protocol mappings; resource services remain responsible for their route.
func NewPager[T any](scope *Scope, request Request, decoder PageDecoder[T], options PagingOptions) (*Pager[T], error) {
	if scope == nil || scope.session == nil || decoder == nil || request.Replay != ReplaySafe || (request.Method != http.MethodGet && request.Method != http.MethodPost) {
		return nil, ErrPagination
	}
	if options.MaxPages == 0 {
		options.MaxPages = 1000
	}
	if options.MaxItems == 0 {
		options.MaxItems = 100000
	}
	if options.MaxPages < 1 || options.MaxItems < 1 {
		return nil, ErrPagination
	}
	switch options.Style {
	case LinkPaging:
		if options.Parameter != "" {
			return nil, ErrPagination
		}
	case TokenPaging, OffsetPaging:
		if strings.TrimSpace(options.Parameter) == "" || strings.ContainsAny(options.Parameter, "\r\n") {
			return nil, ErrPagination
		}
	default:
		return nil, ErrPagination
	}
	profile, err := scope.profile()
	if err != nil {
		return nil, err
	}
	destination, err := profile.destination(request)
	if err != nil {
		return nil, err
	}
	headers, err := requestHeaders(request.Header)
	if err != nil {
		return nil, err
	}
	request.Header = headers
	request.URL = destination.String()
	request.Path = nil
	request.Query = nil
	if request.Body != nil {
		if _, stream := request.Body.(io.Reader); stream {
			return nil, ErrRequest
		}
		body, err := json.Marshal(request.Body)
		if err != nil {
			return nil, ErrRequest
		}
		request.Body = json.RawMessage(slices.Clone(body))
	}
	pager := &Pager[T]{scope: scope, request: request, decoder: decoder, options: options, seen: make(map[string]struct{})}
	if options.Style != LinkPaging {
		values := destination.Query()[options.Parameter]
		if len(values) > 1 {
			return nil, ErrPagination
		}
		if options.Style == OffsetPaging && len(values) == 1 {
			offset, err := strconv.ParseInt(values[0], 10, 64)
			if err != nil || offset < 0 {
				return nil, ErrPagination
			}
			pager.offset = offset
		}
	}
	return pager, nil
}

func (*Pager[T]) String() string     { return "msteams pager (redacted)" }
func (p *Pager[T]) GoString() string { return p.String() }

// Next returns (page, true, nil) for every decoded page, including empty intermediate
// and final pages. Completion is (zero, false, nil). Transport/decoder/continuation
// failures are terminal, retaining any last response but never reporting completion.
func (p *Pager[T]) Next(ctx context.Context) (Page[T], bool, error) {
	if ctx == nil {
		return Page[T]{}, false, ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return Page[T]{}, false, err
	}
	if p == nil || p.scope == nil {
		return Page[T]{}, false, ErrPagination
	}
	if p.failure != nil {
		return Page[T]{Response: p.last}, false, p.failure
	}
	if p.done {
		return Page[T]{}, false, nil
	}
	if p.pages >= p.options.MaxPages {
		return p.fail(ErrPageLimit)
	}
	key, err := p.continuationKey()
	if err != nil {
		return p.fail(err)
	}
	if _, exists := p.seen[key]; exists {
		return p.fail(ErrContinuationLoop)
	}
	// Scope.Do validates every destination before obtaining/attaching credentials.
	response, err := p.scope.Do(ctx, p.request, nil)
	if response != nil {
		p.last = response
	}
	if err != nil {
		return p.fail(err)
	}
	content, err := p.decoder(response.Body())
	if err != nil {
		return p.fail(ErrDecode)
	}
	p.pages++
	p.seen[key] = struct{}{}
	page := Page[T]{Items: content.Items, Response: response}
	if p.options.Style == OffsetPaging {
		if content.Next != "" {
			return p.fail(ErrPagination)
		}
		if len(content.Items) == 0 {
			p.done = true
			return page, true, nil
		}
		if int64(len(content.Items)) > int64(1<<63-1)-p.offset {
			return p.fail(ErrPagination)
		}
		p.offset += int64(len(content.Items))
		p.setQuery(strconv.FormatInt(p.offset, 10))
	} else {
		if content.Next == "" {
			p.done = true
			return page, true, nil
		}
		if p.options.Style == LinkPaging {
			p.request.URL = content.Next
		} else {
			p.setQuery(content.Next)
		}
	}
	page.HasNext = true
	return page, true, nil
}
func (p *Pager[T]) fail(err error) (Page[T], bool, error) {
	p.failure = err
	return Page[T]{Response: p.last}, false, err
}
func (p *Pager[T]) setQuery(value string) {
	// The current URL was validated on construction or by Scope.Do. Token/offset
	// traversal never accepts a server-provided URL and changes only its query key.
	u, _ := url.Parse(p.request.URL)
	values := u.Query()
	values.Set(p.options.Parameter, value)
	u.RawQuery = values.Encode()
	p.request.URL = u.String()
}
func (p *Pager[T]) continuationKey() (string, error) {
	profile, err := p.scope.profile()
	if err != nil {
		return "", err
	}
	destination, err := profile.destination(p.request)
	if err != nil {
		return "", err
	}
	if p.options.Style == TokenPaging {
		return destination.Query().Get(p.options.Parameter), nil
	}
	if p.options.Style == OffsetPaging {
		return strconv.FormatInt(p.offset, 10), nil
	}
	// Canonicalize only the loop-detection key. The opaque wire URL is unchanged.
	destination.RawQuery = destination.Query().Encode()
	return destination.String(), nil
}

// Collect consumes the remaining traversal. On any error it returns the collected
// prefix and last response; it never labels a truncated list as a successful result.
// If the next whole page exceeds MaxItems, that page isn't appended and ErrItemLimit
// is returned. Use Next instead when page-wise processing avoids aggregation limits.
func (p *Pager[T]) Collect(ctx context.Context) (ListResult[T], error) {
	var result ListResult[T]
	for {
		page, ok, err := p.Next(ctx)
		if page.Response != nil {
			result.Response = page.Response
		}
		if err != nil {
			return result, err
		}
		if !ok {
			return result, nil
		}
		result.Pages++
		if len(page.Items) > p.options.MaxItems-len(result.Items) {
			p.failure = ErrItemLimit
			return result, ErrItemLimit
		}
		result.Items = append(result.Items, page.Items...)
	}
}

// DecodeCollection decodes exact JSON field names. Empty itemsField selects a bare
// array; otherwise the array field must be present. Missing/null arrays fail rather
// than turning an unexpected response into truncated success. nextField may be empty
// for offset pages; absent/null continuations mean completion for token/link pages.
func DecodeCollection[T any](body []byte, itemsField, nextField string) (PageContent[T], error) {
	var result PageContent[T]
	rawItems := json.RawMessage(body)
	if itemsField != "" {
		var envelope map[string]json.RawMessage
		if json.Unmarshal(body, &envelope) != nil {
			return result, ErrDecode
		}
		var found bool
		rawItems, found = envelope[itemsField]
		if !found {
			return result, ErrDecode
		}
		if nextField != "" {
			if raw, exists := envelope[nextField]; exists && string(raw) != "null" {
				if json.Unmarshal(raw, &result.Next) != nil {
					return result, ErrDecode
				}
			}
		}
	} else if nextField != "" {
		return result, ErrPagination
	}
	trimmed := bytes.TrimSpace(rawItems)
	if len(trimmed) == 0 || trimmed[0] != '[' || json.Unmarshal(rawItems, &result.Items) != nil {
		return PageContent[T]{}, ErrDecode
	}
	return result, nil
}
