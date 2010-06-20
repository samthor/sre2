
package main

import (
  "unicode"
)

func IsAsciiLower(rune int) bool {
  return rune >= 'a' && rune <= 'z'
}

func IsAsciiUpper(rune int) bool {
  return rune >= 'A' && rune <= 'Z'
}

func IsAsciiDigit(rune int) bool {
  return rune >= '0' && rune <= '9'
}

func IsAsciiAlpha(rune int) bool {
  return IsAsciiLower(rune) || IsAsciiUpper(rune)
}

func IsAsciiAlphaNum(rune int) bool {
  return IsAsciiAlpha(rune) || IsAsciiDigit(rune)
}

func IsAsciiSpace(rune int) bool {
  for _, ch := range "\t\n\v\f\r " {
    if rune == ch {
      return true
    }
  }
  return false
}

var (
  ASCII = map[string] func(rune int) bool {
    "alnum": IsAsciiAlphaNum,
    "alpha": IsAsciiAlpha,
//    "ascii"
    "blank": func(rune int) bool { return rune == '\t' || rune == ' ' },
//    "cntrl"
    "digit": func(rune int) bool { return rune >= '0' && rune <= '9' },
//    "graph"
    "lower": IsAsciiLower,
//    "print"
//    "punct"
    "space": IsAsciiSpace,
    "upper": IsAsciiUpper,
    "word": func(rune int) bool { return IsAsciiAlphaNum(rune) || rune == '_' },
    "xdigit": func(rune int) bool { return unicode.Is(unicode.ASCII_Hex_Digit, rune) },
  }
)
