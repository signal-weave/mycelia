// * String handling utilities

package str

import (
	"fmt"

	"mycelia/environ"
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
	if environ.GetVerbosityLevel() < 3 {
		return
	}
	fmt.Println("[ACTION] - " + s)
}

func WarningPrint(s string) {
	if environ.GetVerbosityLevel() < 2 {
		return
	}
	fmt.Println("[WARNING] - " + s)
}

func ErrorPrint(s string) {
	if environ.GetVerbosityLevel() < 1 {
		return
	}
	fmt.Println("[ERROR] - " + s)
}

func DebugPrintLn(s string) {
	msg := fmt.Sprintf("[DEBUG] - %s", s)
	fmt.Println(msg)
}
