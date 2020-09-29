package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type visitedFilesMap map[string]bool

var errorCount int

func help() {
	fmt.Println("Specify checsum file name")
}

func registerError(e error) {
	fmt.Println(e)
	errorCount++
}

// report missing files
// return the number of files missing
func (data dataMap) reportMissing(visited visitedFilesMap) (deleted int) {
	for file := range data {
		if _, ok := visited[file]; !ok {
			// file not found, dropping from list
			fmt.Println("D", file)
			deleted++
		}
	}
	return
}

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

func readDir(root string, prefix string, callback func(path string)) {
	f, err := os.Open(root)
	if err != nil {
		registerError(err)
		return
	}
	defer f.Close()

	const buflen = 100

	info, err := f.Readdir(buflen)
	for err == nil {
		for _, file := range info {
			name := file.Name()
			path := makePath(prefix, name)
			if file.IsDir() {
				subdir := makePath(root, name)
				readDir(subdir, path, callback) // prefix current path
			} else {
				callback(path)
			}
		}
		info, err = f.Readdir(buflen)
	}
	if err != io.EOF {
		registerError(err)
	}
}

func caclChecksum(file string) (checksum []byte, err error) {
	f, err := os.Open(file)
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

func main() {

	if len(os.Args) < 2 {
		help()
		os.Exit(1)
	}

	data := make(dataMap)
	visited := make(visitedFilesMap)

	dataFile := os.Args[1]
	data.read(dataFile)
	fmt.Printf("Read: %d\n", len(data))

	root := "."
	if len(os.Args) > 2 {
		root = os.Args[2]
	}

	added := 0
	files := make([]string, 0, len(data)*2) // file list for writting
	readDir(root, "", func(file string) {
		visited[file] = true
		files = append(files, file)
		if _, ok := data[file]; !ok {
			path := makePath(root, file)
			if checksum, err := caclChecksum(path); err == nil {
				data[file] = checksum
				fmt.Println("A", file)
				added++
			} else {
				registerError(err)
			}
		}
		// TODO: verify existing checksums with --check
	})

	deleted := data.reportMissing(visited)

	changed := added > 0 || deleted > 0 // don't write file unless changed
	if changed {
		sort.Strings(files)
		data.write(dataFile, files)
	}

	// print stats
	fmt.Printf("Added: %d\n", added)
	fmt.Printf("Deleted: %d\n", deleted)
	fmt.Printf("Errors: %d\n", errorCount)
	if changed {
		fmt.Printf("Written: %d\n", len(files))
	} else {
		fmt.Println("No changes")
	}
}
