package main

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

func main() {

	var params cmdParams
	params.init()

	dataFile, root := params.parse()

	data := make(dataMap)

	inputMod := data.read(dataFile)
	if !params.nostat {
		fmt.Fprintf(os.Stderr, "Read: %d\n", len(data))
	}

	var numWorkers = runtime.GOMAXPROCS(0)
	in, out := startWorkers(numWorkers)

	readDir(root, "", func(file string, mod time.Time) {
		if params.excludes.match(file) {
			stats.reportIf(params.verbose, Skipped, file)
			return
		}
		exists := data.setVisited(file)
		force := params.mode == All || params.mode == Modified && mod.After(inputMod)
		if !exists || force {
			//fmt.Printf("Enqueue %s/%s\n", root, file)
			in <- &checkRequest{root, file} // enqueue checksum calculation
		}
		for {
			select {
			case res := <-out:
				data.updateFrom(res)
			default:
				return
			}
		}
	})

	close(in)

	for res := range out {
		if res == nil {
			if numWorkers--; numWorkers == 0 {
				break
			}
		} else {
			data.updateFrom(res)
		}
	}

	if !params.nodelete {
		data.filter()
	}

	changed := stats.sum([]statKey{Added, Replaced, Deleted}) > 0
	if !params.dry && changed { // don't write file unless anything changed
		outfile := dataFile
		if len(params.outfile) > 0 {
			outfile = params.outfile
		}
		data.writeSorted(outfile)
	}

	if !params.nostat {
		stats.print()
		if changed {
			if params.dry {
				fmt.Fprintf(os.Stderr, "Dry run, not written: %d\n", len(data))
			} else {
				fmt.Fprintf(os.Stderr, "Written: %d\n", len(data))
			}
		} else {
			fmt.Fprintln(os.Stderr, "No changes")
		}
	}
}
