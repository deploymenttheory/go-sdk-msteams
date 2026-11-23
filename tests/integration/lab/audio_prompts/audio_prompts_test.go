package audiopromptsacc

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	audio "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/audio_prompts"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const fixtureService = "audio_prompts"

// silentWAV returns a quarter second of 16 kHz mono 16-bit PCM silence. The
// service stores uploads transcoded, so exported bytes are not compared to it.
func silentWAV() []byte {
	const size = 8000
	samples := make([]byte, size)
	var out bytes.Buffer
	write := func(v any) { _ = binary.Write(&out, binary.LittleEndian, v) }
	out.WriteString("RIFF")
	write(uint32(36 + size))
	out.WriteString("WAVEfmt ")
	write(uint32(16))
	write(uint16(1))
	write(uint16(1))
	write(uint32(16000))
	write(uint32(32000))
	write(uint16(2))
	write(uint16(16))
	out.WriteString("data")
	write(uint32(size))
	out.Write(samples)
	return out.Bytes()
}

func comparePowerShellAudio(t *testing.T, ctx context.Context, fileName, id string, content []byte, deleted bool) {
	t.Helper()
	if os.Getenv("MSTEAMS_LAB_POWERSHELL_ORACLE") != "1" {
		return
	}
	module := os.Getenv("MSTEAMS_LAB_MODULE_PATH")
	if module == "" || !audioID.MatchString(id) {
		t.Fatal("explicit pinned module and owned audio selectors are required")
	}
	command := labkit.OracleCommand(ctx, "read-audio-file.ps1")
	command.WaitDelay = time.Second
	command.Env = []string{}
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR"} {
		if value, ok := os.LookupEnv(key); ok {
			command.Env = append(command.Env, key+"="+value)
		}
	}
	command.Env = append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"),
		"MSTEAMS_LAB_MODULE_PATH="+module, "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"),
		"MSTEAMS_LAB_AUDIO_FIXTURE="+fileName, "MSTEAMS_LAB_AUDIO_ID="+id)
	var output labkit.OracleBuffer
	command.Stdout = &output
	command.Stderr = io.Discard
	defer func() { clear(output.Data) }()
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned PowerShell audio read failed; no output retained")
	}
	var observed struct {
		Success           bool   `json:"success"`
		FileName          string `json:"file_name"`
		ApplicationID     string `json:"application_id"`
		MarkedForDeletion bool   `json:"marked_for_deletion"`
		ContentSHA256     string `json:"content_sha256"`
	}
	if json.Unmarshal(output.Data, &observed) != nil || !observed.Success {
		t.Fatal("pinned PowerShell could not read the owned audio file")
	}
	sum := sha256.Sum256(content)
	if observed.FileName != fileName || observed.ApplicationID != audio.TenantGlobal || observed.MarkedForDeletion != deleted || observed.ContentSHA256 != hex.EncodeToString(sum[:]) {
		t.Fatal("PowerShell and Go audio file observations differ")
	}
	t.Log("PowerShell 8.1.0 / Go audio file parity: passed")
}

// TestAccAudioFileLifecycle uploads one uniquely named TenantGlobal file,
// reads, lists and exports it, then marks it for deletion. It never touches a
// pre-existing lab audio file.
func TestAccAudioFileLifecycle(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("audio file lifecycle requires explicit mutation opt-in and a private journal directory")
	} else if err != nil {
		t.Fatal(err)
	}
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.AudioFiles
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private audio recovery journal unavailable")
	}
	cleanup := audioCleanup{service: service}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"audio-file": cleanup}); err != nil {
			t.Error("audio cleanup incomplete; retained private journal requires recovery")
		}
		if err := journal.Close(); err != nil {
			t.Error("private audio journal close failed")
		}
	})
	fixture, err := journal.FixtureName("audio")
	if err != nil {
		t.Fatal("audio fixture name unavailable")
	}
	entry, err := journal.PlanCreate("audio-file", fixture)
	if err != nil {
		t.Fatal("audio upload was not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	fileName, content := audioFileName(fixture), silentWAV()
	replacements := map[string]string{fixture: "fixture-audio"}

	created, response, err := service.Create(ctx, audio.ImportRequest{FileName: fileName, Content: content})
	if err == nil && created != nil && created.ID != nil && created.GroupID != nil {
		replacements[*created.ID] = "0123456789abcdef0123456789abcdef"
		replacements[*created.GroupID] = "fixture-group"
	}
	labkit.WriteResponseFixture(t, fixtureService, "import", response, replacements)
	if err != nil || response.StatusCode() != 201 || !ownedAudioFile(created, fixture) || created.ID == nil || !audioID.MatchString(*created.ID) {
		t.Fatal("live audio import failed; recovery journal retained")
	}
	id := *created.ID
	if err := journal.Applied(entry, id); err != nil {
		t.Fatal("uploaded audio ID was not journaled")
	}
	if created.MarkedForDeletion() || created.Duration == nil || *created.Duration == "" {
		t.Fatal("import receipt lost its media metadata")
	}
	t.Log("audio import: passed")

	value, response, err := service.GetByID(ctx, "tenantglobal", id, nil)
	if err != nil || !ownedAudioFile(value, fixture) || value.ID == nil || *value.ID != id || value.MarkedForDeletion() {
		t.Fatal("owned audio read-back failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "get", response, replacements)
	minutes := int32(5)
	if _, _, err := service.GetByID(ctx, audio.TenantGlobal, id, &audio.GetOptions{DurationInMinutes: &minutes}); err != nil {
		t.Fatal("audio read with download duration failed")
	}
	if _, _, err := service.GetByID(ctx, audio.OrgAutoAttendant, id, nil); !labkit.IsHTTP(err, 404) {
		t.Fatal("TenantGlobal file was visible under another application")
	}
	t.Log("audio get: passed")

	listed, _, err := service.List(ctx, "")
	if err != nil {
		t.Fatal("TenantGlobal audio list failed")
	}
	found := 0
	for _, item := range listed {
		if item.ID != nil && *item.ID == id && ownedAudioFile(item, fixture) {
			found++
		}
	}
	if found != 1 {
		t.Fatal("owned audio file missing or ambiguous in list")
	}
	t.Log("audio list: passed")

	exported, response, err := service.GetContentByID(ctx, audio.TenantGlobal, id)
	if err != nil || len(exported) < 44 || !bytes.Equal(exported[:4], []byte("RIFF")) || !bytes.Equal(exported[8:12], []byte("WAVE")) {
		t.Fatal("exported audio is not the stored WAV file")
	}
	labkit.WriteResponseFixture(t, fixtureService, "export", response, replacements)
	comparePowerShellAudio(t, ctx, fileName, id, exported, false)
	t.Log("audio export round trip: passed")

	response, err = service.DeleteByID(ctx, id)
	labkit.WriteResponseFixture(t, fixtureService, "delete", response, replacements)
	if err != nil || response.StatusCode() != 204 {
		t.Fatal("live audio delete failed")
	}
	if err := cleanup.absent(ctx, lab.Entry{Fixture: fixture, ResourceID: id}); err != nil {
		t.Fatal("audio deletion mark was not observed")
	}
	value, response, err = service.GetByID(ctx, audio.TenantGlobal, id, nil)
	if err != nil || !value.MarkedForDeletion() {
		t.Fatal("deleted audio read-back lost its deletion mark")
	}
	labkit.WriteResponseFixture(t, fixtureService, "get_deleted", response, replacements)
	// The service answers a repeated delete with a text/plain 400.
	if _, err = service.DeleteByID(ctx, id); !labkit.IsHTTP(err, 400) {
		t.Fatal("repeated audio delete was not rejected")
	}
	t.Log("audio delete/soft-delete observation: passed")
}

// TestAccAudioFileRecover is an explicit operator-selected recovery path for
// an interrupted run. OpenJournal checks tenant/cloud/identity binding locally.
func TestAccAudioFileRecover(t *testing.T) {
	file := os.Getenv("MSTEAMS_LAB_RECOVERY_FILE")
	if file == "" {
		t.Skip("no explicit private recovery journal selected")
	}
	cfg, err := lab.Load()
	if err != nil || cfg.Require(lab.Mutation) != nil {
		t.Fatal("explicit mutation opt-in is required for audio recovery")
	}
	sdk, binding, _ := labkit.SDKClient(t)
	journal, err := lab.OpenJournal(file, binding)
	if err != nil {
		t.Fatal("private bound recovery journal unavailable")
	}
	defer func() {
		if err := journal.Close(); err != nil {
			t.Error("private recovery journal close failed")
		}
	}()
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	if _, err := journal.Recover(ctx, RecoveryHandlers(sdk)); err != nil {
		t.Fatal("audio recovery incomplete; private journal retained")
	}
	t.Log("owned audio recovery: passed")
}
