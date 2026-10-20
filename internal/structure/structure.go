// Package structure reads and writes structure files: the tab-indented tree of
// entries in .crube/structure.txt and .crube/structure_<tag>.txt
// (docs/reference/file-format.md). It only converts between bytes and a Tree;
// finding, matching and writing the file to disk are done elsewhere.
package structure

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

// Entry is one entry of a structure file: a name (a file name, a directory
// name ending in "/", or a pattern) and its description.
type Entry struct {
	Name string
	// Desc is the one-line description. Empty means the entry has none: a
	// line that only gives structure.
	Desc string
	// Detail is the supplement, its lines joined with "\n". Empty means none.
	Detail   string
	Children []*Entry
}

// IsDir reports whether the entry is for directories (its name ends in "/").
func (e *Entry) IsDir() bool {
	return strings.HasSuffix(e.Name, "/")
}

// Tree is the content of one structure file. Entries are those directly under
// the project root, in the order they are written; the order is kept because
// patterns are matched in that order.
type Tree struct {
	Entries []*Entry
}

// ParseError is an error in a structure file. A file with an error is
// rejected as a whole.
type ParseError struct {
	// File is the path of the file. Parse leaves it empty; the caller that
	// read the file sets it.
	File string
	Line int
	Msg  string
}

func (e *ParseError) Error() string {
	if e.File == "" {
		return fmt.Sprintf("line %d: %s", e.Line, e.Msg)
	}
	return fmt.Sprintf("%s:%d: %s", e.File, e.Line, e.Msg)
}

// Parse reads the content of a structure file. Lines may end in LF or CRLF;
// blank lines and trailing white space are ignored.
func Parse(data []byte) (*Tree, error) {
	t := &Tree{}
	p := &parser{frames: []frame{{children: &t.Entries, names: map[string]int{}}}}
	for i, line := range strings.Split(string(data), "\n") {
		if err := p.line(i+1, line); err != nil {
			return nil, err
		}
	}
	p.flush()
	return t, nil
}

// frame is a directory that can still get children: the root, or an entry on
// the path to the last entry read.
type frame struct {
	entry    *Entry // nil for the root
	children *[]*Entry
	names    map[string]int // name -> line, to report duplicates
}

type parser struct {
	// frames[d] is the parent of an entry at depth d.
	frames []frame
	// last is the last entry read; detail lines belong to it.
	last      *Entry
	lastDepth int
	detail    []string
}

func (p *parser) line(n int, line string) error {
	line = strings.TrimRightFunc(line, unicode.IsSpace)
	if line == "" {
		return nil
	}
	body := strings.TrimLeft(line, "\t")
	depth := len(line) - len(body)
	if strings.HasPrefix(body, " ") {
		return &ParseError{Line: n, Msg: "indentation must be tabs"}
	}

	if body == "|" || strings.HasPrefix(body, "| ") {
		if p.last == nil || depth != p.lastDepth+1 {
			return &ParseError{Line: n, Msg: "detail line must come right after its entry, one level deeper"}
		}
		p.detail = append(p.detail, strings.TrimPrefix(body[1:], " "))
		return nil
	}

	p.flush()
	if depth >= len(p.frames) {
		return &ParseError{Line: n, Msg: "indented more than one level deeper than the entry above"}
	}
	parent := p.frames[depth]
	if parent.entry != nil && !parent.entry.IsDir() {
		return &ParseError{Line: n, Msg: fmt.Sprintf("file entry %q cannot have children (directory names end in /)", parent.entry.Name)}
	}

	name, desc, _ := strings.Cut(body, "\t")
	base := strings.TrimSuffix(name, "/")
	if base == "" {
		return &ParseError{Line: n, Msg: "empty name"}
	}
	if strings.Contains(base, "/") {
		return &ParseError{Line: n, Msg: fmt.Sprintf("name %q contains / (write each directory on its own line)", name)}
	}
	if first, ok := parent.names[name]; ok {
		return &ParseError{Line: n, Msg: fmt.Sprintf("duplicate entry %q (first on line %d)", name, first)}
	}
	parent.names[name] = n

	e := &Entry{Name: name, Desc: strings.TrimSpace(desc)}
	*parent.children = append(*parent.children, e)
	p.frames = append(p.frames[:depth+1], frame{entry: e, children: &e.Children, names: map[string]int{}})
	p.last, p.lastDepth = e, depth
	return nil
}

// flush gives the detail lines read so far to the entry they follow.
func (p *parser) flush() {
	if p.last != nil && len(p.detail) > 0 {
		p.last.Detail = strings.Join(p.detail, "\n")
	}
	p.detail = nil
}

// Format writes the tree in the canonical form: tabs for indentation, one tab
// between name and desc, detail lines as "| text" ("|" for an empty line), no
// blank lines and no trailing white space. The order of entries is kept.
func (t *Tree) Format() []byte {
	var b bytes.Buffer
	writeEntries(&b, t.Entries, 0)
	return b.Bytes()
}

func writeEntries(b *bytes.Buffer, entries []*Entry, depth int) {
	indent := strings.Repeat("\t", depth)
	for _, e := range entries {
		b.WriteString(indent)
		b.WriteString(e.Name)
		if desc := strings.TrimSpace(e.Desc); desc != "" {
			b.WriteByte('\t')
			b.WriteString(desc)
		}
		b.WriteByte('\n')
		if e.Detail != "" {
			for _, line := range strings.Split(e.Detail, "\n") {
				b.WriteString(indent)
				b.WriteByte('\t')
				if line = strings.TrimRightFunc(line, unicode.IsSpace); line == "" {
					b.WriteString("|\n")
				} else {
					b.WriteString("| " + line + "\n")
				}
			}
		}
		writeEntries(b, e.Children, depth+1)
	}
}
