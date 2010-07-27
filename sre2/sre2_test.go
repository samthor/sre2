
package sre2

import (
  "testing"
  "unicode"
)

func checkState(t *testing.T, state bool, err string) {
  if !state {
    t.Error(err)
  }
}

func TestSimpleRe(t *testing.T) {
  r := Parse("^(a|b)+c*$")
  checkState(t, !r.RunSimple("abd"), "not a valid match")
  checkState(t, r.RunSimple("a"), "basic string should match")
  checkState(t, !r.RunSimple(""), "empty string should not match")
  checkState(t, r.RunSimple("abcccc"), "longer string should match")
}

// Test all rune classes, as defined in util.go.
func TestRuneClass(t *testing.T) {
  c := NewSingleRuneClass('c')
  checkState(t, c.MatchRune('c'), "should match c")
  checkState(t, !c.MatchRune('d'), "should not match d")

  a := NewAnyRuneClass()
  checkState(t, a.MatchRune(-1), "should match anything, even invalid runes")
  checkState(t, a.MatchRune(1245), "should match anything")

  cr := NewComplexRuneClass()
  cr.Include(unicode.Greek)
  cr.ExcludeRune('Π')
  cr.IncludeRune('A')
  checkState(t, !cr.MatchRune('Π'), "should not match pi")
  checkState(t, cr.MatchRune('Ω'), "should match omega")
  checkState(t, !cr.MatchRune('Z'), "should not match regular latin char")
  checkState(t, cr.MatchRune('A'), "should match included latin char")

  cr = NewComplexRuneClass()
  cr.Exclude(unicode.Cyrillic)
  checkState(t, cr.MatchRune('%'), "should match random char, class is exclude-only")
  cr.IncludeRune('Ж')
  checkState(t, !cr.MatchRune('%'), "should no longer match random char")
}

