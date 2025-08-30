package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"mycelia/globals"
)

//------------------------------------------------------------------------------
// Byte buffer handling for decoding string, byte arrays, and unisnged integer
// values.

// Notably these do not create Mycelia error types because they were very low
// level and the callers give more context creating them instead.
//------------------------------------------------------------------------------

func readU8(r io.Reader, out *uint8) error {
	// endian is irrelevent for 1 byte but Read() requires it.
	return binary.Read(r, binary.BigEndian, out)
}

func readU32(r io.Reader, out *uint32) error {
	return binary.Read(r, binary.BigEndian, out)
}

// Read string up to 65535 characters long.
func readStringU8(r io.Reader) (string, error) {
	n, err := readU8Len(r)
	if err != nil {
		return "", err
	}
	if n == 0 {
		return "", nil
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", fmt.Errorf("read string bytes: %w", err)
	}
	return string(buf), nil
}

// Read string up to 65535 characters long.
func readStringU16(r io.Reader) (string, error) {
	n, err := readU16Len(r)
	if err != nil {
		return "", err
	}
	if n == 0 {
		return "", nil
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", fmt.Errorf("read string bytes: %w", err)
	}
	return string(buf), nil
}

// Read string up to 4b294m characters long.
func readStringU32(r io.Reader) (string, error) {
	n, err := readU32Len(r)
	if err != nil {
		return "", err
	}
	if n == 0 {
		return "", nil
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", fmt.Errorf("read string bytes: %w", err)
	}
	return string(buf), nil
}

// Read bytes up to 65535 bytes long.
func readBytesU16(r io.Reader) ([]byte, error) {
	n, err := readU16Len(r)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("read payload bytes: %w", err)
	}
	return buf, nil
}

// Read bytes up to 4b294m bytes long.
func readBytesU32(r io.Reader) ([]byte, error) {
	n, err := readU32Len(r)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("read payload bytes: %w", err)
	}
	return buf, nil
}

// Read from the io.Reader up to 255 bytes forwards.
func readU8Len(r io.Reader) (uint8, error) {
	var n uint8
	var u8Limit uint8 = 255

	if err := binary.Read(r, binary.BigEndian, &n); err != nil {
		return 0, fmt.Errorf("read length: %w", err)
	}
	// Sanity check - May want to store this value somewhere.
	if n > u8Limit {
		return 0, errors.New("declared length exceeds 255byte safety limit")
	}
	return n, nil
}

// Read from the io.Reader up to 65535 bytes forwards.
func readU16Len(r io.Reader) (uint16, error) {
	var n uint16
	var u16Limit uint16 = 64*globals.BytesInKilobyte - 1 // 64KB - 1B

	if err := binary.Read(r, binary.BigEndian, &n); err != nil {
		return 0, fmt.Errorf("read length: %w", err)
	}
	// Sanity check - May want to store this value somewhere.
	if n > u16Limit {
		return 0, errors.New("declared length exceeds 64KB safety limit")
	}
	return n, nil
}

// Read from the io.Reader up to 4b294m bytes forwards.
func readU32Len(r io.Reader) (uint32, error) {
	var n uint32
	var readLimit uint32 = 64 * globals.BytesInMegabyte // 64MB

	if err := binary.Read(r, binary.BigEndian, &n); err != nil {
		return 0, fmt.Errorf("read length: %w", err)
	}
	// Sanity check - May want to store this value somewhere.
	if n > readLimit {
		return 0, errors.New("declared length exceeds 64MB safety limit")
	}
	return n, nil
}
