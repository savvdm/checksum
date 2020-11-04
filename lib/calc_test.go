package lib

import (
	"bytes"
	"testing"
)

func verifyChecksum(t *testing.T, sum []byte) {
	expected := []byte{
		0x3, 0xcf, 0xd7, 0x43, 0x66, 0x1f, 0x7, 0x97, 0x5f, 0xa2,
		0xf1, 0x22, 0xc, 0x51, 0x94, 0xcb, 0xaf, 0xf4, 0x84, 0x51,
	}

	if !bytes.Equal(expected, sum) {
		t.Errorf("Checksum doesn't match:\nFound:\t%#v\nWanted:\t%#v\n", sum, expected)
	}
}

func checkError(t *testing.T, err error, file string) {
	if err != nil {
		t.Errorf("Can't calc checksum for test file: %v\n", err)
	}
}

func TestCalc(t *testing.T) {
	const file = "test/data.txt"
	sum, err := Calc(file)
	checkError(t, err, file)
	verifyChecksum(t, sum)
}

func TestError(t *testing.T) {
	const file = "test/missing.txt"
	_, err := Calc(file)
	if err == nil {
		t.Error("Did not fail on unexistent file\n")
	}
}

func TestAsyncCalc(t *testing.T) {
	const file = "data.txt"
	in, out := StartWorkers(1)
	in <- &CheckRequest{"test", file}
	close(in)
	var gotResult bool
	for res := range out {
		if res == nil {
			break
		}
		gotResult = true
		if res.File != file {
			t.Errorf("Wrong file name in calc result: %s (expected %s)\n", res.File, file)
		}
		checkError(t, res.Err, file)
		verifyChecksum(t, res.Sum)
	}
	if !gotResult {
		t.Error("Didn't get calculation result\n")
	}
}
