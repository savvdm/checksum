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

type fileData []byte

func (fdat fileData) status() Status {
	return Status(fdat[0])
}

func (fdat fileData) setStatus(status Status) {
	fdat[0] = byte(status)
}

func (fdat fileData) checksum() []byte {
	return fdat[1:]
}

func (fdat fileData) checksumString() string {
	return hex.EncodeToString(fdat.checksum())
}

func (fdat fileData) checksumEqual(checksum []byte) bool {
	return bytes.Equal(fdat.checksum(), checksum)
}

func makeFileData(checksum []byte, status Status) (fdat fileData) {
	fdat = make([]byte, len(checksum)+1)
	fdat.setStatus(status)
	copy(fdat[1:], checksum)
	return
}

// Checksum data in-memory storage
// key - relative path/to/file
// value - fileData structure (status+checksum)
type Data map[string]fileData

// checksum/path separator in the data file
const separator = "  " // Two-space separator used by sha1sum on Linux

func (data Data) Len() int {
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

// set visited flag on existing file
// return false if no such file exists in the data map
func (data Data) MarkVisited(file string) bool {
	if fdat, ok := data[file]; ok {
		fdat.setStatus(Visited)
		return true
	}
	return false
}

// read data map from the given file
func (data Data) Read(fname string) (mod time.Time) {
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
		data[file] = makeFileData(checksum, Read)
	}
	err = scanner.Err()
	check(err, 4)

	return
}

// sort the map's keys and return in a slice
func (data Data) SortKeys() []string {
	files := make([]string, 0, len(data))
	for file := range data {
		files = append(files, file)
	}
	sort.Strings(files)
	return files
}

// sort the map's keys and return in a slice
func (data Data) ReportFiles(files []string, verbose bool) {
	for _, file := range files {
		fdat := data[file]
		switch status := fdat.status(); status {
		case Added, Replaced, Deleted:
			ReportFile(file, status)
		case Checked:
			if verbose {
				ReportFile(file, status)
			}
		}
	}
}

// write out data for the given files,
// in the specified order
func (data Data) WriteFiles(files []string, fname string) {
	f, err := os.Create(fname)
	check(err, 10)
	defer f.Close()

	w := bufio.NewWriter(f)

	// write data
	for _, file := range files {
		fdat, ok := data[file]
		if !ok {
			panic("No checksum for " + file)
		}
		if fdat.status() != Deleted {
			_, err = fmt.Fprintf(w, "%s%s%s\n", fdat.checksumString(), separator, file)
			check(err, 10)
		}
	}

	err = w.Flush()
	check(err, 10)
}

// update checksum for the specified file
func (data Data) Update(file string, checksum []byte) (status Status) {
	fdat, ok := data[file]
	if ok {
		if fdat.checksumEqual(checksum) {
			fdat.setStatus(Checked)
			status = Checked
		} else {
			data[file] = makeFileData(checksum, Replaced)
			status = Replaced
		}
	} else {
		data[file] = makeFileData(checksum, Added)
		status = Added
	}
	return
}

// remove files not found on disk
func (data Data) Filter() (count int) {
	for _, fdat := range data {
		if fdat.status() == Read { // file's not visited
			fdat.setStatus(Deleted)
			count++
		}
	}
	return
}
