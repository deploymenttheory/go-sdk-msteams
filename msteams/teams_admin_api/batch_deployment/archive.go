// Package batchdeployment manages frontline batch deployment submissions and
// their opaque status reads. Archive construction is local and requires no SDK
// client, credentials, filesystem paths or tenant request.
package batchdeployment

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"reflect"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var (
	ErrArchiveInput    = errors.New("invalid deployment archive input")
	ErrCSVRead         = errors.New("deployment CSV source could not be read")
	ErrCSVTooLarge     = errors.New("deployment CSV input exceeds the configured combined byte limit")
	ErrArchiveTooLarge = errors.New("deployment archive exceeds the configured byte limit")
	ErrArchiveWrite    = errors.New("deployment archive could not be constructed")
)

// ArchiveOptions supplies explicit local budgets, not claimed service limits.
// MaxInputBytes bounds the combined original CSV bytes; MaxArchiveBytes includes
// ZIP framing/central directory. Both must be positive.
type ArchiveOptions struct {
	MaxInputBytes   int64
	MaxArchiveBytes int64
}

// Archive is a two-entry deployment ZIP built from original CSV bytes. Its zero
// value is invalid. Explicit Bytes/Encoded access returns data; ordinary formatting
// is redacted. There is no shared temporary file or arbitrary-entry constructor.
type Archive struct{ data []byte }

func (a Archive) Bytes() []byte    { return bytes.Clone(a.data) }
func (a Archive) Encoded() string  { return base64.StdEncoding.EncodeToString(a.data) }
func (a Archive) Size() int64      { return int64(len(a.data)) }
func (Archive) String() string     { return "deployment archive (redacted)" }
func (a Archive) GoString() string { return a.String() }

type archiveBuffer struct {
	bytes.Buffer
	limit int64
}

func (b *archiveBuffer) Write(value []byte) (int, error) {
	if int64(len(value)) > b.limit-int64(b.Len()) {
		return 0, ErrArchiveTooLarge
	}
	return b.Buffer.Write(value)
}
func nilReader(reader io.Reader) bool {
	if reader == nil {
		return true
	}
	value := reflect.ValueOf(reader)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	}
	return false
}

// BuildArchive constructs exactly Teams.csv and Users.csv using Deflate's
// fastest compression. Bytes are not parsed, normalized or merged with a prior
// archive. Readers remain caller-owned and must return promptly or honor ctx;
// this synchronous builder starts no goroutine and does not close them.
func BuildArchive(ctx context.Context, teams, users io.Reader, options ArchiveOptions) (Archive, error) {
	if ctx == nil {
		return Archive{}, client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return Archive{}, err
	}
	if nilReader(teams) || nilReader(users) || options.MaxInputBytes <= 0 || options.MaxArchiveBytes <= 0 {
		return Archive{}, ErrArchiveInput
	}
	buffer := &archiveBuffer{limit: options.MaxArchiveBytes}
	writer := zip.NewWriter(buffer)
	defer func() { _ = writer.Close() }()
	writer.RegisterCompressor(zip.Deflate, func(destination io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(destination, flate.BestSpeed)
	})
	remaining := options.MaxInputBytes
	for _, entry := range []struct {
		name   string
		reader io.Reader
	}{{"Teams.csv", teams}, {"Users.csv", users}} {
		if err := ctx.Err(); err != nil {
			return Archive{}, err
		}
		header := &zip.FileHeader{Name: entry.name, Method: zip.Deflate}
		header.SetMode(0600)
		part, err := writer.CreateHeader(header)
		if err != nil {
			return Archive{}, archiveError(err)
		}
		if err := copyCSV(ctx, part, entry.reader, &remaining); err != nil {
			return Archive{}, err
		}
	}
	if err := writer.Close(); err != nil {
		return Archive{}, archiveError(err)
	}
	if err := ctx.Err(); err != nil {
		return Archive{}, err
	}
	return Archive{data: bytes.Clone(buffer.Bytes())}, nil
}

func archiveError(err error) error {
	if errors.Is(err, ErrArchiveTooLarge) {
		return ErrArchiveTooLarge
	}
	return ErrArchiveWrite
}

func copyCSV(ctx context.Context, destination io.Writer, source io.Reader, remaining *int64) error {
	buffer := make([]byte, 32<<10)
	emptyReads := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		chunk := buffer
		if *remaining < int64(len(buffer)) {
			chunk = buffer[:int(*remaining)+1]
		}
		n, readErr := source.Read(chunk)
		if err := ctx.Err(); err != nil {
			return err
		}
		if n < 0 || n > len(chunk) {
			return ErrCSVRead
		}
		if n > 0 {
			emptyReads = 0
			if int64(n) > *remaining {
				return ErrCSVTooLarge
			}
			written, err := destination.Write(chunk[:n])
			if err != nil || written != n {
				return archiveError(err)
			}
			*remaining -= int64(n)
		} else {
			emptyReads++
		}
		if readErr == io.EOF {
			return nil
		}
		if readErr != nil {
			return ErrCSVRead
		}
		if emptyReads >= 100 {
			return ErrCSVRead
		}
	}
}
