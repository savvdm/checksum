package main

import (
	"flag"
	"fmt"
	"os"
)

type cmdParams struct {
	mode       checkMode
	includes   Patterns
	excludes   Patterns
	verbose    bool
	quiet      bool
	nostat     bool
	dry        bool
	outfile    string
	delete     bool
	cpuprofile string
}

func help() {
	fmt.Printf("Usage: %s [params] checksum_file dir_to_check\n", os.Args[0])
	flag.PrintDefaults()
}

func (params *cmdParams) init() {
	params.mode = Modified // default mode
	flag.Var(&params.mode, "check", "Check mode: new|modified|all")
	flag.Var(&params.includes, "include", "File name pattern to include (default is all)")
	flag.Var(&params.excludes, "exclude", "File name pattern to exclude (default is none)")
	flag.BoolVar(&params.verbose, "v", false, "More detailed output")
	flag.BoolVar(&params.quiet, "q", false, "Less detailed output")
	flag.BoolVar(&params.nostat, "nostat", false, "Don't print stats")
	flag.BoolVar(&params.dry, "n", false, "Don't save changes (dry run)")
	flag.StringVar(&params.outfile, "outfile", "", "Output file name")
	flag.BoolVar(&params.delete, "delete", false, "Delete files missing on disk from the data file")
	flag.StringVar(&params.cpuprofile, "cpuprofile", "", "write cpu profile to file")
}

func (params *cmdParams) parse() (dataFile string, root string) {
	flag.Parse()

	narg := flag.NArg()
	if narg < 1 || narg > 2 {
		help()
		os.Exit(1)
	}

	dataFile = flag.Arg(0)

	if narg > 1 {
		root = flag.Arg(1)
	} else {
		root = "."
	}

	return
}

func (params *cmdParams) reportOk() bool {
	switch params.mode {
	case Modified:
		return !params.quiet
	case All:
		return params.verbose
	}
	return false
}
