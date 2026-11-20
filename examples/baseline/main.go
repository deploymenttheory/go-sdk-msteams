// Command baseline prints the targeted upstream version without authentication.
package main

import (
	"fmt"

	msteams "github.com/deploymenttheory/go-sdk-msteams"
)

func main() {
	fmt.Println("Target MicrosoftTeams PowerShell baseline:", msteams.PowerShellBaseline)
}
