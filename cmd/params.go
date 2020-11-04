package cmd

import (
	"flag"
	"fmt"
	"github.com/savvdm/checksum/lib"
	"os"
)

type Params struct {
	Mode     CheckMode
	Includes lib.Patterns
	Excludes lib.Patterns
	Verbose  bool
	Quiet    bool
	Nostat   bool
	Dry      bool
	Outfile  string
	Delete   bool
	Profile  string
}

func help() {
	fmt.Printf("Usage: %s [params] checksum_file dir_to_check\n", os.Args[0])
	flag.PrintDefaults()
}

func (params *Params) Init() {
	params.Mode = Modified // default mode
	flag.Var(&params.Mode, "check", "Check mode: new|modified|all")
	flag.Var(&params.Includes, "include", "File name pattern to include (default is all)")
	flag.Var(&params.Excludes, "exclude", "File name pattern to exclude (default is none)")
	flag.BoolVar(&params.Verbose, "v", false, "More detailed output")
	flag.BoolVar(&params.Quiet, "q", false, "Less detailed output")
	flag.BoolVar(&params.Nostat, "nostat", false, "Don't print stats")
	flag.BoolVar(&params.Dry, "n", false, "Don't save changes (dry run)")
	flag.StringVar(&params.Outfile, "outfile", "", "Output file name")
	flag.BoolVar(&params.Delete, "delete", false, "Delete files missing on disk from the data file")
	flag.StringVar(&params.Profile, "cpuprofile", "", "write cpu profile to file")
}

func (params *Params) Parse() (dataFile string, root string) {
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

func (params *Params) ReportOK() bool {
	switch params.Mode {
	case Modified:
		return !params.Quiet
	case All:
		return params.Verbose
	}
	return false
}
