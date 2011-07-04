package sre2

// This file describes ASCII character ranges as per "Perl character classes"
// and "ASCII character classes" (POSIX) on the RE2 syntax page, found here:
// http://code.google.com/p/re2/wiki/Syntax

import "unicode"

var posix_groups = map[string]*unicode.RangeTable{
	"alnum": &unicode.RangeTable{
		R16: []unicode.Range16{
			{'0', '9', 1},
			{'A', 'Z', 1},
			{'a', 'z', 1},
		},
	},
	"alpha": &unicode.RangeTable{
		R16: []unicode.Range16{
			{'A', 'Z', 1},
			{'a', 'z', 1},
		},
	},
	"ascii": &unicode.RangeTable{
		R16: []unicode.Range16{
			{0x00, 0x7f, 1},
		},
	},
	"blank": &unicode.RangeTable{
		R16: []unicode.Range16{
			{'\t', '\t', 1},
			{' ', ' ', 1},
		},
	},
	"cntrl": &unicode.RangeTable{
		R16: []unicode.Range16{
			{0x00, 0x1f, 1},
			{0x7f, 0x7f, 1},
		},
	},
	"digit": &unicode.RangeTable{
		R16: []unicode.Range16{
			{'0', '9', 1},
		},
	},
	"graph": &unicode.RangeTable{
		R16: []unicode.Range16{
			{'!', '~', 1},
		},
	},
	"lower": &unicode.RangeTable{
		R16: []unicode.Range16{
			{'a', 'z', 1},
		},
	},
	"print": &unicode.RangeTable{
		R16: []unicode.Range16{
			{' ', '~', 1},
		},
	},
	"punct": &unicode.RangeTable{
		R16: []unicode.Range16{
			{'!', '/', 1},
			{':', '@', 1},
			{'[', '`', 1},
			{'{', '~', 1},
		},
	},
	"space": &unicode.RangeTable{
		R16: []unicode.Range16{
			{'\t', '\r', 1},
			{' ', ' ', 1},
		},
	},
	"upper": &unicode.RangeTable{
		R16: []unicode.Range16{
			{'A', 'Z', 1},
		},
	},
	"word": &unicode.RangeTable{
		R16: []unicode.Range16{
			{'0', '9', 1},
			{'A', 'Z', 1},
			{'a', 'z', 1},
		},
	},
	"xdigit": &unicode.RangeTable{
		R16: []unicode.Range16{
			{'0', '9', 1},
			{'A', 'F', 1},
			{'a', 'f', 1},
		},
	},
}

var perl_groups = map[int]*unicode.RangeTable{
	'd': posix_groups["digit"],
	'w': posix_groups["word"],
	's': &unicode.RangeTable{
		R16: []unicode.Range16{
			{'\t', '\n', 1},
			{'\f', '\r', 1},
			{' ', ' ', 1},
		},
	},
}
