
package main

import (
  "fmt"
  "utf8"
)

const (
  kSplit = iota         // proceed down out & out1
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
}

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
}

/**
 * Generate a new pre-indexed instr.
 */
func (p *parser) instr() *instr {
  if p.inst == len(p.prog) {
    panic("overflow instr buffer")
  }
  i := &instr{p.inst, kSplit, nil, nil, -1, nil}
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
    panic("not yet supported: (")
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
    panic("unsupported: {")
  }
  return start, end
}

func (p *parser) regexp() (start *instr, end *instr) {
  // supports [closure]*
  start = p.instr()
  curr := start

  for p.ch != -1 {
    s, e := p.closure()
    p.out(curr, s)
    curr = e
  }

  end = p.instr()
  p.out(curr, end)
  return start, end
}

/**
 * Generates a simple straight-forward NFA.
 */
func Parse(src string) (prog []*instr) {
  p := parser{src, -1, 0, make([]*instr, 128), 0}
  begin := p.instr()
  match := p.instr()
  match.mode = kMatch

  p.nextc()
  start, end := p.regexp()

  p.out(begin, start)
  p.out(end, match)

  return p.prog
}
