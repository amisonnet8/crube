package status

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/amisonnet8/crube/internal/ignore"
	"github.com/amisonnet8/crube/internal/structure"
)

// project makes a project in a temporary directory: paths ending in "/" are
// directories, others are empty files. It also makes .crube/ with a file in it,
// which must never be reported.
func project(t *testing.T, paths ...string) string {
	t.Helper()
	root := t.TempDir()
	for _, p := range append(paths, ".crube/structure.txt") {
		full := filepath.Join(root, filepath.FromSlash(p))
		if strings.HasSuffix(p, "/") {
			if err := os.MkdirAll(full, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func check(t *testing.T, root, structureText, ignoreText string) *Report {
	t.Helper()
	tree, err := structure.Parse([]byte(structureText))
	if err != nil {
		t.Fatal(err)
	}
	r, err := Check(root, tree, ignore.Parse([]byte(ignoreText)))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// lines shows the items one per line, like the text output of status.
func lines(r *Report) []string {
	var ls []string
	for _, it := range r.Items {
		switch it.Kind {
		case Untracked:
			ls = append(ls, "untracked "+it.Type+" "+it.Path)
		case Orphan:
			ls = append(ls, "orphan "+it.Path)
		default:
			ls = append(ls, "invalid "+it.Path+" "+it.Field)
		}
	}
	return ls
}

func assertItems(t *testing.T, r *Report, want ...string) {
	t.Helper()
	if got := lines(r); !reflect.DeepEqual(got, want) {
		t.Errorf("items:\ngot  %q\nwant %q", got, want)
	}
}

func assertCoverage(t *testing.T, r *Report, all, dirs int) {
	t.Helper()
	want := Coverage{All: Level{OK: all == 0, Missing: all}, Dirs: Level{OK: dirs == 0, Missing: dirs}}
	if r.Coverage != want {
		t.Errorf("coverage = %+v, want %+v", r.Coverage, want)
	}
}

// The example of docs/examples/getting-started.md, steps 2 and 5.
func TestGettingStarted(t *testing.T) {
	root := project(t, "go.mod", "cmd/myapp/main.go", "internal/store/store.go")

	r := check(t, root, "", "")
	assertItems(t, r,
		"untracked dir cmd/",
		"untracked dir cmd/myapp/",
		"untracked dir internal/",
		"untracked dir internal/store/",
		"untracked file go.mod",
		"untracked file cmd/myapp/main.go",
		"untracked file internal/store/store.go",
	)
	assertCoverage(t, r, 7, 4)
	if r.HasProblems() {
		t.Error("untracked paths alone are problems")
	}

	r = check(t, root, "cmd/\tExecutables\n"+
		"\tmyapp/\tCLI entry point of myapp\n"+
		"internal/\tPrivate packages\n"+
		"\tstore/\tPersistence layer (SQLite)\n"+
		"\t\t| Also called: db, database\n", "")
	assertItems(t, r,
		"untracked file go.mod",
		"untracked file cmd/myapp/main.go",
		"untracked file internal/store/store.go",
	)
	assertCoverage(t, r, 3, 0)
}

// The order: in each directory, files first, then subdirectories, both by
// name; directories are reported before files, then orphans, then invalid.
func TestOrder(t *testing.T) {
	root := project(t, "b.txt", "a.txt", "z/", "z/1", "m/x", "m/a/y", "B/k")
	r := check(t, root, "gone\tx\n", "")
	assertItems(t, r,
		"untracked dir B/",
		"untracked dir m/",
		"untracked dir m/a/",
		"untracked dir z/",
		"untracked file a.txt",
		"untracked file b.txt",
		"untracked file B/k",
		"untracked file m/x",
		"untracked file m/a/y",
		"untracked file z/1",
		"orphan gone",
	)
}

func TestPatternsAndStructureOnlyLines(t *testing.T) {
	root := project(t, "data/2025.csv", "data/2026.csv", "data/notes.txt", "src/auth/login.go", "src/auth/login_test.go")
	r := check(t, root, "data/\tData\n"+
		"\t*.csv\tCSV data\n"+
		"src/\n"+ // only structure: src/ itself is not described
		"\tauth/\tAuthentication\n"+
		"\t\t*_test.go\tTests\n", "")
	assertItems(t, r,
		"untracked dir src/",
		"untracked file data/notes.txt",
		"untracked file src/auth/login.go",
	)
	// A described parent does not describe its children.
	assertCoverage(t, r, 3, 1)
}

func TestOrphans(t *testing.T) {
	root := project(t, "src/main.go", "data/a.csv")
	r := check(t, root, "src/\tSource\n"+
		"\tmain.go\tEntry point\n"+
		"\tlegacy/\tOld code\n"+
		"\t\told.go\tNot reported: below an orphan\n"+
		"\t*.rs\tNo Rust file: a pattern that matches nothing\n"+
		"data/\tData\n"+
		"\t*.csv\tCSV\n"+
		"README.md\tGone\n"+
		"src\tA file entry for what is a directory\n", "")
	assertItems(t, r,
		"orphan src/legacy/",
		"orphan src/*.rs",
		"orphan README.md",
		"orphan src",
	)
	if !r.HasProblems() {
		t.Error("orphans are not problems")
	}
	assertCoverage(t, r, 0, 0)
}

func TestEmptyDirectories(t *testing.T) {
	root := project(t, "a.txt", "empty/", "nested/empty/", "logs/x.log", "keep/x")
	r := check(t, root, "a.txt\tA\n"+
		"empty/\tExists, so not an orphan\n"+
		"\tgone/\tAn orphan below an empty directory\n", "*.log\n")
	// empty/ and nested/ hold nothing; logs/ holds only an ignored file.
	assertItems(t, r,
		"untracked dir keep/",
		"untracked file keep/x",
		"orphan empty/gone/",
	)
	assertCoverage(t, r, 2, 1)
}

func TestIgnored(t *testing.T) {
	root := project(t, "node_modules/pkg/index.js", "build/out.bin", "src/a.go", "src/a.gen.go")
	r := check(t, root, "node_modules/\tDependencies (ignored, but they exist)\n"+
		"\tpkg/\tNot looked at\n"+
		"\tgone/\tNot looked at either\n"+
		"src/\tSource\n"+
		"\ta.go\tA\n", "node_modules/\nbuild/\n*.gen.go\n")
	assertItems(t, r)
	assertCoverage(t, r, 0, 0)

	// --no-ignore: everything is looked at, but .crube/ never is.
	tree, err := structure.Parse([]byte("src/\tSource\n\ta.go\tA\n"))
	if err != nil {
		t.Fatal(err)
	}
	r, err = Check(root, tree, ignore.None())
	if err != nil {
		t.Fatal(err)
	}
	assertItems(t, r,
		"untracked dir build/",
		"untracked dir node_modules/",
		"untracked dir node_modules/pkg/",
		"untracked file build/out.bin",
		"untracked file node_modules/pkg/index.js",
		"untracked file src/a.gen.go",
	)
}

func TestSymbolicLinkIsAFile(t *testing.T) {
	root := project(t, "real/x")
	if err := os.Symlink("real", filepath.Join(root, "link")); err != nil {
		t.Skipf("cannot make a symbolic link here: %v", err)
	}
	r := check(t, root, "real/\tReal\n\tx\tX\nlink/\tA directory entry does not match a link\n", "")
	assertItems(t, r,
		"untracked file link",
		"orphan link/",
	)
	assertCoverage(t, r, 1, 0)
}

func TestInvalid(t *testing.T) {
	root := project(t, "a", "b")
	long := strings.Repeat("x", 257)
	detail := strings.Repeat("| "+strings.Repeat("y", 1000)+"\n\t", 5)
	r := check(t, root, "a\t"+long+"\n"+
		"\t"+detail+"\n"+
		"b\tok\n", "")
	assertItems(t, r,
		"invalid a desc",
		"invalid a detail",
	)
	want := []Item{
		{Kind: Invalid, Path: "a", Field: "desc", Width: 257, Limit: 256},
		{Kind: Invalid, Path: "a", Field: "detail", Width: 5000, Limit: 4096},
	}
	if !reflect.DeepEqual(r.Items, want) {
		t.Errorf("items = %+v", r.Items)
	}
	if !r.HasProblems() {
		t.Error("invalid values are not problems")
	}
	// An invalid value does not make a path undescribed.
	assertCoverage(t, r, 0, 0)
}

func TestNothingToReport(t *testing.T) {
	r := check(t, project(t), "", "")
	if r.Items == nil || len(r.Items) != 0 {
		t.Errorf("items = %#v, want an empty, non-nil list (JSON [])", r.Items)
	}
	assertCoverage(t, r, 0, 0)
}
