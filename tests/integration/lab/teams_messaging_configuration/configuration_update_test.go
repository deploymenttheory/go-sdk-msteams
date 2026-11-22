package teamsmessagingconfigurationacc

import (
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func TestAccMessagingConfigurationUpdateRestore(t *testing.T) {
	labkit.RunSingletonUpdate(t, messagingSingleton)
}
