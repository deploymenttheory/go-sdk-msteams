package auth

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
)

// MaxAssertionBytes bounds file- and callback-supplied workload assertions.
const MaxAssertionBytes = 1 << 20

// ErrAssertionUnavailable means the selected source could not provide a nonempty,
// bounded assertion. Its text never includes file paths, material or callback errors.
var ErrAssertionUnavailable = errors.New("workload assertion is unavailable or invalid")

// workloadAssertion reads a fresh assertion for each uncached credential exchange.
// Files are caller-owned, may rotate, and are never read during construction.
func workloadAssertion(ctx context.Context, cfg *Config) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	var assertion string
	if cfg.Assertion != nil {
		value, err := cfg.Assertion(ctx)
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		if err != nil || len(value) > MaxAssertionBytes {
			return "", ErrAssertionUnavailable
		}
		assertion = value
	} else {
		// Reject known nonregular sources before opening them (for example a FIFO).
		// The caller controls this local path and its containing directory.
		info, err := os.Stat(cfg.AssertionFile)
		if err != nil || !info.Mode().IsRegular() || info.Size() > MaxAssertionBytes {
			return "", ErrAssertionUnavailable
		}
		file, err := os.Open(cfg.AssertionFile)
		if err != nil {
			return "", ErrAssertionUnavailable
		}
		defer func() { _ = file.Close() }()
		info, err = file.Stat()
		if err != nil || !info.Mode().IsRegular() {
			return "", ErrAssertionUnavailable
		}
		content, err := io.ReadAll(io.LimitReader(file, MaxAssertionBytes+1))
		if err != nil || len(content) > MaxAssertionBytes {
			return "", ErrAssertionUnavailable
		}
		assertion = string(content)
		clear(content)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	assertion = strings.TrimSpace(assertion)
	if assertion == "" || len(assertion) > MaxAssertionBytes {
		return "", ErrAssertionUnavailable
	}
	// The issuer and Entra validate the assertion. Do not infer trust by decoding JWTs.
	return assertion, nil
}
