package cmd

import (
	"ads-systweak/pkg/analytics"
	"ads-systweak/pkg/tweaks"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var webFlag bool
var versionFlag bool
var dryRunFlag bool
var verboseFlag bool

const Version = "1.0.0"

var rootCmd = &cobra.Command{
	Use:   "ads-systweak",
	Short: "After Dark System Tools - macOS optimization utility",
	Long:  "After Dark System Tools - Unlock your Mac's hidden potential\n\nA free, open-source utility to tweak macOS system settings, available via CLI and Web UI.\nBy After Dark - https://afterdarksys.com\n\nRun 'ads-systweak web' or 'ads-systweak --web' to launch the web interface.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		tweaks.DryRun = dryRunFlag
		tweaks.Verbose = verboseFlag
	},
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			fmt.Printf("ads-systweak version %s\n", Version)
			return
		}
		if webFlag {
			runWeb()
			return
		}
		cmd.Help()
	},
}

func init() {
	// macOS architecture check
	if os.Getenv("GOOS") != "" && os.Getenv("GOOS") != "darwin" {
		fmt.Fprintf(os.Stderr, "Warning: ads-systweak is designed specifically for macOS (darwin).\nYou are currently building/running for %s. Tweaks may not function correctly.\n", os.Getenv("GOOS"))
	}

	rootCmd.PersistentFlags().BoolVar(&dryRunFlag, "dry-run", false, "Simulate changes without modifying the system")
	rootCmd.PersistentFlags().BoolVar(&verboseFlag, "verbose", false, "Enable verbose output to see shell commands being executed")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Print the version number")
	rootCmd.Flags().BoolVar(&webFlag, "web", false, "Launch the web user interface")

	// Initialize analytics
	analyticsEnabled := os.Getenv("ADS_ANALYTICS_ENABLED") == "true"
	analyticsEndpoint := os.Getenv("ADS_ANALYTICS_ENDPOINT")
	if analyticsEndpoint == "" {
		analyticsEndpoint = "https://analytics.afterdarksys.com/events"
	}
	analytics.Initialize(analyticsEndpoint, analyticsEnabled)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
