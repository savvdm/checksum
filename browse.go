package main

import (
	"io"
	"os"
	"strings"
	"time"
)

func makePath(path, name string) string {
	switch {
	case len(path) == 0:
		return name
	case strings.HasSuffix(path, "/"):
		return path + name
	default:
		return path + "/" + name
	}
}

func readDir(root string, prefix string, callback func(path string, mod time.Time)) {
	f, err := os.Open(root)
	if err != nil {
		stats.reportError(err)
		return
	}
	defer f.Close()

	const buflen = 100

	files, err := f.Readdir(buflen)
	for err == nil {
		for _, file := range files {
			name := file.Name()
			path := makePath(prefix, name)
			if file.IsDir() {
				subdir := makePath(root, name)
				readDir(subdir, path, callback) // prefix current path
			} else {
				callback(path, file.ModTime())
			}
		}
		files, err = f.Readdir(buflen)
	}
	if err != io.EOF {
		stats.reportError(err)
	}
}
