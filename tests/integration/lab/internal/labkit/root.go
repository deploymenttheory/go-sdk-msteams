// Package labkit holds helpers shared by the per-service tenant acceptance
// packages under tests/integration/lab. It is test support only: nothing here
// is part of the SDK's public API.
package labkit

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

// Root is the repository root, found by walking up from the test package's
// working directory to go.mod. Paths built from it use forward slashes, which
// the os package accepts on every supported platform.
var Root = findRoot()

// oracleDir holds the pinned PowerShell oracle scripts shared by all services.
var oracleDir = Root + "/tests/integration/lab/oracle"

var oracleScript = regexp.MustCompile(`^[a-z][a-z-]*\.ps1$`)

// OracleCommand prepares pwsh to run one pinned oracle script, named by its
// basename in tests/integration/lab/oracle. Callers set the environment,
// stdin and output buffer. A name that is not a plain script basename resolves
// to a script that does not exist, so the oracle fails rather than running it.
func OracleCommand(ctx context.Context, script string) *exec.Cmd {
	if !oracleScript.MatchString(script) {
		script = "invalid-oracle-script-name.ps1"
	}
	//nolint:gosec // G204: pwsh runs a validated script basename from the repository oracle directory.
	command := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", oracleDir+"/"+script)
	return command
}

func findRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && info.Mode().IsRegular() {
			return filepath.ToSlash(dir)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}
