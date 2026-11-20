// Package lab provides opt-in acceptance-test controls. It is not a public SDK API.
package lab

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Scenario string

const (
	ReadOnly      Scenario = "read-only"
	Mutation      Scenario = "mutation"
	Consequential Scenario = "consequential"
)

var (
	ErrDisabled = errors.New("lab scenario is not enabled")
	ErrConfig   = errors.New("invalid lab configuration")
)

// Config contains enablement only; credentials remain in the caller's auth config.
type Config struct {
	ReadOnly         bool
	Mutations        bool
	Consequential    bool
	JournalDirectory string
	Timeout          time.Duration
}

// Load reads environment only when explicitly called by an acceptance entry point.
func Load() (Config, error) { return Parse(os.Getenv) }

// Parse allows deterministic configuration tests without changing process state.
// Flags accept empty/0/1 only. Mutation runs require a private journal directory.
func Parse(lookup func(string) string) (Config, error) {
	if lookup == nil {
		return Config{}, ErrConfig
	}
	var cfg Config
	for key, target := range map[string]*bool{
		"MSTEAMS_LAB_READ_ONLY": &cfg.ReadOnly, "MSTEAMS_LAB_MUTATIONS": &cfg.Mutations,
		"MSTEAMS_LAB_CONSEQUENTIAL": &cfg.Consequential,
	} {
		switch lookup(key) {
		case "", "0":
		case "1":
			*target = true
		default:
			return Config{}, ErrConfig
		}
	}
	cfg.JournalDirectory = lookup("MSTEAMS_LAB_JOURNAL_DIRECTORY")
	cfg.Timeout = 90 * time.Second
	if raw := lookup("MSTEAMS_LAB_TIMEOUT_SECONDS"); raw != "" {
		seconds, err := strconv.Atoi(raw)
		if err != nil || seconds < 1 || seconds > 300 {
			return Config{}, ErrConfig
		}
		cfg.Timeout = time.Duration(seconds) * time.Second
	}
	if cfg.Mutations && (!cfg.ReadOnly || cfg.JournalDirectory == "") || cfg.Consequential && !cfg.Mutations {
		return Config{}, ErrConfig
	}
	return cfg, nil
}

func (c Config) Require(scenario Scenario) error {
	allowed := false
	switch scenario {
	case ReadOnly:
		allowed = c.ReadOnly
	case Mutation:
		allowed = c.ReadOnly && c.Mutations && c.JournalDirectory != ""
	case Consequential:
		allowed = c.ReadOnly && c.Mutations && c.Consequential && c.JournalDirectory != ""
	default:
		return ErrConfig
	}
	if !allowed {
		return ErrDisabled
	}
	return nil
}

func (Config) String() string   { return "lab enablement (private paths omitted)" }
func (Config) GoString() string { return "lab enablement (private paths omitted)" }
