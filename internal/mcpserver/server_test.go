package mcpserver

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/amisonnet8/crube/internal/core"
)

// project makes a project with src/main.go, src/db/conn.go and a structure
// file, and returns its root.
func project(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src", "db"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"src/main.go", "src/db/conn.go"} {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(f)), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := core.Init(root, ""); err != nil {
		t.Fatal(err)
	}
	structure := "src/\tSource\n\tmain.go\tEntry point\n\t\t| Starts the server.\nold/\tGone\n"
	if err := os.WriteFile(filepath.Join(root, ".crube", "structure.txt"), []byte(structure), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}

// connect starts the server with o and returns a client session to it.
func connect(t *testing.T, o Options) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	st, ct := mcp.NewInMemoryTransports()
	ss, err := New(o).Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ss.Close() })
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

// call calls the tool, and decodes what it returns into out. It returns the
// text of the error when the tool fails ("" when it does not).
func call(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any, out any) string {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	text := res.Content[0].(*mcp.TextContent).Text
	if res.IsError {
		return text
	}
	if err := json.Unmarshal([]byte(text), out); err != nil {
		t.Fatalf("%s: %v: %s", name, err, text)
	}
	return ""
}

func toolNames(t *testing.T, cs *mcp.ClientSession) []string {
	t.Helper()
	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, tool := range res.Tools {
		names = append(names, tool.Name)
	}
	slices.Sort(names)
	return names
}

func TestTools(t *testing.T) {
	root := project(t)
	read := []string{"find", "ls", "show", "status", "tree"}
	if got := toolNames(t, connect(t, Options{Dir: root})); !reflect.DeepEqual(got, read) {
		t.Errorf("without --allow-write: %q, want %q", got, read)
	}
	all := []string{"find", "ls", "move_description", "remove_description", "set_description", "show", "status", "tree"}
	if got := toolNames(t, connect(t, Options{Dir: root, AllowWrite: true})); !reflect.DeepEqual(got, all) {
		t.Errorf("with --allow-write: %q, want %q", got, all)
	}
}

func TestRead(t *testing.T) {
	root := project(t)
	// The server runs in a subdirectory: real paths are still read from the
	// root, and tree has no mark.
	cs := connect(t, Options{Dir: filepath.Join(root, "src", "db")})

	var tree core.TreeResult
	if msg := call(t, cs, "tree", map[string]any{"depth": 1}, &tree); msg != "" {
		t.Fatal(msg)
	}
	if tree.Here != "" || len(tree.Entries) != 2 || tree.Entries[0].Path != "src/" || tree.Entries[0].Here ||
		len(tree.Entries[0].Children) != 0 || tree.Entries[1].Path != "old/" {
		t.Errorf("tree: %+v", tree)
	}

	var ls core.LsResult
	if msg := call(t, cs, "ls", map[string]any{"path": "src"}, &ls); msg != "" {
		t.Fatal(msg)
	}
	wantLs := core.LsResult{Path: "src/", Items: []core.LsItem{
		{Name: "db/", Type: "dir", Entry: "", Desc: ""},
		{Name: "main.go", Type: "file", Entry: "src/main.go", Desc: "Entry point"},
	}}
	if !reflect.DeepEqual(ls, wantLs) {
		t.Errorf("ls: %+v, want %+v", ls, wantLs)
	}

	var show core.ShowResult
	if msg := call(t, cs, "show", map[string]any{"path": "src/main.go"}, &show); msg != "" {
		t.Fatal(msg)
	}
	wantShow := core.ShowResult{Path: "src/main.go", Type: "file", Entry: "src/main.go", Desc: "Entry point", Detail: "Starts the server."}
	if show != wantShow {
		t.Errorf("show: %+v, want %+v", show, wantShow)
	}

	var found struct {
		Query string `json:"query"`
		More  int    `json:"more"`
		Items []struct {
			Path string `json:"path"`
		} `json:"items"`
	}
	if msg := call(t, cs, "find", map[string]any{"query": "server", "limit": 1}, &found); msg != "" {
		t.Fatal(msg)
	}
	if found.Query != "server" || found.More != 0 || len(found.Items) != 1 || found.Items[0].Path != "src/main.go" {
		t.Errorf("find: %+v", found)
	}
}

func TestStatusKinds(t *testing.T) {
	cs := connect(t, Options{Dir: project(t)})
	type report struct {
		Structure string `json:"structure"`
		Items     []struct {
			Kind string `json:"kind"`
			Path string `json:"path"`
		} `json:"items"`
		Coverage struct {
			All struct {
				Missing int `json:"missing"`
			} `json:"all"`
		} `json:"coverage"`
	}
	// All: untracked src/db/ and src/db/conn.go, orphan old/.
	var all report
	if msg := call(t, cs, "status", nil, &all); msg != "" {
		t.Fatal(msg)
	}
	if all.Structure != ".crube/structure.txt" || len(all.Items) != 3 || all.Coverage.All.Missing != 2 {
		t.Errorf("status: %+v", all)
	}
	// Only the orphans; the coverage is still that of the whole project.
	var orphans report
	if msg := call(t, cs, "status", map[string]any{"kinds": []string{"orphan"}}, &orphans); msg != "" {
		t.Fatal(msg)
	}
	if len(orphans.Items) != 1 || orphans.Items[0].Path != "old/" || orphans.Coverage.All.Missing != 2 {
		t.Errorf("status orphan: %+v", orphans)
	}
	if msg := call(t, cs, "status", map[string]any{"kinds": []string{"x"}}, nil); msg != `invalid kind "x" (use untracked, orphan or invalid)` {
		t.Errorf("status x: %q", msg)
	}
}

func TestWrite(t *testing.T) {
	root := project(t)
	cs := connect(t, Options{Dir: root, AllowWrite: true})

	var set core.SetResult
	if msg := call(t, cs, "set_description", map[string]any{"path": "src/db/", "desc": "Database", "detail": "SQLite"}, &set); msg != "" {
		t.Fatal(msg)
	}
	if want := (core.SetResult{Path: "src/db/", Desc: "Database", Detail: "SQLite"}); set != want {
		t.Errorf("set: %+v, want %+v", set, want)
	}
	// Without detail, the detail is kept.
	if msg := call(t, cs, "set_description", map[string]any{"path": "src/db/", "desc": "DB"}, &set); msg != "" {
		t.Fatal(msg)
	}
	if want := (core.SetResult{Path: "src/db/", Desc: "DB", Detail: "SQLite"}); set != want {
		t.Errorf("set without detail: %+v, want %+v", set, want)
	}

	var moved core.MoveResult
	if msg := call(t, cs, "move_description", map[string]any{"from": "src/", "to": "lib/"}, &moved); msg != "" {
		t.Fatal(msg)
	}
	// src/, main.go and db/.
	if want := (core.MoveResult{From: "src/", To: "lib/", Moved: 3}); moved != want {
		t.Errorf("move: %+v, want %+v", moved, want)
	}

	var removed core.RemoveResult
	if msg := call(t, cs, "remove_description", map[string]any{"path": "old/"}, &removed); msg != "" {
		t.Fatal(msg)
	}
	if want := (core.RemoveResult{Path: "old/", Removed: 1}); removed != want {
		t.Errorf("remove: %+v, want %+v", removed, want)
	}

	b, err := os.ReadFile(filepath.Join(root, ".crube", "structure.txt"))
	if err != nil {
		t.Fatal(err)
	}
	want := "lib/\tSource\n\tmain.go\tEntry point\n\t\t| Starts the server.\n\tdb/\tDB\n\t\t| SQLite\n"
	if string(b) != want {
		t.Errorf("structure.txt:\n%q\nwant:\n%q", b, want)
	}
	// The real files are where they were.
	if _, err := os.Stat(filepath.Join(root, "src", "main.go")); err != nil {
		t.Error(err)
	}
}

func TestErrors(t *testing.T) {
	root := project(t)
	cs := connect(t, Options{Dir: root, AllowWrite: true})
	tests := []struct {
		tool string
		args map[string]any
		want string
	}{
		{"show", map[string]any{"path": "nope"}, "nope: no such file"},
		{"tree", map[string]any{"depth": -1}, "invalid depth -1 (use 0 or more)"},
		{"find", map[string]any{"query": "x", "limit": -1}, "invalid limit -1 (use 0 or more)"},
		{"remove_description", map[string]any{"path": "nope"}, "no entry nope"},
		{"set_description", map[string]any{"path": "a", "desc": "two\nlines"}, "desc must be one line (put more in detail)"},
	}
	for _, tt := range tests {
		if got := call(t, cs, tt.tool, tt.args, nil); got != tt.want {
			t.Errorf("%s %v: %q, want %q", tt.tool, tt.args, got, tt.want)
		}
	}

	// No .crube/: every call says so.
	cs = connect(t, Options{Dir: t.TempDir()})
	if got, want := call(t, cs, "tree", nil, nil), "no .crube/ found (run crube init)"; got != want {
		t.Errorf("no .crube/: %q, want %q", got, want)
	}
}
