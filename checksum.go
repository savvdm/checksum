package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type dataMap map[string][]byte
type visitedFilesMap map[string]bool

var errorCount int
var addedCount int
var deletedCount int

const separator = "  " // Two-space separator used by sha1sum on Linux

func help() {
	fmt.Println("Specify checsum file name")
}

func check(e error, code int) {
	if e != nil {
		fmt.Println(e)
		os.Exit(code)
	}
}

func registerError(e error) {
	fmt.Println(e)
	errorCount++
}

// parse single line of the data file
// input line sample (notice: separated by two space chars):
// 2a8c416df19174d4fb421d8c9b9cddfd54914c45  Backup/Geo.tgz
// file path may include spaces
func parseLine(line string) (file string, checksum []byte) {
	fields := strings.SplitN(line, separator, 2)
	if len(fields) != 2 {
		fmt.Printf("Invalid input line: '%s'\n", line)
		os.Exit(5)
	}

	sum, file := fields[0], fields[1]
	if len(sum) == 0 {
		fmt.Printf("Invalid input - checksum is missing: '%s'\n", line)
		os.Exit(5)
	}
	if len(file) == 0 {
		fmt.Printf("Invalid input - file path is missing: '%s'\n", line)
		os.Exit(5)
	}

	checksum, err := hex.DecodeString(sum)
	if err != nil {
		fmt.Printf("%s: %s %s\n", err, sum, file)
		os.Exit(5)
	}

	return
}

// read data map from the given file
//
func (data dataMap) read(csfile string) {
	f, err := os.Open(csfile)
	check(err, 3)
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		file, checksum := parseLine(scanner.Text())
		data[file] = checksum // TODO: check for duplicate keys
	}
	err = scanner.Err()
	check(err, 4)
}

// write data for the given (sorted) files to the specified file
func (data dataMap) write(csfile string, files []string) {
	f, err := os.Create(csfile)
	check(err, 10)
	defer f.Close()

	w := bufio.NewWriter(f)

	// write data
	for _, file := range files {
		strsum := hex.EncodeToString(data[file])
		_, err = fmt.Fprintf(f, "%s%s%s\n", strsum, separator, file)
		check(err, 10)
	}

	err = w.Flush()
	check(err, 10)
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

// build list of files, dropping removed (not visited) files
// dropped files are reported to the user
func (data dataMap) buildFileList(visited visitedFilesMap) []string {
	files := make([]string, 0, len(data))
	for file := range data {
		if _, ok := visited[file]; ok {
			files = append(files, file)
		} else {
			// file not found, dropping from list
			fmt.Println("D", file)
			deletedCount++
		}
	}
	return files
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

	readDir(root, "", func(file string) {
		visited[file] = true
		if _, ok := data[file]; !ok {
			path := makePath(root, file)
			if checksum, err := caclChecksum(path); err == nil {
				data[file] = checksum
				fmt.Println("A", file)
				addedCount++
			} else {
				registerError(err)
			}
		}
		// TODO: verify existing checksums with --check
	})

	files := data.buildFileList(visited)
	changed := addedCount > 0 || deletedCount > 0 // don't write file unless changed
	if changed {
		sort.Strings(files)
		data.write(dataFile, files)
	}

	fmt.Printf("Visited: %d\n", len(visited))
	fmt.Printf("Added: %d\n", addedCount)
	fmt.Printf("Deleted: %d\n", deletedCount)
	if changed {
		fmt.Printf("Written: %d\n", len(files))
	}
	fmt.Printf("Errors: %d\n", errorCount)
}
