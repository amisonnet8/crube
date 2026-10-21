package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// run runs the command line with CRUBE_LANG set to lang ("" for unset).
func run(t *testing.T, lang string, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errb bytes.Buffer
	code = Run(args, Env{
		Getenv: func(k string) string {
			if k == "CRUBE_LANG" {
				return lang
			}
			return ""
		},
		Stdout: &out,
		Stderr: &errb,
	})
	return code, out.String(), errb.String()
}

// newProject makes a project with go.mod and internal/store/store.go, and
// runs crube init in it.
func newProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "internal", "store"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"go.mod", "internal/store/store.go"} {
		if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(f)), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	code, out, errs := run(t, "", "init", "-C", dir)
	if code != 0 || out != "created .crube/structure.txt\ncreated .crube/ignore\n" || errs != "" {
		t.Fatalf("init: %d %q %q", code, out, errs)
	}
	return dir
}

func readStructure(t *testing.T, dir, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, ".crube", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestParse(t *testing.T) {
	allowed := []string{"--json", "--full", "-C", "--detail"}
	tests := []struct {
		name string
		args []string
		opts map[string]string
		pos  []string
		help bool
		err  string
	}{
		{"option after the path", []string{"src", "--full"}, map[string]string{"--full": ""}, []string{"src"}, false, ""},
		{"empty value", []string{"--detail", "", "x"}, map[string]string{"--detail": ""}, []string{"x"}, false, ""},
		{"value that starts with -", []string{"--detail", "-x"}, map[string]string{"--detail": "-x"}, nil, false, ""},
		{"value after =", []string{"--detail=-x", "-C=d"}, map[string]string{"--detail": "-x", "-C": "d"}, nil, false, ""},
		{"after -- is positional", []string{"--json", "--", "-v", "--json"}, map[string]string{"--json": ""}, []string{"-v", "--json"}, false, ""},
		{"lone - is positional", []string{"-"}, map[string]string{}, []string{"-"}, false, ""},
		{"help", []string{"a", "-h"}, map[string]string{}, []string{"a"}, true, ""},
		{"flag with a value", []string{"--json=1"}, nil, nil, false, "--json takes no value"},
		{"missing value", []string{"-C"}, nil, nil, false, "-C needs a value"},
		{"unknown option", []string{"--foo"}, nil, nil, false, "unknown option --foo"},
		{"option of another command", []string{"--limit", "3"}, nil, nil, false, "unknown option --limit"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parse(tt.args, allowed)
			if tt.err != "" {
				if err == nil || err.Error() != tt.err {
					t.Fatalf("err = %v, want %q", err, tt.err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(p.opts, tt.opts) || !reflect.DeepEqual(p.args, tt.pos) || p.help != tt.help {
				t.Errorf("got %v %q %v, want %v %q %v", p.opts, p.args, p.help, tt.opts, tt.pos, tt.help)
			}
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		opts map[string]string
		want int
		err  string
	}{
		{map[string]string{}, 0, ""},
		{map[string]string{"-d": "3"}, 3, ""},
		{map[string]string{"-d": "0"}, 0, `invalid -d "0" (use a number of 1 or more)`},
		{map[string]string{"-d": "x"}, 0, `invalid -d "x" (use a number of 1 or more)`},
	}
	for _, tt := range tests {
		n, err := (&parsed{opts: tt.opts}).count("-d")
		if tt.err != "" {
			if err == nil || err.Error() != tt.err {
				t.Errorf("%v: err = %v, want %q", tt.opts, err, tt.err)
			}
			continue
		}
		if err != nil || n != tt.want {
			t.Errorf("%v: got %d %v, want %d", tt.opts, n, err, tt.want)
		}
	}
}

func TestUsage(t *testing.T) {
	code, out, errs := run(t, "")
	if code != 2 || out != "" || errs != usage {
		t.Errorf("no args: %d %q %q", code, out, errs)
	}
	for _, a := range []string{"-h", "--help", "help"} {
		code, out, errs := run(t, "", a)
		if code != 0 || out != usage || errs != "" {
			t.Errorf("%s: %d %q %q", a, code, out, errs)
		}
	}
	code, out, errs = run(t, "", "ls", "-h")
	if code != 0 || out != "usage: crube ls [path] [--full] [--json] [-C <dir>] [--no-ignore]\n" || errs != "" {
		t.Errorf("ls -h: %d %q %q", code, out, errs)
	}
}

func TestVersion(t *testing.T) {
	for v, want := range map[string]string{"v1.2.3": "crube v1.2.3\n", "": "crube (devel)\n"} {
		var out bytes.Buffer
		code := Run([]string{"--version"}, Env{Getenv: func(string) string { return "" }, Stdout: &out, Stderr: &out, Version: v})
		if code != 0 || out.String() != want {
			t.Errorf("version %q: %d %q, want %q", v, code, out.String(), want)
		}
	}
}

func TestArgumentErrors(t *testing.T) {
	dir := newProject(t)
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"lss"}, `crube: unknown command "lss" (see crube --help)` + "\n"},
		{[]string{"--json", "ls"}, `crube: unknown command "--json" (see crube --help)` + "\n"},
		{[]string{"ls", "--foo"}, "crube: ls: unknown option --foo\n"},
		{[]string{"ls", "-C"}, "crube: ls: -C needs a value\n"},
		{[]string{"ls", "a", "b"}, "crube: ls: too many arguments (usage: crube ls [path] [--full] [--json] [-C <dir>] [--no-ignore])\n"},
		{[]string{"set", "x"}, "crube: set: missing <desc> (usage: crube set <path> <desc> [--detail <text>] [--json] [-C <dir>])\n"},
		{[]string{"find"}, "crube: find: missing <query> (usage: crube find <query> [--limit N] [--json] [-C <dir>] [--no-ignore])\n"},
		{[]string{"init", "--json"}, "crube: init: unknown option --json\n"},
		{[]string{"find", "-C", dir, "--limit", "0", "q"}, `crube: find: invalid --limit "0" (use a number of 1 or more)` + "\n"},
		{[]string{"tree", "-C", dir, "-d", "x"}, `crube: tree: invalid -d "x" (use a number of 1 or more)` + "\n"},
		{[]string{"status", "-C", dir, "--require", "x"}, `crube: status: invalid --require "x" (use all or dirs)` + "\n"},
	}
	for _, tt := range tests {
		code, out, errs := run(t, "", tt.args...)
		if code != 2 || out != "" || errs != tt.want {
			t.Errorf("%q: %d %q %q, want 2 %q", tt.args, code, out, errs, tt.want)
		}
	}
}

func TestCoreErrors(t *testing.T) {
	dir := newProject(t)
	tests := []struct {
		lang string
		args []string
		want string
	}{
		{"", []string{"ls", "--json", "-C", t.TempDir()}, "crube: no .crube/ found (run crube init)\n"},
		{"", []string{"init", "-C", dir}, "crube: nothing to create: .crube/structure.txt already exists\n"},
		{"bad_tag", []string{"ls", "--json", "-C", dir}, `crube: invalid CRUBE_LANG "bad_tag" (use a language tag such as ja or zh-CN)` + "\n"},
		{"", []string{"show", "--json", "-C", dir, "nothing"}, "crube: nothing: no such file\n"},
		{"", []string{"rm", "-C", dir, "nothing"}, "crube: no entry nothing\n"},
	}
	for _, tt := range tests {
		code, out, errs := run(t, tt.lang, tt.args...)
		if code != 2 || out != "" || errs != tt.want {
			t.Errorf("%q: %d %q %q, want 2 %q", tt.args, code, out, errs, tt.want)
		}
	}
}

func TestInitLanguage(t *testing.T) {
	dir := newProject(t)
	code, out, errs := run(t, "ja", "init", "-C", dir)
	if code != 0 || out != "created .crube/structure_ja.txt\n" || errs != "" {
		t.Errorf("%d %q %q", code, out, errs)
	}
}

func TestEdit(t *testing.T) {
	dir := newProject(t)
	steps := [][]string{
		{"set", "-C", dir, "internal/store/", "Persistence layer (SQLite)", "--detail", "Also called: db, database"},
		{"set", "-C", dir, "--", "a.txt", "-v <flag> & more"},
		{"set", "-C", dir, "old/", "Old"},
		{"mv", "-C", dir, "old/", "cmd/"},
		{"set", "-C", dir, "tmp/", "Temporary"},
		{"rm", "-C", dir, "tmp/"},
	}
	for _, args := range steps {
		code, out, errs := run(t, "", args...)
		if code != 0 || out != "" || errs != "" {
			t.Fatalf("%q: %d %q %q", args, code, out, errs)
		}
	}
	want := "internal/\n" +
		"\tstore/\tPersistence layer (SQLite)\n" +
		"\t\t| Also called: db, database\n" +
		"a.txt\t-v <flag> & more\n" +
		"cmd/\tOld\n"
	if got := readStructure(t, dir, "structure.txt"); got != want {
		t.Errorf("structure.txt:\n%s\nwant:\n%s", got, want)
	}

	// --detail "" removes the detail.
	if code, _, errs := run(t, "", "set", "-C", dir, "internal/store/", "Persistence layer (SQLite)", "--detail", ""); code != 0 {
		t.Fatal(errs)
	}
	if got := readStructure(t, dir, "structure.txt"); strings.Contains(got, "|") {
		t.Errorf("detail is left:\n%s", got)
	}
}

func TestEditJSON(t *testing.T) {
	dir := newProject(t)
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"set", "--json", "-C", dir, "internal/store/", "Store", "--detail", "SQLite"},
			"{\n  \"path\": \"internal/store/\",\n  \"desc\": \"Store\",\n  \"detail\": \"SQLite\"\n}\n"},
		// internal/ (added as a parent) and store/ below it.
		{[]string{"mv", "--json", "-C", dir, "internal/", "lib/"},
			"{\n  \"from\": \"internal/\",\n  \"to\": \"lib/\",\n  \"moved\": 2\n}\n"},
		{[]string{"rm", "--json", "-C", dir, "lib/"},
			"{\n  \"path\": \"lib/\",\n  \"removed\": 2\n}\n"},
	}
	for _, tt := range tests {
		code, out, errs := run(t, "", tt.args...)
		if code != 0 || out != tt.want || errs != "" {
			t.Errorf("%q: %d %q %q\nwant %q", tt.args, code, out, errs, tt.want)
		}
	}
}

func TestShowJSON(t *testing.T) {
	dir := newProject(t)
	if code, _, errs := run(t, "", "set", "-C", dir, "internal/store/", "a<b>&c", "--detail", "line 1\nline 2"); code != 0 {
		t.Fatal(errs)
	}
	// The JSON is indented by 2, and keeps <, > and & as they are.
	want := `{
  "path": "internal/store/",
  "type": "dir",
  "entry": "internal/store/",
  "desc": "a<b>&c",
  "detail": "line 1\nline 2",
  "link": ""
}
`
	code, out, errs := run(t, "", "show", "--json", "-C", dir, "internal/store")
	if code != 0 || out != want || errs != "" {
		t.Errorf("%d %q %q\nwant %q", code, out, errs, want)
	}
}

func TestTreeJSON(t *testing.T) {
	dir := newProject(t)
	if code, _, errs := run(t, "", "set", "-C", dir, "internal/store/", "Store"); code != 0 {
		t.Fatal(errs)
	}
	code, out, errs := run(t, "", "tree", "--json", "-C", filepath.Join(dir, "internal", "store"))
	if code != 0 || errs != "" {
		t.Fatalf("%d %q", code, errs)
	}
	var got struct {
		Here    string `json:"here"`
		Entries []struct {
			Name     string `json:"name"`
			Here     bool   `json:"here"`
			Children []struct {
				Name string `json:"name"`
				Here bool   `json:"here"`
			} `json:"children"`
		} `json:"entries"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.Here != "internal/store/" || len(got.Entries) != 1 || got.Entries[0].Name != "internal/" || got.Entries[0].Here ||
		len(got.Entries[0].Children) != 1 || got.Entries[0].Children[0].Name != "store/" || !got.Entries[0].Children[0].Here {
		t.Errorf("tree --json:\n%s", out)
	}

	// -d 1 shows the top level only.
	code, out, _ = run(t, "", "tree", "--json", "-d", "1", "-C", dir)
	if code != 0 || strings.Contains(out, "store/") {
		t.Errorf("tree -d 1:\n%s", out)
	}
}

func TestFindJSON(t *testing.T) {
	dir := newProject(t)
	for _, args := range [][]string{
		{"set", "-C", dir, "internal/", "Private packages"},
		{"set", "-C", dir, "internal/store/", "Persistence layer (SQLite)"},
	} {
		if code, _, errs := run(t, "", args...); code != 0 {
			t.Fatal(errs)
		}
	}
	type found struct {
		Query string `json:"query"`
		More  int    `json:"more"`
		Items []struct {
			Path string `json:"path"`
		} `json:"items"`
	}
	tests := []struct {
		args  []string
		query string
		more  int
		paths []string
	}{
		// Several words are one query: all of them must be found.
		{[]string{"layer", "store"}, "layer store", 0, []string{"internal/store/"}},
		{[]string{"internal"}, "internal", 0, []string{"internal/", "internal/store/"}},
		{[]string{"internal", "--limit", "1"}, "internal", 1, []string{"internal/"}},
	}
	for _, tt := range tests {
		code, out, errs := run(t, "", append([]string{"find", "--json", "-C", dir}, tt.args...)...)
		if code != 0 || errs != "" {
			t.Fatalf("%q: %d %q", tt.args, code, errs)
		}
		var got found
		if err := json.Unmarshal([]byte(out), &got); err != nil {
			t.Fatal(err)
		}
		var paths []string
		for _, it := range got.Items {
			paths = append(paths, it.Path)
		}
		if got.Query != tt.query || got.More != tt.more || !reflect.DeepEqual(paths, tt.paths) {
			t.Errorf("%q: %q %d %q, want %q %d %q", tt.args, got.Query, got.More, paths, tt.query, tt.more, tt.paths)
		}
	}
}

func TestLsJSON(t *testing.T) {
	dir := newProject(t)
	if code, _, errs := run(t, "", "set", "-C", dir, "go.mod", "Module"); code != 0 {
		t.Fatal(errs)
	}
	code, out, errs := run(t, "", "ls", "--json", "--full", "-C", dir)
	if code != 0 || errs != "" {
		t.Fatalf("%d %q", code, errs)
	}
	var got struct {
		Path  string `json:"path"`
		Items []struct {
			Name string `json:"name"`
			Desc string `json:"desc"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.Path != "" || len(got.Items) != 2 || got.Items[0].Name != "go.mod" || got.Items[0].Desc != "Module" || got.Items[1].Name != "internal/" {
		t.Errorf("ls --json:\n%s", out)
	}
}

func TestStatusExitCode(t *testing.T) {
	dir := newProject(t)
	// Nothing described: untracked only, which fails only with --require.
	tests := []struct {
		require string
		want    int
	}{{"", 0}, {"dirs", 1}, {"all", 1}}
	check := func(stage string) {
		t.Helper()
		for _, tt := range tests {
			args := []string{"status", "--json", "-C", dir}
			if tt.require != "" {
				args = append(args, "--require", tt.require)
			}
			code, out, errs := run(t, "", args...)
			if code != tt.want || errs != "" || !strings.Contains(out, `"structure": ".crube/structure.txt"`) {
				t.Errorf("%s %q: %d %q, want %d", stage, args, code, errs, tt.want)
			}
		}
	}
	check("nothing described")

	// The directories described: --require dirs passes, all does not.
	for _, args := range [][]string{
		{"set", "-C", dir, "internal/", "Private packages"},
		{"set", "-C", dir, "internal/store/", "Store"},
	} {
		if code, _, errs := run(t, "", args...); code != 0 {
			t.Fatal(errs)
		}
	}
	tests = []struct {
		require string
		want    int
	}{{"", 0}, {"dirs", 0}, {"all", 1}}
	check("dirs described")

	// An orphan fails even without --require.
	if code, _, errs := run(t, "", "set", "-C", dir, "gone.txt", "Gone"); code != 0 {
		t.Fatal(errs)
	}
	tests = []struct {
		require string
		want    int
	}{{"", 1}, {"dirs", 1}}
	check("orphan")
}
