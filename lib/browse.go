package lib

import (
	"io"
	"os"
	"strings"
	"time"
)

func MakePath(path, name string) string {
	switch {
	case len(path) == 0:
		return name
	case strings.HasSuffix(path, "/"):
		return path + name
	default:
		return path + "/" + name
	}
}

func ReadDir(root string, prefix string, callback func(path string, mod time.Time)) error {
	f, err := os.Open(root)
	if err != nil {
		return err
	}
	defer f.Close()

	const buflen = 100

	files, err := f.Readdir(buflen)
	for err == nil {
		for _, file := range files {
			name := file.Name()
			path := MakePath(prefix, name)
			if file.IsDir() {
				subdir := MakePath(root, name)
				ReadDir(subdir, path, callback) // prefix current path
			} else {
				callback(path, file.ModTime())
			}
		}
		files, err = f.Readdir(buflen)
	}
	if err != io.EOF {
		return err
	}
	return nil
}
