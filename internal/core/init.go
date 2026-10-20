package core

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/amisonnet8/crube/internal/ignore"
	"github.com/amisonnet8/crube/internal/lang"
	"github.com/amisonnet8/crube/internal/project"
)

// Init makes what crube needs, and returns the paths it made (".crube/…") in
// the order they were made.
//
// If no directory from dir up has .crube/, it makes .crube/ in dir with an
// empty structure.txt, an empty file for the language if it is not English,
// and the default ignore. If there is a .crube/ already, it makes only the
// file of the language, if that is missing. It fails when there is nothing to
// make. Files that exist are never overwritten.
func Init(dir, tag string) ([]string, error) {
	name, err := lang.FileName(tag)
	if err != nil {
		return nil, err
	}
	if dir == "" {
		dir = "."
	}

	root, err := project.FindRoot(dir)
	switch {
	case err == nil:
		if err := project.CreateFile(root, name, nil); err != nil {
			if errors.Is(err, fs.ErrExist) {
				return nil, fmt.Errorf("nothing to create: %s already exists", relPath(name))
			}
			return nil, err
		}
		return []string{relPath(name)}, nil
	case !errors.Is(err, project.ErrNotFound):
		return nil, err
	}

	root, err = filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if err := project.MakeDir(root); err != nil {
		return nil, err
	}
	type file struct{ name, data string }
	files := []file{{lang.DefaultFile, ""}}
	if name != lang.DefaultFile {
		files = append(files, file{name, ""})
	}
	files = append(files, file{ignore.FileName, ignore.Default})

	var made []string
	for _, f := range files {
		if err := project.CreateFile(root, f.name, []byte(f.data)); err != nil {
			return made, err
		}
		made = append(made, relPath(f.name))
	}
	return made, nil
}
