// Package textwidth measures text by its display width on a terminal, the
// unit crube uses for the limits of desc and detail and for aligning columns
// (docs/design/data-model.md "Length limits").
package textwidth

import (
	"unicode"

	"golang.org/x/text/width"
)

// Rune returns the display width of r: 2 for wide and fullwidth characters
// (East Asian Width W and F), 0 for combining marks and format characters
// (such as U+0301 and U+200D, which take no column of their own), and 1 for
// everything else. Ambiguous characters (A, such as ○ and ①) count as 1.
func Rune(r rune) int {
	if unicode.In(r, unicode.Mn, unicode.Me, unicode.Cf) {
		return 0
	}
	switch width.LookupRune(r).Kind() {
	case width.EastAsianWide, width.EastAsianFullwidth:
		return 2
	default:
		return 1
	}
}

// String returns the display width of s, the sum of the widths of its runes.
func String(s string) int {
	n := 0
	for _, r := range s {
		n += Rune(r)
	}
	return n
}
