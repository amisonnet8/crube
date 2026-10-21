package core

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/amisonnet8/crube/internal/structure"
)

// SetResult is the result of Set: the entry as written.
type SetResult struct {
	Path   string `json:"path"` // the entry path
	Desc   string `json:"desc"`
	Detail string `json:"detail"`
}

// RemoveResult is the result of Remove.
type RemoveResult struct {
	Path    string `json:"path"`    // the entry path removed
	Removed int    `json:"removed"` // how many entries went: the one at Path and those below it
}

// MoveResult is the result of Move.
type MoveResult struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Moved int    `json:"moved"` // how many entries moved: the one at From and those below it
}

// Set adds or updates the entry at entryPath (such as "data/" or
// "data/*.csv") in the structure file of the language: its desc, and its
// detail when detail is not nil ("" removes the detail). Missing parents are
// added as lines without a desc; a new entry goes after its siblings. It
// returns the entry as written.
func (p *Project) Set(entryPath, desc string, detail *string) (*SetResult, error) {
	entryPath = normEntryPath(entryPath)
	names, err := splitEntryPath(entryPath)
	if err != nil {
		return nil, err
	}
	if strings.ContainsAny(desc, "\r\n") {
		return nil, structure.ErrDescNewline
	}
	desc = strings.TrimSpace(desc)
	if err := structure.CheckDesc(desc); err != nil {
		return nil, err
	}
	var d string
	if detail != nil {
		d = normalizeDetail(*detail)
		if err := structure.CheckDetail(d); err != nil {
			return nil, err
		}
	}

	var e *structure.Entry
	err = p.update(func(tree *structure.Tree) error {
		list := ensureParents(tree, names[:len(names)-1])
		last := names[len(names)-1]
		if i := index(*list, last); i >= 0 {
			e = (*list)[i]
		} else {
			e = &structure.Entry{Name: last}
			*list = append(*list, e)
		}
		e.Desc = desc
		if detail != nil {
			e.Detail = d
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &SetResult{Path: entryPath, Desc: e.Desc, Detail: e.Detail}, nil
}

// Remove removes the entry at entryPath, and every entry below it, from the
// structure file of the language. Files of the project are not touched.
func (p *Project) Remove(entryPath string) (*RemoveResult, error) {
	entryPath = normEntryPath(entryPath)
	names, err := splitEntryPath(entryPath)
	if err != nil {
		return nil, err
	}
	var e *structure.Entry
	err = p.update(func(tree *structure.Tree) error {
		list, i := lookup(tree, names)
		if i < 0 {
			return fmt.Errorf("no entry %s", entryPath)
		}
		e = (*list)[i]
		*list = slices.Delete(*list, i, i+1)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &RemoveResult{Path: entryPath, Removed: count(e)}, nil
}

// Move moves the entry at from, with every entry below it, to the entry path
// to, in the structure file of the language. Missing parents of to are added
// as lines without a desc; the entry goes after its new siblings. Files of
// the project are not touched.
func (p *Project) Move(from, to string) (*MoveResult, error) {
	from, to = normEntryPath(from), normEntryPath(to)
	fromNames, err := splitEntryPath(from)
	if err != nil {
		return nil, err
	}
	toNames, err := splitEntryPath(to)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(from, "/") != strings.HasSuffix(to, "/") {
		return nil, fmt.Errorf("cannot move %s to %s: a directory entry (ending in /) stays a directory entry, and a file entry a file entry", from, to)
	}
	if len(toNames) > len(fromNames) && slices.Equal(toNames[:len(fromNames)], fromNames) {
		return nil, fmt.Errorf("cannot move %s into itself", from)
	}

	var e *structure.Entry
	err = p.update(func(tree *structure.Tree) error {
		list, i := lookup(tree, fromNames)
		if i < 0 {
			return fmt.Errorf("no entry %s", from)
		}
		if _, j := lookup(tree, toNames); j >= 0 {
			return fmt.Errorf("entry %s already exists", to)
		}
		e = (*list)[i]
		*list = slices.Delete(*list, i, i+1)
		dst := ensureParents(tree, toNames[:len(toNames)-1])
		e.Name = toNames[len(toNames)-1]
		*dst = append(*dst, e)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &MoveResult{From: from, To: to, Moved: count(e)}, nil
}

// count is the number of entries in e: e itself and every entry below it.
func count(e *structure.Entry) int {
	n := 1
	for _, c := range e.Children {
		n += count(c)
	}
	return n
}

// osSeparator is the separator of real paths on this OS. It is a variable so
// that tests can act as Windows.
var osSeparator = filepath.Separator

// normEntryPath reads an entry path as a user gives it: on Windows, where "\"
// cannot be in a name, "\" separates names as "/" does. Elsewhere "\" is a
// character of a name, and the path is left as it is.
func normEntryPath(entryPath string) string {
	if osSeparator == '\\' {
		return strings.ReplaceAll(entryPath, `\`, "/")
	}
	return entryPath
}

// splitEntryPath splits an entry path into the names of its entries, each
// with its trailing "/" if it is a directory entry: "src/auth/" is ["src/"
// "auth/"], "data/*.csv" is ["data/" "*.csv"].
func splitEntryPath(entryPath string) ([]string, error) {
	isDir := strings.HasSuffix(entryPath, "/")
	parts := strings.Split(strings.TrimSuffix(entryPath, "/"), "/")
	names := make([]string, len(parts))
	for i, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("invalid entry path %q", entryPath)
		}
		name := part
		if i < len(parts)-1 || isDir {
			name += "/"
		}
		if err := structure.CheckName(name); err != nil {
			return nil, err
		}
		names[i] = name
	}
	return names, nil
}

// normalizeDetail turns CRLF and CR into LF and drops the white space at the
// end of each line, as the structure file will hold it.
func normalizeDetail(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	return strings.Join(lines, "\n")
}

// index returns the position of the entry named name in entries, or -1.
// Names are compared exactly: an entry path names entries, it is not matched
// as a pattern.
func index(entries []*structure.Entry, name string) int {
	return slices.IndexFunc(entries, func(e *structure.Entry) bool { return e.Name == name })
}

// lookup returns the list that holds the entry at names, and its position in
// it; the position is -1 if there is no such entry.
func lookup(tree *structure.Tree, names []string) (*[]*structure.Entry, int) {
	list := &tree.Entries
	for i, name := range names {
		j := index(*list, name)
		if j < 0 {
			return nil, -1
		}
		if i == len(names)-1 {
			return list, j
		}
		list = &(*list)[j].Children
	}
	return nil, -1
}

// ensureParents returns the list of children of the entry at names, adding
// the missing entries on the way as lines without a desc.
func ensureParents(tree *structure.Tree, names []string) *[]*structure.Entry {
	list := &tree.Entries
	for _, name := range names {
		j := index(*list, name)
		if j < 0 {
			*list = append(*list, &structure.Entry{Name: name})
			j = len(*list) - 1
		}
		list = &(*list)[j].Children
	}
	return list
}
