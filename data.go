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

// Checksum data in-memory storage
// key - relative path/to/file
// value[0] - visited flag (file found on disk)
// value[1:] - checksum for the file
type dataMap map[string][]byte

// checksum/path separator in the data file
const separator = "  " // Two-space separator used by sha1sum on Linux

// exit with the specified code in case of error
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

// add checksum for the file
// visited flag initialy cleared (zero)
// existing checksum & flag are silently overriden
func (data dataMap) setValue(file string, checksum []byte, visited bool) {
	value := make([]byte, 1, len(checksum)+1) // visited flag + checksum
	if visited {
		value[0] = 1
	}
	value = append(value, checksum...)
	data[file] = value // note: duplicate files are ignored
}

// set visited flag on existing file
// return false if no such file exists in the data map
func (data dataMap) setVisited(file string) (ok bool) {
	value, ok := data[file]
	if ok {
		value[0] = 1
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
		data.setValue(file, checksum, false) // visited flag initially not set
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
		value, ok := data[file]
		if !ok {
			panic("No checksum for " + file)
		}
		strsum := hex.EncodeToString(value[1:])
		_, err = fmt.Fprintf(f, "%s%s%s\n", strsum, separator, file)
		check(err, 10)
	}

	err = w.Flush()
	check(err, 10)
}

// update checksum for the specified file (and set visited flag)
func (data dataMap) update(file string, checksum []byte) (updated bool) {
	value, ok := data[file]
	if ok {
		if !bytes.Equal(value[1:], checksum) {
			stats.report(Replaced, file)
			copy(value[1:], checksum)
			updated = true
		}
		value[0] = 1 // set visited flag
	} else {
		stats.report(Added, file)
		data.setValue(file, checksum, true) // visited = true
		updated = true
	}
	return
}

// update from async calculation result
func (data dataMap) updateFrom(res *checkResult) {
	//fmt.Printf("Got checksum for %s\n", res.file)
	if res.err == nil {
		data.update(res.file, res.sum)
		stats.register(Checked)
	} else {
		stats.reportError(res.err)
	}
}

// remove files not found on disk
func (data dataMap) filter() {
	for file, value := range data {
		if value[0] == 0 { // file's not visited
			delete(data, file)
			stats.report(Deleted, file)
		}
	}
}
