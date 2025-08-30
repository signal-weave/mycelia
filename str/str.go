// * String handling utilities

package str

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"mycelia/globals"

	"golang.org/x/term"
)

func SprintfLn(formatStr string, args ...string) {
	interfaceArgs := make([]interface{}, len(args))
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

func DebugPrintLn(s string) {
	msg := fmt.Sprintf("[DEBUG] - %s", s)
	fmt.Println(msg)
}

func PrettyPrintStrKeyJson(data map[string]any) {
	b, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		fmt.Println("Could not pretty print json data")
		return
	}
	fmt.Println(string(b))
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

// Prints "-" repeated to fill the terminal lenght if a terminal is being used
// for Stdout, otherwise repeats 80 times.
func PrintAsciiLine() {
	width := getOutputWidth()
	fmt.Println(strings.Repeat("-", width))
}
