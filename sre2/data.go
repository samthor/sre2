
package sre2

import (
  "container/vector"
  "fmt"
  "unicode"
)

// Generic rune matcher. Provides the single MatchRune method.
type RuneClass struct {
  include vector.Vector
  exclude vector.Vector
}

func NewRuneClass() *RuneClass {
  return &RuneClass{}
}

func (c *RuneClass) _Push(negate bool, v interface{}) {
  if !negate {
    c.include.Push(v)
  } else {
    c.exclude.Push(v)
  }
}

func (c *RuneClass) AddRune(negate bool, rune int) {
  c._Push(negate, rune)
}

func (c *RuneClass) AddFunc(negate bool, fn func(rune int) bool) {
  c._Push(negate, fn)
}

func (c *RuneClass) AddRuneRange(negate bool, low int, high int) {
  c._Push(negate, unicode.Range{low, high, 1})
}

func (c *RuneClass) AddUnicodeClass(negate bool, class string) bool {
  found := false

  if len(class) == 1 {
    for key, r := range unicode.Categories {
      if key[0] == class[0] {
        found = true
        c._Push(negate, r)
      }
    }
  } else {
    if r, ok := unicode.Categories[class]; ok {
      c._Push(negate, r)
      found = true
    }
    if r, ok := unicode.Properties[class]; ok {
      c._Push(negate, r)
      found = true
    }
    if r, ok := unicode.Scripts[class]; ok {
      c._Push(negate, r)
      found = true
    }
  }

  return found
}

func (c *RuneClass) AddAsciiClass(negate bool, class string) bool {
  v, found := ASCII[class]
  if found {
    c._Push(negate, v)
  }
  return found
}

func (c *RuneClass) AddRuneClass(negate bool, other *RuneClass) {
  for _, v := range other.include {
    c._Push(negate, v)
  }
  for _, v := range other.exclude {
    c._Push(!negate, v)
  }
}

func (c *RuneClass) MatchRune(rune int) bool {
  // Default is to match. If we find runes to include, then the default will
  // transition to false.
  result := true

  // Search through all included runes, and break if we find a match.
  for _, v := range c.include {
    result = false
    if match(rune, v) {
      result = true
      break
    }
  }

  // If the result could be true, iterate through all excluded runes and fail
  // immediately if we find a counter-example.
  if result {
    for _, v := range c.exclude {
      if match(rune, v) {
        result = false
        break
      }
    }
  }

  return result
}

func match(rune int, v interface{}) bool {
  switch x := v.(type) {
  case int:
    return x == rune
  case func(rune int) bool:
    return x(rune)
  case unicode.Range:
    return rune >= x.Lo && rune <= x.Hi && ((rune - x.Lo) % x.Stride == 0)
  case []unicode.Range:
    return unicode.Is(x, rune)
  }
  panic(fmt.Sprintf("unexpected: %s", v))
}
