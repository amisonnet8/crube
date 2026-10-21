package e2e

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestGettingStarted runs docs/examples/getting-started.md from the top, in
// one shell, as a reader would.
func TestGettingStarted(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, map[string]string{
		"myapp/go.mod":                  "",
		"myapp/cmd/myapp/main.go":       "",
		"myapp/internal/store/store.go": "",
	})
	s := newShell(t, dir)
	for _, b := range readBlocks(t, "docs/examples/getting-started.md") {
		switch {
		case strings.HasPrefix(b.body, "$ "):
			runSession(t, s, b.body)
		case strings.Contains(b.before, "この時点の `.crube/structure.txt`"):
			got, err := os.ReadFile(filepath.Join(dir, "myapp", ".crube", "structure.txt"))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != b.body {
				t.Errorf(".crube/structure.txt:\n%s\nthe docs say:\n%s", got, b.body)
			}
		}
	}
}

// TestCLIExamples runs the command examples of docs/reference/cli.md, each
// block from the project root of the project it is about.
func TestCLIExamples(t *testing.T) {
	for _, b := range readBlocks(t, "docs/reference/cli.md") {
		if !strings.HasPrefix(b.body, "$ ") {
			continue
		}
		t.Run(b.heading+"/"+strings.SplitN(b.body, "\n", 2)[0], func(t *testing.T) {
			var root string
			switch {
			case b.heading == "crube status" && strings.Contains(b.body, "invalid "):
				root = statusProject(t, true)
			case b.heading == "crube status":
				root = project(t, map[string]string{"a.txt": ""}, "a.txt\tA\n")
			default:
				var link bool
				root, link = cliProject(t)
				if !link && strings.Contains(b.body, "crube show current") {
					t.Skip("cannot make a symbolic link here")
				}
			}
			runSession(t, newShell(t, root), b.body)
		})
	}
}

// TestREADMEExample runs the example at the top of README.md, in the project
// of cli.md.
func TestREADMEExample(t *testing.T) {
	root, _ := cliProject(t)
	n := 0
	for _, b := range readBlocks(t, "README.md") {
		if strings.HasPrefix(b.body, "$ ") {
			runSession(t, newShell(t, root), b.body)
			n++
		}
	}
	if n == 0 {
		t.Error("README.md has no example to run")
	}
}

// jsonExample is how to get the output a JSON example of the docs shows.
type jsonExample struct {
	project func(t *testing.T) string
	dir     string // where to run, from the project root
	args    []string
}

func cliRoot(t *testing.T) string {
	root, _ := cliProject(t)
	return root
}

// TestJSONExamples checks every JSON example of the docs against the output
// of the command it shows. An example the test does not know fails, so that a
// new example gets checked too.
func TestJSONExamples(t *testing.T) {
	examples := map[string]jsonExample{
		"docs/reference/cli.md#crube ls/2":                  {cliRoot, "", []string{"ls", "src", "--json"}},
		"docs/reference/cli.md#crube tree/3":                {cliRoot, "src/auth", []string{"tree", "src/", "-d", "2", "--json"}},
		"docs/reference/cli.md#crube find/2":                {cliRoot, "", []string{"find", "セッション", "--json"}},
		"docs/reference/cli.md#crube show/2":                {cliRoot, "", []string{"show", "data/2026.csv", "--json"}},
		"docs/reference/cli.md#crube status/3":              {func(t *testing.T) string { return statusProject(t, true) }, "", []string{"status", "--json"}},
		"docs/reference/cli.md#crube set/2":                 {cliRoot, "", []string{"set", "data/*.csv", "CSVデータ", "--json"}},
		"docs/examples/ai-workflow.md#status --json の出力例/0": {func(t *testing.T) string { return statusProject(t, false) }, "", []string{"status", "--json"}},
	}
	// JSON in the docs that is not the output of a command.
	notOutput := map[string]bool{
		"docs/reference/mcp.md#クライアントへの登録例/0": true,
	}

	seen := map[string]bool{}
	for _, doc := range []string{"docs/reference/cli.md", "docs/reference/mcp.md", "docs/examples/ai-workflow.md"} {
		for _, b := range readBlocks(t, doc) {
			if b.lang != "json" {
				continue
			}
			key := doc + "#" + b.heading + "/" + strconv.Itoa(b.index)
			seen[key] = true
			if notOutput[key] {
				continue
			}
			ex, ok := examples[key]
			if !ok {
				t.Errorf("%s: a JSON example the test does not know; add how to get it to TestJSONExamples", key)
				continue
			}
			t.Run(key, func(t *testing.T) {
				root := ex.project(t)
				args := append([]string{ex.args[0], "-C", filepath.Join(root, filepath.FromSlash(ex.dir))}, ex.args[1:]...)
				got, _ := runCLI(args, nil)
				if !sameJSON(t, got, b.body) {
					t.Errorf("crube %s\ngot:\n%s\nthe docs say:\n%s", strings.Join(ex.args, " "), got, b.body)
				}
			})
		}
	}
	for key := range examples {
		if !seen[key] {
			t.Errorf("%s: no such JSON example in the docs any more", key)
		}
	}
}

// TestErrorExamples checks that each error line in "出力の共通の決まり" of
// cli.md is what crube prints in some case.
func TestErrorExamples(t *testing.T) {
	type errorCase struct {
		project func(t *testing.T) string // nil: no project
		args    []string                  // "-C <root>" is added after the command when project is set
	}
	cases := map[string]errorCase{
		// The bad line is the 12th: 11 good lines, then one indented with spaces.
		"crube: .crube/structure.txt:12: indentation must be tabs": {func(t *testing.T) string {
			return project(t, nil, linesThenBad(11))
		}, []string{"tree"}},
		"crube: no .crube/ found (run crube init)":    {func(t *testing.T) string { return t.TempDir() }, []string{"ls"}},
		"crube: desc is 300 columns wide (limit 256)": {cliRoot, []string{"set", "a.txt", strings.Repeat("x", 300)}},
		"crube: src/foo.go: no such file":             {cliRoot, []string{"show", "src/foo.go"}},
		"crube: build/x: ignored by .crube/ignore (use --no-ignore)": {func(t *testing.T) string {
			root := cliRoot(t)
			write(t, root, map[string]string{"build/x": "", ".crube/ignore": "build/\n"})
			return root
		}, []string{"show", "build/x"}},
		`crube: unknown command "lss" (see crube --help)`:                                                   {nil, []string{"lss"}},
		"crube: ls: unknown option --foo":                                                                   {nil, []string{"ls", "--foo"}},
		"crube: ls: -C needs a value":                                                                       {nil, []string{"ls", "-C"}},
		`crube: find: invalid --limit "x" (use a number of 1 or more)`:                                      {cliRoot, []string{"find", "--limit", "x", "q"}},
		"crube: ls: too many arguments (usage: crube ls [path] [--full] [--json] [-C <dir>] [--no-ignore])": {nil, []string{"ls", "a", "b"}},
		"crube: set: missing <desc> (usage: crube set <path> <desc> [--detail <text>] [--json] [-C <dir>])": {nil, []string{"set", "x"}},
	}

	n := 0
	for _, b := range readBlocks(t, "docs/reference/cli.md") {
		if b.heading != "出力の共通の決まり" || strings.HasPrefix(b.body, "$ ") {
			continue
		}
		for _, line := range strings.Split(strings.TrimSuffix(b.body, "\n"), "\n") {
			n++
			c, ok := cases[line]
			if !ok {
				t.Errorf("an error example the test does not know: %q", line)
				continue
			}
			args := c.args
			if c.project != nil {
				args = append([]string{args[0], "-C", c.project(t)}, args[1:]...)
			}
			got, code := runCLI(args, nil)
			if code != 2 || got != line+"\n" {
				t.Errorf("crube %s: %d %q, the docs say %q", strings.Join(c.args, " "), code, got, line)
			}
		}
	}
	if n != len(cases) {
		t.Errorf("the docs have %d error examples, the test knows %d", n, len(cases))
	}
}

// linesThenBad is a structure file of n good lines (each its own entry) and
// a line indented with spaces after them.
func linesThenBad(n int) string {
	var b strings.Builder
	for i := range n {
		b.WriteString("f" + strconv.Itoa(i) + "\tF\n")
	}
	b.WriteString("  bad\n")
	return b.String()
}
