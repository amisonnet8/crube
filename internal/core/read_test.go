package core

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/amisonnet8/crube/internal/status"
)

const readStructure = "src/\tSource\n" +
	"\tauth/\tAuthentication\n" +
	"\t\t| Entry point is auth.go.\n" +
	"\t\tsession.go\tSessions\n" +
	"\t\t*_test.go\tTests\n" +
	"data/\n" +
	"\t*.csv\tCSV data\n" +
	"build/\tBuild output\n" +
	"go.mod\tModule\n"

// readProject makes a project to read from. The files are empty; build/ is
// ignored.
func readProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, p := range []string{"go.mod", "README.md", "src/auth/session.go", "src/auth/login_test.go", "src/db/conn.go", "data/2026.csv", "build/out.bin", "empty/"} {
		full := filepath.Join(root, filepath.FromSlash(p))
		if strings.HasSuffix(p, "/") {
			mustMkdir(t, full)
			continue
		}
		mustMkdir(t, filepath.Dir(full))
		if err := os.WriteFile(full, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustMkdir(t, filepath.Join(root, ".crube"))
	for name, content := range map[string]string{"structure.txt": readStructure, "ignore": "build/\n"} {
		if err := os.WriteFile(filepath.Join(root, ".crube", name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func mustMkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}

func open(t *testing.T, o Options) *Project {
	t.Helper()
	p, err := Open(o)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLs(t *testing.T) {
	root := readProject(t)
	tests := []struct {
		dir, path string // dir is where the operation runs, relative to root
		wantPath  string
		want      []LsItem
	}{
		{"", "", "", []LsItem{
			{"README.md", "file", "", ""},
			{"data/", "dir", "data/", ""},
			{"empty/", "dir", "", ""},
			{"go.mod", "file", "go.mod", "Module"},
			{"src/", "dir", "src/", "Source"},
		}},
		{"", "src", "src/", []LsItem{
			{"auth/", "dir", "src/auth/", "Authentication"},
			{"db/", "dir", "", ""},
		}},
		{"", "src/auth", "src/auth/", []LsItem{
			{"login_test.go", "file", "src/auth/*_test.go", "Tests"},
			{"session.go", "file", "src/auth/session.go", "Sessions"},
		}},
		{"src", "auth/", "src/auth/", []LsItem{
			{"login_test.go", "file", "src/auth/*_test.go", "Tests"},
			{"session.go", "file", "src/auth/session.go", "Sessions"},
		}},
		{"src/auth", ".", "src/auth/", []LsItem{
			{"login_test.go", "file", "src/auth/*_test.go", "Tests"},
			{"session.go", "file", "src/auth/session.go", "Sessions"},
		}},
		{"src/auth", "..", "src/", []LsItem{
			{"auth/", "dir", "src/auth/", "Authentication"},
			{"db/", "dir", "", ""},
		}},
		{"", "src/db", "src/db/", []LsItem{
			{"conn.go", "file", "", ""},
		}},
		{"", "data", "data/", []LsItem{
			{"2026.csv", "file", "data/*.csv", "CSV data"},
		}},
		{"", "empty", "empty/", []LsItem{}},
	}
	for _, tt := range tests {
		t.Run(tt.dir+":"+tt.path, func(t *testing.T) {
			p := open(t, Options{Dir: filepath.Join(root, filepath.FromSlash(tt.dir))})
			r, err := p.Ls(tt.path)
			if err != nil {
				t.Fatal(err)
			}
			if r.Path != tt.wantPath || !reflect.DeepEqual(r.Items, tt.want) {
				t.Errorf("got %q %+v\nwant %q %+v", r.Path, r.Items, tt.wantPath, tt.want)
			}
		})
	}

	// With --no-ignore, build/ is listed; .crube/ never is.
	r, err := open(t, Options{Dir: root, NoIgnore: true}).Ls("")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, it := range r.Items {
		names = append(names, it.Name)
	}
	if want := []string{"README.md", "build/", "data/", "empty/", "go.mod", "src/"}; !reflect.DeepEqual(names, want) {
		t.Errorf("--no-ignore: %q, want %q", names, want)
	}
}

func TestLsRefused(t *testing.T) {
	root := readProject(t)
	p := open(t, Options{Dir: root})
	for path, msg := range map[string]string{
		"..":            "..: outside the project",
		"../x":          "../x: outside the project",
		"nope":          "nope: no such file",
		"go.mod":        "go.mod: not a directory (use crube show)",
		"build":         "build: ignored by .crube/ignore (use --no-ignore)",
		"build/":        "build/: ignored by .crube/ignore (use --no-ignore)",
		".crube":        ".crube: ignored by .crube/ignore (use --no-ignore)",
		"src/../..":     "src/../..: outside the project",
		"data/2026.csv": "data/2026.csv: not a directory (use crube show)",
	} {
		if _, err := p.Ls(path); err == nil || err.Error() != msg {
			t.Errorf("ls %s: got %v, want %q", path, err, msg)
		}
	}
}

func TestShow(t *testing.T) {
	root := readProject(t)
	p := open(t, Options{Dir: root})
	tests := []struct {
		path string
		want ShowResult
	}{
		{"data/2026.csv", ShowResult{Path: "data/2026.csv", Type: "file", Entry: "data/*.csv", Desc: "CSV data"}},
		{"src/auth", ShowResult{Path: "src/auth/", Type: "dir", Entry: "src/auth/", Desc: "Authentication", Detail: "Entry point is auth.go."}},
		{"src/auth/session.go", ShowResult{Path: "src/auth/session.go", Type: "file", Entry: "src/auth/session.go", Desc: "Sessions"}},
		{"src/db/conn.go", ShowResult{Path: "src/db/conn.go", Type: "file"}},
		{"data", ShowResult{Path: "data/", Type: "dir", Entry: "data/"}},
	}
	for _, tt := range tests {
		r, err := p.Show(tt.path)
		if err != nil {
			t.Errorf("show %s: %v", tt.path, err)
			continue
		}
		if *r != tt.want {
			t.Errorf("show %s:\ngot  %+v\nwant %+v", tt.path, *r, tt.want)
		}
	}

	for path, msg := range map[string]string{
		".":             ".: the project root has no entry (give a path below it)",
		"missing":       "missing: no such file",
		"build/out.bin": "build/out.bin: ignored by .crube/ignore (use --no-ignore)",
		"../x":          "../x: outside the project",
	} {
		if _, err := p.Show(path); err == nil || err.Error() != msg {
			t.Errorf("show %s: got %v, want %q", path, err, msg)
		}
	}
}

func TestShowLink(t *testing.T) {
	root := readProject(t)
	if err := os.Symlink("src", filepath.Join(root, "current")); err != nil {
		t.Skipf("cannot make a symbolic link here: %v", err)
	}
	p := open(t, Options{Dir: root})
	if _, err := p.Set("current", "A link to the source", nil); err != nil {
		t.Fatal(err)
	}
	r, err := p.Show("current")
	if err != nil {
		t.Fatal(err)
	}
	want := ShowResult{Path: "current", Type: "file", Entry: "current", Desc: "A link to the source", Link: "src"}
	if *r != want {
		t.Errorf("got %+v, want %+v", *r, want)
	}
	// ls shows the link as a file, and does not look into it.
	l, err := p.Ls("")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, it := range l.Items {
		if it.Name == "current" {
			found = it.Type == "file" && it.Entry == "current"
		}
	}
	if !found {
		t.Errorf("ls: %+v", l.Items)
	}
}

// paths flattens a tree into "path" lines, with "*" on the marked one and
// two spaces of indent per level.
func paths(nodes []*TreeNode, indent string) []string {
	var ls []string
	for _, n := range nodes {
		line := indent + n.Path
		if n.Here {
			line += " *"
		}
		ls = append(ls, line)
		ls = append(ls, paths(n.Children, indent+"  ")...)
	}
	return ls
}

func TestTree(t *testing.T) {
	root := readProject(t)
	all := []string{"src/", "  src/auth/", "    src/auth/session.go", "    src/auth/*_test.go", "data/", "  data/*.csv", "go.mod"}
	tests := []struct {
		name  string
		dir   string
		o     TreeOptions
		here  string
		above bool
		want  []string
	}{
		{"at the root: no mark", "", TreeOptions{Mark: true}, "", false, all},
		{"an entry of the current directory", "src/auth", TreeOptions{Mark: true}, "src/auth/", false,
			[]string{"src/", "  src/auth/ *", "    src/auth/session.go", "    src/auth/*_test.go", "data/", "  data/*.csv", "go.mod"}},
		{"no entry: the deepest above", "src/db", TreeOptions{Mark: true}, "src/db/", true,
			[]string{"src/ *", "  src/auth/", "    src/auth/session.go", "    src/auth/*_test.go", "data/", "  data/*.csv", "go.mod"}},
		{"a line without a desc", "data", TreeOptions{Mark: true}, "data/", false,
			[]string{"src/", "  src/auth/", "    src/auth/session.go", "    src/auth/*_test.go", "data/ *", "  data/*.csv", "go.mod"}},
		{"nothing matches: no mark", "empty", TreeOptions{Mark: true}, "empty/", false, all},
		{"without the mark (MCP)", "src/auth", TreeOptions{}, "", false, all},
		{"depth 1", "", TreeOptions{Depth: 1}, "", false, []string{"src/", "data/", "go.mod"}},
		{"depth 2", "", TreeOptions{Depth: 2}, "", false, []string{"src/", "  src/auth/", "data/", "  data/*.csv", "go.mod"}},
		{"a subtree", "", TreeOptions{Path: "src/auth/"}, "", false,
			[]string{"src/auth/", "  src/auth/session.go", "  src/auth/*_test.go"}},
		{"a subtree with a depth", "", TreeOptions{Path: "src/", Depth: 2}, "", false,
			[]string{"src/", "  src/auth/"}},
		{"a pattern entry", "", TreeOptions{Path: "data/*.csv"}, "", false, []string{"data/*.csv"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := open(t, Options{Dir: filepath.Join(root, filepath.FromSlash(tt.dir))})
			r, err := p.Tree(tt.o)
			if err != nil {
				t.Fatal(err)
			}
			if got := paths(r.Entries, ""); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tree:\ngot  %q\nwant %q", got, tt.want)
			}
			if r.Here != tt.here || r.Above != tt.above {
				t.Errorf("here %q above %v, want %q %v", r.Here, r.Above, tt.here, tt.above)
			}
		})
	}

	// The nodes carry desc and detail; build/ shows with --no-ignore.
	r, err := open(t, Options{Dir: root, NoIgnore: true}).Tree(TreeOptions{Path: "src/auth/", Depth: 1})
	if err != nil {
		t.Fatal(err)
	}
	want := &TreeNode{Name: "auth/", Path: "src/auth/", Desc: "Authentication", Detail: "Entry point is auth.go.", Children: []*TreeNode{}}
	if len(r.Entries) != 1 || !reflect.DeepEqual(r.Entries[0], want) {
		t.Errorf("node = %+v", r.Entries)
	}
	r, err = open(t, Options{Dir: root, NoIgnore: true}).Tree(TreeOptions{Depth: 1})
	if err != nil {
		t.Fatal(err)
	}
	if got := paths(r.Entries, ""); !reflect.DeepEqual(got, []string{"src/", "data/", "build/", "go.mod"}) {
		t.Errorf("--no-ignore: %q", got)
	}

	if _, err := open(t, Options{Dir: root}).Tree(TreeOptions{Path: "nope/"}); err == nil || err.Error() != "no entry nope/" {
		t.Errorf("tree nope/: %v", err)
	}
}

func TestFindAndStatus(t *testing.T) {
	root := readProject(t)
	p := open(t, Options{Dir: filepath.Join(root, "src")})

	f, err := p.Find("tests", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Items) != 1 || f.Items[0].Path != "src/auth/*_test.go" {
		t.Errorf("find: %+v", f.Items)
	}
	// build/ is ignored, so its entry is not found.
	if f, err := p.Find("build", 0); err != nil || len(f.Items) != 0 {
		t.Errorf("find build: %+v, %v", f, err)
	}

	r, err := p.Status()
	if err != nil {
		t.Fatal(err)
	}
	if r.Structure != ".crube/structure.txt" {
		t.Errorf("structure = %q", r.Structure)
	}
	var got []string
	for _, it := range r.Items {
		got = append(got, it.Kind+" "+it.Type+" "+it.Path)
	}
	// From src/, status still checks the whole project.
	want := []string{
		"untracked dir data/", // its entry has no desc: it only gives structure
		"untracked dir src/db/",
		"untracked file README.md",
		"untracked file src/db/conn.go",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("status:\ngot  %q\nwant %q", got, want)
	}
}

func TestFailed(t *testing.T) {
	clean := &status.Report{Coverage: status.Coverage{All: status.Level{OK: true}, Dirs: status.Level{OK: true}}}
	dirsOnly := &status.Report{Coverage: status.Coverage{All: status.Level{Missing: 2}, Dirs: status.Level{OK: true}}}
	none := &status.Report{Coverage: status.Coverage{All: status.Level{Missing: 3}, Dirs: status.Level{Missing: 1}}}
	orphan := &status.Report{Items: []status.Item{{Kind: status.Orphan, Path: "x"}}, Coverage: clean.Coverage}

	tests := []struct {
		r       *status.Report
		require string
		want    bool
	}{
		{clean, "", false}, {clean, "all", false}, {clean, "dirs", false},
		{dirsOnly, "", false}, {dirsOnly, "all", true}, {dirsOnly, "dirs", false},
		{none, "", false}, {none, "all", true}, {none, "dirs", true},
		{orphan, "", true}, {orphan, "all", true}, {orphan, "dirs", true},
	}
	for i, tt := range tests {
		got, err := Failed(tt.r, tt.require)
		if err != nil || got != tt.want {
			t.Errorf("%d: Failed(require %q) = %v, %v; want %v", i, tt.require, got, err, tt.want)
		}
	}
	if _, err := Failed(clean, "files"); err == nil || err.Error() != `invalid --require "files" (use all or dirs)` {
		t.Errorf("bad --require: %v", err)
	}
}
