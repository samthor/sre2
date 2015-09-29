package sre2

// Note: This file is pulled from Go's regexp package, with non-benchmark tests removed. This is
// useful for comparison with the regexp baseline. Interestingly enough, because these tests all
// work on very 'small' strings, we're actually slower in most cases.
// TODO: Add a string prefix matcher (i.e. not as part of the regexp prog), as this seems to be
// where regexp gets most of its speed gains.

import (
	"strings"
	"testing"
)

func BenchmarkLiteral(b *testing.B) {
	x := strings.Repeat("x", 50) + "y"
	b.StopTimer()
	re := MustParse("y")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if !re.Match(x) {
			println("no match!")
			break
		}
	}
}

func BenchmarkNotLiteral(b *testing.B) {
	x := strings.Repeat("x", 50) + "y"
	b.StopTimer()
	re := MustParse(".y")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if !re.Match(x) {
			println("no match!")
			break
		}
	}
}

func BenchmarkMatchClass(b *testing.B) {
	b.StopTimer()
	x := strings.Repeat("xxxx", 20) + "w"
	re := MustParse("[abcdw]")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if !re.Match(x) {
			println("no match!")
			break
		}
	}
}

func BenchmarkMatchClass_InRange(b *testing.B) {
	b.StopTimer()
	// 'b' is between 'a' and 'c', so the charclass
	// range checking is no help here.
	x := strings.Repeat("bbbb", 20) + "c"
	re := MustParse("[ac]")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if !re.Match(x) {
			println("no match!")
			break
		}
	}
}
/*
func BenchmarkReplaceAll(b *testing.B) {
	x := "abcdefghijklmnopqrstuvwxyz"
	b.StopTimer()
	re := MustParse("[cjrw]")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		re.ReplaceAllString(x, "")
	}
}
*/
func BenchmarkAnchoredLiteralShortNonMatch(b *testing.B) {
	b.StopTimer()
	x := "abcdefghijklmnopqrstuvwxyz"
	re := MustParse("^zbc(d|e)")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		re.Match(x)
	}
}

func BenchmarkAnchoredLiteralLongNonMatch(b *testing.B) {
	b.StopTimer()
	x := strings.Repeat("abcdefghijklmnopqrstuvwxyz", 16)
	re := MustParse("^zbc(d|e)")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		re.Match(x)
	}
}

func BenchmarkAnchoredShortMatch(b *testing.B) {
	b.StopTimer()
	x := "abcdefghijklmnopqrstuvwxyz"
	re := MustParse("^.bc(d|e)")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		re.Match(x)
	}
}

func BenchmarkAnchoredLongMatch(b *testing.B) {
	b.StopTimer()
	x := strings.Repeat("abcdefghijklmnopqrstuvwxyz", 16)
	re := MustParse("^.bc(d|e)")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		re.Match(x)
	}
}

func BenchmarkComplexRe(b *testing.B) {
	b.StopTimer()
	re := MustParse(".*(a|(b))+(#*).+")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		re.Match("aba#hello")
	}
}
