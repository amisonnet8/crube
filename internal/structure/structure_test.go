package structure

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseBuildsTree(t *testing.T) {
	in := "src/\tSource code\n" +
		"\tauth/\tAuthentication\n" +
		"\t\t| Entry point is index.ts.\n" +
		"\t\t|\n" +
		"\t\t| Settings are in config/auth.yaml.\n" +
		"\t\tsession.ts\tSession storage\n" +
		"\t\t*_test.go\tTests\n" +
		"\tmy  file.txt\tA name with spaces\n" +
		"data/\n" +
		"\t*.csv\tCSV data\n"
	want := &Tree{Entries: []*Entry{
		{Name: "src/", Desc: "Source code", Children: []*Entry{
			{
				Name:   "auth/",
				Desc:   "Authentication",
				Detail: "Entry point is index.ts.\n\nSettings are in config/auth.yaml.",
				Children: []*Entry{
					{Name: "session.ts", Desc: "Session storage"},
					{Name: "*_test.go", Desc: "Tests"},
				},
			},
			{Name: "my  file.txt", Desc: "A name with spaces"},
		}},
		{Name: "data/", Children: []*Entry{
			{Name: "*.csv", Desc: "CSV data"},
		}},
	}}

	got, err := Parse([]byte(in))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse:\ngot  %s\nwant %s", dump(got), dump(want))
	}
	if out := string(got.Format()); out != in {
		t.Errorf("Format does not give back the canonical input:\ngot  %q\nwant %q", out, in)
	}
}

func TestParseNormalizes(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string // canonical form
	}{
		{"empty file", "", ""},
		{"only blank lines", "\n\n\t\n", ""},
		{"no newline at the end", "a\tx", "a\tx\n"},
		{"CRLF", "a/\tx\r\n\tb\ty\r\n\t\t| d\r\n", "a/\tx\n\tb\ty\n\t\t| d\n"},
		{"blank lines between entries", "a/\n\n\tb\n\n\t\t| d\n\nc\n", "a/\n\tb\n\t\t| d\nc\n"},
		{"trailing white space", "a/\t x \t\n\tb  \n\t\t| d  \n", "a/\tx\n\tb\n\t\t| d\n"},
		{"several tabs between name and desc", "a\t\t\tx", "a\tx\n"},
		{"tab inside desc is kept", "a\tx\ty", "a\tx\ty\n"},
		{"space before the desc", "a\t   x", "a\tx\n"},
		{"empty detail lines", "a\n\t|\n\t| x\n\t|\n", "a\n\t|\n\t| x\n\t|\n"},
		{"detail text keeps inner spaces", "a\n\t|   x  y\n", "a\n\t|   x  y\n"},
		{"pipe not followed by a space is a name", "|x\tdesc", "|x\tdesc\n"},
		{"same name as file and as directory", "a\na/\n", "a\na/\n"},
		{"same name under different parents", "a/\n\tx\nb/\n\tx\n", "a/\n\tx\nb/\n\tx\n"},
		{"back to a shallower depth", "a/\n\tb/\n\t\tc/\n\t\t\td\ne\n", "a/\n\tb/\n\t\tc/\n\t\t\td\ne\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := Parse([]byte(tt.in))
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if got := string(tree.Format()); got != tt.want {
				t.Errorf("Format:\ngot  %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name string
		in   string
		line int
		msg  string
	}{
		{"spaces as indentation", "a/\n  b\n", 2, "indentation must be tabs"},
		{"space after a tab", "a/\n\t b\n", 2, "indentation must be tabs"},
		{"space before a detail line", "a\n    | d\n", 2, "indentation must be tabs"},
		{"two levels deeper", "a/\n\t\tb\n", 2, "indented more than one level deeper than the entry above"},
		{"first line indented", "\ta\n", 1, "indented more than one level deeper than the entry above"},
		{"child of a file", "a/\n\tb.txt\n\t\tc\n", 3, `file entry "b.txt" cannot have children (directory names end in /)`},
		{"duplicate at the root", "a\tx\nb\na\ty\n", 3, `duplicate entry "a" (first on line 1)`},
		{"duplicate after other children", "d/\n\tx/\n\t\ty\n\tx/\n", 4, `duplicate entry "x/" (first on line 2)`},
		{"detail line first", "| d\na\n", 1, "detail line must come right after its entry, one level deeper"},
		{"detail line at the depth of its entry", "a/\n\tb\n\t| d\n", 3, "detail line must come right after its entry, one level deeper"},
		{"detail line two levels deeper", "a\n\t\t| d\n", 2, "detail line must come right after its entry, one level deeper"},
		{"detail line after a child", "a/\n\tb\n| d\n", 3, "detail line must come right after its entry, one level deeper"},
		{"slash inside a name", "a/b\tx\n", 1, `name "a/b" contains / (write each directory on its own line)`},
		{"slash inside a directory name", "a/b/\n", 1, `name "a/b/" contains / (write each directory on its own line)`},
		{"double slash", "a//\n", 1, `name "a//" contains / (write each directory on its own line)`},
		{"only a slash", "/\tx\n", 1, "empty name"},
		{"space before the separating tab is part of the name", "a/ \tx\n", 1, `name "a/ " contains / (write each directory on its own line)`},
		{"line number counts blank lines", "a/\n\n\n\t\tb\n", 4, "indented more than one level deeper than the entry above"},
		{"line number with CRLF", "a/\r\n\t\tb\r\n", 2, "indented more than one level deeper than the entry above"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := Parse([]byte(tt.in))
			if err == nil {
				t.Fatalf("Parse succeeded: %s", dump(tree))
			}
			var pe *ParseError
			if !errors.As(err, &pe) {
				t.Fatalf("error is %T, want *ParseError: %v", err, err)
			}
			if pe.Line != tt.line || pe.Msg != tt.msg || pe.File != "" {
				t.Errorf("got %+v, want line %d, msg %q", *pe, tt.line, tt.msg)
			}
		})
	}
}

func TestParseErrorMessage(t *testing.T) {
	e := &ParseError{Line: 12, Msg: "indentation must be tabs"}
	if got, want := e.Error(), "line 12: indentation must be tabs"; got != want {
		t.Errorf("without file: got %q, want %q", got, want)
	}
	e.File = ".crube/structure.txt"
	if got, want := e.Error(), ".crube/structure.txt:12: indentation must be tabs"; got != want {
		t.Errorf("with file: got %q, want %q", got, want)
	}
}

func TestFormat(t *testing.T) {
	tree := &Tree{Entries: []*Entry{
		{Name: "a/", Desc: "  spaced desc \t", Detail: "first  \n\n  indented\n", Children: []*Entry{
			{Name: "b"},
			{Name: "c", Desc: "", Detail: "only detail"},
		}},
		{Name: "d", Desc: "x"},
	}}
	want := "a/\tspaced desc\n" +
		"\t| first\n" +
		"\t|\n" +
		"\t|   indented\n" +
		"\t|\n" +
		"\tb\n" +
		"\tc\n" +
		"\t\t| only detail\n" +
		"d\tx\n"
	if got := string(tree.Format()); got != want {
		t.Errorf("Format:\ngot  %q\nwant %q", got, want)
	}
}

func TestIsDir(t *testing.T) {
	for name, want := range map[string]bool{"a/": true, "*/": true, "a": false, "*.go": false} {
		if got := (&Entry{Name: name}).IsDir(); got != want {
			t.Errorf("IsDir(%q) = %v, want %v", name, got, want)
		}
	}
}

// The sample in docs is in the canonical form, so reading and writing it gives
// it back byte for byte.
func TestSampleFiles(t *testing.T) {
	dir := filepath.Join("..", "..", "docs", "examples", "sample-crube")
	for _, name := range []string{"structure.txt", "structure_ja.txt"} {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			tree, err := Parse(data)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if got := string(tree.Format()); got != string(data) {
				t.Errorf("Format differs from the file:\ngot  %q\nwant %q", got, data)
			}
		})
	}

	data, err := os.ReadFile(filepath.Join(dir, "structure.txt"))
	if err != nil {
		t.Fatal(err)
	}
	tree, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	var top []string
	for _, e := range tree.Entries {
		top = append(top, e.Name)
	}
	wantTop := []string{".github/", "cmd/", "config/", "docs/", "internal/", "migrations/", "testdata/", "Makefile", "go.mod"}
	if !reflect.DeepEqual(top, wantTop) {
		t.Errorf("top-level names:\ngot  %q\nwant %q", top, wantTop)
	}
	session := tree.Entries[4].Children[0].Children[0]
	if session.Name != "session.go" {
		t.Fatalf("internal/auth/ first child is %q, want session.go", session.Name)
	}
	wantDetail := "Sessions are stored in Redis with a sliding expiry.\nThe expiry is configured by session.ttl in config/auth.yaml."
	if session.Detail != wantDetail {
		t.Errorf("session.go detail:\ngot  %q\nwant %q", session.Detail, wantDetail)
	}
}

// dump shows a tree on one line, for failure messages.
func dump(t *Tree) string {
	var b strings.Builder
	var walk func([]*Entry)
	walk = func(es []*Entry) {
		b.WriteString("[")
		for i, e := range es {
			if i > 0 {
				b.WriteString(" ")
			}
			b.WriteString(strings.Join([]string{e.Name, e.Desc, e.Detail}, "|"))
			if len(e.Children) > 0 {
				walk(e.Children)
			}
		}
		b.WriteString("]")
	}
	if t != nil {
		walk(t.Entries)
	}
	return b.String()
}
