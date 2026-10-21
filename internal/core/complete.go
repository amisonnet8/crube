package core

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Candidate is an entry path offered to complete the one being typed, with
// the desc of its entry ("" for a real path that has no entry).
type Candidate struct {
	Path string
	Desc string
}

// Complete lists what an entry path typed so far can go on with, one level at
// a time: the entries directly in the directory entry that typed names up to
// its last "/", whose names start with what follows it. Ignored entries are
// left out, as Tree leaves them out (and so is all that is below an ignored
// directory: Ignored looks at the directories above a path). With real (for
// set), the names that are really there and that no entry has exactly are
// offered too, after the entries; ignored ones are left out. The paths are
// from the project root.
func (p *Project) Complete(typed string, real bool) ([]Candidate, error) {
	typed = normEntryPath(typed)
	cut := strings.LastIndex(typed, "/") + 1
	dir, prefix := typed[:cut], typed[cut:]
	var names []string // the entry names of dir, each ending in "/"
	if dir != "" {
		var err error
		if names, err = splitEntryPath(dir); err != nil {
			return nil, err
		}
	}
	plain := make([]string, len(names)) // the same, without their "/"
	for i, n := range names {
		plain[i] = strings.TrimSuffix(n, "/")
	}

	tree, _, err := p.Load()
	if err != nil {
		return nil, err
	}
	ign, err := p.Ignore()
	if err != nil {
		return nil, err
	}

	entries := tree.Entries
	if len(names) > 0 {
		entries = nil
		if list, i := lookup(tree, names); i >= 0 {
			entries = (*list)[i].Children
		}
	}
	var list []Candidate
	for _, e := range entries {
		if !strings.HasPrefix(e.Name, prefix) || ign.Ignored(child(plain, e.Name), e.IsDir()) {
			continue
		}
		list = append(list, Candidate{Path: dir + e.Name, Desc: e.Desc})
	}
	if !real {
		return list, nil
	}

	reals, err := p.realNames(plain)
	if err != nil {
		return nil, err
	}
	for _, name := range reals {
		if !strings.HasPrefix(name, prefix) || index(entries, name) >= 0 ||
			ign.Ignored(child(plain, name), strings.HasSuffix(name, "/")) {
			continue
		}
		list = append(list, Candidate{Path: dir + name})
	}
	return list, nil
}

// child is the names of name (an entry name, or a real name with "/" for a
// directory) under the directories plain.
func child(plain []string, name string) []string {
	return append(plain[:len(plain):len(plain)], strings.TrimSuffix(name, "/"))
}

// realNames lists, in name order, what is directly in the directory plain
// (names from the project root), with "/" after a directory. A directory that
// is not there, or that is reached through a symbolic link, has nothing: what
// is offered stays in the project.
func (p *Project) realNames(plain []string) ([]string, error) {
	path := p.Root
	for _, n := range plain {
		path = filepath.Join(path, n)
		fi, err := os.Lstat(path)
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		if !fi.IsDir() {
			return nil, nil
		}
	}
	list, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(list))
	for _, de := range list {
		name := de.Name()
		if de.IsDir() && de.Type()&fs.ModeSymlink == 0 {
			name += "/"
		}
		names = append(names, name)
	}
	return names, nil
}
