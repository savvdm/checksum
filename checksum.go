package main

import (
	"fmt"
	"github.com/savvdm/checksum/data"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// update from async calculation result
func updateFrom(fs data.FileSum, res *checkResult) {
	//fmt.Printf("Got checksum for %s\n", res.file)
	if res.err == nil {
		switch fs.Update(res.file, res.sum) {
		case data.Added:
			stats.report(Added, res.file)
		case data.Replaced:
			stats.report(Replaced, res.file)
		}
		stats.register(Checked)
	} else {
		stats.reportError(res.err)
	}
}

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

	fs := make(data.FileSum)

	inputMod := fs.Read(dataFile)
	if !params.nostat {
		fmt.Fprintf(os.Stderr, "Read: %d\n", fs.Len())
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
		exists := fs.SetVisited(file)
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
					updateFrom(fs, res)
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
			updateFrom(fs, res)
		}
	}

	// remove files not found on disk
	if params.delete {
		fs.Filter(func(file string) {
			stats.report(Deleted, file)
		})
	}

	// output data
	changed := stats.sum([]statKey{Added, Replaced, Deleted}) > 0
	if !params.dry && changed { // don't write file unless anything changed
		outfile := dataFile
		if len(params.outfile) > 0 {
			outfile = params.outfile
		}
		fs.Write(outfile)
	}

	// report stats
	if !params.nostat {
		stats.print()
		if changed {
			if params.dry {
				fmt.Fprintf(os.Stderr, "Dry run, not written: %d\n", fs.Len())
			} else {
				fmt.Fprintf(os.Stderr, "Written: %d\n", fs.Len())
			}
		} else {
			fmt.Fprintln(os.Stderr, "No changes")
		}
	}
}
