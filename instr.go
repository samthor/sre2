
package main

import (
  "fmt"
  "strconv"
  "unicode"
  "utf8"
)

const (
  kSplit = iota         // proceed down out & out1
  kAltBegin             // begin of alt section, i.e. '('
  kAltEnd               // end of alt section, i.e. ')'
  kRune                 // if match rune, proceed down out
  kCall                 // if matcher passes, proceed down out
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
  rune int              // rune to match (kRune)
  matcher func(rune int) bool   // matcher method (for kCall)
  alt int               // identifier of alt branch (for kAlt{Begin,End})
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
  case kAltBegin:
    str += fmt.Sprintf(" kAltBegin alt=%d", i.alt)
  case kAltEnd:
    str += fmt.Sprintf(" kAltEnd alt=%d", i.alt)
  case kRune:
    str += fmt.Sprintf(" kRune rune=%c", i.rune)
  case kCall:
    str += " kCall meth=?"
  case kMatch:
    str += " kMatch"
  }
  return str + out + "}"
}

type parser struct {
  src string
  ch int
  pos int
  prog []*instr
  inst int
  altpos int
}

/*
 * Generic matcher for consuming instr instances (i.e. kRune/kCall). Does not
 * match anything else.
 */
func (s *instr) match(rune int) bool {
  if s.mode == kRune {
    return s.rune == rune || s.rune == -1
  } else if s.mode == kCall {
    return s.matcher(rune)
  }
  return false
}

/**
 * Generate a new pre-indexed instr.
 */
func (p *parser) instr() *instr {
  if p.inst == len(p.prog) {
    panic("overflow instr buffer")
  }
  i := &instr{p.inst, kSplit, nil, nil, -1, nil, -1}
  p.prog[p.inst] = i
  p.inst += 1
  return i
}

/**
 * Store/return the next character in parser. -1 indicates EOF.
 */
func (p *parser) nextc() int {
  if p.pos >= len(p.src) {
    p.ch = -1
  } else {
		c, w := utf8.DecodeRuneInString(p.src[p.pos:])
		p.ch = c
		p.pos += w
  }
  return p.ch
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

func (p *parser) alt() (start *instr, end *instr) {
  altpos := p.altpos
  p.altpos += 1

  if p.ch != '(' {
    panic("alt must start with '('")
  }

  end = p.instr() // shared end state for alt
  end.mode = kAltEnd
  end.alt = altpos

  p.nextc()
  b_start, b_end := p.regexp()
  start = b_start
  p.out(b_end, end)

  for p.ch == '|' {
    split := p.instr()
    p.out(split, b_start)

    p.nextc()
    b_start, b_end = p.regexp()
    p.out(split, b_start)
    p.out(b_end, end)

    start = split
  }

  if p.ch != ')' {
    panic("alt must end with ')'")
  }

  alt_begin := p.instr()
  alt_begin.mode = kAltBegin
  alt_begin.alt = altpos
  p.out(alt_begin, start)
  return alt_begin, end
}

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
    panic("not yet supported: [")
  case '$':
    panic("not yet supported: end of string")
  case '^':
    panic("not yet supported: start of string")
  case '.':
    start.mode = kRune
  case '\\':
    next := p.nextc()
    start.mode = kRune
    switch next {
    case 'n':
      start.rune = '\n'
    case 't':
      start.rune = '\t'
    default:
      // TODO: limit this to punctuation
      start.rune = next
    }
  default:
    start.mode = kRune
    start.rune = p.ch
  }
  p.nextc()
  return start, end
}

func (p *parser) closure() (start *instr, end *instr) {
  start, end = p.term()

  switch p.ch {
  case '?', '*':
    p_start := start
    p_end := end
    start = p.instr()
    end = p.instr()
    p.out(start, p_start)
    p.out(start, end)
    p.out(p_end, end)
    if p.ch == '*' {
      p.out(end, start)
    }
    p.nextc()
  case '+':
    p_end := end
    end = p.instr()
    p.out(p_end, end)
    p.out(end, start)
    p.nextc()
  case '{':
    count_str := ""
    p.nextc()
    for unicode.IsDigit(p.ch) {
      count_str += fmt.Sprintf("%c", p.ch)
      p.nextc()
    }
    if len(count_str) == 0 {
      panic("{ must be followed by digit")
    }
    if p.ch == '}' {
      // fixed expansion
      count, _ := strconv.Atoi(count_str)
      panic(fmt.Sprintf("can't yet expand to: %d", count))
    } else if p.ch == ',' {
      panic("can't handle anything but {n}")
    } else {
      panic("unexpected char in {}")
    }
  }
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
  // TODO: Clear kSplit recursion. In some cases, kSplit paths may recurse back
  // on themselves. We can remove this and convert it to a single-instr kSplit.

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
  p := parser{src, -1, 0, make([]*instr, 128), 0, 0}
  begin := p.instr()
  match := p.instr()
  match.mode = kMatch

  p.nextc()
  start, end := p.regexp()

  if p.nextc() != -1 {
    panic("could not consume all of regexp!")
  }

  p.out(begin, start)
  p.out(end, match)

  result := p.prog[0:end.idx+1]
  result = cleanup(result)

  return &sregexp{result, p.altpos}
}
