package parsing

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"

	"mycelia/boot"
	"mycelia/commands"
	"mycelia/str"
)

var _ = boot.RuntimeCfg // REQUIRED for global config values.

const unknownCommand = "err"

func ParseLine(line []byte) (string, commands.Command) {
	parts, err := parseTokens(line)
	if err != nil {
		str.WarningPrint("Could not parse data - bad body.")
		return unknownCommand, nil
	}

	if len(parts) < 2 {
		str.WarningPrint("Could not parse data - missing version or command!")
		return unknownCommand, nil
	}

	verStr := parts[0]
	version, err := strconv.Atoi(verStr)
	if err != nil {
		return unknownCommand, nil
	}

	args := parts[1:] // prune off protocol version token.

	// The broker always works off of the same types of command objects.
	// Command objects may evolve over time, adding new fields for new
	// functionality, but the broker should remain compatible with previous
	// client side API versions.

	// If a client is using API ver 1 to communicate with Broker ver 2, then the
	// client should be able to still communicate.
	// This first token of a message is the API version, and this switch runs
	// the corresponding parsing logic.

	// This is mainly because early on there was uncertainty if the protocol and
	// command structure was done right.
	switch version {
	case 1:
		return parseDataV1(args)
	default:
		return unknownCommand, nil
	}
}

// parseTokens reads [varint length][body]... until EOF.
// Length is encoded as unsigned varint (LEB128), like protobuf.
func parseTokens(data []byte) ([]string, error) {
	r := bufio.NewReader(bytes.NewReader(data))
	var out []string

	for {
		// ReadUvarint returns (0, io.EOF) when no more data.
		length, err := binary.ReadUvarint(r)
		if err != nil {
			if errors.Is(err, bytes.ErrTooLarge) {
				return nil, err
			}

			// io.EOF is fine only if we're exactly at the end between fields.
			if err.Error() == "EOF" && r.Buffered() == 0 {
				return out, nil
			}

			return nil, err
		}

		if length == 0 {
			continue
		}

		buf := make([]byte, length)
		n, err := r.Read(buf)
		if err != nil {
			return nil, err
		}
		if uint64(n) != length {
			return nil, errors.New("truncated field body")
		}
		out = append(out, string(buf))
	}
}

// verifyTokenLength reports if the tokens array len is of the given length.
// Will print warning explaining that the command type failed because of
// incorrect argument count.
func verifyTokenLength(tokens []string, length int, cmdName string) bool {
	if len(tokens) != length {
		msg := "%s command failed, expected %d args, got %d isntead"
		wMsg := fmt.Sprintf(msg, cmdName, length, len(tokens))
		str.WarningPrint(wMsg)
		return false
	}
	return true
}
