package project

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// tempDir returns a new temporary directory with symbolic links resolved, so
// that paths compare equal on macOS (/var -> /private/var) and Windows (8.3
// names).
func tempDir(t *testing.T) string {
	t.Helper()
	d, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// symlink makes a symbolic link, or skips the test where that is not allowed
// (Windows without the privilege).
func symlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("cannot make a symbolic link here: %v", err)
	}
}

func names(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var ns []string
	for _, e := range entries {
		ns = append(ns, e.Name())
	}
	return ns
}

func TestFindRoot(t *testing.T) {
	root := tempDir(t)
	mkdir(t, filepath.Join(root, ".crube"))
	mkdir(t, filepath.Join(root, "src", "auth"))

	for _, start := range []string{
		root,
		filepath.Join(root, "src"),
		filepath.Join(root, "src", "auth"),
		filepath.Join(root, ".crube"),
	} {
		got, err := FindRoot(start)
		if err != nil {
			t.Fatalf("FindRoot(%s): %v", start, err)
		}
		if got != root {
			t.Errorf("FindRoot(%s) = %s, want %s", start, got, root)
		}
	}
}

func TestFindRootRelative(t *testing.T) {
	root := tempDir(t)
	mkdir(t, filepath.Join(root, ".crube"))
	mkdir(t, filepath.Join(root, "src"))
	t.Chdir(filepath.Join(root, "src"))

	for _, start := range []string{".", "..", "."} {
		got, err := FindRoot(start)
		if err != nil {
			t.Fatalf("FindRoot(%q): %v", start, err)
		}
		if got != root {
			t.Errorf("FindRoot(%q) = %s, want %s", start, got, root)
		}
	}
}

func TestFindRootNearest(t *testing.T) {
	outer := tempDir(t)
	inner := filepath.Join(outer, "sub")
	mkdir(t, filepath.Join(outer, ".crube"))
	mkdir(t, filepath.Join(inner, ".crube"))
	mkdir(t, filepath.Join(inner, "x"))

	got, err := FindRoot(filepath.Join(inner, "x"))
	if err != nil {
		t.Fatal(err)
	}
	if got != inner {
		t.Errorf("FindRoot = %s, want the nearest %s", got, inner)
	}
}

func TestFindRootSkipsFileNamedCrube(t *testing.T) {
	outer := tempDir(t)
	inner := filepath.Join(outer, "sub")
	mkdir(t, filepath.Join(outer, ".crube"))
	mkdir(t, inner)
	writeFile(t, filepath.Join(inner, ".crube"), "not a directory")

	got, err := FindRoot(inner)
	if err != nil {
		t.Fatal(err)
	}
	if got != outer {
		t.Errorf("FindRoot = %s, want %s (a file named .crube is not a project)", got, outer)
	}
}

func TestFindRootNotFound(t *testing.T) {
	dir := tempDir(t)
	if _, err := os.Stat(filepath.Join(filepath.Dir(dir), ".crube")); err == nil {
		t.Skip("a directory above the temporary directory has .crube/")
	}
	_, err := FindRoot(dir)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("FindRoot = %v, want ErrNotFound", err)
	}
	if got, want := err.Error(), "no .crube/ found (run crube init)"; got != want {
		t.Errorf("message %q, want %q", got, want)
	}
}

func TestWriteFile(t *testing.T) {
	root := tempDir(t)
	mkdir(t, filepath.Join(root, ".crube"))
	path := filepath.Join(root, ".crube", "structure.txt")

	if err := WriteFile(root, "structure.txt", []byte("a\tfirst\n")); err != nil {
		t.Fatalf("create: %v", err)
	}
	if got := readFile(t, path); got != "a\tfirst\n" {
		t.Errorf("after create: %q", got)
	}
	if err := WriteFile(root, "structure.txt", []byte("b\n")); err != nil {
		t.Fatalf("replace: %v", err)
	}
	if got := readFile(t, path); got != "b\n" {
		t.Errorf("after replace: %q", got)
	}
	// No temporary file is left behind.
	if got, want := names(t, filepath.Join(root, ".crube")), []string{"structure.txt"}; !reflect.DeepEqual(got, want) {
		t.Errorf(".crube/ holds %q, want %q", got, want)
	}
	got, err := ReadFile(root, "structure.txt")
	if err != nil || string(got) != "b\n" {
		t.Errorf("ReadFile = %q, %v", got, err)
	}
}

func TestCreateFile(t *testing.T) {
	root := tempDir(t)
	mkdir(t, filepath.Join(root, ".crube"))

	if err := CreateFile(root, "ignore", []byte(".git/\n")); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(root, ".crube", "ignore")); got != ".git/\n" {
		t.Errorf("content %q", got)
	}
	err := CreateFile(root, "ignore", []byte("other\n"))
	if !errors.Is(err, fs.ErrExist) {
		t.Fatalf("second CreateFile = %v, want fs.ErrExist", err)
	}
	if got := readFile(t, filepath.Join(root, ".crube", "ignore")); got != ".git/\n" {
		t.Errorf("an existing file was changed: %q", got)
	}
}

func TestMakeDir(t *testing.T) {
	root := tempDir(t)
	if err := MakeDir(root); err != nil {
		t.Fatal(err)
	}
	if fi, err := os.Stat(filepath.Join(root, ".crube")); err != nil || !fi.IsDir() {
		t.Fatalf(".crube/ was not made: %v", err)
	}
	if err := MakeDir(root); !errors.Is(err, fs.ErrExist) {
		t.Errorf("second MakeDir = %v, want fs.ErrExist", err)
	}
	if err := MakeDir(filepath.Join(root, "missing")); err == nil {
		t.Error("MakeDir under a missing directory succeeded")
	}
}

// ReadFile reads only directly in .crube/: a name that leads elsewhere is
// refused before anything is read.
func TestReadFileRefusesBadNames(t *testing.T) {
	root := tempDir(t)
	mkdir(t, filepath.Join(root, ".crube", "sub"))
	writeFile(t, filepath.Join(root, "secret.txt"), "outside")
	writeFile(t, filepath.Join(root, ".crube", "sub", "x"), "below")
	for _, name := range []string{"../secret.txt", "sub/x", "..", ".", ""} {
		b, err := ReadFile(root, name)
		if err == nil || err.Error() != fmt.Sprintf("invalid file name %q", name) {
			t.Errorf("%q: got %q, %v; want the error for the name", name, b, err)
		}
	}
}

// Every refused write must leave the file system as it was.
func TestWriteRefused(t *testing.T) {
	type write func(root, name string, data []byte) error
	writes := map[string]write{"WriteFile": WriteFile, "CreateFile": CreateFile}

	for wname, w := range writes {
		t.Run(wname, func(t *testing.T) {
			t.Run("bad names", func(t *testing.T) {
				root := tempDir(t)
				mkdir(t, filepath.Join(root, ".crube"))
				for _, name := range []string{"", ".", "..", "../x", "sub/x", `sub\x`, "/abs"} {
					err := w(root, name, []byte("x"))
					var ge *GuardError
					if !errors.As(err, &ge) || ge.Reason != fmt.Sprintf("invalid file name %q", name) {
						t.Errorf("%q: got %v, want a *GuardError for the name", name, err)
					}
				}
				if got := names(t, root); !reflect.DeepEqual(got, []string{".crube"}) {
					t.Errorf("root holds %q", got)
				}
				if got := names(t, filepath.Join(root, ".crube")); len(got) != 0 {
					t.Errorf(".crube/ holds %q", got)
				}
			})

			t.Run("no .crube", func(t *testing.T) {
				root := tempDir(t)
				err := w(root, "structure.txt", []byte("x"))
				assertGuard(t, err, ".crube/structure.txt", "no .crube/ directory")
				if got := names(t, root); len(got) != 0 {
					t.Errorf("root holds %q", got)
				}
			})

			t.Run(".crube is a file", func(t *testing.T) {
				root := tempDir(t)
				writeFile(t, filepath.Join(root, ".crube"), "file")
				err := w(root, "structure.txt", []byte("x"))
				assertGuard(t, err, ".crube/structure.txt", ".crube is not a directory")
			})

			t.Run(".crube is a symbolic link to outside", func(t *testing.T) {
				root := tempDir(t)
				outside := tempDir(t)
				symlink(t, outside, filepath.Join(root, ".crube"))
				err := w(root, "structure.txt", []byte("x"))
				assertGuard(t, err, ".crube/structure.txt", ".crube is a symbolic link")
				if got := names(t, outside); len(got) != 0 {
					t.Errorf("wrote outside the project: %q", got)
				}
			})

			t.Run("target is a symbolic link to outside", func(t *testing.T) {
				root := tempDir(t)
				mkdir(t, filepath.Join(root, ".crube"))
				outside := filepath.Join(tempDir(t), "secret.txt")
				writeFile(t, outside, "keep")
				symlink(t, outside, filepath.Join(root, ".crube", "structure.txt"))
				err := w(root, "structure.txt", []byte("x"))
				assertGuard(t, err, ".crube/structure.txt", "it is a symbolic link")
				if got := readFile(t, outside); got != "keep" {
					t.Errorf("the file outside was changed: %q", got)
				}
			})

			t.Run("target is a dangling symbolic link", func(t *testing.T) {
				root := tempDir(t)
				mkdir(t, filepath.Join(root, ".crube"))
				outside := filepath.Join(tempDir(t), "new.txt")
				symlink(t, outside, filepath.Join(root, ".crube", "structure.txt"))
				err := w(root, "structure.txt", []byte("x"))
				assertGuard(t, err, ".crube/structure.txt", "it is a symbolic link")
				if _, err := os.Lstat(outside); !errors.Is(err, fs.ErrNotExist) {
					t.Errorf("a file was made outside the project: %v", err)
				}
			})

			t.Run("target is a directory", func(t *testing.T) {
				root := tempDir(t)
				mkdir(t, filepath.Join(root, ".crube", "structure.txt"))
				err := w(root, "structure.txt", []byte("x"))
				assertGuard(t, err, ".crube/structure.txt", "it is not a regular file")
			})
		})
	}
}

func assertGuard(t *testing.T, err error, path, reason string) {
	t.Helper()
	var ge *GuardError
	if !errors.As(err, &ge) {
		t.Fatalf("got %v, want a *GuardError", err)
	}
	if ge.Path != path || ge.Reason != reason {
		t.Errorf("got %+v, want path %q reason %q", *ge, path, reason)
	}
}

func TestGuardErrorMessage(t *testing.T) {
	err := &GuardError{Path: ".crube/structure.txt", Reason: ".crube is a symbolic link"}
	if got, want := err.Error(), "refusing to write .crube/structure.txt: .crube is a symbolic link"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
