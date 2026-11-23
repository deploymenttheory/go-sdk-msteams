package callqueuesacc

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	cq "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/call_queues"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const (
	fixtureService = "call_queues"
	// probeQueue is a reserved fixture-form ID that names no call queue.
	probeQueue = "00000000-0000-4000-8000-000000000001"
	// fixtureQueue replaces the live queue ID in captured fixtures.
	fixtureQueue = "00000000-0000-4000-8000-000000000003"
)

type queueObservation struct {
	Success           bool   `json:"success"`
	Name              string `json:"name"`
	RoutingMethod     string `json:"routing_method"`
	AgentAlertTime    int32  `json:"agent_alert_time"`
	AllowOptOut       bool   `json:"allow_opt_out"`
	ConferenceMode    bool   `json:"conference_mode"`
	TimeoutThreshold  int32  `json:"timeout_threshold"`
	OverflowThreshold int32  `json:"overflow_threshold"`
	LanguageID        string `json:"language_id"`
}

var routingNames = map[int32]string{cq.RoutingAttendant: "Attendant", cq.RoutingSerial: "Serial", cq.RoutingRoundRobin: "RoundRobin", cq.RoutingLongestIdle: "LongestIdle"}

// comparePowerShellQueue reads the owned queue with the pinned module and
// compares the projected settings with the Go read.
func comparePowerShellQueue(t *testing.T, ctx context.Context, fixture string, value *cq.CallQueue) {
	t.Helper()
	if os.Getenv("MSTEAMS_LAB_POWERSHELL_ORACLE") != "1" {
		return
	}
	module := os.Getenv("MSTEAMS_LAB_MODULE_PATH")
	if module == "" || !ownedQueue(value, fixture) {
		t.Fatal("explicit pinned module and owned queue selectors are required")
	}
	command := labkit.OracleCommand(ctx, "read-call-queue.ps1")
	command.WaitDelay = time.Second
	command.Env = []string{}
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if v, ok := os.LookupEnv(key); ok {
			command.Env = append(command.Env, key+"="+v)
		}
	}
	command.Env = append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"),
		"MSTEAMS_LAB_MODULE_PATH="+module, "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"),
		"MSTEAMS_LAB_CALL_QUEUE_FIXTURE="+fixture, "MSTEAMS_LAB_CALL_QUEUE_ID="+*value.Identity)
	var output labkit.OracleBuffer
	command.Stdout, command.Stderr = &output, io.Discard
	defer func() { clear(output.Data) }()
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned PowerShell call queue read failed; no output retained")
	}
	var observed queueObservation
	if json.Unmarshal(output.Data, &observed) != nil || !observed.Success {
		t.Fatal("pinned PowerShell could not read the owned call queue")
	}
	if observed.Name != *value.Name || observed.RoutingMethod != routingNames[*value.RoutingMethod] || observed.AgentAlertTime != *value.AgentAlertTime ||
		observed.AllowOptOut != *value.AllowOptOut || observed.ConferenceMode != *value.ConferenceMode || observed.TimeoutThreshold != *value.TimeoutThreshold ||
		observed.OverflowThreshold != *value.OverflowThreshold || observed.LanguageID != *value.LanguageID {
		t.Fatal("PowerShell and Go call queue observations differ")
	}
	t.Log("PowerShell 8.1.0 / Go call queue parity: passed")
}

func tenantReplacements(t *testing.T, binding string) map[string]string {
	t.Helper()
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok || !labkit.DirectoryID.MatchString(tenant) {
		t.Fatal("tenant binding required")
	}
	return map[string]string{tenant: "00000000-0000-4000-8000-0000000000aa", strings.ToLower(tenant): "00000000-0000-4000-8000-0000000000aa"}
}

func ptr[T any](v T) *T { return &v }

// TestAccCallQueueLifecycle creates one uniquely named call queue, reads and
// lists it, replaces its settings from a read-back, and deletes it. It never
// touches a pre-existing queue, agent or resource account.
func TestAccCallQueueLifecycle(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("call queue lifecycle requires explicit mutation opt-in and a private journal directory")
	} else if err != nil {
		t.Fatal(err)
	}
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.CallQueues
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private call queue recovery journal unavailable")
	}
	cleanup := queueCleanup{service: service}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"call-queue": cleanup}); err != nil {
			t.Error("call queue cleanup incomplete; retained private journal requires recovery")
		}
		if err := journal.Close(); err != nil {
			t.Error("private call queue journal close failed")
		}
	})
	fixture, err := journal.FixtureName("queue")
	if err != nil {
		t.Fatal("call queue fixture name unavailable")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := tenantReplacements(t, binding)
	replacements[fixture] = "msteams-sdk-fixture-queue"

	if _, response, err := service.GetByID(ctx, probeQueue); !missingQueue(err) {
		t.Fatal("missing call queue was not a 400 identity error")
	} else {
		labkit.WriteResponseFixture(t, fixtureService, "get_missing", response, replacements)
	}
	t.Log("call queue missing get: passed")

	entry, err := journal.PlanCreate("call-queue", fixture)
	if err != nil {
		t.Fatal("call queue create was not journaled")
	}
	settings := cq.Settings{Name: &fixture, RoutingMethod: ptr(cq.RoutingLongestIdle), AgentAlertTime: ptr(int32(30)), UseDefaultMusicOnHold: ptr(true),
		OverflowThreshold: ptr(int32(50)), TimeoutThreshold: ptr(int32(1200)), LanguageID: ptr("en-US")}
	created, response, err := service.Create(ctx, settings, nil)
	if err == nil && created != nil && created.Identity != nil {
		replacements[*created.Identity] = fixtureQueue
	}
	labkit.WriteResponseFixture(t, fixtureService, "create", response, replacements)
	if err != nil || !ownedQueue(created, fixture) {
		t.Fatal("live call queue create failed; recovery journal retained")
	}
	id := *created.Identity
	if err := journal.Applied(entry, id); err != nil {
		t.Fatal("created call queue ID was not journaled")
	}
	if *created.RoutingMethod != cq.RoutingLongestIdle || created.PresenceAwareRouting == nil || *created.PresenceAwareRouting || *created.AgentAlertTime != 30 {
		t.Fatal("create receipt lost its settings")
	}
	t.Log("call queue create: passed")

	value, response, err := service.GetByID(ctx, id)
	if err != nil || !ownedQueue(value, fixture) || *value.Identity != id || *value.TimeoutThreshold != 1200 || *value.LanguageID != "en-US" {
		t.Fatal("owned call queue read-back failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "get", response, replacements)
	t.Log("call queue get: passed")

	listed, response, err := service.List(ctx, &cq.ListOptions{NameFilter: &fixture, First: ptr(int32(10))})
	if err != nil {
		t.Fatal("call queue list failed")
	}
	found := 0
	for _, item := range listed {
		if item.Identity != nil && *item.Identity == id && ownedQueue(item, fixture) {
			found++
		}
	}
	if found != 1 {
		t.Fatal("owned call queue missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, fixtureService, "list", response, replacements)
	t.Log("call queue list: passed")

	// Set-CsCallQueue sends the whole queue: change three settings on a body
	// rebuilt from the read-back and prove the rest survive.
	update := value.Settings()
	update.AgentAlertTime, update.AllowOptOut, update.RoutingMethod = ptr(int32(45)), ptr(true), ptr(cq.RoutingRoundRobin)
	_, response, err = service.UpdateByID(ctx, id, update, nil)
	labkit.WriteResponseFixture(t, fixtureService, "update", response, replacements)
	if err != nil {
		t.Fatal("live call queue update failed")
	}
	value, response, err = service.GetByID(ctx, id)
	if err != nil || !ownedQueue(value, fixture) || *value.AgentAlertTime != 45 || !*value.AllowOptOut || *value.RoutingMethod != cq.RoutingRoundRobin ||
		*value.TimeoutThreshold != 1200 || *value.OverflowThreshold != 50 || *value.LanguageID != "en-US" || !*value.UseDefaultMusicOnHold {
		t.Fatal("call queue update was not read back with unchanged settings preserved")
	}
	labkit.WriteResponseFixture(t, fixtureService, "get_updated", response, replacements)
	comparePowerShellQueue(t, ctx, fixture, value)
	t.Log("call queue update/read-back: passed")

	response, err = service.DeleteByID(ctx, id)
	labkit.WriteResponseFixture(t, fixtureService, "delete", response, replacements)
	if err != nil {
		t.Fatal("live call queue delete failed")
	}
	if err := cleanup.absent(ctx, lab.Entry{Fixture: fixture, ResourceID: id}); err != nil {
		t.Fatal("call queue deletion was not observed")
	}
	if response, err = service.DeleteByID(ctx, id); !missingQueue(err) {
		t.Fatal("repeated call queue delete was not rejected")
	}
	labkit.WriteResponseFixture(t, fixtureService, "delete_missing", response, replacements)
	t.Log("call queue delete: passed")
}

// TestAccCallQueueRecover is an explicit operator-selected recovery path for
// an interrupted run. OpenJournal checks tenant/cloud/identity binding locally.
func TestAccCallQueueRecover(t *testing.T) {
	file := os.Getenv("MSTEAMS_LAB_RECOVERY_FILE")
	if file == "" {
		t.Skip("no explicit private recovery journal selected")
	}
	cfg, err := lab.Load()
	if err != nil || cfg.Require(lab.Mutation) != nil {
		t.Fatal("explicit mutation opt-in is required for call queue recovery")
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
		t.Fatal("call queue recovery incomplete; private journal retained")
	}
	t.Log("owned call queue recovery: passed")
}
