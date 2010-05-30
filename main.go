
package main

import (
  "fmt"
  "os"
  "flag"
  "regexp"
)

var (
  help *bool = flag.Bool("h", false, "to show help")
  mode *bool = flag.Bool("m", false, "to run in std mode")
  runs *int = flag.Int("runs", 100000, "number of runs to do")
  re *string = flag.String("re", "a*a*a*a*a*aaaaa", "regexp to build")
  s *string = flag.String("s", "aaaaa", "string to match")
)

func main() {
  flag.Parse()
  if *help {
    flag.PrintDefaults()
    return
  }

  if !*mode {
    // use new regexp impl
    r := parse(*re)
    fmt.Fprintln(os.Stderr, "r", r)

    result := false
    for i := 0; i < *runs; i++ {
      result = r.run(*s)
    }

    fmt.Fprintln(os.Stderr, "new result", result)
  } else {
    // use old regexp impl
    r := regexp.MustCompile(*re)
    result := false
    for i := 0; i < *runs; i++ {
      result = r.MatchString(*s)
    }
    fmt.Fprintln(os.Stderr, "std result", result)
  }
}
