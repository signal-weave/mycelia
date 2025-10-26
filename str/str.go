// * String handling utilities

package str

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"mycelia/globals"

	"golang.org/x/term"
)

func SprintfLn(formatStr string, args ...string) {
	interfaceArgs := make([]any, len(args))
	for i, v := range args {
		interfaceArgs[i] = v
	}
	msg := fmt.Sprintf(formatStr, interfaceArgs...)
	fmt.Println(msg)
}

func ActionPrint(s string) {
	if globals.Verbosity < globals.VertAct {
		return
	}
	fmt.Println("[ACTION] - " + s)
}

func WarningPrint(s string) {
	if globals.Verbosity < globals.VerbWrn {
		return
	}
	fmt.Println("[WARNING] - " + s)
}

func ErrorPrint(s string) {
	if globals.Verbosity < globals.VerbErr {
		return
	}
	fmt.Println("[ERROR] - " + s)
}

// Returns the current terminal width if it can be found else 80.
func getOutputWidth() int {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		width, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil {
			return width
		}
	}
	return globals.DefaultTerminalW
}

// PrintAsciiLine prints "-" repeated to fill the terminal length if a terminal
// is being used for Stdout, otherwise repeats 80 columns wide.
func PrintAsciiLine() {
	width := getOutputWidth()
	fmt.Println(strings.Repeat("-", width))
}

// PrintCenteredHeader prints a "-----header-----" to fill the terminal length
// if a terminal is being used for Stdout, otherwise prints 80 columns wide.
func PrintCenteredHeader(header string) {
	width := getOutputWidth()
	vis := utf8.RuneCountInString(header)
	spacer := (width - vis) / 2
	side := strings.Repeat("-", spacer)
	SprintfLn("%s%s%s", side, header, side)
}
