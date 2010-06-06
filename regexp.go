
package main

/* regexp - currently just a list of states */
type sregexp struct {
  prog []*instr
}

func (r *sregexp) addstate(o *StateSet, s *instr) {
  if s == nil || o.Put(s.idx) {
    return // invalid, or already have this state
  }
  st := r.prog[s.idx]
  if st.mode == kSplit {
    r.addstate(o, st.out)
    r.addstate(o, st.out1)
  }
}

func (r *sregexp) next(curr *StateSet, next *StateSet, rune int) (r_curr *StateSet, r_next *StateSet) {
  for _, st := range curr.Get() {
    if r.prog[st].match(rune) {
      r.addstate(next, r.prog[st].out)
    }
  }
  curr.Clear() // clear curr so it can be re-used by caller
  return next, curr
}

func (r *sregexp) run(str string) bool {
  curr := NewStateSet(len(r.prog), len(r.prog))
  next := NewStateSet(len(r.prog), len(r.prog))
  r.addstate(curr, r.prog[0])

  for _, rune := range str {
    //fmt.Fprintf(os.Stderr, "%c\t%b\n", rune, curr.bits[0])
    if curr.Length() == 0 {
      return false // short-circuit failure
    }

    // move along rune paths
    curr, next = r.next(curr, next, rune)
  }

  for _, st := range curr.Get() {
    if r.prog[st].mode == kMatch {
      return true
    }
  }
  return false
}
