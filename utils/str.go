// * String handling utilities

package utils

import (
	"fmt"
)

func SprintfLn(formatStr string, args ...string) {
	msg := fmt.Sprintf(formatStr, args)
	fmt.Println(msg)
}

func ActionPrint(s string) {
	msg := fmt.Sprintf("[ACTION] - %s", s)
	fmt.Println(msg)
}

func ErrorPrint(s string) {
	msg := fmt.Sprintf("[ERROR] - %s", s)
	fmt.Println(msg)
}

func WarningPrint(s string) {
	msg := fmt.Sprintf("[WARNING] - %s", s)
	fmt.Println(msg)
}

func DebugPrintLn(s string) {
	msg := fmt.Sprintf("[DEBUG] - %s", s)
	fmt.Println(msg)
}
