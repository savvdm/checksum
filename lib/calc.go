package lib

import (
	"crypto/sha1"
	"io"
	"os"
)

type CheckRequest struct {
	Root, File string
}

type CheckResult struct {
	File string
	Sum  []byte
	Err  error
}

func Calc(path string) (checksum []byte, err error) {
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

func CalcAll(in chan *CheckRequest, out chan *CheckResult) {
	defer func() { out <- nil }()
	for req := range in {
		var result CheckResult
		result.File = req.File
		path := MakePath(req.Root, req.File)
		result.Sum, result.Err = Calc(path)
		out <- &result
	}
}

func StartWorkers(num int) (in chan *CheckRequest, out chan *CheckResult) {
	in = make(chan *CheckRequest, num*5)
	out = make(chan *CheckResult, num*5)
	for i := 0; i < num; i++ {
		go CalcAll(in, out)
	}
	return
}
