package teamsshiftsapppolicyacc

import (
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func TestAccShiftsAppPolicyUpdateRestore(t *testing.T) {
	labkit.RunSingletonUpdate(t, shiftsAppSingleton)
}
