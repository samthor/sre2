package sre2

func (r *sregexp) Match(src string) bool {
	success, _ := r.run(src, false)
	return success
}

func (r *sregexp) MatchIndex(src string) []int {
	_, capture := r.run(src, true)
	return capture
}

func (r *sregexp) run(src string, submatch bool) (success bool, capture []int) {
	curr := makeStateList(len(r.prog))
	next := makeStateList(len(r.prog))
	parser := NewSafeReader(src)

	// always start with state zero
	curr.addstate(&parser, r.prog[0], submatch, nil)

	for parser.nextCh() != -1 {
		ch := parser.curr()
		if len(curr.states) == 0 {
			return false, nil // no more possible states, short-circuit failure
		}

		// move along rune paths
		for _, st := range curr.states {
			i := r.prog[st.idx]
			if i.match(ch) {
				next.addstate(&parser, i.out, submatch, st.capture)
			}
		}
		curr, next = next, curr
		next.clear() // clear next so it can be re-used
	}

	// search for success state
	for _, st := range curr.states {
		if r.prog[st.idx].mode == iMatch {
			return true, st.capture.list(r.caps)
		}
	}
	return false, nil
}

// stateList is used by regexp.run() to efficiently maintain an ordered list of
// current/next regexp integer states.
type stateList struct {
	sparse []int
	states []state
}

// state represents a state index and captureInfo pair.
type state struct {
	idx     int
	capture *captureInfo
}

// makeStateList builds a new ordered bitset for use in the regexp.
func makeStateList(states int) *stateList {
	return &stateList{make([]int, states), make([]state, 0, states)}
}

// addstate descends through split/alt states and places them all in the
// given stateList.
func (o *stateList) addstate(p *SafeReader, st *instr, submatch bool, capture *captureInfo) {
	if st == nil || o.put(st.idx, capture) {
		return // instr does not exist, or state already in set: fall out
	}

	switch st.mode {
	case iSplit:
		o.addstate(p, st.out, submatch, capture)
		o.addstate(p, st.out1, submatch, capture)
	case iIndexCap:
		if submatch {
			capture = capture.push(p.npos(), st.cid)
		}
		o.addstate(p, st.out, submatch, capture)
	case iBoundaryCase:
		if st.matchBoundaryMode(p.curr(), p.peek()) {
			o.addstate(p, st.out, submatch, capture)
		}
	}
}

// put places the given state into the stateList. Returns true if the state was
// previously set, and false if it was not.
func (o *stateList) put(v int, capture *captureInfo) bool {
	pos := len(o.states)
	if o.sparse[v] < pos && o.states[o.sparse[v]].idx == v {
		return true // already exists
	}

	o.states = o.states[:pos+1]
	o.sparse[v] = pos
	o.states[pos].idx = v
	o.states[pos].capture = capture
	return false
}

// clear resets the stateList to be re-used.
func (o *stateList) clear() {
	o.states = o.states[0:0]
}

// captureInfo represents the submatch information for a given run. This is represented as a linked
// list so that early states can be shared; however there's more cost in GC.
type captureInfo struct {
	c    int          // capture index
	pos  int          // position in string
	prev *captureInfo // previous node in list, or nil
}

// push adds a new head to the existing submatch information, returning it. Note that the receiver
// here may be nil.
func (info *captureInfo) push(pos int, c int) *captureInfo {
	// TODO: If we traverse back and remove previous instances of this capture group, then we might
	// remove information used by other branches.
	return &captureInfo{c, pos, info}
}

// list translates the given submatch state into a concrete []int for use by callers.
func (info *captureInfo) list(size int) (ret []int) {
	ret = make([]int, size<<1)
	for i := 0; i < len(ret); i++ {
		ret[i] = -1
	}
	for info != nil {
		if ret[info.c] == -1 {
			ret[info.c] = info.pos
		}
		info = info.prev
	}
	return ret
}