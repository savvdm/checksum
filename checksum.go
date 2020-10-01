package main

import (
	"bytes"
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
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
func (data dataMap) reportMissing(visited visitedFilesMap, status statusMap) {
	for file := range data {
		if _, ok := visited[file]; !ok {
			// file not found - will not be saved
			status.register(DELETED, file)
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

func readDir(root string, prefix string, callback func(path string, mod time.Time)) {
	f, err := os.Open(root)
	if err != nil {
		registerError(err)
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

type excludePatterns []*regexp.Regexp

func (excludes *excludePatterns) String() string {
	return ""
}

func (excludes *excludePatterns) Set(value string) error {
	*excludes = append(*excludes, regexp.MustCompile(value))
	return nil
}

func (excludes *excludePatterns) match(file string) bool {
	for _, patt := range *excludes {
		if patt.MatchString(file) {
			return true
		}
	}
	return false
}

func main() {

	var excludes excludePatterns

	flag.Var(&excludes, "exclude", "File name pattern to exclude")
	checkAll := flag.Bool("check", false, "Check all files")
	flag.Parse()

	if flag.NArg() < 1 {
		help()
		os.Exit(1)
	}

	data := make(dataMap)
	visited := make(visitedFilesMap)

	dataFile := flag.Arg(0)
	inputMod := data.read(dataFile)
	fmt.Printf("Read: %d\n", len(data))

	root := "."
	if flag.NArg() > 1 {
		root = flag.Arg(1)
	}

	status := make(statusMap)
	files := make([]string, 0, len(data)*2) // file list for writting
	readDir(root, "", func(file string, fileMod time.Time) {
		if excludes.match(file) {
			status.register(SKIPPED, file)
			return
		}
		visited[file] = true
		files = append(files, file)
		if sum, ok := data[file]; !ok || *checkAll || fileMod.After(inputMod) {
			path := makePath(root, file)
			if checksum, err := caclChecksum(path); err != nil {
				registerError(err)
			} else {
				status.register(CHECKED, "") // don't print anything, just count
				if !ok {
					status.register(ADDED, file)
					data[file] = checksum
				} else {
					if !bytes.Equal(sum, checksum) {
						status.register(REPLACED, file)
						data[file] = checksum
					}
				}
			}
		}
		// TODO: verify existing checksums with --check
	})

	data.reportMissing(visited, status)

	changed := status.sum([]string{ADDED, REPLACED, DELETED}) > 0
	if changed { // don't write file unless anything changed
		sort.Strings(files)
		data.write(dataFile, files)
	}

	// print stats
	status.print([]string{ADDED, REPLACED, DELETED, CHECKED, SKIPPED})
	fmt.Printf("Errors: %d\n", errorCount)
	if changed {
		fmt.Printf("Written: %d\n", len(files))
	} else {
		fmt.Println("No changes")
	}
}
