
package sre2

import (
  "utf8"
)

type altpos struct {
  alt int               // alt index
  is_end bool           // end (true) or begin (false)
  pos int               // character pos
  prev *altpos          // previous in stack
}

type pair struct {
  state int
  alt *altpos
}

type m_submatch struct {
  next []pair
  npos int              // next string index
}

func (m *m_submatch) addstate(st *instr, a *altpos) {
  if st == nil {
    return // invalid
  }
  switch st.mode {
  case kSplit:
    m.addstate(st.out, a)
    m.addstate(st.out1, a)
  case kAltBegin:
    a = &altpos{st.alt, false, m.npos, a}
    m.addstate(st.out, a)
  case kAltEnd:
    a = &altpos{st.alt, true, m.npos, a}
    m.addstate(st.out, a)
  case kLeftRight:
    panic("no lr information yet for submatch")
  default:
    // terminal, store (s.idx, altpos) in state
    // note that s.idx won't always be unique (but if both are equal, we could use this)
    pos := len(m.next)
    if pos == cap(m.next) {
      // out of storage, grow to hold onto more states
      hold := m.next
      m.next = make([]pair, pos, pos*2)
      copy(m.next, hold)
    }
    m.next = m.next[0:pos+1]
    m.next[pos] = pair{st.idx, a}
  }
}

// Submatch regexp matcher entry point. Must return both true/false as well
// as information on all submatches.
func (r *sregexp) RunSubMatch(src string) (bool, []int) {
  states_alloc := 64
  m := &m_submatch{make([]pair, 0, states_alloc), 0}
  m.addstate(r.prog[0], nil)
  curr := m.next
  m.next = make([]pair, 0, states_alloc)

  var ch int
  for _, ch = range src {
    m.npos += utf8.RuneLen(ch)

    // move along rune paths
    for _, p := range curr {
      st := r.prog[p.state]
      if st.match(ch) {
        m.addstate(st.out, p.alt)
      }
    }

    curr, m.next = m.next, curr
    m.next = m.next[0:0]
  }

  // Search for a terminal state (in current states). If one is found, allocate
  // and return submatch information for those encountered.
  for _, p := range curr {
    if r.prog[p.state].mode == kMatch {
      alt := make([]int, r.alts*2)
      for i := 0; i < len(alt); i++ {
        // if a particular submatch is not encountered, return -1.
        alt[i] = -1
      }

      a := p.alt
      for a != nil {
        pos := (a.alt * 2)
        if a.is_end {
          pos += 1
        }
        if alt[pos] == -1 {
          alt[pos] = a.pos
        }
        a = a.prev
      }
      return true, alt
    }
  }

  return false, make([]int, 0)
}
