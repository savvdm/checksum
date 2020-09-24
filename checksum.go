package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

var data map[string][]byte
var visited map[string]bool

var errorCount int
var addedCount int

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

func readData(file string) {
	f, err := os.Open(file)
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

func checksumFile(file string) (checksum []byte, err error) {
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

	data = make(map[string][]byte)
	visited = make(map[string]bool)

	readData(os.Args[1])
	fmt.Printf("Read %d checksums\n", len(data))

	root := "."
	if len(os.Args) > 2 {
		root = os.Args[2]
	}

	count := 0
	readDir(root, "", func(file string) {
		visited[file] = true
		if _, ok := data[file]; !ok {
			visited[file] = true
			path := makePath(root, file)
			if checksum, err := checksumFile(path); err == nil {
				data[file] = checksum
				fmt.Printf("A %s\n", file)
				addedCount++
			} else {
				registerError(err)
			}
		}
		// TODO: verify existing checksums with --check
		count++
	})

	fmt.Printf("Visited: %d\n", len(visited))
	fmt.Printf("Added: %d\n", addedCount)
	fmt.Printf("Errors: %d\n", errorCount)
}
