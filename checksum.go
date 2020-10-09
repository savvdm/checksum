package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"time"
)

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

	dataFile, root := params.parse()

	data := make(dataMap)
	visited := make(visitedFiles)

	inputMod := data.read(dataFile)
	if !params.nostat {
		fmt.Printf("Read: %d\n", len(data))
	}

	readDir(root, "", func(file string, mod time.Time) {
		if params.excludes.match(file) {
			stats.reportIf(params.verbose, Skipped, file)
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
		outfile := dataFile
		if len(params.outfile) > 0 {
			outfile = params.outfile
		}
		data.writeSorted(outfile, visited)
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
