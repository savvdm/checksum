package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

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

	readDir(root, "", func(file string, mod time.Time) {
		if excludes.match(file) {
			stats.report(Skipped, file)
			return
		}
		visited[file] = true
		force := *checkAll || mod.After(inputMod)
		if _, exists := data[file]; !exists || force {
			path := makePath(root, file)
			if sum, err := caclChecksum(path); err != nil {
				stats.reportError(err)
			} else {
				data.update(file, sum)
			}
		}
	})

	data.reportMissing(visited)

	changed := stats.sum([]statKey{Added, Replaced, Deleted}) > 0
	if changed { // don't write file unless anything changed
		data.write(dataFile, visited)
	}

	// print stats
	stats.print()
	if changed {
		fmt.Printf("Written: %d\n", len(visited))
	} else {
		fmt.Println("No changes")
	}
}
