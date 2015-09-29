package main

import (
	"fmt"
	"os"
	"flag"
	"regexp"
	"runtime/pprof"

	"github.com/samthor/sre2"
)

var (
	help *bool   = flag.Bool("h", false, "to show help")
	mode *bool   = flag.Bool("m", false, "to run in std mode")
	sub  *bool   = flag.Bool("sub", false, "care about submatches?")
	runs *int    = flag.Int("runs", 100000, "number of runs to do")
	re   *string = flag.String("re", "(a|(b))+", "regexp to build")
	show *bool   = flag.Bool("show", false, "show regexp?")
	s    *string = flag.String("s", "aba", "string to match")
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		return
	}

	if *cpuprofile != "" {
		f, _ := os.Create(*cpuprofile)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	result := false
	alt := []int(nil)

	if !*mode {
		// use sre2
		r := sre2.MustParse(*re)
		if *show {
			r.DebugOut()
		}

		for i := 0; i < *runs; i++ {
			if *sub {
				alt = r.MatchIndex(*s)
				result = (alt != nil)
			} else {
				result = r.Match(*s)
			}
		}
	} else {
		// use existing packaged regexp module
		r := regexp.MustCompile(*re)
		for i := 0; i < *runs; i++ {
			if *sub {
				alt = r.FindStringSubmatchIndex(*s)
				result = (alt != nil)
			} else {
				// NB. This has the same efficiency as FindStringIndex() above, but more closely
				// parallels what we do for SRE2.
				result = r.MatchString(*s)
			}
		}
	}

	fmt.Fprintln(os.Stdout, "result", result, "alt", alt)
}
