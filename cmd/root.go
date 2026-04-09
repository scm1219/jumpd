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

// explorerMode is set via -e/--explorer flag.
// In wrapper mode, opens the selected directory in Windows Explorer instead of cd.
var explorerMode bool

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
  ← →    Previous / Next page
  1-9    Select directory by index
  g+N    Jump to page N (e.g. g3 for page 3)
  Esc    Cancel page jump input
  q      Quit
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
		fmt.Fprintf(cmd.ErrOrStderr(), "No matching directories found under %s:\\ for pattern(s): %s\n", drive, strings.Join(patterns, " > "))
		return nil
	}

	// Interactive selection
	selected := ui.SelectPath(results)
	if selected == "" {
		return nil
	}

	if wrapperMode {
		if explorerMode {
			// Open Windows Explorer at the selected directory, then exit the CMD
			c := exec.Command("explorer", selected)
			_ = c.Start()
		} else {
			// Write path to temp file, the wrapper CMD will cd to it
			tmpPath := filepath.Join(os.TempDir(), "jumpd_path.txt")
			_ = os.WriteFile(tmpPath, []byte(selected), 0644)
		}
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
	rootCmd.Flags().BoolVarP(&explorerMode, "explorer", "e", false, "open selected directory in Windows Explorer")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(rootCmd.ErrOrStderr(), err)
	}
}
