package main

import (
	"crypto/sha1"
	"io"
	"os"
)

type checkRequest struct {
	root, file string
}

type checkResult struct {
	file string
	sum  []byte
	err  error
}

func calc(path string) (checksum []byte, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	h := sha1.New()
	if _, err = io.Copy(h, f); err != nil {
		return
	}

	checksum = h.Sum(nil)
	return
}

func calcChecksums(in chan *checkRequest, out chan *checkResult) {
	defer func() { out <- nil }()
	for req := range in {
		var result checkResult
		result.file = req.file
		path := makePath(req.root, req.file)
		result.sum, result.err = calc(path)
		out <- &result
	}
}
