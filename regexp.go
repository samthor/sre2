
package main

import (
  "utf8"
)

/* regexp - currently just a list of states */
type sregexp struct {
  prog []*instr // list of states
  alts int      // number of marked alts [()'s] in this regexp
}

type pair struct {
  begin int
  end int
}

/**
 * Mutable match state.
 */
type match struct {
  r *sregexp
  src string
  ch int
  cpos int
  npos int
  alt []pair
}

/**
 * Store/return the next character in parser. -1 indicates EOF.
 */
func (m *match) nextc() int {
  if m.npos >= len(m.src) {
    m.ch = -1
  } else {
		c, w := utf8.DecodeRuneInString(m.src[m.npos:])
		m.ch = c
		m.cpos = m.npos
		m.npos += w
  }
  return m.ch
}

func (r *sregexp) addstate(o *StateSet, s *instr, m *match) {
  if s == nil || o.Put(s.idx) {
    return // invalid, or already have this state
  }
  st := r.prog[s.idx]
  if st.mode == kSplit {
    r.addstate(o, st.out, m)
    r.addstate(o, st.out1, m)
  } else if st.mode == kAltBegin {
    // TODO: broken, stores global state, not branch state
    m.alt[st.alt].begin = m.cpos
    r.addstate(o, st.out, m)
  } else if st.mode == kAltEnd {
    // TODO: broken, stores global state, not branch state
    m.alt[st.alt].end = m.npos
    r.addstate(o, st.out, m)
  }
}

func (r *sregexp) next(curr *StateSet, next *StateSet, m *match) (r_curr *StateSet, r_next *StateSet) {
  for _, st := range curr.Get() {
    if r.prog[st].match(m.ch) {
      r.addstate(next, r.prog[st].out, m)
    }
  }
  curr.Clear() // clear curr so it can be re-used by caller
  return next, curr
}

func (r *sregexp) run(src string) (bool, []pair) {
  curr := NewStateSet(len(r.prog), len(r.prog))
  next := NewStateSet(len(r.prog), len(r.prog))

  m := &match{r, src, -1, -1, 0, make([]pair, r.alts)}
  m.nextc()
  if m.cpos != 0 {
    panic("cpos must be zero")
  }

  r.addstate(curr, r.prog[0], m)

  for m.ch != -1 {
    //fmt.Fprintf(os.Stderr, "%c\t%b\n", rune, curr.bits[0])
    if curr.Length() == 0 {
      return false, nil // short-circuit failure
    }

    // move along rune paths
    curr, next = r.next(curr, next, m)
    m.nextc()
  }

  for _, st := range curr.Get() {
    if r.prog[st].mode == kMatch {
      return true, m.alt
    }
  }
  return false, nil
}
