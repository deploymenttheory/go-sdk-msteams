package lab_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
)

type recoveryBackend struct {
	resources   map[string]string
	settings    string
	failRestore bool
	ambiguous   bool
	trace       []string
}

func (b *recoveryBackend) Resolve(ctx context.Context, e lab.Entry) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	b.trace = append(b.trace, "resolve")
	if b.ambiguous {
		return "", lab.ErrUnresolved
	}
	return b.resources[e.Fixture], nil
}
func (b *recoveryBackend) Delete(ctx context.Context, e lab.Entry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b.trace = append(b.trace, "delete")
	for name, id := range b.resources {
		if id == e.ResourceID {
			delete(b.resources, name)
		}
	}
	return nil
}
func (b *recoveryBackend) Restore(ctx context.Context, e lab.Entry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b.trace = append(b.trace, "restore")
	if b.failRestore {
		return errors.New("private remote diagnostic")
	}
	var snapshot struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(e.Snapshot, &snapshot); err != nil {
		return err
	}
	b.settings = snapshot.Value
	return nil
}

func closeJournal(t *testing.T, j *lab.Journal) {
	t.Helper()
	t.Cleanup(func() {
		if err := j.Close(); err != nil {
			t.Error(err)
		}
	})
}

func TestUnitJournalRecoversInterruptedCreationAndSharedSettings(t *testing.T) {
	t.Parallel()
	directory := filepath.Join(t.TempDir(), "private")
	j, err := lab.NewJournal(directory, "fixture-tenant/cloud/identity")
	if err != nil {
		t.Fatal(err)
	}
	closeJournal(t, j)
	name, err := j.FixtureName("team")
	if err != nil {
		t.Fatal(err)
	}
	createID, err := j.PlanCreate("team", name)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(directory, j.FileName())
	// The durable intent exists before the remote creation. Its response will be lost.
	//nolint:gosec // G304: test-owned journal under t.TempDir, never external input.
	raw, err := os.ReadFile(file)
	if err != nil || !strings.Contains(string(raw), name) {
		t.Fatal("creation intent was not durable before send")
	}
	backend := &recoveryBackend{resources: map[string]string{name: "private-created-id"}, settings: "original"}
	if createID == "" {
		t.Fatal("missing creation identity")
	}
	snapshot := json.RawMessage(`{"value":"original"}`)
	restoreID, err := j.PlanRestore("settings", "private-shared-id", snapshot)
	if err != nil {
		t.Fatal(err)
	}
	copy(snapshot, []byte(`{"value":"modified"}`))
	backend.settings = "changed"
	if err := j.Applied(restoreID, "private-shared-id"); err != nil {
		t.Fatal(err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := lab.OpenJournal(file, "different-tenant"); !errors.Is(err, lab.ErrJournal) {
		t.Fatal("cross-tenant recovery accepted")
	}
	reopened, err := lab.OpenJournal(file, "fixture-tenant/cloud/identity")
	if err != nil {
		t.Fatal(err)
	}
	closeJournal(t, reopened)
	backend.failRestore = true
	results, err := reopened.Recover(t.Context(), map[string]lab.Handler{"team": backend, "settings": backend})
	if !errors.Is(err, lab.ErrCleanup) || len(results) != 2 || results[0].Outcome != "unresolved" || results[1].Outcome != "cleaned" || len(backend.resources) != 0 {
		t.Fatal("partial cleanup did not preserve progress")
	}
	if strings.Join(backend.trace, ",") != "restore,resolve,delete" {
		t.Fatal("cleanup order or lost-response reconciliation changed")
	}
	if strings.Contains(fmt.Sprint(err), "private") {
		t.Fatal("cleanup error leaked diagnostics")
	}
	if err := reopened.Close(); err != nil {
		t.Fatal(err)
	}
	last, err := lab.OpenJournal(file, "fixture-tenant/cloud/identity")
	if err != nil {
		t.Fatal(err)
	}
	closeJournal(t, last)
	backend.failRestore = false
	backend.trace = nil
	if _, err := last.Recover(t.Context(), map[string]lab.Handler{"team": backend, "settings": backend}); err != nil {
		t.Fatal(err)
	}
	if backend.settings != "original" || strings.Join(backend.trace, ",") != "restore" {
		t.Fatal("restoration snapshot/progress was lost")
	}
	backend.trace = nil
	if _, err := last.Recover(t.Context(), map[string]lab.Handler{"team": backend, "settings": backend}); err != nil || len(backend.trace) != 0 {
		t.Fatal("completed cleanup repeated")
	}
	for _, formatted := range []string{fmt.Sprint(last), fmt.Sprintf("%#v", last), fmt.Sprintf("%+v", lab.Entry{ResourceID: "private-id", Snapshot: json.RawMessage(`{"private":true}`)})} {
		if strings.Contains(formatted, "private-id") || strings.Contains(formatted, "true") {
			t.Fatal("journal formatting exposed recovery data")
		}
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(file)
		if err != nil || info.Mode().Perm() != 0o600 {
			t.Fatal("journal file is not private")
		}
	}
}

func TestUnitJournalUnknownIdentityAndCancellationRetainIntent(t *testing.T) {
	t.Parallel()
	j, err := lab.NewJournal(filepath.Join(t.TempDir(), "private"), "fixture")
	if err != nil {
		t.Fatal(err)
	}
	closeJournal(t, j)
	name, err := j.FixtureName("channel")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.PlanCreate("channel", name); err != nil {
		t.Fatal(err)
	}
	var nilHandler *recoveryBackend
	if _, err := j.Recover(t.Context(), map[string]lab.Handler{"channel": nilHandler}); !errors.Is(err, lab.ErrCleanup) {
		t.Fatal("missing handler marked clean")
	}
	backend := &recoveryBackend{resources: map[string]string{}}
	backend.ambiguous = true
	if _, err := j.Recover(t.Context(), map[string]lab.Handler{"channel": backend}); !errors.Is(err, lab.ErrCleanup) || strings.Join(backend.trace, ",") != "resolve" {
		t.Fatal("ambiguous identity triggered cleanup")
	}
	backend.ambiguous, backend.trace = false, nil
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := j.Recover(ctx, map[string]lab.Handler{"channel": backend}); !errors.Is(err, context.Canceled) || len(backend.trace) != 0 {
		t.Fatal("canceled cleanup sent requests")
	}
	if _, err := j.Recover(t.Context(), map[string]lab.Handler{"channel": backend}); err != nil {
		t.Fatal("verified absence was not recoverable")
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	backend.trace = nil
	if _, err := j.Recover(t.Context(), map[string]lab.Handler{"channel": backend}); !errors.Is(err, lab.ErrJournal) || len(backend.trace) != 0 {
		t.Fatal("closed journal allowed cleanup without recording")
	}
	if _, err := j.PlanCreate("channel", name); !errors.Is(err, lab.ErrJournal) {
		t.Fatal("closed journal accepted intent")
	}
}

func TestUnitJournalConcurrentPlansSurviveReopen(t *testing.T) {
	t.Parallel()
	directory := filepath.Join(t.TempDir(), "private")
	j, err := lab.NewJournal(directory, "fixture")
	if err != nil {
		t.Fatal(err)
	}
	closeJournal(t, j)
	var group sync.WaitGroup
	failures := make(chan error, 12)
	for i := range 12 {
		group.Add(1)
		go func() {
			defer group.Done()
			name, err := j.FixtureName("team-" + strconv.Itoa(i))
			if err == nil {
				_, err = j.PlanCreate("team", name)
			}
			if err != nil {
				failures <- err
			}
		}()
	}
	group.Wait()
	close(failures)
	for err := range failures {
		t.Fatal(err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := lab.OpenJournal(filepath.Join(directory, j.FileName()), "fixture")
	if err != nil {
		t.Fatal(err)
	}
	closeJournal(t, reopened)
	backend := &recoveryBackend{resources: map[string]string{}}
	results, err := reopened.Recover(t.Context(), map[string]lab.Handler{"team": backend})
	if err != nil || len(results) != 12 || len(backend.trace) != 12 {
		t.Fatal("concurrent plans were lost")
	}
}

func TestUnitAppliedCreationCleansRecordedIdentityWithoutRediscovery(t *testing.T) {
	t.Parallel()
	j, err := lab.NewJournal(filepath.Join(t.TempDir(), "private"), "fixture")
	if err != nil {
		t.Fatal(err)
	}
	closeJournal(t, j)
	name, err := j.FixtureName("team")
	if err != nil {
		t.Fatal(err)
	}
	id, err := j.PlanCreate("team", name)
	if err != nil {
		t.Fatal(err)
	}
	if err := j.Applied(id, "known-id"); err != nil {
		t.Fatal(err)
	}
	backend := &recoveryBackend{resources: map[string]string{name: "known-id"}}
	if _, err := j.Recover(t.Context(), map[string]lab.Handler{"team": backend}); err != nil {
		t.Fatal(err)
	}
	if strings.Join(backend.trace, ",") != "delete" || len(backend.resources) != 0 {
		t.Fatal("recorded creation was rediscovered or not cleaned")
	}
}

func TestUnitJournalNamesAndCorruptDataAreRejected(t *testing.T) {
	t.Parallel()
	directory := filepath.Join(t.TempDir(), "private")
	j, err := lab.NewJournal(directory, "fixture")
	if err != nil {
		t.Fatal(err)
	}
	closeJournal(t, j)
	other, err := lab.NewJournal(directory, "fixture")
	if err != nil {
		t.Fatal(err)
	}
	closeJournal(t, other)
	a, err := j.FixtureName("team")
	if err != nil {
		t.Fatal(err)
	}
	b, err := other.FixtureName("team")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("fixture names collide across runs")
	}
	if _, err := j.PlanCreate("team", b); !errors.Is(err, lab.ErrJournal) {
		t.Fatal("another run's fixture accepted")
	}
	if _, err := j.FixtureName("../bad"); !errors.Is(err, lab.ErrJournal) {
		t.Fatal("invalid fixture label accepted")
	}
	if _, err := j.PlanRestore("settings", "id", json.RawMessage(`{`)); !errors.Is(err, lab.ErrJournal) {
		t.Fatal("invalid snapshot accepted")
	}
	file := filepath.Join(directory, j.FileName())
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(`{"version":1`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := lab.OpenJournal(file, "fixture"); !errors.Is(err, lab.ErrJournal) {
		t.Fatal("truncated journal accepted")
	}
	if runtime.GOOS != "windows" {
		//nolint:gosec // G302: intentionally public test directory verifies rejection; no credentials exist here.
		if err := os.Chmod(directory, 0o755); err != nil {
			t.Fatal(err)
		}
		if _, err := lab.NewJournal(directory, "fixture"); !errors.Is(err, lab.ErrJournal) {
			t.Fatal("public directory accepted")
		}
	}
}
