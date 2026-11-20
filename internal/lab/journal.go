package lab

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

var (
	ErrJournal    = errors.New("private lab journal is unavailable or invalid")
	ErrCleanup    = errors.New("lab cleanup has unresolved entries")
	ErrUnresolved = errors.New("remote fixture identity is unresolved")
	labelPattern  = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}$`)
	runIDPattern  = regexp.MustCompile(`^[a-z2-7]{26,64}$`)
)

// Entry is private recovery data, not a CI artifact. Snapshot must be the minimal
// settings needed to restore a shared resource, never credentials or raw dumps.
type Entry struct {
	ID         string          `json:"id"`
	Kind       string          `json:"kind"`
	Fixture    string          `json:"fixture,omitempty"`
	ResourceID string          `json:"resource_id,omitempty"`
	Action     string          `json:"action"`
	State      string          `json:"state"`
	Snapshot   json.RawMessage `json:"snapshot,omitempty"`
}

func (Entry) String() string   { return "private lab recovery entry" }
func (Entry) GoString() string { return "private lab recovery entry" }

type journalData struct {
	Version int     `json:"version"`
	RunID   string  `json:"run_id"`
	Binding string  `json:"binding_sha256"`
	Entries []Entry `json:"entries"`
}

// Journal serializes updates within one process. A journal has one owner; stop
// its original run before reopening it for recovery. No automatic mutation occurs
// on Open. Close releases local handles and leaves recovery data in place.
type Journal struct {
	mu        sync.Mutex
	recoverMu sync.Mutex
	root      *os.Root
	filename  string
	data      journalData
}

func (*Journal) String() string   { return "private lab cleanup journal" }
func (*Journal) GoString() string { return "private lab cleanup journal" }

// NewJournal creates a private run file. binding must identify the selected
// tenant/cloud/identity; only its digest is retained to reject cross-tenant recovery.
func NewJournal(directory, binding string) (*Journal, error) {
	if binding == "" {
		return nil, ErrJournal
	}
	root, err := privateRoot(directory)
	if err != nil {
		return nil, err
	}
	runID := strings.ToLower(rand.Text())
	j := &Journal{root: root, filename: "msteams-lab-" + runID + ".json", data: journalData{Version: 1, RunID: runID, Binding: bindingDigest(binding), Entries: []Entry{}}}
	if err := j.save(j.data); err != nil {
		_ = root.Close()
		return nil, err
	}
	return j, nil
}

// OpenJournal reads existing private state without performing network calls.
func OpenJournal(file, binding string) (*Journal, error) {
	if binding == "" {
		return nil, ErrJournal
	}
	root, err := privateRoot(filepath.Dir(file))
	if err != nil {
		return nil, err
	}
	fail := func() (*Journal, error) { _ = root.Close(); return nil, ErrJournal }
	name := filepath.Base(file)
	info, err := root.Lstat(name)
	if err != nil || !info.Mode().IsRegular() || runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 || info.Size() > 8<<20 {
		return fail()
	}
	stream, err := root.Open(name)
	if err != nil {
		return fail()
	}
	raw, readErr := io.ReadAll(io.LimitReader(stream, (8<<20)+1))
	closeErr := stream.Close()
	var data journalData
	if readErr != nil || closeErr != nil || len(raw) > 8<<20 || json.Unmarshal(raw, &data) != nil || data.Version != 1 || data.Binding != bindingDigest(binding) || !validData(data) {
		clear(raw)
		return fail()
	}
	clear(raw)
	return &Journal{root: root, filename: name, data: data}, nil
}

func privateRoot(directory string) (*os.Root, error) {
	if directory == "" {
		return nil, ErrJournal
	}
	created := false
	if err := os.Mkdir(directory, 0o700); err == nil {
		created = true
	} else if !errors.Is(err, os.ErrExist) {
		return nil, ErrJournal
	}
	if created && runtime.GOOS != "windows" {
		parent, err := os.OpenRoot(filepath.Dir(filepath.Clean(directory)))
		if err != nil {
			return nil, ErrJournal
		}
		err = syncDirectory(parent)
		_ = parent.Close()
		if err != nil {
			return nil, err
		}
	}
	info, err := os.Lstat(directory)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		return nil, ErrJournal
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, ErrJournal
	}
	return root, nil
}

func bindingDigest(binding string) string {
	digest := sha256.Sum256([]byte(binding))
	return hex.EncodeToString(digest[:])
}

// FileName returns only the local file basename so entry points can locate recovery
// data. Do not print its contents. The directory remains operator configuration.
func (j *Journal) FileName() string { return j.filename }

func (j *Journal) FixtureName(label string) (string, error) {
	if !labelPattern.MatchString(label) {
		return "", ErrJournal
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	return "msteams-sdk-" + j.data.RunID + "-" + label, nil
}

// PlanCreate must persist before sending creation. The unique fixture name allows
// a recovery handler to reconcile a lost response without guessing a resource ID.
func (j *Journal) PlanCreate(kind, fixture string) (string, error) {
	j.mu.Lock()
	runID := j.data.RunID
	j.mu.Unlock()
	if !strings.HasPrefix(fixture, "msteams-sdk-"+runID+"-") {
		return "", ErrJournal
	}
	return j.plan(Entry{Kind: kind, Fixture: fixture, Action: "delete"})
}

// PlanRestore persists the original shared settings before any update is sent.
func (j *Journal) PlanRestore(kind, resourceID string, snapshot json.RawMessage) (string, error) {
	if resourceID == "" || len(snapshot) == 0 || len(snapshot) > 64<<10 || !json.Valid(snapshot) {
		return "", ErrJournal
	}
	return j.plan(Entry{Kind: kind, ResourceID: resourceID, Action: "restore", Snapshot: append(json.RawMessage(nil), snapshot...)})
}

func (j *Journal) plan(entry Entry) (string, error) {
	if !labelPattern.MatchString(entry.Kind) {
		return "", ErrJournal
	}
	entry.ID, entry.State = strings.ToLower(rand.Text()), "planned"
	err := j.update(func(data *journalData) error { data.Entries = append(data.Entries, entry); return nil })
	if err != nil {
		return "", err
	}
	return entry.ID, nil
}

// Applied records the actual ID immediately after a creation/update response,
// including partial-success responses. A missing ID keeps creation unresolved.
func (j *Journal) Applied(id, resourceID string) error {
	if resourceID == "" || len(resourceID) > 2048 {
		return ErrJournal
	}
	return j.update(func(data *journalData) error {
		for i := range data.Entries {
			entry := &data.Entries[i]
			if entry.ID == id && entry.State == "planned" {
				if entry.Action == "restore" && entry.ResourceID != resourceID {
					return ErrJournal
				}
				entry.ResourceID, entry.State = resourceID, "applied"
				return nil
			}
		}
		return ErrJournal
	})
}

// Handler implementations own resource-specific lookup, ownership checks and
// idempotent cleanup. Resolve must return ErrUnresolved for ambiguous matches.
// Empty ID and nil error means verified absence, not a failed lookup.
type Handler interface {
	Resolve(context.Context, Entry) (string, error)
	Delete(context.Context, Entry) error
	Restore(context.Context, Entry) error
}

type CleanupResult struct {
	Kind    string `json:"kind"`
	Outcome string `json:"outcome"`
}

// Recover cleans entries in reverse order and journals each success. Failure
// retains the entry for another attempt and does not block independent entries.
// Use a fresh bounded cleanup context if the test context has been canceled.
func (j *Journal) Recover(ctx context.Context, handlers map[string]Handler) ([]CleanupResult, error) {
	if ctx == nil {
		return nil, ErrConfig
	}
	j.recoverMu.Lock()
	defer j.recoverMu.Unlock()
	j.mu.Lock()
	if j.root == nil {
		j.mu.Unlock()
		return nil, ErrJournal
	}
	entries := cloneData(j.data).Entries
	j.mu.Unlock()
	var results []CleanupResult
	failed := false
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if entry.State == "cleaned" {
			continue
		}
		if err := ctx.Err(); err != nil {
			return results, err
		}
		handler := handlers[entry.Kind]
		result := CleanupResult{Kind: entry.Kind, Outcome: "unresolved"}
		err := ErrUnresolved
		if !nilDependency(handler) {
			if entry.Action == "delete" {
				if entry.ResourceID == "" {
					entry.ResourceID, err = handler.Resolve(ctx, entry)
				} else {
					err = nil
				}
				if err == nil && entry.ResourceID != "" {
					err = handler.Delete(ctx, entry)
				}
			} else {
				err = handler.Restore(ctx, entry)
			}
		}
		if err == nil {
			err = j.update(func(data *journalData) error {
				for k := range data.Entries {
					if data.Entries[k].ID == entry.ID {
						data.Entries[k].State = "cleaned"
						return nil
					}
				}
				return ErrJournal
			})
		}
		if err == nil {
			result.Outcome = "cleaned"
		} else {
			failed = true
		}
		results = append(results, result)
	}
	if err := ctx.Err(); err != nil {
		return results, err
	}
	if failed {
		return results, ErrCleanup
	}
	return results, nil
}

func cloneData(data journalData) journalData {
	data.Entries = append([]Entry(nil), data.Entries...)
	for i := range data.Entries {
		data.Entries[i].Snapshot = append(json.RawMessage(nil), data.Entries[i].Snapshot...)
	}
	return data
}

func validData(data journalData) bool {
	if !runIDPattern.MatchString(data.RunID) || len(data.Entries) > 10000 {
		return false
	}
	seen := make(map[string]bool, len(data.Entries))
	for _, entry := range data.Entries {
		if !labelPattern.MatchString(entry.Kind) || entry.ID == "" || seen[entry.ID] || entry.State != "planned" && entry.State != "applied" && entry.State != "cleaned" {
			return false
		}
		seen[entry.ID] = true
		if entry.State == "applied" && entry.ResourceID == "" {
			return false
		}
		if entry.Action == "restore" {
			if entry.ResourceID == "" || len(entry.Snapshot) == 0 || len(entry.Snapshot) > 64<<10 || !json.Valid(entry.Snapshot) {
				return false
			}
		} else if entry.Action != "delete" || !strings.HasPrefix(entry.Fixture, "msteams-sdk-"+data.RunID+"-") {
			return false
		}
	}
	return true
}

func (j *Journal) update(change func(*journalData) error) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.root == nil {
		return ErrJournal
	}
	data := cloneData(j.data)
	if err := change(&data); err != nil {
		return err
	}
	if err := j.save(data); err != nil {
		return err
	}
	j.data = data
	return nil
}

func (j *Journal) save(data journalData) error {
	if !validData(data) {
		return ErrJournal
	}
	raw, err := json.Marshal(data)
	if err != nil || len(raw) > 8<<20 {
		return ErrJournal
	}
	defer clear(raw)
	temporary := ".write-" + rand.Text()
	stream, err := j.root.OpenFile(temporary, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return ErrJournal
	}
	defer func() { _ = j.root.Remove(temporary) }()
	_, writeErr := stream.Write(raw)
	syncErr := stream.Sync()
	closeErr := stream.Close()
	if writeErr != nil || syncErr != nil || closeErr != nil || j.root.Rename(temporary, j.filename) != nil {
		return ErrJournal
	}
	return syncDirectory(j.root)
}

func syncDirectory(root *os.Root) error {
	// Go does not expose a portable directory flush on Windows. File data is
	// synced there; process-interruption recovery still retains complete JSON.
	if runtime.GOOS == "windows" {
		return nil
	}
	directory, err := root.Open(".")
	if err != nil {
		return ErrJournal
	}
	syncErr := directory.Sync()
	closeErr := directory.Close()
	if syncErr != nil || closeErr != nil {
		return ErrJournal
	}
	return nil
}

func (j *Journal) Close() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.root == nil {
		return nil
	}
	err := j.root.Close()
	j.root = nil
	if err != nil {
		return ErrJournal
	}
	return nil
}
