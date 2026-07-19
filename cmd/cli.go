package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	"ads-systweak/pkg/analytics"
	"ads-systweak/pkg/state"
	"ads-systweak/pkg/tweaks"

	"github.com/spf13/cobra"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(statusCmd)

	// Single tweaks manipulation
	rootCmd.AddCommand(tweakApplyCmd)
	rootCmd.AddCommand(tweakRevertCmd)

	// State-driven staging
	rootCmd.AddCommand(planCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(preapproveCmd)

	// Presets
	presetCmd.AddCommand(presetListCmd)
	presetCmd.AddCommand(presetStageCmd)
	rootCmd.AddCommand(presetCmd)

	// Native Helpers
	rootCmd.AddCommand(speedtestCmd)
	rootCmd.AddCommand(caffeinateCmd)
}

func findTweak(id string) tweaks.Tweak {
	id = strings.ToLower(id)
	for _, t := range tweaks.Registry {
		if strings.ToLower(t.ID()) == id {
			return t
		}
	}
	return nil
}

// ========================
// CLI Direct Tweak Manip
// ========================

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available tweaks",
	Run: func(cmd *cobra.Command, args []string) {
		analytics.TrackAppLaunched("cli")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tCATEGORY\tNAME\tSTATUS")
		fmt.Fprintln(w, "--\t--------\t----\t------")
		for _, t := range tweaks.Registry {
			probe := t.Probe()
			status := string(probe.State)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", t.ID(), t.Category(), t.Name(), status)
		}
		w.Flush()
	},
}

var tweakApplyCmd = &cobra.Command{
	Use:   "force-apply [id]",
	Short: "Immediately force-apply a specific tweak by ID (bypasses state)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		t := findTweak(id)
		if t == nil {
			fmt.Printf("Tweak '%s' not found.\n", id)
			os.Exit(1)
		}
		fmt.Printf("Applying tweak: %s...\n", t.Name())
		if err := t.Apply(); err != nil {
			fmt.Printf("Failed to apply tweak: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Success.")
	},
}

var tweakRevertCmd = &cobra.Command{
	Use:   "force-revert [id]",
	Short: "Immediately force-revert a specific tweak by ID (bypasses state)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		t := findTweak(id)
		if t == nil {
			fmt.Printf("Tweak '%s' not found.\n", id)
			os.Exit(1)
		}
		fmt.Printf("Reverting tweak: %s...\n", t.Name())
		if err := t.Revert(); err != nil {
			fmt.Printf("Failed to revert tweak: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Success.")
	},
}

var statusCmd = &cobra.Command{
	Use:   "status [id]",
	Short: "Check the true operational status of a specific tweak",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		t := findTweak(id)
		if t == nil {
			fmt.Printf("Tweak '%s' not found.\n", id)
			os.Exit(1)
		}
		probe := t.Probe()
		if probe.State != tweaks.ProbeApplied && probe.State != tweaks.ProbeOff {
			fmt.Printf("Status: %s\n", probe.State)
			fmt.Printf("Error checking status: %v\n", probe.Err)
			os.Exit(1)
		}
		if probe.Applied {
			fmt.Println("Status: " + colorGreen + "Applied" + colorReset)
		} else {
			fmt.Println("Status: " + colorYellow + "Default/Off" + colorReset)
		}
	},
}

// ========================
// State-Driven CLI
// ========================

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "View staged changes based on desired config vs actual system state",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := state.LoadConfig()
		if err != nil {
			fmt.Printf("Failed to load configuration: %v\n", err)
			return
		}
		if len(cfg.DesiredState) == 0 {
			fmt.Println("No desired state configured. Use GUI to stage changes or 'preset stage <name>'.")
			return
		}

		fmt.Println("Execution Plan:")
		fmt.Println("----------------------------------------")
		plan := tweaks.BuildPlan(tweaks.Registry, cfg.DesiredState)
		for _, item := range plan {
			action := colorGreen + "Apply" + colorReset
			if item.Action == tweaks.PlanRevert {
				action = colorYellow + "Revert" + colorReset
			} else if item.Action == tweaks.PlanBlocked {
				action = colorRed + "Blocked: " + string(item.Probe.State) + colorReset
				if item.Probe.Err != nil {
					action += " (" + item.Probe.Err.Error() + ")"
				}
			}
			fmt.Printf("[%s] %s -> %s\n", item.Tweak.Category(), item.Tweak.Name(), action)
		}

		if len(plan) == 0 {
			fmt.Println("System matches desired configuration state. No pending changes.")
		} else {
			fmt.Printf("----------------------------------------\nTotal Pending/Blocked Changes: %d\n", len(plan))
			fmt.Println("Run 'ads-systweak apply' to execute.")
		}
	},
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply all staged changes from the plan",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := state.LoadConfig()
		if err != nil {
			fmt.Printf("Failed to load configuration: %v\n", err)
			return
		}

		plan := tweaks.BuildPlan(tweaks.Registry, cfg.DesiredState)
		reader := bufio.NewReader(os.Stdin)

		for _, item := range plan {
			tw := item.Tweak
			if item.Action == tweaks.PlanBlocked {
				fmt.Printf(colorRed+"BLOCKED [%s]: state is %s: %v"+colorReset+"\n", tw.Name(), item.Probe.State, item.Probe.Err)
				continue
			}
			action := "Apply"
			if item.Action == tweaks.PlanRevert {
				action = "Revert"
			}

			preApproved := state.IsPreApproved(cfg, tw.ID(), string(tw.Category()))

			if !preApproved {
				fmt.Printf("Tweak [%s] is pending: %s. Execute? [y/N]: ", tw.Name(), action)
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(strings.ToLower(response))
				if response != "y" && response != "yes" {
					fmt.Println("Skipping.")
					continue
				}
			} else {
				fmt.Printf("Auto-approving (pre-approved) [%s]: %s\n", tw.Name(), action)
			}

			var applyErr error
			if item.Action == tweaks.PlanApply {
				applyErr = tw.Apply()
				if applyErr == nil {
					analytics.TrackTweakApplied(tw.ID(), string(tw.Category()))
				}
			} else {
				applyErr = tw.Revert()
				if applyErr == nil {
					analytics.TrackTweakReverted(tw.ID(), string(tw.Category()))
				}
			}

			if applyErr != nil {
				fmt.Printf(colorRed+"ERROR: %v"+colorReset+"\n", applyErr)
			} else {
				fmt.Println(colorGreen + "SUCCESS." + colorReset)
			}
		}

		if len(plan) == 0 {
			fmt.Println("No pending changes to apply.")
		}
	},
}

var preapproveCmd = &cobra.Command{
	Use:   "preapprove [id/category]",
	Short: "Add a tweak ID or Category to the auto-apply pre-approval list",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := state.LoadConfig()
		if err != nil {
			fmt.Printf("Failed to load configuration: %v\n", err)
			return
		}
		target := args[0]
		if !state.IsPreApproved(cfg, target, "") {
			cfg.PreApproved = append(cfg.PreApproved, target)
			if err := state.SaveConfig(cfg); err != nil {
				fmt.Printf("Failed to save configuration: %v\n", err)
				return
			}
			fmt.Printf("Added '%s' to pre-approval whitelist.\n", target)
		} else {
			fmt.Printf("'%s' is already pre-approved.\n", target)
		}
	},
}

// ========================
// Presets
// ========================

var presetCmd = &cobra.Command{
	Use:   "preset",
	Short: "Manage optimization preset bundles",
}

var presetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available presets",
	Run: func(cmd *cobra.Command, args []string) {
		for _, p := range tweaks.Presets {
			fmt.Printf("- %s (ID: %s)\n  %s\n\n", p.Name, strings.ReplaceAll(strings.ToLower(p.Name), " ", "-"), p.Description)
		}
	},
}

var presetStageCmd = &cobra.Command{
	Use:   "stage [name/id]",
	Short: "Stage an optimization preset into the configuration state",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.ToLower(args[0])
		var match *tweaks.Preset
		for _, p := range tweaks.Presets {
			id := strings.ReplaceAll(strings.ToLower(p.Name), " ", "-")
			if id == query || strings.ToLower(p.Name) == query {
				match = &p
				break
			}
		}

		if match == nil {
			fmt.Printf("Preset '%s' not found.\n", args[0])
			os.Exit(1)
		}

		cfg, err := state.LoadConfig()
		if err != nil {
			fmt.Printf("Failed to load configuration: %v\n", err)
			return
		}

		fmt.Printf("Staging preset: %s...\n", match.Name)
		for _, tID := range match.TweakIDs {
			cfg.DesiredState[tID] = true
			fmt.Printf(" -> staged %s = true\n", tID)
		}
		if err := state.SaveConfig(cfg); err != nil {
			fmt.Printf("Failed to save configuration: %v\n", err)
			return
		}
		analytics.TrackPresetStaged(match.Name, len(match.TweakIDs))
		fmt.Println("Preset staged successfully. Run 'ads-systweak plan' to review or 'ads-systweak apply' to execute.")
	},
}

// ========================
// Native CLI Helpers
// ========================

var speedtestCmd = &cobra.Command{
	Use:   "speedtest",
	Short: "Run Apple's native networkQuality tool to test connection speed",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting native macOS networkQuality speedtest...")
		c := exec.Command("networkQuality")
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			fmt.Printf("Speedtest exited: %v\n", err)
			os.Exit(1)
		}
	},
}

var caffeinateCmd = &cobra.Command{
	Use:   "caffeinate",
	Short: "Prevent the system and display from sleeping (Press Ctrl+C to exit)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Caffeine mode activated. System and display will not sleep.")
		fmt.Println("Press Ctrl+C to stop.")
		c := exec.Command("caffeinate", "-dimsu")
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		_ = c.Run()
		fmt.Println("\nCaffeine mode deactivated.")
	},
}
