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
	if globals.Verbosity < globals.VERB_ACT {
		return
	}
	fmt.Println("[ACTION] - " + s)
}

func WarningPrint(s string) {
	if globals.Verbosity < globals.VERB_WRN {
		return
	}
	fmt.Println("[WARNING] - " + s)
}

func ErrorPrint(s string) {
	if globals.Verbosity < globals.VERB_ERR {
		return
	}
	fmt.Println("[ERROR] - " + s)
}

func PrintByVerbosity(s string, v int) {
	if globals.Verbosity == globals.VERB_NIL {
		return
	}
	if v == globals.VERB_ACT && globals.Verbosity >= v {
		ActionPrint(s)
	}
	if v == globals.VERB_WRN && globals.Verbosity >= v {
		WarningPrint(s)
	}
	if v == globals.VERB_ERR && globals.Verbosity >= v {
		ErrorPrint(s)
	}
}

// Returns the current terminal width if it can be found else 80.
func getOutputWidth() int {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		width, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil {
			return width
		}
	}
	return globals.DEFAULT_TERMINAL_W
}

// Prints "-" repeated to fill the terminal length if a terminal is being used
// for Stdout, otherwise repeats 80 columns wide.
func PrintAsciiLine() {
	width := getOutputWidth()
	fmt.Println(strings.Repeat("-", width))
}

// Prints a "-----header-----" to fill the terminal length if a terminal is
// being used for Stdout, otherwise prints 80 columns wide.
func PrintCenteredHeader(header string) {
	width := getOutputWidth()
	vis := utf8.RuneCountInString(header)
	spacer := (width - vis) / 2
	side := strings.Repeat("-", spacer)
	SprintfLn("%s%s%s", side, header, side)
}
