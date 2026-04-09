// jumpd - Quickly jump to directories on Windows drives with fuzzy matching.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/scm1219/jumpd/cmd"
)

func main() {
	// Show help when no args or help flag
	if len(os.Args) <= 1 {
		cmd.Execute()
		return
	}
	for _, a := range os.Args[1:] {
		if a == "--help" || a == "-h" || a == "--wrapper" {
			cmd.Execute()
			return
		}
	}

	// Always open a new CMD window for the interactive selection.
	// Write commands to a temp .bat file to avoid inline quoting issues with cmd.exe.
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	tmpPath := filepath.Join(os.TempDir(), "jumpd_path.txt")
	tmpBat := filepath.Join(os.TempDir(), "jumpd_run.bat")
	os.Remove(tmpPath)
	os.Remove(tmpBat)

	args := os.Args[1:]
	// Use %%%%i in Sprintf to produce %%i in the bat file (Go %% -> %, so %%%% -> %%)
	bat := fmt.Sprintf(
		"@echo off\r\n\"%s\" --wrapper %s\r\nif exist \"%s\" (for /f \"usebackq\" %%%%i in (\"%s\") do cd /d \"%%%%i\") else exit\r\n",
		exe, strings.Join(args, " "), tmpPath, tmpPath,
	)
	if err := os.WriteFile(tmpBat, []byte(bat), 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	c := exec.Command("cmd.exe", "/c", "start", "", "cmd.exe", "/k", tmpBat)
	_ = c.Start()
}
