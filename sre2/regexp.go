
package sre2

import (
  "fmt"
  "os"
  "strconv"
  "strings"
  "unicode"
  "utf8"
)

/**
 * Regexp definition. Simple, just a list of states.
 */
type sregexp struct {
  prog []*instr         // list of states
  alts int              // number of marked alts [()'s] in this regexp
}

/**
 * Instruction type definitions, for `instr.mode`.
 */
const (
  kSplit = iota         // proceed down out & out1
  kAltBegin             // begin of alt section, i.e. '('
  kAltEnd               // end of alt section, i.e. ')'
  kRuneClass            // if match rune, proceed down out
  kMatch                // success state!
)

/**
 * Single instruction in regexp.
 */
type instr struct {
  idx int               // index of this instr
  mode byte             // mode (as above)
  out *instr            // next instr to process
  out1 *instr           // alt next instr (for kSplit)
  rune runeclass        // rune class
  alt int               // identifier of alt branch (for kAlt{Begin,End})
  alt_id *string        // string identifier of alt branch
}

/**
 * String-representation of an individual instruction.
 */
func (i *instr) str() string {
  str := fmt.Sprintf("{%d", i.idx)
  out := ""
  if i.out != nil {
    out += fmt.Sprintf(" out=%d", i.out.idx)
  }
  switch i.mode {
  case kSplit:
    str += " kSplit"
    if i.out1 != nil {
      out += fmt.Sprintf(" out1=%d", i.out1.idx)
    }
  case kAltBegin, kAltEnd:
    if i.mode == kAltBegin {
      str += " kAltBegin"
    } else {
      str += " kAltEnd"
    }
    str += fmt.Sprintf(" alt=%d", i.alt)
    if i.alt_id != nil {
      str += fmt.Sprintf(" alt_id=%s", *i.alt_id)
    }
  case kRuneClass:
    str += fmt.Sprint(" kRuneClass ", i.rune)
  case kMatch:
    str += " kMatch"
  }
  return str + out + "}"
}

/**
 * DebugOut writes the given regexp to Stderr, for debugging.
 */
func (r *sregexp) DebugOut() {
  for i := 0; i < len(r.prog); i++ {
    fmt.Fprintln(os.Stderr, i, r.prog[i].str())
  }
}

/*
 * Generic matcher for consuming instr instances (i.e. kRune/kCall). Does not
 * match anything else.
 */
func (s *instr) match(rune int) bool {
  return s.mode == kRuneClass && s.rune.MatchRune(rune)
}

/** transient parser state */
type parser struct {
  re *sregexp
  src string
  ch int
  pos int
}

/**
 * Generate a new pre-indexed instr.
 */
func (p *parser) instr() *instr {
  pos := len(p.re.prog)
  if pos == cap(p.re.prog) {
    if pos == 0 {
      panic("should not have cap of zero")
    }
    local := p.re.prog
    p.re.prog = make([]*instr, pos, pos * 2)
    copy(p.re.prog, local)
  }
  p.re.prog = p.re.prog[0:pos+1]
  i := &instr{pos, kSplit, nil, nil, nil, -1, nil}
  p.re.prog[pos] = i
  return i
}

// Store and return the next rune in the parser. -1 is EOF.
func (p *parser) nextc() int {
  if p.pos == -1 {
    p.pos = 0
  } else if p.ch != -1 {
    p.pos += utf8.RuneLen(p.ch)
    if p.pos >= len(p.src) {
      p.ch = -1
      return -1
    }
  } else {
    return -1
  }

  p.ch, _ = utf8.DecodeRuneInString(p.src[p.pos:])
  return p.ch
}

// Jump to the given position within the parser's contained string. The rune
// at this location will be set in p.ch and returned. -1 is EOF.
func (p *parser) jump(pos int) int {
  if pos < 0 {
    panic("can't jump to negative position")
  }

  p.pos = pos
  if p.pos < len(p.src) {
    p.ch, _ = utf8.DecodeRuneInString(p.src[p.pos:])
  } else {
    p.ch = -1
  }
  return p.ch
}

/**
 * Return the literal string from->to some expected characters. Assumes that the
 * cursor is resting on the from character. Will return the parser at the first
 * char past the result.
 */
func (p *parser) literal(start int, end int) (result string, err bool) {
  if p.ch != start {
    return "", true
  }

  result = ""
  for p.nextc() != end {
    if p.ch == -1 {
      return result, true
    }
    result += fmt.Sprintf("%c", p.ch)
  }
  return result, false
}

/**
 * Connect from -> to.
 */
func (p *parser) out(from *instr, to *instr) {
  if from.out == nil {
    from.out = to
  } else if from.mode == kSplit && from.out1 == nil {
    from.out1 = to
  } else {
    panic("can't out")
  }
}

/**
 * Consume some bracketed expression.
 */
func (p *parser) alt() (start *instr, end *instr) {
  use_alts := true
  var alt_id *string
  altpos := p.re.alts
  p.re.alts += 1

  if p.ch != '(' {
    panic("alt must start with '('")
  }

  end = p.instr() // shared end state for alt
  end.mode = kAltEnd
  end.alt = altpos

  p.nextc()
  if p.ch == '?' {
    // TODO: it might be appropriate to move this whole logic outside of alt().
    p.nextc()
    if p.ch == 'P' {
      p.nextc()
      s, err := p.literal('<', '>')
      if err {
        panic("couldn't consume name in < >")
      }
      alt_id = &s
      p.nextc() // move past '>'
    } else {
      // anything but 'P' means flags (unmatched).
      use_alts = false
      outer: for {
        switch p.ch {
        case ':':
          p.nextc() // move past ':'
          break outer // no more flags, process re
        case ')':
          panic("can't yet apply flags to outer")
          break outer // no more flags, ignore re, apply flags to outer
        default:
          panic(fmt.Sprint("flag unsupported:", p.ch))
        }
        p.nextc()
      }
    }
  }

  b_start, b_end := p.regexp()
  start = b_start
  p.out(b_end, end)

  for p.ch == '|' {
    start = p.instr()
    p.out(start, b_start)

    p.nextc()
    b_start, b_end = p.regexp()
    p.out(start, b_start)
    p.out(b_end, end)
    b_start = start
  }

  if p.ch != ')' {
    panic("alt must end with ')'")
  }

  alt_begin := p.instr()
  alt_begin.mode = kAltBegin
  alt_begin.alt = altpos
  p.out(alt_begin, start)

  if !use_alts {
    // clear alts, this is an unmatched group
    alt_begin.mode = kSplit
    end.mode = kSplit
  } else if alt_id != nil {
    // set alt string id
    alt_begin.alt_id = alt_id
    end.alt_id = alt_id
  }

  return alt_begin, end
}

/**
 * Consume a character class, and return a single instr representing this class.
 */
func (p *parser) charclass() runeclass {
  if p.ch != '[' {
    panic("expect charclass to start with [")
  }
  p.nextc() // walk over '['

  class := NewComplexRuneClass()
  negate := false
  if p.ch == '^' {
    negate = !negate
    p.nextc()
  }

  outer: for {
    switch p.ch {
    case '[':
      if p.nextc() != ':' {
        panic("expected ascii class [:")
      }
      _, err := p.literal(':', ':')
      if err {
        panic("could not consume ascii class")
      }
      if p.nextc() != ']' {
        panic("unclosed ascii class")
      }
    case ']':
      // NOTE: caller expects us not to walk over last bracket.
      break outer
    case -1:
      panic("unclosed character class")
    }

    rune := p.ch
    r := make([]unicode.Range, 1)
    if p.nextc() == '-' {
      rune_end := p.nextc()
      if rune >= rune_end {
        panic(fmt.Sprintf("unexpected range: %c >= %c", rune, rune_end))
      }
      r[0] = unicode.Range{rune, rune_end, 1}
      p.nextc() // walk past end of range
    } else {
      r[0] = unicode.Range{rune, rune, 1}
    }
    if negate {
      class.Exclude(r)
    } else {
      class.Include(r)
    }
  }

  return class
}

/**
 * Consume a single term (note that term may include a bracketed expression).
 */
func (p *parser) term() (start *instr, end *instr) {
  start = p.instr()
  end = start

  switch p.ch {
  case -1:
    panic("EOF in term")
  case '*', '+', '{', '?':
    panic("unexpected expansion char")
  case ')', '}', ']':
    panic("unexpected close element")
  case '(':
    start, end = p.alt()
  case '[':
    start.mode = kRuneClass
    start.rune = p.charclass()
  case '$':
    panic("not yet supported: end of string")
  case '^':
    panic("not yet supported: start of string")
  case '.':
    start.mode = kRuneClass
    start.rune = NewAnyRuneClass()
  default:
    ch := p.ch
    if ch == '\\' {
      switch p.nextc() {
      case 'n':
        ch = '\n'
      case 't':
        ch = '\t'
      default:
      panic("only expected \\n or \\t")
      }
    }

    start.mode = kRuneClass
    start.rune = NewSingleRuneClass(ch)
  }
  p.nextc()
  return start, end
}

/**
 * Consume a closure: i.e. ( term + [ repitition ] )
 */
func (p *parser) closure() (start *instr, end *instr) {
  term_pos := p.pos
  start = p.instr()
  end = start
  t_start, t_end := p.term()
  first := true

  var req int
  var opt int
  greedy := true
  switch p.ch {
  case '?':
    req, opt = 0, 1
  case '*':
    req, opt = 0, -1
  case '+':
    req, opt = 1, -1
  case '{':
    r, err := p.literal('{', '}')
    if err {
      panic("couldn't consume repitition {}")
    }
    parts := strings.Split(r, ",", 2)
    // TODO: handle malformed int
    req, _ = strconv.Atoi(parts[0])
    if len(parts) == 2 {
      if len(parts[1]) > 0 {
        // TODO: handle malformed int
        opt, _ = strconv.Atoi(parts[1])
      } else {
        opt = -1
      }
    }
  default:
    return t_start, t_end // nothing to see here
  }

  if p.nextc() == '?' {
    // TODO: explode if opt is 0?
    greedy = false
    p.nextc()
  }
  end_pos := p.pos

  if req < 0 || opt < -1 || req == 0 && opt == 0 {
    panic("invalid req/opt combination")
  }

  // generate required steps
  for i := 0; i < req; i++ {
    if first {
      first = false
    } else {
      p.jump(term_pos)
      t_start, t_end = p.term()
    }
    p.out(end, t_start)
    end = t_end
  }

  // generate optional steps
  if opt == -1 {
    if first {
      first = false
    } else {
      p.jump(term_pos)
      t_start, t_end = p.term()
    }

    helper := p.instr()
    p.out(end, helper)
    if greedy {
      helper.out = t_start // greedily choose optional step
    } else {
      helper.out1 = t_start // optional step is 2nd preference
    }
    p.out(t_end, helper)
    end = helper
  } else {
    real_end := p.instr()

    for i := 0; i < opt; i++ {
      if first {
        first = false
      } else {
        p.jump(term_pos)
        t_start, t_end = p.term()
      }

      helper := p.instr()
      p.out(end, helper)
      if greedy {
        helper.out = t_start
      } else {
        helper.out1 = t_start
      }
      p.out(helper, real_end)

      end = p.instr()
      p.out(t_end, end)
    }

    p.out(end, real_end)
    end = real_end
  }

  p.jump(end_pos)
  return start, end
}

/**
 * Match a regexp (defined as [closure]*) from parser, until either: EOF, |, or
 * ) is encountered.
 */
func (p *parser) regexp() (start *instr, end *instr) {
  start = p.instr()
  curr := start

  for {
    if p.ch == -1 || p.ch == '|' || p.ch == ')' {
      break
    }
    s, e := p.closure()
    p.out(curr, s)
    curr = e
  }

  end = p.instr()
  p.out(curr, end)
  return start, end
}

/**
 * Cleanup the given program. Assumes the given input is a flat slice containing
 * no nil instructions. Will not clean up the first instruction, as it is always
 * the canonical entry point for the regexp.
 *
 * Returns a similarly flat slice containing no nil instructions, however the
 * slice may potentially be smaller.
 */
func cleanup(prog []*instr) []*instr {
  // Detect kSplit recursion. We can remove this and convert it to a single path.
  states := NewStateSet(len(prog), len(prog))
  for i := 1; i < len(prog); i++ {
    states.Clear()
    pi := prog[i]
    var fn func(ci *instr) bool
    fn = func(ci *instr) bool {
      if ci != nil && ci.mode == kSplit {
        if states.Put(ci.idx) {
          // NOTE: I'm not sure if this will ever happen. Panic for now. If we're
          // confident this won't happen, we could move the panic to runtime.
          panic("regexp should never loop")
          return true
        }
        if fn(ci.out) {
          ci.out = nil
        }
        if fn(ci.out1) {
          ci.out1 = nil
        }
        return false
      }
      return false
    }
    fn(pi)
  }

  // Iterate through the program, and remove single-instr kSplits.
  // NB: Don't parse the first instr, it will always be single.
  for i := 1; i < len(prog); i++ {
    pi := prog[i]
    if pi.mode == kSplit && (pi.out1 == nil || pi.out == pi.out1) {
      for j := 0; j < len(prog); j++ {
        if prog[j] == nil {
          continue
        }
        pj := prog[j]
        if pj.out == pi {
          pj.out = pi.out
        }
        if pj.out1 == pi {
          pj.out1 = pi.out
        }
      }
      prog[i] = nil
    }
  }

  // We may now have nil gaps: shift everything up.
  last := 0
  for i := 0; i < len(prog); i++ {
    if prog[i] != nil {
      last = i
    } else {
      // find next non-nil, move here
      var found int
      for found = i; found < len(prog); found++ {
        if prog[found] != nil {
          break
        }
      }
      if found == len(prog) {
        break // no more entries
      }

      // move found to i
      prog[i] = prog[found]
      prog[i].idx = i
      prog[found] = nil
      last = i
    }
  }

  return prog[0:last+1]
}

/**
 * Generates a simple straight-forward NFA.
 */
func Parse(src string) (r *sregexp) {
  // possibly expand this RE all the way to the left
  if src[0] == '^' {
    src = "(" + src[1:len(src)]
  } else {
    src = ".*?(" + src
  }

  // possibly expand this RE to the right
  if src[len(src)-1] == '$' {
    src = src[0:len(src)-1] + ")"
  } else {
    src = src + ").*?"
  }

  re := &sregexp{make([]*instr, 0, 1), 0}
  p := parser{re, src, -1, -1}
  begin := p.instr()
  match := p.instr()
  match.mode = kMatch

  p.nextc()
  start, end := p.regexp()

  if p.ch != -1 {
    panic("could not consume all of regexp!")
  }

  p.out(begin, start)
  p.out(end, match)

  re.prog = cleanup(re.prog)
  return re
}
