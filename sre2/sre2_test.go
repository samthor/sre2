
package sre2

import (
  "testing"
  "unicode"
)

// Check the given state to be true.
func checkState(t *testing.T, state bool, err string) {
  if !state {
    t.Error(err)
  }
}

// Check the equality of two []int slices.
func checkIntSlice(t *testing.T, expected []int, result []int, err string) {
  match := true
  if len(expected) != len(result) {
    match = false
  } else {
    for i := 0; i < len(expected); i++ {
      if expected[i] != result[i] {
        match = false
      }
    }
  }
  checkState(t, match, err)
}

// Run a selection of basic regular expressions against this package.
func TestSimpleRe(t *testing.T) {
  r := Parse("^(a|b)+c*$")
  checkState(t, !r.RunSimple("abd"), "not a valid match")
  checkState(t, r.RunSimple("a"), "basic string should match")
  checkState(t, !r.RunSimple(""), "empty string should not match")
  checkState(t, r.RunSimple("abcccc"), "longer string should match")
}

// Test closure expansion types, such as {..}, ?, +, * etc.
func TestClosureExpansion(t *testing.T) {
  r := Parse("^za?$")
  checkState(t, r.RunSimple("z"), "should match none")
  checkState(t, r.RunSimple("za"), "should match single")
  checkState(t, !r.RunSimple("zaa"), "should not match more")

  r = Parse("^a{2,2}$")
  checkState(t, !r.RunSimple(""), "0 should fail")
  checkState(t, !r.RunSimple("a"), "1 should fail")
  checkState(t, r.RunSimple("aa"), "2 should succeed")
  checkState(t, r.RunSimple("aaa"), "3 should succeed")
  checkState(t, r.RunSimple("aaaa"), "4 should succeed")
  checkState(t, !r.RunSimple("aaaaa"), "5 should fail")

  r = Parse("^a{2}$")
  checkState(t, !r.RunSimple(""), "0 should fail")
  checkState(t, !r.RunSimple("a"), "1 should fail")
  checkState(t, r.RunSimple("aa"), "2 should succeed")
  checkState(t, !r.RunSimple("aaa"), "3 should fail")

  r = Parse("^a{3,}$")
  checkState(t, !r.RunSimple("aa"), "2 should fail")
  checkState(t, r.RunSimple("aaa"), "3 should succeed")
  checkState(t, r.RunSimple("aaaaaa"), "more should succeed")
}

// Test specific greedy/non-greedy closure types.
func TestClosureGreedy(t *testing.T) {
  r := Parse("^(a{0,2}?)(a*)$")
  ok, res := r.RunSubMatch("aaa")
  checkState(t, ok, "should match")
  checkIntSlice(t, []int{0, 3, 0, 0, 0, 3}, res, "did not match expected")

  r = Parse("^(a{0,2})?(a*)$")
  ok, res = r.RunSubMatch("aaa")
  checkState(t, ok, "should match")
  checkIntSlice(t, []int{0, 3, 0, 2, 2, 3}, res, "did not match expected")

  r = Parse("^(a{2,}?)(a*)$")
  ok, res = r.RunSubMatch("aaa")
  checkState(t, ok, "should match")
  checkIntSlice(t, []int{0, 3, 0, 2, 2, 3}, res, "did not match expected")
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

