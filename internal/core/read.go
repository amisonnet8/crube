package core

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/amisonnet8/crube/internal/match"
	"github.com/amisonnet8/crube/internal/search"
	"github.com/amisonnet8/crube/internal/status"
	"github.com/amisonnet8/crube/internal/structure"
)

// The JSON forms below are those of docs/reference/cli.md. Paths are relative
// to the project root, with "/", and end in "/" for directories.

// LsItem is one name directly in the directory listed.
type LsItem struct {
	Name  string `json:"name"`  // ends in "/" for a directory
	Type  string `json:"type"`  // "dir" or "file" (a symbolic link is a file)
	Entry string `json:"entry"` // the path of the entry that matches it; "" for none
	Desc  string `json:"desc"`
}

// LsResult is the result of Ls.
type LsResult struct {
	Path  string   `json:"path"` // the directory listed; "" for the project root
	Items []LsItem `json:"items"`
}

// Ls lists, in name order, what is directly in the directory path (a real
// path, relative to where the operation runs), with the desc of the entry
// that matches each. Ignored names are not listed.
func (p *Project) Ls(path string) (*LsResult, error) {
	rel, names, err := p.realPath(path)
	if err != nil {
		return nil, err
	}
	tree, _, err := p.Load()
	if err != nil {
		return nil, err
	}
	ign, err := p.Ignore()
	if err != nil {
		return nil, err
	}
	fi, err := os.Lstat(filepath.Join(p.Root, filepath.FromSlash(rel)))
	if err != nil {
		return nil, notFound(path, err)
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("%s: not a directory (use crube show)", path)
	}
	if len(names) > 0 && ign.Ignored(names, true) {
		return nil, ignored(path)
	}

	// The entries of the directory, if the directory itself is matched.
	entries := tree.Entries
	prefix := ""
	if len(names) > 0 {
		entries = nil
		if chain := match.Walk(tree.Entries, names, true); len(chain) == len(names) {
			entries = chain[len(chain)-1].Children
			prefix = match.EntryPath(chain)
		}
	}

	list, err := os.ReadDir(filepath.Join(p.Root, filepath.FromSlash(rel)))
	if err != nil {
		return nil, err
	}
	r := &LsResult{Path: dirPath(names), Items: []LsItem{}}
	for _, de := range list {
		isDir := de.IsDir() && de.Type()&fs.ModeSymlink == 0
		if ign.Ignored(append(names[:len(names):len(names)], de.Name()), isDir) {
			continue
		}
		it := LsItem{Name: de.Name(), Type: "file"}
		if isDir {
			it.Name += "/"
			it.Type = "dir"
		}
		if e := match.Child(entries, de.Name(), isDir); e != nil {
			it.Entry = prefix + e.Name
			it.Desc = e.Desc
		}
		r.Items = append(r.Items, it)
	}
	return r, nil
}

// ShowResult is the result of Show.
type ShowResult struct {
	Path   string `json:"path"`
	Type   string `json:"type"`  // "dir" or "file"
	Entry  string `json:"entry"` // the path of the entry that matches it; "" for none
	Desc   string `json:"desc"`
	Detail string `json:"detail"`
	Link   string `json:"link"` // the target of a symbolic link, as written in it but with "/"; "" otherwise
}

// Show returns the desc and the detail of the entry that matches path (a real
// path, relative to where the operation runs).
func (p *Project) Show(path string) (*ShowResult, error) {
	rel, names, err := p.realPath(path)
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("%s: the project root has no entry (give a path below it)", path)
	}
	tree, _, err := p.Load()
	if err != nil {
		return nil, err
	}
	ign, err := p.Ignore()
	if err != nil {
		return nil, err
	}
	full := filepath.Join(p.Root, filepath.FromSlash(rel))
	fi, err := os.Lstat(full)
	if err != nil {
		return nil, notFound(path, err)
	}
	isLink := fi.Mode()&fs.ModeSymlink != 0
	isDir := fi.IsDir() && !isLink
	if ign.Ignored(names, isDir) {
		return nil, ignored(path)
	}

	r := &ShowResult{Path: strings.Join(names, "/"), Type: "file"}
	if isDir {
		r.Path += "/"
		r.Type = "dir"
	}
	if isLink {
		target, err := os.Readlink(full)
		if err != nil {
			return nil, err
		}
		// Windows keeps the target with \ (os.Symlink writes it so).
		r.Link = filepath.ToSlash(target)
	}
	if m := match.Find(tree.Entries, names, isDir); m != nil {
		r.Entry, r.Desc, r.Detail = m.Path, m.Entry.Desc, m.Entry.Detail
	}
	return r, nil
}

// TreeNode is one entry shown by Tree.
type TreeNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"` // the entry path
	Desc     string      `json:"desc"`
	Detail   string      `json:"detail"`
	Here     bool        `json:"here"` // the line that gets the mark of the current place
	Children []*TreeNode `json:"children"`
}

// TreeResult is the result of Tree.
type TreeResult struct {
	// Here is the directory where the operation runs; "" at the project
	// root, and always "" without the mark.
	Here    string      `json:"here"`
	Entries []*TreeNode `json:"entries"`
	// Above is true when the mark is on an entry above Here, because no entry
	// matches Here itself: the CLI then adds "(here: …)" to that line.
	Above bool `json:"-"`
}

// TreeOptions are the options of Tree.
type TreeOptions struct {
	Path  string // an entry path: show only that entry and what is below it
	Depth int    // the number of levels to show; 0 is no limit
	Mark  bool   // mark the entry of the current place (the CLI does; MCP does not)
}

// Tree returns the entries of the structure file as a tree, as they are
// written (the project's files are not looked at). Entries whose path is
// ignored are left out, with what is below them.
func (p *Project) Tree(o TreeOptions) (*TreeResult, error) {
	tree, _, err := p.Load()
	if err != nil {
		return nil, err
	}
	ign, err := p.Ignore()
	if err != nil {
		return nil, err
	}

	top := tree.Entries
	var topNames []string // the names above the entries of top
	if o.Path != "" {
		names, err := splitEntryPath(o.Path)
		if err != nil {
			return nil, err
		}
		list, i := lookup(tree, names)
		if i < 0 {
			return nil, fmt.Errorf("no entry %s", o.Path)
		}
		top = []*structure.Entry{(*list)[i]}
		topNames = names[:len(names)-1]
	}

	r := &TreeResult{Entries: []*TreeNode{}}
	var marked *structure.Entry
	if o.Mark {
		rel, err := filepath.Rel(p.Root, p.dir)
		if err != nil {
			return nil, err
		}
		if here := splitRel(rel); len(here) > 0 {
			r.Here = dirPath(here)
			if chain := match.Walk(tree.Entries, here, true); len(chain) > 0 {
				marked = chain[len(chain)-1]
				r.Above = len(chain) < len(here)
			}
		}
	}

	var build func(entries []*structure.Entry, names []string, level int) []*TreeNode
	build = func(entries []*structure.Entry, names []string, level int) []*TreeNode {
		nodes := []*TreeNode{}
		for _, e := range entries {
			n := append(names[:len(names):len(names)], strings.TrimSuffix(e.Name, "/"))
			if ign.Ignored(n, e.IsDir()) {
				continue
			}
			node := &TreeNode{
				Name:     e.Name,
				Path:     entryPathOf(names, e),
				Desc:     e.Desc,
				Detail:   e.Detail,
				Here:     e == marked,
				Children: []*TreeNode{},
			}
			if o.Depth <= 0 || level < o.Depth {
				node.Children = build(e.Children, n, level+1)
			}
			nodes = append(nodes, node)
		}
		return nodes
	}
	trimmed := make([]string, len(topNames))
	for i, n := range topNames {
		trimmed[i] = strings.TrimSuffix(n, "/")
	}
	r.Entries = build(top, trimmed, 1)
	return r, nil
}

// Find finds entries by their path, desc and detail (see search.Find).
func (p *Project) Find(query string, limit int) (*search.Found, error) {
	tree, _, err := p.Load()
	if err != nil {
		return nil, err
	}
	ign, err := p.Ignore()
	if err != nil {
		return nil, err
	}
	return search.Find(tree, query, limit, ign)
}

// Status compares the project with the structure file of the language (see
// status.Check).
func (p *Project) Status() (*status.Report, error) {
	tree, name, err := p.Load()
	if err != nil {
		return nil, err
	}
	ign, err := p.Ignore()
	if err != nil {
		return nil, err
	}
	r, err := status.Check(p.Root, tree, ign)
	if err != nil {
		return nil, err
	}
	r.Structure = relPath(name)
	return r, nil
}

// Failed reports whether status ends with exit code 1: there is an orphan or
// an invalid value, or the coverage required ("all" or "dirs"; "" for none)
// is not met.
func Failed(r *status.Report, require string) (bool, error) {
	var lv status.Level
	switch require {
	case "":
		return r.HasProblems(), nil
	case "all":
		lv = r.Coverage.All
	case "dirs":
		lv = r.Coverage.Dirs
	default:
		return false, fmt.Errorf("invalid --require %q (use all or dirs)", require)
	}
	return r.HasProblems() || !lv.OK, nil
}

// realPath turns a real path given relative to where the operation runs into
// a path relative to the project root (with "/") and its names. A path outside
// the project is refused.
func (p *Project) realPath(path string) (string, []string, error) {
	abs := path
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(p.dir, path)
	}
	rel, err := filepath.Rel(p.Root, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", nil, fmt.Errorf("%s: outside the project", path)
	}
	names := splitRel(rel)
	return strings.Join(names, "/"), names, nil
}

// splitRel splits a clean relative path of the file system into names; "."
// is no name.
func splitRel(rel string) []string {
	if rel == "." || rel == "" {
		return nil
	}
	return strings.Split(filepath.ToSlash(rel), "/")
}

// dirPath is the path of a directory given as its names: "src/auth/", or ""
// for the project root.
func dirPath(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return strings.Join(names, "/") + "/"
}

// entryPathOf is the entry path of e, an entry under the directories names
// (given without their trailing "/").
func entryPathOf(names []string, e *structure.Entry) string {
	return dirPath(names) + e.Name
}

func notFound(path string, err error) error {
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%s: no such file", path)
	}
	return err
}

func ignored(path string) error {
	return fmt.Errorf("%s: ignored by .crube/ignore (use --no-ignore)", path)
}
