package core

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amisonnet8/crube/internal/structure"
)

// newProject makes a project with .crube/ holding the given files (name ->
// content) and returns it opened with the language tag.
func newProject(t *testing.T, tag string, files map[string]string) *Project {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".crube"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(root, ".crube", name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(Options{Dir: root, Lang: tag})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func content(t *testing.T, p *Project, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(p.Root, ".crube", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func ptr(s string) *string { return &s }

// The steps of docs/examples/getting-started.md, section 3.
func TestSetGettingStarted(t *testing.T) {
	p := newProject(t, "", map[string]string{"structure.txt": ""})
	steps := []struct {
		path, desc string
		detail     *string
	}{
		{"cmd/", "Executables", nil},
		{"cmd/myapp/", "CLI entry point of myapp", nil},
		{"internal/", "Private packages", nil},
		{"internal/store/", "Persistence layer (SQLite)", ptr("Also called: db, database")},
	}
	for _, s := range steps {
		if _, err := p.Set(s.path, s.desc, s.detail); err != nil {
			t.Fatalf("set %s: %v", s.path, err)
		}
	}
	want := "cmd/\tExecutables\n" +
		"\tmyapp/\tCLI entry point of myapp\n" +
		"internal/\tPrivate packages\n" +
		"\tstore/\tPersistence layer (SQLite)\n" +
		"\t\t| Also called: db, database\n"
	if got := content(t, p, "structure.txt"); got != want {
		t.Errorf("structure.txt:\ngot  %q\nwant %q", got, want)
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name   string
		before string
		path   string
		desc   string
		detail *string
		after  string
	}{
		{"missing parents become lines without a desc", "", "a/b/c.txt", "C", nil,
			"a/\n\tb/\n\t\tc.txt\tC\n"},
		{"a pattern", "data/\tData\n", "data/*.csv", "CSV", nil,
			"data/\tData\n\t*.csv\tCSV\n"},
		{"a new entry goes after its siblings", "a/\n\tz\tZ\n\tb/\n\t\tx\tX\n", "a/m", "M", nil,
			"a/\n\tz\tZ\n\tb/\n\t\tx\tX\n\tm\tM\n"},
		{"update keeps the place and the detail", "x\told\n\t| keep\ny\tY\n", "x", "new", nil,
			"x\tnew\n\t| keep\ny\tY\n"},
		{"update replaces the detail", "x\told\n\t| old\n", "x", "new", ptr("first\nsecond"),
			"x\tnew\n\t| first\n\t| second\n"},
		{"an empty detail removes it", "x\told\n\t| old\n", "x", "old", ptr(""),
			"x\told\n"},
		{"an empty desc leaves only structure", "x\told\n", "x", "", nil,
			"x\n"},
		{"desc is trimmed", "", "x", "  spaced  ", nil,
			"x\tspaced\n"},
		{"detail line breaks and trailing spaces", "", "x", "X", ptr("a  \r\n\r\nb\rc"),
			"x\tX\n\t| a\n\t|\n\t| b\n\t| c\n"},
		{"a file entry and a directory entry of one name", "a\tfile\n", "a/", "dir", nil,
			"a\tfile\na/\tdir\n"},
		{"the existing lines are rewritten in the canonical form", "a/  \n\n\tb\t\tB\n", "c", "C", nil,
			"a/\n\tb\tB\nc\tC\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newProject(t, "", map[string]string{"structure.txt": tt.before})
			r, err := p.Set(tt.path, tt.desc, tt.detail)
			if err != nil {
				t.Fatal(err)
			}
			if got := content(t, p, "structure.txt"); got != tt.after {
				t.Errorf("structure.txt:\ngot  %q\nwant %q", got, tt.after)
			}
			// The result is the entry as written.
			tree, _, err := p.Load()
			if err != nil {
				t.Fatal(err)
			}
			names, _ := splitEntryPath(tt.path)
			list, i := lookup(tree, names)
			if i < 0 {
				t.Fatalf("%s is not in the file", tt.path)
			}
			got := (*list)[i]
			if r.Path != tt.path || r.Desc != got.Desc || r.Detail != got.Detail {
				t.Errorf("returned %+v, the file has %+v", *r, *got)
			}
		})
	}
}
func TestSetRefused(t *testing.T) {
	before := "keep\tme\n"
	tests := []struct {
		name   string
		path   string
		desc   string
		detail *string
		msg    string
	}{
		{"desc with a line break", "x", "a\nb", nil, "desc must be one line (put more in detail)"},
		{"desc ending in a line break", "x", "a\n", nil, "desc must be one line (put more in detail)"},
		{"desc too wide", "x", strings.Repeat("あ", 129), nil, "desc is 258 columns wide (limit 256)"},
		{"detail too wide", "x", "X", ptr(strings.Repeat("a", 4097)), "detail is 4097 columns wide (limit 4096)"},
		{"empty path", "", "X", nil, `invalid entry path ""`},
		{"only a slash", "/", "X", nil, `invalid entry path "/"`},
		{"leading slash", "/a", "X", nil, `invalid entry path "/a"`},
		{"double slash", "a//b", "X", nil, `invalid entry path "a//b"`},
		{"name with a tab", "a\tb", "X", nil, `name "a\tb" cannot be written in a structure file`},
		{"name ending in a space", "a/b ", "X", nil, `name "b " cannot be written in a structure file`},
		{"name starting with a space", " a/", "X", nil, `name " a/" cannot be written in a structure file`},
		{"name like a detail line", "| x", "X", nil, `name "| x" cannot be written in a structure file`},
		{"dot dot", "../x", "X", nil, `name "../" cannot be written in a structure file`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newProject(t, "", map[string]string{"structure.txt": before})
			_, err := p.Set(tt.path, tt.desc, tt.detail)
			if err == nil || err.Error() != tt.msg {
				t.Fatalf("got %v, want %q", err, tt.msg)
			}
			if got := content(t, p, "structure.txt"); got != before {
				t.Errorf("the file was changed: %q", got)
			}
		})
	}
}

func TestLanguageFile(t *testing.T) {
	// With structure_ja.txt, CRUBE_LANG=ja reads and writes only it.
	p := newProject(t, "ja", map[string]string{"structure.txt": "x\tEnglish\n", "structure_ja.txt": "x\t日本語\n"})
	if _, err := p.Set("y", "追加", nil); err != nil {
		t.Fatal(err)
	}
	if got := content(t, p, "structure_ja.txt"); got != "x\t日本語\ny\t追加\n" {
		t.Errorf("structure_ja.txt = %q", got)
	}
	if got := content(t, p, "structure.txt"); got != "x\tEnglish\n" {
		t.Errorf("structure.txt was changed: %q", got)
	}

	// Without it, CRUBE_LANG=de uses structure.txt.
	p = newProject(t, "de", map[string]string{"structure.txt": ""})
	if _, err := p.Set("y", "Y", nil); err != nil {
		t.Fatal(err)
	}
	if got := content(t, p, "structure.txt"); got != "y\tY\n" {
		t.Errorf("structure.txt = %q", got)
	}
}

func TestParseErrorNamesTheFile(t *testing.T) {
	p := newProject(t, "ja", map[string]string{"structure.txt": "", "structure_ja.txt": "a/\n  b\n"})
	_, err := p.Set("c", "C", nil)
	var pe *structure.ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("got %v, want a *structure.ParseError", err)
	}
	if got, want := err.Error(), ".crube/structure_ja.txt:2: indentation must be tabs"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := content(t, p, "structure_ja.txt"); got != "a/\n  b\n" {
		t.Errorf("a file with an error was rewritten: %q", got)
	}
}

func TestMissingStructureFile(t *testing.T) {
	p := newProject(t, "", nil)
	_, err := p.Set("a", "A", nil)
	if err == nil || err.Error() != ".crube/structure.txt: no such file (run crube init)" {
		t.Fatalf("got %v", err)
	}
}

func TestRemove(t *testing.T) {
	p := newProject(t, "", map[string]string{"structure.txt": "src/\tS\n\tlegacy/\tL\n\t\told.go\tO\n\tmain.go\tM\ndata/\n\t*.csv\n"})
	// A real file must stay where it is.
	real := filepath.Join(p.Root, "src", "legacy", "old.go")
	if err := os.MkdirAll(filepath.Dir(real), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(real, []byte("package legacy\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	r, err := p.Remove("src/legacy/")
	if err != nil {
		t.Fatal(err)
	}
	// legacy/ and old.go below it.
	if want := (RemoveResult{Path: "src/legacy/", Removed: 2}); *r != want {
		t.Errorf("returned %+v, want %+v", *r, want)
	}
	if got, want := content(t, p, "structure.txt"), "src/\tS\n\tmain.go\tM\ndata/\n\t*.csv\n"; got != want {
		t.Errorf("structure.txt:\ngot  %q\nwant %q", got, want)
	}
	if b, err := os.ReadFile(real); err != nil || string(b) != "package legacy\n" {
		t.Errorf("the real file was touched: %q, %v", b, err)
	}
}

func TestRemoveRefused(t *testing.T) {
	before := "src/\tS\n\tmain.go\tM\n"
	for path, msg := range map[string]string{
		"src/other.go": "no entry src/other.go",
		"src":          "no entry src", // the entry is src/, a directory entry
		"*.go":         "no entry *.go",
		"src/*.go":     "no entry src/*.go", // not matched as a pattern
	} {
		p := newProject(t, "", map[string]string{"structure.txt": before})
		if _, err := p.Remove(path); err == nil || err.Error() != msg {
			t.Errorf("rm %s: got %v, want %q", path, err, msg)
		}
		if got := content(t, p, "structure.txt"); got != before {
			t.Errorf("rm %s changed the file: %q", path, got)
		}
	}
}

func TestMove(t *testing.T) {
	tests := []struct {
		name     string
		before   string
		from, to string
		after    string
		moved    int // the entry and those below it
	}{
		{"a directory with what is below it, after its new siblings",
			"src/\tS\n\tlegacy/\tL\n\t\told.go\tO\n\t\t*_test.go\tT\nlib/\tLib\n\tx\tX\n",
			"src/legacy/", "lib/old/",
			"src/\tS\nlib/\tLib\n\tx\tX\n\told/\tL\n\t\told.go\tO\n\t\t*_test.go\tT\n", 3},
		{"missing parents of the destination are added",
			"a.go\tA\n", "a.go", "cmd/tool/a.go",
			"cmd/\n\ttool/\n\t\ta.go\tA\n", 1},
		{"rename in place goes to the end of the siblings",
			"a\tA\nb\tB\n", "a", "c",
			"b\tB\nc\tA\n", 1},
		{"a file into a new directory of the same name",
			"a\tA\n", "a", "a/a",
			"a/\n\ta\tA\n", 1},
		{"up to the root",
			"x/\n\ty/\tY\n\t\tz\tZ\n", "x/y/", "y/",
			"x/\ny/\tY\n\tz\tZ\n", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newProject(t, "", map[string]string{"structure.txt": tt.before})
			r, err := p.Move(tt.from, tt.to)
			if err != nil {
				t.Fatal(err)
			}
			if got := content(t, p, "structure.txt"); got != tt.after {
				t.Errorf("structure.txt:\ngot  %q\nwant %q", got, tt.after)
			}
			if want := (MoveResult{From: tt.from, To: tt.to, Moved: tt.moved}); *r != want {
				t.Errorf("returned %+v, want %+v", *r, want)
			}
		})
	}
}

func TestMoveRefused(t *testing.T) {
	before := "src/\tS\n\tlegacy/\tL\n\t\told.go\tO\n\tmain.go\tM\nlib/\n"
	tests := []struct{ from, to, msg string }{
		{"src/none.go", "lib/none.go", "no entry src/none.go"},
		{"src/main.go", "src/legacy/old.go", "entry src/legacy/old.go already exists"},
		{"src/main.go", "src/main.go", "entry src/main.go already exists"},
		{"src/legacy/", "lib/", "entry lib/ already exists"},
		{"src/legacy/", "lib/legacy", "cannot move src/legacy/ to lib/legacy: a directory entry (ending in /) stays a directory entry, and a file entry a file entry"},
		{"src/main.go", "lib/main.go/", "cannot move src/main.go to lib/main.go/: a directory entry (ending in /) stays a directory entry, and a file entry a file entry"},
		{"src/", "src/legacy/new/", "cannot move src/ into itself"},
		{"src/legacy/", "src/legacy/x/", "cannot move src/legacy/ into itself"},
		{"src/main.go", "lib/a\tb", `name "a\tb" cannot be written in a structure file`},
	}
	for _, tt := range tests {
		p := newProject(t, "", map[string]string{"structure.txt": before})
		if _, err := p.Move(tt.from, tt.to); err == nil || err.Error() != tt.msg {
			t.Errorf("mv %s %s: got %v, want %q", tt.from, tt.to, err, tt.msg)
		}
		if got := content(t, p, "structure.txt"); got != before {
			t.Errorf("mv %s %s changed the file: %q", tt.from, tt.to, got)
		}
	}
}

func TestEntryPathBackslash(t *testing.T) {
	t.Run("Windows: \\ separates names", func(t *testing.T) {
		old := osSeparator
		osSeparator = '\\'
		t.Cleanup(func() { osSeparator = old })

		p := newProject(t, "", map[string]string{"structure.txt": ""})
		set, err := p.Set(`src\auth\`, "Auth", nil)
		if err != nil {
			t.Fatal(err)
		}
		if set.Path != "src/auth/" {
			t.Errorf("set returned %q", set.Path)
		}
		moved, err := p.Move(`src\auth\`, `lib\auth\`)
		if err != nil {
			t.Fatal(err)
		}
		if moved.From != "src/auth/" || moved.To != "lib/auth/" {
			t.Errorf("mv returned %+v", *moved)
		}
		tree, err := p.Tree(TreeOptions{Path: `lib\`})
		if err != nil {
			t.Fatal(err)
		}
		if len(tree.Entries) != 1 || tree.Entries[0].Path != "lib/" {
			t.Errorf("tree returned %+v", tree.Entries)
		}
		if got, want := content(t, p, "structure.txt"), "src/\nlib/\n\tauth/\tAuth\n"; got != want {
			t.Errorf("structure.txt: got %q, want %q", got, want)
		}
		removed, err := p.Remove(`lib\auth\`)
		if err != nil {
			t.Fatal(err)
		}
		if removed.Path != "lib/auth/" {
			t.Errorf("rm returned %+v", *removed)
		}
	})

	t.Run("elsewhere: \\ is a character of a name", func(t *testing.T) {
		old := osSeparator
		osSeparator = '/'
		t.Cleanup(func() { osSeparator = old })

		p := newProject(t, "", map[string]string{"structure.txt": ""})
		if _, err := p.Set(`a\b`, "X", nil); err != nil {
			t.Fatal(err)
		}
		if got, want := content(t, p, "structure.txt"), "a\\b\tX\n"; got != want {
			t.Errorf("structure.txt: got %q, want %q", got, want)
		}
	})
}
