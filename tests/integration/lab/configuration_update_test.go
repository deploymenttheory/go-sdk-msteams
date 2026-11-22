package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	messaging "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_messaging_configuration"
	migration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_migration_configuration"
	"github.com/go-resty/resty/v2"
)

// These operations are serialized with every tenant mutation in CI. Recovery
// persists only the selected Boolean and restores it through the same public API.
type singletonConfiguration struct {
	kind, field string
	get         func(context.Context) (map[string]any, *resty.Response, error)
	set         func(context.Context, bool) (*resty.Response, error)
}

func messagingSingleton(sdk *msteams.Client) singletonConfiguration {
	s := sdk.TeamsAdminAPI.TeamsMessagingConfigurations
	return singletonConfiguration{"messaging", "EnableVideoMessageCaptions", func(ctx context.Context) (map[string]any, *resty.Response, error) {
		v, r, e := s.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return messagingConfigurationProjection(v), r, nil
	}, func(ctx context.Context, value bool) (*resty.Response, error) {
		_, r, e := s.UpdateByIdentity(ctx, "Global", messaging.UpdateRequest{ScalarSettings: messaging.ScalarSettings{EnableVideoMessageCaptions: policy.Set(value)}})
		return r, e
	}}
}
func migrationSingleton(sdk *msteams.Client) singletonConfiguration {
	s := sdk.TeamsAdminAPI.TeamsMigrationConfigurations
	return singletonConfiguration{"migration", "EnableLegacyClientInterop", func(ctx context.Context) (map[string]any, *resty.Response, error) {
		v, r, e := s.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return migrationConfigurationProjection(v), r, nil
	}, func(ctx context.Context, value bool) (*resty.Response, error) {
		return s.UpdateGlobal(ctx, migration.UpdateRequest{EnableLegacyClientInterop: &value})
	}}
}

type configurationRestore struct{ singletonConfiguration }

func (configurationRestore) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}
func (configurationRestore) Delete(context.Context, lab.Entry) error { return lab.ErrUnresolved }
func (h configurationRestore) Restore(ctx context.Context, entry lab.Entry) error {
	var original struct {
		Value *bool `json:"value"`
	}
	d := json.NewDecoder(bytes.NewReader(entry.Snapshot))
	d.DisallowUnknownFields()
	if entry.ResourceID != "Global" || d.Decode(&original) != nil || d.Decode(new(any)) != io.EOF || original.Value == nil {
		return lab.ErrUnresolved
	}
	// A previous restoration may have succeeded before its receipt was saved.
	current, _, err := h.get(ctx)
	if err != nil {
		return err
	}
	if value, ok := current[h.field].(*bool); ok && value != nil && *value == *original.Value {
		return nil
	}
	if _, err := h.set(ctx, *original.Value); err != nil {
		return err
	}
	return observeUntil(ctx, func(ctx context.Context) (bool, error) {
		current, _, err := h.get(ctx)
		value, ok := current[h.field].(*bool)
		return err == nil && ok && value != nil && *value == *original.Value, err
	})
}

func testSingletonUpdate(t *testing.T, build func(*msteams.Client) singletonConfiguration) {
	t.Helper()
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid configuration mutation enablement")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("configuration updates require explicit mutation enablement and recovery directory")
	} else if err != nil {
		t.Fatal("invalid configuration mutation enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	s := build(sdk)
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private configuration recovery journal unavailable")
	}
	handlers := map[string]lab.Handler{s.kind + "-configuration": configurationRestore{s}}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, handlers); err != nil {
			t.Error("configuration restoration unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("configuration recovery journal close failed")
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	before, response, err := s.get(ctx)
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("configuration original read failed")
	}
	value, ok := before[s.field].(*bool)
	if !ok || value == nil {
		t.Fatal("configuration original Boolean is unavailable; update refused")
	}
	compare := func(settings map[string]any) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !sameTemplateProjection(settings, tenantConfigurationOracle(t, ctx, s.kind, tenant)) {
			t.Fatal("Go and PowerShell configuration update getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go configuration read parity: passed")
	}
	compare(before)
	snapshot, err := json.Marshal(map[string]bool{"value": *value})
	if err != nil {
		t.Fatal("configuration original Boolean could not be journaled")
	}
	entry, err := journal.PlanRestore(s.kind+"-configuration", "Global", snapshot)
	if err != nil {
		t.Fatal("configuration restoration intent not journaled")
	}
	response, err = s.set(ctx, !*value)
	writeResponseFixture(t, "teams_"+s.kind+"_configuration", "update", response, nil)
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("configuration update failed; restoration journal retained")
	}
	if err := journal.Applied(entry, "Global"); err != nil {
		t.Fatal("configuration update receipt not journaled")
	}
	want := make(map[string]any, len(before))
	for k, v := range before {
		want[k] = v
	}
	changed := !*value
	want[s.field] = &changed
	var updated map[string]any
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		read, r, err := s.get(ctx)
		matched := err == nil && sameTemplateProjection(want, read)
		if matched {
			updated = read
			writeResponseFixture(t, "teams_"+s.kind+"_configuration", "update_readback", r, nil)
		}
		return matched, err
	}); err != nil {
		t.Fatal("configuration update or omitted-field preservation did not converge")
	}
	t.Log("configuration update/read-back and omission: passed")
	compare(updated)
	if _, err := journal.Recover(ctx, handlers); err != nil {
		t.Fatal("configuration original Boolean restoration failed")
	}
	restored, response, err := s.get(ctx)
	if err != nil || !sameTemplateProjection(before, restored) {
		t.Fatal("configuration original getter state was not independently restored")
	}
	writeResponseFixture(t, "teams_"+s.kind+"_configuration", "restored_readback", response, nil)
	compare(restored)
	t.Log("configuration restoration/read-back: passed")
}

func TestAccMessagingConfigurationUpdateRestore(t *testing.T) {
	testSingletonUpdate(t, messagingSingleton)
}
func TestAccMigrationConfigurationUpdateRestore(t *testing.T) {
	testSingletonUpdate(t, migrationSingleton)
}

func TestUnitConfigurationRecoveryRejectsForeignTargetsAndSnapshots(t *testing.T) {
	t.Parallel()
	h := configurationRestore{}
	for _, entry := range []lab.Entry{
		{ResourceID: "Tag:foreign", Snapshot: json.RawMessage(`{"value":true}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":null}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":"false"}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":false,"unknown":true}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":false} {}`)},
	} {
		if !errors.Is(h.Restore(t.Context(), entry), lab.ErrUnresolved) {
			t.Fatal("configuration recovery accepted an unsupported target or snapshot")
		}
	}
}

func TestUnitConfigurationRecoveryVerifiesTheOriginalAfterWriting(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		accepted bool
	}{{"persisted", true}, {"accepted-without-persistence", false}} {
		t.Run(tc.name, func(t *testing.T) {
			current, reads, writes := false, 0, 0
			readErr := errors.New("restoration read-back differs")
			h := configurationRestore{singletonConfiguration{field: "setting", get: func(context.Context) (map[string]any, *resty.Response, error) {
				reads++
				if reads > 1 && !current {
					return nil, nil, readErr
				}
				return map[string]any{"setting": &current}, nil, nil
			}, set: func(_ context.Context, value bool) (*resty.Response, error) {
				writes++
				if !value {
					t.Fatal("recovery did not send the journaled original")
				}
				current = tc.accepted
				return nil, nil
			}}}
			err := h.Restore(t.Context(), lab.Entry{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":true}`)})
			if reads != 2 || writes != 1 || (tc.accepted && err != nil) || (!tc.accepted && !errors.Is(err, readErr)) {
				t.Fatal("restoration trusted write acceptance without independent verification")
			}
		})
	}
}
