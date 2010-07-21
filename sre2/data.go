
package sre2

import (
  "container/vector"
  "unicode"
)

// Generic rune matcher.
type runeclass interface {
  MatchRune(rune int) bool
}

// Rune class that matches any rune, i.e. "." regexp.
type any_runeclass struct {}

func (c any_runeclass) MatchRune(rune int) bool {
  return true
}

func NewAnyRuneClass() runeclass {
  return any_runeclass{}
}

// Rune class that matches a single positive rune.
type single_runeclass int

func (c single_runeclass) MatchRune(rune int) bool {
  return rune == int(c)
}

func NewSingleRuneClass(rune int) runeclass {
  if rune <= 0 {
    panic("expected non-zero positive rune")
  }
  return single_runeclass(rune)
}

// Complex rune class; may be used to represent a complete [...] character class
// from regexp. Boils down to included and excluded rune sets.
type complex_runeclass struct {
  include vector.Vector
  exclude vector.Vector
}

func (c complex_runeclass) MatchRune(rune int) bool {
  // Default is to match. If we find runes to include, then the default will
  // transition to false.
  result := true

  // Search through all included runes, and break if we find a match.
  for _, raw := range c.include {
    r, _ := raw.([]unicode.Range)
    result = false
    if unicode.Is(r, rune) {
      result = true
      break
    }
  }

  // If the result could be true, iterate through all excluded runes and fail
  // immediately if we find a counter-example.
  if result {
    for _, raw := range c.exclude {
      r, _ := raw.([]unicode.Range)
      if unicode.Is(r, rune) {
        result = false
        break
      }
    }
  }

  return result
}

func (c complex_runeclass) Include(r []unicode.Range) {
  c.include.Push(r)
}

func (c complex_runeclass) Exclude(r []unicode.Range) {
  c.exclude.Push(r)
}

func NewComplexRuneClass() complex_runeclass {
  return complex_runeclass{make(vector.Vector, 0), make(vector.Vector, 0)}
}
