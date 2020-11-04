package main

import (
	"fmt"
	"github.com/savvdm/checksum/lib"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

func startProfiling(file string) {
	f, err := os.Create(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	pprof.StartCPUProfile(f)
}

func main() {

	var params cmdParams
	params.init()

	dataFile, root := params.parse()

	// setup cpu profiling
	if params.cpuprofile != "" {
		startProfiling(params.cpuprofile)
		defer pprof.StopCPUProfile()
	}

	data := make(lib.FileSum)

	inputMod := data.Read(dataFile)
	if !params.nostat {
		fmt.Fprintf(os.Stderr, "Read: %d\n", data.Len())
	}

	var numWorkers = runtime.GOMAXPROCS(0)
	in, out := lib.StartWorkers(numWorkers)

	var stats lib.StatCounts

	// process checksum result
	update := func(res *lib.CheckResult) {
		if res.Err == nil {
			switch status := data.Update(res.File, res.Sum); status {
			case lib.Added, lib.Replaced:
				stats.Report(status, res.File)
			case lib.Checked:
				stats.ReportIf(params.verbose, status, res.File)
			}
		} else {
			stats.ReportError(res.Err)
		}
	}

	if err := lib.ReadDir(root, "", func(file string, mod time.Time) {
		// check includes (if any)
		if len(params.includes) > 0 && !params.includes.Match(file) {
			return
		}
		// check excludes
		if len(params.excludes) > 0 && params.excludes.Match(file) {
			stats.ReportIf(params.verbose, lib.Skipped, file)
			return
		}
		// mark the file visited (and see if it exists)
		stats.Register(lib.Visited)
		exists := data.MarkVisited(file)
		force := params.mode == All || params.mode == Modified && mod.After(inputMod)
		if !exists || force {
			// enqueue checksum calculation
			// don't block if channel is full
			req := lib.CheckRequest{root, file}
			for {
				select {
				case in <- &req:
					return
				case res := <-out:
					update(res)
				}
			}
		}
	}); err != nil {
		println(err)
		os.Exit(2)
	}

	// no more checksum calculations will be queued
	close(in)

	// read calculated checksums & update lib
	for res := range out {
		if res == nil {
			if numWorkers--; numWorkers == 0 {
				break
			}
		} else {
			update(res)
		}
	}

	// remove files not found on disk
	if params.delete {
		data.Filter(func(file string) {
			stats.Report(lib.Deleted, file)
		})
	}

	// output lib
	changed := stats.IsChanged()
	if !params.dry && changed { // don't write file unless anything changed
		outfile := dataFile
		if len(params.outfile) > 0 {
			outfile = params.outfile
		}
		data.Write(outfile)
	}

	// report stats
	if !params.nostat {
		stats.Print()
		if changed {
			if params.dry {
				fmt.Fprintf(os.Stderr, "Dry run, not written: %d\n", data.Len())
			} else {
				fmt.Fprintf(os.Stderr, "Written: %d\n", data.Len())
			}
		} else {
			fmt.Fprintln(os.Stderr, "No changes")
		}
	}
}
