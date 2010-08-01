
package sre2 

import (
  "utf8"
)

// Simple regexp matcher entry point. Just returns true/false for matching re,
// and completely ignores submatches.
func (r *sregexp) RunSimple(src string) bool {
  curr := NewStateSet(len(r.prog), len(r.prog))
  next := NewStateSet(len(r.prog), len(r.prog))

  // TODO: Fix this very short-term approach to left/right matching!
  var left int = -1
  var right int
  if len(src) > 0 {
    right, _ = utf8.DecodeRuneInString(src)
  } else {
    right = -1
  }

  // always start with state zero
  addstate(curr, r.prog[0], -1, right)

  for p, ch := range src {
    left = right
    p += utf8.RuneLen(ch)
    if p < len(src) {
      right, _ = utf8.DecodeRuneInString(src[p:len(src)])
    } else {
      right = -1
    }

    //fmt.Fprintf(os.Stderr, "%c\t%b\n", rune, curr.bits[0])
    if curr.Length() == 0 {
      return false // no more possible states, short-circuit failure
    }

    // move along rune paths
    for _, st := range curr.Get() {
      i := r.prog[st]
      if i.match(ch) {
        addstate(next, i.out, left, right)
      }
    }
    curr, next = next, curr
    next.Clear() // clear next so it can be re-used
  }

  // search for matching state
  for _, st := range curr.Get() {
    if r.prog[st].mode == kMatch {
      return true
    }
  }
  return false
}

// Helper method - just descends through split/alt states and places them all
// in the given StateSet.
func addstate(set *StateSet, st *instr, left int, right int) {
  if st == nil || set.Put(st.idx) {
    return // invalid
  }
  switch st.mode {
  case kSplit:
    addstate(set, st.out, left, right)
    addstate(set, st.out1, left, right)
  case kAltBegin, kAltEnd:
    // ignore, just walk over
    addstate(set, st.out, left, right)
  case kLeftRight:
    if st.matchLeftRight(left, right) {
      addstate(set, st.out, left, right)
    }
  }
}
