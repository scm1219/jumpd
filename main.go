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

	// Check for -e/--explorer flag (support any position)
	hasExplorer := false
	var filteredArgs []string
	for _, a := range os.Args[1:] {
		if a == "-e" || a == "--explorer" {
			hasExplorer = true
		} else {
			filteredArgs = append(filteredArgs, a)
		}
	}

	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	tmpBat := filepath.Join(os.TempDir(), "jumpd_run.bat")
	os.Remove(tmpBat)

	var bat string
	if hasExplorer {
		// Explorer mode: after selection, CMD window exits automatically
		bat = fmt.Sprintf("@echo off\r\n\"%s\" --wrapper --explorer %s\r\nexit\r\n",
			exe, strings.Join(filteredArgs, " "))
	} else {
		// Default: cd into selected directory
		tmpPath := filepath.Join(os.TempDir(), "jumpd_path.txt")
		os.Remove(tmpPath)
		bat = fmt.Sprintf(
			"@echo off\r\n\"%s\" --wrapper %s\r\nif exist \"%s\" (for /f \"usebackq\" %%%%i in (\"%s\") do (cd /d \"%%%%i\" & for %%%%j in (\"%%%%~nxi\") do title %%%%j & cls)) else exit\r\n",
			exe, strings.Join(filteredArgs, " "), tmpPath, tmpPath,
		)
	}

	if err := os.WriteFile(tmpBat, []byte(bat), 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	c := exec.Command("cmd.exe", "/c", "start", "", "cmd.exe", "/k", tmpBat)
	_ = c.Start()
}
