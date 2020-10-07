package main

import "flag"

type cmdParams struct {
	mode     checkMode
	excludes excludePatterns
	verbose  bool
	quiet    bool
	nostat   bool
	dry      bool
}

func (params *cmdParams) init() {
	params.mode = Modified // default mode
	flag.Var(&params.mode, "check", "Check mode: new|modified|all")
	flag.Var(&params.excludes, "exclude", "File name pattern to exclude")
	flag.BoolVar(&params.verbose, "v", false, "More detailed output")
	flag.BoolVar(&params.quiet, "q", false, "Less detailed output")
	flag.BoolVar(&params.nostat, "nostat", false, "Don't print stats")
	flag.BoolVar(&params.dry, "n", false, "Don't save changes (dry run)")
	flag.Parse()
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
