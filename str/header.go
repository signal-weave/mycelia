package str

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"mycelia/globals"
)

var line1 = "███╗   ███╗██╗   ██╗ ██████╗███████╗██╗     ██╗ █████╗"
var line2 = "████╗ ████║╚██╗ ██╔╝██╔════╝██╔════╝██║     ██║██╔══██╗"
var line3 = "██╔████╔██║ ╚████╔╝ ██║     █████╗  ██║     ██║███████║"
var line4 = "██║╚██╔╝██║  ╚██╔╝  ██║     ██╔══╝  ██║     ██║██╔══██║"
var line5 = "██║ ╚═╝ ██║   ██║   ╚██████╗███████╗███████╗██║██║  ██║"
var line6 = "╚═╝     ╚═╝   ╚═╝    ╚═════╝╚══════╝╚══════╝╚═╝╚═╝  ╚═╝"

var lines = []string{line1, line2, line3, line4, line5, line6}

var producedBy = fmt.Sprintf("A %s product.", globals.Developer)
var disclaimer = "Mycelia is a work-in-progress concurrent message broker."

func printHeader() {
	width := getOutputWidth()
	vis := utf8.RuneCountInString(line1)
	spacer := (width - vis) / 2
	prefix := strings.Repeat(" ", spacer)
	for _, v := range lines {
		fmt.Println(prefix, v)
	}
}

func printVersion(version string) {
	verString := fmt.Sprintf("Running verison: %s", version)
	fmt.Println(verString)
}

func PrintStartupText(version string) {
	PrintAsciiLine()
	printHeader()
	PrintAsciiLine()
	fmt.Println(producedBy)
	printVersion(version)
	PrintAsciiLine()
	fmt.Println(disclaimer)
	PrintAsciiLine()
}
