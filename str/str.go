// * String handling utilities

package str

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"mycelia/errgo"
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
