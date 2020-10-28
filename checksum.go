package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

func main() {

	var params cmdParams
	params.init()

	dataFile, root := params.parse()

	// setup cpu profiling
	if params.cpuprofile != "" {
		f, err := os.Create(params.cpuprofile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	data := make(dataMap)

	inputMod := data.read(dataFile)
	if !params.nostat {
		fmt.Fprintf(os.Stderr, "Read: %d\n", len(data))
	}

	var numWorkers = runtime.GOMAXPROCS(0)
	in, out := startWorkers(numWorkers)

	readDir(root, "", func(file string, mod time.Time) {
		// check includes (if any)
		if len(params.includes) > 0 && !params.includes.match(file) {
			return
		}
		// check excludes
		if len(params.excludes) > 0 && params.excludes.match(file) {
			stats.reportIf(params.verbose, Skipped, file)
			return
		}
		// mark the file visited (and see if it exists)
		stats.register(Visited)
		exists := data.setVisited(file)
		force := params.mode == All || params.mode == Modified && mod.After(inputMod)
		if !exists || force {
			// enqueue checksum calculation
			// don't block if channel is full
			req := checkRequest{root, file}
			for {
				select {
				case in <- &req:
					return
				case res := <-out:
					data.updateFrom(res)
				}
			}
		}
	})

	// no more checksum calculations will be queued
	close(in)

	// read calculated checksums & update data
	for res := range out {
		if res == nil {
			if numWorkers--; numWorkers == 0 {
				break
			}
		} else {
			data.updateFrom(res)
		}
	}

	// remove files not found on disk
	if params.delete {
		data.filter()
	}

	// output data
	changed := stats.sum([]statKey{Added, Replaced, Deleted}) > 0
	if !params.dry && changed { // don't write file unless anything changed
		outfile := dataFile
		if len(params.outfile) > 0 {
			outfile = params.outfile
		}
		data.writeSorted(outfile)
	}

	// report stats
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
