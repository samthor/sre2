package sre2

// This file describes ASCII character ranges as per "Perl character classes"
// and "ASCII character classes" (POSIX) on the RE2 syntax page, found here:
// http://code.google.com/p/re2/wiki/Syntax

import "unicode"

var posix_groups = map[string][]unicode.Range{
	"alnum": {
		unicode.Range{'0', '9', 1},
		unicode.Range{'A', 'Z', 1},
		unicode.Range{'a', 'z', 1},
	},
	"alpha": {
		unicode.Range{'A', 'Z', 1},
		unicode.Range{'a', 'z', 1},
	},
	"ascii": {
		unicode.Range{0x00, 0x7f, 1},
	},
	"blank": {
		unicode.Range{'\t', '\t', 1},
		unicode.Range{' ', ' ', 1},
	},
	"cntrl": {
		unicode.Range{0x00, 0x1f, 1},
		unicode.Range{0x7f, 0x7f, 1},
	},
	"digit": {
		unicode.Range{'0', '9', 1},
	},
	"graph": {
		unicode.Range{'!', '~', 1},
	},
	"lower": {
		unicode.Range{'a', 'z', 1},
	},
	"print": {
		unicode.Range{' ', '~', 1},
	},
	"punct": {
		unicode.Range{'!', '/', 1},
		unicode.Range{':', '@', 1},
		unicode.Range{'[', '`', 1},
		unicode.Range{'{', '~', 1},
	},
	"space": {
		unicode.Range{'\t', '\r', 1},
		unicode.Range{' ', ' ', 1},
	},
	"upper": {
		unicode.Range{'A', 'Z', 1},
	},
	"word": {
		unicode.Range{'0', '9', 1},
		unicode.Range{'A', 'Z', 1},
		unicode.Range{'a', 'z', 1},
	},
	"xdigit": {
		unicode.Range{'0', '9', 1},
		unicode.Range{'A', 'F', 1},
		unicode.Range{'a', 'f', 1},
	},
}

var perl_groups = map[int][]unicode.Range{
	'd': posix_groups["digit"],
	'w': posix_groups["word"],
	's': {
		unicode.Range{'\t', '\n', 1},
		unicode.Range{'\f', '\r', 1},
		unicode.Range{' ', ' ', 1},
	},
}
