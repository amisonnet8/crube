// Package match decides which entry of a structure file an existing path
// matches (docs/reference/file-format.md "照合の規則").
//
// In each directory, an entry whose name is exactly the name wins; otherwise
// the entries are tried in the order they are written, as patterns of
// path.Match, and the first that matches wins. An entry whose name ends in "/"
// matches only directories, any other only files (symbolic links count as
// files). A pattern matches only names directly in its directory; below a
// directory matched by a pattern, matching goes on with that entry's children.
package match

import (
	"path"
	"strings"

	"github.com/amisonnet8/crube/internal/structure"
)

// Child returns the entry among entries (the entries of one directory) that
// matches name, a name directly in that directory, or nil if none does.
// isDir tells whether name is a directory.
func Child(entries []*structure.Entry, name string, isDir bool) *structure.Entry {
	exact := name
	if isDir {
		exact += "/"
	}
	for _, e := range entries {
		if e.Name == exact {
			return e
		}
	}
	for _, e := range entries {
		if e.IsDir() != isDir {
			continue
		}
		// A malformed pattern (such as "[") never matches as a pattern.
		if ok, err := path.Match(strings.TrimSuffix(e.Name, "/"), name); err == nil && ok {
			return e
		}
	}
	return nil
}

// Walk matches the path given as its names, from the project root down, and
// returns the entries matched, one for each name. It stops at the first name
// that no entry matches, so the result is shorter than names when the whole
// path is not matched. All names but the last are directories; isDir tells
// whether the last one is.
func Walk(entries []*structure.Entry, names []string, isDir bool) []*structure.Entry {
	var chain []*structure.Entry
	for i, name := range names {
		e := Child(entries, name, isDir || i < len(names)-1)
		if e == nil {
			break
		}
		chain = append(chain, e)
		entries = e.Children
	}
	return chain
}

// Match is the entry matched by a path, and the path of that entry in the
// structure file (such as "data/*.csv" for data/2026.csv).
type Match struct {
	Entry *structure.Entry
	Path  string
}

// Find returns the entry that the path (its names from the project root)
// matches, or nil if the path, or one of the directories above it, matches no
// entry.
func Find(entries []*structure.Entry, names []string, isDir bool) *Match {
	chain := Walk(entries, names, isDir)
	if len(names) == 0 || len(chain) < len(names) {
		return nil
	}
	return &Match{Entry: chain[len(chain)-1], Path: EntryPath(chain)}
}

// EntryPath joins the names of a chain of entries, from the root down, into
// the path of the last entry in the structure file ("src/auth/", "data/*.csv").
func EntryPath(chain []*structure.Entry) string {
	var b strings.Builder
	for _, e := range chain {
		b.WriteString(e.Name)
	}
	return b.String()
}
