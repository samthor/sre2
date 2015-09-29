# sre2

Implementation of [RE2](https://github.com/google/re2), done natively in Go. Not related to the native implementation. Handles pathological cases with style and does not backtrack.

There are two available matchers: a fast matcher that does not attempt to track submatches, and a slower matcher that does. Internally, sre2 acts only on runes, not bytes; thus, the only missing part of syntax is `\C` (consume a single byte, even in UTF-8 mode).

The code provides a small library with small suite of tests. The package also includes a tiny main test binary, mostly useful for simple tests and for speed comparisons versus the standard regexp module.

This project was previously hosted on [Google Code](https://code.google.com/p/sre2/).

## Usage

```go
// MustParse will panic on compile failure; useful for init()
m := sre2.MustParse(re)
m, err := sre2.Parse(re)

// Simpler matcher just returns true/false
match := m.Match(str)

// Complex matcher returns indexes of found result: match n will be between (n*2,(n*2)+1).
// The 0th match is reserved for the complete found string. On failure, will return nil.
index := m.MatchIndex(str)

// After this example, fooidx will equal: {3, 12, 3, 6, 7, 12}
foo := sre2.MustParse(`(foo+|bar)\w(.*)`)
fooidx := m.MatchIndex("hi fooo test")
```
