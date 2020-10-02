package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"
)

type dataMap map[string][]byte

const separator = "  " // Two-space separator used by sha1sum on Linux

func check(e error, code int) {
	if e != nil {
		fmt.Println(e)
		os.Exit(code)
	}
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
func (data dataMap) read(fname string) (mod time.Time) {
	info, err := os.Stat(fname)
	if os.IsNotExist(err) {
		fmt.Printf("File not found. To create new file, use 'touch %s' command\n", fname)
		os.Exit(3)
	}

	mod = info.ModTime()

	f, err := os.Open(fname)
	check(err, 3)
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		file, checksum := parseLine(scanner.Text())
		data[file] = checksum // TODO: check for duplicate keys
	}
	err = scanner.Err()
	check(err, 4)

	return
}

// write checksum data for the given (sorted) file list to the specified file
func (data dataMap) write(fname string, files []string) {
	f, err := os.Create(fname)
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

// report missing files
// return the number of files missing
func (data dataMap) reportMissing(visited visitedFilesMap) {
	for file := range data {
		if _, ok := visited[file]; !ok {
			// file not found - will not be saved
			stats.report(Deleted, file)
		}
	}
	return
}
