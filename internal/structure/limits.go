package structure

import (
	"errors"
	"fmt"
	"strings"

	"github.com/amisonnet8/crube/internal/textwidth"
)

// The limits of desc and detail, in display width (docs/reference/file-format.md
// "長さの上限"). crube has only these limits; it has no recommended length.
const (
	MaxDescWidth   = 256
	MaxDetailWidth = 4096 // the sum over all lines of the detail
)

// ErrDescNewline is returned by CheckDesc for a desc with a line break: desc is
// shown on one line by ls and tree.
var ErrDescNewline = errors.New("desc must be one line (put more in detail)")

// LimitError is a desc or a detail wider than its limit.
type LimitError struct {
	Field string // "desc" or "detail"
	Width int
	Limit int
}

func (e *LimitError) Error() string {
	return fmt.Sprintf("%s is %d columns wide (limit %d)", e.Field, e.Width, e.Limit)
}

// CheckDesc reports whether desc can be written: one line, at most
// MaxDescWidth wide.
func CheckDesc(desc string) error {
	if strings.ContainsAny(desc, "\r\n") {
		return ErrDescNewline
	}
	if w := textwidth.String(desc); w > MaxDescWidth {
		return &LimitError{Field: "desc", Width: w, Limit: MaxDescWidth}
	}
	return nil
}

// CheckDetail reports whether detail can be written: the widths of its lines
// add up to at most MaxDetailWidth. Line breaks are not counted.
func CheckDetail(detail string) error {
	w := 0
	for _, line := range strings.Split(detail, "\n") {
		w += textwidth.String(line)
	}
	if w > MaxDetailWidth {
		return &LimitError{Field: "detail", Width: w, Limit: MaxDetailWidth}
	}
	return nil
}

// CheckName reports whether name (with a trailing "/" for a directory entry)
// can be written as the name of an entry: the names that the format cannot
// hold (docs/reference/file-format.md "Names that cannot be written") are
// refused, as are an empty name, a "/" inside it, and "." and "..".
func CheckName(name string) error {
	base := strings.TrimSuffix(name, "/")
	if base == "" || base == "." || base == ".." ||
		strings.ContainsAny(base, "/\t\r\n") ||
		strings.HasPrefix(base, " ") || strings.HasSuffix(base, " ") ||
		base == "|" || strings.HasPrefix(base, "| ") {
		return fmt.Errorf("name %q cannot be written in a structure file", name)
	}
	return nil
}
