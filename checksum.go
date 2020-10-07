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
	fmt.Printf("Usage: %s [params] checksum_file dir_to_check\n", os.Args[0])
	flag.PrintDefaults()
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

	var params cmdParams
	params.init()

	if flag.NArg() < 1 {
		help()
		os.Exit(1)
	}

	data := make(dataMap)
	visited := make(visitedFiles)

	dataFile := flag.Arg(0)
	inputMod := data.read(dataFile)
	if !params.nostat {
		fmt.Printf("Read: %d\n", len(data))
	}

	root := "."
	if flag.NArg() > 1 {
		root = flag.Arg(1)
	}

	readDir(root, "", func(file string, mod time.Time) {
		if params.excludes.match(file) {
			stats.reportIf(!params.quiet, Skipped, file)
			return
		}
		visited[file] = true
		force := params.mode == All || params.mode == Modified && mod.After(inputMod)
		if _, exists := data[file]; !exists || force {
			path := makePath(root, file)
			if sum, err := caclChecksum(path); err != nil {
				stats.reportError(err)
			} else {
				if !data.update(file, sum) {
					// checksum not changed
					stats.reportIf(params.reportOk(), Ok, file)
				}
			}
		}
	})

	data.reportMissing(visited)

	changed := stats.sum([]statKey{Added, Replaced, Deleted}) > 0
	if !params.dry && changed { // don't write file unless anything changed
		data.writeSorted(dataFile, visited)
	}

	if !params.nostat {
		stats.print()
		if changed {
			fmt.Printf("Written: %d\n", len(visited))
		} else {
			fmt.Println("No changes")
		}
	}
}
