package london

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
)

// ascii codes used in this package
const (
	asciiSTX = 0x02
	asciiETX = 0x03
	asciiACK = 0x06
	asciiNAK = 0x15
	asciiESC = 0x1b
)

// the lowest subscribe time
const minSubscribeInterval = 50

// methods
const (
	methodSet                = 0x88
	methodSetPercent         = 0x8d
	methodSubscribe          = 0x89
	methodUnsubscribe        = 0x8a
	methodSubscribePercent   = 0x8e
	methodUnsubscribePercent = 0x8f
)

// random constants
const (
	virtualDevice = 0x03
	encodeOffset  = 0x1b80
)

// bytes that need to be encoded
var encodedBytes = []int{
	asciiESC,
	asciiSTX,
	asciiETX,
	asciiACK,
	asciiNAK,
}

type state []byte

// common states
var (
	stateGain     = buildState(0x00, 0x00)
	stateMute     = buildState(0x00, 0x01)
	statePolarity = buildState(0x00, 0x02)
)

func buildState(a, b byte) state {
	return []byte{a, b}
}

func buildCommand(base uint8, state state, block string, data []byte) ([]byte, error) {
	if len(data) != 4 {
		return []byte{}, fmt.Errorf("data must have 4 elements")
	}

	bBlock, err := hex.DecodeString(block)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to decode block: %w", err)
	}

	cmd := []byte{base, 0x00, 0x00, virtualDevice}

	cmd = append(cmd, bBlock...)
	cmd = append(cmd, state...)
	cmd = append(cmd, data...)
	cmd = append(cmd, checksum(cmd))

	cmd, err = encode(cmd)
	if err != nil {
		return []byte{}, fmt.Errorf("unable to encode command: %w", err)
	}

	return cmd, nil
}

func buildSubscribeCommand(base uint8, state state, block string, interval uint32) ([]byte, error) {
	bInterval := make([]byte, 4)
	binary.BigEndian.PutUint32(bInterval, interval)

	return buildCommand(base, state, block, bInterval)
}

func buildUnsubscribeCommand(base uint8, state state, block string) ([]byte, error) {
	return buildSubscribeCommand(base, state, block, 0)
}

func checksum(bytes []byte) byte {
	var sum byte
	for i := range bytes {
		sum = sum ^ bytes[i]
	}

	return sum
}

func encode(b []byte) ([]byte, error) {
	escaped := make([]byte, 2)

	for _, eByte := range encodedBytes {
		binary.BigEndian.PutUint16(escaped, uint16(eByte+encodeOffset))
		b = bytes.Replace(b, []byte{byte(eByte)}, escaped, -1)
	}

	switch {
	case bytes.Contains(b, []byte{asciiSTX}):
		return []byte{}, errors.New("must not contain STX")
	case bytes.Contains(b, []byte{asciiETX}):
		return []byte{}, errors.New("must not contain ETX")
	}

	enc := []byte{asciiSTX}
	enc = append(enc, b...)
	enc = append(enc, asciiETX)

	return enc, nil
}

func decode(b []byte) ([]byte, error) {
	escaped := make([]byte, 2)

	prefix := []byte{asciiSTX}
	suffix := []byte{asciiETX}

	switch {
	case !bytes.HasPrefix(b, prefix):
		return []byte{}, errors.New("must begin with STX")
	case !bytes.HasSuffix(b, suffix):
		return []byte{}, errors.New("must begin with ETX")
	}

	b = bytes.TrimPrefix(b, prefix)
	b = bytes.TrimSuffix(b, suffix)

	switch {
	case bytes.Contains(b, prefix):
		return []byte{}, errors.New("erroneous STX")
	case bytes.Contains(b, suffix):
		return []byte{}, errors.New("erroneous ETX")
	}

	for _, eByte := range encodedBytes {
		binary.BigEndian.PutUint16(escaped, uint16(eByte+encodeOffset))
		b = bytes.Replace(b, escaped, []byte{byte(eByte)}, -1)
	}

	// check the checksum
	sum := checksum(b[:len(b)-1])
	if sum != b[len(b)-1] {
		return []byte{}, errors.New("invalid checksum")
	}

	b = bytes.TrimSuffix(b, []byte{sum})

	return b, nil
}
