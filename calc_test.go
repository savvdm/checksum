package main

import (
	"bytes"
	"testing"
)

func TestCalc(t *testing.T) {
	sum, err := calc("test/data.txt")

	if err != nil {
		t.Errorf("Can't calc checksum for test file: %v\n", err)
	}

	expected := []byte{
		0xa5, 0xc3, 0x41, 0xbe, 0xc5, 0xc8, 0x9e,
		0xd1, 0x67, 0x58, 0x43, 0x50, 0x69, 0xe3,
		0x12, 0x4b, 0x36, 0x85, 0xad, 0x93}

	if !bytes.Equal(expected, sum) {
		t.Errorf("Checksum doesn't match:\nFound:\t%v\nWanted:\t%v\n", sum, expected)
	}
}
