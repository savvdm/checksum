package main

import (
	"bytes"
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

type visitedFilesMap map[string]bool

func help() {
	fmt.Println("Specify checsum file name")
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

	files := make([]string, 0, len(data)*2) // file list for writting
	readDir(root, "", func(file string, fileMod time.Time) {
		if excludes.match(file) {
			stats.report(Skipped, file)
			return
		}
		visited[file] = true
		files = append(files, file)
		if sum, ok := data[file]; !ok || *checkAll || fileMod.After(inputMod) {
			path := makePath(root, file)
			if checksum, err := caclChecksum(path); err != nil {
				stats.reportError(err)
			} else {
				stats.register(Checked)
				if !ok {
					stats.report(Added, file)
					data[file] = checksum
				} else {
					if !bytes.Equal(sum, checksum) {
						stats.report(Replaced, file)
						data[file] = checksum
					}
				}
			}
		}
		// TODO: verify existing checksums with --check
	})

	data.reportMissing(visited)

	changed := stats.sum([]statKey{Added, Replaced, Deleted}) > 0
	if changed { // don't write file unless anything changed
		sort.Strings(files)
		data.write(dataFile, files)
	}

	// print stats
	stats.print()
	if changed {
		fmt.Printf("Written: %d\n", len(files))
	} else {
		fmt.Println("No changes")
	}
}
