// Package ui provides interactive directory selection with paging.
package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const pageSize = 9

// SelectPath presents an interactive selection UI for the given paths.
// Returns the selected path, or empty string if the user quits.
func SelectPath(paths []string) string {
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "No matching directories found.")
		return ""
	}

	if len(paths) == 1 {
		fmt.Println(paths[0])
		return paths[0]
	}

	totalPages := (len(paths) + pageSize - 1) / pageSize
	page := 0

	reader := bufio.NewReader(os.Stdin)

	for {
		displayPage(paths, page, totalPages)

		fmt.Print("\nEnter choice [1-9/n/p/q]: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "\nInput error:", err)
			return ""
		}

		choice := strings.TrimSpace(input)
		if choice == "" {
			continue
		}

		// Quit
		if strings.EqualFold(choice, "q") || strings.EqualFold(choice, "quit") {
			return ""
		}

		// Next page
		if strings.EqualFold(choice, "n") || strings.EqualFold(choice, "next") {
			if page < totalPages-1 {
				page++
			} else {
				fmt.Fprintln(os.Stderr, "Already on the last page.")
			}
			continue
		}

		// Previous page
		if strings.EqualFold(choice, "p") || strings.EqualFold(choice, "prev") {
			if page > 0 {
				page--
			} else {
				fmt.Fprintln(os.Stderr, "Already on the first page.")
			}
			continue
		}

		// Number selection
		num, err := strconv.Atoi(choice)
		if err != nil || num < 1 || num > pageSize {
			fmt.Fprintln(os.Stderr, "Invalid input. Enter 1-9, n (next), p (prev), or q (quit).")
			continue
		}

		start := page * pageSize
		end := start + pageSize
		if end > len(paths) {
			end = len(paths)
		}

		idx := start + num - 1
		if idx < end {
			fmt.Println(paths[idx])
			return paths[idx]
		}

		fmt.Fprintln(os.Stderr, "Index out of range on this page.")
	}
}

// displayPage renders one page of results.
func displayPage(paths []string, page, totalPages int) {
	// Clear screen (Windows compatible)
	fmt.Print("\033[2J\033[H")

	start := page * pageSize
	end := start + pageSize
	if end > len(paths) {
		end = len(paths)
	}

	fmt.Printf("=== jumpd results (page %d/%d, %d total) ===\n\n", page+1, totalPages, len(paths))

	for i := start; i < end; i++ {
		num := i - start + 1
		fmt.Printf("  %d. %s\n", num, paths[i])
	}
}
