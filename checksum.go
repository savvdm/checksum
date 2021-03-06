package main

import (
	"fmt"
	"github.com/savvdm/checksum/cmd"
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

	var params cmd.Params
	params.Init()

	dataFile, root := params.Parse()

	// setup cpu profiling
	if params.Profile != "" {
		startProfiling(params.Profile)
		defer pprof.StopCPUProfile()
	}

	data := make(lib.Data)

	inputMod := data.Read(dataFile)
	if !params.Nostat {
		fmt.Fprintf(os.Stderr, "Read: %d\n", data.Len())
	}

	var numWorkers = runtime.GOMAXPROCS(0)
	in, out := lib.StartWorkers(numWorkers)

	var stats lib.StatCounts

	// process checksum result
	update := func(res *lib.CheckResult) {
		if res.Err == nil {
			data.Update(res.File, res.Sum)
		} else {
			stats.ReportError(res.Err)
		}
	}

	if err := lib.ReadDir(root, "", func(file string, mod time.Time) {
		// check includes (if any)
		if len(params.Includes) > 0 && !params.Includes.Match(file) {
			return
		}
		// check excludes (if any)
		if len(params.Excludes) > 0 && params.Excludes.Match(file) {
			stats.ReportIf(params.Verbose, lib.Skipped, file)
			return
		}
		// mark the file visited (and see if it exists)
		stats.Register(lib.Visited)
		exists := data.MarkVisited(file)
		// see if we need to calculate a checksum
		force := params.Mode == cmd.All || params.Mode == cmd.Modified && mod.After(inputMod)
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
	}); err != nil { // ReadDir failed
		println(err)
		os.Exit(2)
	}

	// no more checksum calculations will be queued
	close(in)

	// read calculated checksums
	for res := range out {
		if res == nil {
			if numWorkers--; numWorkers == 0 {
				break
			}
		} else {
			update(res)
		}
	}

	// sort keys and count stats
	// mark unvisited files for deletion
	files := data.Finalize(params.Delete, &stats)

	// print out file status
	report := func(file string, status lib.Status) {
		lib.ReportStatus(file, status, params.ReportOK())
	}

	// write data & report file status
	changed := stats.HasChanged()
	if !params.Dry && changed { // don't write file unless anything changed
		outfile := dataFile
		if len(params.Outfile) > 0 {
			outfile = params.Outfile
		}
		data.WriteFiles(files, outfile, report)
	} else {
		data.ReportFiles(files, report)
	}

	// report stats
	if !params.Nostat {
		stats.Print()
		if changed {
			if params.Dry {
				fmt.Fprintf(os.Stderr, "Dry run, not written: %d\n", data.Len())
			} else {
				fmt.Fprintf(os.Stderr, "Written: %d\n", data.Len())
			}
		} else {
			fmt.Fprintln(os.Stderr, "No changes")
		}
	}
}
