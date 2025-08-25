// * String handling utilities

package str

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"mycelia/errgo"

	"golang.org/x/term"
)

func getVerbosity() int {
	s := os.Getenv("VERBOSITY")
	i := errgo.ValueOrPanic(strconv.Atoi(s))
	return i
}

func SprintfLn(formatStr string, args ...string) {
	interfaceArgs := make([]interface{}, len(args))
	for i, v := range args {
		interfaceArgs[i] = v
	}
	msg := fmt.Sprintf(formatStr, interfaceArgs...)
	fmt.Println(msg)
}

func ActionPrint(s string) {
	if getVerbosity() < 3 {
		return
	}
	fmt.Println("[ACTION] - " + s)
}

func WarningPrint(s string) {
	if getVerbosity() < 2 {
		return
	}
	fmt.Println("[WARNING] - " + s)
}

func ErrorPrint(s string) {
	if getVerbosity() < 1 {
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

// Prints "-" repeated to fill the terminal lenght if a terminal is being used
// for Stdout, otherwise repeats 80 times.
func PrintAsciiLine() {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		width, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			fmt.Println(strings.Repeat("-", 80))
			return
		} else {
			fmt.Println(strings.Repeat("-", width))
		}
	} else {
		fmt.Println(strings.Repeat("-", 80))
	}
}
