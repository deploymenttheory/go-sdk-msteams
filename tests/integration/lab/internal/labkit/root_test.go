package labkit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnitRootLocatesTheModuleAndOracleScripts(t *testing.T) {
	t.Parallel()
	if info, err := os.Stat(filepath.Join(Root, "go.mod")); err != nil || !info.Mode().IsRegular() {
		t.Fatal("repository root does not contain go.mod")
	}
	if info, err := os.Stat(filepath.Join(oracleDir, "read-team.ps1")); err != nil || !info.Mode().IsRegular() {
		t.Fatal("shared oracle directory does not contain the team oracle")
	}
}

func TestUnitOracleCommandRunsOnlyScriptBasenames(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, script, want string }{
		{"basename", "read-team.ps1", "/read-team.ps1"},
		{"parent traversal", "../read-team.ps1", "/invalid-oracle-script-name.ps1"},
		{"absolute", "/tmp/read-team.ps1", "/invalid-oracle-script-name.ps1"},
		{"other extension", "read-team.sh", "/invalid-oracle-script-name.ps1"},
		{"empty", "", "/invalid-oracle-script-name.ps1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			command := OracleCommand(t.Context(), tc.script)
			last := command.Args[len(command.Args)-1]
			if !strings.HasPrefix(last, oracleDir+"/") || !strings.HasSuffix(last, tc.want) {
				t.Fatal("oracle command selected an unexpected script path")
			}
		})
	}
}
