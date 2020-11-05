package lib

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
type FileSum map[string][]byte

// checksum/path separator in the data file
const separator = "  " // Two-space separator used by sha1sum on Linux

func (data FileSum) Len() int {
	return len(data)
}

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
func (data FileSum) setValue(file string, checksum []byte, visited bool) {
	value := make([]byte, 1, len(checksum)+1) // visited flag + checksum
	if visited {
		value[0] = 1
	}
	value = append(value, checksum...)
	data[file] = value // note: duplicate files are ignored
}

// set visited flag on existing file
// return false if no such file exists in the data map
func (data FileSum) MarkVisited(file string) bool {
	value, ok := data[file]
	if ok {
		value[0] = 1
	}
	return ok
}

// read data map from the given file
func (data FileSum) Read(fname string) (mod time.Time) {
	info, err := os.Stat(fname)
	if os.IsNotExist(err) {
		mod = time.Now()
		return // new file, nothing to load
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
func (data FileSum) sortFiles() []string {
	files := make([]string, 0, len(data))
	for file := range data {
		files = append(files, file)
	}
	sort.Strings(files)
	return files
}

// write checksum data to the specified file
// only files in the visited map are written
func (data FileSum) Write(fname string) {
	files := data.sortFiles()
	data.writeFiles(fname, files)
}

// write out data for the given files,
// in the specified order
func (data FileSum) writeFiles(fname string, files []string) {
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
		_, err = fmt.Fprintf(w, "%s%s%s\n", strsum, separator, file)
		check(err, 10)
	}

	err = w.Flush()
	check(err, 10)
}

// update checksum for the specified file (and set visited flag)
func (data FileSum) Update(file string, checksum []byte) (status StatKey) {
	value, ok := data[file]
	if ok {
		if bytes.Equal(value[1:], checksum) {
			status = Checked
		} else {
			copy(value[1:], checksum)
			status = Replaced
		}
	} else {
		data.setValue(file, checksum, true) // visited = true
		status = Added
	}
	return
}

// remove files not found on disk
func (data FileSum) Filter(deleted func(file string)) {
	for file, value := range data {
		if value[0] == 0 { // file's not visited
			delete(data, file)
			deleted(file)
		}
	}
}
