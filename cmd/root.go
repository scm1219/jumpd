// Package cmd contains the root command for jumpd.
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/scm1219/jumpd/internal/finder"
	"github.com/scm1219/jumpd/internal/ui"
	"github.com/spf13/cobra"
)

// wrapperMode is set when jumpd is relaunched inside a CMD window by main.go.
// In this mode, the selected path is written to a temp file instead of opening a new window.
var wrapperMode bool

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "jumpd <drive> [pattern1] [pattern2] ...",
	Short: "Quickly jump to directories on Windows drives with fuzzy matching",
	Long: `jumpd is a Windows CLI tool for quickly navigating to directories
using fuzzy path matching. Select a directory and it opens a new CMD window
at that location.

USAGE:
  jumpd <drive> [pattern1] [pattern2] ...

  The first argument is the drive letter (e.g., "d" or "d:").
  Remaining arguments are directory name patterns matched level by level.

EXAMPLES:
  jumpd d tools            # List dirs containing "tools" under D:\
  jumpd d tools pickyou    # List dirs containing "pickyou" under D:\*tools*\

INTERACTIVE CONTROLS:
  1-9    Select directory by index (opens new CMD window)
  n/N    Next page
  p/P    Previous page
  q/Q    Quit
`,
	Args: cobra.MinimumNArgs(1),
	RunE: runJumpd,
}

func runJumpd(cmd *cobra.Command, args []string) error {
	// Parse drive letter from first argument
	drive := strings.ToUpper(strings.TrimSuffix(args[0], ":"))
	if len(drive) != 1 || drive[0] < 'A' || drive[0] > 'Z' {
		return fmt.Errorf("invalid drive letter: %q (expected A-Z)", args[0])
	}

	// Remaining arguments are search patterns
	patterns := args[1:]
	if len(patterns) == 0 {
		return fmt.Errorf("at least one search pattern is required\n\nUsage: jumpd <drive> <pattern1> [pattern2] ...")
	}

	// Find matching directories
	results, err := finder.FindDirs(drive, patterns)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "No matching directories found.")
		return nil
	}

	// Interactive selection
	selected := ui.SelectPath(results)
	if selected == "" {
		return nil
	}

	if wrapperMode {
		// Write path to temp file, the wrapper CMD will cd to it
		tmpPath := filepath.Join(os.TempDir(), "jumpd_path.txt")
		_ = os.WriteFile(tmpPath, []byte(selected), 0644)
		return nil
	}

	// Normal mode: open new CMD window at the selected directory
	c := exec.Command("cmd.exe", "/c", "start", "/d", selected, "", "cmd.exe", "/k")
	c.Start()
	return nil
}

// Execute runs the root command.
func Execute() {
	rootCmd.Flags().BoolVar(&wrapperMode, "wrapper", false, "internal: wrapper mode for cmd integration")
	_ = rootCmd.Flags().MarkHidden("wrapper")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(rootCmd.ErrOrStderr(), err)
	}
}
