
package sre2

import (
  "fmt"
  "os"
  "strconv"
  "strings"
  "unicode"
  "utf8"
)

// Regexp definition. Simple, just a list of states and a number of alts.
type sregexp struct {
  prog []*instr         // list of states
  alts int              // number of marked alts [()'s] in this regexp
}

// Writes the given regexp to Stderr, for debugging.
func (r *sregexp) DebugOut() {
  for i := 0; i < len(r.prog); i++ {
    fmt.Fprintln(os.Stderr, i, r.prog[i].String())
  }
}

// Instruction type definitions, for "instr.mode".
const (
  kSplit = iota         // proceed down out & out1
  kAltBegin             // begin of alt section, i.e. '('
  kAltEnd               // end of alt section, i.e. ')'
  kRuneClass            // if match rune, proceed down out
  kMatch                // success state!
)

// Escape constants and their mapping to actual Unicode runes.
var (
  ESCAPES = map[int]int {
    'a': 7, 't': 9, 'n': 10, 'v': 11, 'f': 12, 'r': 13,
  }
)

// Represents a single instruction in any regexp.
type instr struct {
  idx int               // index of this instr
  mode byte             // mode (as above)
  out *instr            // next instr to process
  out1 *instr           // alt next instr (for kSplit)
  rune *RuneClass       // rune class
  alt int               // identifier of alt branch (for kAlt{Begin,End})
  alt_id *string        // string identifier of alt branch
}

// This provides a string-representation of any given instruction.
func (i *instr) String() string {
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

// Matcher method for consuming runes, thus only matches kRuneClass.
func (s *instr) match(rune int) bool {
  return s.mode == kRuneClass && s.rune.MatchRune(rune)
}

// Transient parser state, a combination of regexp and string iterator.
type parser struct {
  re *sregexp
  src string
  ch int
  pos int
}

// Generate a new instruction struct for use in regexp. By default, the instr
// will be of type 'kSplit'.
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

// Return the literal string from->to some expected characters. Assumes that the
// cursor is resting on the 'from' character. This method will return with the
// cursor resting on the 'to' character, or on the final EOF (if not found).
func (p *parser) literal(start int, end int) string {
  if p.ch != start {
    panic(fmt.Sprintf("expected literal to start with: %c", start))
  }

  result := ""
  for p.nextc() != end {
    if p.ch == -1 {
      panic(fmt.Sprintf("expected literal to end with: %c", end))
    }
    result += fmt.Sprintf("%c", p.ch)
  }
  return result
}

// Helper method to connect instr 'from' to instr 'out'.
// TODO: Use safer connection helpers.
func (p *parser) out(from *instr, to *instr) {
  if from.out == nil {
    from.out = to
  } else if from.mode == kSplit && from.out1 == nil {
    from.out1 = to
  } else {
    panic("can't out")
  }
}

// Consume some bracketed expression. At input, expects the cursor to be on '('
// and will return with the cursor just past the matching ')'.
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
      s := p.literal('<', '>')
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
  p.nextc() // Step over the final bracket.

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

// Consume a single character class and provide an implementation of the
// runeclass interface. Will consume all characters that are part of definition.
func (p *parser) class(within_class bool) (class *RuneClass) {
  class = NewRuneClass()
  switch p.ch {
  case '.':
    // TODO: only match newline in 's' mode
    class.AddFunc(false, func(rune int) bool { return true })
    p.nextc()
  case '[':
    if p.nextc() == ':' {
      // Match an ASCII class name.
      ascii_class := p.literal(':', ':')
      negate := false
      if ascii_class[0] == '^' {
        negate = true
        ascii_class = ascii_class[1:len(ascii_class)]
      }
      if p.nextc() != ']' {
        panic("expected closing of ascii class with ':]', got ':'")
      }

      if ok := class.AddAsciiClass(negate, ascii_class); !ok {
        panic(fmt.Sprintf("could not identify ascii class: %s", ascii_class))
      }
    } else {
      if within_class {
        panic("can't match a [...] class within another class")
      }

      negate := false
      if p.ch == '^' {
        negate = true
        p.nextc()
      }

      for p.ch != ']' {
        class.AddRuneClass(negate, p.class(true))
      }
    }
    p.nextc() // Move over final ']'.
  case '\\':
    // Match some escaped character or escaped combination.
    switch p.nextc() {
    case 'x':
      // Match hex character code.
      var hex string
      if p.nextc() == '{' {
        hex = p.literal('{', '}')
      } else {
        hex = fmt.Sprintf("%c%c", p.ch, p.nextc())
      }
      p.nextc() // Step over the end of the hex code.

      // Parse and return the corresponding rune.
      rune, err := strconv.Btoui64(hex, 16)
      if err != nil {
        panic(fmt.Sprintf("couldn't parse hex: %s", hex))
      }

      class.AddRune(false, int(rune))
    case 'p', 'P':
      // Match a Unicode class name.
      negate := (p.ch == 'P')
      unicode_class := fmt.Sprintf("%c", p.nextc())
      if unicode_class[0] == '{' {
        unicode_class = p.literal('{', '}')
      }
      p.nextc() // Step over the end of the hex code.

      // Find and return the class.
      if ok := class.AddUnicodeClass(negate, unicode_class); !ok {
        panic(fmt.Sprintf("could not identify unicode class: %s", unicode_class))
      }
    case 'd', 'D':
      // Match digits.
      p.nextc()
      negate := (p.ch == 'D')
      class.AddUnicodeClass(negate, "Nd")
    case 's', 'S':
      // Match whitespace.
      p.nextc()
      negate := (p.ch == 'S')
      class.AddAsciiClass(negate, "whitespace")
    case 'w', 'W':
      // Match word characters.
      p.nextc()
      negate := (p.ch == 'W')
      class.AddAsciiClass(negate, "word")
    default:
      if escape := ESCAPES[p.ch]; escape != 0 {
        // Literally match '\n', '\r', etc.
        class.AddRune(false, escape)
        p.nextc()
      } else if unicode.Is(_punct, p.ch) {
        // Allow punctuation to be blindly escaped.
        class.AddRune(false, p.ch)
        p.nextc()
      } else if unicode.IsDigit(p.ch) {
        // Match octal character code (begins with digit, up to three digits).
        oct := ""
        for i := 0; i < 3; i++ {
          oct += fmt.Sprintf("%c", p.ch)
          if !unicode.IsDigit(p.nextc()) {
            break
          }
        }

        // Parse and return the corresponding rune.
        rune, err := strconv.Btoui64(oct, 8)
        if err != nil {
          panic(fmt.Sprintf("couldn't parse oct: %s", oct))
        }
        class.AddRune(false, int(rune))
      } else {
        panic(fmt.Sprintf("unsupported escape type: \\%c", p.ch))
      }
    }
  default:
    // Match a single rune literal, or a range (when inside a character class).
    // TODO: Sanity-check and disallow some punctuation.
    rune := p.ch
    if p.nextc() == '-' {
      if !within_class {
        panic(fmt.Sprintf("can't match a range outside class: %c-%c", rune, p.nextc()))
      }
      rune_high := p.nextc()
      if rune_high < rune {
        panic(fmt.Sprintf("unexpected range: %c >= %c", rune, rune_high))
      }
      p.nextc() // Step over the end of the range.
      class.AddRuneRange(false, rune, rune_high)
    } else {
      class.AddRune(false, rune)
    }
  }

  return class
}

// Consume a single term at the current cursor position. This may include a
// bracketed expression. When this function returns, the cursor will have moved
// past the final rune in this term.
func (p *parser) term() (start *instr, end *instr) {
  switch p.ch {
  case -1:
    panic("EOF in term")
  case '*', '+', '{', '?':
    panic(fmt.Sprintf("unexpected expansion char: %c", p.ch))
  case ')', '}', ']':
    panic("unexpected close element")
  case '(':
    return p.alt()
  case '$':
    panic("not yet supported: end of string")
  case '^':
    panic("not yet supported: start of string")
  }

  if p.ch == '\\' {
    pos := p.pos
    switch p.nextc() {
    case 'Q':
      // Match a string literal up to '\E'.
      p.nextc()
      i := strings.Index(p.src[p.pos:len(p.src)], "\\E")
      if i == -1 {
        panic("couldn't find \\E after \\Q")
      }
      literal := p.src[p.pos:p.pos+i]
      p.jump(p.pos + i + 2)

      start = p.instr()
      end = start
      for _, ch := range literal {
        instr := p.instr()
        instr.mode = kRuneClass
        instr.rune = NewRuneClass()
        instr.rune.AddRune(false, ch)

        p.out(end, instr)
        end = instr
      }
      return start, end
    }

    // We didn't find anything interesting: push back to the previous position.
    p.jump(pos)
  }

  // Try to consume a rune class.
  start = p.instr()
  start.mode = kRuneClass
  start.rune = p.class(false)
  return start, start
}

// Safely retrieve a given term from the given position and alt count. If the
// passed first is true, then set it to false and perform a no-op. Otherwise,
// retrieve the new term.
func (p *parser) safe_term(pos int, alt int, first *bool, start **instr, end **instr) {
  if *first {
    *first = false
    return
  }
  p.jump(pos)
  p.re.alts = alt
  *start, *end = p.term()
}

// Consume a closure, defined as (term[repitition]). When this function returns,
// the cursor will be resting past the final rune in this closure.
func (p *parser) closure() (start *instr, end *instr) {

  // Store state of pos/alts in case we have to reparse term.
  base_pos, base_alt := p.pos, p.re.alts

  // Grab first term.
  start = p.instr()
  end = start
  t_start, t_end := p.term()
  first := true // While true, we have a pending term.

  // Req and opt represent the number of required cases, and the number of
  // optional cases, respectively. Opt may be -1 to indicate no optional limit.
  var req int
  var opt int
  greedy := true // Greedily choose an optional step over continuing.
  switch p.ch {
  case '?':
    req, opt = 0, 1
  case '*':
    req, opt = 0, -1
  case '+':
    req, opt = 1, -1
  case '{':
    raw := p.literal('{', '}')
    parts := strings.Split(raw, ",", 2)
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
    greedy = !greedy
    p.nextc()
  }
  end_pos := p.pos

  if req < 0 || opt < -1 || req == 0 && opt == 0 {
    panic("invalid req/opt combination")
  }

  // Generate all required steps.
  for i := 0; i < req; i++ {
    p.safe_term(base_pos, base_alt, &first, &t_start, &t_end)

    p.out(end, t_start)
    end = t_end
  }

  // Generate all optional steps.
  if opt == -1 {
    helper := p.instr()
    p.out(end, helper)
    if greedy {
      helper.out = t_start // greedily choose optional step
    } else {
      helper.out1 = t_start // optional step is 2nd preference
    }
    if end != t_end {
      // This is a little kludgy, but basically only wires up the term to the
      // helper iff it hasn't already been done.
      p.out(t_end, helper)
    }
    end = helper
  } else {
    real_end := p.instr()

    for i := 0; i < opt; i++ {
      p.safe_term(base_pos, base_alt, &first, &t_start, &t_end)

      helper := p.instr()
      p.out(end, helper)
      if greedy {
        helper.out = t_start // greedily choose optional step
      } else {
        helper.out1 = t_start // optional step is 2nd preference
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

// Match a regexp (defined as ([closure]*)) from the parser until either: EOF,
// the literal '|' or the literal ')'. At return, the cursor will still rest
// on this final terminal character.
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

// Cleanup the given program. Assumes the given input is a flat slice containing
// no nil instructions. Will not clean up the first instruction, as it is always
// the canonical entry point for the regexp.
// Returns a similarly flat slice containing no nil instructions, however the
// slice may potentially be smaller.
func cleanup(prog []*instr) []*instr {
  // Detect kSplit recursion. We can remove this and convert it to a single path.
  // This might happen in cases where we loop over some instructions which are
  // not matchers, e.g. \Q\E*.
  states := NewStateSet(len(prog), len(prog))
  for i := 1; i < len(prog); i++ {
    states.Clear()
    pi := prog[i]
    var fn func(ci *instr) bool
    fn = func(ci *instr) bool {
      if ci != nil && ci.mode == kSplit {
        if states.Put(ci.idx) {
          // We've found a recursion.
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

// Generates a simple, straight-forward NFA. Matches an entire regexp from the
// given input string.
func Parse(src string) (r *sregexp) {
  // possibly expand this RE all the way to the left
  if len(src) > 0 && src[0] == '^' {
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
