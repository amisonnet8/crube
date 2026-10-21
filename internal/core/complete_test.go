package core

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestComplete(t *testing.T) {
	root := readProject(t)
	p := open(t, Options{Dir: root})
	tests := []struct {
		typed string
		real  bool
		want  []Candidate
	}{
		// build/ is ignored; .crube/ is always left out.
		{"", false, []Candidate{{"src/", "Source"}, {"data/", ""}, {"go.mod", "Module"}}},
		{"", true, []Candidate{{"src/", "Source"}, {"data/", ""}, {"go.mod", "Module"}, {"README.md", ""}, {"empty/", ""}}},
		{"d", false, []Candidate{{"data/", ""}}},
		{"src/", false, []Candidate{{"src/auth/", "Authentication"}}},
		{"src/", true, []Candidate{{"src/auth/", "Authentication"}, {"src/db/", ""}}},
		{"src/auth/s", false, []Candidate{{"src/auth/session.go", "Sessions"}}},
		// login_test.go is matched by *_test.go, but has no entry of its own.
		{"src/auth/", true, []Candidate{{"src/auth/session.go", "Sessions"}, {"src/auth/*_test.go", "Tests"}, {"src/auth/login_test.go", ""}}},
		{"src/db/", false, nil},
		{"src/db/", true, []Candidate{{"src/db/conn.go", ""}}},
		{"build/", true, nil},
		{"nope/", true, nil},
		{"go.mod/", true, nil},
	}
	for _, tt := range tests {
		got, err := p.Complete(tt.typed, tt.real)
		if err != nil {
			t.Errorf("Complete(%q, %v): %v", tt.typed, tt.real, err)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Complete(%q, %v) = %q, want %q", tt.typed, tt.real, got, tt.want)
		}
	}

	if _, err := p.Complete("a//", false); err == nil {
		t.Error(`Complete("a//") did not fail`)
	}
}

func TestCompleteNoIgnore(t *testing.T) {
	root := readProject(t)
	p := open(t, Options{Dir: root, NoIgnore: true})
	got, err := p.Complete("", true)
	if err != nil {
		t.Fatal(err)
	}
	want := []Candidate{{"src/", "Source"}, {"data/", ""}, {"build/", "Build output"}, {"go.mod", "Module"}, {"README.md", ""}, {"empty/", ""}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

// A real directory reached through a symbolic link is not looked into: what
// is offered stays in the project.
func TestCompleteSymlink(t *testing.T) {
	root := readProject(t)
	if err := os.Symlink("src", filepath.Join(root, "lnk")); err != nil {
		t.Skip("cannot make a symbolic link:", err)
	}
	p := open(t, Options{Dir: root})
	got, err := p.Complete("l", true)
	if err != nil {
		t.Fatal(err)
	}
	if want := []Candidate{{"lnk", ""}}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, err := p.Complete("lnk/", true); err != nil || got != nil {
		t.Errorf(`Complete("lnk/") = %q, %v; want nothing`, got, err)
	}
}

// On Windows an entry path may be typed with "\"; the candidates have "/".
func TestCompleteBackslash(t *testing.T) {
	old := osSeparator
	osSeparator = '\\'
	t.Cleanup(func() { osSeparator = old })
	p := open(t, Options{Dir: readProject(t)})
	got, err := p.Complete(`src\`, false)
	if err != nil {
		t.Fatal(err)
	}
	if want := []Candidate{{"src/auth/", "Authentication"}}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}
