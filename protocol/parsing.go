package protocol

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"mycelia/boot"
	"mycelia/commands"
	"mycelia/str"
)

var _ = boot.RuntimeCfg // REQUIRED for global config values.

var ParseCommandErr = errors.New("unable to parse command")

// parseProtoVer extracts only the protocol version and returns it along with
// a slice that starts at the next byte (i.e., the remainder of the message).
func parseProtoVer(data []byte) (uint8, []byte, error) {
	const u8len = 1
	if len(data) < u8len {
		return 0, nil, io.ErrUnexpectedEOF
	}
	ver := data[0]
	return ver, data[u8len:], nil
}

func ParseLine(line []byte) (commands.Command, error) {
	version, rest, err := parseProtoVer(line)
	if err != nil {
		wMsg := fmt.Sprintf("Read protocol version: %v", err)
		str.WarningPrint(wMsg)
		return nil, ParseCommandErr
	}

	// The broker always works off of the same types of command objects.
	// Command objects may evolve over time, adding new fields for new
	// functionality, but the broker should remain compatible with previous
	// client side API versions.

	// If a client is using API ver 1 to communicate with Broker ver 2, then the
	// client should be able to still communicate.
	// This first token of a message is the API version, and this switch runs
	// the corresponding parsing logic.

	// This is mainly because early on there was uncertainty if the protocol and
	// command structure was done right, and we reserved the ability to update
	// it as we go.
	switch version {
	case 1:
		return decodeV1(rest)
	default:
		return nil, ParseCommandErr
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
