package cli

import (
	"flag"
	"os"
	"testing"

	"mycelia/environ"
)

func TestParseCLIArgs(t *testing.T) {
	// Save and restore original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Reset the flag.CommandLine to allow re-parsing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Simulate CLI input
	os.Args = []string{"cmd", "-address=10.1.2.3", "-port=9999", "-verbosity=2"}

	// Clear the env var before testing
	os.Unsetenv(environ.VERBOSITY_ENV)

	// Run the function
	ParseCLIArgs()

	// Check parsed values
	if Address != "10.1.2.3" {
		t.Errorf("Expected Address to be '10.1.2.3', got '%s'", Address)
	}

	if Port != 9999 {
		t.Errorf("Expected Port to be 9999, got %d", Port)
	}

	expectedVerb := environ.VerbosityStatusMap[2]
	envVerb := os.Getenv(environ.VERBOSITY_ENV)
	if envVerb != expectedVerb {
		t.Errorf("Expected VERBOSITY_ENV to be '%s', got '%s'", expectedVerb, envVerb)
	}
}
