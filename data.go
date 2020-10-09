package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

type dataMap map[string][]byte
type visitedFiles map[string]bool

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

// sort the map's keys and return in a slice
func (data dataMap) sortFiles() []string {
	files := make([]string, 0, len(data))
	for file := range data {
		files = append(files, file)
	}
	sort.Strings(files)
	return files
}

// write checksum data to the specified file
// only files in the visited map are written
func (data dataMap) writeSorted(fname string) {
	files := data.sortFiles()
	data.write(fname, files)
}

// write out data for the given files,
// in the specified order
func (data dataMap) write(fname string, files []string) {
	f, err := os.Create(fname)
	check(err, 10)
	defer f.Close()

	w := bufio.NewWriter(f)

	// write data
	for _, file := range files {
		sum, ok := data[file]
		if !ok {
			panic("No checksum for " + file)
		}
		if sum == nil {
			panic("Empty checksum for " + file)
		}
		strsum := hex.EncodeToString(sum)
		_, err = fmt.Fprintf(f, "%s%s%s\n", strsum, separator, file)
		check(err, 10)
	}

	err = w.Flush()
	check(err, 10)
}

// check sha1 sum for the specified file
// under the specified root
func (data dataMap) update(file string, checksum []byte) (updated bool) {
	sum, exists := data[file]
	if exists {
		if !bytes.Equal(sum, checksum) {
			stats.report(Replaced, file)
			data[file] = checksum
			updated = true
		}
	} else {
		stats.report(Added, file)
		data[file] = checksum
		updated = true
	}
	return updated
}

// remove files not found in the specified map
func (data dataMap) filter(visited visitedFiles) {
	for file := range data {
		if _, ok := visited[file]; !ok {
			delete(data, file)
			stats.report(Deleted, file)
		}
	}
}
