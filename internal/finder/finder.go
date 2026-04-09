// Package finder provides fuzzy directory matching for Windows drives.
package finder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// systemDirs contains directory names that should be skipped during traversal.
var systemDirs = map[string]bool{
	"$Recycle.Bin":            true,
	"System Volume Information": true,
	"Recovery":                 true,
	"PerfLogs":                 true,
	"Documents and Settings":   true,
}

// FindDirs searches for directories under the given drive that match all patterns.
// Patterns are matched level by level: the first pattern matches against top-level
// directories under the drive, the second pattern matches within those results, etc.
// Matching is case-insensitive substring matching.
func FindDirs(drive string, patterns []string) ([]string, error) {
	if len(drive) == 0 {
		return nil, fmt.Errorf("drive letter is required")
	}

	// Normalize drive: "d" -> "D:\", "d:" -> "D:\"
	letter := strings.ToUpper(string(drive[0]))
	basePath := letter + ":\\"

	// Validate drive exists
	if _, err := os.Stat(basePath); err != nil {
		return nil, fmt.Errorf("drive %s is not accessible: %w", letter+":", err)
	}

	if len(patterns) == 0 {
		return nil, fmt.Errorf("at least one search pattern is required")
	}

	// Start with top-level directories under the drive
	var currentDirs []string
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read drive %s: %w", letter+":", err)
	}
	for _, entry := range entries {
		if entry.IsDir() && !isSystemOrHidden(entry.Name()) {
			currentDirs = append(currentDirs, filepath.Join(basePath, entry.Name()))
		}
	}

	// Match each pattern level by level
	for _, pattern := range patterns {
		var matched []string
		patLower := strings.ToLower(pattern)

		for _, dir := range currentDirs {
			base := filepath.Base(dir)
			if strings.Contains(strings.ToLower(base), patLower) {
				// If this is the last pattern, collect the directory itself
				if pattern == patterns[len(patterns)-1] {
					matched = append(matched, dir)
				} else {
					// Expand into subdirectories for next level matching
					subs, err := listDirs(dir)
					if err != nil {
						continue
					}
					matched = append(matched, subs...)
				}
			}
		}
		currentDirs = matched
		if len(currentDirs) == 0 {
			break
		}
	}

	return currentDirs, nil
}

// listDirs returns all non-hidden, non-system subdirectories of the given path.
func listDirs(dir string) ([]string, error) {
	var dirs []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() && !isSystemOrHidden(entry.Name()) {
			dirs = append(dirs, filepath.Join(dir, entry.Name()))
		}
	}
	return dirs, nil
}

// isSystemOrHidden returns true for directories that should be skipped.
func isSystemOrHidden(name string) bool {
	if len(name) > 0 && name[0] == '.' {
		return true
	}
	return systemDirs[name]
}
