// Package ui provides interactive directory selection with paging.
package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

const pageSize = 9

// Windows console API constants
const (
	enableLineInput      uint32 = 0x0002
	enableEchoInput      uint32 = 0x0004
	enableProcessedInput uint32 = 0x0001
)

// Virtual key codes
const (
	vkLeft   = 0x25
	vkUp     = 0x26
	vkRight  = 0x27
	vkDown   = 0x28
	vkHome   = 0x24
	vkEnd    = 0x23
	vkReturn = 0x0D
	vkEscape = 0x1B
	vkBack   = 0x08
	vkC      = 0x43
)

// Windows console input types
const (
	keyEvent = 1
)

type inputRecord struct {
	EventType uint16
	_         [2]byte // padding
	Event     [16]byte
}

type keyEventRecord struct {
	KeyDown         int32
	RepeatCount     uint16
	VirtualKeyCode  uint16
	VirtualScanCode uint16
	Char            uint16
	ControlKeyState uint32
}

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procGetStdHandle           = kernel32.NewProc("GetStdHandle")
	procGetConsoleMode         = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode         = kernel32.NewProc("SetConsoleMode")
	procReadConsoleInputW      = kernel32.NewProc("ReadConsoleInputW")
)

const ctrlPressed uint32 = 0x0008

const stdInputHandle = ^uintptr(9) // -10

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

	// Switch console to raw mode for single-keypress reading
	stdin, err := getStdHandle()
	if err != nil {
		return selectPathFallback(paths)
	}

	oldMode, err := getConsoleMode(stdin)
	if err != nil {
		return selectPathFallback(paths)
	}

	rawMode := oldMode &^ (enableLineInput | enableEchoInput | enableProcessedInput)
	setConsoleMode(stdin, rawMode)
	defer setConsoleMode(stdin, oldMode)

	var jumpBuf strings.Builder
	jumpMode := false

	for {
		displayPage(paths, page, totalPages)

		if jumpMode {
			fmt.Printf("\n  Jump to page [%s_]: ", jumpBuf.String())
		} else {
			fmt.Print("\n  [← ↑ → ↓ page | Home/End | 1-9 select | g jump | q quit]: ")
		}

		ke, ok := readKeyEvent(stdin)
		if !ok || ke.KeyDown == 0 {
			continue
		}

		// Ctrl+C: quit immediately
		if ke.VirtualKeyCode == vkC && ke.ControlKeyState&ctrlPressed != 0 {
			return ""
		}

		ch := string(rune(ke.Char))

		if jumpMode {
			handleJump(ch, ke.VirtualKeyCode, &jumpBuf, &jumpMode, &page, totalPages)
			continue
		}

		switch {
		case ke.VirtualKeyCode == vkLeft, ke.VirtualKeyCode == vkUp:
			if page > 0 {
				page--
			} else {
				page = totalPages - 1
			}

		case ke.VirtualKeyCode == vkRight, ke.VirtualKeyCode == vkDown:
			if page < totalPages-1 {
				page++
			} else {
				page = 0
			}

		case ke.VirtualKeyCode == vkHome:
			page = 0

		case ke.VirtualKeyCode == vkEnd:
			page = totalPages - 1

		case ch == "q" || ch == "Q":
			return ""

		case ch == "g" || ch == "G":
			jumpMode = true
			jumpBuf.Reset()

		case ch >= "1" && ch <= "9":
			if selected := trySelect(paths, page, ch); selected != "" {
				return selected
			}
		}
	}
}

func getStdHandle() (uintptr, error) {
	ret, _, err := procGetStdHandle.Call(stdInputHandle)
	if ret == 0 {
		return 0, err
	}
	return ret, nil
}

func getConsoleMode(handle uintptr) (uint32, error) {
	var mode uint32
	ret, _, err := procGetConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode)))
	if ret == 0 {
		return 0, err
	}
	return mode, nil
}

func setConsoleMode(handle uintptr, mode uint32) error {
	ret, _, err := procSetConsoleMode.Call(handle, uintptr(mode))
	if ret == 0 {
		return err
	}
	return nil
}

func readKeyEvent(stdin uintptr) (*keyEventRecord, bool) {
	var ir inputRecord
	var read uint32
	ret, _, _ := procReadConsoleInputW.Call(stdin, uintptr(unsafe.Pointer(&ir)), 1, uintptr(unsafe.Pointer(&read)))
	if ret == 0 || read == 0 {
		return nil, false
	}
	if ir.EventType != keyEvent {
		return nil, false
	}
	return (*keyEventRecord)(unsafe.Pointer(&ir.Event[0])), true
}

// handleJump processes key events in jump-to-page mode.
func handleJump(ch string, vk uint16, buf *strings.Builder, jumpMode *bool, page *int, totalPages int) {
	switch {
	case vk == vkReturn:
		if buf.Len() > 0 {
			num, err := strconv.Atoi(buf.String())
			if err == nil && num >= 1 && num <= totalPages {
				*page = num - 1
			}
		}
		buf.Reset()
		*jumpMode = false

	case vk == vkEscape:
		buf.Reset()
		*jumpMode = false

	case vk == vkBack:
		if buf.Len() > 0 {
			runes := []rune(buf.String())
			buf.Reset()
			if len(runes) > 1 {
				buf.WriteString(string(runes[:len(runes)-1]))
			}
		} else {
			*jumpMode = false
		}

	case ch >= "0" && ch <= "9":
		buf.WriteString(ch)
	}
}

// trySelect attempts to select an item by number on the current page.
func trySelect(paths []string, page int, ch string) string {
	num := int(ch[0] - '0')
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
	return ""
}

// selectPathFallback provides line-based input when raw console mode is unavailable.
func selectPathFallback(paths []string) string {
	totalPages := (len(paths) + pageSize - 1) / pageSize
	page := 0

	reader := bufio.NewReader(os.Stdin)

	for {
		displayPage(paths, page, totalPages)

		fmt.Print("\nEnter choice [1-9/n/p/home/end/q]: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "\nInput error:", err)
			return ""
		}

		choice := strings.TrimSpace(input)
		if choice == "" {
			continue
		}

		if strings.EqualFold(choice, "q") || strings.EqualFold(choice, "quit") {
			return ""
		}

		if strings.EqualFold(choice, "n") || strings.EqualFold(choice, "next") {
			if page < totalPages-1 {
				page++
			} else {
				page = 0
			}
			continue
		}

		if strings.EqualFold(choice, "p") || strings.EqualFold(choice, "prev") {
			if page > 0 {
				page--
			} else {
				page = totalPages - 1
			}
			continue
		}

		if strings.EqualFold(choice, "home") {
			page = 0
			continue
		}

		if strings.EqualFold(choice, "end") {
			page = totalPages - 1
			continue
		}

		num, err := strconv.Atoi(choice)
		if err != nil || num < 1 || num > pageSize {
			fmt.Fprintln(os.Stderr, "Invalid input. Enter 1-9, n (next), p (prev), home, end, or q (quit).")
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
