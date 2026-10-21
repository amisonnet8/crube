// Package mcpserver is the MCP server of crube (crube mcp,
// docs/reference/mcp.md): each tool calls internal/core and returns what the
// CLI's --json gives, so that the two never behave differently.
package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/amisonnet8/crube/internal/core"
	"github.com/amisonnet8/crube/internal/search"
	"github.com/amisonnet8/crube/internal/status"
)

// Options are what the server runs with.
type Options struct {
	Dir        string // where to look for .crube/ from; "" is the current directory
	Lang       string // the value of CRUBE_LANG when the server started
	AllowWrite bool   // --allow-write: add the tools that write descriptions
	Version    string
}

// The arguments of the tools. A field without omitempty is required.

type treeIn struct {
	Path  string `json:"path,omitempty" jsonschema:"an entry path (such as src/ or data/*.csv): show only that entry and what is below it"`
	Depth int    `json:"depth,omitempty" jsonschema:"the number of levels to show; 0 or omitted for all"`
}

type lsIn struct {
	Path string `json:"path,omitempty" jsonschema:"a directory, relative to the project root; omitted for the root"`
}

type findIn struct {
	Query string `json:"query" jsonschema:"words to find, separated by spaces; all of them must be found"`
	Limit int    `json:"limit,omitempty" jsonschema:"the most results to return; 0 or omitted for all"`
}

type showIn struct {
	Path string `json:"path" jsonschema:"an existing path, relative to the project root"`
}

type statusIn struct {
	Kinds []string `json:"kinds,omitempty" jsonschema:"report only these kinds: untracked, orphan, invalid; omitted for all"`
}

type setIn struct {
	Path   string  `json:"path" jsonschema:"the entry path: a name per level joined with /, a directory ending in / (src/auth/), the last name may be a pattern (data/*.csv)"`
	Desc   string  `json:"desc" jsonschema:"one line: what is there"`
	Detail *string `json:"detail,omitempty" jsonschema:"more lines; omitted keeps the detail as it is, an empty string removes it"`
}

type removeIn struct {
	Path string `json:"path" jsonschema:"the entry path to remove, with every entry below it"`
}

type moveIn struct {
	From string `json:"from" jsonschema:"the entry path the description is at now"`
	To   string `json:"to" jsonschema:"the entry path to move it to (a directory entry stays a directory entry)"`
}

// treeSchema is the output schema of tree, written by hand: the SDK cannot
// make one from core.TreeNode, whose children are TreeNodes again.
var treeSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"here":    map[string]any{"type": "string"},
		"entries": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/$defs/entry"}},
	},
	"required": []string{"here", "entries"},
	"$defs": map[string]any{
		"entry": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name":     map[string]any{"type": "string"},
				"path":     map[string]any{"type": "string"},
				"desc":     map[string]any{"type": "string"},
				"detail":   map[string]any{"type": "string"},
				"here":     map[string]any{"type": "boolean"},
				"children": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/$defs/entry"}},
			},
			"required": []string{"name", "path", "desc", "detail", "here", "children"},
		},
	},
}

// New makes the server with its tools.
func New(o Options) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "crube", Version: o.Version}, nil)

	mcp.AddTool(s, &mcp.Tool{
		Name: "tree",
		Description: "Show the descriptions of the project's files and directories (.crube/) as a tree. " +
			"Use this or find before searching the file system for where something is.",
		OutputSchema: treeSchema,
	}, func(_ context.Context, _ *mcp.CallToolRequest, in treeIn) (*mcp.CallToolResult, *core.TreeResult, error) {
		if err := atLeastZero("depth", in.Depth); err != nil {
			return nil, nil, err
		}
		p, err := open(o)
		if err != nil {
			return nil, nil, err
		}
		r, err := p.Tree(core.TreeOptions{Path: in.Path, Depth: in.Depth})
		return nil, r, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "find",
		Description: "Search the descriptions (entry paths, desc and detail). " +
			"All the words must be found, case-insensitive. Use this before searching the file system.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in findIn) (*mcp.CallToolResult, *search.Found, error) {
		if err := atLeastZero("limit", in.Limit); err != nil {
			return nil, nil, err
		}
		p, err := open(o)
		if err != nil {
			return nil, nil, err
		}
		r, err := p.Find(in.Query, in.Limit)
		return nil, r, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "ls",
		Description: "List a directory of the project with the description of each item. path is relative to the project root.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in lsIn) (*mcp.CallToolResult, *core.LsResult, error) {
		p, err := open(o)
		if err != nil {
			return nil, nil, err
		}
		r, err := p.Ls(in.Path)
		return nil, r, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "show",
		Description: "Show the desc and detail of one existing path.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in showIn) (*mcp.CallToolResult, *core.ShowResult, error) {
		p, err := open(o)
		if err != nil {
			return nil, nil, err
		}
		r, err := p.Show(in.Path)
		return nil, r, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "status",
		Description: "Report paths without a description (untracked), descriptions of paths that are gone (orphan), " +
			"values over the limits (invalid), and coverage.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in statusIn) (*mcp.CallToolResult, *status.Report, error) {
		p, err := open(o)
		if err != nil {
			return nil, nil, err
		}
		r, err := p.Status()
		if err != nil {
			return nil, nil, err
		}
		if err := core.FilterKinds(r, in.Kinds); err != nil {
			return nil, nil, err
		}
		return nil, r, nil
	})

	if !o.AllowWrite {
		return s
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "set_description",
		Description: "Add or change the description of an entry. Writes only .crube/; no project file is touched.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in setIn) (*mcp.CallToolResult, *core.SetResult, error) {
		p, err := open(o)
		if err != nil {
			return nil, nil, err
		}
		r, err := p.Set(in.Path, in.Desc, in.Detail)
		return nil, r, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "remove_description",
		Description: "Remove the description of an entry and of everything below it. The file itself is not deleted.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in removeIn) (*mcp.CallToolResult, *core.RemoveResult, error) {
		p, err := open(o)
		if err != nil {
			return nil, nil, err
		}
		r, err := p.Remove(in.Path)
		return nil, r, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "move_description",
		Description: "Move a description (and those below it) to a new path after a file was moved. The file is not moved.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in moveIn) (*mcp.CallToolResult, *core.MoveResult, error) {
		p, err := open(o)
		if err != nil {
			return nil, nil, err
		}
		r, err := p.Move(in.From, in.To)
		return nil, r, err
	})

	return s
}

// Run serves the tools over stdin and stdout until the client closes them.
func Run(ctx context.Context, o Options) error {
	return New(o).Run(ctx, &mcp.StdioTransport{})
}

// open finds the project for one call, so that changes to .crube/ between
// calls are seen. Real paths are read from the project root.
func open(o Options) (*core.Project, error) {
	p, err := core.Open(core.Options{Dir: o.Dir, Lang: o.Lang})
	if err != nil {
		return nil, err
	}
	return p.AtRoot(), nil
}

func atLeastZero(name string, n int) error {
	if n < 0 {
		return fmt.Errorf("invalid %s %d (use 0 or more)", name, n)
	}
	return nil
}
