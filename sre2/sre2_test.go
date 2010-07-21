
package sre2

import "testing"

func checkState(t *testing.T, state bool, err string) {
  if !state {
    t.Error(err)
  }
}

func TestSimpleRe(t *testing.T) {
  r := Parse("^(a|b)+c*$")
  r.DebugOut()
  checkState(t, !r.RunSimple("abd"), "not a valid match")
  checkState(t, r.RunSimple("a"), "basic string should match")
  checkState(t, r.RunSimple("abcccc"), "longer string should match")
}

func TestRuneClass(t *testing.T) {
  c := NewSingleRuneClass('c')
  checkState(t, c.MatchRune('c'), "should match c")
  checkState(t, !c.MatchRune('d'), "should not match d")
}
