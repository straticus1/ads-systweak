package cmd

import (
	"ads-systweak/pkg/api"

	"github.com/spf13/cobra"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Launch the local web user interface",
	Run: func(cmd *cobra.Command, args []string) {
		runWeb()
	},
}

func init() {
	rootCmd.AddCommand(webCmd)
}

func runWeb() {
	api.StartServer()
}
