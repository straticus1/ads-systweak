package cmd

import (
	"ads-systweak/pkg/analytics"
	"ads-systweak/pkg/tweaks"
	"fmt"
	"os"
	"runtime"

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
	if warning := platformWarning(runtime.GOOS); warning != "" {
		fmt.Fprintln(os.Stderr, warning)
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

func platformWarning(goos string) string {
	if goos == "darwin" {
		return ""
	}
	return fmt.Sprintf("Warning: ads-systweak requires macOS; current operating system is %s. Mutating commands will be unsupported.", goos)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
