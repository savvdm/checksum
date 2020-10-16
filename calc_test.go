package main

import (
	"bytes"
	"testing"
)

func verifyChecksum(t *testing.T, sum []byte) {
	expected := []byte{
		0xa5, 0xc3, 0x41, 0xbe, 0xc5, 0xc8, 0x9e,
		0xd1, 0x67, 0x58, 0x43, 0x50, 0x69, 0xe3,
		0x12, 0x4b, 0x36, 0x85, 0xad, 0x93}

	if !bytes.Equal(expected, sum) {
		t.Errorf("Checksum doesn't match:\nFound:\t%v\nWanted:\t%v\n", sum, expected)
	}
}

func checkError(t *testing.T, err error, file string) {
	if err != nil {
		t.Errorf("Can't calc checksum for test file: %v\n", err)
	}
}

func TestCalc(t *testing.T) {
	const file = "test/data.txt"
	sum, err := calc(file)
	checkError(t, err, file)
	verifyChecksum(t, sum)
}

func TestAsyncCalc(t *testing.T) {
	const file = "data.txt"
	in, out := startWorkers(1)
	in <- &checkRequest{"test", file}
	close(in)
	var gotResult bool
	for res := range out {
		if res == nil {
			break
		}
		gotResult = true
		if res.file != file {
			t.Errorf("Wrong file name in calc result: %s (expected %s)\n", res.file, file)
		}
		checkError(t, res.err, file)
		verifyChecksum(t, res.sum)
	}
	if !gotResult {
		t.Error("Didn't get calculation result\n")
	}
}
