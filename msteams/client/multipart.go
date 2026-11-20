package client

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"reflect"
	"strings"
	"unicode"

	"github.com/go-resty/resty/v2"
)

// MultipartPart describes one ordered form-data part. Resource services supply
// evidenced field names and content types; no backend field names are invented.
// An empty Filename selects a plain form field. Body follows Upload ownership;
// each part must own an independent source (no aliased wrappers).
type MultipartPart struct {
	Name        string
	Filename    string
	ContentType string
	Body        io.ReadCloser
}

// UploadMultipart streams ordered form-data parts and owns every non-nil Body
// immediately, even when another part is invalid. maxBytes bounds the entire wire
// payload, including framing. Multipart headers are generated locally. It starts
// no producer goroutine and never buffers file contents or replays the request.
func (s *Scope) UploadMultipart(ctx context.Context, request Request, parts []MultipartPart, maxBytes int64, result any) (response *resty.Response, err error) {
	owned := make([]*ownedStream, len(parts))
	seen := make(map[io.ReadCloser]*ownedStream, len(parts))
	duplicate := false
	for i, part := range parts {
		if !nilStream(part.Body) {
			if reflect.ValueOf(part.Body).Comparable() {
				if previous, exists := seen[part.Body]; exists {
					owned[i] = previous
					duplicate = true
					continue
				}
			}
			owned[i] = &ownedStream{source: part.Body}
			if reflect.ValueOf(part.Body).Comparable() {
				seen[part.Body] = owned[i]
			}
		}
	}
	defer func() {
		for _, body := range owned {
			if body != nil && body.Close() != nil {
				err = errors.Join(err, ErrStreamClose)
			}
		}
	}()
	if len(parts) == 0 || len(parts) > 100 || duplicate {
		return nil, ErrRequest
	}
	var header bytes.Buffer
	writer := multipart.NewWriter(&header)
	readers := make([]io.Reader, 0, 2*len(parts)+1)
	for i, part := range parts {
		if owned[i] == nil || part.Name == "" || invalidMultipartText(part.Name) || invalidMultipartText(part.Filename) {
			return nil, ErrRequest
		}
		disposition := map[string]string{"name": part.Name}
		if part.Filename != "" {
			disposition["filename"] = part.Filename
		}
		h := textproto.MIMEHeader{"Content-Disposition": {mime.FormatMediaType("form-data", disposition)}}
		if part.ContentType != "" {
			if _, _, parseErr := mime.ParseMediaType(part.ContentType); parseErr != nil || invalidMultipartText(part.ContentType) {
				return nil, ErrRequest
			}
			h.Set("Content-Type", part.ContentType)
		}
		if _, err := writer.CreatePart(h); err != nil {
			return nil, ErrRequest
		}
		readers = append(readers, bytes.NewReader(bytes.Clone(header.Bytes())), owned[i])
		header.Reset()
	}
	if err := writer.Close(); err != nil {
		return nil, ErrRequest
	}
	readers = append(readers, bytes.NewReader(bytes.Clone(header.Bytes())))
	stream := &multipartStream{Reader: io.MultiReader(readers...), parts: owned}
	return s.Upload(ctx, request, stream, writer.FormDataContentType(), maxBytes, result)
}

func invalidMultipartText(value string) bool { return strings.ContainsFunc(value, unicode.IsControl) }

type multipartStream struct {
	io.Reader
	parts []*ownedStream
}

func (s *multipartStream) Close() error {
	var result error
	for _, part := range s.parts {
		if err := part.Close(); err != nil {
			result = errors.Join(result, ErrStreamClose)
		}
	}
	return result
}
