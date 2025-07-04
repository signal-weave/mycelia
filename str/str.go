// * String handling utilities

package str

import (
	"fmt"

	"mycelia/environ"
)

func SprintfLn(formatStr string, args ...string) {
	msg := fmt.Sprintf(formatStr, args)
	fmt.Println(msg)
}

func ActionPrint(s string) {
	if 1 > environ.GetVerbosityLevel() {
		return
	}
	msg := fmt.Sprintf("[ACTION] - %s", s)
	fmt.Println(msg)
}

func WarningPrint(s string) {
	if 2 > environ.GetVerbosityLevel() {
		return
	}
	msg := fmt.Sprintf("[WARNING] - %s", s)
	fmt.Println(msg)
}

func ErrorPrint(s string) {
	if 3 > environ.GetVerbosityLevel() {
		return
	}
	msg := fmt.Sprintf("[ERROR] - %s", s)
	fmt.Println(msg)
}

func DebugPrintLn(s string) {
	msg := fmt.Sprintf("[DEBUG] - %s", s)
	fmt.Println(msg)
}
