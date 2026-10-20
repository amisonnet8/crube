package core

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/amisonnet8/crube/internal/ignore"
	"github.com/amisonnet8/crube/internal/project"
)

func readAll(t *testing.T, dir string) map[string]string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	m := map[string]string{}
	for _, e := range entries {
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		m[e.Name()] = string(b)
	}
	return m
}

func TestInitNewProject(t *testing.T) {
	root := t.TempDir()
	made, err := Init(root, "")
	if err != nil {
		t.Fatal(err)
	}
	// The order of docs/examples/getting-started.md.
	if want := []string{".crube/structure.txt", ".crube/ignore"}; !reflect.DeepEqual(made, want) {
		t.Errorf("made %q, want %q", made, want)
	}
	want := map[string]string{"structure.txt": "", "ignore": ignore.Default}
	if got := readAll(t, filepath.Join(root, ".crube")); !reflect.DeepEqual(got, want) {
		t.Errorf(".crube/ holds %q", got)
	}
}

func TestInitNewProjectWithLanguage(t *testing.T) {
	root := t.TempDir()
	made, err := Init(root, "ja")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{".crube/structure.txt", ".crube/structure_ja.txt", ".crube/ignore"}; !reflect.DeepEqual(made, want) {
		t.Errorf("made %q, want %q", made, want)
	}
}

func TestInitAddsLanguageFile(t *testing.T) {
	root := t.TempDir()
	if _, err := Init(root, ""); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".crube", "ignore"), []byte("custom\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(root, "src", "auth")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	// From a subdirectory: the existing .crube/ above is used.
	made, err := Init(sub, "ja")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{".crube/structure_ja.txt"}; !reflect.DeepEqual(made, want) {
		t.Errorf("made %q, want %q", made, want)
	}
	if _, err := os.Stat(filepath.Join(sub, ".crube")); !errors.Is(err, os.ErrNotExist) {
		t.Error("a nested .crube/ was made")
	}
	want := map[string]string{"structure.txt": "", "structure_ja.txt": "", "ignore": "custom\n"}
	if got := readAll(t, filepath.Join(root, ".crube")); !reflect.DeepEqual(got, want) {
		t.Errorf(".crube/ holds %q", got)
	}
}

func TestInitNothingToCreate(t *testing.T) {
	root := t.TempDir()
	if _, err := Init(root, "ja"); err != nil {
		t.Fatal(err)
	}
	for tag, msg := range map[string]string{
		"":   "nothing to create: .crube/structure.txt already exists",
		"en": "nothing to create: .crube/structure.txt already exists",
		"ja": "nothing to create: .crube/structure_ja.txt already exists",
	} {
		if _, err := Init(root, tag); err == nil || err.Error() != msg {
			t.Errorf("init with %q: got %v, want %q", tag, err, msg)
		}
	}
}

func TestInitRefusesBadLanguage(t *testing.T) {
	root := t.TempDir()
	if _, err := Init(root, "../x"); err == nil {
		t.Fatal("init succeeded")
	}
	if entries, _ := os.ReadDir(root); len(entries) != 0 {
		t.Errorf("made %v", entries)
	}
}

func TestOpen(t *testing.T) {
	root := t.TempDir()
	if _, err := Open(Options{Dir: root}); !errors.Is(err, project.ErrNotFound) {
		t.Errorf("Open without .crube/: %v", err)
	}
	if _, err := Init(root, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(Options{Dir: root, Lang: "a/b"}); err == nil {
		t.Error("Open accepted a bad CRUBE_LANG")
	}
	p, err := Open(Options{Dir: root})
	if err != nil {
		t.Fatal(err)
	}
	if want, _ := filepath.Abs(root); p.Root != want {
		t.Errorf("Root = %s, want %s", p.Root, want)
	}
}
