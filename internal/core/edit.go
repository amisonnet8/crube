package core

import (
	"fmt"
	"os"
	"path"
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
	entryPath, names, err := p.entryBelowRoot(entryPath)
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
		if entryPath, names, err = pickKind(tree, entryPath, names); err != nil {
			return err
		}
		list := ensureParents(tree, names[:len(names)-1])
		last := names[len(names)-1]
		if i := index(*list, last); i >= 0 {
			e = (*list)[i]
		} else {
			// A new file entry: the user did not end the path in "/", so make
			// sure that it is not a real directory rather than guess.
			if !strings.HasSuffix(last, "/") && p.isRealDir(entryPath) {
				return fmt.Errorf("%s: is a directory (use %s/)", entryPath, entryPath)
			}
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
	entryPath, names, err := p.entryBelowRoot(entryPath)
	if err != nil {
		return nil, err
	}
	var e *structure.Entry
	err = p.update(func(tree *structure.Tree) error {
		if entryPath, names, err = pickKind(tree, entryPath, names); err != nil {
			return err
		}
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
	from, fromNames, err := p.entryBelowRoot(from)
	if err != nil {
		return nil, err
	}
	to, toNames, err := p.entryBelowRoot(to)
	if err != nil {
		return nil, err
	}
	var e *structure.Entry
	err = p.update(func(tree *structure.Tree) error {
		if from, fromNames, err = pickKind(tree, from, fromNames); err != nil {
			return err
		}
		// to without a trailing "/" is of the same kind as from.
		if strings.HasSuffix(from, "/") && !strings.HasSuffix(to, "/") {
			to += "/"
			toNames[len(toNames)-1] += "/"
		}
		if strings.HasSuffix(from, "/") != strings.HasSuffix(to, "/") {
			return fmt.Errorf("cannot move %s to %s: a directory entry (ending in /) stays a directory entry, and a file entry a file entry", from, to)
		}
		if len(toNames) > len(fromNames) && slices.Equal(toNames[:len(fromNames)], fromNames) {
			return fmt.Errorf("cannot move %s into itself", from)
		}
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

// entryPath reads an entry path as a user types it: from where the operation
// runs (the project root for the MCP server), as ls reads a real path, so
// that ".", "..", empty names ("a//b") and an absolute path are read the same
// way. Only the names are cleaned: the last one may still be a pattern. A
// path ending in "/", "." or ".." is a directory entry. It returns the entry
// path from the project root and the names of its entries, or "" and no
// names for the project root itself. A path outside the project is refused.
func (p *Project) entryPath(typed string) (string, []string, error) {
	s := normEntryPath(typed)
	if s == "" {
		return "", nil, fmt.Errorf("invalid entry path %q", typed)
	}
	last := s[strings.LastIndex(s, "/")+1:]
	isDir := last == "" || last == "." || last == ".."
	var rel string
	if filepath.IsAbs(s) {
		var err error
		if rel, _, err = p.realPath(s); err != nil {
			return "", nil, err
		}
	} else {
		// With "/" only, so that "\" stays a character of a name where it can
		// be one (normEntryPath has turned it into "/" where it cannot).
		here, err := filepath.Rel(p.Root, p.dir)
		if err != nil {
			return "", nil, err
		}
		rel = path.Join(filepath.ToSlash(here), s)
		if rel == ".." || strings.HasPrefix(rel, "../") {
			return "", nil, fmt.Errorf("%s: outside the project", s)
		}
		if rel == "." {
			rel = ""
		}
	}
	if rel == "" {
		return "", nil, nil
	}
	if isDir {
		rel += "/"
	}
	names, err := splitEntryPath(rel)
	if err != nil {
		return "", nil, err
	}
	return rel, names, nil
}

// entryBelowRoot is entryPath for an operation that needs an entry: the
// project root has none.
func (p *Project) entryBelowRoot(typed string) (string, []string, error) {
	path, names, err := p.entryPath(typed)
	if err == nil && names == nil {
		err = rootHasNoEntry(typed)
	}
	return path, names, err
}

func rootHasNoEntry(path string) error {
	return fmt.Errorf("%s: the project root has no entry (give a path below it)", path)
}

// pickKind lets an entry path that does not end in "/" name a directory
// entry: when the directory entry of that name is there and the file entry
// is not, it returns the directory entry's path and names. When both are
// there, the path is ambiguous and refused. Otherwise the path is returned as
// it is, as a file entry.
func pickKind(tree *structure.Tree, entryPath string, names []string) (string, []string, error) {
	if strings.HasSuffix(entryPath, "/") {
		return entryPath, names, nil
	}
	n := len(names) - 1
	dirNames := append(names[:n:n], names[n]+"/")
	_, file := lookup(tree, names)
	_, dir := lookup(tree, dirNames)
	switch {
	case file >= 0 && dir >= 0:
		return "", nil, fmt.Errorf("%s: both %s and %s/ are entries (add / for the directory)", entryPath, entryPath, entryPath)
	case dir >= 0:
		return entryPath + "/", dirNames, nil
	}
	return entryPath, names, nil
}

// isRealDir tells whether the entry path (from the project root, without a
// trailing "/") is a real directory of the project. A symbolic link is not:
// crube takes it as a file.
func (p *Project) isRealDir(entryPath string) bool {
	fi, err := os.Lstat(filepath.Join(p.Root, filepath.FromSlash(entryPath)))
	return err == nil && fi.IsDir()
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
