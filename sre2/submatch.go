package sre2

type altpos struct {
	pos   int     // position
	index int     // rune index
	prev  *altpos // previous in stack
}

type pair struct {
	state int
	alt   *altpos
}

type m_submatch struct {
	parser SafeReader
	next   []pair
}

func (m *m_submatch) addstate(st *instr, a *altpos) {
	if st == nil {
		return // invalid
	}
	switch st.mode {
	case iSplit:
		m.addstate(st.out, a)
		m.addstate(st.out1, a)
	case iIndexCap:
		m.addstate(st.out, &altpos{st.cid, m.parser.npos(), a})
	case iBoundaryCase:
		if st.matchBoundaryMode(m.parser.curr(), m.parser.peek()) {
			m.addstate(st.out, a)
		}
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
		m.next = m.next[0 : pos+1]
		m.next[pos] = pair{st.idx, a}
	}
}

// MatchIndex is the top-level complex matcher used in sre2, where submatches
// are recorded. This method will return a list of ints indicating the match
// positions; indexes 0:1 represent the entire match, and n*2:n*2+1 store the
// start and end index of each alt. Returns nil if not match found.
func (r *sregexp) MatchIndex(src string) []int {
	states_alloc := 64
	m := &m_submatch{NewSafeReader(src), make([]pair, 0, states_alloc)}
	m.addstate(r.prog[0], nil)
	curr := m.next
	m.next = make([]pair, 0, states_alloc)

	for m.parser.nextCh() != -1 {
		ch := m.parser.curr()

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
		if r.prog[p.state].mode == iMatch {
			alt := make([]int, r.caps*2)
			for i := 0; i < len(alt); i++ {
				// if a particular submatch is not encountered, return -1.
				alt[i] = -1
			}

			a := p.alt
			for a != nil {
				if alt[a.pos] == -1 {
					alt[a.pos] = a.index
				}
				a = a.prev
			}
			return alt
		}
	}

	return nil
}
