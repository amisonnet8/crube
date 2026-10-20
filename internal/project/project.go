// Package project finds the project root (the directory that holds .crube/)
// and is the only place that writes into .crube/.
//
// crube writes nothing outside .crube/ (docs/design/architecture.md "書き込みの
// 範囲"). Every write goes through WriteFile, CreateFile or MakeDir, which
// check right before writing that the target is a regular file directly in
// .crube/, and refuse when .crube or the file is a symbolic link (it could
// lead outside the project).
package project

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// DirName is the name of the directory that holds the structure files.
const DirName = ".crube"

// ErrNotFound is returned by FindRoot when no directory up from the start has
// a .crube/ directory.
var ErrNotFound = errors.New("no .crube/ found (run crube init)")

// GuardError is a write refused because its target is not a regular file
// directly in .crube/.
type GuardError struct {
	Path   string // relative to the project root, with "/" (".crube/structure.txt")
	Reason string
}

func (e *GuardError) Error() string {
	return fmt.Sprintf("refusing to write %s: %s", e.Path, e.Reason)
}

// FindRoot returns the project root: the first directory, from dir up to the
// root of the file system, that has a .crube/ directory. The result is an
// absolute, clean path (symbolic links in it are kept as they are).
func FindRoot(dir string) (string, error) {
	d, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		if fi, err := os.Stat(filepath.Join(d, DirName)); err == nil && fi.IsDir() {
			return d, nil
		}
		parent := filepath.Dir(d)
		if parent == d {
			return "", ErrNotFound
		}
		d = parent
	}
}

// Dir returns the path of .crube/ under root.
func Dir(root string) string {
	return filepath.Join(root, DirName)
}

// ReadFile reads the file name directly in .crube/ under root.
func ReadFile(root, name string) ([]byte, error) {
	if err := checkName(name); err != nil {
		return nil, err
	}
	return os.ReadFile(filepath.Join(Dir(root), name)) //nolint:gosec // name is checked by checkName to be a plain file name in .crube/
}

// MakeDir creates .crube/ under root. It fails if .crube already exists.
func MakeDir(root string) error {
	fi, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", root)
	}
	return os.Mkdir(Dir(root), 0o755) //nolint:gosec // .crube/ holds shared, non-secret data, like any directory of the project
}

// CreateFile creates the file name directly in .crube/ under root with data.
// It fails (with an error that matches fs.ErrExist) if the file exists.
func CreateFile(root, name string, data []byte) (err error) {
	dir, err := guard(root, name)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, name)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644) //nolint:gosec // path is checked by guard to be directly in .crube/
	if err != nil {
		return err
	}
	if _, err = f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return err
	}
	if err = f.Close(); err != nil {
		_ = os.Remove(path)
		return err
	}
	return nil
}

// WriteFile replaces the content of the file name directly in .crube/ under
// root with data, creating the file if needed. It writes a temporary file in
// .crube/ and renames it over the target, so a reader never sees half a file.
func WriteFile(root, name string, data []byte) error {
	dir, err := guard(root, name)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "."+name+".*.tmp")
	if err != nil {
		return err
	}
	// Close before rename or remove: Windows cannot do either on an open file.
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := os.Rename(tmp.Name(), filepath.Join(dir, name)); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	return nil
}

// guard checks, right before a write, that name can be written directly in
// .crube/ under root, and returns the path of .crube/.
func guard(root, name string) (string, error) {
	rel := DirName + "/" + name
	if err := checkName(name); err != nil {
		return "", &GuardError{Path: rel, Reason: err.Error()}
	}
	dir := Dir(root)
	fi, err := os.Lstat(dir)
	if err != nil {
		return "", &GuardError{Path: rel, Reason: "no " + DirName + "/ directory"}
	}
	if fi.Mode()&fs.ModeSymlink != 0 {
		return "", &GuardError{Path: rel, Reason: DirName + " is a symbolic link"}
	}
	if !fi.IsDir() {
		return "", &GuardError{Path: rel, Reason: DirName + " is not a directory"}
	}
	fi, err = os.Lstat(filepath.Join(dir, name))
	switch {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		return "", err
	case fi.Mode()&fs.ModeSymlink != 0:
		return "", &GuardError{Path: rel, Reason: "it is a symbolic link"}
	case !fi.Mode().IsRegular():
		return "", &GuardError{Path: rel, Reason: "it is not a regular file"}
	}
	return dir, nil
}

// checkName accepts only a plain file name: no directory part, not "." or "..".
func checkName(name string) error {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) || filepath.Base(name) != name {
		return fmt.Errorf("invalid file name %q", name)
	}
	return nil
}
