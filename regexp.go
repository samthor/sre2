
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

type altpos struct {
  alt int               // alt index
  is_end bool           // end (true) or begin (false)
  pos int               // character pos
  prev *altpos          // previous in stack
}

/**
 * Mutable match state.
 */
type match struct {
  r *sregexp            // backing regexp
  next *StateSet        // next state set
  cpos int              // current string index
  npos int              // next string index
  allocs int
}

func (m *match) addstate(s *instr, a *altpos) {
  if s == nil {
    return // invalid
  }
  st := m.r.prog[s.idx]
  if st.mode == kSplit {
    m.addstate(st.out, a)
    m.addstate(st.out1, a)
  } else if st.mode == kAltBegin {
    a = &altpos{st.alt, false, m.cpos, a}
    m.allocs += 1
    m.addstate(st.out, a)
  } else if st.mode == kAltEnd {
    a = &altpos{st.alt, true, m.npos, a}
    m.allocs += 1
    m.addstate(st.out, a)
  } else {
    if m.next.Put(s.idx) {
      panic("no dup states")
    }
    // terminal, store altpos in state
  }
}

func (r *sregexp) run(src string) (bool, []pair) {
  m := &match{r, NewStateSet(len(r.prog), len(r.prog)), -1, -1, 0}

  // always start with state zero
  m.addstate(r.prog[0], nil)
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
        m.addstate(i.out, nil)
      }
    }
    curr, m.next = m.next, curr
    m.next.Clear() // clear next so it can be re-used
  }

  for _, st := range curr.Get() {
    if r.prog[st].mode == kMatch {
      return true, nil
    }
  }
  return false, nil
}
