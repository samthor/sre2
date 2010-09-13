
package main

import (
  "fmt"
  "os"
  "flag"
  "regexp"
  "sre2"
)

var (
  help *bool = flag.Bool("h", false, "to show help")
  mode *bool = flag.Bool("m", false, "to run in std mode")
  sub *bool = flag.Bool("sub", false, "care about submatches?")
  runs *int = flag.Int("runs", 100000, "number of runs to do")
  re *string = flag.String("re", "(a|(b))+", "regexp to build")
  s *string = flag.String("s", "aba", "string to match")
)

func main() {
  flag.Parse()
  if *help {
    flag.PrintDefaults()
    return
  }

  if !*mode {
    // use sre2
    r := sre2.MustParse(*re)
    r.DebugOut()

    result := false
    var alt []int
    for i := 0; i < *runs; i++ {
      if *sub {
        result, alt = r.RunSubMatch(*s)
      } else {
        result = r.RunSimple(*s)
      }
    }

    fmt.Fprintln(os.Stdout, "new result", result, "alt", alt)
  } else {
    // use existing packaged regexp module
    r := regexp.MustCompile(*re)
    result := false
    var alt []int
    for i := 0; i < *runs; i++ {
      if *sub {
        alt = r.FindStringIndex(*s)
        result = (alt != nil)
      } else {
        // NB. This has the same efficiency as FindStringIndex() above, but more closely
        // parallels what we do for SRE2.
        result = r.MatchString(*s)
      }
    }
    fmt.Fprintln(os.Stdout, "std result", result, "alt", alt)
  }
}
