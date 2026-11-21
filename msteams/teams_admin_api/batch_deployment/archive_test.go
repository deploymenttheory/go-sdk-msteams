package batchdeployment_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	deployment "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/batch_deployment"
)

var budget = deployment.ArchiveOptions{MaxInputBytes: 4096, MaxArchiveBytes: 8192}

type trackedCSV struct {
	io.Reader
	closes int
}

func (r *trackedCSV) Close() error { r.closes++; return nil }

type readFunc func([]byte) (int, error)

func (f readFunc) Read(p []byte) (int, error) { return f(p) }

func archiveContents(t *testing.T, value deployment.Archive) map[string]string {
	t.Helper()
	raw := value.Bytes()
	reader, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal("archive is not readable ZIP")
	}
	if len(reader.File) != 2 {
		t.Fatal("deployment ZIP entry count differs")
	}
	result := map[string]string{}
	for _, entry := range reader.File {
		if entry.Method != zip.Deflate || entry.Mode().Perm() != 0600 {
			t.Fatal("archive method/mode differs")
		}
		if _, duplicate := result[entry.Name]; duplicate {
			t.Fatal("duplicate ZIP entry")
		}
		body, err := entry.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, readErr := io.ReadAll(body)
		closeErr := body.Close()
		if readErr != nil || closeErr != nil {
			t.Fatal("ZIP entry could not be read/closed")
		}
		result[entry.Name] = string(data)
	}
	return result
}

func TestUnitArchivePreservesOriginalBytesFixedNamesAndReaderOwnership(t *testing.T) {
	t.Parallel()
	teamsText, usersText := "Team Name,Existing Team ID\r\nFixture,\r\n", "User Full Name,User UPN or ID\nFixture 😀,fixture@invalid\n"
	teams, users := &trackedCSV{Reader: strings.NewReader(teamsText)}, &trackedCSV{Reader: strings.NewReader(usersText)}
	value, err := deployment.BuildArchive(t.Context(), teams, users, budget)
	if err != nil {
		t.Fatal(err)
	}
	contents := archiveContents(t, value)
	if contents["Teams.csv"] != teamsText || contents["Users.csv"] != usersText || teams.closes != 0 || users.closes != 0 {
		t.Fatal("CSV bytes/names or borrowed reader ownership differ")
	}
	encoded := value.Encoded()
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || !bytes.Equal(decoded, value.Bytes()) || strings.ContainsAny(encoded, "\r\n") {
		t.Fatal("archive Base64 encoding differs")
	}
	copy := value.Bytes()
	copy[0] ^= 0xff
	if bytes.Equal(copy, value.Bytes()) {
		t.Fatal("caller byte mutation changed stored archive")
	}
	for _, format := range []string{"%v", "%+v", "%#v"} {
		printed := fmt.Sprintf(format, value)
		if strings.Contains(printed, "fixture@invalid") || strings.Contains(printed, encoded) {
			t.Fatal("ordinary formatting exposed archive data")
		}
	}
}

func TestUnitArchiveBudgetsApplyAcrossBothCSVsAndZIPFraming(t *testing.T) {
	t.Parallel()
	input := "original bytes"
	options := deployment.ArchiveOptions{MaxInputBytes: int64(2 * len(input)), MaxArchiveBytes: 8192}
	value, err := deployment.BuildArchive(t.Context(), strings.NewReader(input), strings.NewReader(input), options)
	if err != nil {
		t.Fatal(err)
	}
	options.MaxArchiveBytes = value.Size()
	if _, err := deployment.BuildArchive(t.Context(), strings.NewReader(input), strings.NewReader(input), options); err != nil {
		t.Fatal("exact archive budget failed")
	}
	options.MaxArchiveBytes--
	if result, err := deployment.BuildArchive(t.Context(), strings.NewReader(input), strings.NewReader(input), options); !errors.Is(err, deployment.ErrArchiveTooLarge) || result.Size() != 0 {
		t.Fatal("archive framing escaped its budget")
	}
	options = budget
	options.MaxInputBytes = int64(2*len(input) - 1)
	if result, err := deployment.BuildArchive(t.Context(), strings.NewReader(input), strings.NewReader(input), options); !errors.Is(err, deployment.ErrCSVTooLarge) || result.Size() != 0 {
		t.Fatal("combined CSV budget was not enforced")
	}
}

func TestUnitArchiveRejectsInvalidAndFailedSourcesWithoutReturningPartialData(t *testing.T) {
	t.Parallel()
	var missing *trackedCSV
	for _, tc := range []struct {
		name    string
		reader  io.Reader
		options deployment.ArchiveOptions
		err     error
	}{
		{"nil", nil, budget, deployment.ErrArchiveInput},
		{"typed-nil", missing, budget, deployment.ErrArchiveInput},
		{"zero-input-budget", strings.NewReader(""), deployment.ArchiveOptions{MaxArchiveBytes: 8192}, deployment.ErrArchiveInput},
		{"zero-archive-budget", strings.NewReader(""), deployment.ArchiveOptions{MaxInputBytes: 4096}, deployment.ErrArchiveInput},
		{"source-failure", readFunc(func([]byte) (int, error) { return 0, errors.New("private-reader-canary") }), budget, deployment.ErrCSVRead},
		{"no-progress", readFunc(func([]byte) (int, error) { return 0, nil }), budget, deployment.ErrCSVRead},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value, err := deployment.BuildArchive(t.Context(), tc.reader, strings.NewReader("users"), tc.options)
			if !errors.Is(err, tc.err) || value.Size() != 0 || strings.Contains(err.Error(), "private-reader-canary") {
				t.Fatal("invalid archive source/budget was accepted or leaked")
			}
		})
	}
}

func TestUnitArchiveCancellationStopsBeforeOrDuringInput(t *testing.T) {
	t.Parallel()
	ctx, stop := context.WithCancel(t.Context())
	stop()
	reads := 0
	source := readFunc(func([]byte) (int, error) { reads++; return 0, io.EOF })
	if _, err := deployment.BuildArchive(ctx, source, source, budget); !errors.Is(err, context.Canceled) || reads != 0 {
		t.Fatal("canceled builder read input")
	}
	ctx, stop = context.WithCancel(t.Context())
	defer stop()
	source = readFunc(func([]byte) (int, error) { stop(); return 0, context.Canceled })
	if _, err := deployment.BuildArchive(ctx, source, strings.NewReader("users"), budget); !errors.Is(err, context.Canceled) {
		t.Fatal("input cancellation lost caller classification")
	}
	var missing context.Context
	if _, err := deployment.BuildArchive(missing, strings.NewReader(""), strings.NewReader(""), budget); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("missing context was accepted")
	}
}

func TestUnitArchiveConstructionIsIsolatedAndDoesNotParseCSV(t *testing.T) {
	t.Parallel()
	var group sync.WaitGroup
	type result struct {
		input   string
		archive deployment.Archive
		err     error
	}
	results := make(chan result, 16)
	for i := range 16 {
		group.Go(func() {
			input := fmt.Sprintf("not parsed as CSV; fixture %d\x00", i)
			value, err := deployment.BuildArchive(t.Context(), strings.NewReader(input), strings.NewReader(""), budget)
			results <- result{input: input, archive: value, err: err}
		})
	}
	group.Wait()
	close(results)
	for value := range results {
		if value.err != nil {
			t.Error(value.err)
			continue
		}
		contents := archiveContents(t, value.archive)
		if contents["Teams.csv"] != value.input || contents["Users.csv"] != "" {
			t.Error("concurrent archives shared or normalized input")
		}
	}
}
