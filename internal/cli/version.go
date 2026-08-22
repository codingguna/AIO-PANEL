package cli

import (
	"fmt"

	"github.com/codingguna/aio-panel/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the AIO-PANEL version and build information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.String())
	},
}
