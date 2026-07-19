package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Report macOS version and system-tool capabilities",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Runtime: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		if output, err := exec.Command("/usr/bin/sw_vers").Output(); err == nil {
			facts := parseSWVers(string(output))
			fmt.Printf("macOS: %s (build %s)\n", facts["ProductVersion"], facts["BuildVersion"])
		}
		checks := []string{
			"/usr/bin/defaults", "/usr/bin/kmutil", "/usr/bin/systemextensionsctl",
			"/usr/bin/networkQuality", "/usr/bin/wdutil", "/usr/sbin/networksetup",
			"/usr/bin/pmset", "/usr/bin/tmutil", "/usr/sbin/spctl",
		}
		for _, path := range checks {
			status := "available"
			if _, err := os.Stat(path); err != nil {
				status = "missing"
			}
			fmt.Printf("%-35s %s\n", path, status)
		}
	},
}

func init() { rootCmd.AddCommand(doctorCmd) }

func parseSWVers(output string) map[string]string {
	facts := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		key, value, found := strings.Cut(line, ":")
		if found {
			facts[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	return facts
}
