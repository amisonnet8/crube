package e2e

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/amisonnet8/crube/internal/mcpserver"
)

// mcpCall starts the server in dir, calls one tool and returns the text it
// returns (the JSON of the result, or the error), and whether it failed.
func mcpCall(t *testing.T, dir string, tool string, args map[string]any) (string, bool) {
	t.Helper()
	ctx := context.Background()
	st, ct := mcp.NewInMemoryTransports()
	ss, err := mcpserver.New(mcpserver.Options{Dir: dir, AllowWrite: true}).Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ss.Close() }()
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "e2e", Version: "0"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cs.Close() }()
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: tool, Arguments: args})
	if err != nil {
		t.Fatal(err)
	}
	return res.Content[0].(*mcp.TextContent).Text, res.IsError
}

// TestMCPSameAsCLI checks that each tool returns what the CLI gives with
// --json, run from the project root (where tree has no mark), and fails with
// the same message (the CLI's without "crube: ").
func TestMCPSameAsCLI(t *testing.T) {
	tests := []struct {
		tool string
		args map[string]any
		cli  []string
	}{
		{"tree", nil, []string{"tree"}},
		{"tree", map[string]any{"path": "src/", "depth": 2}, []string{"tree", "src/", "-d", "2"}},
		{"ls", nil, []string{"ls"}},
		{"ls", map[string]any{"path": "src/auth"}, []string{"ls", "src/auth"}},
		{"find", map[string]any{"query": "セッション"}, []string{"find", "セッション"}},
		{"find", map[string]any{"query": "セッション", "limit": 1}, []string{"find", "セッション", "--limit", "1"}},
		{"show", map[string]any{"path": "data/2026.csv"}, []string{"show", "data/2026.csv"}},
		{"show", map[string]any{"path": "src/auth/session.ts"}, []string{"show", "src/auth/session.ts"}},
		{"status", nil, []string{"status"}},
		{"show", map[string]any{"path": "src/nope.go"}, []string{"show", "src/nope.go"}},
		{"tree", map[string]any{"path": "nope/"}, []string{"tree", "nope/"}},
		{"set_description", map[string]any{"path": "src/db/", "desc": "Database", "detail": "SQLite"},
			[]string{"set", "src/db/", "Database", "--detail", "SQLite"}},
		{"set_description", map[string]any{"path": "x", "desc": strings.Repeat("x", 300)},
			[]string{"set", "x", strings.Repeat("x", 300)}},
		{"remove_description", map[string]any{"path": "src/auth/"}, []string{"rm", "src/auth/"}},
		{"move_description", map[string]any{"from": "src/", "to": "lib/"}, []string{"mv", "src/", "lib/"}},
		{"move_description", map[string]any{"from": "src/", "to": "a.go"}, []string{"mv", "src/", "a.go"}},
	}
	for _, tt := range tests {
		name := strings.Join(tt.cli, " ")
		if len(name) > 40 {
			name = name[:40]
		}
		t.Run(name, func(t *testing.T) {
			// A project each, so that a write by one does not change the other.
			cliDir, _ := cliProject(t)
			mcpDir, _ := cliProject(t)

			args := append([]string{tt.cli[0], "-C", cliDir, "--json"}, tt.cli[1:]...)
			cliOut, code := runCLI(args, nil)
			mcpOut, failed := mcpCall(t, mcpDir, tt.tool, tt.args)

			if code == 2 {
				// An error: the same message.
				if !failed || "crube: "+mcpOut+"\n" != cliOut {
					t.Errorf("CLI: %q\nMCP (failed %v): %q", cliOut, failed, mcpOut)
				}
				return
			}
			if failed {
				t.Fatalf("MCP failed: %s (CLI: %s)", mcpOut, cliOut)
			}
			if !sameJSON(t, mcpOut, cliOut) {
				t.Errorf("CLI:\n%s\nMCP:\n%s", cliOut, mcpOut)
			}
			// A write leaves the same structure file.
			a, errA := os.ReadFile(filepath.Join(cliDir, ".crube", "structure.txt"))
			b, errB := os.ReadFile(filepath.Join(mcpDir, ".crube", "structure.txt"))
			if errA != nil || errB != nil || string(a) != string(b) {
				t.Errorf("structure.txt differs:\nCLI:\n%s\nMCP:\n%s", a, b)
			}
		})
	}
}
