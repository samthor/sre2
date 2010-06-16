
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
  curr *StateSet        // current state set
  next *StateSet        // next state set (used internally)
  src string            // complete source string
  ch int                // current unicode rune
  cpos int              // current pos in src
  npos int              // next pos in src
  alt []pair
}

/**
 * Store/return the next character in parser. -1 indicates EOF.
 */
func (m *match) nextc() int {
  if m.npos == -1 {
    // do nothing, we are at EOF
  } else if m.npos >= len(m.src) {
    m.cpos = m.npos
    m.npos = -1
    m.ch = -1
  } else {
		c, w := utf8.DecodeRuneInString(m.src[m.npos:])
		m.ch = c
		m.cpos = m.npos
		m.npos += w
  }
  return m.ch
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

func (m *match) step() {
  for _, st := range m.curr.Get() {
    i := m.r.prog[st]
    if i.match(m.ch) {
      m.addstate(i.out)
    }
  }
  m.curr, m.next = m.next, m.curr
  m.next.Clear() // clear next so it can be re-used
}

func (r *sregexp) run(src string) (bool, []pair) {
  m := &match{r, nil, nil, src, -1, -1, 0, make([]pair, r.alts)}
  m.curr = NewStateSet(len(r.prog), len(r.prog))
  m.next = NewStateSet(len(r.prog), len(r.prog))

  // kick off regexp, assert current pos is start of string
  m.nextc()
  if m.cpos != 0 {
    panic("cpos must be zero")
  }

  m.curr.Put(0) // always start at state zero

  for m.ch != -1 {
    //fmt.Fprintf(os.Stderr, "%c\t%b\n", rune, curr.bits[0])
    if m.curr.Length() == 0 {
      return false, nil // short-circuit failure
    }

    // move along rune paths
    m.step()
    m.nextc()
  }

  for _, st := range m.curr.Get() {
    if r.prog[st].mode == kMatch {
      return true, m.alt
    }
  }
  return false, nil
}
