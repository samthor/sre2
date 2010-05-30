
package main

import (
  "container/list"
)

const (
  kSplit = iota         // proceed down out & out1
  kRune                 // if match rune, proceed down out
  kCall                 // if matcher passes, proceed down out
  kMatch                // success state!
)

type State struct {
  me int                // index of this state
  mode byte             // mode (as above)
  out int               // to next state index
  out1 int              // to alt next state index (only for kSplit)
  rune int              // rune to match (only for kRune)
  matcher func(rune int) bool   // matcher method (for kCall)
}

/* generic matcher, implements definitions of above consts */
func (s *State) match(rune int) bool {
  if s.mode == kRune {
    return s.rune == rune || s.rune == -1
  } else if s.mode == kCall {
    return s.matcher(rune)
  }
  return false
}

/* regexp - currently just a list of states */
type sregexp struct {
  prog []State
}

func (r *sregexp) addstate(o *obitset, s int) {
  if s < 0 || o.Put(s) {
    return // invalid, or already have this state
  }
  st := r.prog[s]
  if st.mode == kSplit {
    r.addstate(o, st.out)
    r.addstate(o, st.out1)
  }
}

func (r *sregexp) next(curr *obitset, next *obitset, rune int) (r_curr *obitset, r_next *obitset) {
  for _, st := range curr.Get() {
    if r.prog[st].match(rune) {
      r.addstate(next, r.prog[st].out)
    }
  }
  curr.Clear() // clear curr so it can be re-used by caller
  return next, curr
}

func (r *sregexp) run(str string) bool {
  curr := NewStateSet(64, 64)
  next := NewStateSet(64, 64)
  r.addstate(curr, 0)

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

/** generates simple NFA */
// TODO: all broken!
func parse(str string) sregexp {
  stack := list.New()
  var next int

  // dummy initial split
  stack.PushBack(&State{next, kSplit, -1, -1, -1, nil})
  next += 1

  for _, rune := range str {
    switch rune {
    case '*':
      // zero or many
      p := stack.Back()
      p.Value.(*State).me = next
      stack.InsertBefore(&State{next - 1, kSplit, next, next + 1, -1, nil}, p)
      p.Value.(*State).out = next - 1
      next += 1
      stack.PushBack(&State{next, kSplit, -1, -1, -1, nil})
    case '?':
      // zero or one
    case '+':
      // one or many (i.e. a+ = aa*)
    case '.':
      // any char
      p := stack.Back()
      p.Value.(*State).out = next
      stack.PushBack(&State{next, kRune, -1, -1, -1, nil})
    default:
      // literal match
      p := stack.Back()
      p.Value.(*State).out = next
      stack.PushBack(&State{next, kRune, -1, -1, rune, nil})
    }
    next += 1
  }
  p := stack.Back()
  p.Value.(*State).out = next
  stack.PushBack(&State{next, kMatch, -1, -1, -1, nil})

  prog := make([]State, stack.Len())
  front := stack.Front()
  for i := 0; i < stack.Len(); i++ {
    prog[i] = *front.Value.(*State)
    front = front.Next()
  }
  return sregexp{prog}
}
