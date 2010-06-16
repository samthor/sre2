
package main

/**
 * Simple regexp matcher entry point. Just returns true/false for matching re,
 * and completely ignores submatches.
 *
 * TODO: Eventually, this should still return the outer submatch, assuming all
 * re's are compiled as such: `.*(<re>).*`
 */
func (r *sregexp) RunSimple(src string) bool {
  curr := NewStateSet(len(r.prog), len(r.prog))
  next := NewStateSet(len(r.prog), len(r.prog))

  // always start with state zero
  addstate(curr, r.prog[0])

  for _, ch := range src {
    //fmt.Fprintf(os.Stderr, "%c\t%b\n", rune, curr.bits[0])
    if curr.Length() == 0 {
      return false // no more possible states, short-circuit failure
    }

    // move along rune paths
    for _, st := range curr.Get() {
      i := r.prog[st]
      if i.match(ch) {
        addstate(next, i.out)
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

/**
 * Helper method - just recurses through split/alt states and places them all
 * in the given StateSet.
 */
func addstate(set *StateSet, st *instr) {
  if st == nil || set.Put(st.idx) {
    return // invalid
  }
  if st.mode == kSplit {
    addstate(set, st.out)
    addstate(set, st.out1)
  } else if st.mode == kAltBegin || st.mode == kAltEnd {
    // ignore, just walk over
    addstate(set, st.out)
  }
}
