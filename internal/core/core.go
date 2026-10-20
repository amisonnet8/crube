// Package core is what the CLI and the MCP server call: each operation of
// crube (init, ls, tree, find, show, status, set, rm, mv) as one function,
// built from the other packages of internal/. The CLI and the MCP server only
// turn arguments into calls and results into output, so that the two never
// behave differently (docs/design/architecture.md).
package core

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/amisonnet8/crube/internal/ignore"
	"github.com/amisonnet8/crube/internal/lang"
	"github.com/amisonnet8/crube/internal/project"
	"github.com/amisonnet8/crube/internal/structure"
)

// Options are what every operation on a project starts from.
type Options struct {
	Dir      string // where to start looking for .crube/ (-C); "" is the current directory
	Lang     string // the value of CRUBE_LANG
	NoIgnore bool   // --no-ignore
}

// Project is a project found from Options.
type Project struct {
	Root     string // the project root, absolute
	dir      string // where the operation runs (o.Dir), absolute: real paths are relative to it
	lang     string
	noIgnore bool
}

// Open finds the project that holds o.Dir, and checks the language.
func Open(o Options) (*Project, error) {
	if err := lang.Check(o.Lang); err != nil {
		return nil, err
	}
	dir := o.Dir
	if dir == "" {
		dir = "."
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	root, err := project.FindRoot(dir)
	if err != nil {
		return nil, err
	}
	return &Project{Root: root, dir: dir, lang: o.Lang, noIgnore: o.NoIgnore}, nil
}

// Ignore returns the matcher of .crube/ignore: with --no-ignore, or without
// the file, only .crube/ is left out.
func (p *Project) Ignore() (*ignore.Matcher, error) {
	if p.noIgnore {
		return ignore.None(), nil
	}
	data, err := project.ReadFile(p.Root, ignore.FileName)
	if errors.Is(err, fs.ErrNotExist) {
		return ignore.Parse(nil), nil
	}
	if err != nil {
		return nil, err
	}
	return ignore.Parse(data), nil
}

// File returns the name, in .crube/, of the structure file that the
// operations read and write for the language.
func (p *Project) File() (string, error) {
	return lang.Select(project.Dir(p.Root), p.lang)
}

// Load reads and parses the structure file of the language. It returns the
// tree and the name of the file in .crube/.
func (p *Project) Load() (*structure.Tree, string, error) {
	name, err := p.File()
	if err != nil {
		return nil, "", err
	}
	data, err := project.ReadFile(p.Root, name)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, "", fmt.Errorf("%s: no such file (run crube init)", relPath(name))
	}
	if err != nil {
		return nil, "", err
	}
	tree, err := structure.Parse(data)
	if err != nil {
		var pe *structure.ParseError
		if errors.As(err, &pe) {
			pe.File = relPath(name)
		}
		return nil, "", err
	}
	return tree, name, nil
}

// update loads the structure file, lets change modify the tree, and writes
// the tree back in the canonical form, through the write guard of project.
// Nothing is written if change fails.
func (p *Project) update(change func(*structure.Tree) error) error {
	tree, name, err := p.Load()
	if err != nil {
		return err
	}
	if err := change(tree); err != nil {
		return err
	}
	return project.WriteFile(p.Root, name, tree.Format())
}

// relPath is the path of a file in .crube/ as shown to people.
func relPath(name string) string {
	return project.DirName + "/" + name
}
