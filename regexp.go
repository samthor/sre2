
package main

import (
  "utf8"
)

/* regexp - currently just a list of states */
type sregexp struct {
  prog []*instr         // list of states
  alts int              // number of marked alts [()'s] in this regexp
}

type pair struct {
  begin int
  end int
}

/**
 * Mutable match state.
 */
type match struct {
  r *sregexp            // backing regexp
  next *StateSet        // next state set
  cpos int              // current string index
  npos int              // next string index
  alt []pair            // alts (TODO: broken)
}

func (m *match) addstate(s *instr) {
  if s == nil || m.next.Put(s.idx) {
    return // invalid, or already have this state
  }
  st := m.r.prog[s.idx]
  if st.mode == kSplit {
    m.addstate(st.out)
    m.addstate(st.out1)
  } else if st.mode == kAltBegin {
    // TODO: broken, stores global state, not branch state
    m.alt[st.alt].begin = m.cpos
    m.addstate(st.out)
  } else if st.mode == kAltEnd {
    // TODO: broken, stores global state, not branch state
    m.alt[st.alt].end = m.npos
    m.addstate(st.out)
  }
}

func (r *sregexp) run(src string) (bool, []pair) {
  m := &match{r, NewStateSet(len(r.prog), len(r.prog)), -1, -1, make([]pair, r.alts)}

  // always start with state zero
  m.addstate(r.prog[0])
  curr := m.next
  m.next = NewStateSet(len(r.prog), len(r.prog))

  var ch int
  for m.cpos, ch = range src {
     m.npos = m.cpos + utf8.RuneLen(ch)

    //fmt.Fprintf(os.Stderr, "%c\t%b\n", rune, curr.bits[0])
    if curr.Length() == 0 {
      return false, nil // short-circuit failure
    }

    // move along rune paths
    for _, st := range curr.Get() {
      i := r.prog[st]
      if i.match(ch) {
        m.addstate(i.out)
      }
    }
    curr, m.next = m.next, curr
    m.next.Clear() // clear next so it can be re-used
  }

  for _, st := range curr.Get() {
    if r.prog[st].mode == kMatch {
      return true, m.alt
    }
  }
  return false, nil
}
