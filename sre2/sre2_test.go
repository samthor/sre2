package sre2

import (
	"fmt"
	"testing"
)

// Check the given state to be true.
func checkState(t *testing.T, state bool, err string) {
	if !state {
		t.Error(err)
	}
}

// Check the equality of two []int slices.
func checkIntSlice(t *testing.T, expected []int, result []int, err string) {
	match := true
	if (expected == nil || result == nil) && (expected != nil && result != nil) {
		match = false
	} else if len(expected) != len(result) {
		match = false
	} else {
		for i := 0; i < len(expected); i++ {
			if expected[i] != result[i] {
				match = false
			}
		}
	}
	checkState(t, match, fmt.Sprintf("%s: got %s, expected %s", err, result, expected))
}

// Run a selection of basic regular expressions against this package.
func TestSimpleRe(t *testing.T) {
	r := MustParse("")
	checkState(t, r.NumAlts() == 0, "blank re should have no alts")
	checkState(t, r.Match(""), "everything should match")
	checkState(t, r.Match("fadsnjkflsdafnas"), "everything should match")

	r = MustParse("^(a|b)+c*$")
	checkState(t, r.NumAlts() == 1, "simple re should have single alt")
	checkState(t, !r.Match("abd"), "not a valid match")
	checkState(t, r.Match("a"), "basic string should match")
	checkState(t, !r.Match(""), "empty string should not match")
	checkState(t, r.Match("abcccc"), "longer string should match")

	r = MustParse("(\\w*)\\s*(\\w*)")
	res := r.MatchIndex("zing hello there")
	checkIntSlice(t, []int{0, 10, 0, 4, 5, 10}, res, "did not match first two words as expected")

	r = MustParse(".*?(\\w+)$")
	res = r.MatchIndex("zing hello there")
	checkIntSlice(t, []int{0, 16, 11, 16}, res, "did not match last word as expected")

	res = r.MatchIndex("\n")
	checkIntSlice(t, res, nil, "should return nil on failed match")
}

// Test parsing an invalid RE returns an error.
func TestInvalidRe(t *testing.T) {
	r, err := Parse("a**")
	checkState(t, err != nil, "must fail parsing")
	checkState(t, r == nil, "regexp must be nil")

	pass := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				pass = true
			}
		}()
		MustParse("z(((a")
	}()
	checkState(t, pass, "should panic")
}

// Test behaviour related to character classes expressed within [...].
func TestCharClass(t *testing.T) {
	r := MustParse("^[\t[:word:]]+$") // Match tabs and word characters.
	checkState(t, r.Match("c"), "non-space should match")
	checkState(t, !r.Match("c t"), "space should not match")
	checkState(t, r.Match("c\tt"), "tab should match")

	r = MustParse("^[:ascii:]*$")
	checkState(t, r.Match(""), "nothing should match")
	checkState(t, r.Match("c"), "ascii should match")
	checkState(t, !r.Match("Π"), "unicode should not match")

	r = MustParse("^\\pN$")
	checkState(t, r.Match("〩"), "character from Nl should match")
	checkState(t, r.Match("¾"), "character from Nu should match")

	r = MustParse("^\\p{Nl}$")
	checkState(t, r.Match("〩"), "character from Nl should match")
	checkState(t, !r.Match("¾"), "character from Nu should not match")

	r = MustParse("^[^. ]$")
	checkState(t, r.Match("\n"), "not everything should match \\n")
	checkState(t, !r.Match(" "), "should match only \\n")

	r = MustParse("^[.\n]$")
	checkState(t, r.Match("\n"), "should match \\n")

	r = MustParse("^\\W$")
	checkState(t, !r.Match("a"), "should not match word char")
	checkState(t, r.Match("!"), "should match non-word")

	r = MustParse("^[abc\\W]$")
	checkState(t, r.Match("a"), "should match 'a'")
	checkState(t, r.Match("!"), "should match '!'")
	checkState(t, !r.Match("d"), "should not match 'd'")

	r = MustParse("^[^abc\\W]$")
	checkState(t, !r.Match("a"), "should not match 'a'")
	checkState(t, !r.Match("%"), "should not match non-word char")
	checkState(t, r.Match("d"), "should match 'd'")

	r = MustParse("^[\\w\\D]$")
	checkState(t, r.Match("a"), "should match regular char 'a'")
	checkState(t, r.Match("2"), "should still match number '2', caught by \\w")

	r = MustParse("^[\\[-\\]]$")
	checkState(t, r.Match("]"), "should match ']'")
	checkState(t, r.Match("["), "should match '['")
	checkState(t, r.Match("\\"), "should match '\\', between [ and ]")
}

// Test regexp generated by escape sequences (e.g. \n, \. etc).
func TestEscapeSequences(t *testing.T) {
	r := MustParse("^\\.\n\\044$") // Match '.\n$'
	checkState(t, r.Match(".\n$"), "should match")
	checkState(t, !r.Match(" \n$"), "space should not match")
	checkState(t, !r.Match("\n\n$"), ". does not match \n by default")
	checkState(t, !r.Match(".\n"), "# should not be treated as end char")

	r = MustParse("^\\x{03a0}\\x25$") // Match 'Π%'.
	checkState(t, r.Match("Π%"), "should match pi+percent")

	r, err := Parse("^\\Π$")
	checkState(t, err != nil && r == nil,
		"should have failed on trying to escape Π, not punctuation")
}

// Tests string literals between \Q...\E.
func TestStringLiteral(t *testing.T) {
	r := MustParse("^\\Qhello\\E$")
	checkState(t, r.Match("hello"), "should match hello")

	r = MustParse("^\\Q.$\\\\E$") // match ".$\\"
	checkState(t, r.Match(".$\\"), "should match")
	checkState(t, !r.Match(" $\\"), "should not match")

	r = MustParse("^a\\Q\\E*b$") // match absolutely nothing between 'ab'
	checkState(t, r.Match("ab"), "should match")
	checkState(t, !r.Match("acb"), "should not match")
}

// Test closure expansion types, such as {..}, ?, +, * etc.
func TestClosureExpansion(t *testing.T) {
	r := MustParse("^za?$")
	checkState(t, r.Match("z"), "should match none")
	checkState(t, r.Match("za"), "should match single")
	checkState(t, !r.Match("zaa"), "should not match more")

	r = MustParse("^a{2,2}$")
	checkState(t, !r.Match(""), "0 should fail")
	checkState(t, !r.Match("a"), "1 should fail")
	checkState(t, r.Match("aa"), "2 should succeed")
	checkState(t, r.Match("aaa"), "3 should succeed")
	checkState(t, r.Match("aaaa"), "4 should succeed")
	checkState(t, !r.Match("aaaaa"), "5 should fail")

	r = MustParse("^a{2}$")
	checkState(t, !r.Match(""), "0 should fail")
	checkState(t, !r.Match("a"), "1 should fail")
	checkState(t, r.Match("aa"), "2 should succeed")
	checkState(t, !r.Match("aaa"), "3 should fail")

	r = MustParse("^a{3,}$")
	checkState(t, !r.Match("aa"), "2 should fail")
	checkState(t, r.Match("aaa"), "3 should succeed")
	checkState(t, r.Match("aaaaaa"), "more should succeed")
}

// Test specific greedy/non-greedy closure types.
func TestClosureGreedy(t *testing.T) {
	r := MustParse("^(a{0,2}?)(a*)$")
	res := r.MatchIndex("aaa")
	checkIntSlice(t, []int{0, 3, 0, 0, 0, 3}, res, "did not match expected")

	r = MustParse("^(a{0,2})?(a*)$")
	res = r.MatchIndex("aaa")
	checkIntSlice(t, []int{0, 3, 0, 2, 2, 3}, res, "did not match expected")

	r = MustParse("^(a{2,}?)(a*)$")
	res = r.MatchIndex("aaa")
	checkIntSlice(t, []int{0, 3, 0, 2, 2, 3}, res, "did not match expected")
}

// Test simple left/right matchers.
func TestLeftRight(t *testing.T) {
	r := MustParse("^.\\b.$")
	checkState(t, r.Match("a "), "left char is word")
	checkState(t, r.Match(" a"), "right char is word")
	checkState(t, !r.Match("  "), "not a boundary")
	checkState(t, !r.Match("aa"), "not a boundary")
}

// Test general flags in sre2.
func TestFlags(t *testing.T) {
	r := MustParse("^(?i:AbC)zz$")
	checkState(t, r.Match("abczz"), "success")
	checkState(t, !r.Match("abcZZ"), "fail, flag should not escape")
	res := r.MatchIndex("ABCzz")
	checkIntSlice(t, []int{0, 5}, res, "should just have a single outer paren")

	r = MustParse("^(?U)(a+)(.+)$")
	res = r.MatchIndex("aaaabb")
	checkIntSlice(t, []int{0, 6, 0, 1, 1, 6}, res, "should be ungreedy")

	r = MustParse("^(?i)a*(?-i)b*$")
	checkState(t, r.Match("AAaaAAaabbbbb"), "success")
	checkState(t, !r.Match("AAaaAAaaBBBa"), "should fail, flag should not escape")

	r = MustParse("(?s)^abc$.^def$")
	checkState(t, !r.Match("abc\ndef"), "multiline mode not on by default")
	r = MustParse("(?ms)^abc$.^def$")
	checkState(t, r.Match("abc\ndef"), "multiline mode works as expected")
}

// Test the behaviour of rune filters.
func TestRuneFilter(t *testing.T) {
	var filter RuneFilter

	filter = MatchRune('#')
	checkState(t, !filter('B'), "should not match random rune")
	checkState(t, filter('#'), "should match configured rune")

	filter = MatchRuneRange('A', 'Z')
	checkState(t, filter('A'), "should match rune 'A' in range")
	checkState(t, filter('B'), "should match rune 'B' in range")

	filter = MatchUnicodeClass("Greek")
	checkState(t, filter('Ω'), "should match omega")
	checkState(t, !filter('Z'), "should not match regular latin rune")

	filter = MatchUnicodeClass("Cyrillic").Not()
	checkState(t, filter('%'), "should match a random non-Cyrillic rune")
	checkState(t, !filter('Ӄ'), "should not match Cyrillic rune")
}

// Test the SafeParser used by much of the code.
func TestStringParser(t *testing.T) {
	src := NewSafeReader("a{bc}d")

	checkState(t, src.curr() == -1, "should not yet be parsing")
	checkState(t, src.nextCh() == 'a', "first char should be a")
	checkState(t, src.nextCh() == '{', "second char should be {")
	lit := src.literal("{", "}")
	checkState(t, lit == "bc", "should equal contained value, got: "+lit)
	checkState(t, src.curr() == 'd', "should now rest on d")
	checkState(t, src.nextCh() == -1, "should be done now")
}
