// Package status compares the project's files with a structure file and
// reports what is not described, what describes nothing, and what is over a
// limit (docs/design/freshness.md). It only reads: it changes neither the
// structure file nor the project.
package status

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/amisonnet8/crube/internal/ignore"
	"github.com/amisonnet8/crube/internal/match"
	"github.com/amisonnet8/crube/internal/structure"
)

// Kinds of items.
const (
	Untracked = "untracked" // exists, but no entry with a desc matches it
	Orphan    = "orphan"    // an entry that no existing path matches
	Invalid   = "invalid"   // a desc or detail over its limit
)

// Item is one finding. Paths are relative to the project root, with "/", and
// end in "/" for directories.
type Item struct {
	Kind  string `json:"kind"`
	Type  string `json:"type,omitempty"`  // untracked: "dir" or "file"
	Path  string `json:"path"`            // untracked: the existing path; orphan, invalid: the entry path
	Field string `json:"field,omitempty"` // invalid: "desc" or "detail"
	Width int    `json:"width,omitempty"` // invalid
	Limit int    `json:"limit,omitempty"` // invalid
}

// Level is whether one level of coverage is met, and how many paths miss it.
type Level struct {
	OK      bool `json:"ok"`
	Missing int  `json:"missing"`
}

// Coverage is how far the descriptions go: every file and directory (All),
// or every directory (Dirs).
type Coverage struct {
	All  Level `json:"all"`
	Dirs Level `json:"dirs"`
}

// Report is the result of Check. Items are in the order of the report:
// untracked directories, untracked files, orphans, invalid values.
type Report struct {
	// Structure is the file that was checked (".crube/structure.txt"). Check
	// leaves it empty; the caller that chose the file sets it.
	Structure string   `json:"structure"`
	Items     []Item   `json:"items"`
	Coverage  Coverage `json:"coverage"`
}

// HasProblems reports whether there is an orphan or an invalid value: the
// findings that make status exit with 1 whatever coverage is required.
func (r *Report) HasProblems() bool {
	for _, it := range r.Items {
		if it.Kind == Orphan || it.Kind == Invalid {
			return true
		}
	}
	return false
}

// Check walks the project under root and compares it with tree. ign decides
// the paths that are left out.
func Check(root string, tree *structure.Tree, ign *ignore.Matcher) (*Report, error) {
	w := &walker{
		root:    root,
		ign:     ign,
		hit:     map[*structure.Entry]bool{},
		reached: map[*structure.Entry]bool{},
	}
	w.reachAll(tree.Entries)
	if _, err := w.walk(nil, tree.Entries); err != nil {
		return nil, err
	}

	r := &Report{Items: []Item{}}
	r.Items = append(r.Items, w.dirs...)
	r.Items = append(r.Items, w.files...)
	w.orphans(tree.Entries, "", r)
	invalid(tree.Entries, "", r)
	r.Coverage.All = level(len(w.dirs) + len(w.files))
	r.Coverage.Dirs = level(len(w.dirs))
	return r, nil
}

func level(missing int) Level {
	return Level{OK: missing == 0, Missing: missing}
}

type walker struct {
	root  string
	ign   *ignore.Matcher
	dirs  []Item // untracked directories, in walk order
	files []Item // untracked files, in walk order
	// hit holds the entries that an existing path matches (ignored and
	// empty ones too: they exist). reached holds the entries whose directory
	// was looked into, so that only they can be orphans.
	hit     map[*structure.Entry]bool
	reached map[*structure.Entry]bool
}

func (w *walker) reachAll(entries []*structure.Entry) {
	for _, e := range entries {
		w.reached[e] = true
	}
}

// walk looks into the directory given by names (from the project root;
// empty for the root) whose entries are the children of the entry matched by
// it. It reports whether the directory holds anything to check: a directory
// with nothing below it but ignored paths and empty directories is left out.
func (w *walker) walk(names []string, entries []*structure.Entry) (bool, error) {
	list, err := os.ReadDir(filepath.Join(append([]string{w.root}, names...)...))
	if err != nil {
		return false, err
	}
	found := false
	var subdirs []string
	// Files first, then subdirectories, both in name order (os.ReadDir sorts).
	for _, de := range list {
		// A symbolic link is not followed: it is a file, whatever it points to.
		isDir := de.IsDir() && de.Type()&os.ModeSymlink == 0
		e := match.Child(entries, de.Name(), isDir)
		if e != nil {
			w.hit[e] = true
		}
		if isDir {
			subdirs = append(subdirs, de.Name())
			continue
		}
		p := append(names[:len(names):len(names)], de.Name())
		if w.ign.Ignored(p, false) {
			continue
		}
		found = true
		if !described(e) {
			w.files = append(w.files, Item{Kind: Untracked, Type: "file", Path: strings.Join(p, "/")})
		}
	}
	for _, name := range subdirs {
		p := append(names[:len(names):len(names)], name)
		if w.ign.Ignored(p, true) {
			continue
		}
		e := match.Child(entries, name, true)
		var children []*structure.Entry
		if e != nil {
			children = e.Children
			w.reachAll(children)
		}
		// Report the directory before what is below it; take it back if it
		// turns out to be empty (then nothing was reported below it, so it is
		// the last item).
		at := len(w.dirs)
		if !described(e) {
			w.dirs = append(w.dirs, Item{Kind: Untracked, Type: "dir", Path: strings.Join(p, "/") + "/"})
		}
		sub, err := w.walk(p, children)
		if err != nil {
			return false, err
		}
		if !sub {
			w.dirs = w.dirs[:at]
			continue
		}
		found = true
	}
	return found, nil
}

// described reports whether a path matched by e counts as described: an
// entry with a desc matches it. A line without a desc only gives structure.
func described(e *structure.Entry) bool {
	return e != nil && e.Desc != ""
}

// orphans adds, in the order of the structure file, the entries that were
// looked for but that no existing path matches. What is below an orphan, or
// below a directory that was not looked into (ignored), is not reported.
func (w *walker) orphans(entries []*structure.Entry, prefix string, r *Report) {
	for _, e := range entries {
		if !w.reached[e] {
			continue
		}
		if !w.hit[e] {
			r.Items = append(r.Items, Item{Kind: Orphan, Path: prefix + e.Name})
			continue
		}
		w.orphans(e.Children, prefix+e.Name, r)
	}
}

// invalid adds, in the order of the structure file, every desc and detail
// over its limit (possible in a file edited by hand).
func invalid(entries []*structure.Entry, prefix string, r *Report) {
	for _, e := range entries {
		p := prefix + e.Name
		for _, err := range []error{structure.CheckDesc(e.Desc), structure.CheckDetail(e.Detail)} {
			var le *structure.LimitError
			if errors.As(err, &le) {
				r.Items = append(r.Items, Item{Kind: Invalid, Path: p, Field: le.Field, Width: le.Width, Limit: le.Limit})
			}
		}
		invalid(e.Children, p, r)
	}
}
