package environ

import (
	"os"
	"testing"
)

func TestGetVerbosityLevel(t *testing.T) {
	testCases := []struct {
		envValue string
		expected int
	}{
		{"NONE", 0},
		{"ACTION", 1},
		{"WARNING", 2},
		{"ERROR", 3},
	}

	for _, tc := range testCases {
		t.Run(tc.envValue, func(t *testing.T) {
			os.Setenv(VERBOSITY_ENV, tc.envValue)
			actual := GetVerbosityLevel()
			if actual != tc.expected {
				t.Errorf("Expected %d for VERBOSITY=%s, got %d",
					tc.expected, tc.envValue, actual)
			}
		})
	}
}

func TestGetVerbosityLevel_Invalid(t *testing.T) {
	os.Setenv(VERBOSITY_ENV, "FOO")
	val := GetVerbosityLevel()
	if val != 0 {
		t.Errorf("Expected default 0 for invalid verbosity, got %d", val)
	}
}
