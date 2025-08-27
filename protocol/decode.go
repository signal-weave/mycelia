package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

func readU32(r io.Reader, out *uint32) error {
	return binary.Read(r, binary.BigEndian, out)
}

func readU8(r io.Reader, out *uint8) error {
	// endian is irrelevent for 1 byte but Read() requires it.
	return binary.Read(r, binary.BigEndian, out)
}

func readString(r io.Reader) (string, error) {
	n, err := readLen(r)
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

func readBytes(r io.Reader) ([]byte, error) {
	n, err := readLen(r)
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

func readLen(r io.Reader) (uint32, error) {
	var n uint32
	if err := binary.Read(r, binary.BigEndian, &n); err != nil {
		return 0, fmt.Errorf("read length: %w", err)
	}
	// Sanity check - May want to store this value somewhere.
	if n > 64*1024*1024 {
		return 0, errors.New("declared length exceeds 64MB safety limit")
	}
	return n, nil
}
