
package sre2

import (
  "container/vector"
  "unicode"
)

type RuneFilter func(rune int) bool

func MatchRune(to_match int) RuneFilter {
  return RuneFilter(func(rune int) bool {
    return rune == to_match
  })
}

func MatchRuneRange(from int, to int) RuneFilter {
  return RuneFilter(func(rune int) bool {
    return rune >= from && rune <= to
  })
}

func MatchUnicodeClass(class string) RuneFilter {
  found := false
  var match vector.Vector
  if len(class) == 1 {
    // A single character is a shorthand request for any category starting with this.
    for key, r := range unicode.Categories {
      if key[0] == class[0] {
        found = true
        match.Push(r)
      }
    }
  } else {
    // Search for the unicode class name inside cats/props/scripts.
    options := []map[string][]unicode.Range{
        unicode.Categories, unicode.Properties, unicode.Scripts}
    for _, option := range options {
      if r, ok := option[class]; ok {
        found = true
        match.Push(r)
      }
    }
  }

  if found {
    return RuneFilter(func(rune int) bool {
      for _, raw := range match {
        r, _ := raw.([]unicode.Range)
        if unicode.Is(r, rune) {
          return true
        }
      }
      return false
    })
  }
  return nil
}

func MatchAsciiClass(class string) RuneFilter {
  r, found := ASCII[class]
  if found {
    return RuneFilter(func(rune int) bool {
      return unicode.Is(r, rune)
    })
  }
  return nil
}

func MergeFilter(filters vector.Vector) RuneFilter {
  return RuneFilter(func(rune int) bool {
    if len(filters) > 0 {
      for _, raw := range filters {
        filter, _ := raw.(RuneFilter)
        if filter(rune) {
          return true
        }
      }
      return false
    }

    // If we haven't merged any filters, don't match (i.e. [] = nothing)
    return false
  })
}

func (r RuneFilter) Not() RuneFilter {
  return RuneFilter(func(rune int) bool {
    return !r(rune)
  })
}

func (r RuneFilter) OptNegate(negate bool) RuneFilter {
  if negate {
    return r.Not()
  }
  return r
}

func (r RuneFilter) IgnoreCase() RuneFilter {
  return RuneFilter(func(rune int) bool {
    return r(unicode.ToLower(rune)) || r(unicode.ToUpper(rune))
  })
}
