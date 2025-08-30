package str

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"mycelia/globals"
)

var line1 string = "███╗   ███╗██╗   ██╗ ██████╗███████╗██╗     ██╗ █████╗"
var line2 string = "████╗ ████║╚██╗ ██╔╝██╔════╝██╔════╝██║     ██║██╔══██╗"
var line3 string = "██╔████╔██║ ╚████╔╝ ██║     █████╗  ██║     ██║███████║"
var line4 string = "██║╚██╔╝██║  ╚██╔╝  ██║     ██╔══╝  ██║     ██║██╔══██║"
var line5 string = "██║ ╚═╝ ██║   ██║   ╚██████╗███████╗███████╗██║██║  ██║"
var line6 string = "╚═╝     ╚═╝   ╚═╝    ╚═════╝╚══════╝╚══════╝╚═╝╚═╝  ╚═╝"

var lines []string = []string{line1, line2, line3, line4, line5, line6}

var producedBy string = fmt.Sprintf("A %s product.", globals.Developer)
var disclaimer string = "Mycelia is a work-in-progress concurrent message broker."

func printHeader() {
	width := getOutputWidth()
	vis := utf8.RuneCountInString(line1)
	spacer := (width - vis) / 2
	prefix := strings.Repeat(" ", spacer)
	for _, v := range lines {
		fmt.Println(prefix, v)
	}
}

func printVersion(major, minor, patch int) {
	brokerVersion := fmt.Sprintf("%d.%d.%d", major, minor, patch)
	verString := fmt.Sprintf("Running verison: %s", brokerVersion)
	fmt.Println(verString)
}

func PrintStartupText(maj, min, patch int) {
	PrintAsciiLine()
	printHeader()
	PrintAsciiLine()
	fmt.Println(producedBy)
	printVersion(maj, min, patch)
	PrintAsciiLine()
	fmt.Println(disclaimer)
	PrintAsciiLine()
}
