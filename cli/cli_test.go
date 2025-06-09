package cli

import (
	"flag"
	"os"
	"testing"
)

func TestParseCLIArgs(t *testing.T) {
	// Save original os.Args and reset after test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Args = []string{"cmd", "-address=10.0.0.1", "-port=9000"}

	ParseCLIArgs()

	if Address != "10.0.0.1" {
		t.Errorf("expected Address to be '10.0.0.1', got '%s'", Address)
	}
	if Port != 9000 {
		t.Errorf("expected Port to be 9000, got %d", Port)
	}
}
