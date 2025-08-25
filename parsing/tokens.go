package parsing

import (
	"fmt"
	"mycelia/str"
	"net/url"
)

// -----------------------------------------------------------------------------
// Herein are utilities for parsing tokens from raw byte arrays.
// -----------------------------------------------------------------------------

// SplitTokensInPlace splits b by ";;" without copying field data.
// Returned slices alias b (zero-copy).
func splitTokens(b []byte) [][]byte {
	var out [][]byte
	end := len(b)
	last := 0
	for i := 0; i+1 < end; {
		if b[i] == ';' && b[i+1] == ';' {
			out = append(out, b[last:i])
			i += 2
			last = i
		} else {
			i++
		}
	}

	// trim trailing CR/LF
	for end > last && (b[end-1] == '\n' || b[end-1] == '\r') {
		end--
	}
	out = append(out, b[last:end])
	return out
}

// Unescape decodes a percent-escaped field (stdlib, fast enough).
func unescape(b []byte) (string, error) {
	return url.PathUnescape(string(b))
}

// Unescapes an entire array of byte tokens.
func unescapeTokens(args [][]byte) ([]string, error) {
	buf := []string{}

	for _, v := range args {
		s, err := unescape(v)
		if err != nil {
			return nil, err
		}
		buf = append(buf, s)
	}

	return buf, nil
}

func verifyTokenLength(tokens []string, length int, cmdName string) bool {
	if len(tokens) != length {
		msg := "%s command failed, expected %d args, got %d isntead"
		wMsg := fmt.Sprintf(msg, cmdName, length, len(tokens))
		str.WarningPrint(wMsg)
		return false
	}
	return true
}
