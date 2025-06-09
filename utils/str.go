// * String handling utilities

package utils

import (
	"fmt"
	"strings"
)

func SprintfLn(formatStr string, args ...string) {
	msg := fmt.Sprintf(formatStr, args)
	fmt.Println(msg)
}

// Same as utils.SprintfLn but will add "  - " to the start of the string with
// spaces equal to the indent number.
func SprintfLnIndent(formatStr string, indent int, args ...string) {
	imsg := strings.Repeat(" ", indent) + "- " + formatStr

	var anyArgs []any = make([]any, len(args))
	for i, v := range args {
		anyArgs[i] = v
	}

	msg := fmt.Sprintf(imsg, anyArgs...)
	fmt.Println(msg)
}

func DebugPrintLn(s string) {
	msg := fmt.Sprintf("[DEBUG] %s", s)
	fmt.Println(msg)
}
