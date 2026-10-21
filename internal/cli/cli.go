// Package cli is the crube command line: it turns arguments into calls to
// internal/core and results into output, and nothing more, so that the CLI
// and the MCP server never behave differently (docs/reference/cli.md).
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/amisonnet8/crube/internal/core"
	"github.com/amisonnet8/crube/internal/lang"
	"github.com/amisonnet8/crube/internal/mcpserver"
)

// Exit codes (docs/reference/cli.md).
const (
	exitOK     = 0
	exitFailed = 1 // status found a problem
	exitError  = 2
)

// Env is what the command line runs with.
type Env struct {
	Getenv  func(string) string
	Stdout  io.Writer
	Stderr  io.Writer
	Version string // set at build time by the release; "" otherwise
	// Width returns the width of the terminal that Stdout is, or 0 when it is
	// not a terminal. nil looks at Stdout itself.
	Width func() int
}

// order is the order of the commands in the usage and in completion.
var order = []string{"init", "ls", "tree", "find", "show", "status", "set", "rm", "mv", "mcp", "completion"}

// usage is what crube --help prints.
var usage = func() string {
	var b strings.Builder
	b.WriteString("usage: crube <command> [arguments]\n\ncommands:\n")
	for _, name := range order {
		fmt.Fprintf(&b, "  %-12s%s\n", name, commands[name].summary)
	}
	b.WriteString("\ncommon options: --json  -C <dir>  --no-ignore\n")
	return b.String()
}()

// command is one subcommand.
type command struct {
	usage    string   // the synopsis after "usage: "
	summary  string   // what it does, in the usage and in completion
	options  []string // the options it accepts
	args     []string // the names of its positional arguments
	min      int      // how many of args are required
	variadic bool     // the last argument takes all the rest
	run      func(r *runner, p *parsed) (int, error)
}

var (
	readOptions = []string{"--json", "-C", "--no-ignore"}
	editOptions = []string{"--json", "-C"}
)

var commands = map[string]*command{
	"init": {
		usage: "crube init [-C <dir>]", summary: "create .crube/", options: []string{"-C"},
		run: (*runner).init,
	},
	"ls": {
		usage: "crube ls [path] [--full] [--json] [-C <dir>] [--no-ignore]", summary: "list a directory with descriptions",
		options: append([]string{"--full"}, readOptions...), args: []string{"path"},
		run: (*runner).ls,
	},
	"tree": {
		usage: "crube tree [path] [-d N] [--json] [-C <dir>] [--no-ignore]", summary: "show the descriptions as a tree",
		options: append([]string{"-d"}, readOptions...), args: []string{"path"},
		run: (*runner).tree,
	},
	"find": {
		usage: "crube find <query> [--limit N] [--json] [-C <dir>] [--no-ignore]", summary: "search paths and descriptions",
		options: append([]string{"--limit"}, readOptions...), args: []string{"query"}, min: 1, variadic: true,
		run: (*runner).find,
	},
	"show": {
		usage: "crube show <path> [--json] [-C <dir>] [--no-ignore]", summary: "show the description of one path",
		options: readOptions, args: []string{"path"}, min: 1,
		run: (*runner).show,
	},
	"status": {
		usage: "crube status [--require all|dirs] [--json] [-C <dir>] [--no-ignore]", summary: "report undescribed, orphan and invalid entries",
		options: append([]string{"--require"}, readOptions...),
		run:     (*runner).status,
	},
	"set": {
		usage: "crube set <path> <desc> [--detail <text>] [--json] [-C <dir>]", summary: "add or change a description",
		options: append([]string{"--detail"}, editOptions...), args: []string{"path", "desc"}, min: 2,
		run: (*runner).set,
	},
	"rm": {
		usage: "crube rm <path> [--json] [-C <dir>]", summary: "remove a description",
		options: editOptions, args: []string{"path"}, min: 1,
		run: (*runner).rm,
	},
	"mv": {
		usage: "crube mv <from> <to> [--json] [-C <dir>]", summary: "move a description",
		options: editOptions, args: []string{"from", "to"}, min: 2,
		run: (*runner).mv,
	},
	"mcp": {
		usage: "crube mcp [--allow-write]", summary: "run as an MCP server over stdio", options: []string{"--allow-write"},
		run: (*runner).mcp,
	},
	"completion": {
		usage: "crube completion <shell>", summary: "print a shell completion script",
		args: []string{"shell"}, min: 1,
		run: (*runner).completion,
	},
}

// Run runs the command line args (without the program name) and returns the
// exit code. Errors go to Stderr as one line "crube: …".
func Run(args []string, env Env) int {
	if len(args) == 0 {
		_, _ = fmt.Fprint(env.Stderr, usage) // nowhere else to report a failure
		return exitError
	}
	switch args[0] {
	case "-h", "--help", "help":
		return say(env, usage)
	case "--version":
		return say(env, "crube "+version(env.Version)+"\n")
	case completeCommand:
		return complete(args[1:], env)
	}

	name := args[0]
	cmd, ok := commands[name]
	if !ok {
		return fail(env, fmt.Errorf("unknown command %q (see crube --help)", name))
	}
	p, err := parse(args[1:], cmd.options)
	if err != nil {
		return fail(env, fmt.Errorf("%s: %w", name, err))
	}
	if p.help {
		return say(env, "usage: "+cmd.usage+"\n")
	}
	switch {
	case len(p.args) < cmd.min:
		return fail(env, fmt.Errorf("%s: missing <%s> (usage: %s)", name, cmd.args[len(p.args)], cmd.usage))
	case len(p.args) > len(cmd.args) && !cmd.variadic:
		return fail(env, fmt.Errorf("%s: too many arguments (usage: %s)", name, cmd.usage))
	}

	r := &runner{env: env, cmd: name}
	code, err := cmd.run(r, p)
	if err != nil {
		return fail(env, err)
	}
	return code
}

// version is the version to show: the one set by the release, else "(devel)".
func version(v string) string {
	if v != "" {
		return v
	}
	return "(devel)"
}

// say writes s to Stdout and ends with exit code 0, or 2 if it cannot.
func say(env Env, s string) int {
	if _, err := io.WriteString(env.Stdout, s); err != nil {
		return fail(env, err)
	}
	return exitOK
}

func fail(env Env, err error) int {
	_, _ = fmt.Fprintln(env.Stderr, "crube:", err) // nowhere else to report a failure
	return exitError
}

// runner runs one command.
type runner struct {
	env Env
	cmd string
}

// open finds the project from the common options.
func (r *runner) open(p *parsed) (*core.Project, error) {
	return core.Open(core.Options{
		Dir:      p.opts["-C"],
		Lang:     r.env.Getenv(lang.EnvVar),
		NoIgnore: p.has("--no-ignore"),
	})
}

// argError marks an error in the arguments with the command, as Run does for
// the errors it finds itself.
func (r *runner) argError(err error) error {
	return fmt.Errorf("%s: %w", r.cmd, err)
}

// writeJSON writes v as indented JSON. HTML characters are kept as they are,
// since descriptions are text for people, not for a web page.
func (r *runner) writeJSON(v any) error {
	enc := json.NewEncoder(r.env.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// output writes v as JSON with --json, else the text that text makes. text
// gets the width to cut descs at: 0 when nothing is to be cut.
func (r *runner) output(p *parsed, v any, text func(b *strings.Builder, width int)) error {
	if p.has("--json") {
		return r.writeJSON(v)
	}
	var b strings.Builder
	text(&b, termWidth(r.env))
	_, err := io.WriteString(r.env.Stdout, b.String())
	return err
}

func (r *runner) init(p *parsed) (int, error) {
	made, err := core.Init(p.opts["-C"], r.env.Getenv(lang.EnvVar))
	for _, m := range made {
		if _, werr := fmt.Fprintln(r.env.Stdout, "created", m); werr != nil {
			return 0, werr
		}
	}
	return exitOK, err
}

func (r *runner) ls(p *parsed) (int, error) {
	proj, err := r.open(p)
	if err != nil {
		return 0, err
	}
	res, err := proj.Ls(arg(p, 0))
	if err != nil {
		return 0, err
	}
	return exitOK, r.output(p, res, func(b *strings.Builder, width int) {
		if p.has("--full") {
			width = 0
		}
		lsText(b, res, width)
	})
}

func (r *runner) tree(p *parsed) (int, error) {
	depth, err := p.count("-d")
	if err != nil {
		return 0, r.argError(err)
	}
	proj, err := r.open(p)
	if err != nil {
		return 0, err
	}
	res, err := proj.Tree(core.TreeOptions{Path: arg(p, 0), Depth: depth, Mark: true})
	if err != nil {
		return 0, err
	}
	return exitOK, r.output(p, res, func(b *strings.Builder, width int) { treeText(b, res, width) })
}

func (r *runner) find(p *parsed) (int, error) {
	limit, err := p.count("--limit")
	if err != nil {
		return 0, r.argError(err)
	}
	proj, err := r.open(p)
	if err != nil {
		return 0, err
	}
	res, err := proj.Find(strings.Join(p.args, " "), limit)
	if err != nil {
		return 0, err
	}
	return exitOK, r.output(p, res, func(b *strings.Builder, width int) { findText(b, res, width) })
}

func (r *runner) show(p *parsed) (int, error) {
	proj, err := r.open(p)
	if err != nil {
		return 0, err
	}
	res, err := proj.Show(arg(p, 0))
	if err != nil {
		return 0, err
	}
	return exitOK, r.output(p, res, func(b *strings.Builder, _ int) { showText(b, res) })
}

func (r *runner) status(p *parsed) (int, error) {
	proj, err := r.open(p)
	if err != nil {
		return 0, err
	}
	res, err := proj.Status()
	if err != nil {
		return 0, err
	}
	failed, err := core.Failed(res, p.opts["--require"])
	if err != nil {
		return 0, r.argError(err)
	}
	if err := r.output(p, res, func(b *strings.Builder, _ int) { statusText(b, res) }); err != nil {
		return 0, err
	}
	if failed {
		return exitFailed, nil
	}
	return exitOK, nil
}

func (r *runner) set(p *parsed) (int, error) {
	proj, err := r.open(p)
	if err != nil {
		return 0, err
	}
	var detail *string
	if d, ok := p.opts["--detail"]; ok {
		detail = &d
	}
	res, err := proj.Set(p.args[0], p.args[1], detail)
	if err != nil {
		return 0, err
	}
	return exitOK, r.edited(p, res)
}

func (r *runner) rm(p *parsed) (int, error) {
	proj, err := r.open(p)
	if err != nil {
		return 0, err
	}
	res, err := proj.Remove(p.args[0])
	if err != nil {
		return 0, err
	}
	return exitOK, r.edited(p, res)
}

func (r *runner) mv(p *parsed) (int, error) {
	proj, err := r.open(p)
	if err != nil {
		return 0, err
	}
	res, err := proj.Move(p.args[0], p.args[1])
	if err != nil {
		return 0, err
	}
	return exitOK, r.edited(p, res)
}

// edited writes the result of set, rm or mv as JSON with --json; without it,
// nothing is written.
func (r *runner) edited(p *parsed, v any) error {
	if !p.has("--json") {
		return nil
	}
	return r.writeJSON(v)
}

// mcp runs the MCP server over stdio until the client closes it. The language
// is checked first, so that a wrong CRUBE_LANG stops the server at once.
func (r *runner) mcp(p *parsed) (int, error) {
	tag := r.env.Getenv(lang.EnvVar)
	if err := lang.Check(tag); err != nil {
		return 0, err
	}
	return exitOK, mcpserver.Run(context.Background(), mcpserver.Options{
		Lang:       tag,
		AllowWrite: p.has("--allow-write"),
		Version:    version(r.env.Version),
	})
}

// arg is the i-th positional argument, or "" when it is not given.
func arg(p *parsed, i int) string {
	if i < len(p.args) {
		return p.args[i]
	}
	return ""
}
